package configapi

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/utils"
)

type DNSConfigRequest struct {
	DNSPort                 string                        `json:"dnsPort,omitempty"`
	DNSFallback             string                        `json:"dnsFallback,omitempty"`
	DNSBlockBlacklist       *bool                         `json:"dnsBlockBlacklist,omitempty"`
	DNSAdditionalBlocklists []string                      `json:"dnsAdditionalBlocklists,omitempty"`
	CustomDNSEntries        []utils.ConstellationDNSEntry `json:"customDNSEntries,omitempty"`
}

// ConfigApiDNS godoc
// @Summary Update DNS configuration
// @Description Patches DNS-related settings including port, fallback, blocklists, and custom entries
// @Tags config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DNSConfigRequest true "DNS configuration fields to update"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/config/dns [patch]
func ConfigApiDNS(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	}

	if req.Method != "PATCH" {
		utils.Error("DNSConfigUpdate: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "DNS001")
		return
	}

	utils.Log("DNSConfigUpdate: Updating DNS configuration")

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	var updateReq DNSConfigRequest
	err := json.NewDecoder(req.Body).Decode(&updateReq)
	if err != nil {
		utils.Error("DNSConfigUpdate: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "DNS002")
		return
	}

	config := utils.ReadConfigFromFile()

	if updateReq.DNSPort != "" {
		config.ConstellationConfig.DNSPort = updateReq.DNSPort
	}
	if updateReq.DNSFallback != "" {
		config.ConstellationConfig.DNSFallback = updateReq.DNSFallback
	}
	if updateReq.DNSBlockBlacklist != nil {
		config.ConstellationConfig.DNSBlockBlacklist = *updateReq.DNSBlockBlacklist
	}
	if updateReq.DNSAdditionalBlocklists != nil {
		config.ConstellationConfig.DNSAdditionalBlocklists = updateReq.DNSAdditionalBlocklists
	}
	if updateReq.CustomDNSEntries != nil {
		config.ConstellationConfig.CustomDNSEntries = updateReq.CustomDNSEntries
	}

	utils.SetBaseMainConfig(config)

	err = utils.TouchDatabase()
	if err != nil {
		utils.Error("DNSConfigUpdate: Failed to touch database", err)
	}

	utils.TriggerEvent(
		"cosmos.settings",
		"DNS settings updated",
		"success",
		"",
		map[string]interface{}{},
	)

	go func() {
		constellation.SendNewDBSyncMessage()
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}
