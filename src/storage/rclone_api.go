package storage

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/accounting"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/fs/rc"

	"github.com/azukaar/cosmos-server/src/utils"
)

// RClone API handlers - replaces RCD HTTP proxy

var rcloneConfigInitialized bool
var rcloneConfigMutex sync.Mutex

func ensureRCloneConfig() {
	rcloneConfigMutex.Lock()
	defer rcloneConfigMutex.Unlock()

	if rcloneConfigInitialized {
		return
	}

	configLocation := utils.CONFIGFOLDER + "rclone.conf"
	config.SetConfigPath(configLocation)
	configfile.Install()
	rcloneConfigInitialized = true
}

// API_RClone_ConfigDump handles /cosmos/rclone/config/dump
func API_RClone_ConfigDump(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	ensureRCloneConfig()

	result := config.DumpRcBlob()
	json.NewEncoder(w).Encode(result)
}

// API_RClone_ConfigCreate handles /cosmos/rclone/config/create
func API_RClone_ConfigCreate(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	var payload struct {
		Name       string                 `json:"name"`
		Type       string                 `json:"type"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		utils.HTTPError(w, "Invalid JSON", http.StatusBadRequest, "RCL001")
		return
	}

	ensureRCloneConfig()

	// Convert parameters to rc.Params
	params := rc.Params{}
	for k, v := range payload.Parameters {
		params[k] = v
	}

	ctx := context.Background()
	_, err := config.CreateRemote(ctx, payload.Name, payload.Type, params, config.UpdateRemoteOpt{})
	if err != nil {
		utils.HTTPError(w, "Failed to create remote: "+err.Error(), http.StatusInternalServerError, "RCL002")
		return
	}

	config.SaveConfig()
	go InitRemoteStorage()

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK"})
}

// API_RClone_ConfigUpdate handles /cosmos/rclone/config/update
func API_RClone_ConfigUpdate(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	var payload struct {
		Name       string                 `json:"name"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		utils.HTTPError(w, "Invalid JSON", http.StatusBadRequest, "RCL003")
		return
	}

	ensureRCloneConfig()

	// Convert parameters to rc.Params
	params := rc.Params{}
	for k, v := range payload.Parameters {
		params[k] = v
	}

	ctx := context.Background()
	_, err := config.UpdateRemote(ctx, payload.Name, params, config.UpdateRemoteOpt{})
	if err != nil {
		utils.HTTPError(w, "Failed to update remote: "+err.Error(), http.StatusInternalServerError, "RCL004")
		return
	}

	config.SaveConfig()
	go InitRemoteStorage()

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK"})
}

// API_RClone_ConfigDelete handles /cosmos/rclone/config/delete
func API_RClone_ConfigDelete(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	var payload struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		utils.HTTPError(w, "Invalid JSON", http.StatusBadRequest, "RCL005")
		return
	}

	ensureRCloneConfig()

	config.DeleteRemote(payload.Name)
	config.SaveConfig()
	go InitRemoteStorage()

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK"})
}

// API_RClone_ConfigSave handles /cosmos/rclone/config/save
func API_RClone_ConfigSave(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	// Config is saved automatically by rclone, just return OK
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK"})
}

// API_RClone_OperationsAbout handles /cosmos/rclone/operations/about (ping storage)
func API_RClone_OperationsAbout(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	var payload struct {
		Fs string `json:"fs"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		utils.HTTPError(w, "Invalid JSON", http.StatusBadRequest, "RCL006")
		return
	}

	ctx := context.Background()
	fsPath := payload.Fs
	if !strings.HasSuffix(fsPath, ":") && !strings.Contains(fsPath, ":/") {
		fsPath = fsPath + ":"
	}

	f, err := fs.NewFs(ctx, fsPath)
	if err != nil {
		utils.HTTPError(w, "Failed to access remote: "+err.Error(), http.StatusInternalServerError, "RCL007")
		return
	}

	doAbout := f.Features().About
	if doAbout == nil {
		utils.HTTPError(w, "About not supported for this remote", http.StatusNotImplemented, "RCL008")
		return
	}

	usage, err := doAbout(ctx)
	if err != nil {
		utils.HTTPError(w, "Failed to get storage info: "+err.Error(), http.StatusInternalServerError, "RCL009")
		return
	}

	result := make(map[string]interface{})
	if usage.Total != nil {
		result["total"] = *usage.Total
	}
	if usage.Used != nil {
		result["used"] = *usage.Used
	}
	if usage.Free != nil {
		result["free"] = *usage.Free
	}
	if usage.Trashed != nil {
		result["trashed"] = *usage.Trashed
	}

	json.NewEncoder(w).Encode(result)
}

// API_RClone_VfsStats handles /cosmos/rclone/vfs/stats
func API_RClone_VfsStats(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	var payload struct {
		Fs string `json:"fs"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		utils.HTTPError(w, "Invalid JSON", http.StatusBadRequest, "RCL009")
		return
	}

	// Get VFS stats from our tracked mounts
	remoteName := strings.TrimSuffix(payload.Fs, ":")
	mountPath := "/mnt/cosmos-storage-" + remoteName
	if utils.IsInsideContainer {
		mountPath = "/mnt/host/mnt/cosmos-storage-" + remoteName
	}

	mountsMutex.RLock()
	mp, exists := liveMounts[mountPath]
	mountsMutex.RUnlock()

	if !exists || mp == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	}

	// Get VFS from mount point
	vfsLayer := mp.VFS
	if vfsLayer == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	}

	// Return VFS stats
	stats := vfsLayer.Stats()
	json.NewEncoder(w).Encode(stats)
}

// API_RClone_CoreStats handles /cosmos/rclone/core/stats
func API_RClone_CoreStats(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	stats := accounting.GlobalStats()
	result, err := stats.RemoteStats(false)
	if err != nil {
		utils.HTTPError(w, "Failed to get stats: "+err.Error(), http.StatusInternalServerError, "RCL010")
		return
	}

	json.NewEncoder(w).Encode(result)
}
