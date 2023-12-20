package docker

import (
	"encoding/json"
	"net/http"
	"os"
	"fmt"
	"strings"
	// "io/ioutil"

	yaml "gopkg.in/yaml.v2"
	"github.com/azukaar/cosmos-server/src/utils"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/mount"
	"github.com/gorilla/mux"
	composeTypes "github.com/compose-spec/compose-go/v2/types"
	// composeLoader "github.com/compose-spec/compose-go/v2/loader"
	// cli "github.com/compose-spec/compose-go/v2/cli"
	// "github.com/docker/compose/v2/pkg/compose"
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

func updateLoseContainer(containerName string, container types.ContainerJSON, form ContainerForm, w http.ResponseWriter, req *http.Request) {
	// Update container settings
	if(form.Image != "") {
		container.Config.Image = form.Image
	}
	if(form.RestartPolicy != "") {
		// THIS IS HACK BECAUSE USER IS NULLABLE, BETTER SOLUTION TO COME
		container.Config.User = form.User		
		container.HostConfig.NetworkMode = containerType.NetworkMode(form.NetworkMode)
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
	_, err := EditContainer(container.ID, container, false)
	if err != nil {
		utils.Error("UpdateContainer: EditContainer", err)
		utils.HTTPError(w, "Internal server error: "+err.Error(), http.StatusInternalServerError, "DS004")
		return
	}

	return 
}

func updateDCContainer(containerName string, container types.ContainerJSON, form ContainerForm, w http.ResponseWriter, req *http.Request) {
	dcWorkingDir := container.Config.Labels["com.docker.compose.project.working_dir"]
	dcServiceName := container.Config.Labels["com.docker.compose.service"]

	project, err := LoadComposeFromName(containerName)
	if err != nil {
		utils.Error("UpdateContainer: LoadComposeFromName", err)
		utils.HTTPError(w, "Internal server error: "+err.Error(), http.StatusInternalServerError, "DS004")
		return
	}

	if project.Services == nil {
		utils.Error("UpdateContainer: project.Services is nil", nil)
		utils.HTTPError(w, "Internal server error: "+err.Error(), http.StatusInternalServerError, "DS004")
		return
	}

	service := project.Services[dcServiceName]

	// CHECK IF NULL ???

	if(form.Image != "") {
		service.Image = form.Image
	}
	if(form.RestartPolicy != "") {
		// THIS IS HACK BECAUSE USER IS NULLABLE, BETTER SOLUTION TO COME
		service.User = form.User
		service.NetworkMode = form.NetworkMode
		service.Restart = form.RestartPolicy
	}
	if(form.Env != nil) {
		service.Environment = ApplyMapDiff(ArrToMap(container.Config.Env), ArrToMap(form.Env), service.Environment)
	}
	if(form.Devices != nil) {
		service.Devices = form.Devices
	}
	if(form.Labels != nil) {
		service.Labels = MappingToLabel(ApplyMapDiff(container.Config.Labels, form.Labels, LabelsToMapping(service.Labels)))
	}
	if(form.Interactive != 0) {
		service.Tty = form.Interactive == 2
		service.StdinOpen = form.Interactive == 2
	}
	if(form.PortBindings != nil) {
		service.Ports = []composeTypes.ServicePortConfig{}
		for port, key := range form.PortBindings {
			service.Ports = append(service.Ports, composeTypes.ServicePortConfig{
				Target: (uint32)(port.Int()),
				Published: key[0].HostPort,
				Protocol: strings.ToUpper(string(port.Proto())),
				// PublishMode: "host",
			})
		}
	}
	utils.Debug(fmt.Sprintf("UpdateContainer: Volume: %v", project.Volumes))	

	if(form.Volumes != nil) {
		service.Volumes = []composeTypes.ServiceVolumeConfig{}

		for _, volume := range form.Volumes {
			source := volume.Source

			if volume.Type == "volume" {
				// check ServiceVolumeConfig for actual name
				volFound := false
				for name, serviceVolume := range project.Volumes {
					if serviceVolume.Name == source {
						source = name
						volFound = true
						break
					}
				}
				if !volFound {
					utils.Log("UpdateContainer: Volume not found: " + source + " - adding as external")
					project.Volumes[source] = composeTypes.VolumeConfig{
						Name: source,
						External: true,
					}
				}
			}
						
			service.Volumes = append(service.Volumes, composeTypes.ServiceVolumeConfig{
				Type: (string)(volume.Type),
				Source: source,
				Target: volume.Target,
			})
		}
	}

	project.Services[dcServiceName] = service

	// Marshal the struct to JSON or yml format.
	data, err := yaml.Marshal(project)
	if err != nil {
		utils.Error("UpdateContainer: Marshal", err)
		utils.HTTPError(w, "Internal server error: "+err.Error(), http.StatusInternalServerError, "DS004")
		return
	}

	// Write the data to the file.
	err = SaveComposeFromName(containerName, data)
	if err != nil {
		utils.Error("UpdateContainer: SaveComposeFromName", err)
		utils.HTTPError(w, "Internal server error: "+err.Error(), http.StatusInternalServerError, "DS004")
		return
	}

	ComposeUp(dcWorkingDir, func(message string, outputType int) {
		utils.Debug(message)
	})
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

		// if label container docker.compose.project 
		if container.Config.Labels["com.docker.compose.project"] != "" {
			utils.Log("UpdateContainer: Updating docker compose container")
			updateDCContainer(containerName, container, form, w, req)
		} else {
			utils.Log("UpdateContainer: Updating lose container")
			updateLoseContainer(containerName, container, form, w, req)
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


func ApplyMapDiff(m1, m2  map[string]string, m3 composeTypes.MappingWithEquals) composeTypes.MappingWithEquals {
	// Process additions and edits
	for key, val2 := range m2 {
		val1, exists := m1[key]
		if !exists || val1 != val2 {
			value := val2
			m3[key] = &value
		}
	}

	// Process deletions
	for key := range m3 {
		if _, exists := m2[key]; !exists {
			// check if exist in m3
			delete(m3, key)
		}
	}

	return m3
}