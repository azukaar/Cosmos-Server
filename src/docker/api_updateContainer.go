package docker

import (
	"encoding/json"
	"net/http"
	"os"
	"os/user"
	"strconv"

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
	// we make this a int so that we can ignore 0
	Interactive    int               `json:"interactive"`
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
			// Create missing folders for bind mounts
			for _, newmount := range form.Volumes {
				if newmount.Type == mount.TypeBind {
					if _, err := os.Stat(newmount.Source); os.IsNotExist(err) {
						err := os.MkdirAll(newmount.Source, 0755)
						if err != nil {
							utils.Error("UpdateService: Unable to create directory for bind mount", err)
							utils.HTTPError(w, "Unable to create directory for bind mount: "+err.Error(), http.StatusInternalServerError, "DS004")
							return
						}
			
						// Change the ownership of the directory to the container.User
						userInfo, err := user.Lookup(container.Config.User)
						if err != nil {
							utils.Error("UpdateService: Unable to lookup user", err)
						} else {
							uid, _ := strconv.Atoi(userInfo.Uid)
							gid, _ := strconv.Atoi(userInfo.Gid)
							err = os.Chown(newmount.Source, uid, gid)
							if err != nil {
								utils.Error("UpdateService: Unable to change ownership of directory", err)
							}
						}	
					}
				}
			}

			container.HostConfig.Mounts = form.Volumes
			container.HostConfig.Binds = []string{}
		}
		if(form.Interactive != 0) {
			container.Config.Tty = form.Interactive == 2
			container.Config.OpenStdin = form.Interactive == 2
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