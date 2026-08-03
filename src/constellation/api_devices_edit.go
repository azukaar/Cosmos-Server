package constellation

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

type DeviceEditRequestJSON struct {
	IsLighthouse   bool     `json:"isLighthouse"`
	IsRelay        bool     `json:"isRelay"`
	IsExitNode     bool     `json:"isExitNode"`
	IsLoadBalancer bool     `json:"isLoadBalancer"`
	Tags           []string `json:"tags" validate:"omitempty,dive,min=1,max=64"`
}

// DeviceEdit_API godoc
// @Summary Edit the current Constellation device properties
// @Tags constellation
// @Accept json
// @Produce json
// @Param body body DeviceEditRequestJSON true "Device edit payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/constellation/edit-device [post]
func DeviceEdit_API(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		utils.Error("DeviceEdit: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES) != nil {
		return
	}

	var request DeviceEditRequestJSON
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		utils.Error("DeviceEdit: Invalid request", err)
		utils.HTTPError(w, "Device Edit Error", http.StatusInternalServerError, "DE001")
		return
	}

	deviceName, err := GetCurrentDeviceName()
	if err != nil {
		utils.Error("DeviceEdit: Error getting current device name", err)
		utils.HTTPError(w, "Device Edit Error: "+err.Error(), http.StatusInternalServerError, "DE002")
		return
	}

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	defer closeDb()

	if errCo != nil {
		utils.Error("Database Connect", errCo)
		utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
		return
	}

	// Non-lighthouses cannot be relay, exit, or load balancer
	if !request.IsLighthouse {
		request.IsRelay = false
		request.IsExitNode = false
		request.IsLoadBalancer = false
	}

	// Normalize tags: trim whitespace, drop empties, dedupe while preserving
	// input order. Keeps the persisted list tidy regardless of how the UI
	// submitted them.
	cleanTags := make([]string, 0, len(request.Tags))
	seenTag := map[string]struct{}{}
	for _, raw := range request.Tags {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		// Reject commas: the UI renders/parses tags as a comma-separated
		// list, so an embedded comma would not round-trip.
		if strings.ContainsAny(t, ",") {
			utils.Error("DeviceEdit: tag contains comma", nil)
			utils.HTTPError(w, "Tag cannot contain commas", http.StatusBadRequest, "DE004")
			return
		}
		if _, dup := seenTag[t]; dup {
			continue
		}
		seenTag[t] = struct{}{}
		cleanTags = append(cleanTags, t)
	}
	request.Tags = cleanTags

	if errV := utils.Validate.Struct(request); errV != nil {
		utils.Error("DeviceEdit: Validation error", errV)
		utils.HTTPError(w, "Device Edit Validation Error: "+errV.Error(), http.StatusBadRequest, "DE005")
		return
	}

	_, err = c.UpdateOne(nil, map[string]interface{}{
		"DeviceName": deviceName,
		"Blocked":    false,
	}, map[string]interface{}{
		"$set": map[string]interface{}{
			"IsLighthouse":   request.IsLighthouse,
			"IsRelay":        request.IsRelay,
			"IsExitNode":     request.IsExitNode,
			"IsLoadBalancer": request.IsLoadBalancer,
			"Tags":           cleanTags,
		},
	})

	if err != nil {
		utils.Error("DeviceEdit: Error updating device", err)
		utils.HTTPError(w, "Device Edit Error: "+err.Error(), http.StatusInternalServerError, "DE003")
		return
	}

	utils.TriggerEvent(
		"cosmos.constellation.device.edit",
		"Device edited",
		"success",
		"",
		map[string]interface{}{
			"deviceName": deviceName,
		})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})

	go func() {
		go SendNewDBSyncMessage()
		time.Sleep(2 * time.Second)
		RestartNebula()
	}()
}
