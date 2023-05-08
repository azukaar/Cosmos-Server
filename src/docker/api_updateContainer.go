package docker

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/azukaar/cosmos-server/src/utils"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/mount"
	"github.com/gorilla/mux"
)

type ContainerForm struct {
	Image          string            `json:"image"`
	RestartPolicy  string            `json:"restartPolicy"`
	Env            []string          `json:"envVars"`
	Labels         map[string]string `json:"labels"`
	PortBindings   nat.PortMap       `json:"portBindings"`
	Volumes        []mount.Mount     `json:"Volumes"`
}

func UpdateContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	utils.Log("UpdateContainer" + "Updating container")

	if req.Method == "POST" {
		errD := Connect()
		if errD != nil {
			utils.Error("UpdateContainer", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "DS002")
			return
		}

		vars := mux.Vars(req)
		containerName := utils.SanitizeSafe(vars["containerId"])
		
		if os.Getenv("HOSTNAME") != "" && containerName == os.Getenv("HOSTNAME") {
			utils.Error("SecureContainerRoute - Container cannot update itself", nil)
			utils.HTTPError(w, "Container cannot update itself", http.StatusBadRequest, "DS003")
			return
		}

		container, err := DockerClient.ContainerInspect(DockerContext, containerName)
		if err != nil {
			utils.Error("UpdateContainer", err)
			utils.HTTPError(w, "Internal server error: "+err.Error(), http.StatusInternalServerError, "DS002")
			return
		}

		var form ContainerForm
		err = json.NewDecoder(req.Body).Decode(&form)
		if err != nil {
			utils.Error("UpdateContainer", err)
			utils.HTTPError(w, "Invalid JSON", http.StatusBadRequest, "DS003")
			return
		}

		// Update container settings
		if(form.Image != "") {
			container.Config.Image = form.Image
		}
		if(form.RestartPolicy != "") {
			container.HostConfig.RestartPolicy = containerType.RestartPolicy{Name: form.RestartPolicy}
		}
		if(form.Env != nil) {
			container.Config.Env = form.Env
		}
		if(form.Labels != nil) {
			container.Config.Labels = form.Labels
		}
		if(form.PortBindings != nil) {
			container.HostConfig.PortBindings = form.PortBindings
			container.Config.ExposedPorts = make(map[nat.Port]struct{})
			for port := range form.PortBindings {
				container.Config.ExposedPorts[port] = struct{}{}
			}
		}
		if(form.Volumes != nil) {
			container.HostConfig.Mounts = form.Volumes
			container.HostConfig.Binds = []string{}
		}

		_, err = EditContainer(container.ID, container)
		if err != nil {
			utils.Error("UpdateContainer: EditContainer", err)
			utils.HTTPError(w, "Internal server error: "+err.Error(), http.StatusInternalServerError, "DS004")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("UpdateContainer: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}