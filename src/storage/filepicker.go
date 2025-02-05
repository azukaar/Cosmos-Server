package storage

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"io/ioutil"
	"syscall"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
)

type StorageInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type DirectoryListing struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
	Ext   string `json:"ext"`
	Created int64  `json:"created"`
	UID   uint32 `json:"uid"`
	GID   uint32 `json:"gid"`
	FullPath string `json:"fullPath"`
}

func ListDirectoryRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		//config := utils.GetMainConfig()
		storage := req.URL.Query().Get("storage")
		path := req.URL.Query().Get("path")

		if storage == "" {
			storage = "local"
			// TODO: Restore when parent button implemented
			// if path == "" {
			// 	path = config.DockerConfig.DefaultDataPath
			// }
		} 

		if path == "" {
			path = "/"
		}

		storages, err := ListStorage()
		if err != nil {
			utils.Error("ListDirectoryRoute: Error listing storages: "+err.Error(), nil)
			utils.HTTPError(w, "Internal server error", http.StatusInternalServerError, "STO002")
			return
		}

		var basePath string
		if storage == "local" {
			basePath = "/"
			if utils.IsInsideContainer {
				basePath = "/mnt/host/"
			}
		} else {
			found := false
			for _, s := range storages {
				if s.Name == storage {
					basePath = s.Path
					found = true
					break
				}
			}
			if !found {
				utils.Error("ListDirectoryRoute: Storage not found: "+storage, nil)
				utils.HTTPError(w, "Storage not found", http.StatusNotFound, "STO001")
				return
			}
		}

		fullPath := filepath.Join(basePath, path)

		directory, err := ListDirectory(fullPath)
		if err != nil {
			utils.Error("ListDirectoryRoute: Error listing directory: "+err.Error(), nil)
			utils.HTTPError(w, "Internal server error", http.StatusInternalServerError, "STO003")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
				"storage":	 storage,
				"path":    path,
				"storages":  storages,
				"directory": directory,
			},
		})
	} else {
		utils.Error("ListDirectoryRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func ListStorage() ([]StorageInfo, error) {
	return append([]StorageInfo{
		{Name: "local", Path: "/"},
	}, CachedRemoteStorageList...), nil
}

func ListDirectory(path string) ([]DirectoryListing, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var listings []DirectoryListing
	for _, file := range files {
		fileType := "file"
		if file.IsDir() {
			fileType = "directory"
		}

		// Get system-specific file info
		sys := file.Sys()
		var uid, gid uint32
		if sys != nil {
			if stat, ok := sys.(*syscall.Stat_t); ok {
				uid = stat.Uid
				gid = stat.Gid
			}
		}

		fullPath := filepath.Join(path, file.Name())

		if utils.IsInsideContainer {
			fullPath = strings.TrimPrefix(fullPath, "/mnt/host")
		}

		listing := DirectoryListing{
			Name:  file.Name(),
			Ext:   filepath.Ext(file.Name()),
			Type:  fileType,
			Size:  file.Size(),
			IsDir: file.IsDir(),
			Created: file.ModTime().Unix(),
			UID:   uid,
			GID:   gid,
			FullPath: fullPath,
		}
		listings = append(listings, listing)
	}

	return listings, nil
}