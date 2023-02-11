package file

import (
	"log"
	"net/http"
	"io"
	"encoding/json"
	"os"
	"../utils"
)

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func FileCopy(w http.ResponseWriter, req *http.Request) {
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
	err := copyFile(filePath, destination)	
	if err != nil {
		log.Fatal(err)
	}

	// return json object	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status": "OK",
	})
}