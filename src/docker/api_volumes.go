package docker

import (
	"context"
	"encoding/json"
	"net/http"
	yaml "gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
	volumeTypes "github.com/docker/docker/api/types/volume"
)

func ListVolumeRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		errD := Connect()
		if errD != nil {
			utils.Error("ManageContainer", errD)
			utils.HTTPError(w, "Internal server error: " + errD.Error(), http.StatusInternalServerError, "LV001")
			return
		}

		// List Docker volumes
		volumes, err := DockerClient.VolumeList(context.Background(), volumeTypes.ListOptions{})
		if err != nil {
			utils.Error("ListVolumeRoute: Error while getting volumes", err)
			utils.HTTPError(w, "Volumes Get Error", http.StatusInternalServerError, "LV002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   volumes,
		})
	} else {
		utils.Error("ListVolumeRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func DeleteVolumeLoose(volumeName string) error {
	utils.Log("Deleting volume: " + volumeName)
	err := DockerClient.VolumeRemove(DockerContext, volumeName, true)
	if err != nil {
		return err
	}
	return nil
}

func DeleteVolumeCompose(volumeName, composeProjectDir, composeFile string) error {
	utils.Log("DeleteVolumeCompose: " + volumeName + " - " + composeProjectDir)

	project, err := LoadCompose(composeProjectDir)
	if err != nil {
		utils.Error("DeleteVolumeCompose: Error while loading compose file", err)
		return err
	}

	deleted := false
	for key, volume := range project.Volumes {
		if volume.Name == volumeName {
			delete(project.Volumes, key)
			deleted = true
		}
	}
	if !deleted {
		delete(project.Volumes, volumeName)
	}

	// Marshal the struct to JSON or yml format.
	data, err := yaml.Marshal(project)
	if err != nil {
		utils.Error("DeleteVolumeCompose: Marshal", err)
		return err
	}

	// Write the data to the file.
	err = SaveCompose(composeFile, data)
	if err != nil {
		utils.Error("DeleteVolumeCompose: SaveCompose", err)
		return err
	}

	ComposeUp(composeProjectDir, func(message string, outputType int) {
		utils.Debug(message)
	})

	return nil
}

func DeleteVolume(volumeName string) error {
	utils.Log("DeleteVolume: " + volumeName)
	volume, err := DockerClient.VolumeInspect(DockerContext, volumeName)
	if err != nil {
		return err
	}

	composeProject := volume.Labels["com.docker.compose.project"]
	
	if composeProject != "" {
		containers, err := ListContainers()
		if err != nil {
			return err
		}
		
		// inspect containers until we find the compose file location
		composeProjectDir := ""
		composeFile := ""

		for _, container := range containers {
			utils.Debug("Checking container: " + container.Names[0])
			containerConfig, err := DockerClient.ContainerInspect(DockerContext, container.Names[0])
			if err != nil {
				return err
			}

			utils.Debug(containerConfig.Config.Labels["com.docker.compose.project"])
			
			if containerConfig.Config.Labels["com.docker.compose.project"] == composeProject {
				composeProjectDir = containerConfig.Config.Labels["com.docker.compose.project.working_dir"]
				composeFile = containerConfig.Config.Labels["com.docker.compose.project.config_files"]
				break
			}
		}

		if composeProjectDir == "" {
			return DeleteVolumeLoose(volumeName)
			// return errors.New("DeleteVolume: Couldn't find docker-compose project directory for volume " + volumeName + ". Aborting to prevent desync.")
		}

		err = DeleteVolumeCompose(volumeName, composeProjectDir, composeFile)
		if err != nil {
			return err
		} else {
			return DeleteVolumeLoose(volumeName)
		}
	} else {
		return DeleteVolumeLoose(volumeName)
	}

	return nil
}


func DeleteVolumeRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "DELETE" {
		// Get the volume name from URL
		vars := mux.Vars(req)
		volumeName := vars["volumeName"]

		errD := Connect()
		if errD != nil {
			utils.Error("DeleteVolumeRoute", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "DV001")
			return
		}

		// Delete the specified Docker volume
		err := DeleteVolume(volumeName)
		if err != nil {
			utils.Error("DeleteVolumeRoute: Error while deleting volume", err)
			utils.HTTPError(w, "Volume Deletion Error " + err.Error(), http.StatusInternalServerError, "DV002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": "Volume deleted successfully",
		})
	} else {
		utils.Error("DeleteVolumeRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

type VolumeCreateRequest struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

func CreateVolumeRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		errD := Connect()
		if errD != nil {
			utils.Error("CreateVolumeRoute", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "CV001")
			return
		}

		var payload VolumeCreateRequest
		err := json.NewDecoder(req.Body).Decode(&payload)
		if err != nil {
			utils.Error("CreateNetworkRoute: Error reading request body", err)
			utils.HTTPError(w, "Error reading request body: "+err.Error(), http.StatusBadRequest, "CN002")
			return
		}

		// Create Docker volume with the provided options
		volumeOptions := volumeTypes.CreateOptions {
			Name:   payload.Name,
			Driver: payload.Driver,
		}

		volume, err := DockerClient.VolumeCreate(context.Background(), volumeOptions)
		if err != nil {
			utils.Error("CreateVolumeRoute: Error while creating volume", err)
			utils.HTTPError(w, "Volume creation error: "+err.Error(), http.StatusInternalServerError, "CV004")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   volume,
		})
	} else {
		utils.Error("CreateVolumeRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func VolumesRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		ListVolumeRoute(w, req)
	} else if req.Method == "POST" {
		CreateVolumeRoute(w, req)
	} else {
		utils.Error("VolumesRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}