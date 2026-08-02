package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSRedirectHandler(t *testing.T) {
	tests := []struct {
		name         string
		requestURL   string
		httpsPort    string
		wantLocation string
	}{
		{
			name:         "standard HTTPS port",
			requestURL:   "http://example.test/dashboard?tab=routes",
			httpsPort:    "443",
			wantLocation: "https://example.test/dashboard?tab=routes",
		},
		{
			name:         "strip HTTP port from hostname",
			requestURL:   "http://example.test:80/cosmos/api/status",
			httpsPort:    "443",
			wantLocation: "https://example.test/cosmos/api/status",
		},
		{
			name:         "configured nonstandard HTTPS port",
			requestURL:   "http://example.test:8080/files/report.txt?download=1",
			httpsPort:    "8443",
			wantLocation: "https://example.test:8443/files/report.txt?download=1",
		},
		{
			name:         "IPv6 hostname",
			requestURL:   "http://[2001:db8::1]/cosmos-ui/",
			httpsPort:    "443",
			wantLocation: "https://[2001:db8::1]/cosmos-ui/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.requestURL, nil)
			res := httptest.NewRecorder()

			httpsRedirectHandler(tt.httpsPort).ServeHTTP(res, req)

			if res.Code != http.StatusPermanentRedirect {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusPermanentRedirect)
			}
			if got := res.Header().Get("Location"); got != tt.wantLocation {
				t.Fatalf("Location = %q, want %q", got, tt.wantLocation)
			}
		})
	}
}
