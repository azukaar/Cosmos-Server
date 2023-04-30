package user

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/pquerna/otp/totp"
)

type User2FACheckRequest struct {
	Token string
}

func Check2FA(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInWeakOnly(w, req) != nil {
		return
	}

	nickname := req.Header.Get("x-cosmos-user")
	
	var request User2FACheckRequest
	errD := json.NewDecoder(req.Body).Decode(&request)
	if errD != nil {
		utils.Error("2FA Error: Invalid User Request", errD)
		utils.HTTPError(w, "2FA Error", http.StatusInternalServerError, "2FA001")
		return
	}

	c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
	if errCo != nil {
		utils.Error("Database Connect", errCo)
		utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
		return
	}

	userInBase := utils.User{}

	err := c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&userInBase)

	if err != nil {
		utils.Error("UserGet: Error while getting user", err)
		utils.HTTPError(w, "User Get Error", http.StatusInternalServerError, "2FA002")
		return
	}

	if(userInBase.MFAKey == "") {
		utils.Error("2FA: User " + nickname + " has no key", nil)
		utils.HTTPError(w, "2FA Error", http.StatusInternalServerError, "2FA003")
		return
	}

	valid := totp.Validate(request.Token, userInBase.MFAKey)

	if valid {
		utils.Log("2FA: User " + nickname + " has valid token")

		if(!userInBase.Was2FAVerified) {
			toSet := map[string]interface{}{
				"Was2FAVerified": true,
			}

			_, err = c.UpdateOne(nil, map[string]interface{}{
				"Nickname": nickname,
			}, map[string]interface{}{
				"$set": toSet,
			})
		
			if err != nil {
				utils.Error("2FA: Cannot update user", err)
				utils.HTTPError(w, "2FA Error", http.StatusInternalServerError, "2FA004")
				return
			}
		}

		SendUserToken(w, userInBase, true)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("2FA: User " + nickname + " has invalid token", nil)
		utils.HTTPError(w, "2FA Error", http.StatusInternalServerError, "2FA005")
		return
	}
}