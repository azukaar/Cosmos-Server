package storage

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"

	"github.com/azukaar/cosmos-server/src/utils"
)

// UnmountRoute handles unmounting filesystem requests
func SnapRAIDRunRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		vars := mux.Vars(req)
		name := vars["name"]
		action := vars["action"]

		config := utils.GetMainConfig()
		snaps := config.Storage.SnapRAIDs
		for _, snap := range snaps {
			if snap.Name == name {
				if action == "sync" {
					RunSnapRAIDSync(snap)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "OK",
					})
					return
				} else if action == "scrub" {
					RunSnapRAIDScrub(snap)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "OK",
					})
					return
				} else if action == "fix" {
					RunSnapRAIDFix(snap)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "OK",
					})
					return
				}	else if action == "enable" {
					ToggleSnapRAID(snap.Name, true)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "OK",
					})
					return
				}	else if action == "disable" {
					ToggleSnapRAID(snap.Name, false)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "OK",
					})
					return
				}	else {
					utils.Error("SnapRAIDRun: Invalid action " + action, nil)
					utils.HTTPError(w, "Invalid action", http.StatusBadRequest, "SNP001")
					return
				}
			}
		}
		
		utils.Error("SnapRAIDRun: SnapRAID not found " + name, nil)
		utils.HTTPError(w, "SnapRAID not found", http.StatusNotFound, "SNP002")
		return
	} else {
		utils.Error("UnmountRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
