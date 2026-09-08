package main

import (
	"testing"

	"github.com/azukaar/cosmos-server/src/utils"
)

func setFaviconTestConfig(t *testing.T) {
	t.Helper()
	prev := utils.GetBaseMainConfig()
	prevHTTPS := utils.IsHTTPS
	t.Cleanup(func() {
		utils.LoadBaseMainConfig(prev)
		utils.IsHTTPS = prevHTTPS
	})
	utils.IsHTTPS = true
	utils.LoadBaseMainConfig(utils.Config{
		HTTPConfig: utils.HTTPConfig{
			Hostname:  "cosmos.example.com",
			HTTPSPort: "443",
			HTTPPort:  "80",
			ProxyConfig: utils.ProxyConfig{
				Routes: []utils.ProxyRouteConfig{
					{Name: "app", Mode: "SERVAPP", Target: "http://myapp:8080"},
					{Name: "ext", Mode: "PROXY", Target: "https://example.org/"},
					{Name: "files", Mode: "STATIC", Target: "/data"},
					{Name: "spa-host", Mode: "SPA", Target: "/data", UseHost: true, Host: "spa.example.com"},
					{Name: "spa-path", Mode: "SPA", Target: "/data", UsePathPrefix: true, PathPrefix: "/spa"},
				},
			},
		},
		OpenIDClients: []utils.OpenIDClient{
			{ID: "grafana", Redirect: "https://grafana.example.com/login/generic_oauth, https://alt.example.com/cb"},
			{ID: "broken", Redirect: "not a url"},
		},
	})
}

func TestUnitFaviconTargetResolvesOnlyConfiguredEntries(t *testing.T) {
	setFaviconTestConfig(t)

	cases := []struct {
		name      string
		route     string
		client    string
		wantURL   string
		wantCont  string
		wantError bool
	}{
		{"servapp route dials the container", "app", "", "http://myapp:8080", "myapp", false},
		{"proxy route fetches its target", "ext", "", "https://example.org/", "", false},
		{"static route has no icon", "files", "", "", "", false},
		{"spa on its own host", "spa-host", "", "https://spa.example.com", "", false},
		{"spa under a path prefix", "spa-path", "", "https://cosmos.example.com/spa", "", false},
		{"openid client uses its first redirect origin", "", "grafana", "https://grafana.example.com", "", false},
		{"unknown route is refused", "nope", "", "", "", true},
		{"unknown client is refused", "", "nope", "", "", true},
		{"unparseable redirect is refused", "", "broken", "", "", true},
		{"a raw URL is not a route", "http://169.254.169.254/", "", "", "", true},
		{"nothing named is refused", "", "", "", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url, container, err := faviconTarget(tc.route, tc.client)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected an error, got url=%q container=%q", url, container)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if url != tc.wantURL || container != tc.wantCont {
				t.Fatalf("got url=%q container=%q, want url=%q container=%q", url, container, tc.wantURL, tc.wantCont)
			}
		})
	}
}

func TestUnitFaviconDialURLSwapsNameForIPOffBridge(t *testing.T) {
	prevInside, prevHost := utils.IsInsideContainer, utils.IsHostNetwork
	prevLookup := utils.GetContainerIPByName
	t.Cleanup(func() {
		utils.IsInsideContainer, utils.IsHostNetwork = prevInside, prevHost
		utils.GetContainerIPByName = prevLookup
	})
	utils.GetContainerIPByName = func(name string) (string, error) {
		if name == "myapp" {
			return "172.18.0.5", nil
		}
		return "", nil
	}

	utils.IsInsideContainer, utils.IsHostNetwork = true, false
	if got := faviconDialURL("http://myapp:8080", "myapp"); got != "http://myapp:8080" {
		t.Fatalf("on a bridge network docker DNS resolves the name, got %q", got)
	}

	utils.IsInsideContainer, utils.IsHostNetwork = false, false
	if got := faviconDialURL("http://myapp:8080/x", "myapp"); got != "http://172.18.0.5:8080/x" {
		t.Fatalf("on the host the name must be swapped for the IP, got %q", got)
	}

	utils.IsInsideContainer, utils.IsHostNetwork = true, true
	if got := faviconDialURL("http://myapp:8080", "myapp"); got != "http://172.18.0.5:8080" {
		t.Fatalf("with host networking the name must be swapped for the IP, got %q", got)
	}

	if got := faviconDialURL("http://unknown:80", "unknown"); got != "http://unknown:80" {
		t.Fatalf("an unresolvable container keeps its name, got %q", got)
	}

	if got := faviconDialURL("https://example.org/", ""); got != "https://example.org/" {
		t.Fatalf("a non-container target is untouched, got %q", got)
	}
}
