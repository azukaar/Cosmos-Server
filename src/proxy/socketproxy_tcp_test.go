package proxy

import (
	"errors"
	"io"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

// ---------------------------------------------------------------------------
// harness
// ---------------------------------------------------------------------------

const testTimeout = 3 * time.Second

func listenLocal(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return ln
}

// startEchoBackend is a real TCP echo server, so the happy path runs over real sockets.
func startEchoBackend(t *testing.T) string {
	t.Helper()
	ln := listenLocal(t)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()
	t.Cleanup(func() { ln.Close() })
	return ln.Addr().String()
}

type tcpProxyHarness struct {
	addr     string
	info     *ProxyInfo
	listener net.Listener
	done     chan struct{}
	stopOnce sync.Once
}

// startTCPProxy runs handleTCPProxy against a throwaway loopback listener.
func startTCPProxy(t *testing.T, pair PortsPair) *tcpProxyHarness {
	t.Helper()

	ln := listenLocal(t)
	pair.From = ln.Addr().String()

	h := &tcpProxyHarness{
		addr:     ln.Addr().String(),
		listener: ln,
		info:     &ProxyInfo{stop: make(chan bool), stopped: make(chan bool)},
		done:     make(chan struct{}),
	}

	go func() {
		defer close(h.done)
		handleTCPProxy(ln, stripTargetProtocol(pair.To), pair, h.info)
	}()

	t.Cleanup(func() {
		h.stop()
		ln.Close()
		h.waitStopped(t)
	})

	return h
}

func (h *tcpProxyHarness) stop() {
	h.stopOnce.Do(func() { close(h.info.stop) })
}

func (h *tcpProxyHarness) waitStopped(t *testing.T) {
	t.Helper()
	select {
	case <-h.done:
	case <-time.After(testTimeout):
		t.Error("handleTCPProxy did not return after the stop signal")
	}
}

func (h *tcpProxyHarness) dialClient(t *testing.T) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", h.addr)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// --- shield seam -----------------------------------------------------------

type shieldSpy struct {
	mu      sync.Mutex
	calls   []string // shieldID, one entry per connection that went through the shield
	block   bool
	blocked []net.Conn
}

func (s *shieldSpy) install(t *testing.T) *shieldSpy {
	t.Helper()
	prev := tcpShieldMiddleware
	tcpShieldMiddleware = func(shieldID string, route utils.ProxyRouteConfig) func(net.Conn) net.Conn {
		return func(conn net.Conn) net.Conn {
			s.mu.Lock()
			s.calls = append(s.calls, shieldID)
			block := s.block
			if block {
				s.blocked = append(s.blocked, conn)
			}
			s.mu.Unlock()
			if block {
				return nil
			}
			return conn
		}
	}
	t.Cleanup(func() { tcpShieldMiddleware = prev })
	return s
}

func (s *shieldSpy) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

// --- dial seam -------------------------------------------------------------

type dialRecord struct {
	network string
	addr    string
	peer    net.Conn // the far end of the pipe handed to the proxy (nil on a failed dial)
	err     error
}

type dialSpy struct {
	mu      sync.Mutex
	n       int
	failOn  map[int]bool // 0-based call index -> return an error
	records chan dialRecord
}

func newDialSpy(t *testing.T) *dialSpy {
	t.Helper()
	d := &dialSpy{failOn: map[int]bool{}, records: make(chan dialRecord, 16)}
	prev := netDial
	netDial = func(network, addr string) (net.Conn, error) {
		d.mu.Lock()
		i := d.n
		d.n++
		fail := d.failOn[i]
		d.mu.Unlock()

		if fail {
			err := errors.New("connection refused (test)")
			d.records <- dialRecord{network: network, addr: addr, err: err}
			return nil, err
		}

		proxySide, peer := net.Pipe()
		d.records <- dialRecord{network: network, addr: addr, peer: peer}
		return proxySide, nil
	}
	t.Cleanup(func() { netDial = prev })
	return d
}

func (d *dialSpy) next(t *testing.T) dialRecord {
	t.Helper()
	select {
	case rec := <-d.records:
		return rec
	case <-time.After(testTimeout):
		t.Fatal("no dial was attempted")
		return dialRecord{}
	}
}

func (d *dialSpy) expectNone(t *testing.T, within time.Duration) {
	t.Helper()
	select {
	case rec := <-d.records:
		t.Fatalf("unexpected dial to %s", rec.addr)
	case <-time.After(within):
	}
}

// expectHungUpOn asserts the proxy actively closed the connection; a read timeout means it leaked open.
func expectHungUpOn(t *testing.T, conn net.Conn) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(testTimeout))
	_, err := conn.Read(make([]byte, 1))
	if err == nil {
		t.Fatal("the connection stayed open and readable")
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		t.Fatal("the connection was left open: the read timed out instead of seeing a close")
	}
}

func plainRoute(name string) utils.ProxyRouteConfig {
	return utils.ProxyRouteConfig{Name: name}
}

// ---------------------------------------------------------------------------
// phase 1: the existing TCP path
// ---------------------------------------------------------------------------

func TestHandleTCPProxy_ProxiesBothDirections(t *testing.T) {
	(&shieldSpy{}).install(t)
	backend := startEchoBackend(t)

	h := startTCPProxy(t, PortsPair{To: backend, route: plainRoute("echo")})

	client := h.dialClient(t)
	if _, err := client.Write([]byte("ping")); err != nil {
		t.Fatalf("write: %v", err)
	}

	client.SetReadDeadline(time.Now().Add(testTimeout))
	buf := make([]byte, 4)
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(buf) != "ping" {
		t.Fatalf("got %q back from the backend, want ping", buf)
	}

	// the second direction again, to prove the copy is not one-shot
	if _, err := client.Write([]byte("pong")); err != nil {
		t.Fatalf("second write: %v", err)
	}
	client.SetReadDeadline(time.Now().Add(testTimeout))
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatalf("second read back: %v", err)
	}
	if string(buf) != "pong" {
		t.Fatalf("got %q back, want pong", buf)
	}
}

func TestHandleTCPProxy_DialFailureClosesClientAndKeepsAccepting(t *testing.T) {
	(&shieldSpy{}).install(t)
	dials := newDialSpy(t)
	dials.failOn[0] = true

	h := startTCPProxy(t, PortsPair{To: "backend:1234", route: plainRoute("refused")})

	failed := h.dialClient(t)
	if rec := dials.next(t); rec.err == nil {
		t.Fatal("expected the first dial to fail")
	}

	// the proxy must hang up on the client it could not serve
	expectHungUpOn(t, failed)

	// ...and the accept loop must still be alive
	second := h.dialClient(t)
	rec := dials.next(t)
	if rec.err != nil {
		t.Fatalf("second dial failed: %v", rec.err)
	}
	if _, err := second.Write([]byte("x")); err != nil {
		t.Fatalf("write on the second connection: %v", err)
	}
	rec.peer.SetReadDeadline(time.Now().Add(testTimeout))
	buf := make([]byte, 1)
	if _, err := io.ReadFull(rec.peer, buf); err != nil {
		t.Fatalf("backend never saw the second connection's data: %v", err)
	}
}

func TestHandleTCPProxy_StopChannelExitsLoop(t *testing.T) {
	(&shieldSpy{}).install(t)
	newDialSpy(t)

	h := startTCPProxy(t, PortsPair{To: "backend:1234", route: plainRoute("stopme")})

	h.stop()
	h.waitStopped(t)
}

func TestHandleTCPProxy_DialsTheLoadBalancerTargetWhenTunnelsExist(t *testing.T) {
	(&shieldSpy{}).install(t)
	dials := newDialSpy(t)

	pair := PortsPair{
		To:    "static:1234",
		route: utils.ProxyRouteConfig{Name: "lb-route"},
		Targets: []utils.TunnelTarget{
			{DeviceName: "node-0", TargetURL: "https://192.168.201.1:8603"},
		},
	}
	h := startTCPProxy(t, pair)

	h.dialClient(t)

	rec := dials.next(t)
	if rec.addr != "192.168.201.1:8603" {
		t.Fatalf("dialed %q, want the tunnel target with its scheme stripped", rec.addr)
	}
	if rec.network != "tcp" {
		t.Fatalf("dialed network %q, want tcp", rec.network)
	}
}

func TestHandleTCPProxy_DialsTheStaticTargetWithoutTunnels(t *testing.T) {
	(&shieldSpy{}).install(t)
	dials := newDialSpy(t)

	h := startTCPProxy(t, PortsPair{To: "tcp://static:1234", route: plainRoute("static-route")})

	h.dialClient(t)

	if rec := dials.next(t); rec.addr != "static:1234" {
		t.Fatalf("dialed %q, want static:1234", rec.addr)
	}
}

// the shield is the only place the client's real address exists for an HTTP-target route
func TestHandleTCPProxy_ShieldAppliedForNonHTTPAndRestrictedRoutes(t *testing.T) {
	tests := []struct {
		name        string
		isHTTPProxy bool
		route       utils.ProxyRouteConfig
		wantShield  bool
	}{
		{"non-http route", false, plainRoute("raw-tcp"), true},
		{"plain http route", true, plainRoute("plain-http"), false},
		{"restricted http route", true, utils.ProxyRouteConfig{Name: "restricted", RestrictToConstellation: true}, true},
		{"whitelisted http route", true, utils.ProxyRouteConfig{Name: "whitelisted", WhitelistInboundIPs: []string{"10.0.0.0/8"}}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shield := (&shieldSpy{}).install(t)
			dials := newDialSpy(t)

			h := startTCPProxy(t, PortsPair{To: "backend:1234", isHTTPProxy: test.isHTTPProxy, route: test.route})
			h.dialClient(t)
			dials.next(t) // the connection reached the dial either way

			if got := shield.count() > 0; got != test.wantShield {
				t.Fatalf("shield applied = %v, want %v", got, test.wantShield)
			}
			if test.wantShield && shield.calls[0] != "proxy-"+h.addr {
				t.Fatalf("shield id = %q, want proxy-%s", shield.calls[0], h.addr)
			}
		})
	}
}

func TestHandleTCPProxy_ShieldBlockClosesClientAndKeepsAccepting(t *testing.T) {
	shield := &shieldSpy{block: true}
	shield.install(t)
	dials := newDialSpy(t)

	h := startTCPProxy(t, PortsPair{To: "backend:1234", route: plainRoute("blocked")})

	blocked := h.dialClient(t)
	dials.expectNone(t, 300*time.Millisecond)

	expectHungUpOn(t, blocked)

	shield.mu.Lock()
	shield.block = false
	shield.mu.Unlock()

	h.dialClient(t)
	if rec := dials.next(t); rec.err != nil {
		t.Fatalf("the accept loop died after a blocked connection: %v", rec.err)
	}
}

// both io.Copy goroutines report into one channel; only one send is ever received
func TestHandleClient_LeavesNoCopyGoroutineBehind(t *testing.T) {
	const sessions = 20

	runtime.GC()
	baseline := runtime.NumGoroutine()

	for i := 0; i < sessions; i++ {
		clientSide, clientPeer := net.Pipe()
		serverSide, serverPeer := net.Pipe()
		info := &ProxyInfo{stop: make(chan bool), stopped: make(chan bool)}

		done := make(chan struct{})
		go func() {
			handleClient(clientSide, serverSide, info)
			close(done)
		}()

		clientPeer.Close()

		select {
		case <-done:
		case <-time.After(testTimeout):
			t.Fatal("handleClient never returned after the client hung up")
		}

		// both directions must be closed once handleClient is done
		serverPeer.SetDeadline(time.Now().Add(testTimeout))
		if _, err := serverPeer.Read(make([]byte, 1)); err == nil {
			t.Fatal("the server side was left open")
		}
		serverPeer.Close()
	}

	deadline := time.Now().Add(testTimeout)
	for {
		leaked := runtime.NumGoroutine() - baseline
		if leaked <= sessions/4 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d goroutines still parked after %d proxied connections (leaked copy goroutines)", leaked, sessions)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
