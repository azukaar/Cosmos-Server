package user

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"

	"../utils" 
)

func UserDelete(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	nickname := vars["nickname"]

	if utils.AdminOrItselfOnly(w, req, nickname) != nil {
		return
	} 

	if(req.Method == "DELETE") {

		c := utils.GetCollection(utils.GetRootAppId(), "users")

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
	} else {
		utils.Error("UserDeletion: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}