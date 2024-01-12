package docker

import (
	"net/http"
	"encoding/json"
	"os"
	"bufio"
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/gorilla/mux"
	contstuff "github.com/docker/docker/api/types/container"
	doctype "github.com/docker/docker/api/types"
)


func DeleteContainerLoose(containerName string) error {
	utils.Log("DeleteContainerLoose: " + containerName)
	err := DockerClient.ContainerRemove(DockerContext, containerName, doctype.ContainerRemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}

func DeleteContainerCompose(containerName, composeProjectDir, composeFile string) error {
	utils.Log("DeleteContainerCompose: " + containerName + " - " + composeProjectDir)

	project, err := LoadCompose(composeProjectDir)
	if err != nil {
		utils.Error("DeleteContainerCompose: Error while loading compose file", err)
		return err
	}

	delete(project.Services, containerName)

	// Marshal the struct to JSON or yml format.
	data, err := yaml.Marshal(project)
	if err != nil {
		utils.Error("DeleteContainerCompose: Marshal", err)
		return err
	}

	// Write the data to the file.
	err = SaveCompose(composeFile, data)
	if err != nil {
		utils.Error("DeleteContainerCompose: SaveCompose", err)
		return err
	}

	ComposeUp(composeProjectDir, func(message string, outputType int) {
		utils.Debug(message)
	})

	return nil
}

func DeleteContainer(containerName string) error {
	utils.Log("DeleteContainer: " + containerName)
	containerConfig, err := DockerClient.ContainerInspect(DockerContext, containerName)
	if err != nil {
		return err
	}

	composeProject := containerConfig.Config.Labels["com.docker.compose.project"]
	composeProjectDir := containerConfig.Config.Labels["com.docker.compose.project.working_dir"]
	composeFile := containerConfig.Config.Labels["com.docker.compose.project.config_files"]
	composeContainerName := containerConfig.Config.Labels["com.docker.compose.service"]

	if composeProject != "" {
		err = DeleteContainerCompose(composeContainerName, composeProjectDir, composeFile)
		if err != nil {
			return err
		} else {
			return DeleteContainerLoose(containerName)
		}
	} else {
		return DeleteContainerLoose(containerName)
	}

	return nil
}

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
	
	if os.Getenv("HOSTNAME") != "" && containerName == os.Getenv("HOSTNAME") && action != "update" && action != "recreate" {
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
			err = DockerClient.ContainerStart(DockerContext, container.ID, doctype.ContainerStartOptions{})
		case "restart":
			err = DockerClient.ContainerRestart(DockerContext, container.ID, contstuff.StopOptions{})
		case "kill":
			err = DockerClient.ContainerKill(DockerContext, container.ID, "")
		case "remove":
			err = DeleteContainer(containerName)
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