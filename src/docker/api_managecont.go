package docker

import (
	"net/http"
	"encoding/json"
	"io"
	"os"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/gorilla/mux"
	// "github.com/docker/docker/client"
	contstuff "github.com/docker/docker/api/types/container"
	doctype "github.com/docker/docker/api/types"
)

func ManageContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	errD := Connect()
	if errD != nil {
		utils.Error("ManageContainer", errD)
		utils.HTTPError(w, "Internal server error: " + errD.Error(), http.StatusInternalServerError, "DS002")
		return
	}

	vars := mux.Vars(req)
	containerName := utils.Sanitize(vars["containerId"])
	// stop, start, restart, kill, remove, pause, unpause, recreate
	action := utils.Sanitize(vars["action"])

	utils.Log("ManageContainer " + containerName)

	if req.Method == "GET" {
		container, err := DockerClient.ContainerInspect(DockerContext, containerName)
		if err != nil {
			utils.Error("ManageContainer", err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DS002")
			return
		}

		imagename := container.Config.Image

		utils.Log("API: " + action + " " + containerName)

		// switch action
		switch action {
		case "stop":
			err = DockerClient.ContainerStop(DockerContext, container.ID, contstuff.StopOptions{})
		case "start":
			err = DockerClient.ContainerStart(DockerContext, container.ID, doctype.ContainerStartOptions{})
		case "restart":
			err = DockerClient.ContainerRestart(DockerContext, container.ID, contstuff.StopOptions{})
		case "kill":
			err = DockerClient.ContainerKill(DockerContext, container.ID, "")
		case "remove":
			err = DockerClient.ContainerRemove(DockerContext, container.ID, doctype.ContainerRemoveOptions{})
		case "pause":
			err = DockerClient.ContainerPause(DockerContext, container.ID)
		case "unpause":
			err = DockerClient.ContainerUnpause(DockerContext, container.ID)
		case "recreate":
			_, err = EditContainer(container.ID, container)
		case "update":
			pull, errPull := DockerClient.ImagePull(DockerContext, imagename, doctype.ImagePullOptions{})
			if errPull != nil {
				utils.Error("Docker Pull", errPull)
				utils.HTTPError(w, "Cannot pull new image", http.StatusBadRequest, "DS004")
				return
			}
			io.Copy(os.Stdout, pull)
			_, err = EditContainer(container.ID, container)
		default:
			utils.HTTPError(w, "Invalid action", http.StatusBadRequest, "DS003")
			return
		}

		if err != nil {
			utils.Error("ManageContainer: " + action, err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DS004")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("ManageContainer: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}