package docker

import (
	"testing"

	doctype "github.com/docker/docker/api/types"
	conttype "github.com/docker/docker/api/types/container"
)

func TestSetAndReadDependsOnLabels(t *testing.T) {
	conf := &conttype.Config{}
	deps := map[string]string{
		"db":    "service_healthy",
		"redis": "service_started",
	}
	SetDependsOnLabels(conf, deps)

	got := DependsOnFromLabels(conf)
	if len(got) != 2 {
		t.Fatalf("expected 2 deps, got %d: %v", len(got), got)
	}
	if got["db"] != "service_healthy" {
		t.Errorf("db condition = %q, want service_healthy", got["db"])
	}
	if got["redis"] != "service_started" {
		t.Errorf("redis condition = %q, want service_started", got["redis"])
	}
}

func TestDependsOnLabelsRoundTripWithEmptyCondition(t *testing.T) {
	conf := &conttype.Config{}
	SetDependsOnLabels(conf, map[string]string{
		"db": "", // compose shorthand (no condition) => service_started
	})
	got := DependsOnFromLabels(conf)
	if got["db"] != "service_started" {
		t.Errorf("empty condition should normalize to service_started, got %q", got["db"])
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

	// healthcheck-less running container counts as healthy (compose parity)
	if !depConditionMet(running, DepConditionHealthy) {
		t.Error("running healthcheck-less container should satisfy service_healthy")
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
		m := map[string]string{}
		for _, d := range deps {
			m[d] = "service_started"
		}
		SetDependsOnLabels(conf, m)
	}
	return doctype.ContainerJSON{Config: conf}
}
