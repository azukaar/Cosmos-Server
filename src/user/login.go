package user

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/azukaar/cosmos-server/src/utils"
)

type LoginRequestJSON struct {
	Nickname string `validate:"required,min=3,max=32,alphanum"`
	Password string `validate:"required,min=8,max=128,containsany=~!@#$%^&*()_+=-{[}]:;"'<>.?/,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
}

func UserLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		time.Sleep(time.Duration(rand.Float64()*2) * time.Second)

		if utils.IsLoggedIn(req) {
			utils.Error("UserLogin: User already logged ing", nil)
			utils.HTTPError(w, "User is already logged in", http.StatusUnauthorized, "UL002")
			return
		}

		var request LoginRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("UserLogin: Invalid User Request", err1)
			utils.HTTPError(w, "User Login Error", http.StatusInternalServerError, "UL001")
			return
		}

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
		defer closeDb()
		if errCo != nil {
			utils.Error("Database Connect", errCo)
			utils.HTTPError(w, "Database Error", http.StatusInternalServerError, "DB001")
			return
		}

		nickname := utils.Sanitize(request.Nickname)
		password := request.Password

		user := utils.User{}

		utils.Debug("UserLogin: Logging user " + nickname)

		err3 := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
		}).Decode(&user)

		if err3 == mongo.ErrNoDocuments {
			bcrypt.CompareHashAndPassword([]byte("$2a$14$4nzsVwEnR3.jEbMTME7kqeCo4gMgR/Tuk7ivNExvXjr73nKvLgHka"), []byte("dummyPassword"))
			utils.Error("UserLogin: User not found", err3)
			utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
			return
		} else if err3 != nil {
			bcrypt.CompareHashAndPassword([]byte("$2a$14$4nzsVwEnR3.jEbMTME7kqeCo4gMgR/Tuk7ivNExvXjr73nKvLgHka"), []byte("dummyPassword"))
			utils.Error("UserLogin: Error while finding user", err3)
			utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
			return
		} else if user.Password == "" {
			utils.Error("UserLogin: User not registered", nil)
			utils.HTTPError(w, "User not registered", http.StatusUnauthorized, "UL002")
			return
		} else {
			err2 := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err2 != nil {
				utils.Error("UserLogin: Encryption error", err2)
				utils.HTTPError(w, "User Logging Error", http.StatusUnauthorized, "UL001")
				return
			}

			if utils.IsEmailEnabled() && utils.IsNotifyLoginEmailEnabled() && user.Email != "" {
				SendLoginNotificationEmail(user.Nickname, user.Email)
			}

			SendUserToken(w, req, user, false)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
			})

			_, errE := c.UpdateOne(nil, map[string]interface{}{
				"Nickname": nickname,
			}, map[string]interface{}{
				"$set": map[string]interface{}{
					"LastLogin": time.Now(),
				},
			})

			if errE != nil {
				utils.Error("UserLogin: Error while updating user last login", errE)
			}
		}
	} else {
		utils.Error("UserLogin: Method not allowed"+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
