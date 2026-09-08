// Community build stub of the Cosmos Pro feature set.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// TagNodesRoute godoc
// @Summary Which nodes a tag set selects
// @Description Returns the nodes whose tags satisfy every requested tag (AND
// @Description semantics; an empty tag set matches every node), each flagged with
// @Description whether it is currently heartbeating, plus the cluster totals the
// @Description UI needs for a "matches N of M nodes" hint (Pro feature).
// @Description Read-only and answerable from any node.
// @Tags deployments
// @Produce json
// @Security BearerAuth
// @Param tags query string false "Comma-separated tags, ANDed together"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/constellation/tag-nodes [get]
func TagNodesRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
