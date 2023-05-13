package configapi

import (
	"encoding/json"
	"net/http"

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

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	var updateReq UpdateRouteRequest
	err := json.NewDecoder(req.Body).Decode(&updateReq)
	if err != nil {
		utils.Error("SettingsUpdate: Invalid Update Request", err)
		utils.HTTPError(w, "Invalid Update Request", http.StatusBadRequest, "UR001")
		return
	}

	config := utils.ReadConfigFromFile()
	routes := config.HTTPConfig.ProxyConfig.Routes
	routeIndex := -1

	if updateReq.Operation != "add" {
		if updateReq.RouteName == "" {
			utils.Error("SettingsUpdate: RouteName must be provided", nil)
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
			utils.Error("SettingsUpdate: Route not found: "+updateReq.RouteName, nil)
			utils.HTTPError(w, "Route not found", http.StatusNotFound, "UR002")
			return
		}
	}

	switch updateReq.Operation {
		case "replace":
			if updateReq.NewRoute == nil {
				utils.Error("SettingsUpdate: NewRoute must be provided for replace operation", nil)
				utils.HTTPError(w, "NewRoute must be provided for replace operation", http.StatusBadRequest, "UR003")
				return
			}
			routes[routeIndex] = *updateReq.NewRoute
		case "move_up":
			if routeIndex > 0 {
				routes[routeIndex-1], routes[routeIndex] = routes[routeIndex], routes[routeIndex-1]
			}
		case "move_down":
			if routeIndex < len(routes)-1 {
				routes[routeIndex+1], routes[routeIndex] = routes[routeIndex], routes[routeIndex+1]
			}
		case "delete":
			routes = append(routes[:routeIndex], routes[routeIndex+1:]...)
		case "add":
			if updateReq.NewRoute == nil {
				utils.Error("SettingsUpdate: NewRoute must be provided for add operation", nil)
				utils.HTTPError(w, "NewRoute must be provided for add operation", http.StatusBadRequest, "UR003")
				return
			}
			routes = append([]utils.ProxyRouteConfig{*updateReq.NewRoute}, routes...)
		default:
			utils.Error("SettingsUpdate: Unsupported operation: "+updateReq.Operation, nil)
			utils.HTTPError(w, "Unsupported operation", http.StatusBadRequest, "UR004")
			return
		}

	config.HTTPConfig.ProxyConfig.Routes = routes
	utils.SaveConfigTofile(config)
	utils.NeedsRestart = true

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}