// Community build stub of the Cosmos Pro feature set: types exist so the shared
// code and the SDK generator compile; handlers answer PRO001 and hooks do nothing.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"sync"
)

// RegistryNodeInfo is what the serving paths need to know about this node. It
// is separate from NodeIdentity because that one is the template-expansion
// identity and deliberately carries no secret.
type RegistryNodeInfo struct {
	DeviceName string
	IP         string
	Tags       []string
	APIKey     string
}

func SetRegistryNodeProvider(f func() RegistryNodeInfo) {
	// Pro feature stub.
}

func SetRegistryClusterHandles(f func() (*sync.RWMutex, nats.JetStreamContext, *nats.Conn)) {
	// Pro feature stub.
}

// RegisterRegistryProtocolRoutes wires the protocol endpoints onto the router.
// Called from InitServer BEFORE proxy.BuildFromConfig so the direct handler
// wins over the materialized loopback route (mux matches in registration
// order). /v2 is claimed before the npm and generic catch-alls, which live at
// the root.
func RegisterRegistryProtocolRoutes(router *mux.Router) {
	// Pro feature stub.
}

// BuildRegistryAccessRoute renders the advertisement route for an access.
// pathPrefix must always be the path space RegisterRegistryProtocolRoutes
// claims for this access. SmartShield applies to external accesses only.
func BuildRegistryAccessRoute(acc RegistryAccess, pathPrefix string) utils.ProxyRouteConfig {
	// Pro feature stub.
	var r0 utils.ProxyRouteConfig
	return r0
}

// StartRegistryServing starts the registry KV watcher and the upload-session
// janitor.
func StartRegistryServing() {
	// Pro feature stub.
}
