package constellation

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/azukaar/cosmos-server/src/utils"
)

// 'constellation-nodes' is the cluster's only liveness signal: an unwritable bucket degrades everything.

func TestNodesKVConfig_FileBackedInHA(t *testing.T) {
	// A majority restarting together strands a memory-backed bucket resolvable but unwritable.
	if got := storageFor(t, true); got != nats.FileStorage {
		t.Fatalf("HA constellation-nodes must be file-backed, got %v", got)
	}
}

func TestNodesKVConfig_MemoryBackedWhenNotHA(t *testing.T) {
	// Single-server installs have no majority to lose and no reason to write heartbeats to disk.
	if got := storageFor(t, false); got != nats.MemoryStorage {
		t.Fatalf("non-HA constellation-nodes should stay in memory, got %v", got)
	}
}

func storageFor(t *testing.T, ha bool) nats.StorageType {
	t.Helper()
	setupTestEnv(t, func(c *utils.Config) {
		c.ConstellationConfig.NATSReplicas = 1
		if ha {
			c.ConstellationConfig.NATSReplicas = 3
		}
	})
	return nodesKVConfig().Storage
}

// An upgraded cluster's bucket already exists, so a storage-type mismatch must be detected to trigger recreation.
func TestKVTopologyOutdated_SpotsAStorageTypeChange(t *testing.T) {
	have := nats.StreamConfig{Storage: nats.MemoryStorage, Replicas: 3}
	want := nats.KeyValueConfig{Storage: nats.FileStorage, Replicas: 3}

	if !kvTopologyOutdated(have, want) {
		t.Fatal("a memory bucket must be recognised as outdated against a file-backed declaration")
	}
}

func TestKVTopologyOutdated_SpotsAReplicaChange(t *testing.T) {
	have := nats.StreamConfig{Storage: nats.FileStorage, Replicas: 1}
	want := nats.KeyValueConfig{Storage: nats.FileStorage, Replicas: 3}

	if !kvTopologyOutdated(have, want) {
		t.Fatal("the pre-existing replica-count check must survive")
	}
}

func TestKVTopologyOutdated_LeavesAMatchingBucketAlone(t *testing.T) {
	// Dropping the bucket costs every node its liveness; a correct shape must never trigger it.
	have := nats.StreamConfig{Storage: nats.FileStorage, Replicas: 3}
	want := nats.KeyValueConfig{Storage: nats.FileStorage, Replicas: 3}

	if kvTopologyOutdated(have, want) {
		t.Fatal("recreated a bucket that already had the declared shape")
	}
}

// Only Storage and Replicas decide whether the bucket survives a restart; other drift is ignored.
func TestKVTopologyOutdated_IgnoresNonStructuralDrift(t *testing.T) {
	have := nats.StreamConfig{Storage: nats.FileStorage, Replicas: 3, MaxAge: time.Minute, Description: "old"}
	want := nats.KeyValueConfig{Storage: nats.FileStorage, Replicas: 3, TTL: 10 * time.Second}

	if kvTopologyOutdated(have, want) {
		t.Fatal("a TTL/description difference must not drop the bucket")
	}
}

func TestDescribeKVTopology_NamesBothShapes(t *testing.T) {
	if got := describeKVTopology(nats.MemoryStorage, 3); got != "memory/R3" {
		t.Fatalf("got %q", got)
	}
	if got := describeKVTopology(nats.FileStorage, 1); got != "file/R1" {
		t.Fatalf("got %q", got)
	}
}
