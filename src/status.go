package main

import (
	"net/http"
	"encoding/json"
	"os"
	"runtime"
	"golang.org/x/sys/cpu"

	"github.com/azukaar/cosmos-server/src/utils" 
	"github.com/azukaar/cosmos-server/src/docker" 
)

func StatusRoute(w http.ResponseWriter, req *http.Request) {
	config := utils.GetMainConfig()

	if !config.NewInstall && (utils.LoggedInOnly(w, req) != nil) {
		return
	}

	if(req.Method == "GET") {
		utils.Log("API: Status")

		databaseStatus := true
		
		if(!config.DisableUserManagement) {
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
				"homepage": config.HomepageConfig,
				"theme": config.ThemeConfig,
				"resources": map[string]interface{}{
					// "ram": utils.GetRAMUsage(),
					// "ramFree": utils.GetAvailableRAM(),
					// "cpu": utils.GetCPUUsage(),
					// "disk": utils.GetDiskUsage(),
					// "network": utils.GetNetworkUsage(),
				},
				"hostmode": utils.IsHostNetwork || os.Getenv("HOSTNAME") == "",
				"database": databaseStatus,
				"docker": docker.DockerIsConnected,
				"backup_status": docker.ExportError,
				"letsencrypt": utils.GetMainConfig().HTTPConfig.HTTPSCertificateMode == "LETSENCRYPT" && utils.GetMainConfig().HTTPConfig.SSLEmail == "",
				"domain": utils.GetMainConfig().HTTPConfig.Hostname == "localhost" || utils.GetMainConfig().HTTPConfig.Hostname == "0.0.0.0",
				"HTTPSCertificateMode": utils.GetMainConfig().HTTPConfig.HTTPSCertificateMode,
				"needsRestart": utils.NeedsRestart,
				"newVersionAvailable": utils.NewVersionAvailable,
				"hostname": utils.GetMainConfig().HTTPConfig.Hostname,
				"CPU": runtime.GOARCH,
				"AVX": cpu.X86.HasAVX,
				"LetsEncryptErrors": utils.LetsEncryptErrors,
				"MonitoringDisabled": utils.GetMainConfig().MonitoringDisabled,
			},
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func CanSendEmail(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "GET") {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
				"canSendEmail": utils.GetMainConfig().EmailConfig.Enabled,
			},
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
