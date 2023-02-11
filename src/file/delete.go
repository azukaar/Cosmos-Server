package file

import (
	"log"
	"net/http"
	"encoding/json"
	"os"
	"../utils"
)

func FileDelete(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	fullPath := req.URL.Query().Get("path")
	if fullPath == "" {
		log.Println("No path specified")
	}

	filePath := utils.GetRealPath(fullPath)

	// delete file
	err := os.Remove(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// return json object
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status": "OK",
	})
}