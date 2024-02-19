package storage

import (
	"net/http"
	"encoding/json"
	"fmt"
	"bufio"
	"os/exec"

	"github.com/azukaar/cosmos-server/src/utils"
)

type FormatDiskJSON struct {
	Disk string `json:"disk"`
	Format string `json:"format"`
}

func isDiskOrPartition(path string) (string, error) {
	cmd := exec.Command("lsblk", "-no", "TYPE", path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
			return "", err
	}

	if err := cmd.Start(); err != nil {
			return "", err
	}

	scanner := bufio.NewScanner(stdout)
	if scanner.Scan() {
			return scanner.Text(), nil
	}
	
	if err := scanner.Err(); err != nil {
			return "", err
	}

	return "", fmt.Errorf("no output from lsblk")
}

func FormatDiskRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		var request FormatDiskJSON
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			utils.Error("FormatDiskRoute: Invalid User Request", err)
			utils.HTTPError(w, "FormatDiskRoute Error", http.StatusInternalServerError, "FD001")
			return
		}

		// Enable streaming of response by setting appropriate headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Type", "text/plain")

		flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
        return
    }

		out, err := FormatDisk(request.Disk, request.Format)
		if err != nil {
			utils.Error("FormatDiskRoute: Error formatting disk", err)
			utils.HTTPError(w, "FormatDiskRoute Error", http.StatusInternalServerError, "FD002")
			return
		}

		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			utils.Log(scanner.Text())
			fmt.Fprintf(w, scanner.Text() + "\n")
			flusher.Flush()
		}

		// Check if the disk is a partition
		diskType, err := isDiskOrPartition(request.Disk)
		utils.Debug("Checking if " + request.Disk + " is a partition: " + diskType)

		if(diskType == "disk") {
			out, err = CreateSinglePartition(request.Disk)
			if err != nil {
				utils.Error("FormatDiskRoute: Error creating partition", err)
				utils.HTTPError(w, "FormatDiskRoute Error", http.StatusInternalServerError, "FD003")
				return
			}
	
			scanner = bufio.NewScanner(out)
			for scanner.Scan() {
				utils.Log(scanner.Text())
				fmt.Fprintf(w, scanner.Text() + "\n")
				flusher.Flush()
			}
	
			out, err = FormatDisk(request.Disk + "1", request.Format)
			if err != nil {
				utils.Error("FormatDiskRoute: Error formatting partition", err)
				utils.HTTPError(w, "FormatDiskRoute Error", http.StatusInternalServerError, "FD002")
				return
			}
	
			scanner = bufio.NewScanner(out)
			for scanner.Scan() {
				utils.Log(scanner.Text())
				fmt.Fprintf(w, scanner.Text() + "\n")
				flusher.Flush()
			}
		}

		utils.Log("FormatDiskRoute - formatted disk " + request.Disk + " with format " + request.Format)
		fmt.Fprintf(w, "[OPERATION SUCCEEDED]")
		flusher.Flush()
		return
	} else {
		utils.Error("FormatDiskRoute: Method not allowed " + req.Method, nil)
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