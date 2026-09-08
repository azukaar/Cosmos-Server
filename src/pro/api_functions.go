// Community build stub of the Cosmos Pro feature set; handlers answer PRO001.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// FunctionRuntimesRoute lists the runtime table.
// @Router /api/constellation/function-runtimes [get]
func FunctionRuntimesRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// FunctionsRoute lists (GET) or creates (POST) functions.
// @Router /api/constellation/functions [get]
// @Router /api/constellation/functions [post]
func FunctionsRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// FunctionsIdRoute reads (GET), replaces (PUT) or deletes (DELETE) a function.
// @Param name path string true "Function name"
// @Router /api/constellation/functions/{name} [get]
// @Router /api/constellation/functions/{name} [put]
// @Router /api/constellation/functions/{name} [delete]
func FunctionsIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// FunctionsDeployRoute pins a version and (re)writes the deployment.
// Body: {"version": "1.2.3"} — empty or "latest" = the registry's latest.
// @Param name path string true "Function name"
// @Router /api/constellation/functions/{name}/deploy [post]
func FunctionsDeployRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// FunctionsVersionsRoute lists the published versions of the function's
// package (newest first) and which one is the registry's latest.
// @Param name path string true "Function name"
// @Router /api/constellation/functions/{name}/versions [get]
func FunctionsVersionsRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// FunctionsInvokeRoute invokes the function from this node (admin test).
// Body: {"method": "POST", "path": "/", "body": "..."}.
// @Param name path string true "Function name"
// @Router /api/constellation/functions/{name}/invoke [post]
func FunctionsInvokeRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
