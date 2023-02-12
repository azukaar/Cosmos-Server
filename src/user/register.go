package user

import (
	"net/http"
	"log"
	"math/rand"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"golang.org/x/crypto/bcrypt"

	"../utils" 
)

func UserRegister(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	if(req.Method == "POST") {
		time.Sleep(time.Duration(rand.Float64()*2)*time.Second)

		nickname := req.FormValue("nickname")
	  registerKey := req.FormValue("registerKey")
	  password := req.FormValue("password")

		err := bcrypt.CompareHashAndPassword([]byte(utils.GetHash()), []byte(password))

		if err != nil {
			log.Println("UserRegister: Encryption error")
			http.Error(w, "User Register Error", http.StatusUnauthorized)
		}

		c := utils.GetCollection(utils.GetRootAppId(), "users")

		user := utils.User{}

		err = c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
			"RegisterKey": registerKey,
			"Password": "",
		}).Decode(&user)

		if err == mongo.ErrNoDocuments {
			log.Println("UserRegister: User not found")
			http.Error(w, "User Register Error", http.StatusNotFound)
		} else if !user.RegisterKeyExp.Before(time.Now()) {
			log.Println("UserRegister: Link expired")
			http.Error(w, "User Register Error", http.StatusNotFound)
		} else if err != nil {
			log.Println("UserRegister: Error while finding user")
			http.Error(w, "User Register Error", http.StatusInternalServerError)
		} else {
			_, err := c.UpdateOne(nil, map[string]interface{}{
				"Nickname": nickname,
				"RegisterKey": registerKey,
				"Password": "",
			}, map[string]interface{}{
				"Password": password,
				"RegisterKey": "",
				"RegisterKeyExp": time.Time{},
			})

			if err != nil {
				log.Println("UserRegister: Error while updating user")
				http.Error(w, "User Register Error", http.StatusInternalServerError)
			}
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status": "OK",
		})
	} else {
		log.Println("UserRegister: Method not allowed" + req.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}