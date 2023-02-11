package utils

import (
	"strings"
	"log"
	"os"
	"net/http"
)

func GetRealPath(fullPath string) string {
	var dataPathsObject = map[string]string{
		"data": "/mnt/d/",
		"diskE": "/mnt/e/",
	}

	path := dataPathsObject[strings.Split(fullPath, "/")[0]]
	if path == "" {
		log.Println("No path specified")
	}

	return path + strings.Join(strings.Split(fullPath, "/")[1:], "/")
}

func SetHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func FileExists(path string) bool {
	_, err := os.Stat(path) 
	if err == nil {
		return true
	}
	log.Println(err)
	return false
}