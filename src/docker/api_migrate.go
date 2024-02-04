package docker

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
)

type migrateToHostMode struct {
	HTTPPort string `json:"http"`
	HTTPSPort string `json:"https"`
}

func MigrateToHostModeRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
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

		SelfAction("host")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("MigrateToHostMode: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}