// Community build stub of the Cosmos Pro feature set; handlers answer PRO001.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// SeaweedFSCreateRequest is the POST body; RestrictToConstellation is a *bool
// so "absent" defaults to true.
type SeaweedFSCreateRequest struct {
	Name                    string   `json:"name"`
	Tags                    []string `json:"tags,omitempty"`
	Image                   string   `json:"image,omitempty"`
	FilerReplicas           int      `json:"filerReplicas,omitempty"`
	IndexMode               string   `json:"indexMode,omitempty"`
	DefaultReplication      string   `json:"defaultReplication,omitempty"`
	VolumeSizeLimitMB       int      `json:"volumeSizeLimitMB,omitempty"`
	MinFreeSpace            string   `json:"minFreeSpace,omitempty"`
	MaxStorageGBPerNode     int      `json:"maxStorageGBPerNode,omitempty"`
	RestrictToConstellation *bool    `json:"restrictToConstellation,omitempty"`
}

func SeaweedFSRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

func SeaweedFSIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// listSeaweedFS godoc
// @Summary List managed SeaweedFS instances
// @Description Returns every instance with heartbeat-derived status, secrets redacted (Pro feature)
// @Tags seaweedfs
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/seaweedfs [get]
func listSeaweedFS(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// SeaweedFSStatusRoute is the unredacted view: full status plus the S3
// credentials and endpoint URLs.
func SeaweedFSStatusRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// createSeaweedFS godoc
// @Summary Create a managed SeaweedFS instance
// @Description Provisions 3 pinned masters on the first three constellation managers,
// @Description a fill-mode volume-server deployment on the chosen tags, a filer+S3
// @Description deployment behind the tunnel LB, and a managed postgres for the filer
// @Description store (Pro feature). Refused when fewer than 3 managers are online.
// @Tags seaweedfs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body SeaweedFSCreateRequest true "Instance to create"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/seaweedfs [post]
func createSeaweedFS(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// deleteSeaweedFS godoc
// @Summary Delete a managed SeaweedFS instance
// @Description Tears down deployments, masters and (by default) the filer database.
// @Description Volume-server data volumes are PRESERVED unless purgeData=true;
// @Description keepFilerDB=true keeps the managed database entirely (Pro feature).
// @Tags seaweedfs
// @Produce json
// @Security BearerAuth
// @Param name path string true "Instance name"
// @Param purgeData query boolean false "Also remove the data volumes on every node"
// @Param keepFilerDB query boolean false "Keep the auto-provisioned filer database"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/seaweedfs/{name} [delete]
func deleteSeaweedFS(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// SeaweedFSRestrictRoute toggles constellation-only access (compose rewrite +
// version bump).
func SeaweedFSRestrictRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// SeaweedFSStorageRoute changes the per-node storage cap. Shrinking below what
// a node already holds deletes nothing; the node just stops receiving new
// volumes until it is back under the ceiling.
func SeaweedFSStorageRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// SeaweedFSJobsRoute replaces the instance's maintenance-job configuration.
func SeaweedFSJobsRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// SeaweedFSRepairRoute starts the "this hardware is gone" repair job on the
// locus node.
func SeaweedFSRepairRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// SeaweedFSDrainRoute evacuates one live volume server (body:
// {"device": "<device name>"}).
func SeaweedFSDrainRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// SeaweedFSUpgradeRoute starts a rolling upgrade (body: {"image": "..."}).
func SeaweedFSUpgradeRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// SeaweedFSReplaceMasterRoute swaps one pinned master for another manager
// (body: {"oldDevice": "...", "newDevice": "..."} — newDevice optional).
func SeaweedFSReplaceMasterRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
