package main

import (
	"net/http"
	"net/url"
	// "fmt"
	"strings"
	"os"
	"io/ioutil"
	"regexp"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type CachedImage struct {
	ContentType string
	ETag string
	Body []byte
}

var cache = make(map[string]CachedImage)

func ExtractFaviconMetaTag(html string) string {
	// Regular expression pattern to match the favicon metatag
	pattern := `<link[^>]*rel="icon"[^>]*href="([^"]+)"[^>]*>`

	// Compile the regular expression pattern
	regex := regexp.MustCompile(pattern)

	// Find the first match in the HTML string
	match := regex.FindStringSubmatch(html)

	if len(match) > 1 {
		// Extract the URL from the matched metatag
		faviconURL := match[1]
		return faviconURL
	}

	return "/favicon.ico"
}

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

		// follow siteurl and check if any redirect. 
		respNew, err := http.Get(siteurl)
		if err != nil {
			utils.Error("FaviconFetch", err)
			sendFallback(w)
			return
		}

		urlBody, err := ioutil.ReadAll(respNew.Body)
		if err != nil {
			utils.Error("FaviconFetch", err)
			sendFallback(w)
			return
		}

		// get favicon meta tag from the response
		faviconURL := ExtractFaviconMetaTag((string)(urlBody))

		// if faviconURL is relative get hostname of the URL and append icon
		if !strings.HasPrefix(faviconURL, "http") {
			u, err := url.Parse(siteurl)
			if err != nil {
				utils.Error("FaviconFetch", err)
				sendFallback(w)
				return
			}
			if !strings.HasPrefix(faviconURL, "/") {
				faviconURL = "/" + faviconURL
			}
			faviconURL = u.Scheme + "://" + u.Host + faviconURL
		}
		
		utils.Log("Favicon: " + faviconURL)

		// Fetch the favicon
		resp, err := http.Get(faviconURL)
		if err != nil {
			utils.Error("FaviconFetch", err)
			sendFallback(w)
			return
		}

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

func SendLogo(w http.ResponseWriter, req *http.Request) {
	pwd,_ := os.Getwd()
	imgsrc := "Logo.png"
	Logo, err := ioutil.ReadFile(pwd + "/" + imgsrc)
	if err != nil {
		utils.Error("Logo", err)
		utils.HTTPError(w, "Favicon", http.StatusInternalServerError, "FA003")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=5")
	w.WriteHeader(http.StatusOK)
	w.Write(Logo)
}