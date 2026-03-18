package proxy

import (
	"net/http"
	"net/http/httputil" 
	"net/url"
	"strings"
	"crypto/tls"
	"os"
	"io/ioutil"
	"strconv"
	"time"
	"context"
	"net"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/constellation"

	"golang.org/x/net/http2"
)


func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}


// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string, AcceptInsecureHTTPSTarget bool, DisableHeaderHardening bool, route utils.ProxyRouteConfig) (*httputil.ReverseProxy, error) {
	var transport http.RoundTripper
	var targetURL *url.URL
	var err error
		
	targetURL, err = url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 5 * time.Second,
	}

	if utils.GetMainConfig().ConstellationConfig.Enabled {
			dialer.Resolver = &net.Resolver{                                            
				PreferGo: true,                                                         
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {    
					currConIp, err := constellation.GetCurrentDeviceIP()
					if err == nil {
						// Try Constellation DNS first
						conn, err := net.Dial(network, currConIp+":53")
						if err == nil {
							return conn, nil
						}
					} 

					// Fallback to system DNS
					return net.Dial(network, address)
				},
			}
			
	}

	if route.UseH2C {
		transport = &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return dialer.DialContext(ctx, network, addr)
			},
		}
	} else {
		customTransport := &http.Transport{
			DialContext: dialer.DialContext,
		}
		if AcceptInsecureHTTPSTarget {
			customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		transport = customTransport
	}

	proxy := &httputil.ReverseProxy{
		Transport:     transport,
		FlushInterval: -1,
	}

	proxy.Director = func(req *http.Request) {
		originalScheme := "http"
		if utils.IsHTTPS {
			originalScheme = "https"
		}
		
		urlQuery := targetURL.RawQuery
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		
		if route.Mode == "SERVAPP" && (!utils.IsInsideContainer || utils.IsHostNetwork) {
			targetHost := targetURL.Hostname()

			targetIP, err := docker.GetContainerIPByName(targetHost)
			if err != nil {
				utils.Error("Director Route", err)
			}
			utils.Debug("Dockerless Target IP: " + targetIP)
			req.URL.Host = targetIP + ":" + targetURL.Port()
		}


		req.URL.Path, req.URL.RawPath = joinURLPath(targetURL, req.URL)


		if urlQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = urlQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = urlQuery + "&" + req.URL.RawQuery
		}
		
		utils.Debug("Request to backend: " + req.URL.String())
		
		req.Header.Set("X-Forwarded-Proto", originalScheme)
		
		if(originalScheme == "https") {
			req.Header.Set("X-Forwarded-Ssl", "on")
		}

		req.Header.Del("X-Forwarded-Port")
		req.Header.Del("X-Forwarded-Host")

		hostname := utils.GetMainConfig().HTTPConfig.Hostname
		if route.Host != "" && route.UseHost {
			hostname = route.Host
		}

		// if route.UsePathPrefix {
		// 	hostname = hostname + route.PathPrefix
		// }

		hostDest := hostname
		hostPort := ""
		if route.OverwriteHostHeader != "" {
			hostDest = route.OverwriteHostHeader
			req.Host = hostDest
		}

		// split port
		hostDestSplit := strings.Split(hostDest, ":")
		hostDestNoPort := hostDest
		if len(hostDestSplit) > 1 {
			hostDestNoPort = hostDestSplit[0]
			hostPort = hostDestSplit[1]
		}

		// req.Header.Set("Host", hostDest)
		// req.Host = hostDest

		req.Header.Set("X-Forwarded-Host", hostDestNoPort)
		
		if hostPort != "" {
			req.Header.Set("X-Forwarded-Port", hostPort)
		}

		// spoof hostname
		if route.SpoofHostname {
			req.Header.Del("X-Forwarded-Port")
			req.Header.Del("X-Forwarded-Host")
			req.Header.Del("host")
			
			req.Header.Set("Host", req.URL.Host)
			req.Host = req.URL.Host
		}

		// Extra headers (applied last so they can overwrite anything)
		if len(route.ExtraHeaders) > 0 {
			clientIP := GetClientID(req, route)

			replacer := strings.NewReplacer(
				"$target", route.Target,
				"$scheme", originalScheme,
				"$protocol", originalScheme,
				"$host", hostname,
				"$origin", hostname,
				"$clientIP", clientIP,
				"$user", req.Header.Get("x-cosmos-user"),
				"$route", route.Name,
				"$path", req.URL.Path,
			)

			for name, value := range route.ExtraHeaders {
				req.Header.Del(name)
				req.Header.Set(name, replacer.Replace(value))
			}
		}

		// Duplicate X-* headers as HTTP_* equivalents (legacy CGI convention)
		if !route.DisableLegacyHTTPHeaders {
			for name, values := range req.Header {
				if strings.HasPrefix(name, "X-") {
					httpName := "HTTP_" + strings.ReplaceAll(strings.ToUpper(name), "-", "_")
					for _, v := range values {
						req.Header.Set(httpName, v)
					}
				}
			}
		}
	}

	proxy.Transport = transport

	proxy.ModifyResponse = func(resp *http.Response) error {
		utils.Debug("Response from backend: " + resp.Status)
		utils.Debug("URL was " + resp.Request.URL.String())
		
		if !DisableHeaderHardening {
			resp.Header.Del("Access-Control-Allow-Origin")
			resp.Header.Del("Access-Control-Allow-Credentials")
			resp.Header.Del("Strict-Transport-Security")
			resp.Header.Del("X-Content-Type-Options")
			resp.Header.Del("Content-Security-Policy")
			resp.Header.Del("X-XSS-Protection")
		}
		
		// if 502
		if resp.StatusCode == 502 {
			// set body
			rb := "502 Bad Gateway. This means your container / backend is not reachable by Cosmos."
			resp.Body = ioutil.NopCloser(strings.NewReader(rb))
			resp.Header.Set("Content-Length", strconv.Itoa(len(rb)))
		}

		return nil
	}

	return proxy, nil
}


func TunnelRouteTo(tunnel utils.ConstellationTunnel, lb *TunnelLoadBalancer) http.Handler {
	type targetProxy struct {
		target  utils.TunnelTarget
		handler http.Handler
	}

	// For self-tunnel, use the original route handler to avoid proxy loop
	currentDeviceName, er12 := constellation.GetCurrentDeviceName()
	if er12 != nil {
		utils.Error("Get Current Device Name for Tunnel Route", er12)
	}

	proxies := make([]targetProxy, 0, len(tunnel.Targets))
	for _, t := range tunnel.Targets {
		if t.DeviceName == currentDeviceName {
			// Find the original config route and use RouteTo directly
			for _, cfgRoute := range utils.GetMainConfig().HTTPConfig.ProxyConfig.Routes {
				if cfgRoute.Name == tunnel.Route.Name {
					proxies = append(proxies, targetProxy{t, RouteTo(cfgRoute)})
					break
				}
			}
		} else {
			route := tunnel.Route
			route.Target = t.TargetURL
			proxy, err := NewProxy(t.TargetURL, route.AcceptInsecureHTTPSTarget, route.DisableHeaderHardening, route)
			if err != nil {
				utils.Error("Create Tunnel Route for "+t.DeviceName, err)
				continue
			}
			proxies = append(proxies, targetProxy{t, proxy})
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(proxies) == 0 {
			http.Error(w, "No tunnel targets available", http.StatusBadGateway)
			return
		}

		targets := make([]utils.TunnelTarget, len(proxies))
		for i, p := range proxies {
			targets[i] = p.target
		}

		sticky := tunnel.Route.LBStickyMode
		stickyKey := ""
		if sticky {
			stickyKey = GetClientID(r, tunnel.Route)
		}

		// For HTTP, check cookie first (survives IP changes on mobile)
		var selected *utils.TunnelTarget
		if sticky {
			if cookie, err := r.Cookie("_cosmos_tunnel_lb"); err == nil && cookie.Value != "" {
				for i, p := range proxies {
					if p.target.DeviceName == cookie.Value {
						selected = &proxies[i].target
						break
					}
				}
				// Cookie device gone — fall through to SelectTarget
			}
		}

		if selected == nil {
			selected = lb.SelectTarget(targets, tunnel.Route.Name, tunnel.Route.LBMode, sticky, stickyKey)
		}
		if selected == nil {
			http.Error(w, "No tunnel targets available", http.StatusBadGateway)
			return
		}

		if sticky {
			http.SetCookie(w, &http.Cookie{
				Name:     "_cosmos_tunnel_lb",
				Value:    selected.DeviceName,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   3600,
			})
		}

		for _, p := range proxies {
			if p.target.DeviceName == selected.DeviceName {
				p.handler.ServeHTTP(w, r)
				return
			}
		}
	})
}

func RouteTo(route utils.ProxyRouteConfig) http.Handler {
	// initialize a reverse proxy and pass the actual backend server url here

	destination := route.Target
	routeType := route.Mode

	if (routeType == "STATIC" || routeType == "SPA") && utils.IsInsideContainer {
		if _, err := os.Stat("/mnt/host"); err == nil {
			destination = "/mnt/host" + destination
		}
	}

	if(routeType == "SERVAPP" || routeType == "PROXY") {
		proxy, err := NewProxy(destination, route.AcceptInsecureHTTPSTarget, route.DisableHeaderHardening, route)
		if err != nil {
				utils.Error("Create Route", err)
		}

		// create a handler function which uses the reverse proxy
		return proxy
	}  else if (routeType == "STATIC") {
		return http.FileServer(http.Dir(destination))
	}  else if (routeType == "SPA") {
		return utils.SPAHandler(destination)
	} else if(routeType == "REDIRECT") {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, destination, 302)
		})
	} else {
		utils.Error("Invalid route type", nil)
		return nil
	}
}