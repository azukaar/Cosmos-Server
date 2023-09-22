package constellation

import (
	"net/http"
	"encoding/json"
	"io/ioutil"

	
	"github.com/azukaar/cosmos-server/src/utils" 
)

func API_ConnectToExisting(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if(req.Method == "POST") {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			utils.Error("API_Restart: Invalid User Request", err)
			utils.HTTPError(w, "API_Restart Error",
				http.StatusInternalServerError, "AR001")
			return	
		}

		config := utils.ReadConfigFromFile()
		config.ConstellationConfig.Enabled = true
		config.ConstellationConfig.SlaveMode = true
		config.ConstellationConfig.DNS = false
		// ConstellationHostname = 

		// output utils.CONFIGFOLDER + "nebula.yml"
		err = ioutil.WriteFile(utils.CONFIGFOLDER + "nebula.yml", body, 0644)
		
		utils.SetBaseMainConfig(config)
		
		RestartNebula()
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
