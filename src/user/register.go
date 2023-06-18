package user

import (
	"net/http"
	"math/rand"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"golang.org/x/crypto/bcrypt"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type RegisterRequestJSON struct {
	Nickname string `validate:"required,min=3,max=32,alphanum"`
	Password string `validate:"required,min=9,max=128,containsany=~!@#$%^&*()_+=-{[}]:;"'<>.?/,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
	RegisterKey string `validate:"required,min=1,max=512,alphanum"`
}

func UserRegister(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "POST") {
		time.Sleep(time.Duration(rand.Float64()*2)*time.Second)
		
		var request RegisterRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("UserRegister: Invalid User Request", err1)
			utils.HTTPError(w, "User Register Error", http.StatusInternalServerError, "UR001")
			return
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("UserRegister: Invalid User Request", errV)
			utils.HTTPError(w, "User Register Error: " + errV.Error(), http.StatusInternalServerError, "UR002")
			return
		}

		nickname := utils.Sanitize(request.Nickname)
		password := request.Password
		registerKey := request.RegisterKey

		utils.Debug("UserRegister: Registering user " + nickname)
				
		hashedPassword, err2 := bcrypt.GenerateFromPassword([]byte(password), 14)

		if err2 != nil {
			utils.Error("UserRegister: Encryption error", err2)
			utils.HTTPError(w, "User Register Error", http.StatusUnauthorized, "UR001")
			return
		}

		c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		user := utils.User{}

		err3 := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
			"RegisterKey": registerKey,
		}).Decode(&user)

		if err3 == mongo.ErrNoDocuments {
			utils.Error("UserRegister: User not found", err3)
			utils.HTTPError(w, "User Register Error", http.StatusInternalServerError, "UR001")
			return
		} else if err3 != nil {
			utils.Error("UserRegister: Error while finding user", err3)
			utils.HTTPError(w, "User Register Error", http.StatusInternalServerError, "UR001")
			return
		} else if user.RegisterKeyExp.Before(time.Now()) {
			utils.Error("UserRegister: Link expired", nil)
			utils.HTTPError(w, "User Register Error", http.StatusInternalServerError, "UR001")
			return
		} else {
			RegisteredAt := user.RegisteredAt
			if RegisteredAt.IsZero() {
				RegisteredAt = time.Now()
			}
			_, err4 := c.UpdateOne(nil, map[string]interface{}{
				"Nickname": nickname,
				"RegisterKey": registerKey,
			}, map[string]interface{}{
				"$set": map[string]interface{}{
					"Password": hashedPassword,
					"RegisterKey": "",
					"RegisterKeyExp": time.Time{},
					"RegisteredAt": RegisteredAt,
					"LastPasswordChangedAt": time.Now(),
					"PassowrdCycle": user.PasswordCycle + 1,
				},
			})

			if err4 != nil {
				utils.Error("UserRegister: Error while updating user", err4)
				utils.HTTPError(w, "User Register Error", http.StatusInternalServerError, "UR001")
				return
			}
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("UserRegister: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}