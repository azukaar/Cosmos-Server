package constellation

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
)

// getNextAvailableIP finds the next available IP in the given CIDR range
// that is not in the usedIPs map. Returns empty string if no IP is available.
// The first IP (network address) and last IP (broadcast) are skipped,
// as well as .1 which is typically reserved for the gateway/lighthouse.
func getNextAvailableIP(usedIPs map[string]bool, cidr string) string {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return ""
	}

	// Start from the first usable IP in the range (skip network address)
	ip := make(net.IP, len(ipnet.IP))
	copy(ip, ipnet.IP)
	ip = ip.To4()
	if ip == nil {
		return ""
	}

	// Iterate through all IPs in the range
	for ipnet.Contains(ip) {
		// Skip network address (.0) and broadcast (.255)
		lastOctet := ip[3]
		if lastOctet != 0 && lastOctet != 255 {
			ipStr := ip.String()
			if !usedIPs[ipStr] {
				return ipStr
			}
		}
		incrementIP(ip)
	}

	return ""
}

// incrementIP increments an IP address by 1
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// GetNextAvailableIP fetches all used IPs from the database and returns the next available IP in the given CIDR range
func GetNextAvailableIP(cidr string) string {
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()

	if errCo != nil {
		utils.Error("GetNextAvailableIP: Database Connect", errCo)
		return ""
	}

	cursor, err := c.Find(nil, map[string]interface{}{
		"Blocked": false,
	})
	if err != nil {
		utils.Error("GetNextAvailableIP: Error fetching devices", err)
		return ""
	}
	defer cursor.Close(nil)

	usedIPs := make(map[string]bool)
	var devices []utils.ConstellationDevice
	if err = cursor.All(nil, &devices); err != nil {
		utils.Error("GetNextAvailableIP: Error decoding devices", err)
		return ""
	}

	for _, device := range devices {
		usedIPs[device.IP] = true
	}

	return getNextAvailableIP(usedIPs, cidr)
}

func API_GetNextIP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		utils.Error("API_GetNextIP: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	cidr := utils.GetMainConfig().ConstellationConfig.IPRange
	if cidr == "" {
		utils.Error("API_GetNextIP: CIDR range not configured", nil)
		utils.HTTPError(w, "CIDR range not configured", http.StatusInternalServerError, "GIP004")
		return
	}

	nextIP := GetNextAvailableIP(cidr)

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
