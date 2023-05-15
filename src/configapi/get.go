package configapi

import (
	"net/http"
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils" 
)

func ConfigApiGet(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	isAdmin := utils.IsAdmin(req)

	if(req.Method == "GET") {
		config := utils.ReadConfigFromFile()

		// delete AuthPrivateKey and TLSKey
		config.HTTPConfig.AuthPrivateKey = ""
		config.HTTPConfig.TLSKey = ""

		if !isAdmin {
			config.MongoDB = "***"
			config.EmailConfig.Password = "***"
			config.EmailConfig.Username = "***"
			config.EmailConfig.Host = "***"

			// filter admin only routes
			filteredRoutes := make([]utils.ProxyRouteConfig, 0)
			for _, route := range config.HTTPConfig.ProxyConfig.Routes {
				if !route.AdminOnly {
					filteredRoutes = append(filteredRoutes, route)
				}
			}
			config.HTTPConfig.ProxyConfig.Routes = filteredRoutes
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": config,
			"updates": utils.UpdateAvailable,
		})
	} else {
		utils.Error("SettingGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
