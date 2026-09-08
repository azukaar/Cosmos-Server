package constellation

import (
	"testing"

	"github.com/azukaar/cosmos-server/src/utils"
)

func tunnelWithLoad(name string, cpu float64, ram float64) utils.ConstellationTunnel {
	return utils.ConstellationTunnel{
		Route: utils.ProxyRouteConfig{Name: name, Host: name + ".test", Tunnel: "_ANY_"},
		Targets: []utils.TunnelTarget{
			{DeviceName: "node-1", TargetURL: "https://192.168.201.2", CPUPercent: cpu, RAMPercent: ram, MonitoringOn: true},
			{DeviceName: "node-0", TargetURL: "https://192.168.201.1", CPUPercent: cpu / 2, RAMPercent: ram, MonitoringOn: true},
		},
	}
}

// A load fluctuation is not a topology change.
func TestSortTunnelsForComparison_IgnoresLoadSamples(t *testing.T) {
	before := []utils.ConstellationTunnel{tunnelWithLoad("registry-acc", 3.5, 41.2)}
	after := []utils.ConstellationTunnel{tunnelWithLoad("registry-acc", 87.9, 63.8)}

	if !utils.JSONEquals(sortTunnelsForComparison(before), sortTunnelsForComparison(after)) {
		t.Fatal("a CPU/RAM change was treated as a tunnel cache change")
	}
}

func TestSortTunnelsForComparison_DetectsRealChanges(t *testing.T) {
	base := []utils.ConstellationTunnel{tunnelWithLoad("registry-acc", 10, 20)}

	tests := []struct {
		name string
		next []utils.ConstellationTunnel
	}{
		{
			name: "a target appears",
			next: func() []utils.ConstellationTunnel {
				changed := tunnelWithLoad("registry-acc", 10, 20)
				changed.Targets = append(changed.Targets, utils.TunnelTarget{DeviceName: "node-2", TargetURL: "https://192.168.201.3"})
				return []utils.ConstellationTunnel{changed}
			}(),
		},
		{
			name: "a target moves address",
			next: func() []utils.ConstellationTunnel {
				changed := tunnelWithLoad("registry-acc", 10, 20)
				changed.Targets[0].TargetURL = "https://192.168.201.9"
				return []utils.ConstellationTunnel{changed}
			}(),
		},
		{
			name: "monitoring switched off",
			next: func() []utils.ConstellationTunnel {
				changed := tunnelWithLoad("registry-acc", 10, 20)
				changed.Targets[0].MonitoringOn = false
				return []utils.ConstellationTunnel{changed}
			}(),
		},
		{
			name: "the route itself changed",
			next: func() []utils.ConstellationTunnel {
				changed := tunnelWithLoad("registry-acc", 10, 20)
				changed.Route.Host = "elsewhere.test"
				return []utils.ConstellationTunnel{changed}
			}(),
		},
		{
			name: "a tunnel disappears",
			next: []utils.ConstellationTunnel{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if utils.JSONEquals(sortTunnelsForComparison(base), sortTunnelsForComparison(test.next)) {
				t.Fatal("a real change was missed")
			}
		})
	}
}

func TestSortTunnelsForComparison_OrderInsensitive(t *testing.T) {
	a := []utils.ConstellationTunnel{tunnelWithLoad("b-route", 1, 2), tunnelWithLoad("a-route", 3, 4)}

	reversed := tunnelWithLoad("a-route", 3, 4)
	reversed.Targets[0], reversed.Targets[1] = reversed.Targets[1], reversed.Targets[0]
	b := []utils.ConstellationTunnel{reversed, tunnelWithLoad("b-route", 1, 2)}

	if !utils.JSONEquals(sortTunnelsForComparison(a), sortTunnelsForComparison(b)) {
		t.Fatal("ordering was treated as a change")
	}
}

// Must not mutate the input: the live cache keeps the load samples.
func TestSortTunnelsForComparison_DoesNotMutateInput(t *testing.T) {
	original := []utils.ConstellationTunnel{tunnelWithLoad("registry-acc", 55.5, 66.6)}

	sortTunnelsForComparison(original)

	if original[0].Targets[0].CPUPercent != 55.5 || original[0].Targets[0].RAMPercent != 66.6 {
		t.Fatalf("input was mutated: %+v", original[0].Targets[0])
	}
}
