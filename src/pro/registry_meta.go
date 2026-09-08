// Community build stub of the Cosmos Pro feature set: types exist so the shared
// code and the SDK generator compile; handlers answer PRO001 and hooks do nothing.

package pro

import (
	"time"
)

// RegistryStats is the per-registry rollup, maintained by CAS increments from
// every serving node and reconciled by the sweep.
type RegistryStats struct {
	SizeBytes    int64     `json:"sizeBytes"`
	PackageCount int64     `json:"packageCount"`
	VersionCount int64     `json:"versionCount"`
	PullCount    int64     `json:"pullCount"`
	PushCount    int64     `json:"pushCount"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
