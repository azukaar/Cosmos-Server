package proxy

import (
	"fmt"
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
	utils.Log("[buildFromConfig][tunnel] Found " + strings.Join([]string{fmt.Sprintf("%d", len(remoteTunnels))}, "") + " remote tunnels")
	for _, tunnel := range remoteTunnels {
		utils.Debug("[buildFromConfig][tunnel] Tunnel route: Name=" + tunnel.Route.Name + " Host=" + tunnel.Route.Host + " UseHost=" + fmt.Sprintf("%v", tunnel.Route.UseHost) + " Target=" + tunnel.Route.Target + " Mode=" + string(tunnel.Route.Mode) + " Disabled=" + fmt.Sprintf("%v", tunnel.Route.Disabled) + " UsePathPrefix=" + fmt.Sprintf("%v", tunnel.Route.UsePathPrefix) + " PathPrefix=" + tunnel.Route.PathPrefix)
		if !tunnel.Route.Disabled && ((strings.HasPrefix(tunnel.Route.Target, "http://") || strings.HasPrefix(tunnel.Route.Target, "https://")) ||
			(tunnel.Route.Mode == "STATIC" || tunnel.Route.Mode == "SPA")) {
			utils.Log("[buildFromConfig][tunnel] Registering tunnel route: " + tunnel.Route.Host + tunnel.Route.PathPrefix + " -> " + tunnel.Route.Target)
			RouterGen(tunnel.Route, router, TunnelRouteTo(tunnel, DefaultTunnelLB))
		} else {
			utils.Debug("[buildFromConfig][tunnel] Skipping tunnel route: Name=" + tunnel.Route.Name + " (Disabled=" + fmt.Sprintf("%v", tunnel.Route.Disabled) + " Target=" + tunnel.Route.Target + " Mode=" + string(tunnel.Route.Mode) + ")")
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