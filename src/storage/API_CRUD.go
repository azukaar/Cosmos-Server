package storage

import (
	"encoding/json"
	"net/http"
	"fmt"
	"strings"

	"github.com/gorilla/mux"

	"github.com/azukaar/cosmos-server/src/utils"
)

// ListDisksRoute godoc
// @Summary List all disks
// @Tags Storage
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/disks [get]
func ListDisksRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
		return
	}

	if req.Method == "GET" {
		disks, err := ListDisks()
		if err != nil {
			utils.Error("ListDisksRoute: " + err.Error(), nil)
			utils.HTTPError(w, "Internal server error", http.StatusInternalServerError, "STO001")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   disks,
		})
	} else {
		utils.Error("ListDisksRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// ListMountsRoute godoc
// @Summary List all mounted filesystems
// @Tags Storage
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/mounts [get]
func ListMountsRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
		return
	}

	if req.Method == "GET" {
		disks, err := ListMounts()
		if err != nil {
			utils.Error("ListMountsRoute: " + err.Error(), nil)
			utils.HTTPError(w, "Internal server error", http.StatusInternalServerError, "STO001")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   disks,
		})
	} else {
		utils.Error("ListMountsRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// Assuming the structure for the mount/unmount request
type MountRequest struct {
	Path       string `json:"path" validate:"required"`
	MountPoint string `json:"mountPoint" validate:"required"`
	Permanent  bool   `json:"permanent"`
	NetDisk    bool   `json:"netDisk"`
	Chown 		 string `json:"chown"`
}

// MountRoute godoc
// @Summary Mount a filesystem
// @Tags Storage
// @Accept json
// @Produce json
// @Param body body MountRequest true "Mount request"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/mount [post]
func MountRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if req.Method == "POST" {
		var request MountRequest
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			utils.Error("MountRoute: Invalid request", err)
			utils.HTTPError(w, "Invalid request: " + err.Error(), http.StatusBadRequest, "MNT001")
			return
		}

		if err := Mount(request.Path, request.MountPoint, request.Permanent, request.NetDisk, request.Chown); err != nil {
			utils.Error("MountRoute: Error mounting", err)
			utils.HTTPError(w, "Error mounting filesystem:" + err.Error(), http.StatusInternalServerError, "MNT002")
			return
		}

		utils.Log(fmt.Sprintf("Mounted %s at %s", request.Path, request.MountPoint))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": fmt.Sprintf("Mounted %s at %s", request.Path, request.MountPoint),
		})
	} else {
		utils.Error("MountRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// UnmountRoute godoc
// @Summary Unmount a filesystem
// @Tags Storage
// @Accept json
// @Produce json
// @Param body body MountRequest true "Unmount request"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/unmount [post]
func UnmountRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if req.Method == "POST" {
		var request MountRequest
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			utils.Error("UnmountRoute: Invalid request", err)
			utils.HTTPError(w, "Invalid request: " + err.Error(), http.StatusBadRequest, "UMNT001")
			return
		}

		if err := Unmount(request.MountPoint, request.Permanent); err != nil {
			utils.Error("UnmountRoute: Error unmounting", err)
			utils.HTTPError(w, "Error unmounting filesystem:" + err.Error(), http.StatusInternalServerError, "UMNT002")
			return
		}

		utils.Log(fmt.Sprintf("Unmounted %s", request.MountPoint))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": fmt.Sprintf("Unmounted %s", request.MountPoint),
		})
	} else {
		utils.Error("UnmountRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// Assuming the structure for the mount/unmount request
type MergeRequest struct {
	Branches   []string `json:"branches" validate:"required"`
	MountPoint string `json:"mountPoint" validate:"required"`
	Permanent  bool   `json:"permanent"`
	Chown 		 string `json:"chown"`
	Opts 		   string `json:"opts"`
}

// MergeRoute godoc
// @Summary Merge multiple filesystems using MergerFS
// @Tags Storage
// @Accept json
// @Produce json
// @Param body body MergeRequest true "Merge request"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/merge [post]
func MergeRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if req.Method == "POST" {
		var request MergeRequest
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			utils.Error("MergeRoute: Invalid request", err)
			utils.HTTPError(w, "Invalid request: " + err.Error(), http.StatusBadRequest, "M001")
			return
		}

		if err := MountMergerFS(request.Branches, request.MountPoint, request.Opts, request.Permanent, request.Chown); err != nil {
			utils.Error("MergeRoute: Error merging", err)
			utils.HTTPError(w, "Error merging filesystem:" + err.Error(), http.StatusInternalServerError, "M002")
			return
		}

		utils.Log(fmt.Sprintf("Merged %s at %s", strings.Join(request.Branches, ":"), request.MountPoint))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": fmt.Sprintf("Merged %s at %s", strings.Join(request.Branches, ":"), request.MountPoint),
		})
	} else {
		utils.Error("MergeRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// CreateSNAPRaidRoute creates a SnapRAID configuration
func createSNAPRaidRoute(w http.ResponseWriter, req *http.Request) {
	var request utils.SnapRAIDConfig
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		utils.Error("CreateSNAPRaidRoute: Invalid request", err)
		utils.HTTPError(w, "Invalid request: " + err.Error(), http.StatusBadRequest, "M001")
		return
	}

	if err := CreateSnapRAID(request, ""); err != nil {
		utils.Error("CreateSNAPRaidRoute: Error merging", err)
		utils.HTTPError(w, "Error merging filesystem:" + err.Error(), http.StatusInternalServerError, "M002")
		return
	}

	utils.Log(fmt.Sprintf("Created SnapRAID %v with %s", request.Data, strings.Join(request.Parity, ":")))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"message": fmt.Sprintf("Created SnapRAID %v with %s", request.Data, strings.Join(request.Parity, ":")),
	})
}

// SnapRAIDEditRoute creates a SnapRAID configuration
func snapRAIDEditRoute(w http.ResponseWriter, req *http.Request) {	
	vars := mux.Vars(req)
	name := vars["name"]

	var request utils.SnapRAIDConfig
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		utils.Error("SnapRAIDEditRoute: Invalid request", err)
		utils.HTTPError(w, "Invalid request: " + err.Error(), http.StatusBadRequest, "M001")
		return
	}

	if err := CreateSnapRAID(request, name); err != nil {
		utils.Error("SnapRAIDEditRoute: Error merging", err)
		utils.HTTPError(w, "Error merging filesystem:" + err.Error(), http.StatusInternalServerError, "M002")
		return
	}

	utils.Log(fmt.Sprintf("Updated SnapRAID %v with %s", request.Data, strings.Join(request.Parity, ":")))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"message": fmt.Sprintf("Updated SnapRAID %v with %s", request.Data, strings.Join(request.Parity, ":")),
	})
}

func snapRAIDDeleteRoute(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]

	if err := DeleteSnapRAID(name); err != nil {
		utils.Error("SnapRAIDDeleteRoute: Error deleting", err)
		utils.HTTPError(w, "Error deleting SnapRAID:" + err.Error(), http.StatusInternalServerError, "M002")
		return
	}

	utils.Log(fmt.Sprintf("Deleted SnapRAID %s", name))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"message": fmt.Sprintf("Deleted SnapRAID %s", name),
	})
}

// SnapRAIDEditRoute godoc
// @Summary Update or delete a SnapRAID configuration by name
// @Tags Storage
// @Accept json
// @Produce json
// @Param name path string true "SnapRAID config name"
// @Param body body utils.SnapRAIDConfig false "Updated SnapRAID config (POST only)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/snapraid/{name} [post]
// @Router /api/snapraid/{name} [delete]
func SnapRAIDEditRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	if req.Method == "POST" {
		snapRAIDEditRoute(w, req)
		return
	} else if req.Method == "DELETE" {
		snapRAIDDeleteRoute(w, req)
		return
	} else {
		utils.Error("SnapRAIDEditRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

type SnapRAIDStatus struct {
	utils.SnapRAIDConfig
	Status string
}

func listSNAPRaidRoute(w http.ResponseWriter, req *http.Request) {	
	config := utils.GetMainConfig()
	snaps := config.Storage.SnapRAIDs
	result := []SnapRAIDStatus{}
	for _, snap := range snaps {
		status, err := RunSnapRAIDStatus(snap)
		if err != nil {
			utils.Error("listSNAPRaidRoute: Error getting status", err)
			status = "Error: " + err.Error()
		}
		final := SnapRAIDStatus{
			snap,
			status,
		}
		result = append(result, final)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data": result,
	})
}

// SNAPRaidCRUDRoute godoc
// @Summary List or create SnapRAID configurations
// @Tags Storage
// @Accept json
// @Produce json
// @Param body body utils.SnapRAIDConfig false "SnapRAID config (POST only)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/snapraid [get]
// @Router /api/snapraid [post]
func SNAPRaidCRUDRoute(w http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
			return
		}
		listSNAPRaidRoute(w, req)
		return
	} else if req.Method == "POST" {
		if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
			return
		}
		createSNAPRaidRoute(w, req)
		return
	} else {
		utils.Error("CreateSNAPRaidRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}