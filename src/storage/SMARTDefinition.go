package storage 

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/analogj/scrutiny/webapp/backend/pkg/thresholds"
)

// ListSmartDef godoc
// @Summary Get SMART attribute definitions for ATA and NVMe drives
// @Tags Storage
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/smart-def [get]
func ListSmartDef(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_RESOURCES_READ) != nil {
		return
	}

	if req.Method == "GET" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": map[string]interface{}{
				"ATA": thresholds.AtaMetadata,
				"NVME": thresholds.NmveMetadata,
			},
		})
	} else {
		utils.Error("ListDisksRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}