package main

import (
	"net/http"
	"encoding/json"
	"runtime"
	"bytes"
	"encoding/gob"
	"os"
	"io"
	"os/exec"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/cpu"

	"github.com/azukaar/cosmos-server/src/utils" 
	"github.com/azukaar/cosmos-server/src/docker" 
	"github.com/azukaar/cosmos-server/src/proxy" 
	"github.com/azukaar/cosmos-server/src/metrics" 
	"github.com/azukaar/cosmos-server/src/market" 
	"github.com/azukaar/cosmos-server/src/cron" 
)

// StatusRoute godoc
// @Summary Get server status
// @Description Returns the current server status including configuration, resource info, and system state
// @Tags system
// @Produce json
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/status [get]
func StatusRoute(w http.ResponseWriter, req *http.Request) {
	config := utils.GetMainConfig()

	if !config.NewInstall && (utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil) {
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

		licenceValid := false 
		licenceNumber := 5
		if utils.FBL != nil && utils.FBL.LValid {
			licenceValid = true
			licenceNumber = utils.FBL.UserNumber
		}

		licenceNodeNumber := 1
		if utils.FBL != nil && utils.FBL.LValid {
			licenceNodeNumber = utils.FBL.CosmosNodeNumber
		}

		absoluteConfigPath := utils.CONFIGFOLDER
		absoluteConfigPath, _ = filepath.Abs(utils.CONFIGFOLDER)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
				"homepage": config.HomepageConfig,
				"theme": config.ThemeConfig,
				"resources": map[string]interface{} {
					// "ram": utils.GetRAMUsage(),
					// "ramFree": utils.GetAvailableRAM(),
					// "cpu": utils.GetCPUUsage(),
					// "disk": utils.GetDiskUsage(),
					// "network": utils.GetNetworkUsage(),
				},
				"containerized": utils.IsInsideContainer,
				"hostmode": utils.IsHostNetwork || !utils.IsInsideContainer || utils.GetMainConfig().DisableHostModeWarning,
				"database": utils.DBStatus,
				"docker": docker.DockerIsConnected,
				"backup_status": docker.ExportError,
				"letsencrypt": utils.GetMainConfig().HTTPConfig.HTTPSCertificateMode == "LETSENCRYPT" && utils.GetMainConfig().HTTPConfig.SSLEmail == "",
				"domain": utils.GetMainConfig().HTTPConfig.Hostname == "localhost" || utils.GetMainConfig().HTTPConfig.Hostname == "0.0.0.0",
				"HTTPSCertificateMode": utils.GetMainConfig().HTTPConfig.HTTPSCertificateMode,
				"CACert": utils.GetMainConfig().HTTPConfig.CACert,
				"needsRestart": utils.NeedsRestart,
				"newVersionAvailable": utils.NewVersionAvailable,
				"hostname": utils.GetMainConfig().HTTPConfig.Hostname,
				"allolocalip": utils.GetMainConfig().HTTPConfig.AllowHTTPLocalIPAccess,
				"CPU": runtime.GOARCH,
				"AVX": cpu.X86.HasAVX,
				"LetsEncryptErrors": utils.LetsEncryptErrors,
				"MonitoringDisabled": utils.GetMainConfig().MonitoringDisabled,
				"Licence": licenceValid,
				"LicenceNumber": licenceNumber,
				"LicenceNodeNumber": licenceNodeNumber,
				"ConfigFolder": absoluteConfigPath,
				"ConstellationName": utils.GetMainConfig().ConstellationConfig.ThisDeviceName,
			},
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// CanSendEmail godoc
// @Summary Check if email sending is enabled
// @Description Returns whether the server has email sending capabilities configured
// @Tags system
// @Produce json
// @Success 200 {object} utils.APIResponse
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/can-send-email [get]
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

// MemStatusRoute godoc
// @Summary Get internal memory status
// @Description Returns memory usage details for internal caches and buffers
// @Tags system
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/_memory [get]
func MemStatusRoute(w http.ResponseWriter, req *http.Request) {
	if (utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil) {
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

// LogsRoute godoc
// @Summary Download server logs
// @Description Returns the server log file as a plain text attachment
// @Tags system
// @Produce plain
// @Security BearerAuth
// @Success 200 {file} binary
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /_logs [get]
func LogsRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN_READ) != nil {
		return
	}

	if(req.Method == "GET") {
		// read log file
		logFile, err := os.Open(utils.CONFIGFOLDER + "cosmos.plain.log")

		if err != nil {
			utils.Error("Logs: Error reading log file", err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "HTTP001")
			return
		}

		defer logFile.Close()

		// read log file
		logBytes, err := io.ReadAll(logFile)

		if err != nil {
			utils.Error("Logs: Error reading log file", err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "HTTP001")
			return
		}

		// send log file
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", "attachment; filename=cosmos.log")
		w.Write(logBytes)
	} else {
		utils.Error("Logs: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// ForceUpdateRoute godoc
// @Summary Force check for server updates
// @Description Triggers a manual check for available server updates
// @Tags system
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/force-server-update [post]
func ForceUpdateRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN) != nil {
		return
	}

	if(req.Method == "POST") {
		utils.Log("API: Force update")
		checkUpdatesAvailable()
	} else {
		utils.Error("Logs: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// restartHostMachineRoute godoc
// @Summary Restart the host machine
// @Description Triggers a restart of the host machine (not available when running inside a container)
// @Tags system
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/restart-server [get]
func restartHostMachineRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN) != nil {
		return
	}

	if(req.Method == "GET") {
		utils.Log("API: Restarting host machine")

		if utils.IsInsideContainer {
			utils.Error("restartHostMachine: Restarting server from inside container is not possible", nil)
			utils.HTTPError(w, "Restarting server from inside container is not possible", http.StatusForbidden, "HTTP001")
			return
		}

		utils.WaitForAllJobs()

		err := restartHostMachine()
		if err != nil {
			utils.Error("restartHostMachine: Error restarting host machine (This usually means Cosmos does not have the permissions to restart your host server)", err)
			utils.HTTPError(w, "Error restarting host machine (This usually means Cosmos does not have the permissions to restart your host server) - " + err.Error(), http.StatusInternalServerError, "HTTP001")
			return
		}

	} else {
		utils.Error("restartHostMachine: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func restartHostMachine() error {
	switch runtime.GOOS {
	case "linux":
			return restartLinux()
	case "windows":
			return restartWindows()
	case "darwin":
			return restartMacOS()
	default:
			return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func restartLinux() error {
	cmd := exec.Command("shutdown", "-r", "now")
	return cmd.Run()
}

func restartWindows() error {
	cmd := exec.Command("shutdown", "/r", "/t", "0")
	return cmd.Run()
}

func restartMacOS() error {
	cmd := exec.Command("shutdown", "-r", "now")
	return cmd.Run()
}