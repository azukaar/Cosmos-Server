package docker

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Ensure the healthcheck test field accepts all docker-compose forms:
// exec-form arrays and shell-form strings (which are wrapped in CMD-SHELL).
// Regression test for azukaar/Cosmos-Server#386:
// "json: cannot unmarshal string into Go struct field
// ContainerCreateRequestContainerHealthcheck.services-healthcheck.test of type []string"
func TestHealthcheckTestUnmarshal(t *testing.T) {
	cases := []struct {
		name  string
		json  string
		want  []string
	}{
		{
			name: "exec form array",
			json: `{"test": ["CMD", "curl", "-f", "http://localhost"]}`,
			want: []string{"CMD", "curl", "-f", "http://localhost"},
		},
		{
			name: "shell form string wraps in CMD-SHELL",
			json: `{"test": "curl -f http://localhost || exit 1"}`,
			want: []string{"CMD-SHELL", "curl -f http://localhost || exit 1"},
		},
		{
			name: "CMD-SHELL string wraps once",
			json: `{"test": "CMD-SHELL echo hi"}`,
			want: []string{"CMD-SHELL", "CMD-SHELL echo hi"},
		},
		{
			name: "empty string test becomes nil",
			json: `{"test": ""}`,
			want: nil,
		},
		{
			name: "missing test stays nil",
			json: `{"interval": "30s"}`,
			want: nil,
		},
		{
			name: "null test stays nil",
			json: `{"test": null}`,
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var h ContainerCreateRequestContainerHealthcheck
			if err := json.Unmarshal([]byte(tc.json), &h); err != nil {
				t.Fatalf("unexpected error unmarshaling %s: %v", tc.json, err)
			}
			if !reflect.DeepEqual(h.Test, tc.want) {
				t.Errorf("Test = %#v, want %#v", h.Test, tc.want)
			}
		})
	}
}

func TestHealthcheckTestUnmarshalInvalid(t *testing.T) {
	var h ContainerCreateRequestContainerHealthcheck
	err := json.Unmarshal([]byte(`{"test": 123}`), &h)
	if err == nil {
		t.Fatal("expected error for numeric test, got nil")
	}
}

// Marshal still emits the canonical array form.
func TestHealthcheckTestMarshal(t *testing.T) {
	h := ContainerCreateRequestContainerHealthcheck{
		Test: []string{"CMD-SHELL", "curl -f http://localhost || exit 1"},
	}
	out, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	want := `{"test":["CMD-SHELL","curl -f http://localhost || exit 1"]}`
	if string(out) != want {
		t.Errorf("marshal = %s, want %s", out, want)
	}
}