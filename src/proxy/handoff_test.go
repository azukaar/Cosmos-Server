package proxy

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
)

// End to end: a client connects to a socket-proxy listener of an HTTP ":<port>"
// route, the connection is handed off to a real http.Server serving routes
// built by RouterGen. Clients appear with a public or a constellation address
// (fakeAddrListener), so loopback trust never hides what the router decides.

const (
	publicClient = "203.0.113.7:40000"
	nebulaClient = "192.168.201.5:40000"
)

type fakeAddrConn struct {
	net.Conn
	remote net.Addr
}

func (c *fakeAddrConn) RemoteAddr() net.Addr { return c.remote }

// fakeAddrListener makes every accepted connection report the given remote address
type fakeAddrListener struct {
	net.Listener
	remote net.Addr
}

func (l *fakeAddrListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &fakeAddrConn{Conn: c, remote: l.remote}, nil
}

func handoffTestConfig(t *testing.T) {
	t.Helper()
	cfg := utils.Config{}
	cfg.MonitoringDisabled = true
	cfg.HTTPConfig.Hostname = "cosmos.example"
	cfg.HTTPConfig.HTTPPort = "80"
	cfg.HTTPConfig.HTTPSPort = "443"
	cfg.HTTPConfig.AcceptAllInsecureHostname = true
	cfg.HTTPConfig.AuthPrivateKey = strings.Repeat("k", 128)
	utils.LoadBaseMainConfig(cfg)

	prev := utils.IsConstellationIP
	utils.IsConstellationIP = func(ip string) bool { return ip == "192.168.201.5" }
	prevMetrics := utils.PushShieldMetrics
	utils.PushShieldMetrics = func(string) {}
	t.Cleanup(func() {
		utils.IsConstellationIP = prev
		utils.PushShieldMetrics = prevMetrics
	})
}

// nameHandler answers with the route name, the address the server saw and the handoff port
func nameHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		port, _ := HandoffPort(r)
		io.WriteString(w, name+"|"+r.RemoteAddr+"|"+port)
	})
}

func handoffRoute(name, host string, restricted bool) utils.ProxyRouteConfig {
	return utils.ProxyRouteConfig{
		Name:                    name,
		UseHost:                 true,
		Host:                    host,
		Mode:                    "STATIC",
		RestrictToConstellation: restricted,
		DisableHeaderHardening:  true,
	}
}

func testCertificate(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{"vault.example.com"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

type handoffStack struct {
	srv       *http.Server
	proxyAddr string // the socket-proxy listener
	port      string
	tls       bool
}

// listenPort reserves the proxy listener first, so routes can be written for its port
func reserveProxyListener(t *testing.T, client string) (net.Listener, string) {
	t.Helper()
	ln := listenLocal(t)
	remote, err := net.ResolveTCPAddr("tcp", client)
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	return &fakeAddrListener{Listener: ln, remote: remote}, port
}

// startHandoffStack serves the routes behind a socket-proxy listener. withTLS
// runs the server as an HTTPS one; plain forces the plain-HTTP handoff.
func startHandoffStack(t *testing.T, ln net.Listener, port string, routes []utils.ProxyRouteConfig, withTLS, plain bool) *handoffStack {
	t.Helper()

	router := mux.NewRouter()
	for _, route := range routes {
		if RouterGen(route, router, nameHandler(route.Name)) == nil {
			t.Fatalf("RouterGen refused %s", route.Name)
		}
	}

	srv := &http.Server{
		Handler:     utils.ClientRealIP(router),
		ConnContext: HandoffConnContext,
	}
	if withTLS {
		cert := testCertificate(t)
		srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
	StartHTTPHandoff(srv, withTLS)
	t.Cleanup(func() {
		srv.Close()
		currentHTTPHandoff.Store(nil)
	})

	info := &ProxyInfo{stop: make(chan bool), stopped: make(chan bool)}
	pair := PortsPair{From: ":" + port, To: "localhost:443", isHTTPProxy: true, plainHTTP: plain, route: routes[0]}
	done := make(chan struct{})
	go func() {
		defer close(done)
		handleTCPProxy(ln, "localhost:443", pair, info)
	}()
	t.Cleanup(func() {
		close(info.stop)
		ln.Close()
		<-done
	})

	return &handoffStack{srv: srv, proxyAddr: ln.Addr().String(), port: port, tls: withTLS && !plain}
}

// get sends one request through the socket-proxy listener with the given Host
func (s *handoffStack) get(t *testing.T, host string) (int, string) {
	t.Helper()
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "tcp", s.proxyAddr)
		},
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: testTimeout}

	scheme := "http"
	if s.tls {
		scheme = "https"
	}
	req, err := http.NewRequest("GET", scheme+"://"+s.proxyAddr+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = host
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request with Host %s: %v", host, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

// The reported bypass: a public ":<port>" route must not be a way into a
// Constellation-only hostname route.
func TestHandoff_PublicPortRouteDoesNotOpenRestrictedHostRoute(t *testing.T) {
	for _, withTLS := range []bool{false, true} {
		handoffTestConfig(t)
		ln, port := reserveProxyListener(t, publicClient)
		s := startHandoffStack(t, ln, port, []utils.ProxyRouteConfig{
			handoffRoute("public-port", ":"+port, false),
			handoffRoute("vault", "vault.example.com", true),
		}, withTLS, false)

		if code, body := s.get(t, "vault.example.com"); code != http.StatusForbidden {
			t.Fatalf("tls=%v: public client reached the restricted route through the port route: %d %q", withTLS, code, body)
		}

		code, body := s.get(t, "anything:"+port)
		if code != http.StatusOK || !strings.HasPrefix(body, "public-port|"+publicClient+"|"+port) {
			t.Fatalf("tls=%v: public port route: %d %q, want it served with the client's real address", withTLS, code, body)
		}
	}
}

func TestHandoff_ConstellationClientStillReachesRestrictedRoutes(t *testing.T) {
	handoffTestConfig(t)
	ln, port := reserveProxyListener(t, nebulaClient)
	s := startHandoffStack(t, ln, port, []utils.ProxyRouteConfig{
		handoffRoute("restricted-port", ":"+port, true),
		handoffRoute("vault", "vault.example.com", true),
	}, true, false)

	if code, body := s.get(t, "vault.example.com"); code != http.StatusOK || !strings.HasPrefix(body, "vault|"+nebulaClient) {
		t.Fatalf("constellation client on the restricted host route: %d %q", code, body)
	}
	if code, body := s.get(t, "anything:"+port); code != http.StatusOK || !strings.HasPrefix(body, "restricted-port|"+nebulaClient) {
		t.Fatalf("constellation client on the restricted port route: %d %q", code, body)
	}
}

// the restriction of a ":<port>" route is now enforced by the router, on the real address
func TestHandoff_RestrictedPortRouteBlocksPublicClient(t *testing.T) {
	handoffTestConfig(t)
	ln, port := reserveProxyListener(t, publicClient)
	s := startHandoffStack(t, ln, port, []utils.ProxyRouteConfig{
		handoffRoute("restricted-port", ":"+port, true),
	}, true, false)

	if code, body := s.get(t, "anything:"+port); code != http.StatusForbidden {
		t.Fatalf("public client on a restricted port route: %d %q", code, body)
	}
}

// the Host header cannot move a request from one port route to another
func TestHandoff_PortRouteOnlyMatchesItsOwnListener(t *testing.T) {
	handoffTestConfig(t)
	ln, port := reserveProxyListener(t, publicClient)
	other := "65000"
	if port == other {
		other = "65001"
	}
	s := startHandoffStack(t, ln, port, []utils.ProxyRouteConfig{
		handoffRoute("this-port", ":"+port, false),
		handoffRoute("other-port", ":"+other, false),
	}, true, false)

	code, body := s.get(t, "anything:"+other)
	if strings.HasPrefix(body, "other-port|") {
		t.Fatalf("a request through :%s reached the :%s route: %d %q", port, other, code, body)
	}
	if code != http.StatusNotFound {
		t.Fatalf("Host for another port: %d %q, want 404", code, body)
	}
}

// with no socket proxy (docker port mapping), port routes keep matching on Host
func TestHandoff_PortRouteMatchesOnHostWithoutHandoff(t *testing.T) {
	handoffTestConfig(t)
	router := mux.NewRouter()
	RouterGen(handoffRoute("port-route", ":9090", false), router, nameHandler("port-route"))

	r := httptest.NewRequest("GET", "http://anything:9090/", nil)
	r.RemoteAddr = publicClient
	w := httptest.NewRecorder()
	utils.ClientRealIP(router).ServeHTTP(w, r)

	if w.Code != http.StatusOK || !strings.HasPrefix(w.Body.String(), "port-route|") {
		t.Fatalf("direct request on a port route: %d %q", w.Code, w.Body.String())
	}
}

// a service on this machine keeps reaching restricted routes over loopback
func TestHandoff_LocalProcessStillPassesRestriction(t *testing.T) {
	handoffTestConfig(t)
	router := mux.NewRouter()
	RouterGen(handoffRoute("vault", "vault.example.com", true), router, nameHandler("vault"))

	for addr, want := range map[string]int{
		"127.0.0.1:5555": http.StatusOK,
		"[::1]:5555":     http.StatusOK,
		publicClient:     http.StatusForbidden,
	} {
		r := httptest.NewRequest("GET", "http://vault.example.com/", nil)
		r.RemoteAddr = addr
		w := httptest.NewRecorder()
		utils.ClientRealIP(router).ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s on a restricted route: %d, want %d", addr, w.Code, want)
		}
	}
}

// AllowHTTPLocalIPAccess: a local-IP port route is served over plain HTTP by
// the full router, even on a server that serves HTTPS (it used to 404)
func TestHandoff_PlainHandoffOnHTTPSServerServesTheRouter(t *testing.T) {
	handoffTestConfig(t)
	ln, port := reserveProxyListener(t, "192.168.1.20:40000")
	s := startHandoffStack(t, ln, port, []utils.ProxyRouteConfig{
		handoffRoute("local-ip", "192.168.1.10:"+port, false),
	}, true, true)

	code, body := s.get(t, "192.168.1.10:"+port)
	if code != http.StatusOK || !strings.HasPrefix(body, "local-ip|192.168.1.20:40000|"+port) {
		t.Fatalf("plain handoff on an HTTPS server: %d %q", code, body)
	}
}

func TestHandoff_ClosedServerRejectsPromptly(t *testing.T) {
	srv := &http.Server{Handler: http.NotFoundHandler(), ConnContext: HandoffConnContext}
	StartHTTPHandoff(srv, false)
	t.Cleanup(func() { currentHTTPHandoff.Store(nil) })

	// let Serve pick the listener up, then shut down as RestartHTTPServer does
	time.Sleep(50 * time.Millisecond)
	srv.Shutdown(context.Background())

	a, b := net.Pipe()
	defer b.Close()
	start := time.Now()
	if err := handOffHTTPConn(a, "8080", false); err == nil {
		t.Fatal("handoff to a shut down server succeeded")
	}
	if time.Since(start) > time.Second {
		t.Fatalf("handoff to a shut down server took %s", time.Since(start))
	}
}

func TestHandoff_NoServerYet(t *testing.T) {
	currentHTTPHandoff.Store(nil)
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	if err := handOffHTTPConn(a, "8080", false); err == nil {
		t.Fatal("handoff without a server succeeded")
	}
}

func TestHandoffConnContext(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	tagged := &handoffConn{Conn: a, listenPort: "8080"}
	for name, c := range map[string]net.Conn{
		"plain": tagged,
		"tls":   tls.Server(tagged, &tls.Config{}),
	} {
		r := httptest.NewRequest("GET", "/", nil).WithContext(HandoffConnContext(context.Background(), c))
		if port, ok := HandoffPort(r); !ok || port != "8080" {
			t.Fatalf("%s: HandoffPort = %q, %v", name, port, ok)
		}
	}

	r := httptest.NewRequest("GET", "/", nil).WithContext(HandoffConnContext(context.Background(), a))
	if _, ok := HandoffPort(r); ok {
		t.Fatal("a connection accepted directly was tagged as handed off")
	}
}

func TestHandoffIsPlain(t *testing.T) {
	tests := []struct {
		host       string
		isHTTPS    bool
		allowLocal bool
		wantPlain  bool
	}{
		{":8080", false, false, true}, // HTTP-only server
		{":8080", true, false, false}, // HTTPS server
		{":8080", true, true, false},  // wildcard host is no local IP
		{"0.0.0.0:8080", true, true, false},
		{"192.168.1.10:8080", true, true, true},   // the condition that never fired before
		{"192.168.1.10:8080", true, false, false}, // local access not allowed
		{"[fd00::1]:8080", true, true, true},
		{"203.0.113.9:8080", true, true, false},
		{"app.example.com:8080", true, true, false},
	}
	for _, test := range tests {
		if got := handoffIsPlain(test.host, test.isHTTPS, test.allowLocal); got != test.wantPlain {
			t.Errorf("handoffIsPlain(%q, https=%v, allowLocal=%v) = %v, want %v", test.host, test.isHTTPS, test.allowLocal, got, test.wantPlain)
		}
	}
}
