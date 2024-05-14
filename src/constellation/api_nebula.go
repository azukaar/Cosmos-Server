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
		"IP": ip + "/24",
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

func API_GetConfig(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if(req.Method == "GET") {
		// read utils.CONFIGFOLDER + "nebula.yml"
		config, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")

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

func API_Restart(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if(req.Method == "GET") {
		RestartNebula()
		utils.RestartHTTPServer()

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

func API_Reset(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
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

func API_GetLogs(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
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
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}