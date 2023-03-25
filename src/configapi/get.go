package configapi

import (
	"net/http"
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils" 
)

func ConfigApiGet(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	} 

	if(req.Method == "GET") {
		config := utils.GetBaseMainConfig()

		// delete AuthPrivateKey and TLSKey
		config.HTTPConfig.AuthPrivateKey = ""
		config.HTTPConfig.TLSKey = ""

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": config,
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
