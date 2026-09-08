package docker

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	doctype "github.com/docker/docker/api/types"
	conttype "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	voltype "github.com/docker/docker/api/types/volume"

	"github.com/azukaar/cosmos-server/src/utils"
)

// DeploymentLabel is the docker label key used to identify resources that
// belong to a scheduler-managed deployment. The value is the deployment name.
const DeploymentLabel = "cosmos-deployment"

// DeploymentVersionLabel is the docker label key holding the integer spec
// version a container was created from (Deployment.Version, stringified). The
// scheduler compares the value a node reports against the desired version to
// decide whether a node is running a stale spec and needs a rolling re-apply.
const DeploymentVersionLabel = "cosmos-deployment-version"

// DeploymentKeepVolumesLabel ("true") marks a deployment as data-bearing: RemoveByDeploymentLabel
// preserves its named volumes regardless of the keepVolumes argument. Stamped on containers AND volumes.
const DeploymentKeepVolumesLabel = "cosmos-deployment-keep-volumes"

// DeploymentRoutesLabel holds the comma-joined proxy route names carried by the deployment's
// compose, so teardown can strip them from config after the KV record is gone.
const DeploymentRoutesLabel = "cosmos-deployment-routes"

// RemoveByDeploymentLabel discovers and tears down every container, network, and
// (unless keepVolumes is set) volume labeled cosmos-deployment=<deploymentName>.
// Mirrors the sequence the servapps delete modal uses (containers → 1s pause →
// networks + volumes in parallel). Per-resource failures are collected, not
// fatal — partial teardown is still progress.
//
// keepVolumes skips the volume teardown: used by the scheduler's re-apply path
// when a deployment's spec version changed, so the new containers reattach to
// the existing volumes and persistent data survives the update. A full delete
// (ActionRemove) passes keepVolumes=false.
func RemoveByDeploymentLabel(deploymentName string, OnLog func(string), keepVolumes bool) []error {
	if OnLog == nil {
		OnLog = func(string) {}
	}

	if err := Connect(); err != nil {
		utils.Error("[SCHED-NODE] RemoveByDeploymentLabel: docker connect failed", err)
		return []error{err}
	}

	labelFilter := filters.NewArgs()
	labelFilter.Add("label", DeploymentLabel+"="+deploymentName)

	var errs []error

	// 1. Containers: kill (best-effort) then remove with Force.
	containers, err := DockerClient.ContainerList(DockerContext, conttype.ListOptions{
		All:     true,
		Filters: labelFilter,
	})
	if err != nil {
		utils.Error("[SCHED-NODE] RemoveByDeploymentLabel: ContainerList failed for "+deploymentName, err)
		OnLog(utils.DoErr("Failed to list containers for deployment %s: %s\n", deploymentName, err.Error()))
		errs = append(errs, err)
	}

	utils.Log(fmt.Sprintf("[SCHED-NODE] remove deployment=%s (label filter) containers=%d", deploymentName, len(containers)))
	OnLog(fmt.Sprintf("Removing deployment %s: %d container(s)\n", deploymentName, len(containers)))

	// a keep-volumes label on any container overrides the caller's keepVolumes
	if !keepVolumes {
		for _, c := range containers {
			if c.Labels[DeploymentKeepVolumesLabel] == "true" {
				keepVolumes = true
				utils.Log("[SCHED-NODE] remove deployment=" + deploymentName + ": keep-volumes label set, preserving volumes")
				OnLog(fmt.Sprintf("Deployment %s is marked keep-volumes: named volumes are preserved\n", deploymentName))
				break
			}
		}
	}

	for _, c := range containers {
		name := c.ID
		if len(c.Names) > 0 {
			name = c.Names[0]
		}

		// Kill is best-effort: a stopped container will return an error we can ignore.
		if killErr := DockerClient.ContainerKill(DockerContext, c.ID, "SIGKILL"); killErr != nil {
			utils.Debug("[SCHED-NODE] ContainerKill " + name + ": " + killErr.Error())
		}

		if rmErr := DockerClient.ContainerRemove(DockerContext, c.ID, conttype.RemoveOptions{Force: true}); rmErr != nil {
			utils.Warn("[SCHED-NODE] failed to remove container " + name + ": " + rmErr.Error())
			OnLog(utils.DoWarn("Failed to remove container %s: %s\n", name, rmErr.Error()))
			errs = append(errs, rmErr)
		} else {
			utils.Debug("[SCHED-NODE] removed container " + name)
			OnLog(fmt.Sprintf("Removed container %s\n", name))
		}
	}

	// 2. Pause to let docker release references — matches the UI's setTimeout(1000).
	time.Sleep(1 * time.Second)

	// 3. Networks and volumes in parallel.
	var wg sync.WaitGroup
	var errMu sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()

		networks, nlErr := DockerClient.NetworkList(DockerContext, doctype.NetworkListOptions{Filters: labelFilter})
		if nlErr != nil {
			utils.Error("[SCHED-NODE] RemoveByDeploymentLabel: NetworkList failed for "+deploymentName, nlErr)
			OnLog(utils.DoErr("Failed to list networks for deployment %s: %s\n", deploymentName, nlErr.Error()))
			errMu.Lock()
			errs = append(errs, nlErr)
			errMu.Unlock()
			return
		}

		for _, n := range networks {
			if rmErr := DockerClient.NetworkRemove(DockerContext, n.ID); rmErr != nil {
				utils.Warn("[SCHED-NODE] failed to remove network " + n.Name + ": " + rmErr.Error())
				OnLog(utils.DoWarn("Failed to remove network %s: %s\n", n.Name, rmErr.Error()))
				errMu.Lock()
				errs = append(errs, rmErr)
				errMu.Unlock()
			} else {
				utils.Debug("[SCHED-NODE] removed network " + n.Name)
				OnLog(fmt.Sprintf("Removed network %s\n", n.Name))
			}
		}
	}()

	if keepVolumes {
		utils.Debug("[SCHED-NODE] RemoveByDeploymentLabel: keepVolumes set, leaving volumes for " + deploymentName)
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()

			volumes, vlErr := DockerClient.VolumeList(DockerContext, voltype.ListOptions{Filters: labelFilter})
			if vlErr != nil {
				utils.Error("[SCHED-NODE] RemoveByDeploymentLabel: VolumeList failed for "+deploymentName, vlErr)
				OnLog(utils.DoErr("Failed to list volumes for deployment %s: %s\n", deploymentName, vlErr.Error()))
				errMu.Lock()
				errs = append(errs, vlErr)
				errMu.Unlock()
				return
			}

			for _, v := range volumes.Volumes {
				if v == nil {
					continue
				}
				// the flag is also stamped on the volume itself: a re-sent remove can arrive after the containers are gone
				if v.Labels[DeploymentKeepVolumesLabel] == "true" {
					utils.Log("[SCHED-NODE] remove deployment=" + deploymentName + ": volume " + v.Name + " is marked keep-volumes, preserving it")
					OnLog(fmt.Sprintf("Volume %s is marked keep-volumes: preserved\n", v.Name))
					continue
				}
				if rmErr := DockerClient.VolumeRemove(DockerContext, v.Name, true); rmErr != nil {
					utils.Warn("[SCHED-NODE] failed to remove volume " + v.Name + ": " + rmErr.Error())
					OnLog(utils.DoWarn("Failed to remove volume %s: %s\n", v.Name, rmErr.Error()))
					errMu.Lock()
					errs = append(errs, rmErr)
					errMu.Unlock()
				} else {
					utils.Debug("[SCHED-NODE] removed volume " + v.Name)
					OnLog(fmt.Sprintf("Removed volume %s\n", v.Name))
				}
			}
		}()
	}

	wg.Wait()

	utils.Log(fmt.Sprintf("[SCHED-NODE] remove deployment=%s done errors=%d", deploymentName, len(errs)))
	OnLog(fmt.Sprintf("Remove deployment %s complete (%d errors)\n", deploymentName, len(errs)))

	return errs
}

// RouteNamesByDeploymentLabel returns the distinct proxy route names recorded on the
// deployment's containers via DeploymentRoutesLabel; called by the scheduler BEFORE teardown.
func RouteNamesByDeploymentLabel(deploymentName string) ([]string, error) {
	if err := Connect(); err != nil {
		return nil, err
	}

	labelFilter := filters.NewArgs()
	labelFilter.Add("label", DeploymentLabel+"="+deploymentName)

	containers, err := DockerClient.ContainerList(DockerContext, conttype.ListOptions{
		All:     true,
		Filters: labelFilter,
	})
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	for _, c := range containers {
		raw := c.Labels[DeploymentRoutesLabel]
		if raw == "" {
			continue
		}
		for _, name := range strings.Split(raw, ",") {
			if name = strings.TrimSpace(name); name != "" {
				seen[name] = struct{}{}
			}
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// ContainerIDsByDeploymentLabel returns the container IDs currently labeled
// cosmos-deployment=<name>. Used by the scheduler's waitForRunning to poll
// container state after an apply.
func ContainerIDsByDeploymentLabel(deploymentName string) ([]string, error) {
	if err := Connect(); err != nil {
		return nil, err
	}

	labelFilter := filters.NewArgs()
	labelFilter.Add("label", DeploymentLabel+"="+deploymentName)

	containers, err := DockerClient.ContainerList(DockerContext, conttype.ListOptions{
		All:     true,
		Filters: labelFilter,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(containers))
	for _, c := range containers {
		ids = append(ids, c.ID)
	}
	return ids, nil
}

// ListDeploymentNamesRunningHere returns the distinct deployment names that have
// at least one actually-running container labeled with DeploymentLabel on this
// host. Called from the heartbeat loop to populate NodeHeartbeat.RunningDeployments.
// Docker is authoritative — no in-memory state is consulted.
//
// Only containers in the "running" state count: a container that was created but
// failed to start (state "created"), or one that has since exited/died, is NOT a
// live replica. Counting those would make the scheduler treat a broken placement
// as healthy and never re-apply it (and the deployments health tab would show it
// as up). We still list All:true so the docker-side status filter, not a default
// running-only list, is what scopes the result — keeping the intent explicit.
// deploymentContainerAlive is the scheduler's notion of a live replica: running, or a lazy container asleep by design.
func deploymentContainerAlive(c doctype.Container) bool {
	return c.State == "running" || IsLazyLabels(c.Labels)
}

func ListDeploymentNamesRunningHere() ([]string, error) {
	if err := Connect(); err != nil {
		return nil, err
	}

	labelFilter := filters.NewArgs()
	labelFilter.Add("label", DeploymentLabel)

	containers, err := DockerClient.ContainerList(DockerContext, conttype.ListOptions{
		All:     true,
		Filters: labelFilter,
	})
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	for _, c := range containers {
		if !deploymentContainerAlive(c) {
			continue
		}
		if name := c.Labels[DeploymentLabel]; name != "" {
			seen[name] = struct{}{}
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	return names, nil
}

// ListDeploymentVersionsRunningHere returns, per deployment with at least one running container on this host
func ListDeploymentVersionsRunningHere() (map[string]int, error) {
	if err := Connect(); err != nil {
		return nil, err
	}

	labelFilter := filters.NewArgs()
	labelFilter.Add("label", DeploymentLabel)

	containers, err := DockerClient.ContainerList(DockerContext, conttype.ListOptions{
		All:     true,
		Filters: labelFilter,
	})
	if err != nil {
		return nil, err
	}

	versions := map[string]int{}
	for _, c := range containers {
		if !deploymentContainerAlive(c) {
			continue
		}
		name := c.Labels[DeploymentLabel]
		if name == "" {
			continue
		}
		// Missing / unparseable label → version 0.
		v := 0
		if raw := c.Labels[DeploymentVersionLabel]; raw != "" {
			if parsed, perr := strconv.Atoi(raw); perr == nil {
				v = parsed
			}
		}
		if existing, seen := versions[name]; !seen || v < existing {
			versions[name] = v
		}
	}
	return versions, nil
}

// DeploymentRunningAtVersion reports whether every named container is alive with this deployment's label at exactly the given spec version.
func DeploymentRunningAtVersion(deploymentName string, containerNames []string, version int) bool {
	if len(containerNames) == 0 {
		return false
	}
	if err := Connect(); err != nil {
		return false
	}

	labelFilter := filters.NewArgs()
	labelFilter.Add("label", DeploymentLabel+"="+deploymentName)

	containers, err := DockerClient.ContainerList(DockerContext, conttype.ListOptions{
		All:     true,
		Filters: labelFilter,
	})
	if err != nil {
		return false
	}

	want := strconv.Itoa(version)
	runningAt := map[string]bool{}
	for _, c := range containers {
		if !deploymentContainerAlive(c) {
			continue
		}
		if c.Labels[DeploymentVersionLabel] != want {
			continue
		}
		for _, n := range c.Names {
			runningAt[strings.TrimPrefix(n, "/")] = true
		}
	}

	for _, name := range containerNames {
		if !runningAt[name] {
			return false
		}
	}
	return true
}

func CleanupExitedDeploymentContainers() {
	if err := Connect(); err != nil {
		utils.Error("CleanupExitedDeploymentContainers: docker connect failed", err)
		return
	}

	labelFilter := filters.NewArgs()
	labelFilter.Add("label", DeploymentLabel)
	labelFilter.Add("status", "exited")

	containers, err := DockerClient.ContainerList(DockerContext, conttype.ListOptions{
		All:     true,
		Filters: labelFilter,
	})
	if err != nil {
		utils.Error("CleanupExitedDeploymentContainers: ContainerList failed", err)
		return
	}

	grace := 30 * time.Minute

	for _, c := range containers {
		// Defensive: the status filter should already scope to exited, but guard
		// in case the daemon returns a broader set.
		if c.State != "" && c.State != "exited" {
			continue
		}
		// a lazy replica is exited by design (stopped by the reaper); the scheduler counts it as alive
		if deploymentContainerAlive(c) {
			continue
		}

		name := c.ID
		if len(c.Names) > 0 {
			name = c.Names[0]
		}
		deployment := c.Labels[DeploymentLabel]

		insp, ierr := DockerClient.ContainerInspect(DockerContext, c.ID)
		if ierr != nil {
			utils.Warn("CleanupExitedDeploymentContainers: inspect " + name + ": " + ierr.Error())
			continue
		}
		if insp.State == nil {
			continue
		}
		finishedAt, perr := time.Parse(time.RFC3339Nano, insp.State.FinishedAt)
		if perr != nil {
			// Unparseable finish time — don't guess, skip rather than risk
			// removing a container that only just stopped.
			utils.Warn("CleanupExitedDeploymentContainers: bad FinishedAt for " + name + ": " + insp.State.FinishedAt)
			continue
		}
		if time.Since(finishedAt) < grace {
			continue
		}

		// RemoveVolumes stays false: container only.
		if rmErr := DockerClient.ContainerRemove(DockerContext, c.ID, conttype.RemoveOptions{Force: true}); rmErr != nil {
			utils.Warn("CleanupExitedDeploymentContainers: failed to remove " + name + ": " + rmErr.Error())
			continue
		}
		utils.Log(fmt.Sprintf("[SCHED-NODE] cleaned up exited container %s (deployment=%s, exited %s)", name, deployment, finishedAt.Format(time.RFC3339)))
	}
}

func ContainerIsRunning(containerID string) (bool, error) {
	if err := Connect(); err != nil {
		return false, err
	}

	insp, err := DockerClient.ContainerInspect(DockerContext, containerID)
	if err != nil {
		return false, err
	}
	if insp.State == nil {
		return false, nil
	}
	return insp.State.Running, nil
}
