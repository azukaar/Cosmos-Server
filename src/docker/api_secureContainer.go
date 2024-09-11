package docker

import (
	"net/http"
	"encoding/json"
	"os"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/gorilla/mux"
)

func SecureContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	vars := mux.Vars(req)
	containerName := utils.SanitizeSafe(vars["containerId"])
	status := utils.Sanitize(vars["status"])
	
	if utils.IsInsideContainer && containerName == os.Getenv("HOSTNAME") {
		utils.Error("SecureContainerRoute - Container cannot update itself", nil)
		utils.HTTPError(w, "Container cannot update itself", http.StatusBadRequest, "DS003")
		return
	}
	
	if(req.Method == "GET") {
		container, err := DockerClient.ContainerInspect(DockerContext, containerName)
		if err != nil {
			utils.Error("ContainerSecureInscpect", err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DS002")
			return
		}

		AddLabels(container, map[string]string{
			"cosmos-force-network-secured": status,
		});

		// change network mode to bridge in case it was set to container
		container.HostConfig.NetworkMode = "bridge"

		utils.Log("API: Set Force network secured "+status+" : " + containerName)

		_, errEdit := EditContainer(container.ID, container, false)
		if errEdit != nil {
			utils.Error("ContainerSecureEdit", errEdit)
			utils.HTTPError(w, "Internal server error: " + errEdit.Error(), http.StatusInternalServerError, "DS003")
			return
		}

		utils.TriggerEvent(
			"cosmos.docker.isolate",
			"Container network isolation changed",
			"success",
			"container@"+containerName,
			map[string]interface{}{
				"container": containerName,
				"status": status,
		})

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}