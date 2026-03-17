package metrics

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
)

// AlertsRoute godoc
// @Summary List or create monitoring alerts
// @Tags Metrics
// @Accept json
// @Produce json
// @Param body body utils.Alert false "Alert config (POST only)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/alerts [get]
// @Router /api/alerts [post]
func AlertsRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		alertsList(w, req)
	} else if req.Method == "POST" {
		alertsCreate(w, req)
	} else {
		utils.Error("AlertsRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

// AlertsIdRoute godoc
// @Summary Get, update, or delete a monitoring alert by name
// @Tags Metrics
// @Accept json
// @Produce json
// @Param name path string true "Alert name"
// @Param body body utils.Alert false "Updated alert config (PUT only)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/alerts/{name} [get]
// @Router /api/alerts/{name} [put]
// @Router /api/alerts/{name} [delete]
func AlertsIdRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		alertsGet(w, req)
	} else if req.Method == "PUT" {
		alertsUpdate(w, req)
	} else if req.Method == "DELETE" {
		alertsDelete(w, req)
	} else {
		utils.Error("AlertsIdRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func alertsList(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	config := utils.GetMainConfig()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   config.MonitoringAlerts,
	})
}

func alertsGet(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	vars := mux.Vars(req)
	name := vars["name"]

	config := utils.GetMainConfig()

	alert, exists := config.MonitoringAlerts[name]
	if !exists {
		utils.HTTPError(w, "Alert not found", http.StatusNotFound, "AL001")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   alert,
	})
}

func alertsCreate(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	var alert utils.Alert
	if err := json.NewDecoder(req.Body).Decode(&alert); err != nil {
		utils.Error("AlertsCreate: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "AL010")
		return
	}

	if alert.Name == "" {
		utils.HTTPError(w, "Alert name is required", http.StatusBadRequest, "AL011")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.MonitoringAlerts == nil {
		config.MonitoringAlerts = make(map[string]utils.Alert)
	}

	if _, exists := config.MonitoringAlerts[alert.Name]; exists {
		utils.HTTPError(w, "Alert with this name already exists", http.StatusConflict, "AL012")
		return
	}

	config.MonitoringAlerts[alert.Name] = alert
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.alerts.created",
		"Alert created",
		"important",
		"alert@"+alert.Name,
		map[string]interface{}{
			"name": alert.Name,
		},
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   alert,
	})
}

func alertsUpdate(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	vars := mux.Vars(req)
	name := vars["name"]

	var alert utils.Alert
	if err := json.NewDecoder(req.Body).Decode(&alert); err != nil {
		utils.Error("AlertsUpdate: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "AL020")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.MonitoringAlerts == nil {
		utils.HTTPError(w, "Alert not found", http.StatusNotFound, "AL021")
		return
	}

	if _, exists := config.MonitoringAlerts[name]; !exists {
		utils.HTTPError(w, "Alert not found", http.StatusNotFound, "AL021")
		return
	}

	// If the name changed, remove the old entry
	if alert.Name != "" && alert.Name != name {
		delete(config.MonitoringAlerts, name)
		config.MonitoringAlerts[alert.Name] = alert
	} else {
		alert.Name = name
		config.MonitoringAlerts[name] = alert
	}

	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.alerts.updated",
		"Alert updated",
		"important",
		"alert@"+name,
		map[string]interface{}{
			"name": name,
		},
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   alert,
	})
}

func alertsDelete(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	vars := mux.Vars(req)
	name := vars["name"]

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.MonitoringAlerts == nil {
		utils.HTTPError(w, "Alert not found", http.StatusNotFound, "AL030")
		return
	}

	if _, exists := config.MonitoringAlerts[name]; !exists {
		utils.HTTPError(w, "Alert not found", http.StatusNotFound, "AL030")
		return
	}

	delete(config.MonitoringAlerts, name)
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.alerts.deleted",
		"Alert deleted",
		"important",
		"alert@"+name,
		map[string]interface{}{
			"name": name,
		},
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}
