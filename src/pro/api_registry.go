// Community build stub of the Cosmos Pro feature set; handlers answer PRO001.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// RegistryCreateRequest is the POST body.
type RegistryCreateRequest struct {
	Name string `json:"name"`
	// Type is docker, npm, static or generic. Required and immutable afterwards.
	Type       string          `json:"type"`
	QuotaBytes int64           `json:"quotaBytes,omitempty"`
	Storage    RegistryStorage `json:"storage"`
}

// RegistrySettingsRequest is the PUT body; the pointer keeps an absent field's
// stored value. The type is immutable and deliberately absent.
type RegistrySettingsRequest struct {
	QuotaBytes *int64 `json:"quotaBytes,omitempty"`
}

func RegistryRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

func RegistryIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// listRegistries godoc
// @Summary List package registries
// @Description Returns every registry with the accesses publishing it and its stored-size rollup, secrets redacted (Pro feature)
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries [get]
func listRegistries(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// getRegistry godoc
// @Summary Get one package registry
// @Description Returns the registry with the accesses publishing it and its stored-size rollup, secrets redacted (Pro feature)
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries/{name} [get]
func getRegistry(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// createRegistry godoc
// @Summary Create a package registry
// @Description Claims the name, provisions the backing bucket and marks the registry ready.
// @Description A registry is typed storage (docker/npm/static/generic): publish it by creating
// @Description an access (Pro feature).
// @Tags registry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body RegistryCreateRequest true "Registry to create"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries [post]
func createRegistry(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// deleteRegistry godoc
// @Summary Delete a package registry
// @Description Removes the record and every metadata key. Refused while an access
// @Description still publishes it. Stored blobs are PRESERVED unless purgeData=true,
// @Description which best-effort empties the backing bucket (the bucket itself is
// @Description left in place) (Pro feature).
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Param purgeData query boolean false "Also delete every stored blob"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries/{name} [delete]
func deleteRegistry(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// RegistrySettingsRoute godoc
// @Summary Update a registry's storage settings
// @Description Replaces the quota. Absent fields keep their stored value (Pro feature).
// @Tags registry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Param body body RegistrySettingsRequest true "Settings to change"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries/{name}/settings [put]
func RegistrySettingsRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
