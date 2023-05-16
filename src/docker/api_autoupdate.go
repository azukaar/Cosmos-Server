package docker

import (
	"net/http"
	"encoding/json"
	"os"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/gorilla/mux"
)

func AutoUpdateContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	vars := mux.Vars(req)
	containerName := utils.SanitizeSafe(vars["containerId"])
	status := utils.Sanitize(vars["status"])
	
	if os.Getenv("HOSTNAME") != "" && containerName == os.Getenv("HOSTNAME") {
		utils.Error("AutoUpdateContainerRoute - Container cannot update itself", nil)
		utils.HTTPError(w, "Container cannot update itself", http.StatusBadRequest, "DS003")
		return
	}
	
	if(req.Method == "GET") {
		container, err := DockerClient.ContainerInspect(DockerContext, containerName)
		if err != nil {
			utils.Error("AutoUpdateContainer Inscpect", err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DS002")
			return
		}

		AddLabels(container, map[string]string{
			"cosmos-auto-update": status,
		});

		utils.Log("API: Set Auto Update "+status+" : " + containerName)

		_, errEdit := EditContainer(container.ID, container)
		if errEdit != nil {
			utils.Error("AutoUpdateContainer Edit", errEdit)
			utils.HTTPError(w, "Internal server error: " + errEdit.Error(), http.StatusInternalServerError, "DS003")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("AutoUpdateContainer: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}