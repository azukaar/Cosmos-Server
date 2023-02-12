package user

import (
	"net/http"
	"log"
	// "io"
	// "os"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	// "golang.org/x/crypto/bcrypt"

	"../utils" 
)

func UserCreate(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	if(req.Method == "POST") {
		nickname := req.FormValue("nickname")

		c := utils.GetCollection(utils.GetRootAppId(), "users")

		user := utils.User{}

		err := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
		}).Decode(&user)

		if err != mongo.ErrNoDocuments {
			log.Println("UserCreation: User already exists")
			http.Error(w, "User Creation Error", http.StatusNotFound)
		} else if err != nil {
			log.Println("UserCreation: Error while finding user")
			http.Error(w, "User Creation Error", http.StatusInternalServerError)
		} else {

			RegisterKey := utils.GenerateRandomString(24)
			RegisterKeyExp := time.Now().Add(time.Hour * 24 * 7)

			_, err := c.InsertOne(nil, map[string]interface{}{
				"Nickname": nickname,
				"Password": "",
				"RegisterKey": RegisterKey,
				"RegisterKeyExp": RegisterKeyExp,
				"Role": utils.USER,
			})

			if err != nil {
				log.Println("UserCreation: Error while creating user")
				http.Error(w, "User Creation Error", http.StatusInternalServerError)
			} 
			
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Status": "OK",
				"RegisterKey": RegisterKey,
				"RegisterKeyExp": RegisterKeyExp,
			})
		}
	} else {
		log.Println("UserCreation: Method not allowed" + req.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}