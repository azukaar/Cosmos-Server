// Community build stub of the Cosmos Pro feature set.

package pro

import (
	"github.com/nats-io/nats.go"
)

// RegisterManagedDBResponder subscribes to this node's managed-DB op subject; self must be the constellation device name verbatim.
func RegisterManagedDBResponder(nc *nats.Conn, self string) (*nats.Subscription, error) {
	// Pro feature stub.
	var r0 *nats.Subscription
	var r1 error
	return r0, r1
}
