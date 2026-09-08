package constellation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
)

type DeviceEditRequestJSON struct {
	// Empty means "this node" (full edit); set means a remote tags-only edit.
	DeviceName     string   `json:"deviceName" validate:"omitempty,min=3,max=32"`
	IsLighthouse   bool     `json:"isLighthouse"`
	IsRelay        bool     `json:"isRelay"`
	IsExitNode     bool     `json:"isExitNode"`
	IsLoadBalancer bool     `json:"isLoadBalancer"`
	Tags           []string `json:"tags" validate:"omitempty,dive,min=1,max=64"`
}

var errTagComma = errors.New("tag cannot contain commas")

func normalizeDeviceTags(raw []string) ([]string, error) {
	clean := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, r := range raw {
		t := strings.TrimSpace(r)
		if t == "" {
			continue
		}
		// commas would not round-trip through the UI's comma-separated tag list
		if strings.ContainsAny(t, ",") {
			return nil, errTagComma
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		clean = append(clean, t)
	}
	return clean, nil
}

// deviceEditFields is the set of columns an edit writes. A remote edit is
// tags-ONLY: the role flags are baked into nebula.yml, so writing a stale or
// defaulted value would restart the mesh and could silently demote a
// lighthouse, relay or load balancer.
func deviceEditFields(request DeviceEditRequestJSON, tags []string, remote bool) map[string]interface{} {
	if remote {
		return map[string]interface{}{
			"Tags": tags,
		}
	}

	// Non-lighthouses cannot be relay, exit, or load balancer
	isRelay, isExitNode, isLoadBalancer := request.IsRelay, request.IsExitNode, request.IsLoadBalancer
	if !request.IsLighthouse {
		isRelay, isExitNode, isLoadBalancer = false, false, false
	}

	return map[string]interface{}{
		"IsLighthouse":   request.IsLighthouse,
		"IsRelay":        isRelay,
		"IsExitNode":     isExitNode,
		"IsLoadBalancer": isLoadBalancer,
		"Tags":           tags,
	}
}

// DeviceEdit_API godoc
// @Summary Edit Constellation device properties (this device, or another device's tags)
// @Tags constellation
// @Accept json
// @Produce json
// @Param body body DeviceEditRequestJSON true "Device edit payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
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

	cleanTags, errT := normalizeDeviceTags(request.Tags)
	if errT != nil {
		utils.Error("DeviceEdit: "+errT.Error(), nil)
		utils.HTTPError(w, "Tag cannot contain commas", http.StatusBadRequest, "DE004")
		return
	}
	request.Tags = cleanTags

	if errV := utils.Validate.Struct(request); errV != nil {
		utils.Error("DeviceEdit: Validation error", errV)
		utils.HTTPError(w, "Device Edit Validation Error: "+errV.Error(), http.StatusBadRequest, "DE005")
		return
	}

	// A device name targets another node (tags-only edit); empty means this node.
	remote := utils.Sanitize(request.DeviceName)
	deviceName := remote

	if remote != "" {
		// Resolve first so an unknown or blocked name is a 404, not a silent no-op.
		device, errD := utils.GetDeviceByName(remote, true)
		if errD != nil {
			utils.Error("DeviceEdit: target device not found: "+remote, errD)
			utils.HTTPError(w, "Device Edit Error: device not found", http.StatusNotFound, "DE006")
			return
		}
		deviceName = device.DeviceName
	} else {
		deviceName, err = GetCurrentDeviceName()
		if err != nil {
			utils.Error("DeviceEdit: Error getting current device name", err)
			utils.HTTPError(w, "Device Edit Error: "+err.Error(), http.StatusInternalServerError, "DE002")
			return
		}
	}

	// restart-or-refresh is no longer decided here: the apply loop diffs the op
	// against the row's pre-image and reacts identically on every node
	err = utils.UpdateDevices(map[string]interface{}{
		"DeviceName": deviceName,
		"Blocked":    false,
	}, deviceEditFields(request, cleanTags, remote != ""))

	if err != nil {
		utils.Error("DeviceEdit: Error updating device", err)
		utils.HTTPStoreError(w, err, "DE003")
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
}
