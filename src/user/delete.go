package user

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"

	"github.com/azukaar/cosmos-server/src/utils" 
)

func UserDelete(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	nickname := vars["nickname"]

	if utils.AdminOrItselfOnly(w, req, nickname) != nil {
		return
	} 

	if(req.Method == "DELETE") {

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
  defer closeDb()
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		utils.Debug("UserDeletion: Deleting user " + nickname)

		_, err := c.DeleteOne(nil, map[string]interface{}{
			"Nickname": nickname,
		})

		if err != nil {
			utils.Error("UserDeletion: Error while deleting user", err)
			utils.HTTPError(w, "User Deletion Error", http.StatusInternalServerError, "UD001")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
		
		go utils.ResyncConstellationNodes()
	} else {
		utils.Error("UserDeletion: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}