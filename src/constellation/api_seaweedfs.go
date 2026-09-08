package constellation

import (
	"net/http"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/azukaar/cosmos-server/src/pro"
)

// Thin bridges keeping `pro` independent of `constellation`; js/nc are read at call time so reconnects are picked up.

func SeaweedFSRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSRoute(w, req, &clientConfigLock, js, nc)
}

func SeaweedFSIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSIdRoute(w, req, &clientConfigLock, js, nc)
}

func SeaweedFSStatusRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSStatusRoute(w, req, &clientConfigLock, js)
}

func SeaweedFSRestrictRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSRestrictRoute(w, req, &clientConfigLock, js)
}

func SeaweedFSJobsRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSJobsRoute(w, req, &clientConfigLock, js)
}

func SeaweedFSStorageRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSStorageRoute(w, req, &clientConfigLock, js)
}

func SeaweedFSBackupRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSBackupRoute(w, req, &clientConfigLock, js)
}

func SeaweedFSBackupRunRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSBackupRunRoute(w, req, &clientConfigLock, js, nc)
}

func SeaweedFSBackupSnapshotsRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSBackupSnapshotsRoute(w, req, &clientConfigLock, js, nc)
}

func SeaweedFSRepairRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSRepairRoute(w, req, &clientConfigLock, js, nc)
}

func SeaweedFSDrainRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSDrainRoute(w, req, &clientConfigLock, js, nc)
}

func SeaweedFSUpgradeRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSUpgradeRoute(w, req, &clientConfigLock, js, nc)
}

func SeaweedFSReplaceMasterRoute(w http.ResponseWriter, req *http.Request) {
	pro.SeaweedFSReplaceMasterRoute(w, req, &clientConfigLock, js, nc)
}

// SeaweedFSList is best-effort: a KV failure returns nil, treated as "unavailable", never "empty cluster".
func SeaweedFSList() []pro.SeaweedFSStatus {
	list, err := pro.ListSeaweedFSWithStatus(&clientConfigLock, js)
	if err != nil {
		return nil
	}
	return list
}

// SeaweedFSClusterHandles hands pro the current cluster handles, read at call time.
func SeaweedFSClusterHandles() (*sync.RWMutex, nats.JetStreamContext, *nats.Conn) {
	return &clientConfigLock, js, nc
}
