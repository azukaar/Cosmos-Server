// Community build stub of the Cosmos Pro feature set; handlers answer PRO001.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// RegistryAccessCreateRequest is the POST body.
type RegistryAccessCreateRequest struct {
	Name               string   `json:"name"`
	Host               string   `json:"host"`
	Registries         []string `json:"registries,omitempty"`
	Internal           bool     `json:"internal,omitempty"`
	AllowAnonymousPull bool     `json:"allowAnonymousPull,omitempty"`
	Tags               []string `json:"tags,omitempty"`
}

// RegistryAccessSettingsRequest is the PUT body; all pointers so absent fields
// keep their stored value.
type RegistryAccessSettingsRequest struct {
	Host               *string   `json:"host,omitempty"`
	Registries         *[]string `json:"registries,omitempty"`
	Internal           *bool     `json:"internal,omitempty"`
	AllowAnonymousPull *bool     `json:"allowAnonymousPull,omitempty"`
	Tags               *[]string `json:"tags,omitempty"`
}

// RegistryAccessTokenCreateRequest mints a deploy token; the raw value is
// returned once and never stored.
type RegistryAccessTokenCreateRequest struct {
	Name       string   `json:"name"`
	Scopes     []string `json:"scopes,omitempty"`
	ExpiryDays int      `json:"expiryDays,omitempty"`
}

func RegistryAccessRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

func RegistryAccessIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// listRegistryAccesses godoc
// @Summary List registry accesses
// @Description Returns every registry endpoint with the nodes currently serving it, token hashes redacted (Pro feature)
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registry-accesses [get]
func listRegistryAccesses(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// getRegistryAccess godoc
// @Summary Get one registry access
// @Description Returns the endpoint with the nodes currently serving it, token hashes redacted (Pro feature)
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Access name"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registry-accesses/{name} [get]
func getRegistryAccess(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// createRegistryAccess godoc
// @Summary Create a registry access
// @Description Publishes one or more registries on a hostname. Every exposed registry
// @Description must share ONE type, so an access serves exactly one protocol (several
// @Description docker registries are fine — they namespace by path; an npm or generic
// @Description access exposes exactly one registry); an empty tag list means every node
// @Description serves it; internal restricts the endpoint to the constellation (Pro feature).
// @Tags registry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body RegistryAccessCreateRequest true "Access to create"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registry-accesses [post]
func createRegistryAccess(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// deleteRegistryAccess godoc
// @Summary Delete a registry access
// @Description Removes the endpoint. The registries it published and everything
// @Description stored in them are untouched (Pro feature).
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Access name"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registry-accesses/{name} [delete]
func deleteRegistryAccess(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// RegistryAccessSettingsRoute godoc
// @Summary Update a registry access
// @Description Replaces the host, exposed registries, visibility, anonymous-pull toggle
// @Description and/or serving tags. Absent fields keep their stored value (Pro feature).
// @Tags registry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Access name"
// @Param body body RegistryAccessSettingsRequest true "Settings to change"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registry-accesses/{name}/settings [put]
func RegistryAccessSettingsRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// RegistryAccessTokensRoute godoc
// @Summary Mint a registry deploy token
// @Description Returns the raw token ONCE — only its sha256 is stored. Scopes are
// @Description pull/push, optionally qualified by protocol (e.g. "docker:push") (Pro feature).
// @Tags registry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Access name"
// @Param body body RegistryAccessTokenCreateRequest true "Token to mint"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registry-accesses/{name}/tokens [post]
func RegistryAccessTokensRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// RegistryAccessTokenIdRoute godoc
// @Summary Delete a registry deploy token
// @Description Revokes the token on every node (Pro feature)
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Access name"
// @Param tokenName path string true "Token name"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registry-accesses/{name}/tokens/{tokenName} [delete]
func RegistryAccessTokenIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
