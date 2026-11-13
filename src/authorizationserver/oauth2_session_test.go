package authorizationserver

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/azukaar/cosmos-server/src/utils"
)

// The ID token issuer is derived from the request host. A host is accepted when
// it is the configured hostname or serves a proxy route, and its scheme follows
// IsHTTPS. The container hostname (used elsewhere as a Docker container id) is
// not a hostname Cosmos serves, so it must not be treated as one here.
func TestNewSessionIssuer(t *testing.T) {
	previous := utils.GetMainConfig()
	previousHTTPS := utils.IsHTTPS
	previousContainer := os.Getenv("HOSTNAME")
	t.Cleanup(func() {
		utils.LoadBaseMainConfig(previous)
		utils.IsHTTPS = previousHTTPS
		os.Setenv("HOSTNAME", previousContainer)
	})

	cfg := previous
	cfg.NewInstall = false
	cfg.HTTPConfig.Hostname = "cosmos.example"
	cfg.HTTPConfig.HTTPSPort = "443"
	cfg.HTTPConfig.HTTPPort = "80"
	cfg.HTTPConfig.ProxyConfig.Routes = []utils.ProxyRouteConfig{
		{Name: "app", Host: "app.example", UseHost: true, AuthEnabled: true},
	}
	utils.LoadBaseMainConfig(cfg)
	utils.IsHTTPS = true
	os.Setenv("HOSTNAME", "cosmos-container")

	for _, c := range []struct {
		name string
		host string
		want string
	}{
		{"configured hostname", "cosmos.example", "https://cosmos.example"},
		{"proxy route host", "app.example", "https://app.example"},
		{"configured hostname on https port", "cosmos.example:443", "https://cosmos.example"},
		{"proxy route on https port", "app.example:443", "https://app.example"},
		{"container hostname is not a served hostname", "cosmos-container", ""},
		{"unknown host", "evil.example", ""},
	} {
		req := httptest.NewRequest("GET", "http://"+c.host+"/oauth2/auth", nil)
		req.Host = c.host

		session := newSession("user", req)

		if c.want == "" {
			if session != nil {
				t.Errorf("%s: expected rejection, got issuer %q", c.name, session.Claims.Issuer)
			}
			continue
		}
		if session == nil {
			t.Errorf("%s: expected issuer %q, got nil session", c.name, c.want)
			continue
		}
		if got := session.Claims.Issuer; got != c.want {
			t.Errorf("%s: issuer = %q, want %q", c.name, got, c.want)
		}
	}
}
