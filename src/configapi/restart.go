package configapi

import (
	"net/http"
	"encoding/json"    
	"github.com/azukaar/cosmos-server/src/utils" 
)

// ConfigApiRestart godoc
// @Summary Restart the server
// @Description Triggers a graceful server restart
// @Tags config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/restart [get]
func ConfigApiRestart(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN) != nil {
		return
	} 

	if(req.Method == "GET") {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
		utils.RestartServer(0)
	} else {
		utils.Error("Restart: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
