package constellation

import (
	"net/http"
	"encoding/json"
		
	"github.com/azukaar/cosmos-server/src/utils" 
)

func DeviceList(w http.ResponseWriter, req *http.Request) {
	// Check for GET method
	if req.Method != "GET" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP002")
		return
	}

	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	isAdmin := utils.IsAdmin(req)
	
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
		nickname := req.Header.Get("x-cosmos-user")
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
	
	// Respond with the list of devices
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "OK",
		"data": devices,
	})
}
