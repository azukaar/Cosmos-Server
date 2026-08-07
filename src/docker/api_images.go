package docker

import (
	"net/http"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 
	
)

// InspectImageRoute godoc
// @Summary Inspect a Docker image by name
// @Tags docker
// @Produce json
// @Param imageName query string true "Name of the Docker image to inspect"
// @Security BearerAuth
// @Success 200 {object} types.ImageInspect
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/images [get]
func InspectImageRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
		return
	}

	errD := Connect()
	if errD != nil {
		utils.Error("InspectImage", errD)
		utils.HTTPError(w, "Internal server error: " + errD.Error(), http.StatusInternalServerError, "DS002")
		return
	}

	imageName := utils.SanitizeSafe(req.URL.Query().Get("imageName"))
	
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