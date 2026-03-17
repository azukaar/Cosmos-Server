package configapi

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
)

func RoutesRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		listRoutes(w, req)
	} else if req.Method == "POST" {
		createRoute(w, req)
	} else {
		utils.Error("RoutesRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func RoutesIdRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		getRoute(w, req)
	} else if req.Method == "PUT" {
		updateRoute(w, req)
	} else if req.Method == "DELETE" {
		deleteRoute(w, req)
	} else {
		utils.Error("RoutesIdRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

// listRoutes godoc
// @Summary List all proxy routes
// @Description Returns all configured proxy routes
// @Tags routes
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]utils.ProxyRouteConfig}
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/routes [get]
func listRoutes(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	config := utils.ReadConfigFromFile()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   config.HTTPConfig.ProxyConfig.Routes,
	})
}

// getRoute godoc
// @Summary Get a proxy route by name
// @Description Returns a single proxy route configuration by its name
// @Tags routes
// @Produce json
// @Security BearerAuth
// @Param name path string true "Route name"
// @Success 200 {object} utils.APIResponse{data=utils.ProxyRouteConfig}
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/routes/{name} [get]
func getRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	name := mux.Vars(req)["name"]
	config := utils.ReadConfigFromFile()

	for _, route := range config.HTTPConfig.ProxyConfig.Routes {
		if route.Name == name {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data":   route,
			})
			return
		}
	}

	utils.HTTPError(w, "Route not found", http.StatusNotFound, "RT001")
}

// createRoute godoc
// @Summary Create a new proxy route
// @Description Creates a new proxy route configuration
// @Tags routes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body utils.ProxyRouteConfig true "Route configuration"
// @Success 200 {object} utils.APIResponse{data=utils.ProxyRouteConfig}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/routes [post]
func createRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	var newRoute utils.ProxyRouteConfig
	err := json.NewDecoder(req.Body).Decode(&newRoute)
	if err != nil {
		utils.Error("CreateRoute: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "RT002")
		return
	}

	if newRoute.Name == "" {
		utils.HTTPError(w, "Route name is required", http.StatusBadRequest, "RT003")
		return
	}

	if !newRoute.UseHost && !newRoute.UsePathPrefix {
		utils.HTTPError(w, "Route must have at least one of UseHost or UsePathPrefix enabled, otherwise it will catch all requests", http.StatusBadRequest, "RT008")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	routes := config.HTTPConfig.ProxyConfig.Routes

	for _, route := range routes {
		if route.Name == newRoute.Name {
			utils.HTTPError(w, "Route with this name already exists", http.StatusConflict, "RT004")
			return
		}
	}

	config.HTTPConfig.ProxyConfig.Routes = append([]utils.ProxyRouteConfig{newRoute}, routes...)
	utils.SetBaseMainConfig(config)

	utils.Log("Route created: " + newRoute.Name)

	utils.TriggerEvent(
		"cosmos.routes",
		"Route created: "+newRoute.Name,
		"success",
		"",
		map[string]interface{}{
	})

	go func() {
		utils.RestartHTTPServer()
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   newRoute,
	})
}

// updateRoute godoc
// @Summary Update a proxy route
// @Description Replaces an existing proxy route configuration by name
// @Tags routes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Route name"
// @Param request body utils.ProxyRouteConfig true "Updated route configuration"
// @Success 200 {object} utils.APIResponse{data=utils.ProxyRouteConfig}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/routes/{name} [put]
func updateRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	name := mux.Vars(req)["name"]

	var updatedRoute utils.ProxyRouteConfig
	err := json.NewDecoder(req.Body).Decode(&updatedRoute)
	if err != nil {
		utils.Error("UpdateRoute: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "RT005")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	routes := config.HTTPConfig.ProxyConfig.Routes
	routeIndex := -1

	for i, route := range routes {
		if route.Name == name {
			routeIndex = i
			break
		}
	}

	if routeIndex == -1 {
		utils.HTTPError(w, "Route not found", http.StatusNotFound, "RT006")
		return
	}

	if !updatedRoute.UseHost && !updatedRoute.UsePathPrefix {
		utils.HTTPError(w, "Route must have at least one of UseHost or UsePathPrefix enabled, otherwise it will catch all requests", http.StatusBadRequest, "RT008")
		return
	}

	routes[routeIndex] = updatedRoute
	config.HTTPConfig.ProxyConfig.Routes = routes
	utils.SetBaseMainConfig(config)

	utils.Log("Route updated: " + name)

	utils.TriggerEvent(
		"cosmos.routes",
		"Route updated: "+name,
		"success",
		"",
		map[string]interface{}{
	})

	go func() {
		utils.RestartHTTPServer()
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   updatedRoute,
	})
}

// deleteRoute godoc
// @Summary Delete a proxy route
// @Description Removes a proxy route configuration by name
// @Tags routes
// @Produce json
// @Security BearerAuth
// @Param name path string true "Route name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/routes/{name} [delete]
func deleteRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	name := mux.Vars(req)["name"]

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	routes := config.HTTPConfig.ProxyConfig.Routes
	routeIndex := -1

	for i, route := range routes {
		if route.Name == name {
			routeIndex = i
			break
		}
	}

	if routeIndex == -1 {
		utils.HTTPError(w, "Route not found", http.StatusNotFound, "RT007")
		return
	}

	routes = append(routes[:routeIndex], routes[routeIndex+1:]...)
	config.HTTPConfig.ProxyConfig.Routes = routes
	utils.SetBaseMainConfig(config)


	utils.Log("Route deleted: " + name)

	utils.TriggerEvent(
		"cosmos.routes",
		"Route deleted: "+name,
		"success",
		"",
		map[string]interface{}{
	})

	go func() {
		utils.RestartHTTPServer()
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}
