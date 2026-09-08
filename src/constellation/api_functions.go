package constellation

import (
	"net/http"

	"github.com/azukaar/cosmos-server/src/pro"
)

func FunctionRuntimesRoute(w http.ResponseWriter, req *http.Request) {
	pro.FunctionRuntimesRoute(w, req, &clientConfigLock, js)
}

func FunctionsRoute(w http.ResponseWriter, req *http.Request) {
	pro.FunctionsRoute(w, req, &clientConfigLock, js)
}

func FunctionsIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.FunctionsIdRoute(w, req, &clientConfigLock, js)
}

func FunctionsDeployRoute(w http.ResponseWriter, req *http.Request) {
	pro.FunctionsDeployRoute(w, req, &clientConfigLock, js)
}

func FunctionsVersionsRoute(w http.ResponseWriter, req *http.Request) {
	pro.FunctionsVersionsRoute(w, req, &clientConfigLock, js)
}

func FunctionsInvokeRoute(w http.ResponseWriter, req *http.Request) {
	pro.FunctionsInvokeRoute(w, req, &clientConfigLock, js)
}
