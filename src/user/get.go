package user

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils" 
)

// UserGet godoc
// @Summary Get a user by nickname
// @Description Returns user details for the specified nickname
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param nickname path string true "User nickname"
// @Success 200 {object} utils.APIResponse{data=utils.User}
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/users/{nickname} [get]
func UserGet(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	nickname := utils.Sanitize(vars["nickname"])

	if nickname == "" && utils.GetAuthContext(req).Nickname != "" {
		nickname = utils.GetAuthContext(req).Nickname
	}
	
	if utils.CheckPermissionsOrSelf(w, req, nickname, utils.PERM_USERS_READ) != nil {
		return
	}

	if(req.Method == "GET") {
		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
  	defer closeDb()
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