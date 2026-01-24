package constellation

import (
	"net/http"
	"encoding/json"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
)

// PublicDeviceInfo represents the limited device information exposed to the public API
type PublicDeviceInfo struct {
	DeviceID   string `json:"id"`
	DeviceName string `json:"name"`
	User       string `json:"user"`
	IP         string `json:"ip"`
	IsLighthouse bool `json:"isLighthouse"`
	IsCosmosNode bool `json:"isCosmosNode"`
	IsRelay bool `json:"isRelay"`
	IsExitNode bool `json:"isExitNode"`
	PublicHostname string `json:"publicHostname"`
	Port string `json:"port"`
}

func DevicePublicList(w http.ResponseWriter, req *http.Request) {
	// Check for GET method
	if req.Method != "GET" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP002")
		return
	}

	// Get authorization header
	auth := req.Header.Get("Authorization")
	if auth == "" {
		http.Error(w, "Unauthorized [1]", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " from auth header
	auth = strings.Replace(auth, "Bearer ", "", 1)

	// Connect to the collection
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()
	if errCo != nil {
		utils.Error("Database Connect", errCo)
		utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
		return
	}

	utils.Log("DevicePublicList: Fetching devices with API key")

	// Find all non-blocked devices that match the API key
	cursor, err := c.Find(nil, map[string]interface{}{
		"Blocked": false,
		"Invisible": false,
	})

	defer cursor.Close(nil)
	
	if err != nil {
		utils.Error("DevicePublicList: Error fetching devices", err)
		utils.HTTPError(w, "Error fetching devices", http.StatusInternalServerError, "DPL001")
		return
	}

	var devices []utils.ConstellationDevice
	if err = cursor.All(nil, &devices); err != nil {
		utils.Error("DevicePublicList: Error decoding devices", err)
		utils.HTTPError(w, "Error decoding devices", http.StatusInternalServerError, "DPL002")
		return
	}

	// Always add the cosmos lighthouse device
	config := utils.GetMainConfig()
	cosmosDevice := utils.ConstellationDevice{
		DeviceName: "cosmos",
		Nickname:   "cosmos",
		IP:         "192.168.201.1",
		IsLighthouse: true,
		IsCosmosNode: true,
		IsRelay: config.ConstellationConfig.IsRelayNode,
		IsExitNode: config.ConstellationConfig.IsExitNode,
		PublicHostname: config.ConstellationConfig.ConstellationHostname,
		Port: "4242",
	}
	devices = append([]utils.ConstellationDevice{cosmosDevice}, devices...)

	// Convert to public device info with limited fields
	publicDevices := make([]PublicDeviceInfo, len(devices))
	for i, device := range devices {
		publicDevices[i] = PublicDeviceInfo{
			DeviceID:   device.DeviceName,
			DeviceName: device.DeviceName,
			User:       device.Nickname,
			IP:         cleanIp(device.IP),
			IsLighthouse: device.IsLighthouse,
			IsCosmosNode: device.IsCosmosNode,
			IsRelay: device.IsRelay,
			IsExitNode: device.IsExitNode,
			PublicHostname: device.PublicHostname,
			Port: device.Port,
		}
	}

	// Respond with the list of public device info
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   publicDevices,
	})
}
