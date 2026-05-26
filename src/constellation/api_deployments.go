package constellation

import (
	"net/http"

	"github.com/azukaar/cosmos-server/src/pro"
)

func DeploymentsRoute(w http.ResponseWriter, req *http.Request) {
	pro.DeploymentsRoute(w, req, &clientConfigLock, js)
}

func DeploymentsIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.DeploymentsIdRoute(w, req, &clientConfigLock, js)
}

func DeploymentsHealthRoute(w http.ResponseWriter, req *http.Request) {
	pro.DeploymentsHealthRoute(w, req, &clientConfigLock, js)
}

func DeploymentsUnbrokeRoute(w http.ResponseWriter, req *http.Request) {
	pro.DeploymentsUnbrokeRoute(w, req, &clientConfigLock, js)
}

func NodesUnbrokeRoute(w http.ResponseWriter, req *http.Request) {
	pro.NodesUnbrokeRoute(w, req, &clientConfigLock, js)
}
