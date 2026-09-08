package docker

// Lazy containers sleep when idle and are woken by the proxy; nothing here may depend on the pro package.

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"

	"github.com/docker/docker/api/types"
	conttype "github.com/docker/docker/api/types/container"
)

const (
	LazyLabel             = "cosmos-lazy"
	LazyIdleLabel         = "cosmos-lazy-idle"
	LazyStartTimeoutLabel = "cosmos-lazy-start-timeout"

	LazyDefaultIdle         = time.Hour
	LazyDefaultStartTimeout = 60 * time.Second
	LazyReaperInterval      = 30 * time.Second
	LazyMaxWakeFailures     = 3
)

// ErrLazyStarting is returned when the wait budget elapsed before the container became ready.
var ErrLazyStarting = errors.New("lazy container is starting")

// Docker calls go through these vars so tests can stub them.

var (
	lazyInspect = func(nameOrID string) (types.ContainerJSON, error) {
		if err := Connect(); err != nil {
			return types.ContainerJSON{}, err
		}
		return DockerClient.ContainerInspect(DockerContext, nameOrID)
	}

	lazyStartContainer = func(nameOrID string) error {
		if err := Connect(); err != nil {
			return err
		}
		return DockerClient.ContainerStart(DockerContext, nameOrID, conttype.StartOptions{})
	}

	// Default StopOptions so docker honours the container's own stop timeout.
	lazyStopContainer = func(nameOrID string) error {
		if err := Connect(); err != nil {
			return err
		}
		return DockerClient.ContainerStop(DockerContext, nameOrID, conttype.StopOptions{})
	}

	lazyRemoveContainer = func(nameOrID string) error {
		if err := Connect(); err != nil {
			return err
		}
		return DockerClient.ContainerRemove(DockerContext, nameOrID, conttype.RemoveOptions{Force: true})
	}

	lazyListContainers = func() ([]types.Container, error) {
		if err := Connect(); err != nil {
			return nil, err
		}
		return DockerClient.ContainerList(DockerContext, conttype.ListOptions{All: true})
	}

	lazyContainerIP = GetContainerIPByName

	lazyDial = func(address string, timeout time.Duration) error {
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}

	// Clock seams.
	lazyNow   = time.Now
	lazySleep = time.Sleep

	// Event seam: TriggerEvent writes to the events DB, which tests do not have.
	lazyTriggerEvent = utils.TriggerEvent

	lazyPollInterval = 250 * time.Millisecond
	lazyDialTimeout  = 2 * time.Second
)

// lazyWake is the single-flight handle for an in-progress wake.
type lazyWake struct {
	done chan struct{}
	woke bool
	err  error
}

type lazyEntry struct {
	name            string
	lazy            bool
	running         bool
	idle            time.Duration
	startTimeout    time.Duration
	lastActivity    time.Time
	openConns       int
	failStreak      int
	stoppedByReaper bool

	wake *lazyWake
}

var (
	lazyMu     sync.Mutex
	lazyStates = map[string]*lazyEntry{}
)

func IsLazyLabels(labels map[string]string) bool {
	return labels != nil && labels[LazyLabel] == "true"
}

// lazyDurationLabel parses a duration label, falling back to def on garbage.
func lazyDurationLabel(labels map[string]string, key string, def time.Duration, containerName string) time.Duration {
	raw := strings.TrimSpace(labels[key])
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		utils.Warn("Lazy container " + containerName + ": invalid " + key + "=\"" + raw + "\", falling back to " + def.String())
		return def
	}
	return d
}

// lazyEntryLocked returns the entry for name, creating it if needed. Caller holds lazyMu.
func lazyEntryLocked(name string) *lazyEntry {
	st := lazyStates[name]
	if st == nil {
		st = &lazyEntry{
			name:         name,
			idle:         LazyDefaultIdle,
			startTimeout: LazyDefaultStartTimeout,
		}
		lazyStates[name] = st
	}
	return st
}

// lazyApplyLabels refreshes an entry's tunables from labels; nil when not lazy.
func lazyApplyLabels(name string, labels map[string]string) *lazyEntry {
	lazyMu.Lock()
	defer lazyMu.Unlock()
	return lazyApplyLabelsLocked(name, labels)
}

func lazyApplyLabelsLocked(name string, labels map[string]string) *lazyEntry {
	if !IsLazyLabels(labels) {
		// label removed: stop tracking
		if st := lazyStates[name]; st != nil && st.wake == nil {
			delete(lazyStates, name)
		} else if st != nil {
			st.lazy = false
		}
		return nil
	}

	st := lazyEntryLocked(name)
	st.lazy = true
	st.idle = lazyDurationLabel(labels, LazyIdleLabel, LazyDefaultIdle, name)
	st.startTimeout = lazyDurationLabel(labels, LazyStartTimeoutLabel, LazyDefaultStartTimeout, name)
	return st
}

// lazyParseStartedAt parses docker's RFC3339Nano StartedAt, falling back to now.
func lazyParseStartedAt(raw string) time.Time {
	if raw == "" {
		return lazyNow()
	}
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil || t.IsZero() || t.Year() <= 1 {
		return lazyNow()
	}
	return t
}

// EnsureLazyAwake wakes a dormant lazy container before the proxy dials it; maxWait <= 0 waits for the full wake.
func EnsureLazyAwake(containerName string, port string, maxWait time.Duration) (bool, error) {
	lazyMu.Lock()
	st := lazyStates[containerName]
	if st == nil || !st.lazy || st.running {
		lazyMu.Unlock()
		return false, nil
	}

	// Single-flight: a burst of concurrent callers triggers exactly one start.
	w := st.wake
	if w == nil {
		w = &lazyWake{done: make(chan struct{})}
		st.wake = w
		go lazyRunWake(containerName, port, w)
	}
	lazyMu.Unlock()

	if maxWait <= 0 {
		<-w.done
		return w.woke, w.err
	}

	timer := time.NewTimer(maxWait)
	defer timer.Stop()

	select {
	case <-w.done:
		return w.woke, w.err
	case <-timer.C:
		// the wake keeps running in the background; the next request finds it up
		return false, ErrLazyStarting
	}
}

func lazyRunWake(containerName string, port string, w *lazyWake) {
	woke, err := lazyDoWake(containerName, port)

	lazyMu.Lock()
	w.woke = woke
	w.err = err
	if st := lazyStates[containerName]; st != nil && st.wake == w {
		st.wake = nil
	}
	lazyMu.Unlock()

	close(w.done)
}

func lazyDoWake(containerName string, port string) (bool, error) {
	insp, err := lazyInspect(containerName)
	if err != nil {
		lazyWakeFailed(containerName, nil, "inspect failed: "+err.Error())
		return false, err
	}

	labels := map[string]string{}
	if insp.Config != nil && insp.Config.Labels != nil {
		labels = insp.Config.Labels
	}

	st := lazyApplyLabels(containerName, labels)
	if st == nil {
		return false, nil
	}

	if insp.State != nil && insp.State.Running {
		lazyWakeSucceeded(containerName, insp.State.StartedAt)
		return false, nil
	}

	if err := lazyStartContainer(containerName); err != nil {
		lazyWakeFailed(containerName, labels, "start failed: "+err.Error())
		return false, err
	}
	// the IP can change across a stop/start; drop the cached one
	ForgetContainerIP(containerName)

	lazyMu.Lock()
	deadline := lazyNow().Add(st.startTimeout)
	lazyMu.Unlock()

	if err := lazyWaitReady(containerName, port, deadline); err != nil {
		lazyWakeFailed(containerName, labels, err.Error())
		return false, err
	}

	lazyWakeSucceeded(containerName, "")

	lazyTriggerEvent(
		"cosmos.container.lazy.started",
		"Lazy container woken up",
		"debug",
		"container@"+containerName,
		map[string]interface{}{
			"container": containerName,
		},
	)

	return true, nil
}

// Readiness is the docker healthcheck when defined, else a TCP accept on port.
func lazyWaitReady(containerName string, port string, deadline time.Time) error {
	for {
		if lazyIsReady(containerName, port) {
			return nil
		}
		if !lazyNow().Before(deadline) {
			return errors.New("timed out waiting for lazy container " + containerName + " to become ready")
		}
		lazySleep(lazyPollInterval)
	}
}

func lazyIsReady(containerName string, port string) bool {
	insp, err := lazyInspect(containerName)
	if err != nil || insp.State == nil {
		return false
	}

	if insp.State.Health != nil {
		// a healthcheck is authoritative: never fall through to the TCP probe
		return insp.State.Health.Status == types.Healthy
	}

	if !insp.State.Running {
		return false
	}

	if port == "" {
		// nothing to probe: a running container is as ready as we can tell
		return true
	}

	// Cosmos-in-container without host networking cannot reach the container network: the dial fails until timeout.
	ip, err := lazyContainerIP(containerName)
	if err != nil || ip == "" {
		return false
	}

	return lazyDial(net.JoinHostPort(ip, port), lazyDialTimeout) == nil
}

func lazyWakeSucceeded(containerName string, startedAt string) {
	lazyMu.Lock()
	defer lazyMu.Unlock()

	st := lazyStates[containerName]
	if st == nil {
		return
	}
	st.running = true
	st.failStreak = 0
	st.stoppedByReaper = false
	// a wake we performed is fresh activity; an already-running container keeps docker's StartedAt
	if startedAt == "" {
		st.lastActivity = lazyNow()
	} else {
		st.lastActivity = lazyParseStartedAt(startedAt)
	}
}

// lazyWakeFailed records a failed wake; after LazyMaxWakeFailures a deployment container is removed and the streak resets.
func lazyWakeFailed(containerName string, labels map[string]string, reason string) {
	lazyMu.Lock()
	st := lazyStates[containerName]
	streak := 0
	if st != nil {
		st.running = false
		st.failStreak++
		streak = st.failStreak
	}
	lazyMu.Unlock()

	utils.Warn("Lazy container " + containerName + " failed to wake (" + reason + "), streak " + strconv.Itoa(streak))

	if streak < LazyMaxWakeFailures {
		return
	}

	if labels != nil && labels[DeploymentLabel] != "" {
		if err := lazyRemoveContainer(containerName); err != nil {
			utils.Warn("Lazy container " + containerName + ": failed to remove after repeated wake failures: " + err.Error())
		} else {
			utils.Log("Lazy container " + containerName + " removed after " + strconv.Itoa(streak) + " failed wakes; the scheduler will re-place it")
		}
	}

	lazyMu.Lock()
	if st = lazyStates[containerName]; st != nil {
		st.failStreak = 0
	}
	lazyMu.Unlock()

	lazyTriggerEvent(
		"cosmos.container.lazy.failed",
		"Lazy container failed to wake up",
		"warning",
		"container@"+containerName,
		map[string]interface{}{
			"container": containerName,
			"reason":    reason,
			"streak":    streak,
		},
	)
}

func LazyTouch(containerName string) {
	lazyMu.Lock()
	defer lazyMu.Unlock()
	if st := lazyStates[containerName]; st != nil && st.lazy {
		st.lastActivity = lazyNow()
	}
}

func LazyConnOpen(containerName string) {
	lazyMu.Lock()
	defer lazyMu.Unlock()
	if st := lazyStates[containerName]; st != nil && st.lazy {
		st.openConns++
		st.lastActivity = lazyNow()
	}
}

func LazyConnClose(containerName string) {
	lazyMu.Lock()
	defer lazyMu.Unlock()
	if st := lazyStates[containerName]; st != nil && st.lazy {
		if st.openConns > 0 {
			st.openConns--
		}
		st.lastActivity = lazyNow()
	}
}

// LazyIsDormant reports whether the container is lazy and currently stopped.
func LazyIsDormant(containerName string) bool {
	lazyMu.Lock()
	defer lazyMu.Unlock()
	st := lazyStates[containerName]
	return st != nil && st.lazy && !st.running
}

// lazyOnContainerEvent keeps the lazy table current from the docker event stream.
func lazyOnContainerEvent(action string, containerID string, containerName string, attributes map[string]string) (suppress bool, levelOverride string) {
	name := strings.TrimPrefix(containerName, "/")
	if name == "" {
		name = containerID
	}

	switch action {
	case "create", "start":
		labels := lazyLabelsForEvent(containerID, attributes)
		st := lazyApplyLabels(name, labels)
		if st == nil {
			return false, ""
		}

		if action == "start" {
			lazyMu.Lock()
			st.running = true
			st.stoppedByReaper = false
			st.failStreak = 0
			// Seed activity so a freshly started container survives the next tick.
			st.lastActivity = lazyNow()
			lazyMu.Unlock()

			// a wake must not drag bootstrap / compose export along every time
			return true, ""
		}
		return false, ""

	case "die", "stop", "kill":
		lazyMu.Lock()
		st := lazyStates[name]
		if st == nil || !st.lazy {
			lazyMu.Unlock()
			return false, ""
		}
		st.running = false
		// one `docker stop` emits kill, die, stop; only `die` consumes the reaper flag
		wasReaped := st.stoppedByReaper
		if action == "die" {
			st.stoppedByReaper = false
		}
		lazyMu.Unlock()

		if action == "die" && wasReaped {
			return false, "debug"
		}
		return false, ""

	case "destroy":
		lazyMu.Lock()
		if st := lazyStates[name]; st != nil && st.wake == nil {
			delete(lazyStates, name)
		}
		lazyMu.Unlock()
		return false, ""
	}

	return false, ""
}

// lazyLabelsForEvent prefers a fresh inspect, falling back to the event's attributes.
func lazyLabelsForEvent(containerID string, attributes map[string]string) map[string]string {
	if containerID != "" {
		if insp, err := lazyInspect(containerID); err == nil && insp.Config != nil && insp.Config.Labels != nil {
			return insp.Config.Labels
		}
	}
	if attributes == nil {
		return map[string]string{}
	}
	return attributes
}

var (
	lazyReaperMu   sync.Mutex
	lazyReaperStop chan struct{}
)

// StartLazyReaper seeds the activity table and starts the idle ticker. Idempotent.
func StartLazyReaper() {
	lazyReaperMu.Lock()
	if lazyReaperStop != nil {
		lazyReaperMu.Unlock()
		return
	}
	stop := make(chan struct{})
	lazyReaperStop = stop
	lazyReaperMu.Unlock()

	lazyRescan()

	go func() {
		ticker := time.NewTicker(LazyReaperInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				lazyReaperTick()
			}
		}
	}()
}

func StopLazyReaper() {
	lazyReaperMu.Lock()
	defer lazyReaperMu.Unlock()
	if lazyReaperStop == nil {
		return
	}
	close(lazyReaperStop)
	lazyReaperStop = nil
}

// lazyRescan rebuilds the lazy table from docker; the event stream keeps it current afterwards.
func lazyRescan() {
	containers, err := lazyListContainers()
	if err != nil {
		utils.Error("Lazy containers: initial rescan failed", err)
		return
	}

	seen := map[string]bool{}

	for _, c := range containers {
		if !IsLazyLabels(c.Labels) {
			continue
		}
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		if name == "" {
			name = c.ID
		}
		seen[name] = true

		st := lazyApplyLabels(name, c.Labels)
		if st == nil {
			continue
		}

		running := c.State == "running"
		startedAt := ""
		if insp, ierr := lazyInspect(c.ID); ierr == nil && insp.State != nil {
			running = insp.State.Running
			startedAt = insp.State.StartedAt
		}

		lazyMu.Lock()
		st.running = running
		// seed from StartedAt so long-idle containers are reaped on the first tick
		if running {
			st.lastActivity = lazyParseStartedAt(startedAt)
		} else {
			st.lastActivity = lazyNow()
		}
		lazyMu.Unlock()
	}

	lazyMu.Lock()
	for name, st := range lazyStates {
		if !seen[name] && st.wake == nil {
			delete(lazyStates, name)
		}
	}
	lazyMu.Unlock()

	utils.Debug("Lazy containers: tracking " + strconv.Itoa(len(seen)) + " container(s)")
}

// lazyReaperTick stops running lazy containers idle longer than their idle window.
func lazyReaperTick() {
	now := lazyNow()

	lazyMu.Lock()
	var candidates []string
	for name, st := range lazyStates {
		if !st.lazy || !st.running || st.wake != nil {
			continue
		}
		if st.openConns > 0 {
			continue
		}
		if now.Sub(st.lastActivity) <= st.idle {
			continue
		}
		// claim before unlocking: the `die` event can land while ContainerStop is in flight
		st.stoppedByReaper = true
		candidates = append(candidates, name)
	}
	lazyMu.Unlock()

	for _, name := range candidates {
		if err := lazyStopContainer(name); err != nil {
			utils.Warn("Lazy containers: failed to stop idle container " + name + ": " + err.Error())
			lazyMu.Lock()
			if st := lazyStates[name]; st != nil {
				st.stoppedByReaper = false
			}
			lazyMu.Unlock()
			continue
		}

		lazyMu.Lock()
		if st := lazyStates[name]; st != nil {
			st.running = false
		}
		lazyMu.Unlock()

		utils.Debug("Lazy containers: stopped idle container " + name)

		lazyTriggerEvent(
			"cosmos.container.lazy.stopped",
			"Lazy container stopped while idle",
			"debug",
			"container@"+name,
			map[string]interface{}{
				"container": name,
			},
		)
	}
}
