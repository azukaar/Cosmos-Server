package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
)

// GetContainerRoute godoc
// @Summary Inspect a single Docker container by ID
// @Tags docker
// @Produce json
// @Param containerId path string true "Container ID or name"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/servapps/{containerId}/ [get]
func GetContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
		return
	}

	vars := mux.Vars(req)
	containerId := vars["containerId"]


	if req.Method == "GET" {
		errD := Connect()
		if errD != nil {
			utils.Error("GetContainerRoute", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "LN001")
			return
		}
		
		// get Docker container
		container, err := DockerClient.ContainerInspect(context.Background(), containerId)
		if err != nil {
			utils.Error("GetContainerRoute: Error while getting container", err)
			utils.HTTPError(w, "Container Get Error: " + err.Error(), http.StatusInternalServerError, "LN002")
			return
		}

		// Mask env var values if user lacks PERM_CREDENTIALS_READ
		if !utils.HasPermission(req, utils.PERM_CREDENTIALS_READ) && container.Config != nil {
			masked := make([]string, len(container.Config.Env))
			for i, env := range container.Config.Env {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					masked[i] = parts[0] + "=***"
				} else {
					masked[i] = env
				}
			}
			container.Config.Env = masked
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   container,
		})
	} else {
		utils.Error("GetContainerRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}