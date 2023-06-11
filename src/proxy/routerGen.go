package proxy

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/azukaar/cosmos-server/src/user"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/go-chi/httprate"
	"github.com/gorilla/mux"
)

func tokenMiddleware(enabled bool, adminOnly bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del("x-cosmos-user")
			r.Header.Del("x-cosmos-role")
			r.Header.Del("x-cosmos-mfa")
			HTTPConfig := utils.GetMainConfig().HTTPConfig
			utils.Log("Config for " + r.Host + " is " + strconv.FormatBool(enabled))
			u, err := user.RefreshUserToken(w, r)

			if err != nil {
				return
			}

			r.Header.Set("x-cosmos-user", u.Nickname)
			r.Header.Set("x-cosmos-role", strconv.Itoa((int)(u.Role)))
			r.Header.Set("x-cosmos-mfa", strconv.Itoa((int)(u.MFAState)))

			ogcookies := r.Header.Get("Cookie")
			cookieRemoveRegex := regexp.MustCompile(`jwttoken=[^;]*;`)
			cookies := cookieRemoveRegex.ReplaceAllString(ogcookies, "")
			r.Header.Set("Cookie", cookies)

			// Replace the token with a application speicfic one
			r.Header.Set("x-cosmos-token", "1234567890")

			if enabled && adminOnly {
				if errT := utils.AdminOnlyWithRedirect(w, r); errT != nil {
					return
				}
			} else if enabled {
				u, err := user.RefreshAppAuthToken(w, r)
				if err == nil {
					utils.Log("cookie was retrieved and verified")
					//set auth and serve
					r.Header.Set("x-cosmos-user", u.Nickname)
					r.Header.Set("x-cosmos-role", strconv.Itoa((int)(u.Role)))
					r.Header.Set("x-cosmos-mfa", strconv.Itoa((int)(u.MFAState)))
				} else {
					rscheme := "https://"
					if r.TLS == nil {
						rscheme = "http://"
					}
					utils.Error("tokenMiddleware: User is not logged in for the app, redirected to main auth", err)
					http.Redirect(w, r, "https://"+HTTPConfig.Hostname+"/ui/login?app_auth=1&app_redir="+rscheme+r.Host+r.URL.Path, http.StatusFound)
					return
				}
			}
			utils.Debug("SERVING: " + HTTPConfig.Hostname + "/" + r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

func RouterGen(route utils.ProxyRouteConfig, router *mux.Router, destination http.Handler) *mux.Route {
	origin := router.Methods("GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD")

	if route.UseHost {
		origin = origin.Host(route.Host)
	}
	
	if route.UsePathPrefix {
		if route.PathPrefix != "" && route.PathPrefix[0] != '/' {
			utils.Error("PathPrefix must start with a /", nil)
		}
		origin = origin.PathPrefix(route.PathPrefix)
	}
	
	if route.UsePathPrefix && route.StripPathPrefix {
		if route.PathPrefix != "" && route.PathPrefix[0] != '/' {
			utils.Error("PathPrefix must start with a /", nil)
		}
		destination = http.StripPrefix(route.PathPrefix, destination)
	}
	
	destination = SmartShieldMiddleware(route.SmartShield)(destination)

	originCORS := route.CORSOrigin

	if originCORS == "" {
		if route.UseHost {
			originCORS = route.Host
		} else {
			originCORS = utils.GetMainConfig().HTTPConfig.Hostname
		}
	}

	if route.UsePathPrefix && !route.StripPathPrefix && (route.Mode == "STATIC" || route.Mode == "SPA") {
		utils.Warn("PathPrefix is used, but StripPathPrefix is false. The route mode is " + (string)(route.Mode) + ". This will likely cause issues with the route. Ignore this warning if you know what you are doing.")
	}

	timeout := route.Timeout

	
	if(!utils.GetMainConfig().HTTPConfig.AcceptAllInsecureHostname) {
		destination = utils.EnsureHostname(destination)
	}

	if timeout > 0 {
		destination = utils.MiddlewareTimeout(timeout * time.Millisecond)(destination)
	}

	throttlePerMinute := route.ThrottlePerMinute

	if throttlePerMinute > 0 {
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

	if route.MaxBandwith > 0 {
		destination = utils.BandwithLimiterMiddleware(route.MaxBandwith)(destination)
	}

	if route.BlockCommonBots {
		destination = BotDetectionMiddleware(destination)
	}

	if route.BlockAPIAbuse {
		destination = utils.BlockPostWithoutReferer(destination)
	}

	destination = tokenMiddleware(route.AuthEnabled, route.AdminOnly)(utils.CORSHeader(originCORS)((destination)))

	origin.Handler(destination)

	utils.Log("Added route: [" + (string)(route.Mode) + "] " + route.Host + route.PathPrefix + " to " + route.Target + "")

	return origin
}
