package docker

import (
	"net/http"
	"strconv"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils" 

	"github.com/gorilla/mux"
)

var maxLimit = 1000

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
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": containers,
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

		// from=initial returns the config the user actually set (container vs
		// image-config diff). from=runtime returns the full live runtime state.
		from := req.URL.Query().Get("from")
		if from == "" {
			from = "initial"
		}

		var service ContainerCreateRequestContainer
		var err error
		if from == "runtime" {
			service, err = ExportContainer(containerID)
		} else {
			service, err = ExportContainerRuntime(containerID)
		}
		if err != nil {
			utils.Error("exportContainer: Error while exporting container", err)
			utils.HTTPError(w, "Container Export Error: "+err.Error(), http.StatusInternalServerError, "EC002")
			return
		}

		// Collect HJSON comments stored as cosmos.compose.<path> labels.
		comments := map[string]string{}
		if ci, ciErr := DockerClient.ContainerInspect(DockerContext, containerID); ciErr == nil {
			if c := getComposeComments(ci.Config); c != nil {
				comments = c
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": service,
			"comments": comments,
		})
	} else {
		utils.Error("exportContainer: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
