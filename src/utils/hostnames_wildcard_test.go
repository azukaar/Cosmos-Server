package utils

import (
	"reflect"
	"testing"
)

// withHostnamesConfig installs a config for GetAllHostnames and restores the
// previous one (and the tunnel-route hook) when the test ends.
func withHostnamesConfig(t *testing.T, hostname string, routes []ProxyRouteConfig, tunnels []ProxyRouteConfig) {
	t.Helper()
	previous := GetMainConfig()
	previousTunnels := GetConstellationTunnelRoutes
	t.Cleanup(func() {
		LoadBaseMainConfig(previous)
		GetConstellationTunnelRoutes = previousTunnels
	})

	config := previous
	config.NewInstall = false
	config.HTTPConfig.Hostname = hostname
	config.HTTPConfig.ProxyConfig.Routes = routes
	LoadBaseMainConfig(config)
	GetConstellationTunnelRoutes = func() []ProxyRouteConfig { return tunnels }
}

func host(h string) ProxyRouteConfig {
	return ProxyRouteConfig{Name: h, UseHost: true, Host: h}
}

func TestGetAllHostnamesCollectsRoutes(t *testing.T) {
	withHostnamesConfig(t, "cluster.example.com:8443",
		[]ProxyRouteConfig{
			host("app.example.com:8080"),
			host("App.example.com:8080"), // not deduped: hostnames are kept as written
			{Name: "path-only", UseHost: false, Host: "ignored.example.com"},
			{Name: "empty", UseHost: true, Host: ""},
			{Name: "list", UseHost: true, Host: "a.example.com,b.example.com"},
			{Name: "spaced", UseHost: true, Host: "a.example.com b.example.com"},
			host("app.example.com:8080"), // exact duplicate
			host("cluster.example.com:8443"),
		},
		[]ProxyRouteConfig{host("tunnel.example.com")},
	)

	for name, tc := range map[string]struct {
		removePorts bool
		want        []string
	}{
		"with ports":    {false, []string{"cluster.example.com:8443", "app.example.com:8080", "App.example.com:8080", "tunnel.example.com"}},
		"without ports": {true, []string{"cluster.example.com", "app.example.com", "App.example.com", "tunnel.example.com"}},
	} {
		if got := GetAllHostnames(tc.removePorts); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v", name, got, tc.want)
		}
	}
}

// a route on ":<port>" or "0.0.0.0:<port>" answers any Host on that port. It
// has no name to certify or publish (Let's Encrypt refuses an empty
// identifier), but EnsureHostname needs the raw entry to let requests through.
func TestGetAllHostnamesPortOnlyRoutes(t *testing.T) {
	withHostnamesConfig(t, "cluster.example.com",
		[]ProxyRouteConfig{host(":8603"), host("0.0.0.0:9000"), host("app.example.com:8080")}, nil)

	want := []string{"cluster.example.com", "app.example.com"}
	if got := GetAllHostnames(true); !reflect.DeepEqual(got, want) {
		t.Errorf("without ports: got %v, want %v", got, want)
	}
	want = []string{"cluster.example.com", ":8603", "0.0.0.0:9000", "app.example.com:8080"}
	if got := GetAllHostnames(false); !reflect.DeepEqual(got, want) {
		t.Errorf("with ports: got %v, want %v", got, want)
	}
}
