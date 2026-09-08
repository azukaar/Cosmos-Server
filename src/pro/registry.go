// Community build stub of the Cosmos Pro feature set: types exist so the shared
// code and the SDK generator compile; handlers answer PRO001 and hooks do nothing.

package pro

import (
	"github.com/nats-io/nats.go"
	"sync"
	"time"
)

// RegistryStorage describes where a registry's blobs live. Exactly one backend
// family's fields are meaningful; the others stay zero.
type RegistryStorage struct {
	Backend string `json:"backend" validate:"required,oneof=seaweedfs s3 local"`
	// SeaweedFS is the managed instance name when Backend is "seaweedfs".
	SeaweedFS string `json:"seaweedfs,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	// External S3 only.
	Endpoint  string `json:"endpoint,omitempty"`
	AccessKey string `json:"accessKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
	Region    string `json:"region,omitempty"`
	Path      string `json:"path,omitempty"`
}

// RegistryToken is a deploy credential stored on the access record so it
// replicates with it; only the hash is kept.
type RegistryToken struct {
	Name        string `json:"name" validate:"required,min=1,max=64"`
	TokenHash   string `json:"tokenHash"`
	TokenSuffix string `json:"tokenSuffix"`
	// Scopes are "pull"/"push", optionally type-qualified ("docker:push").
	Scopes []string `json:"scopes,omitempty"`
	// ExpiresAt zero means never.
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	// LastUsedAt is written lazily by the protocol paths (throttled).
	LastUsedAt time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// RegistryInstance is the persistent record of one registry. Who may reach it
// is an access concern — see RegistryAccess.
type RegistryInstance struct {
	// Name is alphanumeric and lowercased so every derived identifier (bucket
	// name, KV key prefix, OCI repository path) is valid without transformation.
	Name string `json:"name" validate:"required,min=3,max=27,alphanum"`
	// Type is the protocol this registry's contents speak. Immutable.
	Type    string          `json:"type" validate:"required,oneof=docker npm static generic pypi"`
	Storage RegistryStorage `json:"storage"`
	// QuotaBytes caps the registry's stored size; 0 is unlimited.
	QuotaBytes int64 `json:"quotaBytes"`
	// Status: provisioning -> ready; deleting during teardown.
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// RegistryAccess is an endpoint publishing one or more registries of one type.
type RegistryAccess struct {
	Name string `json:"name" validate:"required,min=3,max=27,alphanum"`
	// Host is required: docker clients only address a registry by hostname over
	// TLS, and npm's tarball URLs are absolute.
	Host string `json:"host" validate:"required"`
	// Paths outside Registries 404, never 403 — no existence oracle.
	Registries []string `json:"registries" validate:"required,min=1"`
	// Internal restricts the endpoint to the constellation; the direct mux
	// handler enforces it too, since a request can bypass the route.
	Internal           bool `json:"internal"`
	AllowAnonymousPull bool `json:"allowAnonymousPull"`
	// Unlike SeaweedFS, EMPTY Tags is meaningful and the default: no filter, so
	// every node serves the endpoint.
	Tags   []string        `json:"tags"`
	Tokens []RegistryToken `json:"tokens,omitempty"`
	// Status: ready; deleting during teardown.
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// RegistryStatus is a registry record plus derived data; nothing here is persisted.
type RegistryStatus struct {
	RegistryInstance
	Accesses []string      `json:"accesses"`
	Stats    RegistryStats `json:"stats"`
}

type RegistryAccessStatus struct {
	RegistryAccess
	ServingNodes []string `json:"servingNodes"`
}

// ListRegistriesWithStatus returns records, the accesses publishing them and
// their stats; a failed stats read leaves the zero rollup.
func ListRegistriesWithStatus(lock *sync.RWMutex, js nats.JetStreamContext) ([]RegistryStatus, error) {
	// Pro feature stub.
	var r0 []RegistryStatus
	var r1 error
	return r0, r1
}

func ListRegistryAccessesWithStatus(lock *sync.RWMutex, js nats.JetStreamContext) ([]RegistryAccessStatus, error) {
	// Pro feature stub.
	var r0 []RegistryAccessStatus
	var r1 error
	return r0, r1
}

// RegistryAccessesServedHere returns the endpoints this node serves given its
// constellation tags; a KV failure degrades to an empty list.
func RegistryAccessesServedHere(lock *sync.RWMutex, js nats.JetStreamContext, nodeTags []string) []string {
	// Pro feature stub.
	var r0 []string
	return r0
}
