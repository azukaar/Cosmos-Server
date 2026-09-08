package constellation

import (
	"net/http"

	"github.com/azukaar/cosmos-server/src/pro"
)

// Thin bridges keeping `pro` independent of `constellation`; js/nc are read at call time so reconnects are picked up.

func ManagedDBRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBRoute(w, req, &clientConfigLock, js)
}

func ManagedDBIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBIdRoute(w, req, &clientConfigLock, js, nc)
}

func ManagedDBConnectionRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBConnectionRoute(w, req, &clientConfigLock, js)
}

func ManagedDBDatabasesRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBDatabasesRoute(w, req, &clientConfigLock, js, nc)
}

func ManagedDBDatabaseIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBDatabaseIdRoute(w, req, &clientConfigLock, js, nc)
}

func ManagedDBRotateRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBRotateRoute(w, req, &clientConfigLock, js, nc)
}

func ManagedDBBackupRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBBackupRoute(w, req, &clientConfigLock, js)
}

func ManagedDBBackupRunRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBBackupRunRoute(w, req, &clientConfigLock, js, nc)
}

func ManagedDBBackupSnapshotsRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBBackupSnapshotsRoute(w, req, &clientConfigLock, js, nc)
}

// ManagedDBRestoreRoute must be registered before /databases/{name}, which would otherwise match "restore".
func ManagedDBRestoreRoute(w http.ResponseWriter, req *http.Request) {
	pro.ManagedDBRestoreRoute(w, req, &clientConfigLock, js)
}

// ManagedDatabaseList is best-effort: a KV failure returns nil so ${db.*} placeholders stay literal.
func ManagedDatabaseList() []pro.ManagedDatabase {
	dbs, err := pro.ListManagedDatabases(&clientConfigLock, js)
	if err != nil {
		return nil
	}
	return dbs
}
