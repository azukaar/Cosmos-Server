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
		utils.RestartServer()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK"
		})
	} else {
		utils.Error("Restart: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
