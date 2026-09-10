package docker

import (
	"net/http"
	"strconv"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 

	"github.com/docker/docker/api/types"
	"github.com/gorilla/mux"
)

var maxLimit = 1000

// ContainerWithState adds inspect-only Health/ExitCode to the /api/servapps
// summary: Health is set when a healthcheck exists, ExitCode lets the UI
// distinguish a clean stop (exit 0) from an exited-with-error container.
type ContainerWithState struct {
	types.Container
	Health   string `json:"Health,omitempty"`
	ExitCode *int   `json:"ExitCode,omitempty"`
}

// ListContainersRoute godoc
// @Summary List all Docker containers
// @Tags docker
// @Produce json
// @Param limit query int false "Maximum number of containers to return"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/servapps [get]
func ListContainersRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
		return
	} 

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	// from, _ := req.URL.Query().Get("from")

	if limit == 0 {
		limit = maxLimit
	}
	
	if(req.Method == "GET") {
		containers, err := ListContainers()

		if err != nil {
			utils.Error("ListContainersRoute: Error while getting containers", err)
			utils.HTTPError(w, "Containers Get Error", http.StatusInternalServerError, "DL001")
			return	
		}

		// Enrich with health / exit code (only the states that need them).
		withState := make([]ContainerWithState, 0, len(containers))
		for _, c := range containers {
			entry := ContainerWithState{Container: c}
			if c.State == "running" || c.State == "exited" {
				if inspect, iErr := DockerClient.ContainerInspect(DockerContext, c.ID); iErr == nil && inspect.State != nil {
					if inspect.State.Health != nil {
						entry.Health = inspect.State.Health.Status
					}
					if c.State == "exited" {
						code := inspect.State.ExitCode
						entry.ExitCode = &code
					}
				}
			}
			withState = append(withState, entry)
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": withState,
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// ExportContainerRoute godoc
// @Summary Export a container configuration as a service definition
// @Tags docker
// @Produce json
// @Param containerId path string true "Container ID or name"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/servapps/{containerId}/export [get]
func ExportContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CREDENTIALS_READ) != nil {
		return
	}

	if req.Method == "GET" {
		vars := mux.Vars(req)
		containerID := vars["containerId"]

		errD := Connect()
		if errD != nil {
			utils.Error("exportContainer", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "EC001")
			return
		}

		service, err := ExportContainer(containerID)
		if err != nil {
			utils.Error("exportContainer: Error while exporting container", err)
			utils.HTTPError(w, "Container Export Error: "+err.Error(), http.StatusInternalServerError, "EC002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": service,
		})
	} else {
		utils.Error("exportContainer: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
