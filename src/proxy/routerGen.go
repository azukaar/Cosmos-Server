package proxy

import (
	"net/http"
	"net/http/httputil"  
	"github.com/gorilla/mux"
	"time"
	"github.com/azukaar/cosmos-server/src/utils" 
	"github.com/azukaar/cosmos-server/src/user"
	"strconv"
	"github.com/go-chi/httprate"
	"regexp"
)

func tokenMiddleware(enabled bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("x-cosmos-user", "")
			r.Header.Set("x-cosmos-role", "")
	
			u, err := user.RefreshUserToken(w, r)
	
			if err != nil {
				return
			}
			
			r.Header.Set("x-cosmos-user", u.Nickname)
			r.Header.Set("x-cosmos-role", strconv.Itoa((int)(u.Role)))
			
			ogcookies := r.Header.Get("Cookie")
			cookieRemoveRegex := regexp.MustCompile(`jwttoken=[^;]*;`)
			cookies := cookieRemoveRegex.ReplaceAllString(ogcookies, "")
			r.Header.Set("Cookie", cookies)

			// Replace the token with a application speicfic one
			r.Header.Set("x-cosmos-token", "1234567890")

			if(enabled) {
				utils.LoggedInOnlyWithRedirect(w, r);
			}

			next.ServeHTTP(w, r)
		})
	}
}

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
		tokenMiddleware(route.AuthEnabled)(
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
		)(realDestination)))))

	return origin
}