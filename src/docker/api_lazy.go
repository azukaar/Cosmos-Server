package docker

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"

	containerType "github.com/docker/docker/api/types/container"
	"github.com/gorilla/mux"
)

// LazyContainerRoute godoc
// @Summary Enable or disable lazy start (stop when idle) for a Docker container
// @Tags docker
// @Produce json
// @Param containerId path string true "Container ID or name"
// @Param status path string true "Lazy status (true or false)"
// @Param idle query string false "Idle duration before the container is stopped (Go duration, ex 1h)"
// @Param startTimeout query string false "Readiness timeout after a wake (Go duration, ex 60s)"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/servapps/{containerId}/lazy/{status} [get]
func LazyContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	vars := mux.Vars(req)
	containerName := utils.SanitizeSafe(vars["containerId"])
	status := utils.Sanitize(vars["status"])

	if utils.IsInsideContainer && containerName == os.Getenv("HOSTNAME") {
		utils.Error("LazyContainer - Cosmos cannot be made lazy", nil)
		utils.HTTPError(w, "Cosmos cannot be made lazy", http.StatusBadRequest, "DS001")
		return
	}

	// the wake path dials the container's network, unreachable from a bridge-networked Cosmos
	if status == "true" && utils.IsInsideContainer && !utils.IsHostNetwork {
		utils.Error("LazyContainer - lazy containers require Cosmos on the host network", nil)
		utils.HTTPError(w, "Lazy containers are not supported when Cosmos runs in a container without host networking", http.StatusBadRequest, "DS001")
		return
	}

	if req.Method == "GET" {
		idle := utils.Sanitize(req.URL.Query().Get("idle"))
		startTimeout := utils.Sanitize(req.URL.Query().Get("startTimeout"))

		if idle != "" {
			if _, err := time.ParseDuration(idle); err != nil {
				utils.Error("LazyContainer: invalid idle duration", err)
				utils.HTTPError(w, "Invalid idle duration: "+err.Error(), http.StatusBadRequest, "DS001")
				return
			}
		}

		if startTimeout != "" {
			if _, err := time.ParseDuration(startTimeout); err != nil {
				utils.Error("LazyContainer: invalid startTimeout duration", err)
				utils.HTTPError(w, "Invalid startTimeout duration: "+err.Error(), http.StatusBadRequest, "DS001")
				return
			}
		}

		container, err := DockerClient.ContainerInspect(DockerContext, containerName)
		if err != nil {
			utils.Error("LazyContainer Inspect", err)
			utils.HTTPError(w, "Internal server error: "+err.Error(), http.StatusInternalServerError, "DS002")
			return
		}

		if status == "true" {
			labels := map[string]string{
				LazyLabel: "true",
			}
			if idle != "" {
				labels[LazyIdleLabel] = idle
			}
			if startTimeout != "" {
				labels[LazyStartTimeoutLabel] = startTimeout
			}

			AddLabels(container, labels)

			// docker must not auto-restart a container Cosmos put to sleep
			container.HostConfig.RestartPolicy = containerType.RestartPolicy{
				Name: containerType.RestartPolicyMode("no"),
			}
		} else {
			RemoveLabels(container, []string{
				LazyLabel,
				LazyIdleLabel,
				LazyStartTimeoutLabel,
			})
		}

		utils.Log("API: Set Lazy " + status + " : " + containerName)

		_, errEdit := EditContainer(container.ID, container, false)
		if errEdit != nil {
			utils.Error("LazyContainer Edit", errEdit)
			utils.HTTPError(w, "Internal server error: "+errEdit.Error(), http.StatusInternalServerError, "DS003")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("LazyContainer: Method not allowed"+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
