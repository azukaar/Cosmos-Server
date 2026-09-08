package constellation

import (
	"reflect"
	"testing"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

// tunnelAdvertiser builds a heartbeat advertising one route named "app" under
// the given device name, with a distinguishable config.
func tunnelAdvertiser(deviceName string, lbMode string, sticky bool, auth bool, timeout time.Duration) NodeHeartbeat {
	return NodeHeartbeat{
		DeviceName: deviceName,
		Tunnels: []utils.ProxyRouteConfig{
			{
				Name:         "app",
				Tunnel:       "_ANY_",
				Target:       "https://" + deviceName + ".local",
				LBMode:       lbMode,
				LBStickyMode: sticky,
				AuthEnabled:  auth,
				Timeout:      timeout,
			},
		},
	}
}

func findTunnel(tunnels []utils.ConstellationTunnel, name string) *utils.ConstellationTunnel {
	for i := range tunnels {
		if tunnels[i].Route.Name == name {
			return &tunnels[i]
		}
	}
	return nil
}

// The governing Route must be the lowest advertiser's, whatever the heartbeat order.
func TestUnitMergeTunnelHeartbeatsIsOrderIndependent(t *testing.T) {
	// plain string order: '-' (0x2D) < '_' (0x5F), so "node-a" beats every "node_*"
	heartbeats := []NodeHeartbeat{
		tunnelAdvertiser("node_b", "round_robin", true, true, 30*time.Second),
		tunnelAdvertiser("node-a", "", false, false, 10*time.Second),
		tunnelAdvertiser("node_c", "least_conn", true, false, 20*time.Second),
		tunnelAdvertiser("node_d", "round_robin", false, true, 40*time.Second),
		tunnelAdvertiser("node_e", "", true, true, 50*time.Second),
	}

	orders := map[string][]int{
		"as-listed":     {0, 1, 2, 3, 4},
		"reversed":      {4, 3, 2, 1, 0},
		"winner-last":   {0, 2, 3, 4, 1},
		"winner-middle": {2, 4, 1, 0, 3},
		"winner-first":  {1, 3, 0, 4, 2},
	}

	want := heartbeats[1].Tunnels[0]
	want.Const_IsTunneled = true

	for name, order := range orders {
		t.Run(name, func(t *testing.T) {
			permuted := make([]NodeHeartbeat, 0, len(order))
			for _, i := range order {
				permuted = append(permuted, heartbeats[i])
			}

			tunnels := mergeTunnelHeartbeats(permuted, "node_c")
			if len(tunnels) != 1 {
				t.Fatalf("mergeTunnelHeartbeats() returned %d tunnels, want 1", len(tunnels))
			}

			got := tunnels[0]
			if !reflect.DeepEqual(got.Route, want) {
				t.Errorf("governing Route = %+v, want node-a's %+v", got.Route, want)
			}
			if len(got.Targets) != len(heartbeats) {
				t.Errorf("got %d targets, want %d", len(got.Targets), len(heartbeats))
			}
			if got.Targets[0].DeviceName != "node_c" {
				t.Errorf("local node not first in targets: got %q", got.Targets[0].DeviceName)
			}
		})
	}
}

func TestUnitMergeTunnelHeartbeatsFiltersAndGroups(t *testing.T) {
	heartbeats := []NodeHeartbeat{
		{
			DeviceName: "node_b",
			Tunnels: []utils.ProxyRouteConfig{
				{Name: "app", Tunnel: "_ANY_", Target: "https://b.local", LBMode: "round_robin"},
				{Name: "other", Tunnel: "node_z", Target: "https://b.local/other"},
				{Name: "mine", Tunnel: "node_c", Target: "https://b.local/mine", LBMode: "least_conn"},
			},
		},
		{
			DeviceName: "node_a",
			Tunnels: []utils.ProxyRouteConfig{
				{Name: "app", Tunnel: "_ANY_", Target: "https://a.local"},
			},
		},
	}

	tunnels := mergeTunnelHeartbeats(heartbeats, "node_c")
	if len(tunnels) != 2 {
		t.Fatalf("got %d tunnels, want 2 (route tunneled to another device must be dropped)", len(tunnels))
	}

	app := findTunnel(tunnels, "app")
	if app == nil {
		t.Fatal("route 'app' missing from merge")
	}
	if app.Route.LBMode != "" || app.Route.Target != "https://a.local" {
		t.Errorf("route 'app' governed by %q (LBMode %q), want node_a's config", app.Route.Target, app.Route.LBMode)
	}
	if !app.Route.Const_IsTunneled {
		t.Error("route 'app' lost Const_IsTunneled")
	}
	if len(app.Targets) != 2 {
		t.Errorf("route 'app' has %d targets, want 2", len(app.Targets))
	}

	mine := findTunnel(tunnels, "mine")
	if mine == nil {
		t.Fatal("route explicitly tunneled to this device missing from merge")
	}
	if !mine.Route.Const_IsTunneled {
		t.Error("route 'mine' lost Const_IsTunneled")
	}
}

// Each target must carry its own advertiser's resource sample — the
// load_based LB mode is blind without them.
func TestUnitMergeTunnelHeartbeatsCarriesLoadMetrics(t *testing.T) {
	hbA := tunnelAdvertiser("node_a", "load_based", false, false, 10*time.Second)
	hbA.CPUPercent, hbA.RAMPercent, hbA.MonitoringOn = 20, 70, true
	hbB := tunnelAdvertiser("node_b", "load_based", false, false, 10*time.Second)

	tunnels := mergeTunnelHeartbeats([]NodeHeartbeat{hbA, hbB}, "node_c")
	if len(tunnels) != 1 {
		t.Fatalf("got %d tunnels, want 1", len(tunnels))
	}

	for _, target := range tunnels[0].Targets {
		switch target.DeviceName {
		case "node_a":
			if target.CPUPercent != 20 || target.RAMPercent != 70 || !target.MonitoringOn {
				t.Errorf("node_a target = %+v, want cpu 20 / ram 70 / monitored", target)
			}
		case "node_b":
			if target.MonitoringOn {
				t.Errorf("node_b target = %+v, want MonitoringOn false", target)
			}
		}
	}
}

// A heartbeat with no device name must never govern, and must not stop a named
// advertiser from governing either.
func TestUnitMergeTunnelHeartbeatsIgnoresUnnamedAdvertiser(t *testing.T) {
	heartbeats := []NodeHeartbeat{
		tunnelAdvertiser("", "chaos", true, true, time.Second),
		tunnelAdvertiser("node_a", "round_robin", false, false, 10*time.Second),
	}

	for _, order := range [][]NodeHeartbeat{heartbeats, {heartbeats[1], heartbeats[0]}} {
		tunnels := mergeTunnelHeartbeats(order, "node_c")
		if len(tunnels) != 1 {
			t.Fatalf("got %d tunnels, want 1", len(tunnels))
		}
		if tunnels[0].Route.LBMode != "round_robin" {
			t.Errorf("governing LBMode = %q, want round_robin (unnamed advertiser must not govern)", tunnels[0].Route.LBMode)
		}
	}
}

// Serving tunnels is LB-only — even _ANY_ routes must not land on a
// non-load-balancer, while an LB merges them as before.
func TestUnitTunnelsForNodeGatesOnLoadBalancer(t *testing.T) {
	setupTestEnv(t, func(cfg *utils.Config) {
		cfg.ConstellationConfig.ThisDeviceName = "node-a"
	})
	heartbeats := []NodeHeartbeat{
		tunnelAdvertiser("node_b", "round_robin", false, false, 10*time.Second),
	}

	seedDeviceCache(t, utils.ConstellationDevice{DeviceName: "node-a", IP: "192.168.201.5"})
	if got := tunnelsForNode(heartbeats, "node-a"); len(got) != 0 {
		t.Errorf("non-LB node serves %d tunnels, want 0", len(got))
	}

	seedDeviceCache(t, utils.ConstellationDevice{DeviceName: "node-a", IP: "192.168.201.5", IsLoadBalancer: true})
	if got := tunnelsForNode(heartbeats, "node-a"); len(got) != 1 {
		t.Errorf("LB node serves %d tunnels, want 1", len(got))
	}
}
