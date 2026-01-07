package constellation

import (
	"encoding/json"
	"math/rand"
	"net"
	"strings"
	"sync"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

func compareConfigs(configMap, configMapNew map[string]interface{}) bool {
	configMapStr := fmt.Sprintf("%+v", configMap)
	configMapNewStr := fmt.Sprintf("%+v", configMapNew)

	if string(configMapStr) == string(configMapNewStr) {
		return true
	} else {
		return false
	}
}


func setDefaultConstConfig(configMap map[string]interface{}) map[string]interface{} {
	utils.Debug("setDefaultConstConfig: Setting default values for client nebula.yml")

	if _, exists := configMap["logging"]; !exists {
			configMap["logging"] = map[string]interface{}{
					"format": "text",
					"level":  "info",
			}
	}

	if _, exists := configMap["listen"]; !exists {
			configMap["listen"] = map[string]interface{}{
					"host": "0.0.0.0",
					"port": 4242,
			}
	}

	if _, exists := configMap["punchy"]; !exists {
			configMap["punchy"] = map[string]interface{}{
					"punch":    true,
					"respond":  true,
			}
	}

	if _, exists := configMap["tun"]; !exists {
			configMap["tun"] = map[string]interface{}{
					"dev":                  "nebula1",
					"disabled":             false,
					"drop_local_broadcast": false,
					"drop_multicast":       false,
					"mtu":                  1300,
					"routes":               []interface{}{},
					"tx_queue":             500,
					"unsafe_routes":        []interface{}{},
			}
	}

	if _, exists := configMap["firewall"]; !exists {
			configMap["firewall"] = map[string]interface{}{
					"conntrack": map[string]interface{}{
							"default_timeout": "10m",
							"tcp_timeout":     "12m",
							"udp_timeout":     "3m",
					},
					"inbound": []interface{}{
							map[string]interface{}{
									"host":  "any",
									"port":  "any",
									"proto": "any",
							},
					},
					"inbound_action": "drop",
					"outbound": []interface{}{
							map[string]interface{}{
									"host":  "any",
									"port":  "any",
									"proto": "any",
							},
					},
					"outbound_action": "drop",
			}
	}

	return configMap
}


func GetDeviceConfigSync(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Duration(rand.Float64()*2)*time.Second)

	if(req.Method == "GET") {
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// get authorization header
		auth := req.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Unauthorized [1]", http.StatusUnauthorized)
			return
		}

		// remove "Bearer " from auth header
		auth = strings.Replace(auth, "Bearer ", "", 1)
		
		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
		defer closeDb()
		if errCo != nil {
			utils.Error("Database Connect", errCo)
			utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
			return
		}

		utils.Log("DeviceConfigSync: Fetching devices for IP " + ip)

		cursor, err := c.Find(nil, map[string]interface{}{
			"IP": ip + "/24",
			"APIKey": auth,
			"Blocked": false,
		})
		defer cursor.Close(nil)
		if err != nil {
			utils.Error("DeviceList: Error fetching devices", err)
			utils.HTTPError(w, "Error fetching devices", http.StatusInternalServerError, "DL003")
			return
		}
		
		// if any device is found, return config without keys
		if cursor.Next(nil) {
			d := utils.ConstellationDevice{}
			err := cursor.Decode(&d)
			if err != nil {
				utils.Error("DeviceList: Error decoding device", err)
				utils.HTTPError(w, "Error decoding device", http.StatusInternalServerError, "DL004")
				return
			}

			configYml, err := getYAMLClientConfig(d.DeviceName, utils.CONFIGFOLDER + "nebula.yml", "", "", "", "", utils.ConstellationDevice{
				Nickname: d.Nickname,
				DeviceName: d.DeviceName,
				PublicKey: "",
				IP: d.IP,
				IsLighthouse: d.IsLighthouse,
				IsCosmosNode: d.IsCosmosNode,
				IsRelay: d.IsRelay,
				PublicHostname: d.PublicHostname,
				Port: d.Port,
				APIKey: "",
			}, false, false)

			if err != nil {
				utils.Error("DeviceConfigSync: Error marshalling nebula.yml", err)
				utils.HTTPError(w, "Error marshalling nebula.yml", http.StatusInternalServerError, "DCS003")
				return
			}

			// Respond with the list of devices
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "OK",
				"data": string(configYml),
			})
		} else {
			utils.Error("DeviceConfigSync: Unauthorized [2]", nil)
			utils.HTTPError(w, "Unauthorized [2]", http.StatusUnauthorized, "DCS001")
			return
		}
		
	} else {
		utils.Error("DeviceConfigSync: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

type DeviceResyncRequest struct {
	Nickname string `json:"nickname",validate:"required,min=3,max=32,alphanum"`
	DeviceName string `json:"deviceName",validate:"required,min=3,max=32,alphanum"`
}

func GetDeviceConfigForSync(nickname, deviceName string) ([]byte, error) {
	utils.Log("DeviceConfigSync: Fetching config for " + deviceName + " from " + nickname)

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()
	if errCo != nil {
		utils.Error("Database Connect", errCo)
		return nil, errCo
	}

	cursor, err := c.Find(nil, map[string]interface{}{
		"DeviceName": deviceName,
		"Nickname": nickname,
	})
	defer cursor.Close(nil)
	if err != nil {
		utils.Error("DeviceList: Error fetching devices", err)
		return nil, err
	}
	
	// if any device is found, return config without keys
	if cursor.Next(nil) {
		utils.Log("DeviceConfigSync: Device found " + deviceName)

		d := utils.ConstellationDevice{}
		err := cursor.Decode(&d)
		if err != nil {
			utils.Error("DeviceList: Error decoding device", err)
			return nil, err
		}

		configYml, err := getYAMLClientConfig(d.DeviceName, utils.CONFIGFOLDER + "nebula.yml", "", "", "", "", utils.ConstellationDevice{
			Nickname: d.Nickname,
			DeviceName: d.DeviceName,
			PublicKey: "",
			IP: d.IP,
			IsLighthouse: d.IsLighthouse,
			IsCosmosNode: d.IsCosmosNode,
			IsRelay: d.IsRelay,
			PublicHostname: d.PublicHostname,
			Port: d.Port,
			APIKey: "",
		}, false, false)

		if err != nil {
			utils.Error("DeviceConfigSync: Error marshalling nebula.yml", err)
			return nil, err
		}

		// respond in JSON (from yaml)
		configJSON, err := json.Marshal(map[string]interface{}{
			"status":  "OK",
			"data": string(configYml),
		})

		if err != nil {
			utils.Error("DeviceConfigSync: Error marshalling nebula.yml to JSON", err)
			return nil, err
		}

		return configJSON, nil		
	} else {
		utils.Error("DeviceConfigSync: Unauthorized [2]", nil)
		return nil, errors.New("Unauthorized")
	}
}

func GetDeviceConfigManualSync(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "POST") {
		var request DeviceResyncRequest
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("DeviceConfigManualSync: Invalid User Request", err1)
			utils.HTTPError(w, "Device Creation Error",
				http.StatusInternalServerError, "DB001")
			return
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("DeviceConfigManualSync: Invalid User Request", errV)
			utils.HTTPError(w, "Device Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "DB002")
			return 
		}
		
		nickname := utils.Sanitize(request.Nickname)
		deviceName := utils.Sanitize(request.DeviceName)
		
		if utils.AdminOrItselfOnly(w, req, nickname) != nil {
			return
		}

		utils.Log("DeviceConfigManualSync: Resync Device " + deviceName)

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
  	defer closeDb()
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		cursor, err := c.Find(nil, map[string]interface{}{
			"Nickname": nickname,
			"DeviceName": deviceName,
		})
		defer cursor.Close(nil)
		if err != nil {
			utils.Error("DeviceList: Error fetching devices", err)
			utils.HTTPError(w, "Error fetching devices", http.StatusInternalServerError, "DL003")
			return
		}
		
		// if any device is found, return config without keys
		if cursor.Next(nil) {
			d := utils.ConstellationDevice{}
			err := cursor.Decode(&d)
			if err != nil {
				utils.Error("DeviceList: Error decoding device", err)
				utils.HTTPError(w, "Error decoding device", http.StatusInternalServerError, "DL004")
				return
			}

			configYml, err := getYAMLClientConfig(d.DeviceName, utils.CONFIGFOLDER + "nebula.yml", "", "", "", "", utils.ConstellationDevice{
				Nickname: d.Nickname,
				DeviceName: d.DeviceName,
				PublicKey: "",
				IP: d.IP,
				IsLighthouse: d.IsLighthouse,
				IsCosmosNode: d.IsCosmosNode,
				IsRelay: d.IsRelay,
				PublicHostname: d.PublicHostname,
				Port: d.Port,
				APIKey: "",
			}, true, true)

			if err != nil {
				utils.Error("DeviceConfigSync: Error marshalling nebula.yml", err)
				utils.HTTPError(w, "Error marshalling nebula.yml", http.StatusInternalServerError, "DCS003")
				return
			}

			// Respond with the list of devices
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "OK",
				"data": string(configYml),
			})
		} else {
			utils.Error("DeviceConfigSync: Unauthorized [2]", nil)
			utils.HTTPError(w, "Unauthorized [2]", http.StatusUnauthorized, "DCS001")
			return
		}
		
	} else {
		utils.Error("DeviceConfigManualSync: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func GetNATSCredentials(isMaster bool) (string, string, error) {
	if isMaster {
		return MASTERUSER, MASTERPWD, nil
	}
	
	nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")
	if err != nil {
		utils.Error("GetNATSCredentials: error while reading nebula.yml", err)
		return "", "", err
	}

	configMap := make(map[string]interface{})
	err = yaml.Unmarshal(nebulaFile, &configMap)
	if err != nil {
		utils.Error("GetNATSCredentials: Invalid slave config file for resync", err)
		return "", "", err
	}

	if configMap["cstln_api_key"] == nil || configMap["cstln_device_name"] == nil {
		utils.Error("GetNATSCredentials: Invalid slave config file for resync", nil)
		return "", "", errors.New("Invalid slave config file for resync")
	}

	apiKey := configMap["cstln_api_key"].(string)
	deviceName := configMap["cstln_device_name"].(string)

	return deviceName, apiKey, nil
}

func SlaveConfigSync(newConfig string) (bool, error) {
	if !utils.GetMainConfig().ConstellationConfig.Enabled || !utils.GetMainConfig().ConstellationConfig.SlaveMode {
		return false, nil
	}
	
	utils.Log("SlaveConfigSync: Resyncing config")
	
	nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")
	if err != nil {
		utils.Error("SlaveConfigSync: error while reading nebula.yml", err)
		return false, err
	}

	configMap := make(map[string]interface{})
	err = yaml.Unmarshal(nebulaFile, &configMap)
	if err != nil {
		utils.Error("SlaveConfigSync: Invalid slave config file for resync", err)
		return false, err
	}

	rawEndpoint := configMap["cstln_config_endpoint"]
	apiKey := configMap["cstln_api_key"]

	if rawEndpoint == nil  || apiKey == nil {
		utils.Error("SlaveConfigSync: Invalid slave config file for resync", nil)
		return false, errors.New("Invalid slave config file for resync")
	}

	endpoint := rawEndpoint.(string)
	endpoint += "cosmos/api/constellation/config-sync"

	// utils.Log("SlaveConfigSync: Fetching config from " + endpoint.(string))

	// fetch the config from the endpoint with Authorization header
	// req, err := http.NewRequest("GET", endpoint.(string), nil)
	// if err != nil {
	// 	utils.Error("SlaveConfigSync: Error creating request", err)
	// 	return false, err
	// }

	body := newConfig

	if body == "" {
		utils.Log("SlaveConfigSync: Fetching config from NATS")

		body, err = SendNATSMessage(NATSClientTopic + ".constellation.config", "SYNC")
		if err != nil {
			utils.Error("SlaveConfigSync: Error fetching config", err)
			return false, err
		}
	} else {
		utils.Log("SlaveConfigSync: Config already received. Sync now")
	}

	type Response struct {
    Status string      `json:"status"`
    Data   interface{} `json:"data"` // Use interface{} if the data type can vary
	}
	var jsonresp Response
	err = json.Unmarshal([]byte(body), &jsonresp)
	if err != nil {
			utils.Error("SlaveConfigSync: Error unmarshalling JSON", err)
			return false, err
	}

	if jsonresp.Status != "OK" {
		utils.Error("SlaveConfigSync: Error fetching config", nil)
		return false, errors.New("Error fetching config")
	}

	// get jsonbody.data
	databody := jsonresp.Data

	// parse the response and re-apply cstln_api_key, pki cert, key and ca
	var configMapNew map[string]interface{}
	err = yaml.Unmarshal([]byte(databody.(string)), &configMapNew)
	if err != nil {
		utils.Error("SlaveConfigSync: Invalid slave config file for resync", err)
		return false, err
	}

	configMapNew = setDefaultConstConfig(configMapNew)

	configMapNew["cstln_api_key"] = apiKey

	pkiMap, ok := configMapNew["pki"].(map[string]interface{})
	if !ok {
		pkiMap = make(map[string]interface{})
	}

	pkiMap["cert"] = configMap["pki"].(map[interface{}]interface{})["cert"]
	pkiMap["key"] = configMap["pki"].(map[interface{}]interface{})["key"]
	pkiMap["ca"] = configMap["pki"].(map[interface{}]interface{})["ca"]

	configMapNew["pki"] = pkiMap
	
	// apply tunnels 

	// Define the slice to hold the unmarshalled tunnels.
	var tunnels []utils.ProxyRouteConfig
	config := utils.ReadConfigFromFile()

	// Convert the `cstln_tunnels` part back to YAML string.
	tunnelsData, err := yaml.Marshal(configMapNew["cstln_tunnels"])
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

	// extract cstln_is_cosmos_node from config
	isCosmosNode, ok := configMapNew["cstln_is_cosmos_node"].(bool)
	if ok && isCosmosNode {
		utils.SetIsCosmosNode(isCosmosNode)
	}

	// write the new config back to file

	configYml, err := yaml.Marshal(configMapNew)
	if err != nil {
		utils.Error("SlaveConfigSync: Error marshalling new config", err)
		return false, err
	}

	err = ioutil.WriteFile(utils.CONFIGFOLDER + "nebula.yml", configYml, 0644)
	if err != nil {
		utils.Error("SlaveConfigSync: Error writing new config", err)
		return false, err
	}

	utils.Log("SlaveConfigSync: Config resynced")
	
	utils.SetBaseMainConfig(config)
	
	utils.TriggerEvent(
		"cosmos.settings",
		"Settings updated",
		"success",
		"",
		map[string]interface{}{
			"from": "Constellation",
	})

	// if the config change, restart
	if !compareConfigs(configMap, configMapNew) {
		return true, nil
	}

	return false, nil
}

var clientSyncLock = sync.Mutex{}

func TriggerClientResync() error {
	clientSyncLock.Lock()
	defer clientSyncLock.Unlock()

	utils.Log("TriggerClientResync: Resyncing all clients")

	// get al clients
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()

	if errCo != nil {
		utils.Error("TriggerClientResync: Database Connect", errCo)
		return errCo
	}

	cursor, err := c.Find(nil, map[string]interface{}{
		"Blocked": false,
	})

	defer cursor.Close(nil)

	if err != nil {
		utils.Error("TriggerClientResync: Error fetching devices", err)
		return err
	}

	for cursor.Next(nil) {
		d := utils.ConstellationDevice{}
		err := cursor.Decode(&d)
		if err != nil {
			utils.Error("TriggerClientResync: Error decoding device", err)
			continue
		}

		// send message to client
		username := sanitizeNATSUsername(d.DeviceName)
		if d.IsLighthouse {
			body, err := GetDeviceConfigForSync(d.Nickname, d.DeviceName)
			if err != nil {
				utils.Error("TriggerClientResync: Error getting device config for sync", err)
				continue
			}

			err = PublishNATSMessage("cosmos."+username+".constellation.config.resync", (string)(body))
			

			if err != nil {
				utils.Error("TriggerClientResync: Error sending resync message to client", err)
			}

			if !utils.GetMainConfig().ConstellationConfig.DoNotSyncNodes {
				SendSyncPayload(username)
			}
			
			utils.Log("TriggerClientResync: Resync message sent to " + username)
		}
	}

	return nil
}
