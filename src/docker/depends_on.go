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
	// its network namespace, and are currently stopped. Scoped to the same
	// stack, and only depends_on entries with restart:true or network_mode
	// references are honored (docker-compose parity).
	startedStack := ""
	if started.ContainerJSONBase != nil && started.Config != nil {
		startedStack = ContainerStack(started.Config)
	}

	stoppedDependents := []string{}
	seen := map[string]bool{}
	var collect func(depName string)
	collect = func(depName string) {
		for cName, cInfo := range byName {
			if seen[cName] {
				continue
			}
			deps := DependsOnFromLabels(cInfo.full.Config)
			depEntry, hasDep := deps[depName]
			sharesNetNS := strings.HasPrefix(string(cInfo.full.HostConfig.NetworkMode), "container:"+depName) ||
				strings.HasPrefix(string(cInfo.full.HostConfig.NetworkMode), "service:"+depName)
			if !hasDep && !sharesNetNS {
				continue
			}

			if sharesNetNS {
				// network-mode dependents are hard-bound to the target: always
				// cascade, regardless of stack (scope like compose's project)
				seen[cName] = true
				if !cInfo.full.State.Running {
					stoppedDependents = append(stoppedDependents, cName)
				}
				collect(cName[1:])
				continue
			}

			// depends_on: only cascade same-stack dependents with restart:true
			if !depEntry.Restart {
				continue
			}
			if startedStack != "" && ContainerStack(cInfo.full.Config) != startedStack {
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

// DependsOnFieldFromLabels converts the persisted compose depends_on label back
// into the depends_on *field* shape (map[service]DependsOnCont) that Cosmos's
// compose model exposes. This is what the user should see/edit: the field, not
// the label. Mirrors compose.go's projectFromName reconstruction.
func DependsOnFieldFromLabels(conf *conttype.Config) map[string]ContainerCreateRequestContainerDependsOnCont {
	out := map[string]ContainerCreateRequestContainerDependsOnCont{}
	if conf == nil || conf.Labels == nil {
		return out
	}
	for dep, entry := range DependsOnFromLabels(conf) {
		out[dep] = ContainerCreateRequestContainerDependsOnCont{
			Condition: entry.Condition,
			Restart:   strconv.FormatBool(entry.Restart),
		}
	}
	return out
}

// stripInternalDependsOnLabel removes compose's internal
// com.docker.compose.depends_on label from a labels map without mutating the
// input. The label is an internal serialization detail: the source of truth
// exposed to the user is the depends_on field.
func stripInternalDependsOnLabel(labels map[string]string) map[string]string {
	if labels == nil {
		return nil
	}
	if _, ok := labels[composeDependenciesLabel]; !ok {
		return labels
	}
	out := make(map[string]string, len(labels)-1)
	for k, v := range labels {
		if k == composeDependenciesLabel {
			continue
		}
		out[k] = v
	}
	return out
}

// ---------------------------------------------------------------------------
// Stack membership + restart-of-dependents
//
// docker-compose (pkg/compose/restart.go prepareRestartProject) restart
// behavior, which we mirror:
//   - depends_on entries with restart:true make the dependent restart together
//     with the dependency (entries with restart:false are pruned from the
//     graph before restarting)
//   - network_mode / ipc / pid: service:X dependents MUST be re-created when
//     the target restarts (their shared namespace dies) — compose handles this
//     as an implicit edge, not via restart:true
//   - everything runs in dependency order, dependencies first
//
// We additionally scope the cascade to containers of the SAME stack, so a
// single restarted container does not rebuild unrelated containers that happen
// to reference it.
// ---------------------------------------------------------------------------

// stackLabelCandidates are the label keys (in priority order) that Cosmos and
// docker-compose use to mark a container's stack/project membership.
var stackLabelCandidates = []string{
	"cosmos-stack",      // dash variant (referenced by user + backups.go)
	"cosmos.stack",      // dot variant (what the compose editor writes)
	"cosmos.stack.main", // main marker
	"cosmos-stack-main",
	"com.docker.compose.project", // real docker-compose
}

// ContainerStack returns the stack/project name a container belongs to, or ""
// if it is standalone. It accepts every variant Cosmos and compose use.
func ContainerStack(conf *conttype.Config) string {
	if conf == nil || conf.Labels == nil {
		return ""
	}
	// main markers don't carry a name; look for the actual stack/project name
	for _, key := range []string{"cosmos-stack", "cosmos.stack", "com.docker.compose.project"} {
		if v := conf.Labels[key]; v != "" {
			return v
		}
	}
	return ""
}

// containerStackMatches reports whether two container configs belong to the
// same stack. A standalone container (no stack) matches nothing but itself.
func containerStackMatches(a, b *conttype.Config) bool {
	sa, sb := ContainerStack(a), ContainerStack(b)
	if sa == "" || sb == "" {
		return false
	}
	return sa == sb
}

// restartDependentsNeeded decides whether a dependent must be (re)started when
// its dependency is restarted, mirroring docker-compose:
//   - network_mode/ipc/pid namespace sharing => always restart
//   - depends_on with restart:true            => restart
//   - depends_on with restart:false           => no restart
func restartDependentsNeeded(depEntry dependsOnEntry, networkMode string) bool {
	if NetworkModeContainerRef(networkMode) || NetworkModeServiceRef(networkMode) {
		return true
	}
	return depEntry.Restart
}

// RestartStackDependents restarts (or re-creates) every same-stack container
// that depends on containerName, either through depends_on (restart:true) or
// by sharing its network namespace (network_mode: container:/service:), in
// dependency order. It mirrors docker-compose restart semantics and is scoped
// to the stack to avoid surprising restarts of unrelated containers.
//
// Returns the names of containers it restarted (for logging/UI).
func RestartStackDependents(containerName string) []string {
	restarted := []string{}

	target, err := DockerClient.ContainerInspect(DockerContext, containerName)
	if err != nil {
		utils.Error("RestartStackDependents: cannot inspect "+containerName, err)
		return restarted
	}
	targetStack := ContainerStack(target.Config)
	if targetStack == "" {
		// no stack to scope against: only restart direct network_mode dependents
		// anywhere (namespace sharing is independent of stack), but be safe and
		// require the same stack like compose's project scope.
		utils.Debug("RestartStackDependents: " + containerName + " is not part of a stack; nothing to cascade")
		return restarted
	}

	containers, err := ListContainers()
	if err != nil {
		utils.Error("RestartStackDependents", err)
		return restarted
	}

	// Build same-stack container set.
	byName := map[string]depContainerInfo{}
	for _, c := range containers {
		full, err := DockerClient.ContainerInspect(DockerContext, c.ID)
		if err != nil {
			continue
		}
		if full.ID == target.ID {
			continue
		}
		if ContainerStack(full.Config) != targetStack {
			continue // scope to same stack
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

	// Collect dependents.
	dependentNames := []string{}
	for name, cInfo := range byName {
		full := cInfo.full
		deps := DependsOnFromLabels(full.Config)
		if entry, ok := deps[target.Name[1:]]; ok {
			if restartDependentsNeeded(entry, string(full.HostConfig.NetworkMode)) {
				utils.Log(fmt.Sprintf("RestartStackDependents: %s depends_on %s (restart=%t) -> restarting", name, target.Name[1:], entry.Restart))
				dependentNames = append(dependentNames, name)
			}
			continue
		}
		// network_mode / ipc / pid namespace sharing (implicit edge)
		nm := string(full.HostConfig.NetworkMode)
		if NetworkModeContainerRef(nm) || NetworkModeServiceRef(nm) {
			refTarget := NetworkModeRefTarget(nm)
			if refTarget == target.Name[1:] || refTarget == target.ID {
				utils.Log(fmt.Sprintf("RestartStackDependents: %s shares network namespace with %s -> restarting", name, target.Name[1:]))
				dependentNames = append(dependentNames, name)
			}
			continue
		}
		// cosmos-force-network-mode label (durable network_mode reference)
		labelMode := GetLabel(full, "cosmos-force-network-mode")
		labelTarget := NetworkModeRefTarget(labelMode)
		if labelTarget != "" && (labelTarget == target.Name[1:] || labelTarget == target.ID) {
			utils.Log(fmt.Sprintf("RestartStackDependents: %s shares network namespace with %s (label) -> restarting", name, target.Name[1:]))
			dependentNames = append(dependentNames, name)
		}
	}

	if len(dependentNames) == 0 {
		return restarted
	}

	// Restart in dependency order (dependencies first), respecting each
	// dependent's own depends_on waits.
	ordered := OrderByDependencies(dependentNames, depContainerConfigs(byName))
	for _, name := range ordered {
		full := byName[name].full
		utils.Log(fmt.Sprintf("RestartStackDependents: restarting %s", name))
		if errW := WaitForDependsOn(DockerContext, full.ID); errW != nil {
			utils.Error("RestartStackDependents: dependency wait failed for "+name, errW)
			continue
		}
		if errR := DockerClient.ContainerRestart(DockerContext, full.ID, conttype.StopOptions{}); errR != nil {
			utils.Error("RestartStackDependents: cannot restart "+name, errR)
			continue
		}
		restarted = append(restarted, name)
	}
	return restarted
}
