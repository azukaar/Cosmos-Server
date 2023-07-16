package main

import (
	"net/http"
	"net/url"
	// "fmt"
	"strings"
	"os"
	"io/ioutil"
	"encoding/json"
	"path"
	"time"
	"context"

	"go.deanishe.net/favicon"

	"github.com/azukaar/cosmos-server/src/utils" 
)

type CachedImage struct {
	ContentType string
	ETag string
	Body []byte
}

func httpGetWithTimeout(url string) (*http.Response, error) {
	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

var IconCache = make(map[string]CachedImage)

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

var IconCacheLock = make(chan bool, 1)

func GetFavicon(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	// get url from query string
	escsiteurl := req.URL.Query().Get("q")
	
	IconCacheLock <- true
	defer func() { <-IconCacheLock }()
	
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
		if _, ok := IconCache[siteurl]; ok {
			utils.Debug("Favicon in cache")
			resp := IconCache[siteurl]
			sendImage(w, resp)
			return
		}

		var icons []*favicon.Icon
		var defaultIcons = []*favicon.Icon{
			&favicon.Icon{URL: "favicon.png", Width: 0},
			&favicon.Icon{URL: "/favicon.png", Width: 0},
			&favicon.Icon{URL: "favicon.ico", Width: 0},
			&favicon.Icon{URL: "/favicon.ico", Width: 0},
		}

		// follow siteurl and check if any redirect. 
		
		respNew, err := httpGetWithTimeout(siteurl)

		if err != nil {
			utils.Error("FaviconFetch", err)
			icons = append(icons, defaultIcons...)
		} else {
			siteurl = respNew.Request.URL.String()
			icons, err = favicon.Find(siteurl)

			if err != nil || len(icons) == 0 {
				icons = append(icons, defaultIcons...)
			} else {
				// Check if icons list is missing any default values
				for _, defaultIcon := range defaultIcons {
						found := false
						for _, icon := range icons {
								if icon.URL == defaultIcon.URL {
										found = true
										break
								}
						}
						if !found {
							icons = append(icons, defaultIcon)
						}
				}
			}
		}

		for _, icon := range icons {
			if icon.Width <= 256 {

				iconURL := icon.URL
				u, err := url.Parse(siteurl)
				if err != nil {
					utils.Debug("FaviconFetch failed to parse " + err.Error())
					continue
				}
				
				if !strings.HasPrefix(iconURL, "http") {
					if strings.HasPrefix(iconURL, ".") {
						// Relative URL starting with "."
						// Resolve the relative URL based on the base URL
						baseURL := u.Scheme + "://" + u.Host
						iconURL = baseURL + iconURL[1:]
					} else if strings.HasPrefix(iconURL, "/") {
						// Relative URL starting with "/"
						// Append the relative URL to the base URL
						iconURL = u.Scheme + "://" + u.Host + iconURL
					} else {
						// Relative URL without starting dot or slash
						// Construct the absolute URL based on the current page's URL path
						baseURL := u.Scheme + "://" + u.Host
						baseURLPath := path.Dir(u.Path)
						iconURL = baseURL + baseURLPath + "/" + iconURL
					}
				}
				
				utils.Debug("Favicon Trying to fetch " + iconURL)

				// Fetch the favicon
				resp, err := httpGetWithTimeout(iconURL)
				if err != nil {
					utils.Debug("FaviconFetch" + err.Error())
					continue
				}

				// check if 200 and if image 
				if resp.StatusCode != 200 {
					utils.Debug("FaviconFetch - " + iconURL + " - not 200 ")
					continue
				} else if !strings.Contains(resp.Header.Get("Content-Type"), "image") && !strings.Contains(resp.Header.Get("Content-Type"), "octet-stream") {
					utils.Debug("FaviconFetch - " + iconURL + " - not image ")
					continue
				} else {
					utils.Log("Favicon found " + iconURL)

					// Cache the response 
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						utils.Debug("FaviconFetch - cant read " + err.Error())
						continue
					}
					
					finalImage := CachedImage{
						ContentType: resp.Header.Get("Content-Type"),
						ETag: resp.Header.Get("ETag"),
						Body: body,
					}

					IconCache[siteurl] = finalImage

					sendImage(w, IconCache[siteurl])
					return
				}
			}
	} 
	utils.Log("Favicon final fallback")
	sendFallback(w)
	return
	
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

		resp, err := httpGetWithTimeout(siteurl)
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