package file

import (
	"log"
	"net/http"
	"encoding/json"
	"os"
	"../utils"
)

func FileMove(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	fullPath := req.URL.Query().Get("path")
	if fullPath == "" {
		log.Println("No path specified")
	}

	filePath := utils.GetRealPath(fullPath)

	fullDestination := req.URL.Query().Get("destination")
	if fullDestination == "" {
		log.Println("No destination specified")
	}

	destination := utils.GetRealPath(fullDestination)

	// copy file to destination

	err := os.Rename(filePath, destination)
	if err != nil {
		log.Fatal(err)
	}

	// return json object	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status": "OK",
	})
}