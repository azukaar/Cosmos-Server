package docker

import (
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils" 
)

func ContainersIdRoute(w http.ResponseWriter, req *http.Request) {
	if (req.Method == "GET") {
		// ContainerGet(w, req)
	} else {
		utils.Error("UserRoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func ContainersRoute(w http.ResponseWriter, req *http.Request) {
  if (req.Method == "POST") {
		// CreateContainer(w, req)
	} else if (req.Method == "GET") {
		ListContainersRoute(w, req)
	} else {
		utils.Error("UserRoute: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}