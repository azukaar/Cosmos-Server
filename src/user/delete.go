package user

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"

	"github.com/azukaar/cosmos-server/src/utils" 
)

// UserDelete godoc
// @Summary Delete a user
// @Description Deletes a user account by nickname
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param nickname path string true "User nickname"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/users/{nickname} [delete]
func UserDelete(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	nickname := utils.Sanitize(vars["nickname"])

	if utils.CheckPermissionsOrSelf(w, req, nickname, utils.PERM_USERS) != nil {
		return
	} 

	if utils.FBL.AgentMode {
		utils.Error("User: Agents cannot manage users. Use a manager server", nil)
		utils.HTTPError(w, "User Creation Error: Agents cannot manage users. Use a manager server",
			http.StatusInternalServerError, "UC001")
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