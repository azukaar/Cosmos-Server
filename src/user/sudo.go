package user

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/azukaar/cosmos-server/src/utils"
)

type SudoRequestJSON struct {
	Password string `validate:"required,min=8,max=128,containsany=~!@#$%^&*()_+=-{[}]:;"'<>.?/,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
}

func UserSudo(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		time.Sleep(time.Duration(rand.Float64()*2) * time.Second)

		if utils.LoggedInOnly(w, req) != nil {
			return
		}
		
		userNickname := req.Header.Get("x-cosmos-user")
		role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
		userRole, _ := strconv.Atoi(req.Header.Get("x-cosmos-user-role"))

		if role < utils.USER || userRole < utils.ADMIN {
			utils.Error("UserSudo: User cannot sudo", nil)
			utils.HTTPError(w, "User cannot sudo", http.StatusUnauthorized, "HTTP005")
			return
		}

		var request SudoRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("UserSudo: Invalid User Request", err1)
			utils.HTTPError(w, "User Sudo Error", http.StatusInternalServerError, "UL001")
			return
		}

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
		defer closeDb()
		if errCo != nil {
			utils.Error("Database Connect", errCo)
			utils.HTTPError(w, "Database Error", http.StatusInternalServerError, "DB001")
			return
		}

		nickname := utils.Sanitize(userNickname)
		password := request.Password

		user := utils.User{}

		utils.Debug("UserSudo: Logging user " + nickname)

		err3 := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
		}).Decode(&user)

		if err3 == mongo.ErrNoDocuments {
			bcrypt.CompareHashAndPassword([]byte("$2a$14$4nzsVwEnR3.jEbMTME7kqeCo4gMgR/Tuk7ivNExvXjr73nKvLgHka"), []byte("dummyPassword"))
			utils.Error("UserSudo: User not found", err3)
			utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
			return
		} else if err3 != nil {
			bcrypt.CompareHashAndPassword([]byte("$2a$14$4nzsVwEnR3.jEbMTME7kqeCo4gMgR/Tuk7ivNExvXjr73nKvLgHka"), []byte("dummyPassword"))
			utils.Error("UserSudo: Error while finding user", err3)
			utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
			return
		} else if user.Password == "" {
			utils.Error("UserSudo: User not registered", nil)
			utils.HTTPError(w, "User not registered", http.StatusInternalServerError, "UL002")
			return
		} else {
			err2 := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err2 != nil {
				utils.Error("UserSudo: Encryption error", err2)
				utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
				return
			}

			SendUserToken(w, req, user, true, user.Role)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
			})
		}
	} else {
		utils.Error("UserLogin: Method not allowed"+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
