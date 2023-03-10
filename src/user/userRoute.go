package user

import (
	"net/http"
	"../utils" 
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