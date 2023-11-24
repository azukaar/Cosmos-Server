package main

import (
	"io/ioutil"
	"os"
	"net/http"
	"path/filepath"
	"io"
	"encoding/json"
	"strings"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils" 
	"github.com/azukaar/cosmos-server/src/docker" 
)

var validExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".svg":  true,
	".webp": true,
	".tiff": true,
	".avif": true,
}

func UploadImage(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}
	
	vars := mux.Vars(req)
	originalName := vars["name"]

	name := originalName + "-" + utils.GenerateRandomString(6)

	// if name includes / or ..
	if filepath.Clean(name) != name || strings.Contains(name, "/") {
		utils.HTTPError(w, "Invalid file name", http.StatusBadRequest, "FILE002")
		return
	}

	if(req.Method == "POST") {
		// if the uploads directory does not exist, create it
		if _, err := os.Stat(utils.CONFIGFOLDER + "/uploads"); os.IsNotExist(err) {
			os.Mkdir(utils.CONFIGFOLDER + "/uploads", 0750)
		}

		// parse the form data
		err := req.ParseMultipartForm(1 << 20)
		if err != nil {
			utils.HTTPError(w, "Error parsing form data", http.StatusInternalServerError, "FORM001")
			return
		}

		// retrieve the file part of the form
		file, header, err := req.FormFile("image")
		if err != nil {
			utils.HTTPError(w, "Error retrieving file from form data", http.StatusInternalServerError, "FORM002")
			return
		}
		defer file.Close()
		
		// get the file extension
		ext := filepath.Ext(header.Filename)
		
		if !validExtensions[ext] {
			utils.HTTPError(w, "Invalid file extension " + ext, http.StatusBadRequest, "FILE001")
			return
		}

		// create a new file in the config directory
		dst, err := os.Create(utils.CONFIGFOLDER + "/uploads/" + name + ext)
		if err != nil {
			utils.HTTPError(w, "Error creating destination file", http.StatusInternalServerError, "FILE004")
			return
		}
		defer dst.Close()

		// copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			utils.HTTPError(w, "Error writing to destination file", http.StatusInternalServerError, "FILE005")
			return
		}

		// return a response to the client
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
				"path": "/cosmos/api/image/" + name + ext,
				"filename": header.Filename,
				"size": header.Size,
				"extension": ext,
			},
		})

	} else {
		utils.Error("UploadBackground: Method not allowed - " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func GetImage(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	utils.Log("API: GetImage")

	vars := mux.Vars(req)
	name := vars["name"]

	// if name includes / or ..
	if filepath.Clean(name) != name || strings.Contains(name, "/") {
		utils.Error("GetBackground: Invalid file name - " + name, nil)
		utils.HTTPError(w, "Invalid file name", http.StatusBadRequest, "FILE002")
		return
	}

	// get the file extension
	ext := filepath.Ext(name)

	if !validExtensions[ext] {
		utils.Error("GetBackground: Invalid file extension - " + ext, nil)
		utils.HTTPError(w, "Invalid file extension", http.StatusBadRequest, "FILE001")
		return
	}

	if(req.Method == "GET") {
		// get the background image
		bg, err := ioutil.ReadFile(utils.CONFIGFOLDER + "/uploads/" + name)
		if err != nil {
			utils.Error("GetBackground: Error reading image - " + name, err)
			utils.HTTPError(w, "Error reading image", http.StatusInternalServerError, "FILE003")
			return
		}

		// return a response to the client
		w.Header().Set("Content-Type", "image/" + ext)
		w.Write(bg)

	} else {
		utils.Error("GetBackground: Method not allowed - " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func imageCleanUp() {
	utils.Log("Image cleanup")

	config := utils.GetMainConfig()
	images := map[string]bool{}
	images[config.HomepageConfig.Background] = true

	for _, route := range config.HTTPConfig.ProxyConfig.Routes {
		if(route.Icon != "") {
			images[route.Icon] = true
		}
	}

	// get containers
	containers, err := docker.ListContainers()

	if err != nil {
		utils.Error("Image cleanup: Error getting containers", err)
		return
	}

	for _, container := range containers {
		if(container.Labels["cosmos-icon"] != "") {
			images[container.Labels["cosmos-icon"]] = true
		}
	}

	fmt.Println(images)

	// if the uploads directory does not exist, return
	if _, err := os.Stat(utils.CONFIGFOLDER + "/uploads"); os.IsNotExist(err) {
		return
	}

	// get the files in the uploads directory
	files, err := ioutil.ReadDir(utils.CONFIGFOLDER + "/uploads")
	if err != nil {
		utils.Error("Image cleanup: Error reading directory", err)
		return
	}

	// loop through the files
	base := "/cosmos/api/image/"
	for _, f := range files {
		if(!images[base + f.Name()]) {
			err := os.Remove(utils.CONFIGFOLDER + "/uploads/" + f.Name())
			if err != nil {
				utils.Error("Image cleanup: Error removing file", err)
			}
		}
	}
}