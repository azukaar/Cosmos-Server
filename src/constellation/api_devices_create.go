package constellation

import (
	"net/http"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type DeviceCreateRequestJSON struct {
	DeviceName string `json:"deviceName",validate:"required,min=3,max=32,alphanum"`
	IP string `json:"ip",validate:"required,ipv4"`
	PublicKey string `json:"publicKey",omitempty`

	// for devices only
	Nickname string `json:"nickname",validate:"max=32,alphanum",omitempty`
	Invisible bool `json:"invisible",omitempty`

	// for lighthouse only
	IsLighthouse bool `json:"isLighthouse",omitempty`
	IsCosmosNode bool `json:"isCosmosNode",omitempty`
	IsRelay bool `json:"isRelay",omitempty`
	IsLoadBalancer bool `json:"isLoadBalancer",omitempty`
	IsExitNode bool `json:"isExitNode",omitempty`
	PublicHostname string `json:"PublicHostname",omitempty`
	Port string `json:"port",omitempty`

	// internal
	APIKey string `json:"-"`
}

func DeviceCreate_API(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "POST") {
		var request DeviceCreateRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("ConstellationDeviceCreation: Invalid User Request", err1)
			utils.HTTPError(w, "Device Creation Error",
				http.StatusInternalServerError, "DC001")
			return 
		}

		if !utils.FBL.LValid && !utils.FBL.IsCosmosNode {
			utils.Error("ConstellationDeviceCreation: No valid licence found to use Constellation.", nil)
			utils.HTTPError(w, "Device Creation Error: No valid licence found to use Constellation.",
				http.StatusInternalServerError, "DC001")
			return
		}

		nickname := utils.Sanitize(request.Nickname)

		if utils.AdminOrItselfOnly(w, req, nickname) != nil {
			return
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("DeviceCreation: Invalid User Request", errV)
			utils.HTTPError(w, "Device Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "DC002")
			return 
		}

		cert, key, _, request, err := DeviceCreate(request)
		if err != nil {
			utils.Error("DeviceCreation: Error creating device", err)
			utils.HTTPError(w, "Device Creation Error: " + err.Error(),
				http.StatusInternalServerError, "DC003")
			return
		}

		APIKey := request.APIKey
		deviceName := request.DeviceName
		
		capki, err := getCApki()
		if err != nil {
			utils.Error("DeviceCreation: Error while reading CA", err)
			utils.HTTPError(w, "Device Creation Error: " + err.Error(),
				http.StatusInternalServerError, "DC003")
			return
		}

		if err == nil {
			// read configYml from config/nebula.yml
			configYml, err := getYAMLClientConfig(deviceName, utils.CONFIGFOLDER + "nebula.yml", capki, cert, key, APIKey, utils.ConstellationDevice{
				Nickname: nickname,
				DeviceName: deviceName,
				PublicKey: key,
				IP: request.IP,
				IsLighthouse: request.IsLighthouse,
				IsCosmosNode: request.IsCosmosNode,
				IsRelay: request.IsRelay,
				IsExitNode: request.IsExitNode,
				IsLoadBalancer: request.IsLoadBalancer,
				PublicHostname: request.PublicHostname,
				Port: request.Port,
				APIKey: APIKey,
			}, true, true)


			lightHousesList := []utils.ConstellationDevice{}
			if request.IsLighthouse {
				lightHousesList, err = GetAllLightHouses()
			}

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
					"IsCosmosNode": request.IsCosmosNode,
					"IsLoadBalancer": request.IsLoadBalancer,
					"IsRelay": request.IsRelay,
					"IsExitNode": request.IsExitNode,
					"PublicHostname": request.PublicHostname,
					"Port": request.Port,
					"LighthousesList": lightHousesList,
					"Invisible": request.Invisible,
				},
			})
			
			utils.ResyncConstellationNodes()
			time.Sleep(2 * time.Second)
			go RestartNebula()
		} else {
			utils.Error("DeviceCreation: Error creating device", err)
			utils.HTTPError(w, "Device Creation Error: " + err.Error(),
				http.StatusInternalServerError, "DC004")
			return
		}
	} else {
		utils.Error("DeviceCreation: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func DeviceCreate(request DeviceCreateRequestJSON) (string, string, string, DeviceCreateRequestJSON, error) {
	nickname := utils.Sanitize(request.Nickname)
	deviceName := utils.Sanitize(request.DeviceName)
	APIKey := utils.GenerateRandomString(32)

	utils.Log("ConstellationDeviceCreation: Creating Device " + deviceName)

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()
	
	if errCo != nil {
		return "", "", "", DeviceCreateRequestJSON{}, errCo
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
			return "", "", "", DeviceCreateRequestJSON{}, err
		}

		// Cosmos nodes are also lighthouses
		if request.IsCosmosNode {
			request.IsLighthouse = true
		}

		// Check cosmos node and devices limit
		if request.IsCosmosNode {
			totalClientLimit := 10 * int64(utils.GetNumberUsers())

			count, errCount := c.CountDocuments(nil, map[string]interface{}{
				"IsCosmosNode": true,
				"Blocked": false,
			})
			if errCount != nil {
				return "", "", "", DeviceCreateRequestJSON{}, errCount
			}

			countDevices, errCountDevices := c.CountDocuments(nil, map[string]interface{}{
				"Blocked": false,
			})
			if errCountDevices != nil {
				return "", "", "", DeviceCreateRequestJSON{}, errCountDevices
			}

			if countDevices >= totalClientLimit {
				return "", "", "", DeviceCreateRequestJSON{}, errors.New("DeviceCreation: Device limit reached")
			}

			if count >= int64(utils.GetNumberCosmosNode()) {
				return "", "", "", DeviceCreateRequestJSON{}, errors.New("DeviceCreation: Cosmos Node limit reached")
			}
		}

		if request.IsLighthouse && request.Nickname != "" {
			return "", "", "", DeviceCreateRequestJSON{}, errors.New("DeviceCreation: Lighthouse cannot belong to a user")
		}

		if err != nil {
			return "", "", "", DeviceCreateRequestJSON{}, err
		}

		if (!request.IsLighthouse ) {
			request.IsRelay = false
			request.IsExitNode = false
			request.Invisible = false
			request.IsLoadBalancer = false
		}

		_, err3 := c.InsertOne(nil, map[string]interface{}{
			"Nickname": nickname,
			"DeviceName": deviceName,
			"PublicKey": key,
			"IP": request.IP,
			"IsLighthouse": request.IsLighthouse,
			"IsCosmosNode": request.IsCosmosNode,
			"IsRelay": request.IsRelay,
			"IsExitNode": request.IsExitNode,
			"PublicHostname": request.PublicHostname,
			"IsLoadBalancer": request.IsLoadBalancer,
			"Port": request.Port,
			"Fingerprint": fingerprint,
			"APIKey": APIKey,
			"Blocked": false,
			"Invisible": request.Invisible,
		})

		if err3 != nil {
			return "", "", "", DeviceCreateRequestJSON{}, err3
		} 

		request.Nickname = nickname
		request.DeviceName = deviceName
		request.PublicKey = key
		request.IsLoadBalancer = request.IsLoadBalancer
		request.IsExitNode = request.IsExitNode
		request.IsLoadBalancer = request.IsLoadBalancer
		request.Invisible = request.Invisible
		request.APIKey = APIKey	

		return cert, key, fingerprint, request, nil
	} else if err2 == nil {
		return "", "", "", DeviceCreateRequestJSON{}, errors.New("DeviceCreation: Device with this name already exists")
	} else {
		return "", "", "", DeviceCreateRequestJSON{}, err2
	}
}