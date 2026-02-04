package proxy

import (
	"github.com/gorilla/mux"
	"net/http"

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
		routeConfig := tunnel.Route 
		if !routeConfig.Disabled {
			RouterGen(routeConfig, router, RouteTo(routeConfig))
		}
	}

	remoteConfigs := utils.GetMainConfig().RemoteStorage
	for _, shares := range remoteConfigs.Shares {
			route := shares.Route
			if route.Disabled {
				continue
			}
			RouterGen(route, router, RouteTo(route))
	}

	for i := len(config.Routes)-1; i >= 0; i-- {
		routeConfig := config.Routes[i]
		if !routeConfig.Disabled {
			RouterGen(routeConfig, router, RouteTo(routeConfig))
		}
	}
	
	return router
}