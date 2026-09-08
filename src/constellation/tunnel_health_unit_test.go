package constellation

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
)

// backendServer starts an http server closed at the end of the test.
func backendServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func healthyByName(name string) bool {
	return isTunnelBackendHealthy(utils.ProxyRouteConfig{Name: name})
}

func TestUnitProbeTunnelBackend(t *testing.T) {
	srv := backendServer(t)
	addr := srv.Listener.Addr().String()

	if !probeTunnelBackend(addr) {
		t.Fatalf("probeTunnelBackend(%q) = false on a live backend", addr)
	}

	srv.Close()
	if probeTunnelBackend(addr) {
		t.Errorf("probeTunnelBackend(%q) = true after the backend was closed", addr)
	}
}

func TestUnitTunnelProbeAddress(t *testing.T) {
	setupTestEnv(t, nil)

	tests := []struct {
		name    string
		route   utils.ProxyRouteConfig
		want    string
		verdict probeVerdict
	}{
		{"explicit port", utils.ProxyRouteConfig{Mode: "PROXY", Target: "http://backend:8080"}, "backend:8080", probeDial},
		{"default http port", utils.ProxyRouteConfig{Mode: "PROXY", Target: "http://backend"}, "backend:80", probeDial},
		{"default https port", utils.ProxyRouteConfig{Mode: "PROXY", Target: "https://backend"}, "backend:443", probeDial},
		{"static", utils.ProxyRouteConfig{Mode: "STATIC", Target: "/var/www"}, "", probeSkip},
		{"spa", utils.ProxyRouteConfig{Mode: "SPA", Target: "/var/www"}, "", probeSkip},
		{"redirect", utils.ProxyRouteConfig{Mode: "REDIRECT", Target: "https://elsewhere.com"}, "", probeSkip},
		{"non http scheme", utils.ProxyRouteConfig{Mode: "PROXY", Target: "unix:///var/run/x.sock"}, "", probeSkip},
		{"empty target", utils.ProxyRouteConfig{Mode: "PROXY", Target: ""}, "", probeSkip},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addr, verdict := tunnelProbeAddress(test.route)
			if verdict != test.verdict || addr != test.want {
				t.Errorf("tunnelProbeAddress() = (%q, %v), want (%q, %v)", addr, verdict, test.want, test.verdict)
			}
		})
	}
}

// stubContainer replaces the docker seams for one test.
func stubContainer(t *testing.T, live docker.ContainerLiveness, liveErr error, ip string, ipErr error) {
	t.Helper()
	prevLive, prevIP, prevDormant := probeContainerLiveness, probeContainerIP, probeContainerDormant
	probeContainerLiveness = func(string) (docker.ContainerLiveness, error) { return live, liveErr }
	probeContainerIP = func(string) (string, error) { return ip, ipErr }
	probeContainerDormant = func(string) bool { return false }
	t.Cleanup(func() {
		probeContainerLiveness, probeContainerIP, probeContainerDormant = prevLive, prevIP, prevDormant
	})
}

// A dormant (asleep) lazy container must stay advertised so a request can wake it.
func TestUnitTunnelProbeAddressDormantLazyStaysAdvertised(t *testing.T) {
	setupTestEnv(t, nil)
	route := utils.ProxyRouteConfig{Mode: "SERVAPP", Target: "http://app:8080"}
	stubContainer(t, docker.ContainerLiveness{Exists: true, Running: false}, nil, "", nil)
	probeContainerDormant = func(name string) bool { return name == "app" }

	addr, verdict := tunnelProbeAddress(route)
	if verdict != probeSkip || addr != "" {
		t.Fatalf("tunnelProbeAddress() = (%q, %v), want (\"\", probeSkip)", addr, verdict)
	}

	// once awake it is probed like any other container
	probeContainerDormant = func(string) bool { return false }
	if _, verdict := tunnelProbeAddress(route); verdict != probeDown {
		t.Fatalf("awake-but-stopped verdict = %v, want probeDown", verdict)
	}
}

func TestUnitTunnelProbeAddressServapp(t *testing.T) {
	setupTestEnv(t, nil)
	route := utils.ProxyRouteConfig{Mode: "SERVAPP", Target: "http://app:8080"}

	up := docker.ContainerLiveness{Exists: true, Running: true, Uptime: time.Minute}

	tests := []struct {
		name    string
		live    docker.ContainerLiveness
		liveErr error
		ip      string
		ipErr   error
		want    string
		verdict probeVerdict
	}{
		{"running and settled", up, nil, "172.17.0.9", nil, "172.17.0.9:8080", probeDial},
		{"container gone", docker.ContainerLiveness{}, nil, "", nil, "", probeDown},
		{"restarting", docker.ContainerLiveness{Exists: true}, nil, "", nil, "", probeDown},
		{"running but too young",
			docker.ContainerLiveness{Exists: true, Running: true, Uptime: tunnelProbeMinUptime - time.Millisecond},
			nil, "172.17.0.9", nil, "", probeDown},
		{"running but no IP", up, nil, "", nil, "", probeDown},
		{"docker unreachable", docker.ContainerLiveness{}, errors.New("daemon down"), "", nil, "", probeSkip},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stubContainer(t, test.live, test.liveErr, test.ip, test.ipErr)
			addr, verdict := tunnelProbeAddress(route)
			if verdict != test.verdict || addr != test.want {
				t.Errorf("tunnelProbeAddress() = (%q, %v), want (%q, %v)", addr, verdict, test.want, test.verdict)
			}
		})
	}
}

// A crash-looping SERVAPP container must stay withdrawn, never reverting to unknown-is-healthy.
func TestUnitSweepTunnelBackendsCrashLoopStaysWithdrawn(t *testing.T) {
	setupTestEnv(t, func(cfg *utils.Config) {
		cfg.HTTPConfig.ProxyConfig.Routes = []utils.ProxyRouteConfig{
			{Name: "looping", Mode: "SERVAPP", Target: "http://app:8080", Tunnel: "_ANY_"},
		}
	})

	// restarting: no IP at all, the state the container is in ~99% of the time
	stubContainer(t, docker.ContainerLiveness{Exists: true}, nil, "", errors.New("no IP address found"))

	sweepTunnelBackends()
	sweepTunnelBackends()
	if healthyByName("looping") {
		t.Fatal("crash-looping route still healthy after 2 sweeps")
	}

	// the ~0.5s window where the container IS running, before it dies again
	stubContainer(t, docker.ContainerLiveness{Exists: true, Running: true, Uptime: 500 * time.Millisecond},
		nil, "172.17.0.9", nil)
	sweepTunnelBackends()
	if healthyByName("looping") {
		t.Error("route re-advertised during the brief running window of a crash loop")
	}

	// back to restarting — must not revert to unknown-is-healthy
	stubContainer(t, docker.ContainerLiveness{Exists: true}, nil, "", errors.New("no IP address found"))
	for i := 0; i < 3; i++ {
		sweepTunnelBackends()
		if healthyByName("looping") {
			t.Fatalf("route re-advertised on sweep %d while the container was restarting", i+1)
		}
	}
}

// Docker being unreachable must not withdraw a healthy backend.
func TestUnitSweepTunnelBackendsDockerOutageFailsOpen(t *testing.T) {
	setupTestEnv(t, func(cfg *utils.Config) {
		cfg.HTTPConfig.ProxyConfig.Routes = []utils.ProxyRouteConfig{
			{Name: "app", Mode: "SERVAPP", Target: "http://app:8080", Tunnel: "_ANY_"},
		}
	})

	stubContainer(t, docker.ContainerLiveness{}, errors.New("daemon down"), "", nil)
	for i := 0; i < 3; i++ {
		sweepTunnelBackends()
	}
	if !healthyByName("app") {
		t.Error("route withdrawn because docker could not be reached, want fail-open")
	}
}

func TestUnitSweepTunnelBackendsWithdrawsDeadBackend(t *testing.T) {
	live := backendServer(t)
	dead := backendServer(t)
	deadTarget := dead.URL
	dead.Close()

	setupTestEnv(t, func(cfg *utils.Config) {
		cfg.HTTPConfig.ProxyConfig.Routes = []utils.ProxyRouteConfig{
			{Name: "live", Mode: "PROXY", Target: live.URL, Tunnel: "_ANY_"},
			{Name: "dead", Mode: "PROXY", Target: deadTarget, Tunnel: "_ANY_"},
			{Name: "static", Mode: "STATIC", Target: "/var/www", Tunnel: "_ANY_"},
			{Name: "not-tunneled", Mode: "PROXY", Target: deadTarget},
		}
	})

	// before any sweep everything is advertised
	if !healthyByName("dead") {
		t.Error("dead route is unhealthy before the first sweep, want fail-open")
	}

	sweepTunnelBackends()
	if !healthyByName("dead") {
		t.Error("dead route withdrawn after a single failed sweep, want the debounce to hold it")
	}

	sweepTunnelBackends()
	if healthyByName("dead") {
		t.Error("dead route still healthy after 2 failed sweeps")
	}
	if !healthyByName("live") {
		t.Error("live route marked unhealthy")
	}
	if !healthyByName("static") {
		t.Error("static route marked unhealthy, it has no probeable backend")
	}

	tunnelHealthMux.RLock()
	_, tracked := tunnelHealth["not-tunneled"]
	_, staticTracked := tunnelHealth["static"]
	tunnelHealthMux.RUnlock()
	if tracked {
		t.Error("non-tunneled route was probed")
	}
	if staticTracked {
		t.Error("static route was probed")
	}
}

func TestUnitSweepTunnelBackendsRecovers(t *testing.T) {
	srv := backendServer(t)
	target := srv.URL
	addr := srv.Listener.Addr().String()
	srv.Close()

	setupTestEnv(t, func(cfg *utils.Config) {
		cfg.HTTPConfig.ProxyConfig.Routes = []utils.ProxyRouteConfig{
			{Name: "flappy", Mode: "PROXY", Target: target, Tunnel: "_ANY_"},
		}
	})

	sweepTunnelBackends()
	sweepTunnelBackends()
	if healthyByName("flappy") {
		t.Fatal("route still healthy after 2 failed sweeps")
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Skip("could not rebind the backend port:", err)
	}
	defer listener.Close()

	sweepTunnelBackends()
	if !healthyByName("flappy") {
		t.Error("route not restored on the first successful sweep")
	}
}

func TestUnitGetAllTunneledRoutesSkipsUnhealthyBackends(t *testing.T) {
	live := backendServer(t)
	dead := backendServer(t)
	deadTarget := dead.URL
	dead.Close()

	setupTestEnv(t, func(cfg *utils.Config) {
		cfg.ConstellationConfig.ThisDeviceName = "node-a"
		cfg.HTTPConfig.ProxyConfig.Routes = []utils.ProxyRouteConfig{
			{Name: "live", Mode: "PROXY", Target: live.URL, Tunnel: "_ANY_"},
			{Name: "dead", Mode: "PROXY", Target: deadTarget, Tunnel: "_ANY_"},
		}
	})
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "node-a",
		"cstln_ip":          "192.168.201.1",
		"cstln_api_key":     "test-key",
	})

	names := func() []string {
		out := []string{}
		for _, route := range GetAllTunneledRoutes() {
			out = append(out, route.Name)
		}
		return out
	}

	if got := names(); len(got) != 2 {
		t.Fatalf("before the first sweep GetAllTunneledRoutes() advertised %v, want both routes", got)
	}

	sweepTunnelBackends()
	sweepTunnelBackends()

	got := names()
	if len(got) != 1 || got[0] != "live" {
		t.Errorf("GetAllTunneledRoutes() = %v, want [live] once the dead backend is withdrawn", got)
	}
}
