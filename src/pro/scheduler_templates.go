// Community build stub of the Cosmos Pro feature set; handlers answer PRO001 and hooks do nothing.

package pro

// NodeIdentity is exposed to template expansion as ${node_*} variables.
type NodeIdentity struct {
	// human-visible constellation device name
	DeviceName string
	// wire identity (constellation-nodes KV key and per-node NATS subjects); identical to DeviceName
	SanitizedName string
	// Nebula IP, e.g. "100.64.0.3"
	IP string
	// node mode: 0 = free, 1 = pro client, 2 = lighthouse/relay; exposed as ${node_mode}
	CosmosNode int
}

// SetNodeIdentityProvider is wired by the constellation package at startup.
func SetNodeIdentityProvider(f func() NodeIdentity) {
	// Pro feature stub.
}

// SetMountedStorageProvider is wired by constellation to storage.CachedRemoteStorageList.
func SetMountedStorageProvider(f func() []string) {
	// Pro feature stub.
}

// SetManagedDatabaseProvider is wired by constellation to ListManagedDatabases.
func SetManagedDatabaseProvider(f func() []ManagedDatabase) {
	// Pro feature stub.
}

// SetSeaweedFSProvider is wired by constellation to ListSeaweedFSWithStatus.
func SetSeaweedFSProvider(f func() []SeaweedFSStatus) {
	// Pro feature stub.
}
