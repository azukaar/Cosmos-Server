package authorizationserver

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/azukaar/cosmos-server/src/utils"
)

// Discovery must derive every endpoint - and the issuer - from the host the
// client used, never from the configured hostname. Mixing the two produces a
// document whose authorization endpoint lies outside the issuer origin, which
// clients reject.
func TestDiscoverEndpointUsesRequestHost(t *testing.T) {
	previous := utils.GetMainConfig()
	previousHTTPS := utils.IsHTTPS
	t.Cleanup(func() {
		utils.LoadBaseMainConfig(previous)
		utils.IsHTTPS = previousHTTPS
	})

	cfg := previous
	cfg.NewInstall = false
	cfg.HTTPConfig.Hostname = "192.0.2.1"
	cfg.HTTPConfig.ProxyConfig.Routes = nil
	utils.LoadBaseMainConfig(cfg)
	utils.IsHTTPS = true

	req := httptest.NewRequest("GET", "https://idp.example.com/.well-known/openid-configuration", nil)
	req.Host = "idp.example.com"
	rw := httptest.NewRecorder()

	discoverEndpoint(rw, req)

	if rw.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(rw.Body.Bytes(), &doc); err != nil {
		t.Fatalf("failed to decode discovery doc: %v", err)
	}

	const origin = "https://idp.example.com"

	for _, key := range []string{
		"issuer",
		"authorization_endpoint",
		"token_endpoint",
		"jwks_uri",
		"userinfo_endpoint",
		"revocation_endpoint",
		"end_session_endpoint",
	} {
		got, _ := doc[key].(string)
		if !strings.HasPrefix(got, origin) {
			t.Errorf("%s = %q, want it under %q", key, got, origin)
		}
		if strings.Contains(got, cfg.HTTPConfig.Hostname) {
			t.Errorf("%s = %q, leaks the configured hostname", key, got)
		}
	}

	got, _ := doc["authorization_endpoint"].(string)
	if got != origin+"/cosmos-ui/openid" {
		t.Errorf("authorization_endpoint = %q, want %s/cosmos-ui/openid", got, origin)
	}
}
