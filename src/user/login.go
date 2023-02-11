package user

import (
	"net/http"
	"log"
	"encoding/json"

	"../utils" 
)

func UserLogin(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	if(req.Method == "POST") {
		nickname := req.FormValue("nickname")
	  password := req.FormValue("password")

		log.Println("UserLogin: nickname: " + nickname)
		log.Println("UserLogin: password: " + password) // im just testing ok dont panic
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status": "OK",
		})
	} else {
		log.Println("UserLogin: Method not allowed" + req.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}