package file

import (
	"log"
	"net/http"
	"io/ioutil"
	"os"
	"encoding/json"
	"../utils"
)

func FileList(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	fullPath := req.URL.Query().Get("path")
	if fullPath == "" {
		log.Println("No path specified")
	}

	filePath := utils.GetRealPath(fullPath)

	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		log.Fatal(err)
	}
	
	// get folder stats
	folderStats, err := os.Stat(filePath)

	// add file FileStats to json object
	var fileStats [](utils.FileStats)
	for _, file := range files {
		fileStats = append(fileStats, utils.FileStats{
			Name: file.Name(),
			Path: fullPath + "/" + file.Name(),
			Size: file.Size(),
			Mode: file.Mode(),
			ModTime: file.ModTime(),
			IsDir: file.IsDir(),
		})
	}

	// return json object
	// return json object with metadata FileStat and content
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Metadata": utils.FileStats{
			Name: folderStats.Name(),
			Size: folderStats.Size(),
			Mode: folderStats.Mode(),
			ModTime: folderStats.ModTime(),
			IsDir: folderStats.IsDir(),
		},
		"Content": fileStats,
	})
}