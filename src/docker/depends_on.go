package docker

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	doctype "github.com/docker/docker/api/types"
	conttype "github.com/docker/docker/api/types/container"

	"github.com/azukaar/cosmos-server/src/utils"
)

// Dependency conditions, mirroring docker-compose's depends_on condition
// field exactly (compose-go types.go). A dependency without an explicit
// condition defaults to service_started, exactly like docker-compose.
const (
	DepConditionStarted          = "service_started"    // default
	DepConditionRunningOrHealthy = "running_or_healthy" // compose CLI's implicit fallback
	DepConditionHealthy          = "service_healthy"
	DepConditionCompletedSuccess = "service_completed_successfully"
)

// composeDependenciesLabel is the SAME label docker-compose itself uses to
// persist the depends_on field on each container
// (pkg/api/labels.go: com.docker.compose.depends_on), value format:
//
//	"<service>:<condition>:<restart>,<service2>:<condition2>:<restart2>,..."
//
// Matching compose exactly means a stack that was created by docker-compose
// and then imported into Cosmos keeps working, and vice-versa. This is the
// only durable, per-container record of the depends_on graph (compose stores
// it nowhere else).
const composeDependenciesLabel = "com.docker.compose.depends_on"

// dependencyTimeout is the max time we wait for a dependency condition
// (service_healthy / service_completed_successfully / running_or_healthy) to
// be satisfied before giving up. service_started is NOT waited on here — it is
// handled purely by the DAG start order, same as docker-compose
// (waitDependencies returns immediately for ServiceConditionStarted).
const dependencyTimeout = 10 * time.Minute

// DependsOnConfig mirrors compose-go's DependsOnConfig map value.
type dependsOnEntry struct {
	Condition string
	Restart   bool
	Required  bool
}

// ValidDepCondition reports whether cond is a condition we understand.
func ValidDepCondition(cond string) bool {
	switch cond {
	case DepConditionStarted, DepConditionRunningOrHealthy, DepConditionHealthy, DepConditionCompletedSuccess:
		return true
	}
	return false
}

// NormalizeDepCondition maps an empty/unknown condition to service_started
// (compose-default), with one exception: compose's CLI treats a missing/empty
// condition as running_or_healthy (its `ServiceConditionRunningOrHealthy`
// const). We keep empty => service_started when we WRITE the label (compose-go
// writes whatever was parsed, default service_started), but when we READ back
// a legacy label fragment without a condition we fall back to running_or_healthy
// exactly like compose.go does.
func NormalizeDepCondition(cond string) string {
	if cond == "" {
		return DepConditionStarted
	}
	if !ValidDepCondition(cond) {
		return DepConditionStarted
	}
	return cond
}

// DependsOnFromLabels reconstructs the depends_on map from a container's
// Config.Labels, exactly like docker-compose's projectFromName does
// (pkg/compose/compose.go). Returns an empty map when there is no label.
func DependsOnFromLabels(conf *conttype.Config) map[string]dependsOnEntry {
	deps := map[string]dependsOnEntry{}
	if conf == nil || conf.Labels == nil {
		return deps
	}

	raw := conf.Labels[composeDependenciesLabel]
	if raw == "" {
		return deps
	}

	for _, dc := range strings.Split(raw, ",") {
		dc = strings.TrimSpace(dc)
		if dc == "" {
			continue
		}
		dcArr := strings.Split(dc, ":")
		condition := DepConditionRunningOrHealthy
		restart := true
		required := true
		dependency := dcArr[0]

		if len(dcArr) > 1 {
			condition = dcArr[1]
			if len(dcArr) > 2 {
				restart, _ = strconv.ParseBool(dcArr[2])
			}
		}

		deps[dependency] = dependsOnEntry{
			Condition: condition,
			Restart:   restart,
			Required:  required,
		}
	}
	return deps
}

// SetDependsOnLabels writes the depends_on graph as compose's
// com.docker.compose.depends_on label, in compose's own
// "<svc>:<cond>:<restart>,..." format. Call this right before
// ContainerCreate so the graph survives re-creates and can be inspected at
// runtime (compose does the same).
func SetDependsOnLabels(conf *conttype.Config, deps map[string]dependsOnEntry) {
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

	parts := make([]string, 0, len(keys))
	for _, service := range keys {
		d := deps[service]
		if d.Condition == "" {
			d.Condition = DepConditionStarted
		}
		parts = append(parts, fmt.Sprintf("%s:%s:%t", service, d.Condition, d.Restart))
	}
	conf.Labels[composeDependenciesLabel] = strings.Join(parts, ",")
}

// DepConditions returns the map service->condition for convenience (used by
// the HTTP/ordering layer).
func DepConditions(conf *conttype.Config) map[string]string {
	out := map[string]string{}
	for k, v := range DependsOnFromLabels(conf) {
		out[k] = v.Condition
	}
	return out
}

// depConditionMet reports whether dep (as currently inspected) satisfies the
// condition cond. Models compose's isServiceHealthy / isServiceCompleted.
func depConditionMet(dep doctype.ContainerJSON, cond string) bool {
	if dep.ContainerJSONBase == nil || dep.State == nil {
		return false
	}

	switch cond {
	case DepConditionStarted, DepConditionRunningOrHealthy:
		// compose's checkDependencyRunningOrHealthy: healthy if running, or
		// falls back to "running" when there's no healthcheck.
		if dep.State.Status == "exited" {
			return false
		}
		noHealthcheck := dep.Config == nil || dep.Config.Healthcheck == nil ||
			(len(dep.Config.Healthcheck.Test) > 0 && dep.Config.Healthcheck.Test[0] == "NONE")
		if noHealthcheck {
			return dep.State.Running
		}
		return dep.State.Health != nil && dep.State.Health.Status == doctype.Healthy
	case DepConditionHealthy:
		if dep.State.Status == "exited" {
			return false
		}
		if dep.Config == nil || dep.Config.Healthcheck == nil ||
			(len(dep.Config.Healthcheck.Test) > 0 && dep.Config.Healthcheck.Test[0] == "NONE") {
			// compose with fallbackRunning=false errors; for robustness treat a
			// healthcheck-less running container as unhealthy-ish but keep
			// waiting is wrong -> treat as not-met so dependent keeps waiting.
			return false
		}
		return dep.State.Health != nil && dep.State.Health.Status == doctype.Healthy
	case DepConditionCompletedSuccess:
		return !dep.State.Running && dep.State.ExitCode == 0
	default:
		return dep.State.Running
	}
}

// WaitForDepCondition polls depName until cond is met, or returns an error
// after dependencyTimeout. A missing dependency is an error (compose:
// "%s is missing dependency %s").
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
// depends_on graph persisted on the container's compose label. It is used by
// the runtime paths (restart / recreate / update). mirroring compose, only
// conditions other than service_started are waited on: service_started is
// already guaranteed by start ordering.
func WaitForDependsOn(ctx context.Context, containerName string) error {
	current, err := DockerClient.ContainerInspect(ctx, containerName)
	if err != nil {
		return fmt.Errorf("depends_on: cannot inspect container %q: %w", containerName, err)
	}

	deps := DependsOnFromLabels(current.Config)
	if len(deps) == 0 {
		return nil
	}

	// Deterministic order (sorted) so logs are stable.
	names := make([]string, 0, len(deps))
	for name := range deps {
		names = append(names, name)
	}
	sort.Strings(names)

	waiting := []string{}
	for _, name := range names {
		if deps[name].Condition != DepConditionStarted {
			waiting = append(waiting, name)
		}
	}
	if len(waiting) == 0 {
		return nil
	}

	utils.Log(fmt.Sprintf("depends_on: container %s waiting for %d dependencies: %s",
		containerName, len(waiting), strings.Join(waiting, ", ")))

	for _, depName := range waiting {
		cond := deps[depName].Condition
		utils.Log(fmt.Sprintf("depends_on: waiting for %s (%s)", depName, cond))
		if err := WaitForDepCondition(ctx, depName, cond); err != nil {
			return err
		}
	}
	return nil
}

// ReorderDependedOn is the boot-up cascade: after containerID starts, any
// container whose depends_on graph references it (directly or transitively)
// gets started again in dependency order, plus any stopped container that
// shares its network namespace (network_mode: container:) since that reference
// requires the target to exist. Best-effort and non-blocking: running
// dependents are left alone, failures are logged not fatal, runs in its own
// goroutine (see onDockerStarted).
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

	// Find containers that (transitively) depend on the started one, or share
	// its network namespace, and are currently stopped.
	stoppedDependents := []string{}
	seen := map[string]bool{}
	var collect func(depName string)
	collect = func(depName string) {
		for cName, cInfo := range byName {
			if seen[cName] {
				continue
			}
			deps := DependsOnFromLabels(cInfo.full.Config)
			_, hasDep := deps[depName]
			sharesNetNS := strings.HasPrefix(string(cInfo.full.HostConfig.NetworkMode), "container:"+depName)
			if !hasDep && !sharesNetNS {
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
