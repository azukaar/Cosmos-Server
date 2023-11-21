package docker

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
	filters "github.com/docker/docker/api/types/filters"
	volumeTypes"github.com/docker/docker/api/types/volume"
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
		volumes, err := DockerClient.VolumeList(context.Background(), filters.Args{})
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
		err := DockerClient.VolumeRemove(context.Background(), volumeName, true)
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
		volumeOptions := volumeTypes.VolumeCreateBody{
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