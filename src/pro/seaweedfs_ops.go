// Community build stub of the Cosmos Pro feature set: types exist so the shared
// code and the SDK generator compile; handlers answer PRO001 and hooks do nothing.

package pro

import (
	"github.com/nats-io/nats.go"
)

// RegisterSeaweedFSResponder subscribes this node to its own op subject.
func RegisterSeaweedFSResponder(nc *nats.Conn, self string) (*nats.Subscription, error) {
	// Pro feature stub.
	var r0 *nats.Subscription
	var r1 error
	return r0, r1
}
