package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/docker/docker/api/types"
	network "github.com/docker/docker/api/types/network"
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

		// if network is connected to a single container, detach it
		network, err := DockerClient.NetworkInspect(context.Background(), networkID, types.NetworkInspectOptions{})
		if err != nil {
			utils.Error("DeleteNetworkRoute: Error while getting network", err)
			utils.HTTPError(w, "Network Get Error: "+err.Error(), http.StatusInternalServerError, "DN002")
			return
		}

		if len(network.Containers) == 1 {
			for containerID := range network.Containers {
				err = DockerClient.NetworkDisconnect(context.Background(), networkID, containerID, true)
				if err != nil {
					utils.Error("DeleteNetworkRoute: Error while detaching network", err)
					utils.HTTPError(w, "Network Detach Error: "+err.Error(), http.StatusInternalServerError, "DN002")
					return
				}
			}
		}

		// Delete the specified Docker network
		err = DockerClient.NetworkRemove(context.Background(), networkID)
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

func NetworkContainerRoutes(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		ListContainerNetworks(w, req)
	} else if req.Method == "DELETE" {
		DetachNetwork(w, req)
		} else if req.Method == "POST" {
		AttachNetwork(w, req)
	} else {
		utils.Error("NetworkContainerRoutes: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func NetworkRoutes(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		ListNetworksRoute(w, req)
	} else if req.Method == "POST" {
		CreateNetworkRoute(w, req)
	} else {
		utils.Error("NetworkContainerRoutes: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func AttachNetwork(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		vars := mux.Vars(req)
		containerID := vars["containerId"]
		networkID := vars["networkId"]

		errD := Connect()
		if errD != nil {
			utils.Error("AttachNetwork", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "AN001")
			return
		}

		err := DockerClient.NetworkConnect(context.Background(), networkID, containerID, nil)
		if err != nil {
			utils.Error("AttachNetwork: Error while attaching network", err)
			utils.HTTPError(w, "Network Attach Error: "+err.Error(), http.StatusInternalServerError, "AN002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": "Network attached successfully",
		})
	} else {
		utils.Error("AttachNetwork: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func DetachNetwork(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "DELETE" {
		vars := mux.Vars(req)
		containerID := vars["containerId"]
		networkID := vars["networkId"]

		errD := Connect()
		if errD != nil {
			utils.Error("DetachNetwork", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "DN001")
			return
		}

		if os.Getenv("HOSTNAME") != "" && networkID == "bridge" && containerID == os.Getenv("HOSTNAME") {
			utils.Error("DetachNetwork - Cannot disconnect self from bridge", nil)
			utils.HTTPError(w, "Cannot disconnect self from bridge", http.StatusBadRequest, "DS003")
			return
		}

		err := DockerClient.NetworkDisconnect(context.Background(), networkID, containerID, true)
		if err != nil {
			utils.Error("DetachNetwork: Error while detaching network", err)
			utils.HTTPError(w, "Network Detach Error: "+err.Error(), http.StatusInternalServerError, "DN002")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"message": "Network detached successfully",
		})
	} else {
		utils.Error("DetachNetwork: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func ListContainerNetworks(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		vars := mux.Vars(req)
		containerID := vars["containerId"]

		errD := Connect()
		if errD != nil {
			utils.Error("ListContainerNetworks", errD)
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

		container, err := DockerClient.ContainerInspect(context.Background(), containerID)
		if err != nil {
			utils.Error("ListContainerNetworks: Error while getting container", err)
			utils.HTTPError(w, "Container Get Error: "+err.Error(), http.StatusInternalServerError, "LN002")
			return
		}
		containerNetworks := container.NetworkSettings.Networks

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
				"networks": networks,
				"containerNetworks": containerNetworks,
			},
		})
	} else {
		utils.Error("ListContainerNetworks: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

type createNetworkPayload struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
	AttachCosmos bool `json:"attachCosmos"`
	ParentInterface string `json:"parentInterface"`
	Subnet string `json:"subnet"`
}

func CreateNetworkRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "POST" {
		errD := Connect()
		if errD != nil {
			utils.Error("CreateNetworkRoute", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "CN001")
			return
		}

		var payload createNetworkPayload
		err := json.NewDecoder(req.Body).Decode(&payload)
		if err != nil {
			utils.Error("CreateNetworkRoute: Error reading request body", err)
			utils.HTTPError(w, "Error reading request body: "+err.Error(), http.StatusBadRequest, "CN002")
			return
		}

		networkCreate := types.NetworkCreate{
			CheckDuplicate: true,
			Driver:         payload.Driver,
			Options: map[string]string{
				"parent": payload.ParentInterface,
			},
		}

		if payload.Subnet != "" {
			networkCreate.IPAM = &network.IPAM{
				Driver: "default",
				Config: []network.IPAMConfig{
					{
						Subnet: payload.Subnet,
					},
				},
			}
		}

		resp, err := CreateReasonableNetwork(payload.Name, networkCreate)
		if err != nil {
			utils.Error("CreateNetworkRoute: Error while creating network", err)
			utils.HTTPError(w, "Network Create Error: " + err.Error(), http.StatusInternalServerError, "CN004")
			return
		}

		if payload.AttachCosmos {
			// Attach network to cosmos
			err = AttachNetworkToCosmos(resp.ID)
			if err != nil {
				utils.Error("CreateNetworkRoute: Error while attaching network to cosmos", err)
				utils.HTTPError(w, "Network Attach Error: " + err.Error(), http.StatusInternalServerError, "CN005")
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   resp,
		})
	} else {
		utils.Error("CreateNetworkRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}