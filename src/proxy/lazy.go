package proxy

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/metrics"
	"github.com/azukaar/cosmos-server/src/utils"
)

// LazyWakeBudget bounds how long a request waits for a dormant container before being told to retry.
const LazyWakeBudget = 30 * time.Second

// lazyTargetContainer returns the container name and port a SERVAPP route dials, or "" when it targets none.
func lazyTargetContainer(route utils.ProxyRouteConfig) (name string, port string) {
	if route.Mode != "SERVAPP" {
		return "", ""
	}
	target, err := url.Parse(route.Target)
	if err != nil {
		return "", ""
	}
	port = target.Port()
	if port == "" {
		port = "80"
		if target.Scheme == "https" {
			port = "443"
		}
	}
	return target.Hostname(), port
}

// lazyWakeForDial wakes a dormant lazy container before a dial; a var so tests can stub docker.
var lazyWakeForDial = func(containerName string, port string) error {
	woke, err := docker.EnsureLazyAwake(containerName, port, LazyWakeBudget)
	if woke {
		recordLazyColdStart(containerName)
	}
	return err
}

// lazyIsDormant is a package variable so tests can stub the docker interaction.
var lazyIsDormant = docker.LazyIsDormant

// ProbeParam marks the status HEAD the UI sends to a route (see HostChip).
// ProbeHeader carries the answer when Cosmos replies instead of the app.
const ProbeParam = "__cosmos_probe"
const ProbeHeader = "X-Cosmos-Container"

func isLazyStarting(err error) bool {
	return errors.Is(err, docker.ErrLazyStarting)
}

// lazyTrackConn marks a connection in flight for the idle reaper; call the returned release exactly once.
var lazyTrackConn = func(containerName string) func() {
	docker.LazyTouch(containerName)
	docker.LazyConnOpen(containerName)
	return func() { docker.LazyConnClose(containerName) }
}

func recordLazyColdStart(containerName string) {
	if utils.GetMainConfig().MonitoringDisabled {
		return
	}
	metrics.PushSetMetric("container."+containerName+".cold-starts", 1, metrics.DataDef{
		Max:          0,
		Period:       time.Second * 30,
		Label:        "Cold starts " + containerName,
		AggloType:    "sum",
		SetOperation: "sum",
		Object:       "container@" + containerName,
	})
}

// lazyMiddleware wakes the route's dormant container before the reverse proxy dials it and holds it awake for the request's lifetime.
func lazyMiddleware(route utils.ProxyRouteConfig) func(http.Handler) http.Handler {
	name, port := lazyTargetContainer(route)
	if name == "" {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// the UI status probe must not wake a sleeping container: answer for it instead
			probe := false
			if q := r.URL.Query(); q.Has(ProbeParam) {
				probe = true
				q.Del(ProbeParam)
				r.URL.RawQuery = q.Encode()
			}
			if probe && lazyIsDormant(name) {
				if origin := r.Header.Get("Origin"); origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Expose-Headers", ProbeHeader)
				}
				w.Header().Set(ProbeHeader, "sleeping")
				w.Header().Set("Cache-Control", "no-store")
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			// deferred so the hold covers websocket upgrades and long downloads;
			// a status probe is not activity and must not reset the idle timer
			if !probe {
				release := lazyTrackConn(name)
				defer release()
			}

			err := lazyWakeForDial(name, port)

			if err == nil {
				// the Director resolves the container IP per request
				next.ServeHTTP(w, r)
				return
			}

			if isLazyStarting(err) {
				writeLazyStarting(w, r)
				return
			}

			utils.Error("Lazy container "+name+" failed to start", err)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("lazy container failed to start"))
		})
	}
}

const lazyStartingRetrySeconds = "3"

func writeLazyStarting(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Refresh", lazyStartingRetrySeconds)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(lazyStartingPage))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Retry-After", lazyStartingRetrySeconds)
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte("lazy container starting, retry shortly"))
}

// Self-contained on purpose: served while the backend is down, so no external assets.
const lazyStartingPage = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="refresh" content="` + lazyStartingRetrySeconds + `">
<title>Starting application</title>
<style>
html,body{height:100%;margin:0}
body{display:flex;align-items:center;justify-content:center;background:#101418;color:#e6e9ee;font-family:system-ui,-apple-system,"Segoe UI",Roboto,Helvetica,Arial,sans-serif}
.box{max-width:26rem;padding:2rem;text-align:center}
h1{margin:0 0 .75rem;font-size:1.25rem;font-weight:600}
p{margin:0 0 1.25rem;font-size:.9rem;line-height:1.5;color:#a7b0bd}
.spinner{width:2.25rem;height:2.25rem;margin:0 auto 1.5rem;border:3px solid rgba(255,255,255,.15);border-top-color:#4b9fff;border-radius:50%;animation:spin 1s linear infinite}
@keyframes spin{to{transform:rotate(360deg)}}
@media (prefers-reduced-motion:reduce){.spinner{animation:none}}
</style>
</head>
<body>
<div class="box">
<div class="spinner"></div>
<h1>Starting the application</h1>
<p>This application was asleep and is waking up. This page refreshes automatically.</p>
</div>
</body>
</html>
`
