package docker

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	conttype "github.com/docker/docker/api/types/container"
)

// harness

type lazyEventRecord struct {
	id     string
	level  string
	object string
	data   map[string]interface{}
}

type lazyHarness struct {
	t *testing.T

	mu sync.Mutex

	inspects int
	starts   int
	stops    int
	removes  int
	lists    int
	ips      int
	dials    int

	events []lazyEventRecord
}

// newLazyHarness stubs every docker seam and restores the originals on cleanup.
func newLazyHarness(t *testing.T) *lazyHarness {
	t.Helper()

	h := &lazyHarness{t: t}

	origInspect := lazyInspect
	origStart := lazyStartContainer
	origStop := lazyStopContainer
	origRemove := lazyRemoveContainer
	origList := lazyListContainers
	origIP := lazyContainerIP
	origDial := lazyDial
	origNow := lazyNow
	origSleep := lazySleep
	origTrigger := lazyTriggerEvent
	origPoll := lazyPollInterval
	origDialTimeout := lazyDialTimeout

	lazyInspect = func(string) (types.ContainerJSON, error) {
		h.bump(&h.inspects)
		t.Fatalf("unexpected lazyInspect call")
		return types.ContainerJSON{}, nil
	}
	lazyStartContainer = func(string) error {
		h.bump(&h.starts)
		t.Fatalf("unexpected lazyStartContainer call")
		return nil
	}
	lazyStopContainer = func(string) error {
		h.bump(&h.stops)
		t.Fatalf("unexpected lazyStopContainer call")
		return nil
	}
	lazyRemoveContainer = func(string) error {
		h.bump(&h.removes)
		t.Fatalf("unexpected lazyRemoveContainer call")
		return nil
	}
	lazyListContainers = func() ([]types.Container, error) {
		h.bump(&h.lists)
		t.Fatalf("unexpected lazyListContainers call")
		return nil, nil
	}
	lazyContainerIP = func(string) (string, error) {
		h.bump(&h.ips)
		t.Fatalf("unexpected lazyContainerIP call")
		return "", nil
	}
	lazyDial = func(string, time.Duration) error {
		h.bump(&h.dials)
		t.Fatalf("unexpected lazyDial call")
		return nil
	}
	lazyTriggerEvent = func(id, label, level, object string, data map[string]interface{}) {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.events = append(h.events, lazyEventRecord{id: id, level: level, object: object, data: data})
	}

	lazyPollInterval = 2 * time.Millisecond
	lazyDialTimeout = 20 * time.Millisecond

	lazyMu.Lock()
	lazyStates = map[string]*lazyEntry{}
	lazyMu.Unlock()

	t.Cleanup(func() {
		lazyInspect = origInspect
		lazyStartContainer = origStart
		lazyStopContainer = origStop
		lazyRemoveContainer = origRemove
		lazyListContainers = origList
		lazyContainerIP = origIP
		lazyDial = origDial
		lazyNow = origNow
		lazySleep = origSleep
		lazyTriggerEvent = origTrigger
		lazyPollInterval = origPoll
		lazyDialTimeout = origDialTimeout

		lazyMu.Lock()
		lazyStates = map[string]*lazyEntry{}
		lazyMu.Unlock()
	})

	return h
}

func (h *lazyHarness) bump(counter *int) {
	h.mu.Lock()
	*counter++
	h.mu.Unlock()
}

func (h *lazyHarness) counts() (inspects, starts, stops, removes, dials int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.inspects, h.starts, h.stops, h.removes, h.dials
}

func (h *lazyHarness) eventsOf(id string) []lazyEventRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := []lazyEventRecord{}
	for _, e := range h.events {
		if e.id == id {
			out = append(out, e)
		}
	}
	return out
}

func (h *lazyHarness) onInspect(fn func(name string, n int) (types.ContainerJSON, error)) {
	lazyInspect = func(name string) (types.ContainerJSON, error) {
		h.mu.Lock()
		h.inspects++
		n := h.inspects
		h.mu.Unlock()
		return fn(name, n)
	}
}

func (h *lazyHarness) onStart(fn func(name string) error) {
	lazyStartContainer = func(name string) error {
		h.bump(&h.starts)
		if fn == nil {
			return nil
		}
		return fn(name)
	}
}

func (h *lazyHarness) onStop(fn func(name string) error) {
	lazyStopContainer = func(name string) error {
		h.bump(&h.stops)
		if fn == nil {
			return nil
		}
		return fn(name)
	}
}

func (h *lazyHarness) onRemove(fn func(name string) error) {
	lazyRemoveContainer = func(name string) error {
		h.bump(&h.removes)
		if fn == nil {
			return nil
		}
		return fn(name)
	}
}

func (h *lazyHarness) onList(fn func() ([]types.Container, error)) {
	lazyListContainers = func() ([]types.Container, error) {
		h.bump(&h.lists)
		return fn()
	}
}

func (h *lazyHarness) onIP(ip string) {
	lazyContainerIP = func(string) (string, error) {
		h.bump(&h.ips)
		return ip, nil
	}
}

func (h *lazyHarness) onDial(fn func(addr string, n int) error) {
	lazyDial = func(addr string, _ time.Duration) error {
		h.mu.Lock()
		h.dials++
		n := h.dials
		h.mu.Unlock()
		return fn(addr, n)
	}
}

func seedLazy(name string, mutate func(*lazyEntry)) *lazyEntry {
	lazyMu.Lock()
	defer lazyMu.Unlock()
	st := &lazyEntry{
		name:         name,
		lazy:         true,
		idle:         LazyDefaultIdle,
		startTimeout: LazyDefaultStartTimeout,
		lastActivity: time.Now(),
	}
	if mutate != nil {
		mutate(st)
	}
	lazyStates[name] = st
	return st
}

func getLazy(name string) lazyEntry {
	lazyMu.Lock()
	defer lazyMu.Unlock()
	st := lazyStates[name]
	if st == nil {
		return lazyEntry{}
	}
	return *st
}

func inspectJSON(running bool, labels map[string]string, health *types.Health, startedAt string) types.ContainerJSON {
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running:   running,
				StartedAt: startedAt,
				Health:    health,
			},
		},
		Config: &conttype.Config{Labels: labels},
	}
}

func lazyLabels(extra map[string]string) map[string]string {
	l := map[string]string{LazyLabel: "true"}
	for k, v := range extra {
		l[k] = v
	}
	return l
}

// labels

func TestIsLazyLabels(t *testing.T) {
	if IsLazyLabels(nil) {
		t.Fatal("nil labels must not be lazy")
	}
	if IsLazyLabels(map[string]string{}) {
		t.Fatal("empty labels must not be lazy")
	}
	if IsLazyLabels(map[string]string{LazyLabel: "1"}) {
		t.Fatal("only \"true\" marks a container lazy")
	}
	if !IsLazyLabels(map[string]string{LazyLabel: "true"}) {
		t.Fatal("cosmos-lazy=true must be lazy")
	}
}

func TestLazyDurationLabelFallsBackOnGarbage(t *testing.T) {
	newLazyHarness(t)

	if got := lazyDurationLabel(map[string]string{}, LazyIdleLabel, LazyDefaultIdle, "c"); got != LazyDefaultIdle {
		t.Fatalf("missing label: got %v", got)
	}
	if got := lazyDurationLabel(map[string]string{LazyIdleLabel: "nonsense"}, LazyIdleLabel, LazyDefaultIdle, "c"); got != LazyDefaultIdle {
		t.Fatalf("garbage label: got %v", got)
	}
	if got := lazyDurationLabel(map[string]string{LazyIdleLabel: "-5m"}, LazyIdleLabel, LazyDefaultIdle, "c"); got != LazyDefaultIdle {
		t.Fatalf("negative label: got %v", got)
	}
	if got := lazyDurationLabel(map[string]string{LazyIdleLabel: "90s"}, LazyIdleLabel, LazyDefaultIdle, "c"); got != 90*time.Second {
		t.Fatalf("valid label: got %v", got)
	}
}

// fast path

func TestEnsureLazyAwakeFastPathTouchesNoDocker(t *testing.T) {
	h := newLazyHarness(t)

	if woke, err := EnsureLazyAwake("unknown", "80", time.Second); woke || err != nil {
		t.Fatalf("unknown: got (%v, %v)", woke, err)
	}

	seedLazy("plain", func(st *lazyEntry) { st.lazy = false })
	if woke, err := EnsureLazyAwake("plain", "80", time.Second); woke || err != nil {
		t.Fatalf("non-lazy: got (%v, %v)", woke, err)
	}

	seedLazy("awake", func(st *lazyEntry) { st.running = true })
	if woke, err := EnsureLazyAwake("awake", "80", time.Second); woke || err != nil {
		t.Fatalf("running: got (%v, %v)", woke, err)
	}

	inspects, starts, stops, removes, dials := h.counts()
	if inspects|starts|stops|removes|dials != 0 {
		t.Fatalf("fast path hit docker: inspects=%d starts=%d stops=%d removes=%d dials=%d",
			inspects, starts, stops, removes, dials)
	}
}

// single-flight

func TestEnsureLazyAwakeSingleFlight(t *testing.T) {
	h := newLazyHarness(t)

	var started atomic.Bool
	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(started.Load(), lazyLabels(nil), nil, ""), nil
	})
	h.onStart(func(string) error {
		// Slow enough that all 20 callers pile onto the same single-flight.
		time.Sleep(20 * time.Millisecond)
		started.Store(true)
		return nil
	})
	h.onIP("10.0.0.5")
	h.onDial(func(string, int) error { return nil })

	seedLazy("burst", nil)

	const callers = 20
	var wg sync.WaitGroup
	results := make([]bool, callers)
	errs := make([]error, callers)

	release := make(chan struct{})
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-release
			results[i], errs[i] = EnsureLazyAwake("burst", "8080", 5*time.Second)
		}(i)
	}
	close(release)
	wg.Wait()

	_, starts, _, _, _ := h.counts()
	if starts != 1 {
		t.Fatalf("expected exactly 1 ContainerStart for a burst of %d callers, got %d", callers, starts)
	}
	for i := 0; i < callers; i++ {
		if errs[i] != nil {
			t.Fatalf("caller %d: %v", i, errs[i])
		}
		if !results[i] {
			t.Fatalf("caller %d: expected woke=true", i)
		}
	}

	if st := getLazy("burst"); !st.running {
		t.Fatal("container should be marked running after a successful wake")
	}

	if len(h.eventsOf("cosmos.container.lazy.started")) != 1 {
		t.Fatalf("expected one lazy.started event, got %d", len(h.eventsOf("cosmos.container.lazy.started")))
	}
	if ev := h.eventsOf("cosmos.container.lazy.started"); ev[0].level != "debug" || ev[0].object != "container@burst" {
		t.Fatalf("bad started event: %+v", ev[0])
	}
}

// readiness

func TestReadinessUsesHealthcheckWhenPresent(t *testing.T) {
	h := newLazyHarness(t)

	var started atomic.Bool
	var polls atomic.Int32
	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		if !started.Load() {
			return inspectJSON(false, lazyLabels(nil), nil, ""), nil
		}
		n := polls.Add(1)
		status := types.Starting
		if n >= 4 {
			status = types.Healthy
		}
		return inspectJSON(true, lazyLabels(nil), &types.Health{Status: status}, ""), nil
	})
	h.onStart(func(string) error { started.Store(true); return nil })

	seedLazy("healthy-app", nil)

	woke, err := EnsureLazyAwake("healthy-app", "8080", 5*time.Second)
	if err != nil || !woke {
		t.Fatalf("got (%v, %v)", woke, err)
	}
	if polls.Load() < 4 {
		t.Fatalf("expected to poll health until healthy, polls=%d", polls.Load())
	}

	// A healthcheck is authoritative: the TCP probe must never be used.
	_, _, _, _, dials := h.counts()
	if dials != 0 {
		t.Fatalf("healthcheck path must not TCP-dial, got %d dials", dials)
	}
}

func TestReadinessUsesTCPWhenNoHealthcheck(t *testing.T) {
	h := newLazyHarness(t)

	var started atomic.Bool
	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(started.Load(), lazyLabels(nil), nil, ""), nil
	})
	h.onStart(func(string) error { started.Store(true); return nil })
	h.onIP("172.17.0.9")

	gotAddr := make(chan string, 8)
	h.onDial(func(addr string, n int) error {
		select {
		case gotAddr <- addr:
		default:
		}
		if n < 3 {
			return errors.New("connection refused")
		}
		return nil
	})

	seedLazy("tcp-app", nil)

	woke, err := EnsureLazyAwake("tcp-app", "9000", 5*time.Second)
	if err != nil || !woke {
		t.Fatalf("got (%v, %v)", woke, err)
	}
	if addr := <-gotAddr; addr != "172.17.0.9:9000" {
		t.Fatalf("dialled %q, want 172.17.0.9:9000", addr)
	}
	if _, _, _, _, dials := h.counts(); dials != 3 {
		t.Fatalf("expected 3 dials (2 refused + 1 accepted), got %d", dials)
	}
}

// ErrLazyStarting

func TestEnsureLazyAwakeReturnsErrLazyStartingAndFinishesInBackground(t *testing.T) {
	h := newLazyHarness(t)

	readyAt := time.Now().Add(150 * time.Millisecond)
	var started atomic.Bool

	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(started.Load(), lazyLabels(nil), nil, ""), nil
	})
	h.onStart(func(string) error { started.Store(true); return nil })
	h.onIP("10.1.2.3")
	h.onDial(func(string, int) error {
		if time.Now().Before(readyAt) {
			return errors.New("connection refused")
		}
		return nil
	})

	seedLazy("slowboot", func(st *lazyEntry) { st.startTimeout = 5 * time.Second })

	woke, err := EnsureLazyAwake("slowboot", "8080", 10*time.Millisecond)
	if !errors.Is(err, ErrLazyStarting) {
		t.Fatalf("expected ErrLazyStarting, got (%v, %v)", woke, err)
	}
	if woke {
		t.Fatal("ErrLazyStarting must report woke=false")
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		if getLazy("slowboot").running {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background wake never completed")
		}
		time.Sleep(5 * time.Millisecond)
	}

	if _, starts, _, _, _ := h.counts(); starts != 1 {
		t.Fatalf("expected 1 start, got %d", starts)
	}

	if woke, err := EnsureLazyAwake("slowboot", "8080", time.Second); woke || err != nil {
		t.Fatalf("after wake: got (%v, %v)", woke, err)
	}
}

// failure streak

func TestWakeFailureStreakRemovesDeploymentContainer(t *testing.T) {
	h := newLazyHarness(t)

	labels := lazyLabels(map[string]string{DeploymentLabel: "media"})
	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(false, labels, nil, ""), nil
	})
	h.onStart(func(string) error { return errors.New("no such image") })
	h.onRemove(func(string) error { return nil })

	seedLazy("broken", nil)

	for i := 1; i <= LazyMaxWakeFailures; i++ {
		woke, err := EnsureLazyAwake("broken", "80", 5*time.Second)
		if woke || err == nil {
			t.Fatalf("attempt %d: expected failure, got (%v, %v)", i, woke, err)
		}

		_, _, _, removes, _ := h.counts()
		st := getLazy("broken")
		if i < LazyMaxWakeFailures {
			if removes != 0 {
				t.Fatalf("attempt %d: removed too early", i)
			}
			if st.failStreak != i {
				t.Fatalf("attempt %d: failStreak=%d", i, st.failStreak)
			}
		} else {
			if removes != 1 {
				t.Fatalf("attempt %d: expected exactly one remove, got %d", i, removes)
			}
			if st.failStreak != 0 {
				t.Fatalf("streak must reset after giving up, got %d", st.failStreak)
			}
		}
	}

	failed := h.eventsOf("cosmos.container.lazy.failed")
	if len(failed) != 1 {
		t.Fatalf("expected 1 lazy.failed event, got %d", len(failed))
	}
	if failed[0].level != "warning" {
		t.Fatalf("lazy.failed must be warning, got %q", failed[0].level)
	}
	if failed[0].object != "container@broken" {
		t.Fatalf("bad object %q", failed[0].object)
	}
	if failed[0].data["streak"] != LazyMaxWakeFailures {
		t.Fatalf("streak in data = %v", failed[0].data["streak"])
	}
	if reason, _ := failed[0].data["reason"].(string); reason == "" {
		t.Fatal("reason missing from lazy.failed data")
	}
}

func TestWakeFailureStreakLeavesNonDeploymentContainerStopped(t *testing.T) {
	h := newLazyHarness(t)

	labels := lazyLabels(nil)
	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(false, labels, nil, ""), nil
	})
	h.onStart(func(string) error { return errors.New("boom") })
	// lazyRemoveContainer is left as the "must not be called" stub.

	seedLazy("standalone", nil)

	for i := 0; i < LazyMaxWakeFailures; i++ {
		if _, err := EnsureLazyAwake("standalone", "80", 5*time.Second); err == nil {
			t.Fatalf("attempt %d: expected error", i)
		}
	}

	if _, _, _, removes, _ := h.counts(); removes != 0 {
		t.Fatalf("a container without %s must never be removed, removes=%d", DeploymentLabel, removes)
	}
	if len(h.eventsOf("cosmos.container.lazy.failed")) != 1 {
		t.Fatalf("expected the failed event anyway, got %d", len(h.eventsOf("cosmos.container.lazy.failed")))
	}
	if st := getLazy("standalone"); st.failStreak != 0 || st.running {
		t.Fatalf("after giving up: streak=%d running=%v", st.failStreak, st.running)
	}
}

func TestWakeFailureStreakResetsOnSuccess(t *testing.T) {
	h := newLazyHarness(t)

	var started atomic.Bool
	var failStarts atomic.Bool
	failStarts.Store(true)

	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(started.Load(), lazyLabels(nil), nil, ""), nil
	})
	h.onStart(func(string) error {
		if failStarts.Load() {
			return errors.New("transient")
		}
		started.Store(true)
		return nil
	})
	h.onIP("10.0.0.1")
	h.onDial(func(string, int) error { return nil })

	seedLazy("flaky", nil)

	for i := 1; i < LazyMaxWakeFailures; i++ {
		if _, err := EnsureLazyAwake("flaky", "80", 5*time.Second); err == nil {
			t.Fatalf("attempt %d: expected error", i)
		}
	}
	if st := getLazy("flaky"); st.failStreak != LazyMaxWakeFailures-1 {
		t.Fatalf("streak before success = %d", st.failStreak)
	}

	failStarts.Store(false)
	if woke, err := EnsureLazyAwake("flaky", "80", 5*time.Second); !woke || err != nil {
		t.Fatalf("recovery: got (%v, %v)", woke, err)
	}
	if st := getLazy("flaky"); st.failStreak != 0 {
		t.Fatalf("a successful wake must reset the streak, got %d", st.failStreak)
	}
	if len(h.eventsOf("cosmos.container.lazy.failed")) != 0 {
		t.Fatal("must not give up before the streak reaches the maximum")
	}
}

// activity tracking

func TestLazyActivityTrackingIsNoOpForUnknownContainers(t *testing.T) {
	newLazyHarness(t)

	LazyTouch("ghost")
	LazyConnOpen("ghost")
	LazyConnClose("ghost")

	lazyMu.Lock()
	n := len(lazyStates)
	lazyMu.Unlock()
	if n != 0 {
		t.Fatalf("activity calls must not create entries, got %d", n)
	}
}

func TestLazyConnCountNeverGoesNegative(t *testing.T) {
	newLazyHarness(t)
	seedLazy("app", nil)

	LazyConnClose("app")
	LazyConnClose("app")
	if st := getLazy("app"); st.openConns != 0 {
		t.Fatalf("openConns=%d, want 0", st.openConns)
	}

	LazyConnOpen("app")
	LazyConnOpen("app")
	if st := getLazy("app"); st.openConns != 2 {
		t.Fatalf("openConns=%d, want 2", st.openConns)
	}
	LazyConnClose("app")
	if st := getLazy("app"); st.openConns != 1 {
		t.Fatalf("openConns=%d, want 1", st.openConns)
	}
}

func TestLazyIsDormant(t *testing.T) {
	newLazyHarness(t)

	seedLazy("asleep", nil)
	seedLazy("up", func(st *lazyEntry) { st.running = true })
	seedLazy("notlazy", func(st *lazyEntry) { st.lazy = false })

	if !LazyIsDormant("asleep") {
		t.Fatal("stopped lazy container must be dormant")
	}
	if LazyIsDormant("up") {
		t.Fatal("running lazy container must not be dormant")
	}
	if LazyIsDormant("notlazy") {
		t.Fatal("non-lazy container must never be dormant")
	}
	if LazyIsDormant("missing") {
		t.Fatal("unknown container must never be dormant")
	}
}

// reaper

func TestReaperStopsOnlyIdleUnusedContainers(t *testing.T) {
	h := newLazyHarness(t)

	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	lazyNow = func() time.Time { return now }

	stopped := map[string]bool{}
	var stopMu sync.Mutex
	h.onStop(func(name string) error {
		stopMu.Lock()
		stopped[name] = true
		stopMu.Unlock()
		return nil
	})

	seedLazy("idle", func(st *lazyEntry) {
		st.running = true
		st.idle = 10 * time.Minute
		st.lastActivity = now.Add(-30 * time.Minute)
	})
	seedLazy("busy", func(st *lazyEntry) {
		st.running = true
		st.idle = 10 * time.Minute
		st.lastActivity = now.Add(-30 * time.Minute)
		st.openConns = 1 // a hijacked websocket holds it awake
	})
	seedLazy("recent", func(st *lazyEntry) {
		st.running = true
		st.idle = 10 * time.Minute
		st.lastActivity = now.Add(-1 * time.Minute)
	})
	seedLazy("alreadydown", func(st *lazyEntry) {
		st.running = false
		st.idle = 10 * time.Minute
		st.lastActivity = now.Add(-30 * time.Minute)
	})
	seedLazy("notlazy", func(st *lazyEntry) {
		st.lazy = false
		st.running = true
		st.idle = 10 * time.Minute
		st.lastActivity = now.Add(-30 * time.Minute)
	})

	lazyReaperTick()

	stopMu.Lock()
	defer stopMu.Unlock()
	if len(stopped) != 1 || !stopped["idle"] {
		t.Fatalf("reaper stopped %v, want only [idle]", stopped)
	}

	st := getLazy("idle")
	if st.running {
		t.Fatal("reaped container must be marked not running")
	}
	if !st.stoppedByReaper {
		t.Fatal("reaped container must carry stoppedByReaper")
	}

	ev := h.eventsOf("cosmos.container.lazy.stopped")
	if len(ev) != 1 || ev[0].level != "debug" || ev[0].object != "container@idle" {
		t.Fatalf("bad lazy.stopped events: %+v", ev)
	}
}

func TestReaperClearsFlagWhenStopFails(t *testing.T) {
	h := newLazyHarness(t)

	now := time.Now()
	lazyNow = func() time.Time { return now }
	h.onStop(func(string) error { return errors.New("docker is unhappy") })

	seedLazy("stubborn", func(st *lazyEntry) {
		st.running = true
		st.idle = time.Minute
		st.lastActivity = now.Add(-time.Hour)
	})

	lazyReaperTick()

	st := getLazy("stubborn")
	if st.stoppedByReaper {
		t.Fatal("a failed stop must not leave stoppedByReaper set, or a later crash would be logged as routine")
	}
	if !st.running {
		t.Fatal("a failed stop must leave the container marked running")
	}
}

func TestRescanSeedsLastActivityFromStartedAt(t *testing.T) {
	h := newLazyHarness(t)

	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	lazyNow = func() time.Time { return now }

	oldStart := now.Add(-4 * time.Hour).Format(time.RFC3339Nano)
	freshStart := now.Add(-1 * time.Minute).Format(time.RFC3339Nano)

	h.onList(func() ([]types.Container, error) {
		return []types.Container{
			{ID: "id-old", Names: []string{"/veteran"}, State: "running", Labels: lazyLabels(map[string]string{LazyIdleLabel: "30m"})},
			{ID: "id-new", Names: []string{"/newborn"}, State: "running", Labels: lazyLabels(map[string]string{LazyIdleLabel: "30m"})},
			{ID: "id-off", Names: []string{"/dormant"}, State: "exited", Labels: lazyLabels(nil)},
			{ID: "id-plain", Names: []string{"/plain"}, State: "running", Labels: map[string]string{}},
		}, nil
	})
	h.onInspect(func(name string, _ int) (types.ContainerJSON, error) {
		switch name {
		case "id-old":
			return inspectJSON(true, lazyLabels(map[string]string{LazyIdleLabel: "30m"}), nil, oldStart), nil
		case "id-new":
			return inspectJSON(true, lazyLabels(map[string]string{LazyIdleLabel: "30m"}), nil, freshStart), nil
		case "id-off":
			return inspectJSON(false, lazyLabels(nil), nil, ""), nil
		}
		return types.ContainerJSON{}, errors.New("unexpected inspect of " + name)
	})

	lazyRescan()

	lazyMu.Lock()
	_, hasPlain := lazyStates["plain"]
	lazyMu.Unlock()
	if hasPlain {
		t.Fatal("non-lazy containers must not be tracked")
	}

	if st := getLazy("veteran"); !st.lastActivity.Equal(now.Add(-4 * time.Hour)) {
		t.Fatalf("veteran lastActivity=%v, want StartedAt", st.lastActivity)
	}
	if st := getLazy("newborn"); !st.lastActivity.Equal(now.Add(-1 * time.Minute)) {
		t.Fatalf("newborn lastActivity=%v, want StartedAt", st.lastActivity)
	}
	if st := getLazy("veteran"); st.idle != 30*time.Minute {
		t.Fatalf("idle label not applied, got %v", st.idle)
	}
	if st := getLazy("dormant"); st.running {
		t.Fatal("exited container must be tracked as not running")
	}

	stopped := map[string]bool{}
	var stopMu sync.Mutex
	h.onStop(func(name string) error {
		stopMu.Lock()
		stopped[name] = true
		stopMu.Unlock()
		return nil
	})

	lazyReaperTick()

	stopMu.Lock()
	defer stopMu.Unlock()
	if !stopped["veteran"] {
		t.Fatal("a container idle since long before startup must be reaped on the first tick")
	}
	if stopped["newborn"] {
		t.Fatal("a freshly started container must never be reaped on the first tick")
	}
}

func TestStartStopLazyReaperIsIdempotent(t *testing.T) {
	h := newLazyHarness(t)

	h.onList(func() ([]types.Container, error) { return nil, nil })

	StartLazyReaper()
	StartLazyReaper()
	defer StopLazyReaper()

	h.mu.Lock()
	lists := h.lists
	h.mu.Unlock()
	if lists != 1 {
		t.Fatalf("StartLazyReaper must be idempotent, rescanned %d times", lists)
	}

	StopLazyReaper()
	StopLazyReaper() // must not panic
}

// docker event hook

func TestLazyEventHookTracksLifecycle(t *testing.T) {
	h := newLazyHarness(t)

	labels := lazyLabels(map[string]string{LazyIdleLabel: "5m", LazyStartTimeoutLabel: "90s"})
	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(false, labels, nil, ""), nil
	})

	suppress, level := lazyOnContainerEvent("create", "abc", "app", labels)
	if suppress || level != "" {
		t.Fatalf("create: got (%v, %q)", suppress, level)
	}
	st := getLazy("app")
	if !st.lazy || st.running {
		t.Fatalf("after create: %+v", st)
	}
	if st.idle != 5*time.Minute || st.startTimeout != 90*time.Second {
		t.Fatalf("labels not applied: idle=%v startTimeout=%v", st.idle, st.startTimeout)
	}

	suppress, level = lazyOnContainerEvent("start", "abc", "app", labels)
	if !suppress {
		t.Fatal("start of a lazy container must suppress bootstrap/export")
	}
	if level != "" {
		t.Fatalf("start level override = %q", level)
	}
	if st = getLazy("app"); !st.running || st.lastActivity.IsZero() {
		t.Fatalf("after start: %+v", st)
	}

	lazyOnContainerEvent("destroy", "abc", "app", labels)
	lazyMu.Lock()
	_, still := lazyStates["app"]
	lazyMu.Unlock()
	if still {
		t.Fatal("destroy must drop the entry")
	}
}

func TestLazyEventHookDoesNotSuppressNonLazyStarts(t *testing.T) {
	h := newLazyHarness(t)

	plain := map[string]string{}
	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(true, plain, nil, ""), nil
	})

	suppress, level := lazyOnContainerEvent("start", "xyz", "regular", plain)
	if suppress {
		t.Fatal("a non-lazy start must keep its bootstrap and export")
	}
	if level != "" {
		t.Fatalf("level override = %q", level)
	}
}

func TestLazyEventHookDowngradesDieOnlyWhenReaped(t *testing.T) {
	newLazyHarness(t)

	seedLazy("reaped", func(st *lazyEntry) {
		st.running = true
		st.stoppedByReaper = true
	})
	suppress, level := lazyOnContainerEvent("die", "id1", "reaped", lazyLabels(nil))
	if suppress {
		t.Fatal("die must not suppress side effects")
	}
	if level != "debug" {
		t.Fatalf("reaper-initiated die level = %q, want debug", level)
	}
	st := getLazy("reaped")
	if st.running {
		t.Fatal("die must mark the container not running")
	}
	if st.stoppedByReaper {
		t.Fatal("the flag must be cleared after one die")
	}

	// A second die with no reaper stop in between is a crash.
	_, level = lazyOnContainerEvent("die", "id1", "reaped", lazyLabels(nil))
	if level != "" {
		t.Fatalf("crash die level override = %q, want none (warning)", level)
	}

	_, level = lazyOnContainerEvent("die", "id2", "someone-else", map[string]string{})
	if level != "" {
		t.Fatalf("non-lazy die level override = %q", level)
	}
}

// one `docker stop` emits kill, die, stop in that order.
func TestLazyEventHookDowngradesDieAcrossRealStopSequence(t *testing.T) {
	newLazyHarness(t)

	seedLazy("reaped", func(st *lazyEntry) {
		st.running = true
		st.stoppedByReaper = true
	})

	if _, level := lazyOnContainerEvent("kill", "id1", "reaped", lazyLabels(nil)); level != "" {
		t.Fatalf("kill level override = %q, want none", level)
	}
	if !getLazy("reaped").stoppedByReaper {
		t.Fatal("kill must not consume the reaper flag")
	}
	if _, level := lazyOnContainerEvent("die", "id1", "reaped", lazyLabels(nil)); level != "debug" {
		t.Fatalf("die after reaper kill level = %q, want debug", level)
	}
	if _, level := lazyOnContainerEvent("stop", "id1", "reaped", lazyLabels(nil)); level != "" {
		t.Fatalf("stop level override = %q, want none", level)
	}
	if getLazy("reaped").stoppedByReaper {
		t.Fatal("flag must be consumed by die")
	}
}

func TestLazyWakeForgetsCachedIP(t *testing.T) {
	h := newLazyHarness(t)

	var started atomic.Bool
	h.onInspect(func(_ string, _ int) (types.ContainerJSON, error) {
		return inspectJSON(started.Load(), lazyLabels(nil), nil, ""), nil
	})
	h.onStart(func(string) error { started.Store(true); return nil })
	h.onIP("10.0.0.9")
	h.onDial(func(string, int) error { return nil })
	seedLazy("app", nil)

	cache.Set("app", "10.0.0.1", time.Minute)
	if _, err := EnsureLazyAwake("app", "8080", 5*time.Second); err != nil {
		t.Fatalf("wake: %v", err)
	}
	if ip, found := cache.Get("app"); found {
		t.Fatalf("cached IP %q must be dropped on wake", ip)
	}
}
