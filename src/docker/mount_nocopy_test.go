package docker

import (
	"encoding/json"
	"testing"
)

func TestNoCopyGating(t *testing.T) {
	cases := []struct {
		name           string
		serverVersion  string
		explicitNoCopy bool
		wantNoCopy     bool
	}{
		{"engine 29.6.1 no explicit", "29.6.1", false, false},
		{"engine 29.5.0 no explicit", "29.5.0", false, false},
		{"engine 29.4.3 no explicit (workaround on)", "29.4.3", false, true},
		{"engine 26.0.0 no explicit (workaround on)", "26.0.0", false, true},
		{"engine 29.6.1 explicit true", "29.6.1", true, true},
		{"engine 29.4.3 explicit true", "29.4.3", true, true},
	}
	for _, c := range cases {
		old := dockerServerVersion
		dockerServerVersion = c.serverVersion
		cm := CosmosMount{Type: "volume", Source: "v", Target: "/f", SubPath: "a/b.conf"}
		if c.explicitNoCopy {
			cm.NoCopy = true
		}
		dm := cm.ToDockerMount()
		got := dm.VolumeOptions != nil && dm.VolumeOptions.NoCopy
		dockerServerVersion = old
		if got != c.wantNoCopy {
			t.Errorf("%s: NoCopy=%v want %v", c.name, got, c.wantNoCopy)
		}
	}
}

func TestNoCopyJSONRoundTrip(t *testing.T) {
	cm := CosmosMount{Type: "volume", Source: "v", Target: "/f", SubPath: "a/b.conf", NoCopy: true}
	dm := cm.ToDockerMount()
	b, _ := json.Marshal(dm)
	if !strContains(string(b), `"NoCopy":true`) {
		t.Errorf("ToDockerMount JSON missing NoCopy: %s", b)
	}
	back := FromDockerMount(dm)
	if !back.NoCopy {
		t.Errorf("NoCopy lost in FromDockerMount round-trip")
	}
}

func TestDockerVersionAtLeast(t *testing.T) {
	cases := []struct{ ver, min string; want bool }{
		{"29.6.1", "29.5.0", true},
		{"29.5.0", "29.5.0", true},
		{"29.5.1", "29.5.0", true},
		{"29.4.3", "29.5.0", false},
		{"26.0.0", "29.5.0", false},
		{"28.0.0", "29.5.0", false},
		{"30.0.0", "29.5.0", true},
	}
	for _, c := range cases {
		if got := dockerVersionAtLeast(c.ver, c.min); got != c.want {
			t.Errorf("dockerVersionAtLeast(%q,%q)=%v want %v", c.ver, c.min, got, c.want)
		}
	}
}

func strContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
