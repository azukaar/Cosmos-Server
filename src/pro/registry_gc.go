// Community build stub of the Cosmos Pro feature set: types exist so the shared
// code and the SDK generator compile; handlers answer PRO001 and hooks do nothing.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// StartRegistryGC starts the reconciler that keeps this node's registry GC jobs
// in sync with leadership.
func StartRegistryGC() {
	// Pro feature stub.
}

// RegistryGCRoute godoc
// @Summary Run a registry's garbage collection now
// @Description Starts a mark-and-sweep of the registry's blob store and reconciles its
// @Description stored-size and package counters. Returns 202 immediately; the outcome
// @Description arrives as a cosmos.registry.gc event. Blobs younger than 24 hours are
// @Description never collected, whether or not anything references them (Pro feature).
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Success 202 {object} utils.APIResponse
// @Router /api/constellation/registries/{name}/gc [post]
func RegistryGCRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
