package proxy

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
)

const lazyTestTarget = "http://lazyapp:8080"

type wakeCall struct {
	name string
	port string
}

// stubLazy swaps the docker-facing hooks for the test and returns the observed calls and counters.
func stubLazy(t *testing.T, wake func(name, port string) error) (calls *[]wakeCall, opens *int32, releases *int32) {
	t.Helper()

	origWake := lazyWakeForDial
	origTrack := lazyTrackConn
	t.Cleanup(func() {
		lazyWakeForDial = origWake
		lazyTrackConn = origTrack
	})

	seen := []wakeCall{}
	var opened, released int32

	lazyWakeForDial = func(name, port string) error {
		seen = append(seen, wakeCall{name, port})
		if wake == nil {
			return nil
		}
		return wake(name, port)
	}
	lazyTrackConn = func(name string) func() {
		atomic.AddInt32(&opened, 1)
		return func() { atomic.AddInt32(&released, 1) }
	}

	return &seen, &opened, &released
}

func lazyTestRoute() utils.ProxyRouteConfig {
	return utils.ProxyRouteConfig{
		Name:   "lazy-route",
		Mode:   "SERVAPP",
		Target: lazyTestTarget,
	}
}

func okHandler(hit *int32) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has(ProbeParam) {
			panic("probe marker must be stripped before the backend sees the request")
		}
		atomic.AddInt32(hit, 1)
		w.Write([]byte("BACKEND"))
	})
}

func TestLazyMiddlewarePassesThroughWhenAwake(t *testing.T) {
	calls, opens, releases := stubLazy(t, nil)

	var hits int32
	h := lazyMiddleware(lazyTestRoute())(okHandler(&hits))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://lazy.example/", nil))

	if w.Code != http.StatusOK || w.Body.String() != "BACKEND" {
		t.Fatalf("request did not reach the backend: %d %q", w.Code, w.Body.String())
	}
	if len(*calls) != 1 || (*calls)[0] != (wakeCall{"lazyapp", "8080"}) {
		t.Fatalf("unexpected wake calls: %v", *calls)
	}
	if hits != 1 {
		t.Fatalf("backend hit %d times, want 1", hits)
	}
	if *opens != 1 || *releases != 1 {
		t.Fatalf("conn tracking opens=%d releases=%d, want 1/1", *opens, *releases)
	}
}

func TestLazyMiddlewareStartingServesHTMLPage(t *testing.T) {
	_, _, releases := stubLazy(t, func(string, string) error { return docker.ErrLazyStarting })

	var hits int32
	h := lazyMiddleware(lazyTestRoute())(okHandler(&hits))

	r := httptest.NewRequest("GET", "http://lazy.example/", nil)
	r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status is %d, want %d", w.Code, http.StatusBadGateway)
	}
	if hits != 0 {
		t.Fatal("request reached the backend while the container was starting")
	}
	if got := w.Header().Get("Refresh"); got != "3" {
		t.Fatalf("Refresh header is %q, want %q", got, "3")
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type is %q, want text/html", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, `<meta http-equiv="refresh" content="3">`) {
		t.Fatalf("HTML page has no meta refresh:\n%s", body)
	}
	if !strings.Contains(strings.ToLower(body), "starting") {
		t.Fatalf("HTML page does not say the app is starting:\n%s", body)
	}
	// self-contained: no external asset may be referenced
	for _, forbidden := range []string{"<script", "src=", "<link", "http://", "https://"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("HTML page references an external asset (%q):\n%s", forbidden, body)
		}
	}
	if *releases != 1 {
		t.Fatalf("releases=%d, want 1", *releases)
	}
}

func TestLazyMiddlewareStartingServesPlainText(t *testing.T) {
	_, _, releases := stubLazy(t, func(string, string) error { return docker.ErrLazyStarting })

	var hits int32
	h := lazyMiddleware(lazyTestRoute())(okHandler(&hits))

	r := httptest.NewRequest("GET", "http://lazy.example/api", nil)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status is %d, want %d", w.Code, http.StatusBadGateway)
	}
	if hits != 0 {
		t.Fatal("request reached the backend while the container was starting")
	}
	if got := w.Header().Get("Retry-After"); got != "3" {
		t.Fatalf("Retry-After header is %q, want %q", got, "3")
	}
	if got := w.Body.String(); got != "lazy container starting, retry shortly" {
		t.Fatalf("body is %q", got)
	}
	if strings.Contains(w.Header().Get("Content-Type"), "html") {
		t.Fatalf("non-HTML client got an HTML content type: %q", w.Header().Get("Content-Type"))
	}
	if *releases != 1 {
		t.Fatalf("releases=%d, want 1", *releases)
	}
}

func TestLazyMiddlewareGenericFailure(t *testing.T) {
	_, _, releases := stubLazy(t, func(string, string) error { return io.ErrUnexpectedEOF })

	var hits int32
	h := lazyMiddleware(lazyTestRoute())(okHandler(&hits))

	// an HTML client must still get the plain failure, not the "starting" page
	r := httptest.NewRequest("GET", "http://lazy.example/", nil)
	r.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status is %d, want %d", w.Code, http.StatusBadGateway)
	}
	if hits != 0 {
		t.Fatal("request reached the backend after a failed wake")
	}
	if got := w.Body.String(); got != "lazy container failed to start" {
		t.Fatalf("body is %q", got)
	}
	if got := w.Header().Get("Retry-After"); got != "" {
		t.Fatalf("failed wake advertised a retry: %q", got)
	}
	if *releases != 1 {
		t.Fatalf("releases=%d, want 1", *releases)
	}
}

func TestLazyMiddlewareIgnoresNonServappRoutes(t *testing.T) {
	calls, opens, _ := stubLazy(t, nil)

	for _, mode := range []utils.ProxyMode{"PROXY", "STATIC", "SPA", "REDIRECT"} {
		route := lazyTestRoute()
		route.Mode = mode

		var hits int32
		h := lazyMiddleware(route)(okHandler(&hits))

		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://lazy.example/", nil))

		if hits != 1 || w.Body.String() != "BACKEND" {
			t.Fatalf("%s route did not pass through: %d %q", mode, w.Code, w.Body.String())
		}
	}

	if len(*calls) != 0 {
		t.Fatalf("non-SERVAPP routes woke a container: %v", *calls)
	}
	if *opens != 0 {
		t.Fatalf("non-SERVAPP routes tracked %d connections", *opens)
	}
}

func TestLazyMiddlewareReleasesOnPanic(t *testing.T) {
	_, opens, releases := stubLazy(t, nil)

	h := lazyMiddleware(lazyTestRoute())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("backend handler exploded")
	}))

	func() {
		defer func() {
			if recover() == nil {
				t.Error("panic did not propagate out of the middleware")
			}
		}()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "http://lazy.example/", nil))
	}()

	if *opens != 1 || *releases != 1 {
		t.Fatalf("opens=%d releases=%d after a panic, want 1/1", *opens, *releases)
	}
}

// hijackRecorder stands in for the websocket upgrade path (ServeHTTP blocks until the hijacked copies finish).
type hijackRecorder struct {
	*httptest.ResponseRecorder
	conn net.Conn
}

func (h *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return h.conn, bufio.NewReadWriter(bufio.NewReader(h.conn), bufio.NewWriter(h.conn)), nil
}

func TestLazyMiddlewareHoldsHijackedConnection(t *testing.T) {
	_, opens, releases := stubLazy(t, nil)

	serverSide, clientSide := net.Pipe()

	h := lazyMiddleware(lazyTestRoute())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		// mirrors switchProtocolCopier: stays here until the peer goes away
		io.Copy(io.Discard, conn)
		conn.Close()
	}))

	rec := &hijackRecorder{httptest.NewRecorder(), serverSide}
	done := make(chan struct{})
	go func() {
		h.ServeHTTP(rec, httptest.NewRequest("GET", "http://lazy.example/ws", nil))
		close(done)
	}()

	// while the upgraded connection is live, the container must stay held
	deadline := time.After(2 * time.Second)
	for atomic.LoadInt32(opens) == 0 {
		select {
		case <-deadline:
			t.Fatal("connection was never tracked")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	clientSide.Write([]byte("hello"))
	if got := atomic.LoadInt32(releases); got != 0 {
		t.Fatalf("container released while the websocket was still open (releases=%d)", got)
	}

	clientSide.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after the connection closed")
	}

	if atomic.LoadInt32(opens) != 1 || atomic.LoadInt32(releases) != 1 {
		t.Fatalf("opens=%d releases=%d for one upgraded request, want 1/1",
			atomic.LoadInt32(opens), atomic.LoadInt32(releases))
	}
}

// The wake must sit behind every RouterGen gate: an anonymous request must never touch the container.
func TestLazyMiddlewareIsInsideTheAuthGate(t *testing.T) {
	calls, opens, _ := stubLazy(t, nil)

	cfg := utils.Config{}
	cfg.MonitoringDisabled = true
	cfg.HTTPConfig.Hostname = "cosmos.example"
	cfg.HTTPConfig.HTTPPort = "80"
	cfg.HTTPConfig.HTTPSPort = "443"
	cfg.HTTPConfig.AcceptAllInsecureHostname = true
	cfg.HTTPConfig.AuthPrivateKey = strings.Repeat("k", 128)
	utils.LoadBaseMainConfig(cfg)

	route := lazyTestRoute()
	route.Host = "lazy.example"
	route.UseHost = true
	route.AuthEnabled = true
	route.DisableHeaderHardening = true

	router := mux.NewRouter()
	if RouterGen(route, router, RouteTo(route)) == nil {
		t.Fatal("RouterGen refused the route")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "http://lazy.example/", nil))

	if w.Code == http.StatusOK {
		t.Fatalf("anonymous request was served (status %d)", w.Code)
	}
	if len(*calls) != 0 {
		t.Fatalf("anonymous request woke the container: %v", *calls)
	}
	if *opens != 0 {
		t.Fatalf("anonymous request tracked %d connections", *opens)
	}
}

func TestLazyMiddlewareProbeDoesNotWakeDormantContainer(t *testing.T) {
	calls, opens, _ := stubLazy(t, nil)
	origDormant := lazyIsDormant
	t.Cleanup(func() { lazyIsDormant = origDormant })
	dormant := true
	lazyIsDormant = func(name string) bool { return dormant }

	var hit int32
	h := lazyMiddleware(lazyTestRoute())(okHandler(&hit))

	r := httptest.NewRequest("HEAD", "http://lazy.example/?"+ProbeParam+"=1", nil)
	r.Header.Set("Origin", "https://cosmos.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("dormant probe: got status %d, want 503", w.Code)
	}
	if got := w.Header().Get(ProbeHeader); got != "sleeping" {
		t.Fatalf("dormant probe: %s = %q, want sleeping", ProbeHeader, got)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://cosmos.example" {
		t.Fatalf("dormant probe: CORS origin = %q", got)
	}
	if len(*calls) != 0 || atomic.LoadInt32(opens) != 0 || atomic.LoadInt32(&hit) != 0 {
		t.Fatalf("dormant probe must not wake, track or reach the backend: wakes=%d opens=%d hits=%d", len(*calls), *opens, hit)
	}

	// awake: the probe is ordinary traffic and reaches the app
	dormant = false
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("HEAD", "http://lazy.example/?"+ProbeParam+"=1", nil))
	if w.Code != http.StatusOK || atomic.LoadInt32(&hit) != 1 {
		t.Fatalf("awake probe: got status %d hits=%d, want 200 and one hit", w.Code, hit)
	}

	// no marker: a dormant container is woken as usual
	dormant = true
	before := len(*calls)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://lazy.example/", nil))
	if len(*calls) != before+1 {
		t.Fatalf("plain request must wake: wakes went %d -> %d", before, len(*calls))
	}
}
