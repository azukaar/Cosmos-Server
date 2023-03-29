package main

import (
	"net/http"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 
	"github.com/azukaar/cosmos-server/src/docker" 
)

func StatusRoute(w http.ResponseWriter, req *http.Request) {
	if !utils.GetMainConfig().NewInstall && (utils.AdminOnly(w, req) != nil) {
		return
	}

	if(req.Method == "GET") {
		utils.Log("API: Status")

		databaseStatus := true
		
		if(!utils.GetMainConfig().DisableUserManagement) {
			err := utils.DB()
			if err != nil {
				utils.Error("Status: Database error", err)
				databaseStatus = false
			}
		} else {
			utils.Log("Status: User management is disabled, skipping database check")
		}

		if(!docker.DockerIsConnected) {
			ed := docker.Connect()
			if ed != nil {
				utils.Error("Status: Docker error", ed)
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
				"database": databaseStatus,
				"docker": docker.DockerIsConnected,
				"letsencrypt": utils.GetMainConfig().HTTPConfig.HTTPSCertificateMode == "LETSENCRYPT" && utils.GetMainConfig().HTTPConfig.SSLEmail == "",
				"domain": utils.GetMainConfig().HTTPConfig.Hostname == "localhost" || utils.GetMainConfig().HTTPConfig.Hostname == "0.0.0.0",
				"HTTPSCertificateMode": utils.GetMainConfig().HTTPConfig.HTTPSCertificateMode,
			},
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}