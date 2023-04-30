package user

import (
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils" 
)

func UsersIdRoute(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "DELETE") {
		UserDelete(w, req)
	} else if (req.Method == "GET") {
		UserGet(w, req)
	} else if (req.Method == "PATCH") {
		UserEdit(w, req)
	} else {
		utils.Error("UserRoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func UsersRoute(w http.ResponseWriter, req *http.Request) {
  if (req.Method == "POST") {
		UserCreate(w, req)
	} else if (req.Method == "GET") {
		UserList(w, req)
	} else {
		utils.Error("UserRoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func API2FA(w http.ResponseWriter, req *http.Request) {
	if(req.Method == "POST") {
		Check2FA(w, req)
	} else if (req.Method == "GET") {
		New2FA(w, req)
	} else if (req.Method == "DELETE") {
		Delete2FA(w, req)
	} else {
		utils.Error("API2FARoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}