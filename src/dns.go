package main

import (
	"net/http"
	"encoding/json"
	"net"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils" 
)

// CheckDNSRoute godoc
// @Summary Check DNS configuration for a URL
// @Description Verifies that DNS is correctly configured for the given URL
// @Tags system
// @Produce json
// @Security BearerAuth
// @Param url query string true "URL to check DNS for"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/dns-check [get]
func CheckDNSRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil {
		return
	}

	if(req.Method == "GET") {
		url := utils.SanitizeSafe(req.URL.Query().Get("url"))
		url = strings.Split(url, ":")[0]
		
		if url == "" {			
			utils.Error("CheckDNS", nil)
			utils.HTTPError(w, "Internal server error: No URL requested", http.StatusInternalServerError, "DNS001")
			return
		}

		// if ends with .local
		if !strings.HasSuffix(url, ".local") {

			errDNS := utils.CheckDNS(url)

			if errDNS != nil {
				utils.Error("CheckDNS", errDNS)
				utils.HTTPError(w, "DNS Check error: " + errDNS.Error(), http.StatusInternalServerError, "DNS002")
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("CheckDNS: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}


// GetDNSRoute godoc
// @Summary Resolve DNS for a URL
// @Description Performs a DNS lookup for the given URL and returns the resolved IPv4 address
// @Tags system
// @Produce json
// @Param url query string true "URL to resolve"
// @Success 200 {object} utils.APIResponse{data=string}
// @Failure 405 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/dns [get]
func GetDNSRoute(w http.ResponseWriter, req *http.Request) {
	if !utils.GetMainConfig().NewInstall && (utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil) {
		return
	}

	if(req.Method == "GET") {
		url := utils.SanitizeSafe(req.URL.Query().Get("url"))
		url = strings.Split(url, ":")[0]
		
		if url == "" {			
			utils.Error("CheckDNS", nil)
			utils.HTTPError(w, "Internal server error: No URL requested", http.StatusInternalServerError, "DNS001")
			return
		}

		ips, err := net.LookupIP(url)
		if err != nil {
			utils.Error("CheckDNS", err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DNS001")
			return
		}

		for _, ip := range ips {
			if ip.To4() != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "OK",
					"data": ip.String(),
				})
				return
			}
		}

		utils.Error("CheckDNS: No DNS entry found. Did you point the domain to your server?", nil)
		utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DNS001")
	} else {
		utils.Error("CheckDNS: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}