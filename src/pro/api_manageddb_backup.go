// Community build stub of the Cosmos Pro feature set; handlers answer PRO001.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// ManagedDBBackupRequest is the configure-backup body.
type ManagedDBBackupRequest struct {
	Enabled         bool   `json:"enabled"`
	Repository      string `json:"repository" validate:"omitempty,max=512"`
	Crontab         string `json:"crontab" validate:"omitempty,max=128"`
	CrontabForget   string `json:"crontabForget" validate:"omitempty,max=128"`
	RetentionPolicy string `json:"retentionPolicy" validate:"omitempty,max=512"`
}

// ManagedDBRestoreRequest is the restore-as-a-new-instance body.
type ManagedDBRestoreRequest struct {
	// SourceName's backup configuration (repository AND password) is used to
	// read the snapshot; the source instance itself is never touched.
	SourceName string `json:"sourceName" validate:"required"`
	// SnapshotID is a restic snapshot id, long or short; "latest" works.
	SnapshotID string `json:"snapshotId" validate:"required"`
	// NewName is the instance to create. It must not exist.
	NewName string `json:"newName" validate:"required,min=3,max=48"`
	// Databases restricts what is restored; empty means every logical database.
	Databases []string `json:"databases"`
	// Port and RestrictToConstellation default as in a plain create (allocate a
	// port; restrict to the constellation).
	Port                    int   `json:"port" validate:"omitempty,min=1024,max=65535"`
	RestrictToConstellation *bool `json:"restrictToConstellation"`
}

// ManagedDBBackupRoute handles the backup configuration of one instance.
func ManagedDBBackupRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// ManagedDBBackupRunRoute triggers a backup now.
func ManagedDBBackupRunRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// ManagedDBBackupSnapshotsRoute lists an instance's snapshots.
func ManagedDBBackupSnapshotsRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// ManagedDBRestoreRoute restores a snapshot into a brand-new instance.
func ManagedDBRestoreRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// configureManagedDBBackup godoc
// @Summary Configure backups for a managed database
// @Description Sets the restic repository and schedules for an instance's backups
// @Description (Pro feature). The repository is interpreted on the instance's HOME
// @Description node and is initialised lazily, on the first run. The repository
// @Description password is generated once and never rotated by this endpoint, since
// @Description it is the only key to the existing snapshots.
// @Tags databases
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Param body body ManagedDBBackupRequest true "Backup configuration"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name}/backup [put]
func configureManagedDBBackup(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// unconfigureManagedDBBackup godoc
// @Summary Remove a managed database's backup configuration
// @Description Clears the instance's backup settings and deregisters its scheduled
// @Description jobs (Pro feature). The restic repository and every snapshot in it
// @Description are left untouched — but the password is dropped with the record, so
// @Description recovering those snapshots afterwards means recovering the password
// @Description from the backup password log.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name}/backup [delete]
func unconfigureManagedDBBackup(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// runManagedDBBackupNow godoc
// @Summary Back up a managed database now
// @Description Starts a backup of the instance on its home node and returns
// @Description immediately (Pro feature): a dump plus an upload is unbounded work.
// @Description Progress and outcome are reported through the instance's events and
// @Description through the snapshot listing.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 502 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name}/backup/run [post]
func runManagedDBBackupNow(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// listManagedDBSnapshots godoc
// @Summary List a managed database's backup snapshots
// @Description Returns the restic snapshots tagged for this instance, newest first
// @Description (Pro feature). An empty list is returned — not an error — while the
// @Description repository has not been written to yet.
// @Tags databases
// @Produce json
// @Security BearerAuth
// @Param name path string true "Database instance name"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 502 {object} utils.HTTPErrorResult
// @Router /api/constellation/databases/{name}/backup/snapshots [get]
func listManagedDBSnapshots(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
