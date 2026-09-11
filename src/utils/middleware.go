package utils

import (
	"context"
	"net/http"
	"time"
	"net"
	"bufio"
	"strings"
	"fmt"
	"sync"
	"os"
	"sync/atomic"
	"io"

	"github.com/mxk/go-flowrate/flowrate"
	"github.com/oschwald/geoip2-golang"
)

// https://github.com/go-chi/chi/blob/master/middleware/timeout.go

var PushShieldMetrics func(string)

type safeInt struct {
	val int64
}

var BannedIPs = sync.Map{}

// Close connection right away if banned (save resources)

func IncrementIPAbuseCounter(ip string) {
	// Load or store a new *safeInt
	actual, _ := BannedIPs.LoadOrStore(ip, &safeInt{})
	counter := actual.(*safeInt)

	// Increment the counter using atomic for concurrent access
	atomic.AddInt64(&counter.val, 1)
}

func GetIPAbuseCounter(ip string) int64 {
	// Load the *safeInt
	actual, ok := BannedIPs.Load(ip)
	if !ok {
			return 0
	}
	counter := actual.(*safeInt)

	// Load the value using atomic for concurrent access
	return atomic.LoadInt64(&counter.val)
}

func ClientRealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := GetClientIP(r)
		if clientID == "" {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "ClientID", clientID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func BlockBannedIPs(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, ok := r.Context().Value("ClientID").(string)
        if !ok || ip == "" {
					if hj, ok := w.(http.Hijacker); ok {
							conn, _, err := hj.Hijack()
							if err == nil {
									conn.Close()
							}
					}
					return
        }

				nbAbuse := GetIPAbuseCounter(ip)

        if nbAbuse > 275 {
					Warn("IP " + ip + " has " + fmt.Sprintf("%d", nbAbuse) + " abuse(s) and will soon be banned.")
				}

        if nbAbuse > 300 {
					if hj, ok := w.(http.Hijacker); ok {
							conn, _, err := hj.Hijack()
							if err == nil {
									conn.Close()
							}
					}
					return
				}

        next.ServeHTTP(w, r)
    })
}

func CleanBannedIPs() {
	BannedIPs.Range(func(key, value interface{}) bool {
		BannedIPs.Delete(key)
		return true
	})
}

func MiddlewareTimeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					Error("Request Timeout. Cancelling.", ctx.Err())
					HTTPError(w, "Gateway Timeout", 
						http.StatusGatewayTimeout, "HTTP002")
					return 
				}
			}()

			w.Header().Set("X-Timeout-Duration", timeout.String())

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

type responseWriter struct {
	http.ResponseWriter
	*flowrate.Writer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func BandwithLimiterMiddleware(max int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if(max > 0) {
				fw := flowrate.NewWriter(w, max)
				w = &responseWriter{w, fw}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func SetSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if(IsHTTPS) {
			// TODO: Add preload if we have a valid certificate
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
				
		next.ServeHTTP(w, r)
	})
}

func SetCosmosHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {		
		w.Header().Set("X-Served-By-Cosmos", "1")
		
		next.ServeHTTP(w, r)
	})
}

// originUnderDomain reports whether reqOrigin (e.g. "https://sub.example.com")
// is the Cosmos hostname itself or one of its subdomains. Host-based app routes
// (app.sub.example.com) are reached cross-origin from the Cosmos UI, so the CORS
// header must allow the UI's origin - but only origins under the Cosmos domain,
// never arbitrary third-party sites.
func originUnderDomain(reqOrigin string, hostname string) bool {
	if reqOrigin == "" || hostname == "" {
		return false
	}
	h := strings.TrimPrefix(strings.TrimPrefix(reqOrigin, "https://"), "http://")
	if i := strings.Index(h, "/"); i >= 0 {
		h = h[:i]
	}
	if i := strings.Index(h, ":"); i >= 0 {
		h = h[:i]
	}
	hostname = strings.TrimPrefix(strings.TrimPrefix(hostname, "https://"), "http://")
	if i := strings.Index(hostname, ":"); i >= 0 {
		hostname = hostname[:i]
	}
	return h == hostname || strings.HasSuffix(h, "."+hostname)
}

func CORSHeader(origin string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if origin != "" {
				hostname := GetMainConfig().HTTPConfig.Hostname
				reqOrigin := r.Header.Get("Origin")
				// If the request comes from the Cosmos UI (its hostname or a
				// subdomain), allow that exact origin so the browser can read the
				// app response. This makes host-based app routes usable from a
				// Cosmos UI on another subdomain.
				if reqOrigin != "" && originUnderDomain(reqOrigin, hostname) {
					w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Vary", "Origin")
			}

			next.ServeHTTP(w, r)
		})
	}
}

func PublicCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Vary", "Origin")

		next.ServeHTTP(w, r)
	})
}

// headProbeWriter wraps http.ResponseWriter so that for a cross-origin HEAD
// probe the Access-Control-Allow-Origin emitted is always the requesting
// origin, regardless of what inner middleware (e.g. the per-route CORSHeader)
// may have set to the app's own host. Without this the browser would compare
// the UI origin against the app host and block the status.
type headProbeWriter struct {
	http.ResponseWriter
	origin     string
	wrote      bool
}

func (h *headProbeWriter) WriteHeader(code int) {
	if h.origin != "" && !h.wrote {
		h.Header().Set("Access-Control-Allow-Origin", h.origin)
		h.Header().Set("Vary", "Origin")
		h.wrote = true
	}
	h.ResponseWriter.WriteHeader(code)
}

func (h *headProbeWriter) Write(b []byte) (int, error) {
	if h.origin != "" && !h.wrote {
		h.Header().Set("Access-Control-Allow-Origin", h.origin)
		h.Header().Set("Vary", "Origin")
		h.wrote = true
	}
	return h.ResponseWriter.Write(b)
}

// Flush forwards to the underlying ResponseWriter when it supports
// http.Flusher. Streaming routes (create service, image pull, container
// update) rely on w.(http.Flusher) succeeding; without this forwarding the
// wrapper would hide Flusher and progress logs would stop streaming
// line-by-line (buffered, arriving in chunks).
func (h *headProbeWriter) Flush() {
	if flusher, ok := h.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// ReadFrom forwards io.ReaderFrom to the underlying ResponseWriter when it
// supports it. Without this method, wrappers that select their behaviour from
// the writer's capability set degrade badly: chi's Logger middleware
// (middleware.NewWrapResponseWriter) checks Flusher && Hijacker &&
// io.ReaderFrom to decide between its httpFancyWriter (which forwards
// Hijack) and its flushWriter (which does NOT implement http.Hijacker). A
// headProbeWriter that only advertised Flusher+Hijacker made chi pick
// flushWriter, and every WebSocket upgrade through the logger then failed
// with "response does not implement http.Hijacker" / http.ErrNotSupported.
func (h *headProbeWriter) ReadFrom(r io.Reader) (int64, error) {
	if h.origin != "" && !h.wrote {
		h.Header().Set("Access-Control-Allow-Origin", h.origin)
		h.Header().Set("Vary", "Origin")
		h.wrote = true
	}
	if rf, ok := h.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(struct{ io.Writer }{h}, r)
}

// Hijack forwards to the underlying ResponseWriter when it supports
// http.Hijacker, so WebSocket/terminal upgrades keep working through the
// wrapper. It first tries the direct writer, then falls back to walking the
// Unwrap() chain so that a non-Hijacker wrapper (e.g. chi's flushWriter)
// sitting between us and the real writer does not break upgrades.
func (h *headProbeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w := h.ResponseWriter
	for {
		if hijacker, ok := w.(http.Hijacker); ok {
			return hijacker.Hijack()
		}
		unwrapper, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return nil, nil, http.ErrNotSupported
		}
		w = unwrapper.Unwrap()
	}
}

// Unwrap lets http.ResponseController and middleware down the chain reach
// the underlying writer (e.g. for Flush, Hijack, and other optional
// interfaces) instead of being cut off by this wrapper.
func (h *headProbeWriter) Unwrap() http.ResponseWriter {
	return h.ResponseWriter
}

// HeadProbeCORS lets the Cosmos UI read responses from host-based app routes
// (app.*.com) that live on a different origin. It applies to any request that
// carries a browser Origin under the Cosmos domain - including an auth-gate 302
// to OpenID login, which is produced inside tokenMiddleware before the per-route
// CORSHeader runs and would otherwise carry no Access-Control-Allow-Origin.
//
// Safe by construction: the origin is only echoed when it is the configured
// Cosmos hostname or one of its subdomains (originUnderDomain). Third-party
// origins are ignored. The wrapper re-asserts ACAO at write time so no inner
// middleware can clobber it.
func HeadProbeCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			hostname := GetMainConfig().HTTPConfig.Hostname
			if originUnderDomain(origin, hostname) {
				h := &headProbeWriter{ResponseWriter: w, origin: origin}
				next.ServeHTTP(h, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func AcceptHeader(accept string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", accept)

			next.ServeHTTP(w, r)
		})
	}
}

// GetIPLocation returns the ISO country code for a given IP address.
func GetIPLocation(ip string) (string, error) {
	geoDB, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		return "", err
	}
	defer geoDB.Close()

	parsedIP := net.ParseIP(ip)
	record, err := geoDB.Country(parsedIP)
	if err != nil {
		return "", err
	}

	return record.Country.IsoCode, nil
}

// BlockByCountryMiddleware returns a middleware function that blocks requests from specified countries.
func BlockByCountryMiddleware(blockedCountries []string, CountryBlacklistIsWhitelist bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, ok := r.Context().Value("ClientID").(string)
			if !ok || ip == "" {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			countryCode, err := GetIPLocation(ip)

			if err == nil {
				config := GetMainConfig()

				if CountryBlacklistIsWhitelist {
					if countryCode != "" {
						Debug("Country code: " + countryCode)
						blocked := true
						for _, blockedCountry := range blockedCountries {
							if config.ServerCountry != countryCode && countryCode == blockedCountry {
								blocked = false
							}
						}

						if blocked {
							PushShieldMetrics("geo")
							IncrementIPAbuseCounter(ip)

							TriggerEvent(
								"cosmos.proxy.shield.geo",
								"Proxy Shield Geo blocked",
								"warning",
								"",
								map[string]interface{}{
								"clientID": ip,
								"country": countryCode,
								"hostname": r.Host,
								"url": r.URL.String(),
							})

							http.Error(w, "Access denied", http.StatusForbidden)
							return
						}
					} else {
						Debug("Missing geolocation information to block IPs")
					}
				} else {
					for _, blockedCountry := range blockedCountries {
						if config.ServerCountry != countryCode && countryCode == blockedCountry {
							PushShieldMetrics("geo")
							IncrementIPAbuseCounter(ip)
	
							TriggerEvent(
								"cosmos.proxy.shield.geo",
								"Proxy Shield Geo blocked",
								"warning",
								"",
								map[string]interface{}{
								"clientID": ip,
								"country": countryCode,
								"hostname": r.Host,
								"url": r.URL.String(),
							})

							http.Error(w, "Access denied", http.StatusForbidden)
							return
						}
					}
				}
			} else {
				Debug("Missing geolocation information to block IPs")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// blockPostWithoutReferer blocks POST requests without a Referer header
func BlockPostWithoutReferer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" || r.Method == "DELETE" {
			referer := r.Header.Get("Referer")
			if referer == "" {
				PushShieldMetrics("referer")
				Error("Blocked POST request without Referer header", nil)
				http.Error(w, "Bad Request: Invalid request.", http.StatusBadRequest)

				ip, _, _ := net.SplitHostPort(r.RemoteAddr)
				if ip != "" {
					TriggerEvent(
						"cosmos.proxy.shield.referer",
						"Proxy Shield Referer blocked",
						"warning",
						"",
						map[string]interface{}{
						"clientID": ip,
						"hostname": r.Host,
						"url": r.URL.String(),
					})

					IncrementIPAbuseCounter(ip)
				}

				return
			}
		}

		// If it's not a POST request or the POST request has a Referer header, pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}

func EnsureHostname(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		og := GetMainConfig().HTTPConfig.Hostname
		ni := GetMainConfig().NewInstall

		if ni || og == "0.0.0.0" {
			next.ServeHTTP(w, r)
			return
		}

		hostnames := GetAllHostnames(false, false)

		reqHostNoPort := strings.Split(r.Host, ":")[0]
		reqPort := ""
		if i := strings.LastIndex(r.Host, ":"); i != -1 {
			reqPort = r.Host[i+1:]
		}

		isOk := false

		for _, hostname := range hostnames {
			hostnameNoPort := strings.Split(hostname, ":")[0]
			if hostnameNoPort == "" {
				// A ":<port>" route Host is a port-wildcard, so any request Host on that port is legitimate.
				if p := strings.TrimPrefix(hostname, ":"); p != "" && p == reqPort {
					isOk = true
				}
				continue
			}
			if reqHostNoPort == hostnameNoPort {
				isOk = true
			}
		}
		
		if(GetMainConfig().HTTPConfig.AllowHTTPLocalIPAccess) {
			if(IsLocalIP(reqHostNoPort)) {
				isOk = true
			}
		}
		
		if !isOk {
			PushShieldMetrics("hostname")
			Error("Invalid Hostname " + r.Host + " for request. Expecting one of " + fmt.Sprintf("%v", hostnames), nil)
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Bad Request: Invalid hostname. Use your domain instead of your IP to access your server. Check logs if more details are needed.", http.StatusBadRequest)
			
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip != "" {
				TriggerEvent(
					"cosmos.proxy.shield.hostname",
					"Proxy Shield hostname blocked",
					"warning",
					"",
					map[string]interface{}{
					"clientID": ip,
					"hostname": r.Host,
					"url": r.URL.String(),
				})
				IncrementIPAbuseCounter(ip)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}

func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !HasPermission(r, PERM_ADMIN_READ) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func EnsureHostnameCosmosAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		og := GetMainConfig().HTTPConfig.Hostname
		ni := GetMainConfig().NewInstall

		isLogin := !strings.HasPrefix(r.URL.Path, "/cosmos/api") ||
						   strings.HasPrefix(r.URL.Path, "/cosmos/api/login") ||
						   strings.HasPrefix(r.URL.Path, "/cosmos/api/status") ||
							 strings.HasPrefix(r.URL.Path, "/cosmos/api/password-reset") ||
							 strings.HasPrefix(r.URL.Path, "/cosmos/api/mfa") ||
							 strings.HasPrefix(r.URL.Path, "/cosmos/api/can-send-email") ||
							 strings.HasPrefix(r.URL.Path, "/cosmos/api/me")

		if ni || og == "0.0.0.0" || isLogin {
			next.ServeHTTP(w, r)
			return
		}
		
		reqHostNoPort := strings.Split(r.Host, ":")[0]
		
		if(GetMainConfig().HTTPConfig.AllowHTTPLocalIPAccess) {
			if(IsLocalIP(reqHostNoPort)) {
				next.ServeHTTP(w, r)
				return
			}
		}

		if og != reqHostNoPort {
			PushShieldMetrics("hostname")
			Error("Invalid Hostname " + r.Host + " for API request to " + r.URL.Path, nil)
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Bad Request: Invalid hostname. Use your domain instead of your IP to access your server. Check logs if more details are needed.", http.StatusBadRequest)
			
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip != "" {
				TriggerEvent(
					"cosmos.proxy.shield.hostname",
					"Proxy Shield hostname blocked",
					"warning",
					"",
					map[string]interface{}{
					"clientID": ip,
					"hostname": r.Host,
					"url": r.URL.String(),
				})
				IncrementIPAbuseCounter(ip)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}

func IsValidHostname(hostname string) bool {
	og := GetMainConfig().HTTPConfig.Hostname
	ni := GetMainConfig().NewInstall

	if ni || og == "0.0.0.0" || GetMainConfig().HTTPConfig.AcceptAllInsecureHostname {
		return true
	}


	hostnames := GetAllHostnames(false, false)

	reqHostNoPort := strings.Split(hostname, ":")[0]
	
	isOk := false

	for _, hostname := range hostnames {
		hostnameNoPort := strings.Split(hostname, ":")[0]
		if reqHostNoPort == hostnameNoPort {
			isOk = true
		}
	}
	
	if(GetMainConfig().HTTPConfig.AllowHTTPLocalIPAccess) {
		if(IsLocalIP(reqHostNoPort)) {
			isOk = true
		}
	}
		
	return isOk
}

func IPInRange(ipStr, cidrStr string) (bool, error) {
	_, cidrNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		// If not a CIDR range, try exact IP match
		return ipStr == cidrStr, nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("parse IP: invalid IP address")
	}

	return cidrNet.Contains(ip), nil
}

// CheckIPAccess reports whether a peer passes the constellation restriction and inbound whitelist. Only a Nebula device IP satisfies restrictToConstellation.
func CheckIPAccess(clientIP string, remoteAddr string, restrictToConstellation bool, whitelistIPs []string) bool {
	return checkIPAccess(clientIP, remoteAddr, restrictToConstellation, whitelistIPs, false)
}

// CheckRouteIPAccess additionally lets a local peer (IsLocalPeer) satisfy restrictToConstellation; the whitelist is never relaxed.
func CheckRouteIPAccess(clientIP string, remoteAddr string, restrictToConstellation bool, whitelistIPs []string) bool {
	return checkIPAccess(clientIP, remoteAddr, restrictToConstellation, whitelistIPs, true)
}

func checkIPAccess(clientIP string, remoteAddr string, restrictToConstellation bool, whitelistIPs []string, allowLocalPeer bool) bool {
	isUsingWhiteList := len(whitelistIPs) > 0
	isInWhitelist := false
	isInConstellation := IsConstellationIP(remoteAddr)
	isLocalPeer := allowLocalPeer && IsLocalPeer(remoteAddr)

	for _, ipRange := range whitelistIPs {
		if strings.Contains(ipRange, "/") {
			if ok, _ := IPInRange(clientIP, ipRange); ok {
				isInWhitelist = true
				break
			}
		} else if clientIP == ipRange {
			isInWhitelist = true
			break
		}
	}

	if restrictToConstellation {
		return isInConstellation || isInWhitelist || isLocalPeer
	}
	if isUsingWhiteList {
		return isInWhitelist
	}
	return true
}

func Restrictions(RestrictToConstellation bool, WhitelistInboundIPs []string) func(next http.Handler) http.Handler {
	return restrictions(RestrictToConstellation, WhitelistInboundIPs, false)
}

// RestrictionsAllowLocalHop is Restrictions that lets a loopback peer pass unchecked: on host-wildcard ":<port>" routes it is the socket-proxy listener's already-vetted internal hop, not a client.
func RestrictionsAllowLocalHop(RestrictToConstellation bool, WhitelistInboundIPs []string) func(next http.Handler) http.Handler {
	return restrictions(RestrictToConstellation, WhitelistInboundIPs, true)
}

func restrictions(RestrictToConstellation bool, WhitelistInboundIPs []string, allowLocalHop bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, ok := r.Context().Value("ClientID").(string)
		if !ok || ip == "" {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)

		if allowLocalHop && (remoteAddr == "127.0.0.1" || remoteAddr == "::1") {
			next.ServeHTTP(w, r)
			return
		}

		if !CheckRouteIPAccess(ip, remoteAddr, RestrictToConstellation, WhitelistInboundIPs) {
			PushShieldMetrics("ip-whitelists")

			TriggerEvent(
				"cosmos.proxy.shield.whitelist",
				"Proxy Shield IP blocked by whitelist",
				"warning",
				"",
				map[string]interface{}{
				"clientID": ip,
				"hostname": r.Host,
				"url": r.URL.String(),
			})

			IncrementIPAbuseCounter(ip)
			Error("Request from " + ip + " is blocked because of restrictions", nil)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
		})
	}
}

func ContentTypeMiddleware(contentType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

func SPAHandler(targetFolder string) http.Handler {
	// pwd,_ := os.Getwd()
	fs := http.FileServer(http.Dir(targetFolder))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Debug("Serving SPA from " + targetFolder + r.URL.Path)
		// if file does not exist or is a directory, serve index.html
		if stat, err := os.Stat(targetFolder + r.URL.Path); os.IsNotExist(err) || stat.IsDir() {
			Debug("Serving SPA index.html")
			http.ServeFile(w, r, targetFolder + "/index.html")
		} else {
			Debug("Serving SPA from " + targetFolder + r.URL.Path)
			fs.ServeHTTP(w, r)
		}
	})
}