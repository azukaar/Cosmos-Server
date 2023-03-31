package docker

import (
	"net/http"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/gorilla/mux"
)

func SecureContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	vars := mux.Vars(req)
	containerName := utils.Sanitize(vars["containerId"])
	status := utils.Sanitize(vars["status"])
	
	if(req.Method == "GET") {
		container, err := DockerClient.ContainerInspect(DockerContext, containerName)
		if err != nil {
			utils.Error("ContainerSecure", err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DS002")
			return
		}

		AddLabels(container, map[string]string{
			"cosmos-force-network-secured": status,
		});

		utils.Log("API: Set Force network secured "+status+" : " + containerName)

		_, errEdit := EditContainer(container.ID, container)
		if errEdit != nil {
			utils.Error("ContainerSecure", errEdit)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DS003")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}