package constellation

import (
	"net/http"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	
	"github.com/azukaar/cosmos-server/src/utils" 
)

type DeviceCreateRequestJSON struct {
	Nickname string `json:"nickname",validate:"required,min=3,max=32,alphanum"`
	DeviceName string `json:"deviceName",validate:"required,min=3,max=32,alphanum"`
	IP string `json:"ip",validate:"required,ipv4"`
	PublicKey string `json:"publicKey",omitempty`
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

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("DeviceCreation: Invalid User Request", errV)
			utils.HTTPError(w, "Device Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "DC002")
			return 
		}
		
		nickname := utils.Sanitize(request.Nickname)
		deviceName := utils.Sanitize(request.DeviceName)
		
		if utils.AdminOrItselfOnly(w, req, nickname) != nil {
			return
		}

		c, errCo := utils.GetCollection(utils.GetRootAppId(), "devices")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		device := utils.Device{}

		utils.Debug("ConstellationDeviceCreation: Creating Device " + deviceName)
		
		err2 := c.FindOne(nil, map[string]interface{}{
			"DeviceName": deviceName,
		}).Decode(&device)

		if err2 == mongo.ErrNoDocuments {
			cert, key, err := generateNebulaCert(deviceName, request.IP, false)

			if err != nil {
				utils.Error("DeviceCreation: Error while creating Device", err)
				utils.HTTPError(w, "Device Creation Error: " + err.Error(),
					http.StatusInternalServerError, "DC001")
				return
			}

			_, err3 := c.InsertOne(nil, map[string]interface{}{
				"Nickname": nickname,
				"DeviceName": deviceName,
				"PublicKey": cert,
				"PrivateKey": key,
				"IP": request.IP,
			})

			if err3 != nil {
				utils.Error("DeviceCreation: Error while creating Device", err3)
				utils.HTTPError(w, "Device Creation Error: " + err.Error(),
					http.StatusInternalServerError, "DC004")
				return 
			} 

			// read configYml from config/nebula.yml
			configYml, err := getYAMLClientConfig(deviceName, utils.CONFIGFOLDER + "nebula.yml")
			if err != nil {
				utils.Error("DeviceCreation: Error while reading config", err)
				utils.HTTPError(w, "Device Creation Error: " + err.Error(),
					http.StatusInternalServerError, "DC005")
				return
			}

			capki, err := getCApki()
			if err != nil {
				utils.Error("DeviceCreation: Error while reading ca.crt", err)
				utils.HTTPError(w, "Device Creation Error: " + err.Error(),
					http.StatusInternalServerError, "DC006")
				return
			}
			
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data": map[string]interface{}{
					"Nickname": nickname,
					"DeviceName": deviceName,
					"PublicKey": cert,
					"PrivateKey": key,
					"IP": request.IP,
					"Config": configYml,
					"CA": capki,
				},
			})
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