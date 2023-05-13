package docker

import (
	"net/http"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 
	
	"github.com/gorilla/mux"
)

func InspectImageRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	errD := Connect()
	if errD != nil {
		utils.Error("InspectImage", errD)
		utils.HTTPError(w, "Internal server error: " + errD.Error(), http.StatusInternalServerError, "DS002")
		return
	}

	vars := mux.Vars(req)
	imageName := utils.SanitizeSafe(vars["imageName"])
	
	utils.Log("InspectImage " + imageName)

	if req.Method == "GET" {
		image, _, err := DockerClient.ImageInspectWithRaw(DockerContext, imageName)
		if err != nil {
			utils.Error("InspectImage", err)
			utils.HTTPError(w, "Internal server error: " + err.Error(), http.StatusInternalServerError, "DS002")
			return
		}

		json.NewEncoder(w).Encode(image)
	} else {
		utils.Error("InspectImage: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}