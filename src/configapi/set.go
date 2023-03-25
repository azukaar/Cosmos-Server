package configapi

import (
	"net/http"
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils" 
)

func ConfigApiSet(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	} 

	if(req.Method == "PUT") {
		var request utils.Config
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("SettingsUpdate: Invalid User Request", err1)
			utils.HTTPError(w, "User Creation Error", 
				http.StatusInternalServerError, "UC001")
			return 
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("SettingsUpdate: Invalid User Request", errV)
			utils.HTTPError(w, "User Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "UC003")
			return 
		}

		// restore AuthPrivateKey and TLSKey
		config := utils.GetBaseMainConfig()
		request.HTTPConfig.AuthPrivateKey = config.HTTPConfig.AuthPrivateKey
		request.HTTPConfig.TLSKey = config.HTTPConfig.TLSKey

		utils.SaveConfigTofile(request)

		// if err != nil {
		// 	utils.Error("SettingsUpdate: Error saving config to file", err)
		// 	utils.HTTPError(w, "Error saving config to file",
		// 		http.StatusInternalServerError, "CS001")
		// 	return
		// }

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingsUpdate: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
