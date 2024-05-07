package proxy

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils"
)

func BuildFromConfig(router *mux.Router, config utils.ProxyConfig) *mux.Router {

	router.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	ConstellationConfig := utils.GetMainConfig().ConstellationConfig
	
	// if constellation slave
	if ConstellationConfig.Enabled && ConstellationConfig.SlaveMode {
		for i := len(ConstellationConfig.Tunnels)-1; i >= 0; i-- {
			routeConfig := ConstellationConfig.Tunnels[i]
			if !routeConfig.Disabled {
				RouterGen(routeConfig, router, RouteTo(routeConfig))
			}
		}
	}

	for i := len(config.Routes)-1; i >= 0; i-- {
		routeConfig := config.Routes[i]
		if !routeConfig.Disabled {
			RouterGen(routeConfig, router, RouteTo(routeConfig))
		}
	}
	
	return router
}