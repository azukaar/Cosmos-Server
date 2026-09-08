// Community build stub of the Cosmos Pro feature set; handlers answer PRO001.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// ManagedDBCreateRequest is the create-instance body.
type ManagedDBCreateRequest struct {
	Name   string `json:"name" validate:"required,min=3,max=48,alphanum"`
	Engine string `json:"engine" validate:"omitempty,oneof=postgres"`
	// Image overrides the pinned default; never inferred from a running container.
	Image string `json:"image" validate:"omitempty,max=128"`
	// Port overrides proxy listen-port allocation. Zero means "allocate one".
	Port int `json:"port" validate:"omitempty,min=1024,max=65535"`
	// Pointer: the default is true, so a plain bool cannot tell "client sent
	// false" from "client omitted the field".
	RestrictToConstellation *bool `json:"restrictToConstellation"`
}

// ManagedDBUpdateRequest is the edit-instance body.
type ManagedDBUpdateRequest struct {
	RestrictToConstellation *bool `json:"restrictToConstellation"`
}

// ManagedDBLogicalRequest is the create-logical-database / rotate body.
type ManagedDBLogicalRequest struct {
	Database string `json:"database" validate:"required,min=1,max=48"`
	Role     string `json:"role" validate:"omitempty,min=1,max=48"`
}

// ManagedDBRoute handles the collection endpoint.
func ManagedDBRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// ManagedDBIdRoute handles the per-instance endpoint.
func ManagedDBIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// ManagedDBConnectionRoute returns the unredacted connection info for one instance.
func ManagedDBConnectionRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// ManagedDBDatabasesRoute handles the logical-database collection of one instance.
func ManagedDBDatabasesRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// ManagedDBDatabaseIdRoute handles one logical database of one instance.
func ManagedDBDatabaseIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// ManagedDBRotateRoute rotates one logical database role's password.
func ManagedDBRotateRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// listManagedDBs godoc
// @Summary List all managed databases
// @Description Returns every managed database record in the cluster, with live/down
// @Description status merged in from node heartbeats (Pro feature). Secrets are
// @Description redacted; use the connection endpoint to read them.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 503 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases [get]
func listManagedDBs(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// createManagedDB godoc
// @Summary Provision a managed database instance
// @Description Creates a single-node postgres container on THIS node and records
// @Description it in the cluster store (Pro feature). Not scheduled: the node that
// @Description serves this request becomes the instance's home node.
// @Tags databases
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ManagedDBCreateRequest true "Instance definition"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases [post]
func createManagedDB(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// getManagedDB godoc
// @Summary Get a managed database
// @Description Returns one record with live status and, when the home node answers,
// @Description its container/server state (Pro feature). Secrets are redacted.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name} [get]
func getManagedDB(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// updateManagedDB godoc
// @Summary Update a managed database instance
// @Description Changes settings that do not require recreating the container —
// @Description currently only the constellation restriction, which is re-applied to
// @Description the instance's proxy route on its home node (Pro feature).
// @Tags databases
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Param body body ManagedDBUpdateRequest true "Updated settings"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 502 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name} [put]
func updateManagedDB(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// getManagedDBConnection godoc
// @Summary Get connection info for a managed database
// @Description Returns the superuser credentials and connection URL for an instance,
// @Description plus one URL per logical database (Pro feature). This is the only
// @Description endpoint that returns secrets.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name}/connection [get]
func getManagedDBConnection(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// deleteManagedDB godoc
// @Summary Delete a managed database instance
// @Description Stops and removes the instance's container on its home node and drops
// @Description the record (Pro feature). The data volume is KEPT unless
// @Description ?removeVolume=true is passed.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Param removeVolume query bool false "Also destroy the data volume"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 502 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name} [delete]
func deleteManagedDB(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// createLogicalDB godoc
// @Summary Create a logical database and role
// @Description Creates one application database inside an instance, owned by a
// @Description freshly generated role scoped to it (Pro feature). Returns the
// @Description credentials once, in the response.
// @Tags databases
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Param body body ManagedDBLogicalRequest true "Logical database definition"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name}/databases [post]
func createLogicalDB(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// dropLogicalDB godoc
// @Summary Delete a logical database and its role
// @Description Drops one application database inside an instance, terminating its
// @Description open sessions first (Pro feature). Destroys that database's data.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Param database path string true "Logical database name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name}/databases/{database} [delete]
func dropLogicalDB(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// rotateLogicalDB godoc
// @Summary Rotate a logical database role's password
// @Description Generates a new password for one application role and returns the new
// @Description connection URL (Pro feature). Existing sessions keep working until
// @Description they reconnect.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Param database path string true "Logical database name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name}/databases/{database}/rotate [post]
func rotateLogicalDB(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
