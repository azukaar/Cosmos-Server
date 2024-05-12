package user

import (
	"net/http"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 
)


func Me(w http.ResponseWriter, req *http.Request) {
  if (req.Method == "GET") {
		if utils.LoggedInOnly(w, req) != nil {
			return
		}

		nickname := req.Header.Get("x-cosmos-user")

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
  	defer closeDb()
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

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
		user.Email = ""
		user.RegisterKey = ""
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": user,
		})
	} else {
		utils.Error("UserRoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}