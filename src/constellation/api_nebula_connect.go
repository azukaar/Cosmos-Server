package constellation

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	
	"github.com/azukaar/cosmos-server/src/utils" 
)

func API_ConnectToExisting(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if(req.Method == "POST") {
		utils.ConfigLock.Lock()
		defer utils.ConfigLock.Unlock()

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
		config.ConstellationConfig.DNSDisabled = true

		// extract "tunnels" from config 
		// and append the new tunnel to it
		
		var configMap map[string]interface{}

		err = yaml.Unmarshal(body, &configMap)
		if err != nil {
			utils.Error("API_ConnectToExisting: Invalid User Request", err)
			utils.HTTPError(w, "API_ConnectToExisting Error",
				http.StatusInternalServerError, "ACE001")
			return	
		}

    // Define the slice to hold the unmarshalled tunnels.
    var tunnels []utils.ProxyRouteConfig

    // Convert the `cstln_tunnels` part back to YAML string.
    tunnelsData, err := yaml.Marshal(configMap["cstln_tunnels"])
    if err != nil {
        utils.Error("Error marshalling tunnels data back to YAML", err)
    } else {
			// Unmarshal the YAML string into the specific struct.
			err = yaml.Unmarshal(tunnelsData, &tunnels)

			if err != nil {
					utils.Error("Error unmarshalling tunnels YAML", err)
			}	
		}

		config.ConstellationConfig.Tunnels = tunnels

		// output utils.CONFIGFOLDER + "nebula.yml"
		err = ioutil.WriteFile(utils.CONFIGFOLDER + "nebula.yml", body, 0644)
		
		utils.SetBaseMainConfig(config)
		
		utils.TriggerEvent(
			"cosmos.settings",
			"Settings updated",
			"success",
			"",
			map[string]interface{}{
				"from": "Constellation",
		})
	
		RestartNebula()
		utils.RestartHTTPServer()
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
