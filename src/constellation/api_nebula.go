package constellation

import (
	"net/http"
	"net"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"	
	"errors"
	
	"github.com/azukaar/cosmos-server/src/utils" 
)

// TODO: Cache this
func CheckConstellationToken(req *http.Request) error {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return errors.New("Invalid request")
	}

	// get authorization header
	auth := req.Header.Get("x-cstln-auth")
	if auth == "" {
		return errors.New("Unauthorized: No Authorization header")
	}

	// remove "Bearer " from auth header
	auth = strings.Replace(auth, "Bearer ", "", 1)
	
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()
	if errCo != nil {
		return errCo
	}

	utils.Log("DeviceConfigSync: Fetching devices for IP " + ip)

	cursor, err := c.Find(nil, map[string]interface{}{
		"IP": ip,
		"APIKey": auth,
		"Blocked": false,
	})
	defer cursor.Close(nil)
	if err != nil {
		return err
	}
	
	// if any device is found, return config without keys
	if cursor.Next(nil) {
		return nil
	}

	return errors.New("Unauthorized: Client not found")
}

// API_GetConfig godoc
// @Summary Get the current Nebula configuration
// @Tags constellation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/constellation/config [get]
func API_GetConfig(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if(req.Method == "GET") {
		// read utils.CONFIGFOLDER + "nebula.yml"
		config, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula-temp.yml")

		if err != nil {
			utils.Error("SettingGet: error while reading nebula.yml", err)
			utils.HTTPError(w, "Error while reading nebula.yml", http.StatusInternalServerError, "HTTP002")
			return
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": string(config),
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// API_Restart godoc
// @Summary Restart the Nebula VPN service and HTTP server
// @Tags constellation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/constellation/restart [get]
func API_Restart(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if(req.Method == "GET") {
		RestartNebula()
		go utils.RestartHTTPServer()

		utils.Log("Constellation: nebula restarted")
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// API_Reset godoc
// @Summary Reset the Nebula VPN configuration
// @Tags constellation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/constellation/reset [get]
func API_Reset(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if(req.Method == "GET") {
		ResetNebula()

		utils.Log("Constellation: nebula reset")
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// API_GetLogs godoc
// @Summary Get Nebula VPN service logs
// @Tags constellation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/constellation/logs [get]
func API_GetLogs(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if(req.Method == "GET") {
		logs, err := os.ReadFile(utils.CONFIGFOLDER+"nebula.log")
		if err != nil {
			utils.Error("Error reading file:", err)
			return
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": string(logs),
		})
	} else {
		utils.Error("API_GetLogs: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// API_Ping godoc
// @Summary Check if the NATS client connection is alive
// @Tags constellation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/constellation/ping [get]
func API_Ping(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil {
		return
	}

	if(req.Method == "GET") {
		isConnected := PingNATSClient()
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": isConnected,
		})
	} else {
		utils.Error("API_Ping: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}