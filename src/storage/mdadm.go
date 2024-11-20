package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"sync"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
)

// RaidConfig represents the configuration for creating a new RAID array
type RaidConfig struct {
	Name     string   `json:"name"`
	Level    int      `json:"level"`    // 0, 1, 5, 6, 10
	Devices  []string `json:"devices"`  // List of device paths
	Spares   []string `json:"spares"`   // List of spare device paths
	Metadata string   `json:"metadata"` // Metadata version (default: 1.2)
	Filesystem string `json:"filesystem"` // Filesystem to use (default: ext4)
}

var raidOpMutex sync.Mutex

// CreateRaidArray creates a new RAID array using mdadm
func CreateRaidArray(config RaidConfig) error {
	if len(config.Devices) < 2 {
		return fmt.Errorf("at least 2 devices are required for RAID")
	}

	// Validate RAID level
	validLevels := map[int]bool{0: true, 1: true, 5: true, 6: true, 10: true}
	if !validLevels[config.Level] {
		return fmt.Errorf("invalid RAID level: %d", config.Level)
	}

	// Default metadata version if not specified
	if config.Metadata == "" {
		config.Metadata = "1.2"
	}

	// Default filesystem if not specified
	if config.Filesystem == "" {
		config.Filesystem = "ext4"
	}

	// unmout all devices
	for _, device := range config.Devices {
		Unmount(device, true)
	}

	// Construct the mdadm command
	cmdArgs := []string{
			"--create",
			"--run",
			"--verbose",
			"/dev/md/" + config.Name,
			"--level=" + fmt.Sprint(config.Level),
			"--metadata=" + config.Metadata,
			"--raid-devices=" + fmt.Sprint(len(config.Devices)),
	}

	// Add each device as a separate argument
	cmdArgs = append(cmdArgs, config.Devices...)

	// If there are spares, append them as well
	if len(config.Spares) > 0 {
			cmdArgs = append(cmdArgs, "--spare-devices="+fmt.Sprint(len(config.Spares)))
			cmdArgs = append(cmdArgs, config.Spares...)
	}

	output, err := utils.Exec("mdadm", cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to create RAID array: %v, output: %s", err, string(output))
	}

	utils.Log("[MDADM] RAID array created successfully: " + output)

	// Update mdadm.conf
	output, err = utils.Exec("mdadm", "--detail", "--scan")
	if err != nil {
		return fmt.Errorf("failed to update mdadm configuration: %v", err)
	}

	time.Sleep(2 * time.Second)

	utils.Log("[MDADM] Updated mdadm configuration: " + output)

	// Format the RAID array
	ioreadr, err := FormatDisk("/dev/md/" + config.Name, config.Filesystem)
	if err != nil {
		return fmt.Errorf("failed to format RAID array: %v", err)
	}
	output = ""
	scanner := bufio.NewScanner(ioreadr)
	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}

	utils.Log("[MDADM] Formatted RAID array: " + output)

	return nil
}

// DeleteRaidArray removes a RAID array
func DeleteRaidArray(name string) error {
	cmd := exec.Command("mdadm", "--stop", "/dev/md/"+name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop RAID array: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("mdadm", "--remove", "/dev/md/"+name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove RAID array: %v, output: %s", err, string(output))
	}

	// Update mdadm.conf
	cmd = exec.Command("mdadm", "--detail", "--scan")
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update mdadm configuration: %v", err)
	}

	return nil
}

// AddDeviceToRaid adds a new device to an existing RAID array
func AddDeviceToRaid(name string, device string) error {
	cmd := exec.Command("mdadm", "--add", "/dev/md/"+name, device)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add device to RAID array: %v, output: %s", err, string(output))
	}
	return nil
}

// ReplaceDeviceInRaid replaces a failed device in a RAID array
func ReplaceDeviceInRaid(name string, oldDevice string, newDevice string) error {
	// First, mark the old device as failed
	cmd := exec.Command("mdadm", "/dev/md/"+name, "--fail", oldDevice)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to mark device as failed: %v, output: %s", err, string(output))
	}

	// Remove the failed device
	cmd = exec.Command("mdadm", "/dev/md/"+name, "--remove", oldDevice)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove failed device: %v, output: %s", err, string(output))
	}

	// Add the new device
	cmd = exec.Command("mdadm", "/dev/md/"+name, "--add", newDevice)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add new device: %v, output: %s", err, string(output))
	}

	return nil
}

// GetRaidInfo returns detailed information about all RAID arrays
func GetRaidInfo() ([]map[string]interface{}, error) {
	// First get the list of arrays
	cmd := exec.Command("mdadm", "--detail", "--scan")
	scanOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to scan RAID arrays: %v", err)
	}

	arrays := []map[string]interface{}{}
	scanner := bufio.NewScanner(strings.NewReader(string(scanOutput)))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ARRAY") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			devicePath := parts[1]
			
			// Get detailed information for each array
			cmd = exec.Command("mdadm", "--detail", devicePath)
			detailOutput, err := cmd.CombinedOutput()
			if err != nil {
				continue
			}

			details := parseRaidDetails(string(detailOutput))
			details["device"] = devicePath
			arrays = append(arrays, details)
		}
	}

	return arrays, nil
}

// parseRaidDetails parses the output of mdadm --detail
func parseRaidDetails(output string) map[string]interface{} {
	details := make(map[string]interface{})
	var devices []map[string]string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				details[key] = value
			}
		}

		// Parse device information
		if strings.Contains(line, "/dev/") {
			device := make(map[string]string)
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.HasPrefix(field, "/dev/") {
					device["path"] = field
				} else if field == "active" || field == "faulty" || field == "spare" {
					device["status"] = field
				}
			}
			if device["path"] != "" {
				devices = append(devices, device)
			}
		}
	}

	details["devices"] = devices
	return details
}

// HTTP Routes

func RaidCreateRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method != "POST" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	raidOpMutex.Lock()
	defer raidOpMutex.Unlock()

	var config RaidConfig
	if err := json.NewDecoder(req.Body).Decode(&config); err != nil {
		utils.HTTPError(w, "Invalid request body", http.StatusBadRequest, "RAID001")
		return
	}

	if err := CreateRaidArray(config); err != nil {
		utils.Error("RaidCreateRoute: Error creating RAID array", err)
		utils.HTTPError(w, "Error creating RAID array: "+err.Error(), http.StatusInternalServerError, "RAID002")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "OK",
		"message": "RAID array created successfully",
	})
}

func RaidDeleteRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method != "DELETE" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	raidOpMutex.Lock()
	defer raidOpMutex.Unlock()

	vars := mux.Vars(req)
	name := vars["name"]

	if err := DeleteRaidArray(name); err != nil {
		utils.Error("RaidDeleteRoute: Error deleting RAID array", err)
		utils.HTTPError(w, "Error deleting RAID array: "+err.Error(), http.StatusInternalServerError, "RAID003")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "OK",
		"message": "RAID array deleted successfully",
	})
}

type RaidDeviceRequest struct {
	Device string `json:"device"`
}

func RaidAddDeviceRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method != "POST" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	raidOpMutex.Lock()
	defer raidOpMutex.Unlock()

	vars := mux.Vars(req)
	name := vars["name"]

	var request RaidDeviceRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		utils.HTTPError(w, "Invalid request body", http.StatusBadRequest, "RAID004")
		return
	}

	if err := AddDeviceToRaid(name, request.Device); err != nil {
		utils.Error("RaidAddDeviceRoute: Error adding device to RAID array", err)
		utils.HTTPError(w, "Error adding device: "+err.Error(), http.StatusInternalServerError, "RAID005")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "OK",
		"message": "Device added successfully",
	})
}

type RaidReplaceDeviceRequest struct {
	OldDevice string `json:"oldDevice"`
	NewDevice string `json:"newDevice"`
}

func RaidReplaceDeviceRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method != "POST" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	raidOpMutex.Lock()
	defer raidOpMutex.Unlock()

	vars := mux.Vars(req)
	name := vars["name"]

	var request RaidReplaceDeviceRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		utils.HTTPError(w, "Invalid request body", http.StatusBadRequest, "RAID006")
		return
	}

	if err := ReplaceDeviceInRaid(name, request.OldDevice, request.NewDevice); err != nil {
		utils.Error("RaidReplaceDeviceRoute: Error replacing device in RAID array", err)
		utils.HTTPError(w, "Error replacing device: "+err.Error(), http.StatusInternalServerError, "RAID007")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "OK",
		"message": "Device replaced successfully",
	})
}

func RaidListRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method != "GET" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	arrays, err := GetRaidInfo()
	if err != nil {
		utils.Error("RaidListRoute: Error listing RAID arrays", err)
		utils.HTTPError(w, "Error listing RAID arrays: "+err.Error(), http.StatusInternalServerError, "RAID008")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   arrays,
	})
}
// GetRaidStatus returns the status of a specific RAID array
func GetRaidStatus(name string) (map[string]interface{}, error) {
	cmd := exec.Command("mdadm", "--detail", "/dev/md/"+name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get RAID status: %v", err)
	}
	
	details := parseRaidDetails(string(output))
	details["device"] = "/dev/md/" + name
	return details, nil
}

// ResizeRaidArray resizes a RAID array to use all available space
func ResizeRaidArray(name string) error {
	cmd := exec.Command("mdadm", "--grow", "--size=max", "/dev/md/"+name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to resize RAID array: %v, output: %s", err, string(output))
	}
	return nil
}

func RaidStatusRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method != "GET" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	vars := mux.Vars(req)
	name := vars["name"]

	status, err := GetRaidStatus(name)
	if err != nil {
		utils.Error("RaidStatusRoute: Error getting RAID status", err)
		utils.HTTPError(w, "Error getting RAID status: "+err.Error(), http.StatusInternalServerError, "RAID009")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   status,
	})
}

func RaidResizeRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method != "POST" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	raidOpMutex.Lock()
	defer raidOpMutex.Unlock()

	vars := mux.Vars(req)
	name := vars["name"]

	if err := ResizeRaidArray(name); err != nil {
		utils.Error("RaidResizeRoute: Error resizing RAID array", err)
		utils.HTTPError(w, "Error resizing RAID array: "+err.Error(), http.StatusInternalServerError, "RAID010")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "OK",
		"message": "RAID array resized successfully",
	})
}