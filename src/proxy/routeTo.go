package proxy

import (
	"net/http"
	"net/http/httputil" 
	"net/url"
	spa "github.com/roberthodgen/spa-server"
	"github.com/azukaar/cosmos-server/src/utils"
)

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
			return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.ModifyResponse = func(resp *http.Response) error {
		utils.Debug("Response from backend: " + resp.Status)
		utils.Debug("URL was " + resp.Request.URL.String())

		return nil
	}

	return proxy, nil
}


func RouteTo(route utils.ProxyRouteConfig) http.Handler {
	// initialize a reverse proxy and pass the actual backend server url here

	destination := route.Target
	routeType := route.Mode

	if(routeType == "SERVAPP" || routeType == "PROXY") {
		proxy, err := NewProxy(destination)
		if err != nil {
				utils.Error("Create Route", err)
		}

		// create a handler function which uses the reverse proxy
		return proxy
	}  else if (routeType == "STATIC") {
		return http.FileServer(http.Dir(destination))
	}  else if (routeType == "SPA") {
		return spa.SpaHandler(destination, "index.html")	
	} else if(routeType == "REDIRECT") {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, destination, 302)
		})
	} else {
		utils.Error("Invalid route type", nil)
		return nil
	}
}