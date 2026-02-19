package constellation

import (
	"net/http"
	"encoding/json"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

type DeviceEditRequestJSON struct {
	IsLighthouse   bool   `json:"isLighthouse"`
	IsRelay        bool   `json:"isRelay"`
	IsExitNode     bool   `json:"isExitNode"`
	IsLoadBalancer bool   `json:"isLoadBalancer"`
}

func DeviceEdit_API(w http.ResponseWriter, req *http.Request) {
	if utils.FBL.AgentMode {
		utils.Error("Constellation: Agents cannot manage devices. Use a manager server", nil)
		utils.HTTPError(w, "Constellation Error: Agents cannot manage devices. Use a manager server",
			http.StatusInternalServerError, "UC001")
		return
	}
	
	if req.Method != "POST" {
		utils.Error("DeviceEdit: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	if utils.AdminOnly(w, req) != nil {
		return
	}

	var request DeviceEditRequestJSON
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		utils.Error("DeviceEdit: Invalid request", err)
		utils.HTTPError(w, "Device Edit Error", http.StatusInternalServerError, "DE001")
		return
	}

	deviceName, err := GetCurrentDeviceName()
	if err != nil {
		utils.Error("DeviceEdit: Error getting current device name", err)
		utils.HTTPError(w, "Device Edit Error: "+err.Error(), http.StatusInternalServerError, "DE002")
		return
	}

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()

	if errCo != nil {
		utils.Error("Database Connect", errCo)
		utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
		return
	}

	// Non-lighthouses cannot be relay, exit, or load balancer
	if !request.IsLighthouse {
		request.IsRelay = false
		request.IsExitNode = false
		request.IsLoadBalancer = false
	}

	_, err = c.UpdateOne(nil, map[string]interface{}{
		"DeviceName": deviceName,
		"Blocked":    false,
	}, map[string]interface{}{
		"$set": map[string]interface{}{
			"IsLighthouse":   request.IsLighthouse,
			"IsRelay":        request.IsRelay,
			"IsExitNode":     request.IsExitNode,
			"IsLoadBalancer": request.IsLoadBalancer,
		},
	})

	if err != nil {
		utils.Error("DeviceEdit: Error updating device", err)
		utils.HTTPError(w, "Device Edit Error: "+err.Error(), http.StatusInternalServerError, "DE003")
		return
	}

	utils.TriggerEvent(
		"cosmos.constellation.device.edit",
		"Device edited",
		"success",
		"",
		map[string]interface{}{
			"deviceName": deviceName,
		})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})

	go func() {
		SendNewDBSyncMessage()
		time.Sleep(2 * time.Second)
		RestartNebula()
	}()
}
