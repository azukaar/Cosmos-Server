package main

import (
	"net/http"
	"encoding/json"
	"net"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils" 
)

func CheckDNSRoute(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
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

		errDNS := utils.CheckDNS(url)

		if errDNS != nil {
			utils.Error("CheckDNS", errDNS)
			utils.HTTPError(w, "DNS Check error: " + errDNS.Error(), http.StatusInternalServerError, "DNS002")
			return
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


func GetDNSRoute(w http.ResponseWriter, req *http.Request) {
	if !utils.GetMainConfig().NewInstall && (utils.LoggedInOnly(w, req) != nil) {
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