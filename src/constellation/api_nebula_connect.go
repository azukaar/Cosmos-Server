package constellation

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
	"gopkg.in/yaml.v2"

	"github.com/azukaar/cosmos-server/src/utils"
)

func API_NewConstellation(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		utils.ConfigLock.Lock()
		defer utils.ConfigLock.Unlock()

		utils.Log("API_NewConstellation: creating new Constellation")

		var request struct {
			DeviceName string `json:"deviceName"`
		}

		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			utils.Error("API_NewConstellation: Invalid User Request", err)
			utils.HTTPError(w, "API_NewConstellation Error",
				http.StatusInternalServerError, "ANC001")
			return
		}

		if request.DeviceName == "" {
			utils.Error("API_NewConstellation: Device name is required", nil)
			utils.HTTPError(w, "Device name is required",
				http.StatusBadRequest, "ANC002")
			return
		}
	
		// check if ca.crt exists
		if _, err = os.Stat(utils.CONFIGFOLDER + "ca.crt"); os.IsNotExist(err) {
			utils.Log("Constellation: ca.crt not found, generating...")
			// generate ca.crt
			
			errG := generateNebulaCACert("Cosmos - " + utils.GetMainConfig().ConstellationConfig.ConstellationHostname)
			if errG != nil {
				utils.Error("Constellation: error while generating ca.crt", errG)
			}
		}

		utils.Log("Constellation: cosmos.crt generating...")
		// generate cosmos.crt
		_,_,_,errG := generateNebulaCert("cosmos", "192.168.201.1", "", true)
		if errG != nil {
			utils.Error("Constellation: error while generating cosmos.crt", errG)
		}
	
		DeviceCreateRequest := DeviceCreateRequestJSON{
			DeviceName: request.DeviceName,
			IP: "192.168.201.1",
			IsLighthouse: true,
			IsCosmosNode: true,
			Nickname: "",
		}

		_, _, _, response, err := DeviceCreate(DeviceCreateRequest)

		if err != nil {
			utils.Error("API_NewConstellation: Error creating lighthouse device", err)
			utils.HTTPError(w, "API_NewConstellation Error: " + err.Error(),
				http.StatusInternalServerError, "ANC003")
			return
		}

		deviceName := response.DeviceName

		config := utils.ReadConfigFromFile()
		config.ConstellationConfig.Enabled = true
		config.ConstellationConfig.ThisDeviceName = deviceName

		utils.SetBaseMainConfig(config)

		utils.TriggerEvent(
			"cosmos.settings",
			"Settings updated",
			"success",
			"",
			map[string]interface{}{
				"from": "Constellation",
			})

		utils.Log("API_NewConstellation: Constellation created with device name: " + deviceName)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})

		go func() {
			RestartNebula()
			utils.RestartHTTPServer()
		}()
	} else {
		utils.Error("API_NewConstellation: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func API_ConnectToExisting(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if(req.Method == "POST") {
		utils.ConfigLock.Lock()
		defer utils.ConfigLock.Unlock()

		utils.Log("API_ConnectToExisting: connecting to an external Constellation")

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
		
		var configMap map[string]interface{}

		err = yaml.Unmarshal(body, &configMap)
		if err != nil {
			utils.Error("API_ConnectToExisting: Invalid User Request", err)
			utils.HTTPError(w, "API_ConnectToExisting Error",
				http.StatusInternalServerError, "ACE001")
			return	
		}

		configMap = setDefaultConstConfig(configMap)

		configMapString, err := yaml.Marshal(configMap)
		if err != nil {
			utils.Error("API_ConnectToExisting: Invalid User Request", err)
			utils.HTTPError(w, "API_ConnectToExisting Error",
				http.StatusInternalServerError, "ACE002")
			return
		}

		// output utils.CONFIGFOLDER + "nebula.yml"
		err = ioutil.WriteFile(utils.CONFIGFOLDER + "nebula.yml", configMapString, 0644)
		
		if deviceNameVal, ok := configMap["cstln_device_name"]; ok {
			config.ConstellationConfig.ThisDeviceName = deviceNameVal.(string)
		} else {
			utils.Error("API_ConnectToExisting: device name not found in config", nil)
			utils.HTTPError(w, "API_ConnectToExisting Error: device name not found in config",
				http.StatusInternalServerError, "ACE003")
			return
		}

		// read values into main config
		if exitNodeVal, ok := configMap["cstln_is_exit_node"]; ok {
			if exitNodeStr, ok := exitNodeVal.(string); ok {
				config.ConstellationConfig.IsExitNode = (exitNodeStr == "true")
			} else if exitNodeBool, ok := exitNodeVal.(bool); ok {
				config.ConstellationConfig.IsExitNode = exitNodeBool
			}
		}
		
		utils.SetBaseMainConfig(config)
		
		utils.TriggerEvent(
			"cosmos.settings",
			"Settings updated",
			"success",
			"",
			map[string]interface{}{
				"from": "Constellation",
		})
	

		utils.Log("API_ConnectToExisting: connected to an external Constellation")
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
		
		go func() {
			RestartNebula()
			utils.RestartHTTPServer()
		}()
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
