package user

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils" 
)

type EditRequestJSON struct {
	Email string `validate:"email"`
}

func UserEdit(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	nickname := vars["nickname"]

	if utils.AdminOrItselfOnly(w, req, nickname) != nil {
		return
	} 

	if(req.Method == "PATCH") {
		var request EditRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("UserEdit: Invalid User Request", err1)
			utils.HTTPError(w, "User Edit Error", http.StatusInternalServerError, "UL001")
			return
		}

		// Validate request
		err2 := utils.Validate.Struct(request)
		if err2 != nil {
			utils.Error("UserEdit: Invalid User Request", err2)
			utils.HTTPError(w, "User request invalid: " + err2.Error(), http.StatusInternalServerError, "UL002")
			return
		}
		
		c := utils.GetCollection(utils.GetRootAppId(), "users")

		utils.Debug("UserEdit: Edit user " + nickname)

		toSet := map[string]interface{}{}
		if request.Email != "" {
			
			if utils.AdminOnly(w, req) != nil {
				return
			} 

			toSet["Email"] = request.Email
		}

		_, err := c.UpdateOne(nil, map[string]interface{}{
			"Nickname": nickname,
		}, map[string]interface{}{
			"$set": toSet,
		})

		if err != nil {
			utils.Error("UserEdit: Error while getting user", err)
			utils.HTTPError(w, "User Edit Error", http.StatusInternalServerError, "UE001")
			return
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("UserEdit: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}