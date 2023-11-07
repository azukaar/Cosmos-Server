package user

import (
	"net/http"
	"encoding/json"
	"time"
	"math/rand"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type PasswordResetRequestJSON struct {
	Nickname string `validate:"required,min=3,max=32,alphanum"`
	Email string `validate:"required,min=3,max=32,alphanum"`
}

func ResetPassword(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "POST") {
		if !utils.IsEmailEnabled() {
			utils.HTTPError(w, "Email is not enabled", http.StatusInternalServerError, "PR007")
			return
		}

		time.Sleep(time.Duration(rand.Float64()*2)*time.Second)

		var request PasswordResetRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("PasswordReset: Invalid User Request", err1)
			utils.HTTPError(w, "User Send Invite Error", http.StatusInternalServerError, "PR001")
			return
		}

		nickname := utils.Sanitize(request.Nickname)

		utils.Debug("Sending password reset to: " + nickname)
		
		c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		user := utils.User{}

		err := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
			"Email": request.Email,
		}).Decode(&user)

		if err != nil {
			utils.Error("PasswordReset: Error while finding user", err)
			utils.HTTPError(w, "User Send Invite Error", http.StatusInternalServerError, "PR001")
			return
		} else {
			RegisterKeyExp := time.Now().Add(time.Hour * 24 * 7)
			RegisterKey := utils.GenerateRandomString(48)

			utils.Debug(RegisterKey)
			utils.Debug(RegisterKeyExp.String())

			_, err := c.UpdateOne(nil, map[string]interface{}{
				"Nickname": nickname,
			}, map[string]interface{}{
				"$set": map[string]interface{}{
					"RegisterKeyExp": RegisterKeyExp,
					"RegisterKey": RegisterKey,
				},
			})

			if err != nil {
				utils.Error("PasswordReset: Error while updating user", err)
				utils.HTTPError(w, "User Send Invite Error", http.StatusInternalServerError, "PR001")
				return
			}

			utils.Debug("Sending an email to " + user.Email)
			url := utils.GetServerURL() + ("cosmos-ui/register?t=1&nickname="+user.Nickname+"&key=" + RegisterKey)
			
			errEm := SendPasswordEmail(user.Nickname, user.Email, url) 

			if errEm != nil {
				utils.Error("PasswordReset: Error while sending email", errEm)
				utils.HTTPError(w, "User Send Invite Error", http.StatusInternalServerError, "PR002")
				return
			}

			utils.TriggerEvent(
				"cosmos.user.passwordreset",
				"Password reset sent",
				"success",
				"",
				map[string]interface{}{
					"nickname": user.Nickname,
			})

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
			})
		}
	} else {
		utils.Error("PasswordReset: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}