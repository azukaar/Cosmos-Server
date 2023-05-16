package main

import (
	"net/http"
	"net/url"
	// "fmt"
	"strings"
	"os"
	"io/ioutil"
	"regexp"
	"path"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type CachedImage struct {
	ContentType string
	ETag string
	Body []byte
}

var cache = make(map[string]CachedImage)

func ExtractFaviconMetaTag(filename string) string {
	// Read the contents of the file
	htmlBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "/favicon.ico"
	}
	html := string(htmlBytes)

	// Regular expression pattern to match the favicon metatag
	pattern := `<link[^>]*rel="icon"[^>]*(?:sizes="([^"]+)")?[^>]*href="([^"]+)"[^>]*>|<meta[^>]*name="msapplication-TileImage"[^>]*content="([^"]+)"[^>]*>`

	// Compile the regular expression pattern
	regex := regexp.MustCompile(pattern)

	// Find all matches in the HTML string
	matches := regex.FindAllStringSubmatch(html, -1)

	var faviconURL string

	// Iterate over the matches to find the appropriate favicon
	for _, match := range matches {
		sizes := match[1]
		href := match[2]
		msAppTileImage := match[3]

		// Check if the meta tag specifies msapplication-TileImage
		if msAppTileImage != "" {
			faviconURL = msAppTileImage
			break
		}

		// Check if the sizes attribute contains 96x96
		if strings.Contains(sizes, "96x96") {
			faviconURL = href
			break
		}

		// Check if the sizes attribute contains 64x64
		if strings.Contains(sizes, "64x64") {
			faviconURL = href
			continue
		}

		// Check if the sizes attribute contains 32x32
		if strings.Contains(sizes, "32x32") {
			faviconURL = href
			continue
		}

		// If no sizes specified, set faviconURL to the first match without sizes
		if faviconURL == "" && sizes == "" {
			faviconURL = href
		}
	}

	// If a favicon URL is found, return it
	if faviconURL != "" {
		return faviconURL
	}

	// Return an error if no favicon URL is found
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

			if strings.HasPrefix(faviconURL, ".") {
				// Relative URL starting with "."
				// Resolve the relative URL based on the base URL
				baseURL := u.Scheme + "://" + u.Host
				faviconURL = baseURL + faviconURL[1:]
			} else if strings.HasPrefix(faviconURL, "/") {
				// Relative URL starting with "/"
				// Append the relative URL to the base URL
				faviconURL = u.Scheme + "://" + u.Host + faviconURL
			} else {
				// Relative URL without starting dot or slash
				// Construct the absolute URL based on the current page's URL path
				baseURL := u.Scheme + "://" + u.Host
				baseURLPath := path.Dir(u.Path)
				faviconURL = baseURL + baseURLPath + "/" + faviconURL
			}
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