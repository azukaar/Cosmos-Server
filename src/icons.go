package main

import (
	"net/http"
	"net/url"
	// "fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"strconv"

	"github.com/azukaar/cosmos-server/src/utils" 
	"go.deanishe.net/favicon"
)

type CachedImage struct {
	ContentType string
	ETag string
	Body []byte
}

var cache = make(map[string]CachedImage)

func sendImage(w http.ResponseWriter, image CachedImage) {
		// Copy the response to the output
		w.Header().Set("Content-Type", image.ContentType)
		w.Header().Set("ETag", image.ETag)
		w.Header().Set("Cache-Control", "max-age=86400")
		w.WriteHeader(http.StatusOK)
		w.Write(image.Body)
}

func sendFallback(w http.ResponseWriter) {
	// Send the fallback image
	pwd,_ := os.Getwd()
	imgsrc := "cosmos_gray.png"
	fallback, err := ioutil.ReadFile(pwd + "/" + imgsrc)
	if err != nil {
		utils.Error("Favicon: fallback", err)
		utils.HTTPError(w, "Favicon", http.StatusInternalServerError, "FA003")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=5")
	w.WriteHeader(http.StatusOK)
	w.Write(fallback)
}


func GetFavicon(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	// get url from query string
	escsiteurl := req.URL.Query().Get("q")
	
	// URL decode
	siteurl, err := url.QueryUnescape(escsiteurl)
	if err != nil {
		utils.Error("Favicon: URL decode", err)
		utils.HTTPError(w, "URL decode", http.StatusInternalServerError, "FA002")
		return
	}

	
	if(req.Method == "GET") { 
		utils.Log("Fetch favicon for " + siteurl)

		// Check if we have the favicon in cache
		if _, ok := cache[siteurl]; ok {
			utils.Debug("Favicon in cache")
			resp := cache[siteurl]
			sendImage(w, resp)
			return
		}

		icons, err := favicon.Find(siteurl)
		utils.Debug("Found Favicon: " + strconv.Itoa(len(icons)))
		if err != nil {
			utils.Error("FaviconFetch", err)
			sendFallback(w)
			return
		}
		
		if len(icons) == 0 {
			utils.Error("FaviconFetch", err)
			sendFallback(w)
			return
		}

		iconIndex := len(icons)-1
		iconChanged := false

		for i, icon := range icons {
			utils.Debug("Favicon Width: " + icon.URL + " " + strconv.Itoa(icon.Width))
			if icon.Width <= 256 {
				iconIndex = i
				iconChanged = true
				break
			}
		}
		if !iconChanged {
			iconIndex = 0
		}
		icon := icons[iconIndex]

		utils.Log("Favicon: " + icon.URL)

		// Fetch the favicon
		resp, err := http.Get(icon.URL)
		if err != nil {
			utils.Error("FaviconFetch", err)
			sendFallback(w)
			return
		}

		// save the body to a file
		// out, err := os.Create("favicon.ico")
		// if err != nil {
		// 	utils.Error("FaviconFetch", err)
		// 	utils.HTTPError(w, "Favicon Fetch", http.StatusInternalServerError, "FA001")
		// 	return
		// }
		// defer out.Close()
		// _, err = io.Copy(out, resp.Body)
		// if err != nil {
		// 	utils.Error("FaviconFetch", err)
		// 	utils.HTTPError(w, "Favicon Fetch", http.StatusInternalServerError, "FA001")
		// 	return
		// }

		// Cache the response 
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			utils.Error("FaviconFetch", err)
			sendFallback(w)
			return
		}
		
		finalImage := CachedImage{
			ContentType: resp.Header.Get("Content-Type"),
			ETag: resp.Header.Get("ETag"),
			Body: body,
		}

		cache[siteurl] = finalImage

		sendImage(w, cache[siteurl])
	} else {
		utils.Error("Favicon: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func PingURL(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	// get url from query string
	escsiteurl := req.URL.Query().Get("q")
	
	// URL decode
	siteurl, err := url.QueryUnescape(escsiteurl)
	if err != nil {
		utils.Error("Ping: URL decode", err)
		utils.HTTPError(w, "Ping URL decode", http.StatusInternalServerError, "FA002")
		return
	}

	
	if(req.Method == "GET") { 
		utils.Log("Ping for " + siteurl)

		resp, err := http.Get(siteurl)
		if err != nil {
			utils.Error("Ping", err)
			utils.HTTPError(w, "URL decode", http.StatusInternalServerError, "PI0001")
			return
		}
		
		if resp.StatusCode >= 500 {
			utils.Error("Ping", err)
			utils.HTTPError(w, "URL decode", http.StatusInternalServerError, "PI0002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
			},
		})
	} else {
		utils.Error("Favicon: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}