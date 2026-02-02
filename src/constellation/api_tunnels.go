package constellation

import (
	"net/http"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils"
)

func TunnelList(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP002")
		return
	}

	if utils.AdminOnly(w, req) != nil {
		return
	}

	tunnels := GetLocalTunnelCache()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   tunnels,
	})
}
