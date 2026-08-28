package docker

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	doctype "github.com/docker/docker/api/types"
	conttype "github.com/docker/docker/api/types/container"

	"github.com/azukaar/cosmos-server/src/utils"
)

// Container dependency conditions, mirroring docker-compose's depends_on
// condition field. A dependency without an explicit condition is treated as
// service_started (docker-compose default).
const (
	DepConditionStarted          = "service_started"
	DepConditionHealthy          = "service_healthy"
	DepConditionCompletedSuccess = "service_completed_successfully"

	// Cosmos-native extension (same semantics as compose, different key so
	// compose specs keep working unchanged).
	depConditionRestart = "restart"
)

// Labels used to persist the dependency graph on the container itself so the
// runtime restart/recreate paths (which go through EditContainer, not
// CreateService) can honor depends_on too.
const (
	depLabelPrefix = "cosmos.depends_on."
	depLabelList   = "cosmos.depends_on.list"
)

// dependencyTimeout is the maximum time we wait for a dependency condition
// (service_healthy / service_completed_successfully) to be satisfied before
// giving up. service_started only waits for the container to be running,
// which is fast.
const dependencyTimeout = 10 * time.Minute

func depLabelFor(service string) string {
	return depLabelPrefix + service
}

// ValidDepCondition reports whether cond is a condition value we understand.
// Unknown conditions degrade to service_started (compose ignores unknown
// values for ordering; we do the same but keep the value in labels so the
// graph is still visible).
func ValidDepCondition(cond string) bool {
	switch cond {
	case DepConditionStarted, DepConditionHealthy, DepConditionCompletedSuccess, depConditionRestart:
		return true
	}
	return false
}

// NormalizeDepCondition maps an empty/invalid condition to service_started.
func NormalizeDepCondition(cond string) string {
	if !ValidDepCondition(cond) {
		return DepConditionStarted
	}
	return cond
}

// DependsOnFromLabels reconstructs the depends_on map from the labels Cosmos
// writes at container creation. Returns an empty map when the container has
// no dependency labels (the common case).
func DependsOnFromLabels(conf *conttype.Config) map[string]string {
	deps := map[string]string{}
	if conf == nil || conf.Labels == nil {
		return deps
	}

	list := conf.Labels[depLabelList]
	if list == "" {
		return deps
	}

	for _, service := range strings.Split(list, ",") {
		service = strings.TrimSpace(service)
		if service == "" {
			continue
		}
		cond := conf.Labels[depLabelFor(service)]
		deps[service] = NormalizeDepCondition(cond)
	}
	return deps
}

// SetDependsOnLabels writes the depends_on graph as labels on the container
// config (mutating conf.Labels). Call this right before ContainerCreate so
// the graph survives re-creates and can be inspected at runtime.
func SetDependsOnLabels(conf *conttype.Config, deps map[string]string) {
	if conf == nil {
		return
	}
	if conf.Labels == nil {
		conf.Labels = make(map[string]string)
	}

	keys := make([]string, 0, len(deps))
	for service := range deps {
		keys = append(keys, service)
	}
	sort.Strings(keys)

	conf.Labels[depLabelList] = strings.Join(keys, ",")
	for _, service := range keys {
		conf.Labels[depLabelFor(service)] = NormalizeDepCondition(deps[service])
	}
}

// depConditionMet reports whether dep (as currently inspected) satisfies the
// condition cond. state may be nil for a container that no longer exists.
func depConditionMet(dep doctype.ContainerJSON, cond string) bool {
	if cond == "" {
		cond = DepConditionStarted
	}

	if dep.ContainerJSONBase == nil || dep.State == nil {
		return false
	}

	switch cond {
	case DepConditionStarted:
		return dep.State.Running
	case DepConditionHealthy:
		if !dep.State.Running {
			return false
		}
		// A container without a healthcheck never becomes "healthy"; treat it
		// as healthy once it is running (same as compose's service_healthy
		// behavior for healthcheck-less images).
		if dep.Config == nil || dep.Config.Healthcheck == nil {
			return true
		}
		return dep.State.Health != nil && dep.State.Health.Status == doctype.Healthy
	case DepConditionCompletedSuccess:
		return !dep.State.Running && dep.State.ExitCode == 0
	default:
		// Unknown condition: fall back to started.
		return dep.State.Running
	}
}

// WaitForDepCondition polls depName until cond is met, or returns an error
// after dependencyTimeout. A missing dependency is an error (mirrors compose
// depending on a service that does not exist).
func WaitForDepCondition(ctx context.Context, depName string, cond string) error {
	if cond == "" {
		cond = DepConditionStarted
	}

	deadline := time.Now().Add(dependencyTimeout)

	for {
		dep, err := DockerClient.ContainerInspect(ctx, depName)
		if err != nil {
			return fmt.Errorf("depends_on: cannot inspect dependency %q: %w", depName, err)
		}
		if depConditionMet(dep, cond) {
			utils.Debug(fmt.Sprintf("depends_on: %s condition %q met for %s", depName, cond, depName))
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("depends_on: dependency %q did not satisfy condition %q within %s", depName, cond, dependencyTimeout)
		}
		time.Sleep(1 * time.Second)
	}
}

// WaitForDependsOn waits for every dependency of containerName, using the
// dependency graph persisted on the container's labels. It is used by the
// runtime paths (restart / recreate / update) that do not go through
// CreateService's pre-start ordering. Missing dependency containers are
// considered an error and surfaced.
func WaitForDependsOn(ctx context.Context, containerName string) error {
	current, err := DockerClient.ContainerInspect(ctx, containerName)
	if err != nil {
		return fmt.Errorf("depends_on: cannot inspect container %q: %w", containerName, err)
	}

	deps := DependsOnFromLabels(current.Config)
	if len(deps) == 0 {
		return nil
	}

	// Deterministic order (sorted labels) so logs are stable.
	names := make([]string, 0, len(deps))
	for name := range deps {
		names = append(names, name)
	}
	sort.Strings(names)

	utils.Log(fmt.Sprintf("depends_on: container %s waiting for %d dependencies: %s",
		containerName, len(names), strings.Join(names, ", ")))

	for _, depName := range names {
		cond := deps[depName]
		utils.Log(fmt.Sprintf("depends_on: waiting for %s (%s)", depName, cond))
		if err := WaitForDepCondition(ctx, depName, cond); err != nil {
			return err
		}
	}
	return nil
}

// ReorderDependedOn is the boot-up cascade: after containerID starts, any
// container whose depends_on graph references it (directly or transitively)
// gets started again in dependency order. Docker knows nothing about
// depends_on, so without this a dependent that crashed (or never started on
// boot) would stay down even though its dependency is back.
//
// It is deliberately best-effort and non-blocking:
//   - only containers that are stopped AND depend on containerID are started;
//     running dependents are left alone (no restarts, no churn)
//   - failures are logged, never fatal
//   - runs in its own goroutine (see onDockerStarted)
func ReorderDependedOn(containerID string) {
	startedName := ""
	started, err := DockerClient.ContainerInspect(DockerContext, containerID)
	if err == nil && started.ContainerJSONBase != nil && started.Name != "" {
		startedName = started.Name[1:]
	}

	containers, err := ListContainers()
	if err != nil {
		utils.Debug("ReorderDependedOn: cannot list containers: " + err.Error())
		return
	}

	// Build a name -> (config, stopped?, running?) map once.
	byName := map[string]depContainerInfo{}
	for _, c := range containers {
		full, err := DockerClient.ContainerInspect(DockerContext, c.ID)
		if err != nil {
			continue
		}
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
		} else if full.ContainerJSONBase != nil {
			name = full.Name
		}
		if name == "" {
			continue
		}
		byName[name] = depContainerInfo{full: full, name: name}
	}

	// Find all containers that (transitively) depend on the started one and
	// are currently stopped.
	stoppedDependents := []string{}
	seen := map[string]bool{}
	var collect func(name string)
	collect = func(depName string) {
		for cName, cInfo := range byName {
			if seen[cName] {
				continue
			}
			deps := DependsOnFromLabels(cInfo.full.Config)
			if _, ok := deps[depName]; !ok {
				continue
			}
			seen[cName] = true
			if !cInfo.full.State.Running {
				stoppedDependents = append(stoppedDependents, cName)
			}
			collect(cName[1:]) // transitive
		}
	}
	collect(startedName)

	if len(stoppedDependents) == 0 {
		utils.Debug("ReorderDependedOn: no stopped dependents for " + startedName)
		return
	}

	utils.Log(fmt.Sprintf("ReorderDependedOn: %s started; waking %d stopped dependents", startedName, len(stoppedDependents)))

	// Start them in dependency order, waiting for each one's own deps.
	ordered := OrderByDependencies(stoppedDependents, depContainerConfigs(byName))
	for _, name := range ordered {
		cInfo := byName[name]
		utils.Log(fmt.Sprintf("ReorderDependedOn: starting %s after %s", name, startedName))
		if errW := WaitForDependsOn(DockerContext, cInfo.full.ID); errW != nil {
			utils.Error("ReorderDependedOn: dependency wait failed for "+name, errW)
			continue
		}
		if errS := DockerClient.ContainerStart(DockerContext, cInfo.full.ID, conttype.StartOptions{}); errS != nil {
			utils.Error("ReorderDependedOn: cannot start "+name, errS)
		}
	}
}

// depContainerInfo carries the few fields ReorderDependedOn needs per container.
type depContainerInfo struct {
	full doctype.ContainerJSON
	name string
}

// depContainerConfigs adapts map[string]depContainerInfo into the
// map[string]types.ContainerJSON that OrderByDependencies expects.
func depContainerConfigs(byName map[string]depContainerInfo) map[string]doctype.ContainerJSON {
	out := make(map[string]doctype.ContainerJSON, len(byName))
	for k, v := range byName {
		out[k] = v.full
	}
	return out
}
