package proxy

import (
	"testing"

	"github.com/azukaar/cosmos-server/src/utils"
)

func localClaim(port string, to string) PortClaim {
	return PortClaim{
		Pair:    PortsPair{From: port, To: to, route: utils.ProxyRouteConfig{Name: "local-" + port}},
		IsLocal: true,
	}
}

func tunnelClaim(port string, to string) PortClaim {
	return PortClaim{
		Pair:    PortsPair{From: port, To: to, route: utils.ProxyRouteConfig{Name: "tunnel-" + port, Tunnel: "_ANY_"}},
		IsLocal: false,
	}
}

func acceptedFor(t *testing.T, accepted []PortsPair, port string) PortsPair {
	t.Helper()
	for _, pair := range accepted {
		if pair.From == port {
			return pair
		}
	}
	t.Fatalf("port %s was not accepted at all; got %+v", port, accepted)
	return PortsPair{}
}

func TestResolvePortClaims_LocalBeatsTunnelWhicheverOrder(t *testing.T) {
	tests := []struct {
		name   string
		claims []PortClaim
	}{
		{"local first", []PortClaim{localClaim(":8603", "seaweed:8333"), tunnelClaim(":8603", "https://192.168.201.1")}},
		{"tunnel first", []PortClaim{tunnelClaim(":8603", "https://192.168.201.1"), localClaim(":8603", "seaweed:8333")}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accepted, shadowed, conflicts := ResolvePortClaims(test.claims)

			if len(accepted) != 1 {
				t.Fatalf("expected exactly one accepted claim, got %d", len(accepted))
			}
			if got := acceptedFor(t, accepted, ":8603").To; got != "seaweed:8333" {
				t.Fatalf("the local route lost the port: To=%q", got)
			}
			if len(conflicts) != 0 {
				t.Fatalf("a tunnel losing to a local route is not a conflict, got %v", conflicts)
			}
			if len(shadowed) != 1 || shadowed[0] != ":8603" {
				t.Fatalf("expected :8603 reported as shadowed, got %v", shadowed)
			}
		})
	}
}

// Only same-kind collisions are real misconfiguration.
func TestResolvePortClaims_ReportsGenuineConflicts(t *testing.T) {
	accepted, shadowed, conflicts := ResolvePortClaims([]PortClaim{
		localClaim(":8080", "a:80"),
		localClaim(":8080", "b:80"),
	})

	if len(accepted) != 1 || acceptedFor(t, accepted, ":8080").To != "a:80" {
		t.Fatalf("first local claim should win, got %+v", accepted)
	}
	if len(conflicts) != 1 || conflicts[0] != ":8080" {
		t.Fatalf("expected :8080 reported as a conflict, got %v", conflicts)
	}
	if len(shadowed) != 0 {
		t.Fatalf("two local routes are a conflict, not a shadow, got %v", shadowed)
	}
}

func TestResolvePortClaims_TwoTunnelsOnOnePortConflict(t *testing.T) {
	_, shadowed, conflicts := ResolvePortClaims([]PortClaim{
		tunnelClaim(":9000", "https://192.168.201.1"),
		tunnelClaim(":9000", "https://192.168.201.2"),
	})

	if len(conflicts) != 1 || conflicts[0] != ":9000" {
		t.Fatalf("expected :9000 reported as a conflict, got %v", conflicts)
	}
	if len(shadowed) != 0 {
		t.Fatalf("no local route was involved, got shadowed %v", shadowed)
	}
}

func TestResolvePortClaims_TunnelKeptWhenNoLocalClaim(t *testing.T) {
	accepted, shadowed, conflicts := ResolvePortClaims([]PortClaim{
		localClaim(":8603", "seaweed:8333"),
		tunnelClaim(":9000", "https://192.168.201.2"),
	})

	if len(accepted) != 2 {
		t.Fatalf("expected both ports accepted, got %+v", accepted)
	}
	if got := acceptedFor(t, accepted, ":9000").To; got != "https://192.168.201.2" {
		t.Fatalf("tunnel-only port lost its target: %q", got)
	}
	if len(shadowed) != 0 || len(conflicts) != 0 {
		t.Fatalf("distinct ports must not collide: shadowed=%v conflicts=%v", shadowed, conflicts)
	}
}

func TestResolvePortClaims_PreservesTunnelTargets(t *testing.T) {
	claim := tunnelClaim(":9000", "https://192.168.201.2")
	claim.Pair.Targets = []utils.TunnelTarget{
		{DeviceName: "node-0", TargetURL: "https://192.168.201.1"},
		{DeviceName: "node-1", TargetURL: "https://192.168.201.2"},
	}

	accepted, _, _ := ResolvePortClaims([]PortClaim{claim})

	if len(accepted) != 1 || len(accepted[0].Targets) != 2 {
		t.Fatalf("tunnel targets were dropped: %+v", accepted)
	}
}

func TestResolvePortClaims_EmptyInput(t *testing.T) {
	accepted, shadowed, conflicts := ResolvePortClaims(nil)

	if len(accepted) != 0 || len(shadowed) != 0 || len(conflicts) != 0 {
		t.Fatal("nothing in, nothing out")
	}
}
