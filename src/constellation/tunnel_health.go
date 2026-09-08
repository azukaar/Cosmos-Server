package constellation

import (
	"errors"
	"net"
	"net/url"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
)

const (
	tunnelProbeInterval = 5 * time.Second
	tunnelProbeTimeout  = 2 * time.Second
	// consecutive failed sweeps before a backend is withdrawn — a single blip
	// would otherwise restart the HTTP server of every load balancer twice
	tunnelProbeFailuresToWithdraw = 2
	// min uptime before a SERVAPP container counts as serving — a crash-looper is briefly "running"
	tunnelProbeMinUptime = 5 * time.Second
)

// probeVerdict is what a sweep should do with a route; only "cannot probe" may fail open.
type probeVerdict int

const (
	// probeSkip: nothing probeable — keep prior health (deliberately the zero value)
	probeSkip probeVerdict = iota
	// probeDial: dial the address and let the result speak
	probeDial
	// probeDown: known not to be serving, counted as a failed probe
	probeDown
)

// docker call seams, so the probe rules can be unit-tested without a daemon
var (
	probeContainerLiveness = docker.GetContainerLiveness
	probeContainerIP       = docker.GetContainerIPByName
	probeContainerDormant  = docker.LazyIsDormant
)

type tunnelProbeState struct {
	healthy bool
	fails   int
}

var tunnelHealth = map[string]tunnelProbeState{}
var tunnelHealthMux sync.RWMutex

var tunnelProberStopChan chan struct{}
var tunnelProberLock sync.Mutex

// isTunnelBackendHealthy reports whether the local backend of a tunneled route
// answered the last probes. Unknown routes are healthy: the prober has not run
// yet, or the route has no probeable backend.
func isTunnelBackendHealthy(route utils.ProxyRouteConfig) bool {
	tunnelHealthMux.RLock()
	defer tunnelHealthMux.RUnlock()

	state, exists := tunnelHealth[route.Name]
	if !exists {
		return true
	}
	return state.healthy
}

// tunnelProbeAddress resolves the host:port the proxy director would dial for
// a tunneled route's backend, and what the sweep should do with it.
func tunnelProbeAddress(route utils.ProxyRouteConfig) (string, probeVerdict) {
	// static content is served by cosmos itself, a redirect never reaches a backend
	if route.Mode == "STATIC" || route.Mode == "SPA" || route.Mode == "REDIRECT" {
		return "", probeSkip
	}

	target, err := url.Parse(route.Target)
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") {
		return "", probeSkip
	}

	host := target.Hostname()
	if host == "" {
		return "", probeSkip
	}

	port := target.Port()
	if port == "" {
		port = "80"
		if target.Scheme == "https" {
			port = "443"
		}
	}

	// same condition as the proxy director: a SERVAPP hostname is a container
	// name that only Docker can resolve when Cosmos is not on its network
	if route.Mode == "SERVAPP" && (!utils.IsInsideContainer || utils.IsHostNetwork) {
		// a dormant lazy container is asleep by design and gets woken by the
		// proxy on the next request — withdrawing the route would break that
		if probeContainerDormant(host) {
			return "", probeSkip
		}

		live, err := probeContainerLiveness(host)
		if err != nil {
			// docker unreachable is not evidence of anything: fail open
			return "", probeSkip
		}
		if !live.Exists || !live.Running || live.Uptime < tunnelProbeMinUptime {
			return "", probeDown
		}

		ip, err := probeContainerIP(host)
		if err != nil || ip == "" {
			return "", probeDown
		}
		host = ip
	}

	return net.JoinHostPort(host, port), probeDial
}

// isBackendDownError distinguishes a backend that is unambiguously not
// listening from a probe we could not carry out; anything unclassified fails
// open so a broken prober never withdraws a live backend.
func isBackendDownError(err error) bool {
	if os.IsTimeout(err) || errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		// a stopped container disappears from Docker's embedded DNS; the proxy
		// would fail to resolve it exactly the same way
		return true
	}
	return false
}

func probeTunnelBackend(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, tunnelProbeTimeout)
	if err == nil {
		conn.Close()
		return true
	}
	return !isBackendDownError(err)
}

// sweepTunnelBackends probes every tunneled route's backend once and rebuilds
// the health cache, dropping routes that no longer exist in the config.
func sweepTunnelBackends() {
	routes := utils.GetMainConfig().HTTPConfig.ProxyConfig.Routes

	type probeResult struct {
		name string
		up   bool
	}
	results := make(chan probeResult, len(routes))
	pending := 0
	fresh := map[string]bool{}

	for _, route := range routes {
		if route.Tunnel == "" || route.Disabled {
			continue
		}
		addr, verdict := tunnelProbeAddress(route)
		switch verdict {
		case probeSkip:
			continue
		case probeDown:
			// record as failed so the route stays withdrawn rather than
			// reverting to unknown-is-healthy
			fresh[route.Name] = false
		case probeDial:
			pending++
			go func(name string, addr string) {
				results <- probeResult{name: name, up: probeTunnelBackend(addr)}
			}(route.Name, addr)
		}
	}

	for i := 0; i < pending; i++ {
		result := <-results
		fresh[result.name] = result.up
	}

	tunnelHealthMux.Lock()
	defer tunnelHealthMux.Unlock()

	next := map[string]tunnelProbeState{}
	for name, up := range fresh {
		previous := tunnelHealth[name]
		if up {
			next[name] = tunnelProbeState{healthy: true}
			continue
		}
		state := tunnelProbeState{healthy: true, fails: previous.fails + 1}
		if state.fails >= tunnelProbeFailuresToWithdraw {
			state.healthy = false
			if previous.healthy || previous.fails == 0 {
				utils.Warn("[constellation] Tunneled route '" + name + "' backend is not answering, withdrawing it from the constellation")
			}
		}
		next[name] = state
	}
	tunnelHealth = next
}

func StartTunnelProber() {
	StopTunnelProber()

	tunnelProberLock.Lock()
	stopChan := make(chan struct{})
	tunnelProberStopChan = stopChan
	tunnelProberLock.Unlock()

	go func() {
		ticker := time.NewTicker(tunnelProbeInterval)
		defer ticker.Stop()

		sweepTunnelBackends()

		for {
			select {
			case <-stopChan:
				utils.Debug("[constellation] Tunnel backend prober stopped")
				return
			case <-ticker.C:
				sweepTunnelBackends()
			}
		}
	}()
}

func StopTunnelProber() {
	tunnelProberLock.Lock()
	if tunnelProberStopChan != nil {
		close(tunnelProberStopChan)
		tunnelProberStopChan = nil
	}
	tunnelProberLock.Unlock()

	// forget results so a restarted constellation advertises everything until
	// the first sweep lands
	tunnelHealthMux.Lock()
	tunnelHealth = map[string]tunnelProbeState{}
	tunnelHealthMux.Unlock()
}
