package constellation

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
	
	"github.com/azukaar/cosmos-server/src/utils" 
)

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