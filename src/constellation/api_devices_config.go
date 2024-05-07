package constellation

import (
	"net/http"
	"encoding/json"
	"math/rand"
	"time"
	"net"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"reflect"
	
	"github.com/azukaar/cosmos-server/src/utils" 
)

func DeviceConfigSync(w http.ResponseWriter, req *http.Request) {
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
		utils.Log("DeviceConfigSync: Fetching devices for APIKey " + auth)

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
				IsRelay: d.IsRelay,
				PublicHostname: d.PublicHostname,
				Port: d.Port,
				APIKey: "",
			}, false)

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

func DeviceConfigManualSync(w http.ResponseWriter, req *http.Request) {
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
				IsRelay: d.IsRelay,
				PublicHostname: d.PublicHostname,
				Port: d.Port,
				APIKey: "",
			}, true)

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

func SlaveConfigSync() {
	nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")
	if err != nil {
		utils.Error("SlaveConfigSync: error while reading nebula.yml", err)
		return
	}

	configMap := make(map[string]interface{})
	err = yaml.Unmarshal(nebulaFile, &configMap)
	if err != nil {
		utils.Error("SlaveConfigSync: Invalid slave config file for resync", err)
		return
	}

	endpoint := configMap["cstln_config_endpoint"]
	apiKey := configMap["cstln_api_key"]

	if endpoint == nil  || apiKey == nil {
		utils.Error("SlaveConfigSync: Invalid slave config file for resync", nil)
		return
	}

	// fetch the config from the endpoint with Authorization header
	req, err := http.NewRequest("GET", endpoint.(string), nil)
	if err != nil {
		utils.Error("SlaveConfigSync: Error creating request", err)
		return
	}

	req.Header = map[string][]string{
		"Authorization": {"Bearer " + apiKey.(string)},
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		utils.Error("SlaveConfigSync: Error fetching config", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		utils.Error("SlaveConfigSync: Error fetching config", nil)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Error("SlaveConfigSync: Error reading response", err)
		return
	}

	// parse the response and re-apply cstln_api_key, pki cert, key and ca
	var configMapNew map[string]interface{}
	err = yaml.Unmarshal(body, &configMapNew)
	if err != nil {
		utils.Error("SlaveConfigSync: Invalid slave config file for resync", err)
		return
	}

	configMapNew["cstln_api_key"] = apiKey
	
	pkiMap, ok := configMapNew["pki"].(map[string]interface{})
	if !ok {
			utils.Error("SlaveConfigSync: Error asserting pki type", nil)
			return
	}

	pkiMap["cert"] = configMapNew["pki"].(map[string]interface{})["cert"]
	pkiMap["key"] = configMapNew["pki"].(map[string]interface{})["key"]
	pkiMap["ca"] = configMapNew["pki"].(map[string]interface{})["ca"]

	configMapNew["pki"] = pkiMap

	// write the new config back to file

	configYml, err := yaml.Marshal(configMapNew)
	if err != nil {
		utils.Error("SlaveConfigSync: Error marshalling new config", err)
		return
	}

	err = ioutil.WriteFile(utils.CONFIGFOLDER + "nebula.yml", configYml, 0644)
	if err != nil {
		utils.Error("SlaveConfigSync: Error writing new config", err)
		return
	}

	utils.Log("SlaveConfigSync: Config resynced")

	// if the config change, restart
	if !reflect.DeepEqual(configMap, configMapNew) {
		RestartNebula()
		utils.RestartHTTPServer()
	}
}