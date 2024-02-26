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
	Password string `json:"password"`
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
		// Enable streaming of response by setting appropriate headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Type", "text/plain")

		flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
        return
    }

		var request FormatDiskJSON
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			utils.Error("FormatDiskRoute: Invalid User Request", err)
			fmt.Fprintf(w, "[OPERATION FAILED] FormatDiskRoute  Syntax Error")
			http.Error(w, "FormatDiskRoute  Syntax Error", http.StatusBadRequest)

			return
		}
			
		nickname := req.Header.Get("x-cosmos-user")

		errp := utils.CheckPassword(nickname, request.Password)
		if errp != nil {
			utils.Error("FormatDiskRoute: Invalid User Request", errp)
			fmt.Fprintf(w, "[OPERATION FAILED] Wrong password supplied. Try again")
			http.Error(w, "Wrong password supplied. Try again", http.StatusUnauthorized)
			return
		}

		out, err := FormatDisk(request.Disk, request.Format)
		if err != nil {
			utils.Error("FormatDiskRoute: Error formatting disk", err)
			fmt.Fprintf(w, "[OPERATION FAILED] Error formatting disk: " + err.Error())
			http.Error(w, "Error formatting disk", http.StatusInternalServerError)
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
				fmt.Fprintf(w, "[OPERATION FAILED] Error creating partition: " + err.Error())
				http.Error(w, "Error creating partition", http.StatusInternalServerError)
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
				fmt.Fprintf(w, "[OPERATION FAILED] Error formatting partition: " + err.Error())
				http.Error(w, "Error formatting partition", http.StatusInternalServerError)
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

