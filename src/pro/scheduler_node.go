package pro

import (
	"github.com/nats-io/nats.go"
)

func RegisterNodeDispatchHandler(nc *nats.Conn, selfSanitized string) error {
	// Pro feature stub: cluster dispatch is only available in Cosmos Pro.
	return nil
}
