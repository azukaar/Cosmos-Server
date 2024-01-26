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
	"context"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils" 
	"github.com/azukaar/cosmos-server/src/docker" 

	"go.mongodb.org/mongo-driver/mongo"
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

func MigratePre013() {
	// read background from config.json
	config := utils.ReadConfigFromFile()
	bg := config.HomepageConfig.Background

	utils.Debug("MigratePre013: background is" + bg)

	// if the background start with /cosmos/api/background/
	if(strings.HasPrefix(bg, "/cosmos/api/background/")) {
		// get the background name
		ext := strings.TrimPrefix(bg, "/cosmos/api/background/")
		bgPath := utils.CONFIGFOLDER + "/background." + ext

		utils.Debug("MigratePre013: background path is" + bgPath)

		// if the background file exists
		if _, err := os.Stat(bgPath); err == nil {
			// move the background file to uploads
			err := os.Rename(bgPath, utils.CONFIGFOLDER + "/uploads/background." + ext)
			if err != nil {
				utils.Error("MigratePre013: Error moving background file", err)
			}

			// update the background path
			config.HomepageConfig.Background = "/cosmos/api/image/background." + ext
			utils.SetBaseMainConfig(config)
		}
	}
}

func CommitMigrationPre014() {
	config := utils.ReadConfigFromFile()
	config.LastMigration = "0.14.0"
	utils.SetBaseMainConfig(config)
}

func MigratePre014() {
	config := utils.ReadConfigFromFile()
	if config.LastMigration != "" {
		c, err := utils.CompareSemver(config.LastMigration, "0.14.0")
		if err != nil {
			utils.Fatal("Can't read field 'lastMigration' from config", err)
		}
		if(c > 0) {
			return
		}
	}

	if _, err := os.Stat(utils.CONFIGFOLDER + "database"); err != nil {
		utils.Log("MigratePre014: No database found, trying to migrate pre-0.14 data")

		cu2, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
		if errCo != nil {
			// Assuming we dont need to migrate
			utils.Log("MigratePre014: No database, assuming we dont need to migrate")
			utils.Debug(errCo.Error())
			CommitMigrationPre014()
			return
		}

		cd2, errCo := utils.GetCollection(utils.GetRootAppId(), "devices")
		if errCo != nil {
			// Assuming we dont need to migrate
			utils.Log("MigratePre014: No database, assuming we dont need to migrate")
			utils.Debug(errCo.Error())
			CommitMigrationPre014()
			return
		}

		utils.Log("MigratePre014: Migrating pre-0.14 data")

		cu, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")

		if errCo != nil {
			utils.Fatal("Error trying to migrate pre-0.14 data [1]", errCo)
			return
		}

		// copy users from cu2 to cu
		cursor, err := cu2.Find(nil, map[string]interface{}{})
		if err != nil && err != mongo.ErrNoDocuments {
			utils.Fatal("Error trying to migrate pre-0.14 data [3]", err)
			return
		}
		defer cursor.Close(nil)

		users := []utils.User{}

		if err = cursor.All(nil, &users); err != nil {
			utils.Fatal("Error trying to migrate pre-0.14 data [4]", err)
			return
		}

		for _, user := range users {
			_, err := cu.InsertOne(nil, user)
			if err != nil {
				utils.Fatal("Error trying to migrate pre-0.14 data [5]", err)
				return
			}
		}
		
		closeDb()

		utils.Log("MigratePre014: Migrated " + fmt.Sprint(len(users)) + " users")
		
		
		cd, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
		defer closeDb()

		if errCo != nil {
			utils.Fatal("Error trying to migrate pre-0.14 data [2]", errCo)
			return
		}

		// copy devices from cd2 to cd
		cursor, err = cd2.Find(nil, map[string]interface{}{})
		if err != nil && err != mongo.ErrNoDocuments {
			utils.Fatal("Error trying to migrate pre-0.14 data [6]", err)
			return
		}

		defer cursor.Close(nil)

		devices := []utils.ConstellationDevice{}

		if err = cursor.All(context.Background(), &devices); err != nil {
			utils.Fatal("Error trying to migrate pre-0.14 data [7]", err)
			return
		}

		for _, device := range devices {
			_, err := cd.InsertOne(context.Background(), device)
			if err != nil {
				utils.Fatal("Error trying to migrate pre-0.14 data [8]", err)
				return
			}
		}
		
		utils.Log("MigratePre014: Migrated " + fmt.Sprint(len(devices)) + " devices")

		CommitMigrationPre014()
	}
}