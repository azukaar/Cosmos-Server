package constellation

import (
	"net/http"
	"encoding/json"
	"strconv"

	"github.com/azukaar/cosmos-server/src/utils"
)

func API_GetNextIP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		utils.Error("API_GetNextIP: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()

	if errCo != nil {
		utils.Error("API_GetNextIP: Database Connect", errCo)
		utils.HTTPError(w, "Database error", http.StatusInternalServerError, "DB001")
		return
	}

	// Get all used IPs
	cursor, err := c.Find(nil, map[string]interface{}{
		"Blocked": false,
	})
	if err != nil {
		utils.Error("API_GetNextIP: Error fetching devices", err)
		utils.HTTPError(w, "Error fetching devices", http.StatusInternalServerError, "GIP001")
		return
	}
	defer cursor.Close(nil)

	usedIPs := make(map[string]bool)
	var devices []utils.ConstellationDevice
	if err = cursor.All(nil, &devices); err != nil {
		utils.Error("API_GetNextIP: Error decoding devices", err)
		utils.HTTPError(w, "Error decoding devices", http.StatusInternalServerError, "GIP002")
		return
	}

	for _, device := range devices {
		usedIPs[device.IP] = true
	}

	// Find next available IP starting from 192.168.201.2
	// Skip 192.168.201.1 as it's typically reserved for the first lighthouse
	nextIP := ""
	for j := 201; j <= 255; j++ {
		for i := 2; i <= 254; i++ {
			ip := "192.168." + strconv.Itoa(j) + "." + strconv.Itoa(i)
			if !usedIPs[ip] {
				nextIP = ip
				break
			}
		}
		if nextIP != "" {
			break
		}
	}

	if nextIP == "" {
		utils.Error("API_GetNextIP: No available IPs", nil)
		utils.HTTPError(w, "No available IP addresses", http.StatusInternalServerError, "GIP003")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   nextIP,
	})
}
