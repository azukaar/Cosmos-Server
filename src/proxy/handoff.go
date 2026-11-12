package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

// HTTP ":<port>" routes: the socket proxy hands the accepted connection to the
// running http.Server in-process instead of re-dialing localhost. The request
// keeps the client's real address, and its context carries the listener port.

var errHandoffClosed = errors.New("http handoff listener closed")

// how long a handed-off connection waits for the server to accept it
const handoffTimeout = 5 * time.Second

// handoffListener is a net.Listener fed by the socket proxy.
type handoffListener struct {
	conns chan net.Conn
	done  chan struct{}
	once  sync.Once
}

func newHandoffListener() *handoffListener {
	return &handoffListener{conns: make(chan net.Conn), done: make(chan struct{})}
}

func (l *handoffListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.conns:
		return c, nil
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *handoffListener) Close() error {
	l.once.Do(func() { close(l.done) })
	return nil
}

func (l *handoffListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}
}

func (l *handoffListener) deliver(c net.Conn) error {
	timer := time.NewTimer(handoffTimeout)
	defer timer.Stop()
	select {
	case l.conns <- c:
		return nil
	case <-l.done:
		return errHandoffClosed
	case <-timer.C:
		return errors.New("http handoff timed out")
	}
}

// handoffConn is a client connection handed to the HTTP server, tagged with
// the port of the socket-proxy listener that accepted it.
type handoffConn struct {
	net.Conn
	listenPort string
}

type httpHandoff struct {
	tls   *handoffListener // nil when the server does not serve HTTPS
	plain *handoffListener
}

var currentHTTPHandoff atomic.Pointer[httpHandoff]

// StartHTTPHandoff serves the handed-off connections with srv, over TLS
// (srv.TLSConfig) when withTLS, and over plain HTTP in every case. Call it
// before the socket proxy starts; srv.Shutdown closes the listeners.
func StartHTTPHandoff(srv *http.Server, withTLS bool) {
	h := &httpHandoff{plain: newHandoffListener()}
	if withTLS {
		h.tls = newHandoffListener()
		go func() {
			logHandoffServeError(srv.ServeTLS(h.tls, "", ""))
			h.tls.Close()
		}()
	}
	go func() {
		logHandoffServeError(srv.Serve(h.plain))
		h.plain.Close()
	}()
	currentHTTPHandoff.Store(h)
}

func logHandoffServeError(err error) {
	if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
		utils.Error("[SocketProxy] HTTP handoff server stopped", err)
	}
}

// handOffHTTPConn passes a socket-proxy client to the HTTP server. plain
// forces plain HTTP, otherwise TLS whenever the server serves it.
func handOffHTTPConn(c net.Conn, listenPort string, plain bool) error {
	h := currentHTTPHandoff.Load()
	if h == nil {
		return errHandoffClosed
	}
	l := h.plain
	if !plain && h.tls != nil {
		l = h.tls
	}
	return l.deliver(&handoffConn{Conn: c, listenPort: listenPort})
}

type handoffPortKey struct{}

// HandoffConnContext is the http.Server ConnContext that tags requests coming
// from the socket proxy with the port they were accepted on.
func HandoffConnContext(ctx context.Context, c net.Conn) context.Context {
	if tc, ok := c.(*tls.Conn); ok {
		c = tc.NetConn()
	}
	if hc, ok := c.(*handoffConn); ok {
		return context.WithValue(ctx, handoffPortKey{}, hc.listenPort)
	}
	return ctx
}

// HandoffPort returns the socket-proxy port a request came in on, false for a
// request accepted directly by the HTTP server.
func HandoffPort(r *http.Request) (string, bool) {
	port, ok := r.Context().Value(handoffPortKey{}).(string)
	return port, ok
}
