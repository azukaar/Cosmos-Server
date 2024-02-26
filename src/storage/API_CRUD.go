package storage

import (
	"encoding/json"
	"net/http"
	"fmt"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
)

func ListDisksRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
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

func ListMountsRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
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
	Path       string `json:"path"`
	MountPoint string `json:"mountPoint"`
	Permanent  bool   `json:"permanent"`
	Chown 		 string `json:"chown"`
}

// MountRoute handles mounting filesystem requests
func MountRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		var request MountRequest
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			utils.Error("MountRoute: Invalid request", err)
			utils.HTTPError(w, "Invalid request: " + err.Error(), http.StatusBadRequest, "MNT001")
			return
		}

		if err := Mount(request.Path, request.MountPoint, request.Permanent, request.Chown); err != nil {
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

// UnmountRoute handles unmounting filesystem requests
func UnmountRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
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
	Branches   []string `json:"branches"`
	MountPoint string `json:"mountPoint"`
	Permanent  bool   `json:"permanent"`
	Chown 		 string `json:"chown"`
	Opts 		   string `json:"opts"`
}

// MergeRoute handles merging filesystem requests
func MergeRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
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