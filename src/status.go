package main

import (
	"net/http"
	"encoding/json"
	"os"
	"runtime"
	"bytes"
	"encoding/gob"

	"golang.org/x/sys/cpu"

	"github.com/azukaar/cosmos-server/src/utils" 
	"github.com/azukaar/cosmos-server/src/docker" 
	"github.com/azukaar/cosmos-server/src/proxy" 
	"github.com/azukaar/cosmos-server/src/metrics" 
	"github.com/azukaar/cosmos-server/src/market" 
	"github.com/azukaar/cosmos-server/src/cron" 
)

func StatusRoute(w http.ResponseWriter, req *http.Request) {
	config := utils.GetMainConfig()

	if !config.NewInstall && (utils.LoggedInOnly(w, req) != nil) {
		return
	}

	if(req.Method == "GET") {
		utils.Log("API: Status")
		
		if config.NewInstall {
			if(!config.DisableUserManagement) {
				err := utils.DB()
				if err != nil {
					utils.Error("Status: Database error", err)
				}
			} else {
				utils.Log("Status: User management is disabled, skipping database check")
			}
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
				"containerized": os.Getenv("HOSTNAME") != "",
				"hostmode": utils.IsHostNetwork || os.Getenv("HOSTNAME") == "" || utils.GetMainConfig().DisableHostModeWarning,
				"database": utils.DBStatus,
				"docker": docker.DockerIsConnected,
				"backup_status": docker.ExportError,
				"letsencrypt": utils.GetMainConfig().HTTPConfig.HTTPSCertificateMode == "LETSENCRYPT" && utils.GetMainConfig().HTTPConfig.SSLEmail == "",
				"domain": utils.GetMainConfig().HTTPConfig.Hostname == "localhost" || utils.GetMainConfig().HTTPConfig.Hostname == "0.0.0.0",
				"HTTPSCertificateMode": utils.GetMainConfig().HTTPConfig.HTTPSCertificateMode,
				"needsRestart": utils.NeedsRestart,
				"newVersionAvailable": utils.NewVersionAvailable,
				"hostname": utils.GetMainConfig().HTTPConfig.Hostname,
				"allolocalip": utils.GetMainConfig().HTTPConfig.AllowHTTPLocalIPAccess,
				"CPU": runtime.GOARCH,
				"AVX": cpu.X86.HasAVX,
				"LetsEncryptErrors": utils.LetsEncryptErrors,
				"MonitoringDisabled": utils.GetMainConfig().MonitoringDisabled,
				"Licence": utils.FBL.LValid,
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

func getRealSizeOf(v interface{}) (int) {
	b := new(bytes.Buffer)
	// if map return number of keys
	if m, ok := v.(map[interface{}]interface{}); ok {
		return len(m)
	}
	// if array return number of elements
	if a, ok := v.([]interface{}); ok {
		return len(a)
	}

	if err := gob.NewEncoder(b).Encode(v); err != nil {
			utils.Error("MemStatusRoute: Error encoding", err)
			return 0
	}
	
	return b.Len()
}
func getRealSizeOf2(v interface{}) (int) {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(v); err != nil {
			utils.Error("MemStatusRoute: Error encoding", err)
			return 0
	}
	
	return b.Len()
}

func MemStatusRoute(w http.ResponseWriter, req *http.Request) {
	if (utils.LoggedInOnly(w, req) != nil) {
		return
	}

	if(req.Method == "GET") {


		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
				"IconCache": getRealSizeOf(IconCache),
				"UpdateAvailable": getRealSizeOf(utils.UpdateAvailable),
				"LetsEncryptErrors": getRealSizeOf(utils.LetsEncryptErrors),
				"BannedIPs": getRealSizeOf2(utils.BannedIPs),
				"WriteBuffer": getRealSizeOf2(utils.GetWriteBuffer()),
				"Shield": proxy.GetShield(),
				"ActiveProxies": getRealSizeOf2(proxy.GetActiveProxies()),
				"Markets": getRealSizeOf(market.GetCachedMarket()),
				"Metrics": getRealSizeOf(metrics.GetDataBuffer()),
				"CRON": getRealSizeOf(cron.GetJobsList()),
			},
		})
	} else {
		utils.Error("MemRoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}