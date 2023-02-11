package user

import (
	"net/http"
	"log"
	// "io"
	// "os"
	"encoding/json"

	"../utils" 
)

func UserRegister(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	if(req.Method == "GET") {
		// sedn form
	}
	if(req.Method == "POST") {
		// check origin

		// check password strength

		user := &utils.User{
			Nickname: req.FormValue("nickname"),
			Password: req.FormValue("password"),
			Role: (utils.Role)(req.FormValue("role")),
		}

		err := utils.Validate.Struct(user)

		if(err != nil) {
			log.Fatal(err)
		}

		req.ParseForm()
	}

	// return json object	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status": "OK",
	})
}