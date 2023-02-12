package user

import (
	"net/http"
	"log"
	"math/rand"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt"

	"../utils" 
)

func UserLogin(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	if(req.Method == "POST") {
		time.Sleep(time.Duration(rand.Float64()*2)*time.Second)

		nickname := req.FormValue("nickname")
	  password := req.FormValue("password")

		err := bcrypt.CompareHashAndPassword([]byte(utils.GetHash()), []byte(password))

		if err != nil {
			log.Println("UserLogin: Encryption error")
			http.Error(w, "User Logging Error", http.StatusUnauthorized)
		}

		c := utils.GetCollection(utils.GetRootAppId(), "users")

		err = c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
			"Password": password,
		}).Decode(&utils.User{})

		if err == mongo.ErrNoDocuments {
			log.Println("UserLogin: User not found")
			http.Error(w, "User Logging Error", http.StatusNotFound)
		} else if err != nil {
			log.Println("UserLogin: Error while finding user")
			http.Error(w, "User Logging Error", http.StatusInternalServerError)
		} else {
			token := jwt.New(jwt.SigningMethodEdDSA)
			claims := token.Claims.(jwt.MapClaims)
			claims["exp"] = time.Now().Add(30 * 24 * time.Hour)
			claims["authorized"] = true
			claims["nickname"] = nickname

			tokenString, err := token.SignedString(utils.GetPrivateAuthKey())

			if err != nil {
				log.Println("UserLogin: Error while signing token")
				http.Error(w, "User Logging Error", http.StatusInternalServerError)
			}

			expiration := time.Now().Add(30 * 24 * time.Hour)

    	cookie := http.Cookie{
				Name: "jwttoken",
				Value: tokenString,
				Expires: expiration,
			}

			http.SetCookie(w, &cookie)
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status": "OK",
		})
	} else {
		log.Println("UserLogin: Method not allowed" + req.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
