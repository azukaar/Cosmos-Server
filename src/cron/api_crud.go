package cron

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
)

// CronConfigRoute godoc
// @Summary List or create CRON job configurations
// @Tags Cron
// @Accept json
// @Produce json
// @Param body body utils.CRONConfig false "CRON config (POST only)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/cron [get]
// @Router /api/cron [post]
func CronConfigRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		cronConfigList(w, req)
	} else if req.Method == "POST" {
		cronConfigCreate(w, req)
	} else {
		utils.Error("CronConfigRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

// CronConfigIdRoute godoc
// @Summary Get, update, or delete a CRON job configuration by name
// @Tags Cron
// @Accept json
// @Produce json
// @Param name path string true "CRON job name"
// @Param body body utils.CRONConfig false "Updated CRON config (PUT only)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/cron/{name} [get]
// @Router /api/cron/{name} [put]
// @Router /api/cron/{name} [delete]
func CronConfigIdRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		cronConfigGet(w, req)
	} else if req.Method == "PUT" {
		cronConfigUpdate(w, req)
	} else if req.Method == "DELETE" {
		cronConfigDelete(w, req)
	} else {
		utils.Error("CronConfigIdRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func cronConfigList(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
		return
	}

	config := utils.GetMainConfig()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   config.CRON,
	})
}

func cronConfigGet(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
		return
	}

	vars := mux.Vars(req)
	name := vars["name"]

	config := utils.GetMainConfig()

	cronJob, exists := config.CRON[name]
	if !exists {
		utils.HTTPError(w, "CRON job not found", http.StatusNotFound, "CR001")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   cronJob,
	})
}

func cronConfigCreate(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	var cronJob utils.CRONConfig
	if err := json.NewDecoder(req.Body).Decode(&cronJob); err != nil {
		utils.Error("CronConfigCreate: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "CR010")
		return
	}

	if cronJob.Name == "" {
		utils.HTTPError(w, "CRON job name is required", http.StatusBadRequest, "CR011")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.CRON == nil {
		config.CRON = make(map[string]utils.CRONConfig)
	}

	if _, exists := config.CRON[cronJob.Name]; exists {
		utils.HTTPError(w, "CRON job with this name already exists", http.StatusConflict, "CR012")
		return
	}

	config.CRON[cronJob.Name] = cronJob
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.cron.config.created",
		"CRON job config created",
		"important",
		"cron@"+cronJob.Name,
		map[string]interface{}{
			"name": cronJob.Name,
		},
	)

	go (func() {
		InitJobs()
		InitScheduler()
	})()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   cronJob,
	})
}

func cronConfigUpdate(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	vars := mux.Vars(req)
	name := vars["name"]

	var cronJob utils.CRONConfig
	if err := json.NewDecoder(req.Body).Decode(&cronJob); err != nil {
		utils.Error("CronConfigUpdate: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "CR020")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.CRON == nil {
		utils.HTTPError(w, "CRON job not found", http.StatusNotFound, "CR021")
		return
	}

	if _, exists := config.CRON[name]; !exists {
		utils.HTTPError(w, "CRON job not found", http.StatusNotFound, "CR021")
		return
	}

	// If the name changed, remove the old entry
	if cronJob.Name != "" && cronJob.Name != name {
		delete(config.CRON, name)
		config.CRON[cronJob.Name] = cronJob
	} else {
		cronJob.Name = name
		config.CRON[name] = cronJob
	}

	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.cron.config.updated",
		"CRON job config updated",
		"important",
		"cron@"+name,
		map[string]interface{}{
			"name": name,
		},
	)

	go (func() {
		InitJobs()
		InitScheduler()
	})()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   cronJob,
	})
}

func cronConfigDelete(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	vars := mux.Vars(req)
	name := vars["name"]

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.CRON == nil {
		utils.HTTPError(w, "CRON job not found", http.StatusNotFound, "CR030")
		return
	}

	if _, exists := config.CRON[name]; !exists {
		utils.HTTPError(w, "CRON job not found", http.StatusNotFound, "CR030")
		return
	}

	delete(config.CRON, name)
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.cron.config.deleted",
		"CRON job config deleted",
		"important",
		"cron@"+name,
		map[string]interface{}{
			"name": name,
		},
	)

	go (func() {
		InitJobs()
		InitScheduler()
	})()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}
