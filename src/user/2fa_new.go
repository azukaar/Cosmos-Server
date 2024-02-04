package user

import (
	"encoding/json"
	"net/http"
	"time"
	"math/rand"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/pquerna/otp/totp"
)

func New2FA(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInWeakOnly(w, req) != nil {
		return
	}
	time.Sleep(time.Duration(rand.Float64()*2)*time.Second)

	nickname := req.Header.Get("x-cosmos-user")

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Cosmos " + utils.GetMainConfig().HTTPConfig.Hostname,
		AccountName: nickname,
	})

	if err != nil {
		utils.Error("2FA: Cannot generate key", err)
		utils.HTTPError(w, "2FA Error", http.StatusInternalServerError, "2FA001")
		return
	}

	utils.Log("2FA: New key generated for " + nickname)
	
	toSet := map[string]interface{}{
		"MFAKey": key.Secret(),
		"Was2FAVerified": false,
	}

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
  defer closeDb()
	if errCo != nil {
		utils.Error("Database Connect", errCo)
		utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
		return
	}

	userInBase := utils.User{}

	err = c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&userInBase)

	if err != nil {
		utils.Error("UserGet: Error while getting user", err)
		utils.HTTPError(w, "User Get Error", http.StatusInternalServerError, "UD001")
		return
	}

	if(userInBase.MFAKey != "" && userInBase.Was2FAVerified) {
		if utils.LoggedInOnly(w, req) != nil {
			return
		}
	}

	_, err = c.UpdateOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}, map[string]interface{}{
		"$set": toSet,
	})

	if err != nil {
		utils.Error("2FA: Cannot update user", err)
		utils.HTTPError(w, "2FA Error", http.StatusInternalServerError, "2FA002")
		return
	}

	utils.Log("2FA: User " + nickname + " updated")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data": map[string]interface{}{
			"key": key.URL(),
		},
	})
}