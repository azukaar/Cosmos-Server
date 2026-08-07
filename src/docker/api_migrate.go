package docker

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
)

type migrateToHostMode struct {
	HTTPPort string `json:"http" validate:"required"`
	HTTPSPort string `json:"https" validate:"required"`
}

// MigrateToHostModeRoute godoc
// @Summary Migrate Cosmos container to host network mode
// @Tags docker
// @Accept json
// @Produce json
// @Param body body migrateToHostMode true "HTTP and HTTPS port configuration"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/migrate-host [post]
func MigrateToHostModeRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if req.Method == "POST" {
		errD := Connect()
		if errD != nil {
			utils.Error("MigrateToHostMode", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "CN001")
			return
		}

		var payload migrateToHostMode
		err := json.NewDecoder(req.Body).Decode(&payload)
		if err != nil {
			utils.Error("MigrateToHostMode: Error reading request body", err)
			utils.HTTPError(w, "Error reading request body: "+err.Error(), http.StatusBadRequest, "CN002")
			return
		}

		config := utils.ReadConfigFromFile()
		config.HTTPConfig.HTTPPort = payload.HTTPPort
		config.HTTPConfig.HTTPSPort = payload.HTTPSPort
		utils.SetBaseMainConfig(config)
		
		utils.TriggerEvent(
			"cosmos.settings",
			"Settings updated",
			"success",
			"",
			map[string]interface{}{
		})

		err = SelfAction("host")
		if err != nil {
			utils.Error("MigrateToHostMode: Error migrating to host mode", err)
			utils.HTTPError(w, "Error migrating to host mode: "+err.Error(), http.StatusInternalServerError, "CN003")
			return
		}
		
		utils.SetBaseMainConfig(config)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("MigrateToHostMode: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}