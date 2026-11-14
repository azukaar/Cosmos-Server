package docker

import (
	"testing"
)

func TestNetworkModeContainerRef(t *testing.T) {
	cases := map[string]bool{
		"container:foo":    true,
		"container:abc123": true,
		"service:foo":      false,
		"bridge":           false,
		"host":             false,
		"":                 false,
		"cosmos-net":       false,
	}
	for in, want := range cases {
		if got := NetworkModeContainerRef(in); got != want {
			t.Errorf("NetworkModeContainerRef(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNetworkModeServiceRef(t *testing.T) {
	cases := map[string]bool{
		"service:foo":   true,
		"container:foo": false,
		"bridge":        false,
		"":              false,
	}
	for in, want := range cases {
		if got := NetworkModeServiceRef(in); got != want {
			t.Errorf("NetworkModeServiceRef(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNetworkModeRefTarget(t *testing.T) {
	cases := map[string]string{
		"container:foo":    "foo",
		"container:abc123": "abc123",
		"service:bar":      "bar",
		"bridge":           "",
		"host":             "",
		"":                 "",
		"container:":       "",
	}
	for in, want := range cases {
		if got := NetworkModeRefTarget(in); got != want {
			t.Errorf("NetworkModeRefTarget(%q) = %q, want %q", in, got, want)
		}
	}
}

// ContainerRefToName must leave non-reference modes and unresolvable
// references untouched, and normalize resolvable refs to container:<name>.
// The "newName" cases simulate Docker's behavior of accepting both a name and
// an ID prefix for the same container — without a daemon we cannot exercise
// ResolveContainerRefToName, so we cover the non-resolution paths here and the
// resolution path via the daemon-backed integration flows.
func TestContainerRefToNamePassthrough(t *testing.T) {
	cases := map[string]string{
		"bridge":     "bridge",
		"host":       "host",
		"default":    "default",
		"cosmos-net": "cosmos-net",
		"":           "",
		"container:": "container:",
		"service:":   "service:",
	}
	for in, want := range cases {
		if got := ContainerRefToName(in); got != want {
			t.Errorf("ContainerRefToName(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestCombineLabelTarget mimics RecreateDepedencies' label matching: after
// normalization the label holds container:<name>, and a dependent must be
// recreated when that name (or a compose service alias) matches the recreated
// container's name.
func TestCombineLabelTarget(t *testing.T) {
	containerName := "app-db"
	cases := []struct {
		label   string
		wantHit bool
	}{
		{"container:" + containerName, true},
		{"service:" + containerName, true},
		{"container:other", false},
		{"service:other", false},
		{"container:" + containerName + ":extra", false},
		{"", false},
	}
	for _, c := range cases {
		target := NetworkModeRefTarget(c.label)
		hit := target != "" && target == containerName
		if hit != c.wantHit {
			t.Errorf("label %q: hit=%v want %v (target=%q)", c.label, hit, c.wantHit, target)
		}
	}
}
