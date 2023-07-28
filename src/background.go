package main

import (
	"io/ioutil"
	"os"
	"net/http"
	"path/filepath"
	"io"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils" 
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

func UploadBackground(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if(req.Method == "POST") {
		// parse the form data
		err := req.ParseMultipartForm(1 << 20)
		if err != nil {
			utils.HTTPError(w, "Error parsing form data", http.StatusInternalServerError, "FORM001")
			return
		}

		// retrieve the file part of the form
		file, header, err := req.FormFile("background")
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
		dst, err := os.Create("/config/background" + ext)
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

func GetBackground(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	vars := mux.Vars(req)
	ext := vars["ext"]

	if !validExtensions["." + ext] {
		utils.HTTPError(w, "Invalid file extension", http.StatusBadRequest, "FILE001")
		return
	}

	if(req.Method == "GET") {
		// get the background image
		bg, err := ioutil.ReadFile("/config/background." + ext)
		if err != nil {
			utils.HTTPError(w, "Error reading background image", http.StatusInternalServerError, "FILE003")
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