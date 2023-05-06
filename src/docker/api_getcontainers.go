package docker

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
)

func GetContainerRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
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