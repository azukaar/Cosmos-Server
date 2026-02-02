package proxy

import (
	"net/http"
	"regexp"
	"strconv"
	"time"
	"net/url"

	"github.com/azukaar/cosmos-server/src/user"
	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/go-chi/httprate"
	"github.com/gorilla/mux"
)

func tokenMiddleware(route utils.ProxyRouteConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			enabled := route.AuthEnabled
			adminOnly := route.AdminOnly

			// bypass auth if from Constellation tunnel
			if ((enabled && r.Header.Get("x-cosmos-user") != "") || !enabled) {
				remoteAddr, _ := utils.SplitIP(r.RemoteAddr)
				
				isConstIP := utils.IsConstellationIP(remoteAddr)
				isConstTokenValid := constellation.CheckConstellationToken(r) == nil

				if isConstIP && isConstTokenValid {
					utils.Debug("Bypassing auth for Constellation tunnel")
					r.Header.Del("x-cstln-auth")

					next.ServeHTTP(w, r)
					return
				}
			}
			
			r.Header.Del("x-cosmos-user")
			r.Header.Del("x-cosmos-role")
			r.Header.Del("x-cosmos-user-role")
			r.Header.Del("x-cosmos-mfa")
			r.Header.Del("x-cstln-auth")

			role, u, err := user.RefreshUserToken(w, r)

			if err != nil {
				return
			}

			r.Header.Set("x-cosmos-user", u.Nickname)
			r.Header.Set("x-cosmos-role", strconv.Itoa((int)(role)))
			r.Header.Set("x-cosmos-user-role", strconv.Itoa((int)(u.Role)))
			r.Header.Set("x-cosmos-mfa", strconv.Itoa((int)(u.MFAState)))

			ogcookies := r.Header.Get("Cookie")
			cookieRemoveRegex := regexp.MustCompile(`\s?jwttoken=[^;]*;?\s?`)
			cookies := cookieRemoveRegex.ReplaceAllString(ogcookies, "")
			r.Header.Set("Cookie", cookies)

			// Replace the token with a application speicfic one
			//r.Header.Set("x-cosmos-token", "1234567890")

			if enabled && adminOnly {
				if errT := AdminOnlyWithRedirect(w, r, route); errT != nil {
					return
				}
			} else if enabled {
				if errT := LoggedInOnlyWithRedirect(w, r, route); errT != nil {
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AddConstellationToken(route utils.ProxyRouteConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If the request is from a Constellation tunnel, add the token
			apiKey, _ := constellation.GetCurrentDeviceAPIKey()
			if route._IsTunneled {
				// Add the token
				r.Header.Set("x-cstln-auth", apiKey)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RouterGen(route utils.ProxyRouteConfig, router *mux.Router, destination http.Handler) *mux.Route {
	origin := router.NewRoute()

	if route.UseHost {
		origin = origin.Host(route.Host)

		if route.Mode == "SERVAPP" || route.Mode == "PROXY" || route.Mode == "REDIRECT" {
			// if Scheme is not http/https, discard
			urlRoute, err := url.Parse(route.Target)
			if err != nil {
				utils.Error("Invalid target URL: "+route.Target, err)
				return nil
			}

			if urlRoute.Scheme != "http" && urlRoute.Scheme != "https" {
				return nil
			}
		}
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

	destination = AddConstellationToken(route)(destination)

	for filter := range route.AddionalFilters {
		if route.AddionalFilters[filter].Type == "header" {
			origin = origin.Headers(route.AddionalFilters[filter].Name, route.AddionalFilters[filter].Value)
		} else if route.AddionalFilters[filter].Type == "query" {
			origin = origin.Queries(route.AddionalFilters[filter].Name, route.AddionalFilters[filter].Value)
		} else if route.AddionalFilters[filter].Type == "method" {
			origin = origin.Methods(route.AddionalFilters[filter].Value)
		} else {
			utils.Error("Unknown filter type: "+route.AddionalFilters[filter].Type, nil)
		}
	}
	
	destination = utils.Restrictions(route.RestrictToConstellation, route.WhitelistInboundIPs)(destination)
	
	if route.BlockCommonBots {
		destination = BotDetectionMiddleware(destination)
	}

	if route.BlockAPIAbuse {
		destination = utils.BlockPostWithoutReferer(destination)
	}

	destination = SmartShieldMiddleware(route.Name, route)(destination)

	originCORS := route.CORSOrigin

	if originCORS == "" && !route.DisableHeaderHardening {
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

	
	if(!utils.GetMainConfig().HTTPConfig.AcceptAllInsecureHostname && route.UseHost) {
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

	if !route.DisableHeaderHardening {
		destination = utils.SetSecurityHeaders(utils.CORSHeader(originCORS)(destination))
	}

	destination = tokenMiddleware(route)(utils.SetCosmosHeader(destination))

	origin.Handler(destination)

	utils.Log("Added route: [" + (string)(route.Mode) + "] " + route.Host + route.PathPrefix + " to " + route.Target + "")

	return origin
}
