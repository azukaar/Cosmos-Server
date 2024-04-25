package constellation

import (
	"net/http"
	"encoding/json"
	
	"github.com/azukaar/cosmos-server/src/utils" 
)

type DeviceBlockRequestJSON struct {
	Nickname string `json:"nickname",validate:"required,min=3,max=32,alphanum"`
	DeviceName string `json:"deviceName",validate:"required,min=3,max=32,alphanum"`
  Block bool `json:"block",omitempty`
}

func DeviceBlock(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "POST") {
		var request DeviceBlockRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("ConstellationDeviceBlocking: Invalid User Request", err1)
			utils.HTTPError(w, "Device Creation Error",
				http.StatusInternalServerError, "DB001")
			return 
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("DeviceBlocking: Invalid User Request", errV)
			utils.HTTPError(w, "Device Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "DB002")
			return 
		}
		
		nickname := utils.Sanitize(request.Nickname)
		deviceName := utils.Sanitize(request.DeviceName)
		
		if utils.AdminOrItselfOnly(w, req, nickname) != nil {
			return
		}

		utils.Log("ConstellationDeviceBlocking: Blocking Device " + deviceName)

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
  	defer closeDb()
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		device := utils.Device{}

		utils.Debug("ConstellationDeviceBlocking: Blocking Device " + deviceName)
		
		err2 := c.FindOne(nil, map[string]interface{}{
			"DeviceName": deviceName,
			"Nickname": nickname,
			"Blocked": false,
		}).Decode(&device)

		if err2 == nil {
			utils.Debug("ConstellationDeviceBlocking: Found Device " + deviceName)

			_, err3 := c.UpdateOne(nil, map[string]interface{}{
				"DeviceName": deviceName,
				"Nickname": nickname,
			}, map[string]interface{}{
				"$set": map[string]interface{}{
					"Blocked": request.Block,
				},
			})

			if err3 != nil {
				utils.Error("DeviceBlocking: Error while updating device", err3)
				utils.HTTPError(w, "Device Creation Error: " + err3.Error(),
					 http.StatusInternalServerError, "DB001")
				return
			}

			if request.Block {
				utils.Log("ConstellationDeviceBlocking: Device " + deviceName + " blocked")
			} else {
				utils.Log("ConstellationDeviceBlocking: Device " + deviceName + " unblocked")
			}
			
			RestartNebula()
		} else {
			utils.Error("DeviceBlocking: Error while finding device", err2)
			utils.HTTPError(w, "Device Creation Error: " + err2.Error(),
				 http.StatusInternalServerError, "DB001")
			return 
		}

		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("DeviceBlocking: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}