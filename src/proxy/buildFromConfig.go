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
	
	// Proxy RClone
	if utils.ProxyRClone {
		data64Auth := utils.Base64Encode(utils.ProxyRCloneUser + ":" + utils.ProxyRClonePwd)
		rcloneRoute := utils.ProxyRouteConfig{
			Name: "RClone",
			Mode: "PROXY",
			UseHost: false,
			Target: "http://localhost:5573",
			UsePathPrefix: true,
			PathPrefix: "/cosmos/rclone",
			AuthEnabled: true,
			AdminOnly: true,
			ExtraHeaders: map[string]string{
				"Authorization": "Basic " + data64Auth,
			},
		}
		RouterGen(rcloneRoute, router, RouteTo(rcloneRoute))
	}

	ConstellationConfig := utils.GetMainConfig().ConstellationConfig
	
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