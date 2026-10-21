package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/utils"
)

const testPeerIP = "10.42.0.9"

// two-node tunnel: "self" is this node, "peer" is remote and comes first so plain LB always picks it
func newTunnelTestHandler(t *testing.T) (http.Handler, *[]string, *[]string) {
	t.Helper()

	localSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("LOCAL"))
	}))
	t.Cleanup(localSrv.Close)

	remoteMarkers := []string{}
	remoteAuth := []string{}
	remoteSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteMarkers = append(remoteMarkers, r.Header.Get(TunnelLBForwardedHeader))
		remoteAuth = append(remoteAuth, r.Header.Get("x-cstln-auth"))
		w.Write([]byte("REMOTE"))
	}))
	t.Cleanup(remoteSrv.Close)

	cfg := utils.Config{}
	cfg.ConstellationConfig.Enabled = true
	cfg.ConstellationConfig.ThisDeviceName = "self"
	cfg.HTTPConfig.ProxyConfig.Routes = []utils.ProxyRouteConfig{{
		Name:   "tunnel-route",
		Mode:   "PROXY",
		Target: localSrv.URL,
	}}
	utils.LoadBaseMainConfig(cfg)

	constellation.NebulaStarted.Store(true)
	constellation.CachedDeviceNames = map[string]string{"peer": testPeerIP, "self": testSelfIP}
	constellation.CachedDevices = map[string]utils.ConstellationDevice{
		"peer": {DeviceName: "peer", IP: testPeerIP},
		"self": {DeviceName: "self", IP: testSelfIP, APIKey: testSelfAPIKey},
	}
	t.Cleanup(func() {
		constellation.NebulaStarted.Store(false)
		constellation.CachedDeviceNames = map[string]string{}
		constellation.CachedDevices = map[string]utils.ConstellationDevice{}
	})

	tunnel := utils.ConstellationTunnel{
		Route: utils.ProxyRouteConfig{
			Name:         "tunnel-route",
			Mode:         "PROXY",
			Target:       localSrv.URL,
			LBMode:       "first",
			LBStickyMode: true,
		},
		Targets: []utils.TunnelTarget{
			{DeviceName: "peer", TargetURL: remoteSrv.URL},
			{DeviceName: "self", TargetURL: localSrv.URL},
		},
	}

	return TunnelRouteTo(tunnel, &TunnelLoadBalancer{}), &remoteMarkers, &remoteAuth
}

func serveTunnelRequest(t *testing.T, handler http.Handler, remoteAddr string, marker bool) *http.Response {
	t.Helper()

	r := httptest.NewRequest("GET", "http://tunnel.example/", nil)
	r.RemoteAddr = remoteAddr
	if marker {
		r.Header.Set(TunnelLBForwardedHeader, "1")
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w.Result()
}

func bodyOf(t *testing.T, resp *http.Response) string {
	t.Helper()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return string(b)
}

func TestTunnelRouteToServesForwardedRequestLocally(t *testing.T) {
	handler, remoteMarkers, _ := newTunnelTestHandler(t)

	resp := serveTunnelRequest(t, handler, testPeerIP+":4242", true)

	if body := bodyOf(t, resp); body != "LOCAL" {
		t.Fatalf("forwarded request was re-balanced: got %q, want %q", body, "LOCAL")
	}
	if len(*remoteMarkers) != 0 {
		t.Fatalf("forwarded request hit a remote node again: %v", *remoteMarkers)
	}
	for _, c := range resp.Cookies() {
		if c.Name == "_cosmos_tunnel_lb" {
			t.Fatalf("forwarded request set a sticky cookie: %q", c.Value)
		}
	}
}

// a peer's hop onto a ":<port>" route arrives through the socket proxy from
// loopback; re-balancing it bounced requests between two nodes forever
func TestTunnelRouteToServesLoopbackForwardedRequestLocally(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:4242", "[::1]:4242"} {
		handler, remoteMarkers, _ := newTunnelTestHandler(t)

		resp := serveTunnelRequest(t, handler, addr, true)

		if body := bodyOf(t, resp); body != "LOCAL" {
			t.Fatalf("%s: forwarded request was re-balanced: got %q, want %q", addr, body, "LOCAL")
		}
		if len(*remoteMarkers) != 0 {
			t.Fatalf("%s: forwarded request hit a remote node again: %v", addr, *remoteMarkers)
		}
	}
}

func TestTunnelRouteToIgnoresSpoofedMarker(t *testing.T) {
	handler, remoteMarkers, _ := newTunnelTestHandler(t)

	resp := serveTunnelRequest(t, handler, "1.2.3.4:5678", true)

	if body := bodyOf(t, resp); body != "REMOTE" {
		t.Fatalf("spoofed marker forced local serving: got %q, want %q", body, "REMOTE")
	}
	if len(*remoteMarkers) != 1 {
		t.Fatalf("expected one forward to the remote node, got %v", *remoteMarkers)
	}
}

func TestTunnelRouteToMarksForwardedRequest(t *testing.T) {
	handler, remoteMarkers, _ := newTunnelTestHandler(t)

	resp := serveTunnelRequest(t, handler, "1.2.3.4:5678", false)

	if body := bodyOf(t, resp); body != "REMOTE" {
		t.Fatalf("got %q, want %q", body, "REMOTE")
	}
	if len(*remoteMarkers) != 1 || (*remoteMarkers)[0] != "1" {
		t.Fatalf("remote node did not receive the forwarded marker: %v", *remoteMarkers)
	}
}

func TestTunnelRouteToAuthenticatesForwardedRequest(t *testing.T) {
	handler, _, remoteAuth := newTunnelTestHandler(t)

	resp := serveTunnelRequest(t, handler, "1.2.3.4:5678", false)

	if body := bodyOf(t, resp); body != "REMOTE" {
		t.Fatalf("got %q, want %q", body, "REMOTE")
	}
	if len(*remoteAuth) != 1 || (*remoteAuth)[0] != testSelfAPIKey {
		t.Fatalf("remote node did not receive this node's constellation token: %v", *remoteAuth)
	}
}

// guards the reported symptom: a direct overlay client must get a non-empty sticky key and a stable cookie
// (the cross-node sticky store itself needs NATS JetStream, so only the cookie pins here)
func TestTunnelRouteToPinsDirectOverlayClient(t *testing.T) {
	handler, _, _ := newTunnelTestHandler(t)

	first := serveTunnelRequest(t, handler, testSelfIP+":5000", false)
	bodyOf(t, first)
	second := serveTunnelRequest(t, handler, testSelfIP+":5001", false)
	bodyOf(t, second)

	pin := stickyCookieOf(t, first)
	if pin == "" {
		t.Fatal("no sticky cookie issued for a direct overlay client")
	}
	if got := stickyCookieOf(t, second); got != pin {
		t.Fatalf("sticky target rotated between requests: %q then %q", pin, got)
	}
	if key := clientIDOf(t, testSelfIP+":5000", "", ""); key != testSelfIP {
		t.Fatalf("sticky key for a direct overlay client is %q, want %q", key, testSelfIP)
	}
}

func stickyCookieOf(t *testing.T, resp *http.Response) string {
	t.Helper()

	for _, c := range resp.Cookies() {
		if c.Name == "_cosmos_tunnel_lb" {
			return c.Value
		}
	}
	return ""
}
