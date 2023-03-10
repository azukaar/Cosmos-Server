package user

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"../utils" 
)

func ConfigApiGet(w http.ResponseWriter, req *http.Request) {
	if AdminOnly(w, req) != nil {
		return
	} 

	if(req.Method == "GET") {
		config := utils.GetBaseMainConfig()

		// delete AuthPrivateKey and TLSKey
		config.AuthPrivateKey = ""
		config.TLSKey = ""

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
