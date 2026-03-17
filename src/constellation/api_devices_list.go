package constellation

import (
	"net/http"
	"encoding/json"
		
	"github.com/azukaar/cosmos-server/src/utils" 
)

// DeviceList godoc
// @Summary List Constellation devices for the current user (or all if admin)
// @Tags constellation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/constellation/devices [get]
func DeviceList(w http.ResponseWriter, req *http.Request) {
	// Check for GET method
	if req.Method != "GET" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP002")
		return
	}

	if utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil {
		return
	}

	isAdmin := utils.HasPermission(req, utils.PERM_RESOURCES_READ)
	
	// Connect to the collection
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
    defer closeDb()
	if errCo != nil {
		utils.Error("Database Connect", errCo)
		utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
		return
	}
	
	var devices []utils.ConstellationDevice
	
	// Check if user is an admin
	if isAdmin {
		// If admin, get all devices
		cursor, err := c.Find(nil, map[string]interface{}{})
		defer cursor.Close(nil)
		if err != nil {
			utils.Error("DeviceList: Error fetching devices", err)
			utils.HTTPError(w, "Error fetching devices", http.StatusInternalServerError, "DL001")
			return
		}
		
		if err = cursor.All(nil, &devices); err != nil {
			utils.Error("DeviceList: Error decoding devices", err)
			utils.HTTPError(w, "Error decoding devices", http.StatusInternalServerError, "DL002")
			return
		}
	} else {
		// If not admin, get user's devices based on their nickname
		nickname := utils.GetAuthContext(req).Nickname
		cursor, err := c.Find(nil, map[string]interface{}{"Nickname": nickname})
		defer cursor.Close(nil)
		if err != nil {
			utils.Error("DeviceList: Error fetching devices", err)
			utils.HTTPError(w, "Error fetching devices", http.StatusInternalServerError, "DL003")
			return
		}
		
		if err = cursor.All(nil, &devices); err != nil {
			utils.Error("DeviceList: Error decoding devices", err)
			utils.HTTPError(w, "Error decoding devices", http.StatusInternalServerError, "DL004")
			return
		}
	}
	
	n, _ := GetCurrentDeviceName()

	// If list is empty, include the current device at least
	if (devices == nil || len(devices) == 0) && n != "" {
		currentDevice, err := GetCurrentDevice()
		if err == nil {
			devices = []utils.ConstellationDevice{currentDevice}
		}
	}

	// Respond with the list of devices
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "OK",
		"data": devices,
		"currentDeviceName": n,
	})
}
