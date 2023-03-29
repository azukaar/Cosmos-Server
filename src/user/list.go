package user

import (
	"net/http"
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils" 
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"math"
)

var maxLimit = 1000

func UserList(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	} 

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	// from, _ := req.URL.Query().Get("from")

	if limit == 0 {
		limit = maxLimit
	}
	
	if(req.Method == "GET") {
		c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
		}

		utils.Debug("UserList: List user ")

		userList := []utils.User{}	

		l := int64(math.Max((float64)(maxLimit), (float64)(limit)))

		fOpt := options.FindOptions{
			Limit: &l,
		}

		// TODO: Implement pagination

		cursor, errDB := c.Find(
			nil,
			map[string]interface{}{
				// "_id": map[string]interface{}{
				// 	"$gt": from,
				// },
			},
			&fOpt,
		)

		if errDB != nil {
			utils.Error("UserList: Error while getting user", errDB)
			utils.HTTPError(w, "User Get Error", http.StatusInternalServerError, "UL001")
			return
		}

		for cursor.Next(nil) {
			user := utils.User{}
			errDec := cursor.Decode(&user)
			if errDec != nil {
				utils.Error("UserList: Error while decoding user", errDec)
				utils.HTTPError(w, "User Get Error", http.StatusInternalServerError, "UL001")
				return
			}
			user.Link = "/api/user/" + user.Nickname
			userList = append(userList, user)
		}

		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": userList,
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}