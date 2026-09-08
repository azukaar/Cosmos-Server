package proxy

import (
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
)

// --- lazy seams ------------------------------------------------------------

type socketWakeCall struct {
	name string
	port string
}

type lazySpy struct {
	mu sync.Mutex

	wakes      []socketWakeCall
	wakeErr    error
	resolved   []string // container names passed to the IP resolver
	ips        []string // IP returned per resolution, last one repeats
	resolveErr error

	opens    int
	releases int

	events chan string // "wake" / "resolve" / "open" / "release", in order
}

func newLazySpy(t *testing.T, ips ...string) *lazySpy {
	t.Helper()
	if len(ips) == 0 {
		ips = []string{"172.17.0.5"}
	}
	s := &lazySpy{ips: ips, events: make(chan string, 32)}

	prevWake := lazyWakeForDial
	prevResolve := dockerGetContainerIPByName
	prevTouch := lazyTrackConn

	lazyWakeForDial = func(name string, port string) error {
		s.mu.Lock()
		s.wakes = append(s.wakes, socketWakeCall{name, port})
		err := s.wakeErr
		s.mu.Unlock()
		s.events <- "wake"
		return err
	}

	dockerGetContainerIPByName = func(name string) (string, error) {
		s.mu.Lock()
		i := len(s.resolved)
		s.resolved = append(s.resolved, name)
		err := s.resolveErr
		ip := s.ips[len(s.ips)-1]
		if i < len(s.ips) {
			ip = s.ips[i]
		}
		s.mu.Unlock()
		s.events <- "resolve"
		if err != nil {
			return "", err
		}
		return ip, nil
	}

	lazyTrackConn = func(name string) func() {
		s.mu.Lock()
		s.opens++
		s.mu.Unlock()
		s.events <- "open"
		var once sync.Once
		return func() {
			once.Do(func() {
				s.mu.Lock()
				s.releases++
				s.mu.Unlock()
				s.events <- "release"
			})
		}
	}

	t.Cleanup(func() {
		lazyWakeForDial = prevWake
		dockerGetContainerIPByName = prevResolve
		lazyTrackConn = prevTouch
	})

	return s
}

func (s *lazySpy) counts() (wakes, resolves, opens, releases int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.wakes), len(s.resolved), s.opens, s.releases
}

func (s *lazySpy) nextEvent(t *testing.T) string {
	t.Helper()
	select {
	case e := <-s.events:
		return e
	case <-time.After(testTimeout):
		t.Fatal("no lazy event happened")
		return ""
	}
}

func (s *lazySpy) waitReleases(t *testing.T, want int) {
	t.Helper()
	deadline := time.Now().Add(testTimeout)
	for {
		_, _, _, releases := s.counts()
		if releases == want {
			return
		}
		if releases > want {
			t.Fatalf("release ran %d times, want %d", releases, want)
		}
		if time.Now().After(deadline) {
			t.Fatalf("release ran %d times, want %d", releases, want)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func lazyPair(container string, port string) PortsPair {
	return PortsPair{
		To:            "tcp://" + container + ":" + port,
		route:         utils.ProxyRouteConfig{Name: "lazy-" + container, Mode: "SERVAPP", Target: "tcp://" + container + ":" + port},
		lazyContainer: container,
		lazyPort:      port,
	}
}

// ---------------------------------------------------------------------------
// per-connection wake + resolution
// ---------------------------------------------------------------------------

func TestHandleTCPProxy_WakesTheContainerBeforeDialing(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t, "172.17.0.9")
	dials := newDialSpy(t)

	h := startTCPProxy(t, lazyPair("sleepy", "5432"))
	h.dialClient(t)

	rec := dials.next(t)
	if rec.addr != "172.17.0.9:5432" {
		t.Fatalf("dialed %q, want the freshly resolved 172.17.0.9:5432", rec.addr)
	}

	lazy.mu.Lock()
	defer lazy.mu.Unlock()
	if len(lazy.wakes) != 1 || lazy.wakes[0] != (socketWakeCall{"sleepy", "5432"}) {
		t.Fatalf("wake calls = %+v, want one call for sleepy:5432", lazy.wakes)
	}
	if len(lazy.resolved) != 1 || lazy.resolved[0] != "sleepy" {
		t.Fatalf("resolver calls = %v, want one for sleepy", lazy.resolved)
	}
}

func TestHandleTCPProxy_WakeHappensBeforeResolveAndDial(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t)
	dials := newDialSpy(t)

	h := startTCPProxy(t, lazyPair("ordered", "80"))
	h.dialClient(t)
	dials.next(t)

	if got := lazy.nextEvent(t); got != "wake" {
		t.Fatalf("first lazy event = %q, want wake", got)
	}
	if got := lazy.nextEvent(t); got != "resolve" {
		t.Fatalf("second lazy event = %q, want resolve", got)
	}
	if got := lazy.nextEvent(t); got != "open" {
		t.Fatalf("third lazy event = %q, want open", got)
	}
}

func TestHandleTCPProxy_StartingContainerClosesClientWithoutDialing(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t)
	lazy.wakeErr = docker.ErrLazyStarting
	dials := newDialSpy(t)

	h := startTCPProxy(t, lazyPair("cold", "8080"))
	client := h.dialClient(t)

	if got := lazy.nextEvent(t); got != "wake" {
		t.Fatalf("first lazy event = %q, want wake", got)
	}
	dials.expectNone(t, 300*time.Millisecond)

	expectHungUpOn(t, client)

	_, resolves, opens, releases := lazy.counts()
	if resolves != 0 {
		t.Fatalf("resolved the IP of a container that is not up yet (%d calls)", resolves)
	}
	if opens != 0 || releases != 0 {
		t.Fatalf("tracked a connection that never happened: opens=%d releases=%d", opens, releases)
	}

	// the loop keeps serving once the container is up
	lazy.mu.Lock()
	lazy.wakeErr = nil
	lazy.mu.Unlock()

	h.dialClient(t)
	if rec := dials.next(t); rec.err != nil {
		t.Fatalf("the accept loop died after a starting container: %v", rec.err)
	}
}

func TestHandleTCPProxy_WakeFailureClosesClientWithoutDialing(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t)
	lazy.wakeErr = errors.New("docker is on fire")
	dials := newDialSpy(t)

	h := startTCPProxy(t, lazyPair("broken", "8080"))
	client := h.dialClient(t)

	lazy.nextEvent(t)
	dials.expectNone(t, 300*time.Millisecond)

	expectHungUpOn(t, client)
}

func TestHandleTCPProxy_ResolveFailureClosesClientWithoutDialing(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t)
	lazy.resolveErr = errors.New("no such container")
	dials := newDialSpy(t)

	h := startTCPProxy(t, lazyPair("ghost", "8080"))
	client := h.dialClient(t)

	lazy.nextEvent(t) // wake
	lazy.nextEvent(t) // resolve
	dials.expectNone(t, 300*time.Millisecond)

	expectHungUpOn(t, client)

	if _, _, opens, releases := lazy.counts(); opens != releases {
		t.Fatalf("connection tracking leaked: opens=%d releases=%d", opens, releases)
	}
}

func TestHandleTCPProxy_ResolvesTheIPPerConnection(t *testing.T) {
	(&shieldSpy{}).install(t)
	newLazySpy(t, "172.17.0.2", "172.17.0.7")
	dials := newDialSpy(t)

	h := startTCPProxy(t, lazyPair("restarts", "5432"))

	h.dialClient(t)
	if rec := dials.next(t); rec.addr != "172.17.0.2:5432" {
		t.Fatalf("first connection dialed %q, want 172.17.0.2:5432", rec.addr)
	}

	h.dialClient(t)
	if rec := dials.next(t); rec.addr != "172.17.0.7:5432" {
		t.Fatalf("second connection dialed %q — the IP was not re-resolved", rec.addr)
	}
}

func TestHandleTCPProxy_TracksTheConnectionForItsWholeLife(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t)
	dials := newDialSpy(t)

	h := startTCPProxy(t, lazyPair("tracked", "5432"))

	client := h.dialClient(t)
	rec := dials.next(t)

	// the connection is up: tracked, not released
	deadline := time.Now().Add(testTimeout)
	for {
		_, _, opens, releases := lazy.counts()
		if opens == 1 && releases == 0 {
			break
		}
		if opens > 1 || releases > 0 {
			t.Fatalf("unexpected tracking state opens=%d releases=%d", opens, releases)
		}
		if time.Now().After(deadline) {
			t.Fatalf("the connection was never tracked: opens=%d", opens)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// client hangs up -> both directions close -> release runs once
	client.Close()
	rec.peer.Close()
	lazy.waitReleases(t, 1)

	// give any duplicate release a chance to show up
	time.Sleep(100 * time.Millisecond)
	if _, _, opens, releases := lazy.counts(); opens != 1 || releases != 1 {
		t.Fatalf("release did not run exactly once: opens=%d releases=%d", opens, releases)
	}
}

func TestHandleTCPProxy_ReleasesTheTrackWhenTheDialFails(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t)
	dials := newDialSpy(t)
	dials.failOn[0] = true

	h := startTCPProxy(t, lazyPair("unreachable", "5432"))
	h.dialClient(t)
	dials.next(t)

	lazy.waitReleases(t, 1)
}

// ---------------------------------------------------------------------------
// paths that must stay untouched
// ---------------------------------------------------------------------------

func TestHandleTCPProxy_NonServappNeverWakes(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t)
	dials := newDialSpy(t)

	// no lazyContainer: plain route, or docker DNS resolves the name inside a container
	h := startTCPProxy(t, PortsPair{To: "tcp://static:1234", route: plainRoute("plain")})
	h.dialClient(t)

	if rec := dials.next(t); rec.addr != "static:1234" {
		t.Fatalf("dialed %q, want the baked-in static:1234", rec.addr)
	}
	if wakes, resolves, opens, _ := lazy.counts(); wakes != 0 || resolves != 0 || opens != 0 {
		t.Fatalf("a non-SERVAPP route touched the lazy path: wakes=%d resolves=%d opens=%d", wakes, resolves, opens)
	}
}

// A tunnel target lives on another node, which wakes its own containers.
func TestHandleTCPProxy_TunnelTargetNeverWakes(t *testing.T) {
	(&shieldSpy{}).install(t)
	lazy := newLazySpy(t)
	dials := newDialSpy(t)

	pair := lazyPair("remote-too", "5432")
	pair.route.Name = "lazy-with-tunnel"
	pair.Targets = []utils.TunnelTarget{{DeviceName: "node-1", TargetURL: "tcp://192.168.201.4:5432"}}

	h := startTCPProxy(t, pair)
	h.dialClient(t)

	if rec := dials.next(t); rec.addr != "192.168.201.4:5432" {
		t.Fatalf("dialed %q, want the tunnel target", rec.addr)
	}
	if wakes, resolves, opens, _ := lazy.counts(); wakes != 0 || resolves != 0 || opens != 0 {
		t.Fatalf("a tunnel hop touched the lazy path: wakes=%d resolves=%d opens=%d", wakes, resolves, opens)
	}
}

// A blocked connection must never wake a sleeping container.
func TestHandleTCPProxy_BlockedConnectionNeverWakes(t *testing.T) {
	(&shieldSpy{block: true}).install(t)
	lazy := newLazySpy(t)
	dials := newDialSpy(t)

	h := startTCPProxy(t, lazyPair("protected", "5432"))
	h.dialClient(t)

	dials.expectNone(t, 300*time.Millisecond)
	if wakes, _, _, _ := lazy.counts(); wakes != 0 {
		t.Fatalf("a shield-blocked connection woke the container (%d wakes)", wakes)
	}
}

// ---------------------------------------------------------------------------
// what addPortPair bakes into the pair
// ---------------------------------------------------------------------------

func dockerlessMode(t *testing.T) {
	t.Helper()
	prevInside, prevHost := utils.IsInsideContainer, utils.IsHostNetwork
	utils.IsInsideContainer = false
	utils.IsHostNetwork = false
	t.Cleanup(func() {
		utils.IsInsideContainer, utils.IsHostNetwork = prevInside, prevHost
	})
}

func TestSocketDestination_KeepsTheContainerNameForDockerlessTCP(t *testing.T) {
	dockerlessMode(t)
	lazy := newLazySpy(t)

	route := utils.ProxyRouteConfig{Name: "db", Mode: "SERVAPP", Target: "tcp://postgres:5432"}
	dest, container, port := socketDestination(route, route.Target)

	if container != "postgres" || port != "5432" {
		t.Fatalf("lazy target = %q:%q, want postgres:5432", container, port)
	}
	if dest != "tcp://postgres:5432" {
		t.Fatalf("destination = %q, want the target kept as-is", dest)
	}
	if _, resolves, _, _ := lazy.counts(); resolves != 0 {
		t.Fatalf("resolved the container IP at build time (%d calls) — it is stale by the first connection", resolves)
	}
}

// Inside a container, docker's own DNS resolves the name, so nothing is deferred.
func TestSocketDestination_InsideContainerKeepsTheName(t *testing.T) {
	prevInside, prevHost := utils.IsInsideContainer, utils.IsHostNetwork
	utils.IsInsideContainer = true
	utils.IsHostNetwork = false
	t.Cleanup(func() { utils.IsInsideContainer, utils.IsHostNetwork = prevInside, prevHost })
	lazy := newLazySpy(t)

	route := utils.ProxyRouteConfig{Name: "db", Mode: "SERVAPP", Target: "tcp://postgres:5432"}
	dest, container, port := socketDestination(route, route.Target)

	if container != "" || port != "" {
		t.Fatalf("lazy target = %q:%q, want none inside a container", container, port)
	}
	if dest != "tcp://postgres:5432" {
		t.Fatalf("destination = %q, want the target untouched", dest)
	}
	if _, resolves, _, _ := lazy.counts(); resolves != 0 {
		t.Fatalf("resolved the container IP inside a container (%d calls)", resolves)
	}
}

// UDP still resolves at build time: there is no accept to hang a wake off.
func TestSocketDestination_UDPStillResolvesAtBuildTime(t *testing.T) {
	dockerlessMode(t)
	lazy := newLazySpy(t, "172.17.0.4")

	route := utils.ProxyRouteConfig{Name: "dns", Mode: "SERVAPP", Target: "udp://coredns:53"}
	dest, container, port := socketDestination(route, route.Target)

	if container != "" || port != "" {
		t.Fatalf("UDP must not be lazy, got %q:%q", container, port)
	}
	if dest != "udp://172.17.0.4:53" {
		t.Fatalf("destination = %q, want udp://172.17.0.4:53", dest)
	}
	if _, resolves, _, _ := lazy.counts(); resolves != 1 {
		t.Fatalf("resolver calls = %d, want 1", resolves)
	}
}

func TestSocketDestination_NonServappUntouched(t *testing.T) {
	dockerlessMode(t)
	lazy := newLazySpy(t)

	route := utils.ProxyRouteConfig{Name: "raw", Mode: "PROXY", Target: "tcp://10.0.0.5:5432"}
	dest, container, port := socketDestination(route, route.Target)

	if dest != "tcp://10.0.0.5:5432" || container != "" || port != "" {
		t.Fatalf("got %q %q %q, want the target untouched", dest, container, port)
	}
	if wakes, resolves, _, _ := lazy.counts(); wakes != 0 || resolves != 0 {
		t.Fatalf("a PROXY route touched docker: wakes=%d resolves=%d", wakes, resolves)
	}
}

// End to end over real sockets: a dormant container wakes, gets a new IP, and the connection is proxied to it.
func TestHandleTCPProxy_WakesThenProxiesToTheNewIP(t *testing.T) {
	(&shieldSpy{}).install(t)
	backend := startEchoBackend(t)
	host, port, err := net.SplitHostPort(backend)
	if err != nil {
		t.Fatal(err)
	}
	lazy := newLazySpy(t, host)

	h := startTCPProxy(t, lazyPair("woken", port))

	client := h.dialClient(t)
	if _, err := client.Write([]byte("hi")); err != nil {
		t.Fatalf("write: %v", err)
	}
	client.SetReadDeadline(time.Now().Add(testTimeout))
	buf := make([]byte, 2)
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(buf) != "hi" {
		t.Fatalf("got %q, want hi", buf)
	}

	client.Close()
	lazy.waitReleases(t, 1)
}
