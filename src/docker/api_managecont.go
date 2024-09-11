package docker

import (
	"net/http"
	"encoding/json"
	"os"
	"bufio"
	"fmt"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/gorilla/mux"
	contstuff "github.com/docker/docker/api/types/container"
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
	containerName := utils.SanitizeSafe(vars["containerId"])
	// stop, start, restart, kill, remove, pause, unpause, recreate
	action := utils.Sanitize(vars["action"])
	
	if utils.IsInsideContainer && containerName == os.Getenv("HOSTNAME") && action != "update" && action != "recreate" {
		utils.Error("ManageContainer - Container cannot update itself", nil)
		utils.HTTPError(w, "Container cannot update itself", http.StatusBadRequest, "DS003")
		return
	}

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
			err = DockerClient.ContainerStart(DockerContext, container.ID, contstuff.StartOptions{})
		case "restart":
			err = DockerClient.ContainerRestart(DockerContext, container.ID, contstuff.StopOptions{})
		case "kill":
			err = DockerClient.ContainerKill(DockerContext, container.ID, "")
		case "remove":
			err = DockerClient.ContainerRemove(DockerContext, container.ID, contstuff.RemoveOptions{})
		case "pause":
			err = DockerClient.ContainerPause(DockerContext, container.ID)
		case "unpause":
			err = DockerClient.ContainerUnpause(DockerContext, container.ID)
		case "recreate":
			_, err = RecreateContainer(container.Name, container)
		case "update":
			out, errPull := DockerPullImage(imagename)
			if errPull != nil {
				utils.Error("Docker Pull", errPull)
				utils.HTTPError(w, "Cannot pull new image", http.StatusBadRequest, "DS004")
				return
			}
			defer out.Close()
			
			// Enable streaming of response by setting appropriate headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Transfer-Encoding", "chunked")

			flusher, ok := w.(http.Flusher)

			if !ok {
				utils.Error("Container Update - Cannot stream response", nil)
				utils.HTTPError(w, "Cannot stream response", http.StatusInternalServerError, "DS004")
				return
			}

			// wait for image pull to finish
			scanner := bufio.NewScanner(out)
			for scanner.Scan() {
				utils.Log(scanner.Text())
				fmt.Fprintf(w, scanner.Text() + "\n")
				flusher.Flush()
			}

			utils.Log("Container Update - Image pulled " + imagename)

			_, err = RecreateContainer(container.Name, container)

			if err != nil {
				utils.Error("Container Update - EditContainer", err)
				utils.HTTPError(w, "[OPERATION FAILED] Cannot recreate container", http.StatusBadRequest, "DS004")
				return
			}

			utils.UpdateAvailable["/" + containerName] = false
			fmt.Fprintf(w, "[OPERATION SUCCEEDED]")
			flusher.Flush()
			return
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