package proxy

import (
	"net/http"
	"net/http/httputil"  
	"github.com/gorilla/mux"
	"time"
	"../utils" 
	"github.com/go-chi/httprate"
)

func RouterGen(route utils.ProxyRouteConfig, router *mux.Router, destination *httputil.ReverseProxy) *mux.Route {
	var realDestination http.Handler
	realDestination = destination

	origin := router.Methods("GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD")

	if(route.UseHost) {
		origin = origin.Host(route.Host)
	}

	if(route.UsePathPrefix) {
		origin = origin.PathPrefix(route.PathPrefix)
	}
	
	if(route.UsePathPrefix && route.StripPathPrefix) {
		realDestination = http.StripPrefix(route.PathPrefix, destination)
	}
	timeout := route.Timeout
	
	if(timeout == 0) {
		timeout = 10000
	}

	throttlePerMinute := route.ThrottlePerMinute

	if(throttlePerMinute == 0) {
		throttlePerMinute = 60
	}

	originCORS := route.CORSOrigin

	if originCORS == "" {
		if route.UseHost {
			originCORS = route.Host
		} else {
			originCORS = utils.GetMainConfig().HTTPConfig.Hostname
		}
	}

	origin.Handler(
		utils.CORSHeader(originCORS)(
		utils.MiddlewareTimeout(timeout * time.Millisecond)(
		httprate.Limit(throttlePerMinute, 1*time.Minute, 
			httprate.WithKeyFuncs(httprate.KeyByIP),
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				utils.Error("Too many requests. Throttling", nil)
				utils.HTTPError(w, "Too many requests", 
					http.StatusTooManyRequests, "HTTP003")
				return 
			}),
		)(realDestination))))

	return origin
}