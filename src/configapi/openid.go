package configapi

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
)

func OpenIDRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		listOpenIDClients(w, req)
	} else if req.Method == "POST" {
		createOpenIDClient(w, req)
	} else {
		utils.Error("OpenIDRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func OpenIDIdRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		getOpenIDClient(w, req)
	} else if req.Method == "PUT" {
		updateOpenIDClient(w, req)
	} else if req.Method == "DELETE" {
		deleteOpenIDClient(w, req)
	} else {
		utils.Error("OpenIDIdRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

// listOpenIDClients godoc
// @Summary List all OpenID clients
// @Description Returns all configured OpenID Connect client configurations
// @Tags openid
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]utils.OpenIDClient}
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/openid [get]
func listOpenIDClients(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	config := utils.ReadConfigFromFile()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   config.OpenIDClients,
	})
}

// getOpenIDClient godoc
// @Summary Get an OpenID client by ID
// @Description Returns a single OpenID Connect client configuration
// @Tags openid
// @Produce json
// @Security BearerAuth
// @Param id path string true "OpenID client ID"
// @Success 200 {object} utils.APIResponse{data=utils.OpenIDClient}
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/openid/{id} [get]
func getOpenIDClient(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	id := mux.Vars(req)["id"]
	config := utils.ReadConfigFromFile()

	for _, client := range config.OpenIDClients {
		if client.ID == id {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data":   client,
			})
			return
		}
	}

	utils.HTTPError(w, "OpenID client not found", http.StatusNotFound, "OID001")
}

// createOpenIDClient godoc
// @Summary Create an OpenID client
// @Description Creates a new OpenID Connect client configuration
// @Tags openid
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body utils.OpenIDClient true "OpenID client configuration"
// @Success 200 {object} utils.APIResponse{data=utils.OpenIDClient}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/openid [post]
func createOpenIDClient(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	var newClient utils.OpenIDClient
	err := json.NewDecoder(req.Body).Decode(&newClient)
	if err != nil {
		utils.Error("CreateOpenIDClient: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "OID002")
		return
	}

	if newClient.ID == "" {
		utils.HTTPError(w, "OpenID client ID is required", http.StatusBadRequest, "OID003")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	clients := config.OpenIDClients

	for _, client := range clients {
		if client.ID == newClient.ID {
			utils.HTTPError(w, "OpenID client with this ID already exists", http.StatusConflict, "OID004")
			return
		}
	}

	config.OpenIDClients = append(config.OpenIDClients, newClient)
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.settings",
		"OpenID client created: "+newClient.ID,
		"success",
		"",
		map[string]interface{}{
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   newClient,
	})
}

// updateOpenIDClient godoc
// @Summary Update an OpenID client
// @Description Replaces an existing OpenID Connect client configuration
// @Tags openid
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "OpenID client ID"
// @Param request body utils.OpenIDClient true "Updated OpenID client configuration"
// @Success 200 {object} utils.APIResponse{data=utils.OpenIDClient}
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/openid/{id} [put]
func updateOpenIDClient(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	id := mux.Vars(req)["id"]

	var updatedClient utils.OpenIDClient
	err := json.NewDecoder(req.Body).Decode(&updatedClient)
	if err != nil {
		utils.Error("UpdateOpenIDClient: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "OID005")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	clients := config.OpenIDClients
	clientIndex := -1

	for i, client := range clients {
		if client.ID == id {
			clientIndex = i
			break
		}
	}

	if clientIndex == -1 {
		utils.HTTPError(w, "OpenID client not found", http.StatusNotFound, "OID006")
		return
	}

	clients[clientIndex] = updatedClient
	config.OpenIDClients = clients
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.settings",
		"OpenID client updated: "+id,
		"success",
		"",
		map[string]interface{}{
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   updatedClient,
	})
}

// deleteOpenIDClient godoc
// @Summary Delete an OpenID client
// @Description Removes an OpenID Connect client configuration by ID
// @Tags openid
// @Produce json
// @Security BearerAuth
// @Param id path string true "OpenID client ID"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/openid/{id} [delete]
func deleteOpenIDClient(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	id := mux.Vars(req)["id"]

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	clients := config.OpenIDClients
	clientIndex := -1

	for i, client := range clients {
		if client.ID == id {
			clientIndex = i
			break
		}
	}

	if clientIndex == -1 {
		utils.HTTPError(w, "OpenID client not found", http.StatusNotFound, "OID007")
		return
	}

	clients = append(clients[:clientIndex], clients[clientIndex+1:]...)
	config.OpenIDClients = clients
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.settings",
		"OpenID client deleted: "+id,
		"success",
		"",
		map[string]interface{}{
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}
