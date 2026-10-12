package utils

import (
	"net/http/httptest"
	"testing"
)

func TestIsStreamingEndpoint(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/cosmos/api/docker-service", true},
		{"/cosmos/api/docker-service/extra", true},
		{"/cosmos/api/servapps/myapp/manage/update", true},
		{"/cosmos/api/servapps/myapp/manage/recreate", true},
		{"/cosmos/api/servapps/myapp/update", true},
		{"/cosmos/api/images/pull", true},
		{"/cosmos/api/images/pull-if-missing", true},
		// Non-streaming endpoints keep the short timeout.
		{"/cosmos/api/servapps", false},
		{"/cosmos/api/servapps/myapp", false},
		{"/cosmos/api/servapps/myapp/logs", false},
		{"/cosmos/api/servapps/myapp/terminal/attach", false},
		{"/cosmos/api/config", false},
		{"/cosmos/api/markets", false},
		{"/cosmos/api/users", false},
	}
	for _, c := range cases {
		req := httptest.NewRequest("POST", c.path, nil)
		if got := IsStreamingEndpoint(req); got != c.want {
			t.Errorf("IsStreamingEndpoint(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}