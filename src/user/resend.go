package user

import (
	"net/http"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type InviteRequestJSON struct {
	Nickname string `validate:"required,min=3,max=32,alphanum"`
	FormType string
}

func UserResendInviteLink(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "POST") {
		var request InviteRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("UserInvite: Invalid User Request", err1)
			utils.HTTPError(w, "User Send Invite Error", http.StatusInternalServerError, "US001")
			return
		}

		nickname := utils.Sanitize(request.Nickname)

		if utils.AdminOnly(w, req) != nil {
			return
		}

		utils.Debug("Re-Sending an invite to " + nickname)
		
		c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		user := utils.User{}

		err := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
		}).Decode(&user)

		if err == mongo.ErrNoDocuments {
			utils.Error("UserInvite: User not found", err)
			utils.HTTPError(w, "User Send Invite Error", http.StatusNotFound, "US001")
			return
		} else if err != nil {
			utils.Error("UserInvite: Error while finding user", err)
			utils.HTTPError(w, "User Send Invite Error", http.StatusInternalServerError, "US001")
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
				utils.Error("UserInvite: Error while updating user", err)
				utils.HTTPError(w, "User Send Invite Error", http.StatusInternalServerError, "US001")
				return
			}

			emailWasSent := false

			if utils.IsEmailEnabled() && user.Email != "" {
				utils.Debug("Sending an email to " + user.Email)
				url := utils.GetServerURL() + ("cosmos-ui/register?t="+request.FormType+"&nickname="+user.Nickname+"&key=" + RegisterKey)
				
				var errEm error 
				
				if request.FormType == "2" {
					errEm = SendInviteEmail(user.Nickname, user.Email, url)
				} else {
					errEm = SendAdminPasswordEmail(user.Nickname, user.Email, url)
				}

				if errEm != nil {
					utils.Error("UserInvite: Error while sending email", errEm)
				}
				emailWasSent = true
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data": map[string]interface{}{
					"registerKey": RegisterKey,
					"registerKeyExp": RegisterKeyExp,
					"emailWasSent": emailWasSent,
				},
			})
		}
	} else {
		utils.Error("UserInvite: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}