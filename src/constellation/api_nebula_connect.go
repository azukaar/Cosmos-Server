package constellation

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"gopkg.in/yaml.v2"

	"github.com/azukaar/cosmos-server/src/utils"
)

// API_NewConstellation godoc
// @Summary Create a new Constellation VPN network
// @Tags constellation
// @Accept json
// @Produce json
// @Param body body object true "Constellation creation payload (deviceName, isLighthouse, hostname, ipRange)"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/constellation/create [post]
func API_NewConstellation(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
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
			NATSReplicas int `json:"natsReplicas,omitempty"`
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
			PublicHostname: request.Hostname,
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
		if request.NATSReplicas > 0 && utils.IsPro() {
			config.ConstellationConfig.NATSReplicas = request.NATSReplicas
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

// ConnectToExisting applies a Nebula YAML config to connect this node to an
// existing Constellation network. It returns the updated config. The caller
// is responsible for persisting the config and restarting Nebula.
func ConnectToExisting(yamlBody []byte, config utils.Config) (utils.Config, error) {
	utils.Log("ConnectToExisting: connecting to an external Constellation")

	config.ConstellationConfig.Enabled = true

	var configMap map[string]interface{}
	err := yaml.Unmarshal(yamlBody, &configMap)
	if err != nil {
		return config, err
	}

	configMap = setDefaultConstConfig(configMap)

	configMapString, err := yaml.Marshal(configMap)
	if err != nil {
		return config, err
	}

	err = ioutil.WriteFile(utils.CONFIGFOLDER+"nebula.yml", configMapString, 0600)
	if err != nil {
		return config, err
	}

	if deviceNameVal, ok := configMap["cstln_device_name"]; ok {
		config.ConstellationConfig.ThisDeviceName = deviceNameVal.(string)
	} else {
		return config, errors.New("device name not found in constellation config")
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

	utils.Log("ConnectToExisting: connected to an external Constellation")
	return config, nil
}

// API_ConnectToExisting godoc
// @Summary Connect this node to an existing Constellation VPN network
// @Tags constellation
// @Accept application/x-yaml
// @Produce json
// @Param body body string true "Nebula YAML configuration"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/constellation/connect [post]
func API_ConnectToExisting(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if(req.Method == "POST") {
		utils.ConfigLock.Lock()
		defer utils.ConfigLock.Unlock()

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			utils.Error("API_ConnectToExisting: Invalid User Request", err)
			utils.HTTPError(w, "API_ConnectToExisting Error",
				http.StatusInternalServerError, "AR001")
			return
		}

		config := utils.ReadConfigFromFile()
		config, err = ConnectToExisting(body, config)
		if err != nil {
			utils.Error("API_ConnectToExisting: Error", err)
			utils.HTTPError(w, "API_ConnectToExisting Error: "+err.Error(),
				http.StatusInternalServerError, "ACE001")
			return
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
