package configapi

import (
	"net/http"
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils"
)

func ConfigApiSet(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
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

		// restore fields that are never sent to the client or are masked with ***
		config := utils.ReadConfigFromFile()
		request.HTTPConfig.AuthPrivateKey = config.HTTPConfig.AuthPrivateKey
		request.HTTPConfig.AuthPublicKey = config.HTTPConfig.AuthPublicKey
		request.HTTPConfig.TLSCert = config.HTTPConfig.TLSCert
		request.HTTPConfig.TLSKey = config.HTTPConfig.TLSKey
		request.NewInstall = config.NewInstall

		// restore API token as we cannot edit it here
		request.APITokens = config.APITokens

		// restore DNS if user cannot read credentials, as they are sent as empty and we don't want to override them
		canReadCredentials := utils.HasPermission(req, utils.PERM_CREDENTIALS_READ)
		if !canReadCredentials {
			request.HTTPConfig.DNSChallengeConfig = config.HTTPConfig.DNSChallengeConfig
		}

		// restore credential fields if they were masked (sent as "***")
		if request.MongoDB == "***" {
			request.MongoDB = config.MongoDB
		}
		if request.EmailConfig.Password == "***" {
			request.EmailConfig.Password = config.EmailConfig.Password
		}
		if request.EmailConfig.Username == "***" {
			request.EmailConfig.Username = config.EmailConfig.Username
		}
		if request.EmailConfig.Host == "***" {
			request.EmailConfig.Host = config.EmailConfig.Host
		}
		if request.Database.Password == "***" {
			request.Database.Password = config.Database.Password
		}
		if request.Database.Username == "***" {
			request.Database.Username = config.Database.Username
		}
		if request.Licence == "***" {
			request.Licence = config.Licence
		}
		if request.ServerToken == "***" {
			request.ServerToken = config.ServerToken
		}

		utils.SetBaseMainConfig(request)
		
		utils.TriggerEvent(
			"cosmos.settings",
			"Settings updated",
			"success",
			"",
			map[string]interface{}{
		})

		go utils.SoftRestartServer()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingsUpdate: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
