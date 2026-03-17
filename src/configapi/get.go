package configapi

import (
	"net/http"
	"encoding/json"
	"os"
	"io/ioutil"
	"github.com/azukaar/cosmos-server/src/utils" 
)

// ConfigApiGet godoc
// @Summary Get server configuration
// @Description Returns the full server configuration (sensitive fields masked for non-credential users)
// @Tags config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=utils.Config}
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/config [get]
func ConfigApiGet(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil {
		return
	}

	isAdmin := utils.HasPermission(req, utils.PERM_CONFIGURATION_READ)
	canReadCredentials := utils.HasPermission(req, utils.PERM_CREDENTIALS_READ)

	if(req.Method == "GET") {
		config := utils.ReadConfigFromFile()

		// delete AuthPrivateKey and TLSKey
		config.HTTPConfig.AuthPrivateKey = ""
		config.HTTPConfig.TLSKey = ""

		if !canReadCredentials {
			config.MongoDB = "***"
			config.EmailConfig.Password = "***"
			config.EmailConfig.Username = "***"
			config.EmailConfig.Host = "***"
			config.Database.Password = "***"
			config.Database.Username = "***"
			config.HTTPConfig.DNSChallengeConfig = map[string]string{}
			config.Licence = "***"
			config.ServerToken = "***"
		}

		if !isAdmin {
			// filter admin only routes
			filteredRoutes := make([]utils.ProxyRouteConfig, 0)
			for _, route := range config.HTTPConfig.ProxyConfig.Routes {
				if !route.AdminOnly {
					filteredRoutes = append(filteredRoutes, route)
				}
			}
			config.HTTPConfig.ProxyConfig.Routes = filteredRoutes
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": config,
			"updates": utils.UpdateAvailable,
			"hostname": os.Getenv("HOSTNAME"),
			"isAdmin": isAdmin,
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// BackupFileApiGet godoc
// @Summary Get backup file
// @Description Returns the backup cosmos-compose JSON file
// @Tags config
// @Produce json
// @Security BearerAuth
// @Success 200 {file} binary
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/get-backup [get]
func BackupFileApiGet(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CREDENTIALS_READ) != nil {
		return
	}

	if req.Method == "GET" {
		// read file
		path := utils.CONFIGFOLDER + "backup.cosmos-compose.json"

		file, err := os.Open(path)
		if err != nil {
			utils.Error("BackupFileApiGet: Error opening file", err)
			utils.HTTPError(w, "Error opening file", http.StatusInternalServerError, "HTTP002")
			return
		}
		defer file.Close()

		// read content
		content, err := ioutil.ReadAll(file)
		if err != nil {
			utils.Error("BackupFileApiGet: Error reading file content", err)
			utils.HTTPError(w, "Error reading file content", http.StatusInternalServerError, "HTTP003")
			return
		}

		// set the content type header, so that the client knows it's receiving a JSON file
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(content)
		if err != nil {
			utils.Error("BackupFileApiGet: Error writing to response", err)
		}

	} else {
		utils.Error("BackupFileApiGet: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
