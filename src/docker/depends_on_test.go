package docker

import (
	"testing"

	doctype "github.com/docker/docker/api/types"
	conttype "github.com/docker/docker/api/types/container"
)

func TestSetAndReadDependsOnComposeLabel(t *testing.T) {
	conf := &conttype.Config{}
	deps := map[string]dependsOnEntry{
		"db":    {Condition: "service_healthy", Restart: true},
		"redis": {Condition: "service_started", Restart: true},
	}
	SetDependsOnLabels(conf, deps)

	// Must be written in docker-compose's own label + format.
	if conf.Labels[composeDependenciesLabel] == "" {
		t.Fatal("compose depends_on label not written")
	}
	if _, ok := conf.Labels["cosmos.depends_on.db"]; ok {
		t.Fatal("should NOT write custom cosmos.depends_on.* labels anymore")
	}

	got := DependsOnFromLabels(conf)
	if len(got) != 2 {
		t.Fatalf("expected 2 deps, got %d: %v", len(got), got)
	}
	if got["db"].Condition != "service_healthy" {
		t.Errorf("db condition = %q, want service_healthy", got["db"].Condition)
	}
	if got["redis"].Condition != "service_started" {
		t.Errorf("redis condition = %q, want service_started", got["redis"].Condition)
	}
	if !got["db"].Restart {
		t.Error("db restart should be true")
	}
}

// A label written by real docker-compose (com.docker.compose.depends_on,
// "dep:condition:restart" format) must parse correctly.
func TestParseComposeLabel(t *testing.T) {
	conf := &conttype.Config{Labels: map[string]string{
		composeDependenciesLabel: "db:service_healthy:true,redis:service_started:false,worker:service_completed_successfully:true",
	}}
	got := DependsOnFromLabels(conf)
	if len(got) != 3 {
		t.Fatalf("expected 3 deps, got %d: %v", len(got), got)
	}
	if got["db"].Condition != "service_healthy" || !got["db"].Restart {
		t.Errorf("db = %+v", got["db"])
	}
	if got["redis"].Condition != "service_started" || got["redis"].Restart {
		t.Errorf("redis = %+v", got["redis"])
	}
	// legacy fragment without condition -> running_or_healthy (compose parity)
	conf2 := &conttype.Config{Labels: map[string]string{
		composeDependenciesLabel: "db:service_healthy:true,cache",
	}}
	got2 := DependsOnFromLabels(conf2)
	if got2["cache"].Condition != "running_or_healthy" {
		t.Errorf("legacy no-condition dep should default to running_or_healthy, got %q", got2["cache"].Condition)
	}
}

func TestDependsOnFromLabelsNoLabels(t *testing.T) {
	got := DependsOnFromLabels(&conttype.Config{})
	if len(got) != 0 {
		t.Fatalf("expected no deps for label-less config, got %v", got)
	}
	got = DependsOnFromLabels(nil)
	if len(got) != 0 {
		t.Fatalf("expected no deps for nil config, got %v", got)
	}
}

func newTestContainer(state *doctype.ContainerState, conf *conttype.Config) doctype.ContainerJSON {
	return doctype.ContainerJSON{
		ContainerJSONBase: &doctype.ContainerJSONBase{State: state},
		Config:            conf,
	}
}

func TestDepConditionMet(t *testing.T) {
	running := newTestContainer(&doctype.ContainerState{Running: true}, &conttype.Config{})
	stoppedOK := newTestContainer(&doctype.ContainerState{Running: false, ExitCode: 0}, &conttype.Config{})
	stoppedFail := newTestContainer(&doctype.ContainerState{Running: false, ExitCode: 1}, &conttype.Config{})
	noState := doctype.ContainerJSON{}

	if !depConditionMet(running, DepConditionStarted) {
		t.Error("running should satisfy service_started")
	}
	if depConditionMet(stoppedOK, DepConditionStarted) {
		t.Error("stopped should NOT satisfy service_started")
	}

	// healthcheck-less running container counts as healthy for
	// running_or_healthy (compose parity: fallback to running)
	if !depConditionMet(running, DepConditionRunningOrHealthy) {
		t.Error("running healthcheck-less container should satisfy running_or_healthy")
	}

	// service_healthy with NO healthcheck: compose errors / treats as unmet
	if depConditionMet(running, DepConditionHealthy) {
		t.Error("running healthcheck-less container should NOT satisfy service_healthy")
	}

	// running with a healthcheck but no health status: not healthy yet
	hc := newTestContainer(&doctype.ContainerState{Running: true}, &conttype.Config{Healthcheck: &conttype.HealthConfig{Test: []string{"CMD", "true"}}})
	if depConditionMet(hc, DepConditionHealthy) {
		t.Error("running with healthcheck but no status should NOT be healthy")
	}
	hcHealthy := newTestContainer(&doctype.ContainerState{Running: true, Health: &doctype.Health{Status: doctype.Healthy}}, &conttype.Config{Healthcheck: &conttype.HealthConfig{Test: []string{"CMD", "true"}}})
	if !depConditionMet(hcHealthy, DepConditionHealthy) {
		t.Error("healthy container should satisfy service_healthy")
	}

	if !depConditionMet(stoppedOK, DepConditionCompletedSuccess) {
		t.Error("exited-0 should satisfy service_completed_successfully")
	}
	if depConditionMet(running, DepConditionCompletedSuccess) {
		t.Error("running should NOT satisfy service_completed_successfully")
	}
	if depConditionMet(stoppedFail, DepConditionCompletedSuccess) {
		t.Error("exited-1 should NOT satisfy service_completed_successfully")
	}
	if depConditionMet(noState, DepConditionStarted) {
		t.Error("nil state should not satisfy any condition")
	}
}

func TestOrderByDependencies(t *testing.T) {
	byName := map[string]doctype.ContainerJSON{
		"/app":    containerWithDeps("db", "redis"),
		"/db":     containerWithDeps(""),
		"/redis":  containerWithDeps(""),
		"/worker": containerWithDeps("app"),
	}

	names := []string{"/app", "/db", "/worker", "/redis"}
	ordered := OrderByDependencies(names, byName)

	// dependencies must appear before dependents
	pos := map[string]int{}
	for i, n := range ordered {
		pos[n] = i
	}
	if pos["/db"] > pos["/app"] {
		t.Errorf("db should start before app; order=%v", ordered)
	}
	if pos["/redis"] > pos["/app"] {
		t.Errorf("redis should start before app; order=%v", ordered)
	}
	if pos["/app"] > pos["/worker"] {
		t.Errorf("app should start before worker; order=%v", ordered)
	}
	if len(ordered) != 4 {
		t.Errorf("expected 4 containers, got %d: %v", len(ordered), ordered)
	}
}

func TestOrderByDependenciesCycleFallback(t *testing.T) {
	byName := map[string]doctype.ContainerJSON{
		"/a": containerWithDeps("b"),
		"/b": containerWithDeps("a"),
	}
	names := []string{"/a", "/b"}
	ordered := OrderByDependencies(names, byName)
	// cycle => keep original order, no infinite loop
	if len(ordered) != 2 {
		t.Fatalf("expected 2 containers, got %d: %v", len(ordered), ordered)
	}
}

func containerWithDeps(deps ...string) doctype.ContainerJSON {
	conf := &conttype.Config{Labels: map[string]string{}}
	if len(deps) > 0 && deps[0] != "" {
		m := map[string]dependsOnEntry{}
		for _, d := range deps {
			m[d] = dependsOnEntry{Condition: "service_started", Restart: true}
		}
		SetDependsOnLabels(conf, m)
	}
	return doctype.ContainerJSON{Config: conf}
}

// Regression test for the pihole/tailscale failure: a service with
// network_mode: container:tailscale (external target, not in the batch) must
// NOT be treated as an unsatisfied dependency.
func TestReOrderServicesExternalNetworkMode(t *testing.T) {
	serviceMap := map[string]ContainerCreateRequestContainer{
		"pihole": {
			Name:        "pihole",
			NetworkMode: "container:tailscale",
		},
	}
	order, _, err := ReOrderServices(serviceMap)
	if err != nil {
		t.Fatalf("external network_mode target should not error, got: %v", err)
	}
	if len(order) != 1 || order[0].Name != "pihole" {
		t.Fatalf("expected pihole in start order, got %+v", order)
	}
}

// A service with network_mode referencing an in-batch service must be ordered
// after it (soft constraint), and should not fail.
func TestReOrderServicesInBatchNetworkMode(t *testing.T) {
	serviceMap := map[string]ContainerCreateRequestContainer{
		"app": {
			Name:        "app",
			NetworkMode: "container:db",
		},
		"db": {
			Name: "db",
		},
	}
	order, _, err := ReOrderServices(serviceMap)
	if err != nil {
		t.Fatalf("in-batch network_mode should not error, got: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("expected 2 in order, got %d: %+v", len(order), order)
	}
	// db must start before app
	pos := map[string]int{}
	for i, c := range order {
		pos[c.Name] = i
	}
	if pos["db"] > pos["app"] {
		t.Fatalf("db should start before app; order=%v", order)
	}
}

// Depends_on must still be respected as a hard ordering constraint.
func TestReOrderServicesDependsOn(t *testing.T) {
	serviceMap := map[string]ContainerCreateRequestContainer{
		"app": {
			Name: "app",
			DependsOn: map[string]ContainerCreateRequestContainerDependsOnCont{
				"db": {Condition: "service_started"},
			},
		},
		"db": {Name: "db"},
	}
	order, _, err := ReOrderServices(serviceMap)
	if err != nil {
		t.Fatalf("depends_on should not error, got: %v", err)
	}
	pos := map[string]int{}
	for i, c := range order {
		pos[c.Name] = i
	}
	if pos["db"] > pos["app"] {
		t.Fatalf("db should start before app via depends_on; order=%v", order)
	}
}

// Circular depends_on must still error.
func TestReOrderServicesCircular(t *testing.T) {
	serviceMap := map[string]ContainerCreateRequestContainer{
		"a": {
			Name: "a",
			DependsOn: map[string]ContainerCreateRequestContainerDependsOnCont{
				"b": {Condition: "service_started"},
			},
		},
		"b": {
			Name: "b",
			DependsOn: map[string]ContainerCreateRequestContainerDependsOnCont{
				"a": {Condition: "service_started"},
			},
		},
	}
	_, _, err := ReOrderServices(serviceMap)
	if err == nil {
		t.Fatal("circular depends_on should error")
	}
}

// The depends_on *field* must be reconstructed from the label, and the label
// itself hidden, so the source of truth the user sees/edits is the field.
func TestDependsOnFieldFromLabelsAndHide(t *testing.T) {
	conf := &conttype.Config{Labels: map[string]string{
		composeDependenciesLabel: "db:service_healthy:true,redis:service_started:false",
		"some.other.label":       "value",
	}}
	field := DependsOnFieldFromLabels(conf, nil)
	if len(field) != 2 {
		t.Fatalf("expected 2 field entries, got %d: %v", len(field), field)
	}
	if field["db"].Condition != "service_healthy" || field["db"].Restart != "true" {
		t.Errorf("db = %+v", field["db"])
	}
	if field["redis"].Condition != "service_started" || field["redis"].Restart != "false" {
		t.Errorf("redis = %+v", field["redis"])
	}

	// The internal label must be hidden; other labels kept.
	labels := stripInternalDependsOnLabel(conf.Labels)
	if _, ok := labels[composeDependenciesLabel]; ok {
		t.Fatal("com.docker.compose.depends_on label should be hidden")
	}
	if labels["some.other.label"] != "value" {
		t.Errorf("other labels should be preserved, got %v", labels)
	}

	// Original labels map must not be mutated.
	if _, ok := conf.Labels[composeDependenciesLabel]; !ok {
		t.Fatal("original labels map should not be mutated")
	}
}

// No label -> empty field, and strip is a no-op that returns the same map.
func TestDependsOnFieldNoLabel(t *testing.T) {
	conf := &conttype.Config{Labels: map[string]string{"a": "b"}}
	if len(DependsOnFieldFromLabels(conf, nil)) != 0 {
		t.Fatal("no label should mean no depends_on field")
	}
	labels := stripInternalDependsOnLabel(conf.Labels)
	if labels["a"] != "b" {
		t.Fatalf("labels should be unchanged: %v", labels)
	}
}

func TestContainerStackVariants(t *testing.T) {
	cases := []struct {
		labels map[string]string
		want   string
	}{
		{map[string]string{"cosmos-stack": "dashstack"}, "dashstack"},
		{map[string]string{"cosmos.stack": "mystack"}, "mystack"},
		{map[string]string{"com.docker.compose.project": "proj"}, "proj"},
		{map[string]string{"com.docker.compose.project": "proj", "cosmos.stack": "other"}, "other"}, // cosmos wins priority
		{map[string]string{"cosmos.stack.main": "true"}, ""},                                        // main marker alone -> no name
		{nil, ""},
	}
	for i, c := range cases {
		conf := &conttype.Config{Labels: c.labels}
		if got := ContainerStack(conf); got != c.want {
			t.Errorf("case %d: ContainerStack = %q, want %q", i, got, c.want)
		}
	}
}

func TestRestartDependentsNeeded(t *testing.T) {
	// depends_on restart:true -> restart
	if !restartDependentsNeeded(dependsOnEntry{Restart: true}, "") {
		t.Error("depends_on restart:true should restart dependent")
	}
	// depends_on restart:false -> no restart
	if restartDependentsNeeded(dependsOnEntry{Restart: false}, "") {
		t.Error("depends_on restart:false should NOT restart dependent")
	}
	// network_mode container: (even restart:false) -> restart (namespace shared)
	if !restartDependentsNeeded(dependsOnEntry{Restart: false}, "container:db") {
		t.Error("network_mode container: should restart dependent even with restart:false")
	}
	if !restartDependentsNeeded(dependsOnEntry{Restart: false}, "service:db") {
		t.Error("network_mode service: should restart dependent even with restart:false")
	}
	// plain depends_on no restart -> default false -> no restart
	if restartDependentsNeeded(dependsOnEntry{}, "bridge") {
		t.Error("default depends_on (restart false) should NOT restart")
	}
}

// Service-name <-> container-name resolution (docker-compose compatibility).
func TestResolveDepKey(t *testing.T) {
	mk := func(name, svc string) doctype.ContainerJSON {
		full := doctype.ContainerJSON{
			ContainerJSONBase: &doctype.ContainerJSONBase{Name: "/" + name},
			Config:            &conttype.Config{Labels: map[string]string{}},
		}
		if svc != "" {
			full.Config.Labels["com.docker.compose.service"] = svc
		}
		return full
	}

	idx := &containerNameIndex{byName: map[string]doctype.ContainerJSON{}, byService: map[string]doctype.ContainerJSON{}}
	idx.byName["my-db"] = mk("my-db", "db")
	idx.byName["my-web"] = mk("my-web", "web")
	idx.byService["db"] = idx.byName["my-db"]
	idx.byService["web"] = idx.byName["my-web"]

	// Cosmos-style key (container name) -> unchanged
	if got := resolveDepKey(idx, "my-db"); got != "my-db" {
		t.Errorf("container-name key: got %q want my-db", got)
	}
	// docker-compose-style key (service name) -> resolved to container name
	if got := resolveDepKey(idx, "db"); got != "my-db" {
		t.Errorf("service-name key db: got %q want my-db", got)
	}
	if got := resolveDepKey(idx, "web"); got != "my-web" {
		t.Errorf("service-name key web: got %q want my-web", got)
	}
	// unresolvable -> unchanged
	if got := resolveDepKey(idx, "missing"); got != "missing" {
		t.Errorf("missing key: got %q want missing", got)
	}
	// nil index -> unchanged
	if got := resolveDepKey(nil, "db"); got != "db" {
		t.Errorf("nil index: got %q want db", got)
	}
}

func TestDependsOnIncludesTarget(t *testing.T) {
	deps := map[string]dependsOnEntry{
		"db":    {Condition: "service_healthy", Restart: true},
		"my-db": {Condition: "service_started", Restart: true},
	}
	// container-name match
	if _, ok := DependsOnIncludesTarget(deps, "my-db", "db"); !ok {
		t.Error("should match by container name my-db")
	}
	// service-name match
	if entry, ok := DependsOnIncludesTarget(deps, "custom-db", "db"); !ok || entry.Condition != "service_healthy" {
		t.Errorf("should match by service name db, got %+v ok=%v", entry, ok)
	}
	// no match
	if _, ok := DependsOnIncludesTarget(deps, "web", "web"); ok {
		t.Error("should not match web")
	}
}
