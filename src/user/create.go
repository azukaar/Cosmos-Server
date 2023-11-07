package user

import (
	"net/http"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type CreateRequestJSON struct {
	Nickname string `validate:"required,min=3,max=32,alphanum"`
	Email string `validate:"omitempty,email"`
}

func UserCreate(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	} 

	if(req.Method == "POST") {
		var request CreateRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("UserCreation: Invalid User Request", err1)
			utils.HTTPError(w, "User Creation Error", 
				http.StatusInternalServerError, "UC001")
			return 
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("UserCreation: Invalid User Request", errV)
			utils.HTTPError(w, "User Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "UC003")
			return 
		}
		
		nickname := utils.Sanitize(request.Nickname)
		email := utils.Sanitize(request.Email)

		c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		user := utils.User{}

		utils.Debug("UserCreation: Creating user " + nickname)

		err2 := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
		}).Decode(&user)

		if err2 == mongo.ErrNoDocuments {
			RegisterKey := utils.GenerateRandomString(24)
			RegisterKeyExp := time.Now().Add(time.Hour * 24 * 7)

			_, err3 := c.InsertOne(nil, map[string]interface{}{
				"Nickname": nickname,
				"Email": email,
				"Password": "",
				"RegisterKey": RegisterKey,
				"RegisterKeyExp": RegisterKeyExp,
				"Role": utils.USER,
				"PasswordCycle": 0,
				"CreatedAt": time.Now(),
			})

			if err3 != nil {
				utils.Error("UserCreation: Error while creating user", err3)
				utils.HTTPError(w, "User Creation Error", 
					http.StatusInternalServerError, "UC001")
				return 
			} 
			
			utils.TriggerEvent(
				"cosmos.user.create",
				"User created",
				"success",
				"",
				map[string]interface{}{
					"nickname": nickname,
			})

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data": map[string]interface{}{
					"registerKey": RegisterKey,
					"registerKeyExp": RegisterKeyExp,
				},
			})
		} else if err2 == nil {
			utils.Error("UserCreation: User already exists", nil)
			utils.HTTPError(w, "User already exists", http.StatusConflict, "UC002")
		  return 
		} else {
			utils.Error("UserCreation: Error while finding user", err2)
			utils.HTTPError(w, "User Creation Error", http.StatusInternalServerError, "UC001")
			return 
		}
	} else {
		utils.Error("UserCreation: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}