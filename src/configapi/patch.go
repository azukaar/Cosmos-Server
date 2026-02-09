package configapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
)

type UpdateRouteRequest struct {
	RouteName string `json:"routeName"`
	Operation string `json:"operation"`
	NewRoute  *utils.ProxyRouteConfig `json:"newRoute,omitempty"`
}

func ConfigApiPatch(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	utils.Log("RouteSettingsUpdate: Patching config")

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	var updateReq UpdateRouteRequest
	err := json.NewDecoder(req.Body).Decode(&updateReq)
	if err != nil {
		utils.Error("RouteSettingsUpdate: Invalid Update Request", err)
		utils.HTTPError(w, "Invalid Update Request", http.StatusBadRequest, "UR001")
		return
	}

	config := utils.ReadConfigFromFile()
	routes := config.HTTPConfig.ProxyConfig.Routes
	routeIndex := -1

	if updateReq.Operation != "add" {
		if updateReq.RouteName == "" {
			utils.Error("RouteSettingsUpdate: RouteName must be provided", nil)
			utils.HTTPError(w, "RouteName must be provided", http.StatusBadRequest, "UR002")
			return
		}
	
		for i, route := range routes {
			if route.Name == updateReq.RouteName {
				routeIndex = i
				break
			}
		}
	
		if routeIndex == -1 {
			utils.Error("RouteSettingsUpdate: Route not found: "+updateReq.RouteName, nil)
			utils.HTTPError(w, "Route not found", http.StatusNotFound, "UR002")
			return
		}
	}

	switch updateReq.Operation {
		case "replace":
			utils.Log("RouteSettingsUpdate: Replacing route: "+updateReq.RouteName)
			if updateReq.NewRoute == nil {
				utils.Error("RouteSettingsUpdate: NewRoute must be provided for replace operation", nil)
				utils.HTTPError(w, "NewRoute must be provided for replace operation", http.StatusBadRequest, "UR003")
				return
			}
			routes[routeIndex] = *updateReq.NewRoute
		case "move_up":
			utils.Log("RouteSettingsUpdate: Moving up route: "+updateReq.RouteName)
			if routeIndex > 0 {
				routes[routeIndex-1], routes[routeIndex] = routes[routeIndex], routes[routeIndex-1]
			}
		case "move_down":
			utils.Log("RouteSettingsUpdate: Moving down route: "+updateReq.RouteName)
			if routeIndex < len(routes)-1 {
				routes[routeIndex+1], routes[routeIndex] = routes[routeIndex], routes[routeIndex+1]
			}
		case "delete":
			utils.Log("RouteSettingsUpdate: Deleting route: "+updateReq.RouteName)
			routes = append(routes[:routeIndex], routes[routeIndex+1:]...)
		case "add":
			utils.Log("RouteSettingsUpdate: Adding route")
			if updateReq.NewRoute == nil {
				utils.Error("RouteSettingsUpdate: NewRoute must be provided for add operation", nil)
				utils.HTTPError(w, "NewRoute must be provided for add operation", http.StatusBadRequest, "UR003")
				return
			}
			routes = append([]utils.ProxyRouteConfig{*updateReq.NewRoute}, routes...)
		default:
			utils.Error("RouteSettingsUpdate: Unsupported operation: "+updateReq.Operation, nil)
			utils.HTTPError(w, "Unsupported operation", http.StatusBadRequest, "UR004")
			return
		}

	config.HTTPConfig.ProxyConfig.Routes = routes
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.settings",
		"Settings updated",
		"success",
		"",
		map[string]interface{}{
	})
	
	go (func () {
		utils.RestartHTTPServer()
	})()
	
	if updateReq.NewRoute != nil && updateReq.NewRoute.Mode == "SERVAPP" {
		utils.Log("RouteSettingsUpdate: Service needs update: "+updateReq.NewRoute.Target)

		target := updateReq.NewRoute.Target
		tokens := strings.Split(target, ":")
		name := tokens[1][2:]
		
		utils.Log("RouteSettingsUpdate: Service needs update: "+name)

		// TODO CACHE BURST IN IP RESOLUTION
		// utils.ReBootstrapContainer(name)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}