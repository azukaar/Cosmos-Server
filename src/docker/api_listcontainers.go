package docker

import (
	"net/http"
	"strconv"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 
)

var maxLimit = 1000

func ListContainersRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	} 

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	// from, _ := req.URL.Query().Get("from")

	if limit == 0 {
		limit = maxLimit
	}
	
	if(req.Method == "GET") {
		containers, err := ListContainers()

		if err != nil {
			utils.Error("ListContainersRoute: Error while getting containers", err)
			utils.HTTPError(w, "Containers Get Error", http.StatusInternalServerError, "DL001")
			return	
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": containers,
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}