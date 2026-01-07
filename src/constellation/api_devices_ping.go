package constellation

import (
	"net/http"
	"encoding/json"
	"os/exec"
	"runtime"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
)

func DevicePing(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP002")
		return
	}

	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	// Extract device ID from URL parameters
	vars := mux.Vars(req)
	deviceID := vars["id"]

	if deviceID == "" {
		utils.HTTPError(w, "Device ID is required", http.StatusBadRequest, "DP001")
		return
	}

	// Connect to the collection
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()
	if errCo != nil {
		utils.Error("Database Connect", errCo)
		utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
		return
	}

	var device utils.ConstellationDevice

	// Find the device by DeviceName
	err := c.FindOne(nil, map[string]interface{}{
		"DeviceName": deviceID,
	}).Decode(&device)

	if err != nil {
		utils.Error("DevicePing: Device not found", err)
		utils.HTTPError(w, "Device not found", http.StatusNotFound, "DP002")
		return
	}

	// Check if the device IP exists
	if device.IP == "" {
		utils.HTTPError(w, "Device has no IP address", http.StatusBadRequest, "DP003")
		return
	}

	// Ping the device IP
	pingResult := pingIP(device.IP)

	// Respond with the ping result
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data": map[string]interface{}{
			"deviceName": device.DeviceName,
			"ip":         device.IP,
			"reachable":  pingResult,
		},
	})
}

func pingIP(ip string) bool {
	var cmd *exec.Cmd

	// remove the CIDR suffix if present
	if len(ip) > 3 && ip[len(ip)-3] == '/' {
		ip = ip[:len(ip)-3]
	}

	// Use platform-specific ping command
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "10000", ip)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "10", ip)
	}

	output, err := cmd.CombinedOutput()

	utils.Debug("Ping - Pinging IP "+ip+" - Reachable: \nOutput:\n"+string(output))
	
	return err == nil
}
