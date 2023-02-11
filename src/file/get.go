package file

import (
	"log"
	"net/http"
	"io"
	"strings"
	"os"
	// "bytes"
	// "mime/multipart"
	"bufio"
	"strconv"
	"../utils"
)

func getExtension(path string) string {
	return strings.Split(path, ".")[len(strings.Split(path, ".")) - 1]
}

func getContentType(path string) string {
	switch getExtension(path) {
	case "html":
		return "text/html"
	case "css":
		return "text/css"
	case "js":
		return "application/javascript"
	case "png":
		return "image/png"
	case "jpg":
		return "image/jpeg"
	case "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "svg":
		return "image/svg+xml"
	case "mp4":
		return "video/mp4"
	case "mkv":
		return "video/x-matroska"
	case "webm":
		return "video/webm"
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "ogg":
		return "audio/ogg"
	default:
		return "text/plain"
	}
}

func FileGet(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	fullPath := req.URL.Query().Get("path")
	if fullPath == "" {
		log.Println("No path specified")
	}

	filePath := utils.GetRealPath(fullPath)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	/*fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}*/

	// set  header content type depending on file type 
	w.Header().Set("Content-Type", getContentType(filePath))

	// multipart file send
	if getExtension(filePath) == "mp4" || getExtension(filePath) == "mkv" || getExtension(filePath) == "webm" {
		w.Header().Set("Content-Disposition", "attachment; filename=" + filePath)
	}

	// open stat file
	stat, err := os.Stat(filePath)

	
	buffer := bufio.NewReader(file)
	
	// set content-length 
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	//copy buffer to client
	io.Copy(w, buffer)

	/*
	// get file stats
	fileStats, err := os.Stat(filePath)

	// return json object with metadata FileStat and content
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Metadata": FileStats{
			Name: fileStats.Name(),
			Path: filePath,
			Size: fileStats.Size(),
			Mode: fileStats.Mode(),
			ModTime: fileStats.ModTime(),
			IsDir: fileStats.IsDir(),
		},
		"Content": string(file),
	})*/
}