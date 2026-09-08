// Community build stub of the Cosmos Pro feature set; handlers answer PRO001 and hooks do nothing.

package pro

import (
	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

type Deployment struct {
	Name string `json:"name" validate:"required,min=3,max=64,alphanum"`
	// Exactly one of the three replica modes (fixed Replicas, autoscale Min/MaxReplicas, fill ReplicaFill)
	// must be set; ValidateReplicaConfig enforces the exclusivity.
	Replicas int `json:"replicas,omitempty" validate:"omitempty,min=1"`
	// MinReplicas/MaxReplicas select load-based autoscaling clamped to [min,max], one step per cooldown.
	MinReplicas int `json:"minReplicas,omitempty" validate:"omitempty,min=1"`
	MaxReplicas int `json:"maxReplicas,omitempty" validate:"omitempty,min=1"`
	// ReplicaFill: one replica on every alive node matching Tags (DaemonSet-style).
	ReplicaFill bool `json:"replicaFill,omitempty"`
	// fill sub-mode: "full" (default) = every eligible node; "bare" = autoscale between 1 and the
	// eligible set; "empty" = bare with lazy scale-to-zero.
	ReplicaFillMode string `json:"replicaFillMode,omitempty" validate:"omitempty,oneof=full bare empty"`
	// Strategy selects which PlacementStrategy the scheduler uses for this
	// deployment. Empty is treated as "round-robin" for back-compat with
	// KV entries written before this field existed.
	Strategy string `json:"strategy" validate:"omitempty,oneof=round-robin least-busy"`
	// Tags filter eligible placement nodes. A deployment with Tags=["gpu"]
	// will only land on nodes whose ConstellationDevice.Tags contains "gpu".
	// Multiple tags are AND'd: ["gpu","nvme"] requires both. Empty means no
	// filter — any node is eligible.
	Tags []string `json:"tags,omitempty" validate:"omitempty,dive,min=1,max=64"`
	// Storage lists RCLONE remote names this deployment depends on. Checked
	// node-side in executeApply before docker.CreateService runs — a missing
	// remote produces StatusFail and flows through the existing fail-streak
	// quarantine path. Not a placement filter: RCLONE config is cluster-synced
	// via constellation, so every eligible node has the same remote set.
	// ${storage.NAME} in compose fields resolves to the mount path on apply.
	Storage []string                          `json:"storage,omitempty" validate:"omitempty,dive,min=1,max=64"`
	Compose docker.DockerServiceCreateRequest `json:"compose" validate:"required"`
	// PreserveVolumesOnRemove keeps the deployment's named volumes on disk on replica removal or
	// full delete; stamped as the cosmos-deployment-keep-volumes label on containers and volumes.
	PreserveVolumesOnRemove bool `json:"preserveVolumesOnRemove,omitempty"`
	// Owner marks a system-owned deployment (e.g. "seaweedfs:<instance>"); the HTTP API refuses user create/update/delete on it.
	Owner string `json:"owner,omitempty"`
	// Version is a server-assigned monotonic integer bumped on every create/update, stamped on containers
	// as cosmos-deployment-version so the scheduler can detect stale specs.
	Version int `json:"version"`
}

func DeploymentsRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

func DeploymentsIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// listDeployments godoc
// @Summary List all cluster deployments
// @Description Returns all deployment definitions from the constellation-deployments KV (Pro feature)
// @Tags deployments
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/constellation/deployments [get]
func listDeployments(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// createDeployment godoc
// @Summary Create a new cluster deployment
// @Description Creates a new deployment definition in the constellation-deployments KV (Pro feature)
// @Tags deployments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body Deployment true "Deployment definition"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/constellation/deployments [post]
func createDeployment(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// getDeployment godoc
// @Summary Get a cluster deployment by name
// @Description Returns a single deployment definition (Pro feature)
// @Tags deployments
// @Produce json
// @Security BearerAuth
// @Param name path string true "Deployment name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/deployments/{name} [get]
func getDeployment(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// updateDeployment godoc
// @Summary Update a cluster deployment
// @Description Updates an existing deployment definition (Pro feature)
// @Tags deployments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Deployment name"
// @Param body body Deployment true "Updated deployment"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/deployments/{name} [put]
func updateDeployment(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// deleteDeployment godoc
// @Summary Delete a cluster deployment
// @Description Removes a deployment from the constellation-deployments KV (Pro feature)
// @Tags deployments
// @Produce json
// @Security BearerAuth
// @Param name path string true "Deployment name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/deployments/{name} [delete]
func deleteDeployment(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// DeploymentHealth is the per-deployment cluster status returned by the health
// endpoint. `desired` is the configured replica count; `actual` is how many
// nodes currently report running it (from each node's constellation-nodes
// heartbeat); `broken` mirrors the scheduler's quarantine state.
type DeploymentHealth struct {
	// Desired is mode-dependent: the fixed count, the alive tag-matching node count in full-fill,
	// or the current actual clamped to [min,max] in autoscale.
	Desired int `json:"desired"`
	Actual  int `json:"actual"`
	// echo of the deployment's replica-mode config so the UI needn't re-fetch the spec
	MinReplicas int  `json:"minReplicas,omitempty"`
	MaxReplicas int  `json:"maxReplicas,omitempty"`
	ReplicaFill bool `json:"replicaFill,omitempty"`
	// echoes the fill sub-mode ("" = full)
	ReplicaFillMode string   `json:"replicaFillMode,omitempty"`
	Nodes           []string `json:"nodes"`
	Broken          bool     `json:"broken"`
	BrokenReason    string   `json:"brokenReason,omitempty"`
	// Version is the desired (current) spec version. UpToDate counts how many of
	// the nodes reporting this deployment are running that version. Updating is
	// true while a spec change is still rolling out — i.e. at least one reporting
	// node is on an older version. A deployment is only fully healthy when
	// Actual == Desired AND UpToDate == Actual (Updating == false).
	Version  int  `json:"version"`
	UpToDate int  `json:"upToDate"`
	Updating bool `json:"updating"`
}

// DeploymentsHealthRoute godoc
// @Summary Cluster health for all deployments
// @Description Returns desired vs. actual replica placement per deployment plus
// @Description scheduler quarantine state (Pro feature). Read-only: safe to call
// @Description from any node, not just the scheduler leader.
// @Tags deployments
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 503 {object} utils.HTTPErrorResult
// @Router /api/constellation/deployments/health [get]
func DeploymentsHealthRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// DeploymentsUnbrokeRoute godoc
// @Summary Clear a deployment's quarantine
// @Description Removes a deployment from the scheduler's quarantine set so the
// @Description next reconcile cycle attempts to place it again (Pro feature).
// @Tags deployments
// @Produce json
// @Security BearerAuth
// @Param name path string true "Deployment name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/constellation/deployments/{name}/unbroke [post]
func DeploymentsUnbrokeRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// NodesUnbrokeRoute godoc
// @Summary Clear a node's quarantine
// @Description Removes a node from the scheduler's quarantine set and resets its
// @Description fail streak so the next reconcile cycle is free to place work on it
// @Description again (Pro feature).
// @Tags deployments
// @Produce json
// @Security BearerAuth
// @Param name path string true "Node (device) name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Router /api/constellation/nodes/{name}/unbroke [post]
func NodesUnbrokeRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
