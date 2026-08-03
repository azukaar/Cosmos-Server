package pro

import (
	"sync"

	"github.com/nats-io/nats.go"
)

func ClientHeartbeatInit(clientConfigLock *sync.RWMutex, js nats.JetStreamContext, replicas int) {
	// Pro feature stub: clustered deployments are only available in Cosmos Pro.
}
