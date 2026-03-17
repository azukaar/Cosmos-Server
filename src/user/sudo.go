package user

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/azukaar/cosmos-server/src/utils"
)

type SudoRequestJSON struct {
	Password string `validate:"required,min=8,max=128,containsany=~!@#$%^&*()_+=-{[}]:;"'<>.?/,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
}

// UserSudo godoc
// @Summary Elevate to sudo mode
// @Description Re-authenticates the user with their password to grant elevated (sudo) permissions
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SudoRequestJSON true "Password for sudo elevation"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/sudo [post]
func UserSudo(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		time.Sleep(time.Duration(rand.Float64()*2) * time.Second)

		if utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil {
			return
		}
		
		authCtx := utils.GetAuthContext(req)
		userNickname := authCtx.Nickname

		if !utils.PermissionsHaveSudo(authCtx.Permissions) {
			utils.Error("UserSudo: User cannot sudo", nil)
			utils.HTTPError(w, "User cannot sudo", http.StatusUnauthorized, "HTTP005")
			return
		}

		var request SudoRequestJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("UserSudo: Invalid User Request", err1)
			utils.HTTPError(w, "User Sudo Error", http.StatusInternalServerError, "UL001")
			return
		}

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
		defer closeDb()
		if errCo != nil {
			utils.Error("Database Connect", errCo)
			utils.HTTPError(w, "Database Error", http.StatusInternalServerError, "DB001")
			return
		}

		nickname := utils.Sanitize(userNickname)
		password := request.Password

		user := utils.User{}

		utils.Debug("UserSudo: Logging user " + nickname)

		err3 := c.FindOne(nil, map[string]interface{}{
			"Nickname": nickname,
		}).Decode(&user)

		if err3 == mongo.ErrNoDocuments {
			bcrypt.CompareHashAndPassword([]byte("$2a$14$4nzsVwEnR3.jEbMTME7kqeCo4gMgR/Tuk7ivNExvXjr73nKvLgHka"), []byte("dummyPassword"))
			utils.Error("UserSudo: User not found", err3)
			utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
			return
		} else if err3 != nil {
			bcrypt.CompareHashAndPassword([]byte("$2a$14$4nzsVwEnR3.jEbMTME7kqeCo4gMgR/Tuk7ivNExvXjr73nKvLgHka"), []byte("dummyPassword"))
			utils.Error("UserSudo: Error while finding user", err3)
			utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
			return
		} else if user.Password == "" {
			utils.Error("UserSudo: User not registered", nil)
			utils.HTTPError(w, "User not registered", http.StatusInternalServerError, "UL002")
			return
		} else {
			err2 := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err2 != nil {
				utils.Error("UserSudo: Encryption error", err2)
				utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
				return
			}

			SendUserToken(w, req, user, true, true)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
			})
		}
	} else {
		utils.Error("UserLogin: Method not allowed"+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
