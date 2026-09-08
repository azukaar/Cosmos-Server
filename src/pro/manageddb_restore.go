// Community build stub of the Cosmos Pro feature set.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// restoreManagedDB godoc
// @Summary Restore a managed database backup into a new instance
// @Description Restores one snapshot of an instance's backup into a BRAND-NEW
// @Description instance on the node serving this request (Pro feature). The source
// @Description instance is never modified. Each logical database in the snapshot is
// @Description recreated with a FRESHLY generated role password — the roles in the
// @Description backup's globals.sql carry password hashes the cluster cannot
// @Description reproduce, so that file is kept for manual recovery only.
// @Description The repository must be reachable from this node.
// @Description Partial failures are reported per database in the response body; the
// @Description new instance is kept either way.
// @Tags databases
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ManagedDBRestoreRequest true "Restore request"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/restore [post]
func restoreManagedDB(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
