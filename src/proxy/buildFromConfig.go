package proxy

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/constellation"
)

func BuildFromConfig(router *mux.Router, config utils.ProxyConfig) *mux.Router {
	router.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	remoteTunnels := constellation.GetLocalTunnelCache()
	for _, tunnel := range remoteTunnels {
		if !tunnel.Route.Disabled && ((strings.HasPrefix(tunnel.Route.Target, "http://") || strings.HasPrefix(tunnel.Route.Target, "https://")) ||
			(tunnel.Route.Mode == "STATIC" || tunnel.Route.Mode == "SPA")) {
			RouterGen(tunnel.Route, router, TunnelRouteTo(tunnel, DefaultTunnelLB))
		}
	}

	remoteConfigs := utils.GetMainConfig().RemoteStorage
	for _, shares := range remoteConfigs.Shares {
			route := shares.Route
			if route.Disabled {
				continue
			}
			if !strings.HasPrefix(route.Target, "http://") && !strings.HasPrefix(route.Target, "https://") {
				continue
			}
			RouterGen(route, router, RouteTo(route))
	}

	for i := len(config.Routes)-1; i >= 0; i-- {
		routeConfig := config.Routes[i]
		if !routeConfig.Disabled && ((strings.HasPrefix(routeConfig.Target, "http://") || strings.HasPrefix(routeConfig.Target, "https://")) ||
			(routeConfig.Mode == "STATIC" || routeConfig.Mode == "SPA")) {
			RouterGen(routeConfig, router, RouteTo(routeConfig))
		}
	}
	
	return router
}