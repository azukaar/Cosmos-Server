package main

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"go.deanishe.net/favicon"

	"github.com/azukaar/cosmos-server/src/docker"
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
type result struct {
	IconURL     string
	CachedImage CachedImage
	Error       error
}

func faviconTarget(routeName string, clientID string) (siteurl string, container string, err error) {
	config := utils.GetMainConfig()

	if routeName != "" {
		for _, route := range config.HTTPConfig.ProxyConfig.Routes {
			if route.Name != routeName {
				continue
			}
			switch route.Mode {
			case "SERVAPP":
				target, perr := url.Parse(route.Target)
				if perr != nil {
					return "", "", perr
				}
				return route.Target, target.Hostname(), nil
			case "PROXY":
				return route.Target, "", nil
			case "STATIC":
				return "", "", nil
			default:
				// Served by Cosmos itself: fetch through the public origin.
				origin := strings.TrimSuffix(utils.GetServerURL(""), "/")
				if route.UseHost {
					scheme := "http://"
					if utils.IsHTTPS {
						scheme = "https://"
					}
					origin = scheme + route.Host
				}
				if route.UsePathPrefix {
					origin += route.PathPrefix
				}
				return origin, "", nil
			}
		}
		return "", "", errors.New("no route named " + routeName)
	}

	if clientID != "" {
		for _, client := range config.OpenIDClients {
			if client.ID != clientID {
				continue
			}
			// Clients may list several redirect URIs; the first one is the application's origin.
			first := strings.TrimSpace(strings.Split(client.Redirect, ",")[0])
			redirect, perr := url.Parse(first)
			if perr != nil {
				return "", "", perr
			}
			if redirect.Scheme == "" || redirect.Host == "" {
				return "", "", errors.New("OpenID client " + clientID + " has no usable redirect URI")
			}
			return redirect.Scheme + "://" + redirect.Host, "", nil
		}
		return "", "", errors.New("no OpenID client named " + clientID)
	}

	return "", "", errors.New("favicon request names neither a route nor an OpenID client")
}

func faviconDialURL(siteurl string, container string) string {
	if container == "" || (utils.IsInsideContainer && !utils.IsHostNetwork) {
		return siteurl
	}
	ip, _ := utils.GetContainerIPByName(container)
	if ip == "" {
		return siteurl
	}
	parsed, err := url.Parse(siteurl)
	if err != nil {
		return siteurl
	}
	host := ip
	if port := parsed.Port(); port != "" {
		host += ":" + port
	}
	parsed.Host = host
	return parsed.String()
}

// GetFavicon godoc
// @Summary Get favicon for a route or OpenID client
// @Description Fetches and caches the favicon of a configured route or OpenID client, returning it as an image. Exactly one of route or openid must be given.
// @Tags system
// @Produce octet-stream
// @Security BearerAuth
// @Param route query string false "Name of the configured route"
// @Param openid query string false "ID of the configured OpenID client"
// @Success 200 {file} binary
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/favicon [get]
func GetFavicon(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil {
		return
	}

	if req.Method != "GET" {
		utils.Error("Favicon: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	siteurl, container, err := faviconTarget(req.URL.Query().Get("route"), req.URL.Query().Get("openid"))
	if err != nil {
		utils.Error("Favicon: target", err)
		utils.HTTPError(w, "Favicon target", http.StatusNotFound, "FA001")
		return
	}
	if siteurl == "" {
		sendFallback(w)
		return
	}

	// Keyed by the configured URL, not the dialed address: container IPs and redirects change and must not invalidate a cached icon.
	cacheKey := siteurl

	IconCacheLock <- true
	defer func() { <-IconCacheLock }()

	if cached, ok := IconCache[cacheKey]; ok {
		utils.Debug("Favicon in cache")
		sendImage(w, cached)
		return
	}

	// A dormant lazy container must never be woken for an icon; serve the placeholder.
	if container != "" && docker.LazyIsDormant(container) {
		utils.Debug("Favicon: " + container + " is dormant, serving fallback")
		sendFallback(w)
		return
	}

	siteurl = faviconDialURL(siteurl, container)
	utils.Log("Fetch favicon for " + siteurl)

	var icons []*favicon.Icon
	var defaultIcons = []*favicon.Icon{
		{URL: "/favicon.ico", Width: 0},
		{URL: "/favicon.png", Width: 0},
		{URL: "favicon.ico", Width: 0},
		{URL: "favicon.png", Width: 0},
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

	// Create a channel to collect favicon fetch results
	resultsChan := make(chan result)
	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Loop through each icon and start a goroutine to fetch it
	for _, icon := range icons {
		if icon.Width <= 256 {
			wg.Add(1)
			go func(icon *favicon.Icon) {
				defer wg.Done()
				fetchAndCacheIcon(icon, siteurl, resultsChan)
			}(icon)
		}
	}

	// Close the results channel when all fetches are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect the results
	for result := range resultsChan {
		IconCache[cacheKey] = result.CachedImage
		sendImage(w, IconCache[cacheKey])
		return
	}

	utils.Log("Favicon final fallback")
	sendFallback(w)
}

// fetchAndCacheIcon is a helper function to fetch and cache the icon
func fetchAndCacheIcon(icon *favicon.Icon, baseSiteURL string, resultsChan chan<- result) {
	iconURL := icon.URL
	u, err := url.Parse(baseSiteURL)
	if err != nil {
			utils.Debug("FaviconFetch failed to parse " + err.Error())
			return
	}

	if !strings.HasPrefix(iconURL, "http") {
			// Process the iconURL to make it absolute
			iconURL = resolveIconURL(iconURL, u)
	}

	utils.Debug("Favicon Trying to fetch " + iconURL)

	// Fetch the favicon
	resp, err := httpGetWithTimeout(iconURL)
	if err != nil {
			utils.Debug("FaviconFetch - " + err.Error())
			return
	}
	defer resp.Body.Close()

	// Check if response is successful and content type is image
	if resp.StatusCode != 200 || (!strings.Contains(resp.Header.Get("Content-Type"), "image") && !strings.Contains(resp.Header.Get("Content-Type"), "octet-stream")) {
			utils.Debug("FaviconFetch - " + iconURL + " - not 200 or not image ")
			return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
			utils.Debug("FaviconFetch - can't read " + err.Error())
			return
	}

	// Prepare the cached image
	cachedImage := CachedImage{
			ContentType: resp.Header.Get("Content-Type"),
			ETag:        resp.Header.Get("ETag"),
			Body:        body,
	}

	// Send the result back via the channel
	resultsChan <- result{IconURL: iconURL, CachedImage: cachedImage}
}

// resolveIconURL processes the iconURL to make it an absolute URL if it is relative
func resolveIconURL(iconURL string, baseURL *url.URL) string {
	if strings.HasPrefix(iconURL, ".") {
			// Relative URL starting with "."
			// Resolve the relative URL based on the base URL
			return baseURL.Scheme + "://" + baseURL.Host + iconURL[1:]
	} else if strings.HasPrefix(iconURL, "/") {
			// Relative URL starting with "/"
			// Append the relative URL to the base URL
			return baseURL.Scheme + "://" + baseURL.Host + iconURL
	} else {
			// Relative URL without starting dot or slash
			// Construct the absolute URL based on the current page's URL path
			baseURLPath := path.Dir(baseURL.Path)
			if baseURLPath == "." {
					baseURLPath = ""
			}
			return baseURL.Scheme + "://" + baseURL.Host + baseURLPath + "/" + iconURL
	}
}

// SendLogo godoc
// @Summary Get Cosmos logo
// @Description Returns the Cosmos server logo as a PNG image
// @Tags system
// @Produce png
// @Success 200 {file} binary
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /logo [get]
func SendLogo(w http.ResponseWriter, req *http.Request) {
	pwd,_ := os.Getwd()
	imgsrc := "Logo2.png"
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