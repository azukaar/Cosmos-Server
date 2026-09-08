package pro

import (
	"sync"

	"github.com/nats-io/nats.go"
)

type SchedulerOptions struct {
	FollowerOnly bool
}

func StartScheduler(lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn, self string, strategies map[string]PlacementStrategy) {
	// Pro feature stub: the cluster scheduler is only available in Cosmos Pro.
}

func StopScheduler() {
	// Pro feature stub.
}

func DefaultStrategies() map[string]PlacementStrategy {
	return map[string]PlacementStrategy{}
}

func GetLeaderName(lock *sync.RWMutex, nc *nats.Conn, js nats.JetStreamContext) (string, bool) {
	return "", false
}

func StartSchedulerWithOptions(lock *sync.RWMutex, js nats.JetStreamContext, nc *nats.Conn, self string, strategies map[string]PlacementStrategy, opts SchedulerOptions) {
}
