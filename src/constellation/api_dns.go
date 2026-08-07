package constellation

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
)

func DNSEntriesRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		listDNSEntries(w, req)
	} else if req.Method == "POST" {
		createDNSEntry(w, req)
	} else {
		utils.Error("DNSEntriesRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func DNSEntriesIdRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		getDNSEntry(w, req)
	} else if req.Method == "PUT" {
		updateDNSEntry(w, req)
	} else if req.Method == "DELETE" {
		deleteDNSEntry(w, req)
	} else {
		utils.Error("DNSEntriesIdRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

// listDNSEntries godoc
// @Summary List all custom Constellation DNS entries
// @Tags constellation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/constellation/dns [get]
func listDNSEntries(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	config := utils.ReadConfigFromFile()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   config.ConstellationConfig.CustomDNSEntries,
	})
}

// getDNSEntry godoc
// @Summary Get a single Constellation DNS entry by key
// @Tags constellation
// @Produce json
// @Param key path string true "DNS entry key"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/dns/{key} [get]
func getDNSEntry(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION_READ) != nil {
		return
	}

	key := mux.Vars(req)["key"]
	config := utils.ReadConfigFromFile()

	for _, entry := range config.ConstellationConfig.CustomDNSEntries {
		if entry.Key == key {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data":   entry,
			})
			return
		}
	}

	utils.HTTPError(w, "DNS entry not found", http.StatusNotFound, "DNS001")
}

// createDNSEntry godoc
// @Summary Create a new Constellation DNS entry
// @Tags constellation
// @Accept json
// @Produce json
// @Param body body utils.ConstellationDNSEntry true "DNS entry to create"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/constellation/dns [post]
func createDNSEntry(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	var newEntry utils.ConstellationDNSEntry
	err := json.NewDecoder(req.Body).Decode(&newEntry)
	if err != nil {
		utils.Error("CreateDNSEntry: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "DNS002")
		return
	}

	if newEntry.Key == "" {
		utils.HTTPError(w, "DNS entry key is required", http.StatusBadRequest, "DNS003")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	entries := config.ConstellationConfig.CustomDNSEntries

	for _, entry := range entries {
		if entry.Key == newEntry.Key {
			utils.HTTPError(w, "DNS entry with this key already exists", http.StatusConflict, "DNS004")
			return
		}
	}

	config.ConstellationConfig.CustomDNSEntries = append(config.ConstellationConfig.CustomDNSEntries, newEntry)
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.settings",
		"DNS entry created: "+newEntry.Key,
		"success",
		"",
		map[string]interface{}{
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   newEntry,
	})
}

// updateDNSEntry godoc
// @Summary Update an existing Constellation DNS entry
// @Tags constellation
// @Accept json
// @Produce json
// @Param key path string true "DNS entry key"
// @Param body body utils.ConstellationDNSEntry true "Updated DNS entry"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/dns/{key} [put]
func updateDNSEntry(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	key := mux.Vars(req)["key"]

	var updatedEntry utils.ConstellationDNSEntry
	err := json.NewDecoder(req.Body).Decode(&updatedEntry)
	if err != nil {
		utils.Error("UpdateDNSEntry: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "DNS005")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	entries := config.ConstellationConfig.CustomDNSEntries
	entryIndex := -1

	for i, entry := range entries {
		if entry.Key == key {
			entryIndex = i
			break
		}
	}

	if entryIndex == -1 {
		utils.HTTPError(w, "DNS entry not found", http.StatusNotFound, "DNS006")
		return
	}

	entries[entryIndex] = updatedEntry
	config.ConstellationConfig.CustomDNSEntries = entries
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.settings",
		"DNS entry updated: "+key,
		"success",
		"",
		map[string]interface{}{
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   updatedEntry,
	})
}

// deleteDNSEntry godoc
// @Summary Delete a Constellation DNS entry by key
// @Tags constellation
// @Produce json
// @Param key path string true "DNS entry key"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/dns/{key} [delete]
func deleteDNSEntry(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	key := mux.Vars(req)["key"]

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()
	entries := config.ConstellationConfig.CustomDNSEntries
	entryIndex := -1

	for i, entry := range entries {
		if entry.Key == key {
			entryIndex = i
			break
		}
	}

	if entryIndex == -1 {
		utils.HTTPError(w, "DNS entry not found", http.StatusNotFound, "DNS007")
		return
	}

	entries = append(entries[:entryIndex], entries[entryIndex+1:]...)
	config.ConstellationConfig.CustomDNSEntries = entries
	utils.SetBaseMainConfig(config)

	utils.TriggerEvent(
		"cosmos.settings",
		"DNS entry deleted: "+key,
		"success",
		"",
		map[string]interface{}{
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}
