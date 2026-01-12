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

	// if strings.HasPrefix(targetHost, "http://unix://") {
	// 	// Unix socket handling
	// 	socketPath := strings.TrimPrefix(targetHost, "http://unix://")
	// 	transport = &http.Transport{
	// 		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
	// 			return net.Dial("unix", socketPath)
	// 		},
	// 	}
	// 	// Use a dummy URL for the director
	// 	targetURL, _ = url.Parse("http://unix-socket")
	// } else {
		// Regular HTTP/HTTPS handling
		targetURL, err = url.Parse(targetHost)
		if err != nil {
			return nil, err
		}

		customTransport := &http.Transport{}

		if utils.GetMainConfig().ConstellationConfig.Enabled && utils.GetMainConfig().ConstellationConfig.SlaveMode {
			customTransport = &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 5 * time.Second,
					Resolver: &net.Resolver{
						PreferGo: true,
						Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
							return net.Dial(network, "192.168.201.1:53")
						},
					},
				}).DialContext,
			}
		}

		if AcceptInsecureHTTPSTarget {
			customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

		transport = customTransport
	// }

	proxy := &httputil.ReverseProxy{
		Transport: transport,
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
				utils.Error("Create Route", err)
			}
			utils.Debug("Dockerless Target IP: " + targetIP)
			req.URL.Host = targetIP + ":" + targetURL.Port()
		}

		utils.Debug("Request to backend: " + req.URL.String())

		req.URL.Path, req.URL.RawPath = joinURLPath(targetURL, req.URL)
		if urlQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = urlQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = urlQuery + "&" + req.URL.RawQuery
		}
		
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

		// Extra headers
		for name, value := range route.ExtraHeaders {
			req.Header.Del(name)
			req.Header.Set(name, value)
		}
	}

	proxy.Transport = transport

	proxy.ModifyResponse = func(resp *http.Response) error {
		utils.Debug("Response from backend: " + resp.Status)
		utils.Debug("URL was " + resp.Request.URL.String())

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