package docker

import (
	"encoding/json"
	"net/http"
	"os"
	"fmt"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/mount"
	"github.com/gorilla/mux"
)

type ContainerForm struct {
	Image          string            `json:"image"`
	User 				   string            `json:"user"`
	RestartPolicy  string            `json:"restartPolicy"`
	Env            []string          `json:"envVars"`
	Devices        []string `json:"devices"`
	Labels         map[string]string `json:"labels"`
	PortBindings   nat.PortMap       `json:"portBindings"`
	Volumes        []mount.Mount     `json:"Volumes"`
	// we make this a int so that we can ignore 0
	Interactive    int               `json:"interactive"`
	NetworkMode 	 string            `json:"networkMode"`
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
			// THIS IS HACK BECAUSE USER IS NULLABLE, BETTER SOLUTION TO COME
			container.Config.User = form.User
			container.HostConfig.RestartPolicy = containerType.RestartPolicy{Name: form.RestartPolicy}
		}
		if(form.Env != nil) {
			container.Config.Env = form.Env
		}
		
		if(form.Devices != nil) {
			container.HostConfig.Devices = []containerType.DeviceMapping{}
			for _, device := range form.Devices {
				deviceHost := strings.Split(device, ":")[0]
				deviceCont := deviceHost

				if strings.Contains(device, ":") {
					deviceCont = strings.Split(device, ":")[1]
				}

				container.HostConfig.Devices = append(container.HostConfig.Devices, containerType.DeviceMapping{
					PathOnHost:        deviceHost,
					PathInContainer:   deviceCont,
					CgroupPermissions: "rwm",
				})
			}
		}

		if(form.Labels != nil) {
			container.Config.Labels = form.Labels
		}

		if(form.PortBindings != nil) {
			utils.Debug(fmt.Sprintf("UpdateContainer: PortBindings: %v", form.PortBindings))
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
		if(form.Interactive != 0) {
			container.Config.Tty = form.Interactive == 2
			container.Config.OpenStdin = form.Interactive == 2
		}
		if(form.NetworkMode != "") {
			container.HostConfig.NetworkMode = containerType.NetworkMode(form.NetworkMode)
			// if not bridge, remove mac address
			if form.NetworkMode != "bridge" &&
				 form.NetworkMode != "default" {
					container.Config.MacAddress = ""
			}
		}

		_, err = EditContainer(container.ID, container, false)
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