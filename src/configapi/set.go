package configapi

import (
	"net/http"
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/authorizationserver"
	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/cron"
	"github.com/azukaar/cosmos-server/src/storage"
	"github.com/azukaar/cosmos-server/src/backups"
)

func ConfigApiSet(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	} 

	if(req.Method == "PUT") {
		var request utils.Config
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("SettingsUpdate: Invalid User Request", err1)
			utils.HTTPError(w, "User Creation Error", 
				http.StatusInternalServerError, "UC001")
			return 
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("SettingsUpdate: Invalid User Request", errV)
			utils.HTTPError(w, "User Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "UC003")
			return 
		}

		// restore AuthPrivateKey and TLSKey
		config := utils.ReadConfigFromFile()
		request.HTTPConfig.AuthPrivateKey = config.HTTPConfig.AuthPrivateKey
		request.HTTPConfig.AuthPublicKey = config.HTTPConfig.AuthPublicKey
		request.HTTPConfig.TLSCert = config.HTTPConfig.TLSCert
		request.HTTPConfig.TLSKey = config.HTTPConfig.TLSKey
		request.NewInstall = config.NewInstall

		utils.SetBaseMainConfig(request)
		
		utils.TriggerEvent(
			"cosmos.settings",
			"Settings updated",
			"success",
			"",
			map[string]interface{}{
		})

		utils.InitFBL()
		utils.DisconnectDB()
		authorizationserver.Init()
		go (func() {
			storage.Restart()
			constellation.RestartNebula()
			utils.RestartHTTPServer()
			cron.InitJobs()
			cron.InitScheduler()
			backups.InitBackups()
		})()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingsUpdate: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
