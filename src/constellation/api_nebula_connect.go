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
			IsLighthouse bool   `json:"isLighthouse"`
			Hostname string `json:"hostname"`
			IPRange string `json:"ipRange"`
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

		if request.Hostname == "" {
			utils.Error("API_NewConstellation: Hostname is required", nil)
			utils.HTTPError(w, "Hostname is required",
				http.StatusBadRequest, "ANC004")
			return
		}

		if request.IPRange == "" {
			request.IPRange = "192.168.201.0/24"
		}
		
		utils.Log("Constellation: exporting nebula.yml...")
		err = ExportDefaultConfigToYAML(utils.CONFIGFOLDER + "nebula.yml")

		if err != nil {
			utils.Error("Constellation: error while exporting nebula.yml", err)
		}

		// check if ca.crt exists
		if _, errCA := os.Stat(utils.CONFIGFOLDER + "ca.crt"); errCA == nil {
			utils.Log("Constellation: ca.crt found, deleting...")
			// delete ca.crt
			errD := os.Remove(utils.CONFIGFOLDER + "ca.crt")
			if errD != nil {
				utils.Error("Constellation: error while deleting ca.crt", errD)
			}
			os.Remove(utils.CONFIGFOLDER + "ca.key")
		}

		errG := generateNebulaCACert("Cosmos - " + request.Hostname)
		if errG != nil {
			utils.Error("Constellation: error while generating ca.crt", errG)
		}

		ip := GetNextAvailableIP(request.IPRange)

		utils.Log("Constellation: cosmos.crt generating with ip " + ip)
		
		// generate cosmos.crt
		_,_,_,errG = generateNebulaCert(request.DeviceName, "cosmos", ip, "", true)
		if errG != nil {
			utils.Error("Constellation: error while generating cosmos.crt", errG)
		}
	
		DeviceCreateRequest := DeviceCreateRequestJSON{
			DeviceName: request.DeviceName,
			IP: ip,
			IsLighthouse: request.IsLighthouse,
			CosmosNode: 2,
			IsRelay: true,
			IsExitNode: true,
			IsLoadBalancer: true,
			Nickname: "",
			PublicHostname: utils.GetMainConfig().ConstellationConfig.ConstellationHostname,
			Port: "4242",
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
		config.ConstellationConfig.ConstellationHostname = request.Hostname
		config.ConstellationConfig.IPRange = request.IPRange
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

		if publicHostnameVal, ok := configMap["cstln_public_hostname"]; ok {
			config.ConstellationConfig.ConstellationHostname = publicHostnameVal.(string)
		}

		if licence, ok := configMap["cstln_server_licence"]; ok {
			config.Licence = licence.(string)
		}

		if cosmosNode, ok := configMap["cstln_cosmos_node"]; ok {
			config.AgentMode = cosmosNode.(int) == 1
		}

		if ipRange, ok := configMap["cstln_ip_range"]; ok {
			config.ConstellationConfig.IPRange = ipRange.(string)
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
