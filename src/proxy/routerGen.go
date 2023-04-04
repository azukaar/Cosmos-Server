package proxy

import (
	"net/http"
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

func RouterGen(route utils.ProxyRouteConfig, router *mux.Router, destination http.Handler) *mux.Route {
	origin := router.Methods("GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD")

	if(route.UseHost) {
		origin = origin.Host(route.Host)
	}

	if(route.UsePathPrefix) {
		if(route.PathPrefix != "" && route.PathPrefix[0] != '/') {
			utils.Error("PathPrefix must start with a /", nil)
		}
		origin = origin.PathPrefix(route.PathPrefix)
	}
	
	if(route.UsePathPrefix && route.StripPathPrefix) {
		if(route.PathPrefix != "" && route.PathPrefix[0] != '/') {
			utils.Error("PathPrefix must start with a /", nil)
		}
		destination = http.StripPrefix(route.PathPrefix, destination)
	}

	originCORS := route.CORSOrigin

	if originCORS == "" {
		if route.UseHost {
			originCORS = route.Host
		} else {
			originCORS = utils.GetMainConfig().HTTPConfig.Hostname
		}
	}

	if(route.UsePathPrefix && !route.StripPathPrefix && (route.Mode == "STATIC" || route.Mode == "SPA")) {
		utils.Warn("PathPrefix is used, but StripPathPrefix is false. The route mode is " + (string)(route.Mode) + ". This will likely cause issues with the route. Ignore this warning if you know what you are doing.")
	}
	
	timeout := route.Timeout
	
	if(timeout > 0) {
		destination = utils.MiddlewareTimeout(timeout * time.Millisecond)(destination)
	}

	throttlePerMinute := route.ThrottlePerMinute

	if(throttlePerMinute > 0) {
		throtthleTime := time.Minute
		destination = httprate.Limit(throttlePerMinute, throtthleTime, 
			httprate.WithKeyFuncs(httprate.KeyByIP),
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				utils.Error("Too many requests. Throttling", nil)
				utils.HTTPError(w, "Too many requests", 
					http.StatusTooManyRequests, "HTTP003")
				return 
			}),
		)(destination)
	}

	origin.Handler(tokenMiddleware(route.AuthEnabled)(utils.CORSHeader(originCORS)((destination))))

	utils.Log("Added route: ["+ (string)(route.Mode)  + "] " + route.Host + route.PathPrefix + " to " + route.Target + "")

	return origin
}