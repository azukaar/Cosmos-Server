package user

import (
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils" 
)


func Me(w http.ResponseWriter, req *http.Request) {
  if (req.Method == "GET") {
		UserGet(w, req)
	} else {
		utils.Error("UserRoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}