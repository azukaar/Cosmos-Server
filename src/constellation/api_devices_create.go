package constellation

import (
	"net/http"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type DeviceCreateRequestJSON struct {
	DeviceName string `json:"deviceName",validate:"required,min=3,max=32,alphanum"`
	IP string `json:"ip",validate:"required,ipv4"`
	PublicKey string `json:"publicKey",omitempty`
	
	// for devices only
	Nickname string `json:"nickname",validate:"max=32,alphanum",omitempty`
	
	// for lighthouse only
	IsLighthouse bool `json:"isLighthouse",omitempty`
	IsRelay bool `json:"isRelay",omitempty`
	PublicHostname string `json:"PublicHostname",omitempty`
	Port string `json:"port",omitempty`
}

func DeviceCreate(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "POST") {
		var request DeviceCreateRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("ConstellationDeviceCreation: Invalid User Request", err1)
			utils.HTTPError(w, "Device Creation Error",
				http.StatusInternalServerError, "DC001")
			return 
		}

		if !utils.FBL.LValid {
			utils.Error("ConstellationDeviceCreation: No valid licence found to use Constellation.", nil)
			utils.HTTPError(w, "Device Creation Error: No valid licence found to use Constellation.",
				http.StatusInternalServerError, "DC001")
			return
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("DeviceCreation: Invalid User Request", errV)
			utils.HTTPError(w, "Device Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "DC002")
			return 
		}
		
		nickname := utils.Sanitize(request.Nickname)
		deviceName := utils.Sanitize(request.DeviceName)
		APIKey := utils.GenerateRandomString(32)
		
		if utils.AdminOrItselfOnly(w, req, nickname) != nil {
			return
		}

		utils.Log("ConstellationDeviceCreation: Creating Device " + deviceName)

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
    defer closeDb()
		
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		device := utils.Device{}

		utils.Debug("ConstellationDeviceCreation: Creating Device " + deviceName)
		
		err2 := c.FindOne(nil, map[string]interface{}{
			"DeviceName": deviceName,
			"Blocked": false,
		}).Decode(&device)

		if err2 == mongo.ErrNoDocuments {

			cert, key, fingerprint, err := generateNebulaCert(deviceName, request.IP, request.PublicKey, false)

			if err != nil {
				utils.Error("DeviceCreation: Error while creating Device", err)
				utils.HTTPError(w, "Device Creation Error: " + err.Error(),
					http.StatusInternalServerError, "DC001")
				return
			}

			if request.IsLighthouse && request.Nickname != "" {
				utils.Error("DeviceCreation: Lighthouse cannot belong to a user", nil)
				utils.HTTPError(w, "Device Creation Error: Lighthouse cannot have a nickname",
					http.StatusInternalServerError, "DC003")
				return
			}

			if err != nil {
				utils.Error("DeviceCreation: Error while getting fingerprint", err)
				utils.HTTPError(w, "Device Creation Error: " + err.Error(),
					http.StatusInternalServerError, "DC007")
				return
			}

			_, err3 := c.InsertOne(nil, map[string]interface{}{
				"Nickname": nickname,
				"DeviceName": deviceName,
				"PublicKey": key,
				"IP": request.IP,
				"IsLighthouse": request.IsLighthouse,
				"IsRelay": request.IsRelay,
				"PublicHostname": request.PublicHostname,
				"Port": request.Port,
				"Fingerprint": fingerprint,
				"APIKey": APIKey,
				"Blocked": false,
			})

			if err3 != nil {
				utils.Error("DeviceCreation: Error while creating Device", err3)
				utils.HTTPError(w, "Device Creation Error: " + err3.Error(),
					http.StatusInternalServerError, "DC004")
				return 
			} 

			capki, err := getCApki()
			if err != nil {
				utils.Error("DeviceCreation: Error while reading ca.crt", err)
				utils.HTTPError(w, "Device Creation Error: " + err.Error(),
					http.StatusInternalServerError, "DC006")
				return
			}

			lightHousesList := []utils.ConstellationDevice{}
			if request.IsLighthouse {
				lightHousesList, err = GetAllLightHouses()
			}

			
			// read configYml from config/nebula.yml
			configYml, err := getYAMLClientConfig(deviceName, utils.CONFIGFOLDER + "nebula.yml", capki, cert, key, APIKey, utils.ConstellationDevice{
				Nickname: nickname,
				DeviceName: deviceName,
				PublicKey: key,
				IP: request.IP,
				IsLighthouse: request.IsLighthouse,
				IsRelay: request.IsRelay,
				PublicHostname: request.PublicHostname,
				Port: request.Port,
				APIKey: APIKey,
			}, true, true)


			if err != nil {
				utils.Error("DeviceCreation: Error while reading config", err)
				utils.HTTPError(w, "Device Creation Error: " + err.Error(),
					http.StatusInternalServerError, "DC005")
				return
			}

			utils.TriggerEvent(
				"cosmos.constellation.device.create",
				"Device created",
				"success",
				"",
				map[string]interface{}{
					"deviceName": deviceName,
					"nickname": nickname,
					"publicKey": key,
					"ip": request.IP,
			})

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data": map[string]interface{}{
					"Nickname": nickname,
					"DeviceName": deviceName,
					"PublicKey": key,
					"PrivateKey": cert,
					"IP": request.IP,
					"Config": configYml,
					"CA": capki,
					"IsLighthouse": request.IsLighthouse,
					"IsRelay": request.IsRelay,
					"PublicHostname": request.PublicHostname,
					"Port": request.Port,
					"LighthousesList": lightHousesList,
				},
			})
			
			go RestartNebula()
		} else if err2 == nil {
			utils.Error("DeviceCreation: Device already exists", nil)
			utils.HTTPError(w, "Device name already exists", http.StatusConflict, "DC002")
		  return 
		} else {
			utils.Error("DeviceCreation: Error while finding device", err2)
			utils.HTTPError(w, "Device Creation Error: " + err2.Error(),
				 http.StatusInternalServerError, "DC001")
			return 
		}
	} else {
		utils.Error("DeviceCreation: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}