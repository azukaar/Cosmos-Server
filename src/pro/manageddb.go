// Community build stub of the Cosmos Pro feature set.

package pro

import (
	"github.com/nats-io/nats.go"
	"sync"
	"time"
)

// ManagedDatabase is the persistent record of one managed database instance.
type ManagedDatabase struct {
	Name    string `json:"name" validate:"required,min=3,max=48,alphanum"`
	Engine  string `json:"engine" validate:"omitempty,oneof=postgres"`
	Image   string `json:"image"`
	Version string `json:"version,omitempty"`
	// HomeNode is the constellation device name, used verbatim as the
	// constellation-nodes KV key and in NATS subjects.
	HomeNode     string `json:"homeNode"`
	HomeNodeName string `json:"homeNodeName,omitempty"`
	// HomeIP is the home node's Nebula IP; ${db.<name>.host} expands to it.
	HomeIP string `json:"homeIP"`
	// Port is the Cosmos socket proxy listen port, not a docker port publish.
	Port int `json:"port"`
	// RestrictToConstellation blocks connections from outside the constellation; defaults to true on create.
	RestrictToConstellation bool `json:"restrictToConstellation"`
	// SuperUser/SuperPassword are stored in the clear; redacted from the list API.
	SuperUser     string                      `json:"superUser"`
	SuperPassword string                      `json:"superPassword"`
	CreatedAt     time.Time                   `json:"createdAt"`
	Databases     map[string]ManagedLogicalDB `json:"databases,omitempty"`
	// Backup is nil until configured; interpreted entirely on the home node.
	Backup *ManagedDBBackup `json:"backup,omitempty"`
}

// ManagedDBBackup is the backup configuration of one instance.
type ManagedDBBackup struct {
	// Enabled gates scheduled jobs only; manual runs and snapshot listing still work.
	Enabled bool `json:"enabled"`
	// Repository is a restic repository (path or rclone:<remote>:<path>), resolved on the home node.
	Repository string `json:"repository"`
	// Password is generated once and never regenerated; redacted like the superuser password.
	Password string `json:"password"`
	// Crontab and CrontabForget are 6-field (seconds-first) schedules, validated at configure time.
	Crontab       string `json:"crontab"`
	CrontabForget string `json:"crontabForget"`
	// RetentionPolicy is passed verbatim to `restic forget` as arguments.
	RetentionPolicy string    `json:"retentionPolicy"`
	ConfiguredAt    time.Time `json:"configuredAt"`
}

// ManagedLogicalDB is one application's database-and-role pair inside an instance.
type ManagedLogicalDB struct {
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListManagedDatabases returns every stored record, sorted by name.
func ListManagedDatabases(lock *sync.RWMutex, js nats.JetStreamContext) ([]ManagedDatabase, error) {
	// Pro feature stub.
	var r0 []ManagedDatabase
	var r1 error
	return r0, r1
}
