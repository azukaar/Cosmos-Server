package backups

import (
	"fmt"
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"os"
	
	"github.com/azukaar/cosmos-server/src/utils"
)

func AddBackupRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		var request utils.SingleBackupConfig
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			utils.Error("AddBackup: Invalid request", err)
			utils.HTTPError(w, "Invalid request: "+err.Error(), http.StatusBadRequest, "BCK001")
			return
		}

		config := utils.GetMainConfig()
		if _, exists := config.Backup.Backups[request.Name]; exists {
			utils.HTTPError(w, "Backup already exists", http.StatusBadRequest, "BCK002")
			return
		}

		// Check repository status
		repoInfo, err := os.Stat(request.Repository)
		if err != nil && !os.IsNotExist(err) {
			utils.Error("AddBackup: Failed to check repository", err)
			utils.HTTPError(w, "Failed to check repository: "+err.Error(), http.StatusInternalServerError, "BCK003")
			return
		}

		var password string
		isNewRepo := false

		if os.IsNotExist(err) {
			isNewRepo = true
		} else if repoInfo.IsDir() {
			files, err := os.ReadDir(request.Repository)
			if err != nil {
				utils.Error("AddBackup: Failed to read repository directory", err)
				utils.HTTPError(w, "Failed to read repository directory: "+err.Error(), http.StatusInternalServerError, "BCK005")
				return
			}
			isNewRepo = len(files) == 0
		}

		if isNewRepo {
			password = utils.GenerateRandomString(16)
			if err := CreateRepository(request.Repository, password); err != nil {
				utils.Error("AddBackup: Failed to create repository", err)
				utils.HTTPError(w, "Failed to create repository: "+err.Error(), http.StatusInternalServerError, "BCK004")
				return
			}
		} else {
			found := false
			for _, backup := range config.Backup.Backups {
				if backup.Repository == request.Repository {
					password = backup.Password
					found = true
					break
				}
			}

			if !found {
				utils.Error("AddBackup: Cannot find password for existing repository", nil)
				utils.HTTPError(w, "Repository exists but password not found", http.StatusBadRequest, "BCK006")
				return
			}

			if err := CheckRepository(request.Repository, password); err != nil {
				utils.Error("AddBackup: Invalid repository", err)
				utils.HTTPError(w, "Invalid repository: "+err.Error(), http.StatusInternalServerError, "BCK007")
				return
			}
		}

		request.Password = password
		if config.Backup.Backups == nil {
			config.Backup.Backups = make(map[string]utils.SingleBackupConfig)
		}
		config.Backup.Backups[request.Name] = request
		utils.SetBaseMainConfig(config)
		InitBackups()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "OK",
			"message": fmt.Sprintf("Added backup %s", request.Name),
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func EditBackupRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		var request utils.SingleBackupConfig
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			utils.Error("EditBackup: Invalid request", err)
			utils.HTTPError(w, "Invalid request: "+err.Error(), http.StatusBadRequest, "BCK001")
			return
		}

		config := utils.GetMainConfig()
		if _, exists := config.Backup.Backups[request.Name]; !exists {
			utils.HTTPError(w, "Backup does not exist", http.StatusBadRequest, "BCK002")
			return
		}

		current := config.Backup.Backups[request.Name]

		current.Crontab = request.Crontab
		current.CrontabForget = request.CrontabForget
		current.RetentionPolicy = request.RetentionPolicy

		config.Backup.Backups[request.Name] = current
		utils.SetBaseMainConfig(config)
		InitBackups()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "OK",
			"message": fmt.Sprintf("Edited backup %s", request.Name),
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func RemoveBackupRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "DELETE" {
		vars := mux.Vars(req)
		name := vars["name"]
		var request struct {
			DeleteRepo bool `json:"deleteRepo"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			utils.HTTPError(w, "Invalid request: " + err.Error(), http.StatusBadRequest, "BCK001")
			return
		}

		config := utils.GetMainConfig()
		backup, exists := config.Backup.Backups[name]
		if !exists {
			utils.HTTPError(w, "Backup not found", http.StatusNotFound, "BCK004")
			return
		}

		// Check if other backups use same repository
		otherBackupsUsingRepo := false
		for n, b := range config.Backup.Backups {
			if n != name && b.Repository == backup.Repository {
				otherBackupsUsingRepo = true
				break
			}
		}

		if request.DeleteRepo {
			if otherBackupsUsingRepo {
				if err := DeleteByTag(backup.Repository, backup.Password, backup.Name); err != nil {
					utils.Error("RemoveBackup: Failed to delete snapshots", err)
					utils.HTTPError(w, "Failed to delete snapshots: " + err.Error(), http.StatusInternalServerError, "BCK005")
					return
				}
			} else {
				if err := DeleteRepository(backup.Repository); err != nil {
					utils.Error("RemoveBackup: Failed to delete repository", err)
					utils.HTTPError(w, "Failed to delete repository: " + err.Error(), http.StatusInternalServerError, "BCK006")
					return
				}
			}
		} else {
			if err := DeleteByTag(backup.Repository, backup.Password, backup.Name); err != nil {
				utils.Error("RemoveBackup: Failed to delete snapshots", err)
				utils.HTTPError(w, "Failed to delete snapshots: " + err.Error(), http.StatusInternalServerError, "BCK005")
				return
			}
		}

		delete(config.Backup.Backups, name)
		utils.SetBaseMainConfig(config)
		InitBackups()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": fmt.Sprintf("Removed backup %s", name),
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func ListSnapshotsRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		vars := mux.Vars(req)
		name := vars["name"]
		
		backup, exists := utils.GetMainConfig().Backup.Backups[name]
		if !exists {
			utils.HTTPError(w, "Backup not found", http.StatusNotFound, "BCK004")
			return
		}

		output, err := ListSnapshotsWithFilters(backup.Repository, backup.Password, []string{backup.Name}, "", "")
		if err != nil {
			utils.Error("ListSnapshots: Failed to list snapshots", err)
			utils.HTTPError(w, "Failed to list snapshots: "+err.Error(), http.StatusInternalServerError, "BCK006")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var outputJSON []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &outputJSON); err != nil {
			utils.Error("ListSnapshots: Failed to parse snapshots", err)
			utils.HTTPError(w, "Failed to parse snapshots: "+err.Error(), http.StatusInternalServerError, "BCK007")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":	 outputJSON,
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func ListFoldersRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		vars := mux.Vars(req)
		name := vars["name"]
		snapshot := vars["snapshot"]
		path := req.URL.Query().Get("path")
		
		backup, exists := utils.GetMainConfig().Backup.Backups[name]
		if !exists {
			utils.HTTPError(w, "Backup not found", http.StatusNotFound, "BCK004")
			return
		}

		output, err := ListDirectory(backup.Repository, backup.Password, snapshot, path)
		if err != nil {
			utils.Error("ListFolders: Failed to list folders", err)
			utils.HTTPError(w, "Failed to list folders: "+err.Error(), http.StatusInternalServerError, "BCK007")
			return
		}

		outputF := SplitJSONObjects(output)
		
		w.Header().Set("Content-Type", "application/json")
		var outputJSON []map[string]interface{}
		if err := json.Unmarshal([]byte(outputF), &outputJSON); err != nil {
			utils.Error("ListSnapshots: Failed to parse snapshots", err)
			utils.HTTPError(w, "Failed to parse snapshots: "+err.Error(), http.StatusInternalServerError, "BCK007")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":	 outputJSON,
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func RestoreBackupRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		vars := mux.Vars(req)
		name := vars["name"]

		var request struct {
			SnapshotID string   `json:"snapshotId"`
			Target     string   `json:"target"`
			Include    []string `json:"include,omitempty"`
		}

		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			utils.Error("RestoreBackup: Invalid request", err)
			utils.HTTPError(w, "Invalid request: "+err.Error(), http.StatusBadRequest, "BCK008")
			return
		}

		backup, exists := utils.GetMainConfig().Backup.Backups[name]
		if !exists {
			utils.HTTPError(w, "Backup not found", http.StatusNotFound, "BCK004")
			return
		}

		// Verify snapshot belongs to this backup
		snapshots, err := ListSnapshotsWithFilters(backup.Repository, backup.Password, []string{backup.Name}, "", "")
		if err != nil {
			utils.Error("RestoreBackup: Failed to verify snapshot", err)
			utils.HTTPError(w, "Failed to verify snapshot: "+err.Error(), http.StatusInternalServerError, "BCK009")
			return
		}

		var snapshotsArray []map[string]interface{}
		if err := json.Unmarshal([]byte(snapshots), &snapshotsArray); err != nil {
			utils.Error("RestoreBackup: Failed to parse snapshots", err)
			utils.HTTPError(w, "Failed to parse snapshots: "+err.Error(), http.StatusInternalServerError, "BCK010")
			return
		}

		snapshotFound := false
		for _, s := range snapshotsArray {
			if s["id"].(string) == request.SnapshotID {
				snapshotFound = true
				break
			}
		}

		if !snapshotFound {
			utils.HTTPError(w, "Snapshot not found for this backup", http.StatusNotFound, "BCK011")
			return
		}

		CreateRestoreJob(RestoreConfig{
			Repository: backup.Repository,
			Password:   backup.Password,
			SnapshotID: request.SnapshotID,
			Target:     request.Target,
			Name:       backup.Name,
			Include:    request.Include,
			OriginalSource: backup.Source,
		})

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": "Restore job created",
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func ListSnapshotsRouteFromRepo(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		vars := mux.Vars(req)
		name := vars["name"]
		
		backup, exists := utils.GetMainConfig().Backup.Backups[name]
		if !exists {
			utils.HTTPError(w, "Backup not found", http.StatusNotFound, "BCK004")
			return
		}

		output, err := ListSnapshotsWithFilters(backup.Repository, backup.Password, []string{}, "", "")
		if err != nil {
			utils.Error("ListSnapshots: Failed to list snapshots", err)
			utils.HTTPError(w, "Failed to list snapshots: "+err.Error(), http.StatusInternalServerError, "BCK006")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var outputJSON []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &outputJSON); err != nil {
			utils.Error("ListSnapshots: Failed to parse snapshots", err)
			utils.HTTPError(w, "Failed to parse snapshots: "+err.Error(), http.StatusInternalServerError, "BCK007")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":	 outputJSON,
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func ListRepos(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		config := utils.GetMainConfig()
		repos := map[string]utils.SingleBackupConfig{}
		results := map[string]interface{}{}

		for _, backup := range config.Backup.Backups {
			found := false
			if _, exists := repos[backup.Repository]; exists {
				found = true
			}
			
			if !found {
				repos[backup.Repository] = backup

				output, err := StatsRepository(backup.Repository, backup.Password)
				if err != nil {
					utils.Error("ListRepos: Failed to get repository stats", err)
					results[backup.Repository] = map[string]interface{}{
						"status": "error",
						"id": backup.Name,
						"error": err.Error(),
					}
				} else {
					var outputJSON map[string]interface{}
					if err := json.Unmarshal([]byte(output), &outputJSON); err != nil {
						utils.Error("ListRepos: Failed to parse repository stats", err)
						results[backup.Repository] = map[string]interface{}{
							"status": "error",
							"id": backup.Name,
							"error": err.Error(),
						}
					} else {
						results[backup.Repository] = map[string]interface{}{
							"status": "ok",
							"id": backup.Name,
							"stats": outputJSON,
							"path": backup.Repository,
						}
					}
				}
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":	 results,
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func ForgetSnapshotRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "DELETE" {
		vars := mux.Vars(req)
		name := vars["name"]
		snapshot := vars["snapshot"]

		config := utils.GetMainConfig()
		backup, exists := config.Backup.Backups[name]
		if !exists {
			utils.HTTPError(w, "Backup not found", http.StatusNotFound, "BCK004")
			return
		}

		err := ForgetSnapshot(backup.Repository, backup.Password, snapshot)
		if err != nil {
			utils.Error("ForgetSnapshotRoute: Failed to forget snapshot", err)
			utils.HTTPError(w, "Failed to forget snapshot: "+err.Error(), http.StatusInternalServerError, "BCK008")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "OK",
			"message": fmt.Sprintf("Forgot snapshot %s from repository %s", snapshot, backup.Repository),
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func StatsRepositorySubfolderRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		vars := mux.Vars(req)
		name := vars["name"]
		snapshot := vars["snapshot"]
		path := req.URL.Query().Get("path")

		backup, exists := utils.GetMainConfig().Backup.Backups[name]
		if !exists {
			utils.HTTPError(w, "Backup not found", http.StatusNotFound, "BCK004")
			return
		}

		output, err := StatsRepositorySubfolder(backup.Repository, backup.Password, snapshot, path)
		if err != nil {
			utils.Error("StatsRepositorySubfolder: Failed to get repository stats", err)
			utils.HTTPError(w, "Failed to get repository stats: "+err.Error(), http.StatusInternalServerError, "BCK009")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var outputJSON map[string]interface{}
		if err := json.Unmarshal([]byte(output), &outputJSON); err != nil {
			utils.Error("StatsRepositorySubfolder: Failed to parse repository stats", err)
			utils.HTTPError(w, "Failed to parse repository stats: "+err.Error(), http.StatusInternalServerError, "BCK010")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":	 outputJSON,
		})
	} else {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}