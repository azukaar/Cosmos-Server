package configapi

import (
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils" 
)

func ConfigRoute(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "GET") {
		ConfigApiGet(w, req)
	} else if (req.Method == "PUT") {
		ConfigApiSet(w, req)
	} else {
		utils.Error("UserRoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}