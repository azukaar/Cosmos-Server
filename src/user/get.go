package user

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils" 
)

func UserGet(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	nickname := utils.Sanitize(vars["nickname"])

	if nickname == "" && req.Header.Get("x-cosmos-user") != "" {
		nickname = req.Header.Get("x-cosmos-user")
	}
	
	if utils.AdminOrItselfOnly(w, req, nickname) != nil {
		return
	}

	if(req.Method == "GET") {

		c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		utils.Debug("UserGet: Get user " + nickname)

		user := utils.User{}

		err := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
		}).Decode(&user)

		if err != nil {
			utils.Error("UserGet: Error while getting user", err)
			utils.HTTPError(w, "User Get Error", http.StatusInternalServerError, "UD001")
			return
		}

		user.Link = "/api/user/" + user.Nickname
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": user,
		})
	} else {
		utils.Error("UserGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}