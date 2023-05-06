package docker

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/docker/docker/api/types"
)

func ListNetworksRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		errD := Connect()
		if errD != nil {
			utils.Error("ListNetworksRoute", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "LN001")
			return
		}

		// List Docker networks
		networks, err := DockerClient.NetworkList(context.Background(), types.NetworkListOptions{})
		if err != nil {
			utils.Error("ListNetworksRoute: Error while getting networks", err)
			utils.HTTPError(w, "Networks Get Error: " + err.Error(), http.StatusInternalServerError, "LN002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   networks,
		})
	} else {
		utils.Error("ListNetworksRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}


func DeleteNetworkRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "DELETE" {
		// Get the network ID from URL
		vars := mux.Vars(req)
		networkID := vars["networkID"]

		errD := Connect()
		if errD != nil {
			utils.Error("DeleteNetworkRoute", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "DN001")
			return
		}

		// Delete the specified Docker network
		err := DockerClient.NetworkRemove(context.Background(), networkID)
		if err != nil {
			utils.Error("DeleteNetworkRoute: Error while deleting network", err)
			utils.HTTPError(w, "Network Deletion Error: " + err.Error(), http.StatusInternalServerError, "DN002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": "Network deleted successfully",
		})
	} else {
		utils.Error("DeleteNetworkRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}