package utils

import "testing"

// stubLocalDocker replaces the docker seam for one test.
func stubLocalDocker(t *testing.T, subnet func(string) bool) {
	t.Helper()
	prev := IsLocalDockerIP
	IsLocalDockerIP = subnet
	t.Cleanup(func() { IsLocalDockerIP = prev })
}

func stubContainerMode(t *testing.T, inside bool, hostNet bool) {
	t.Helper()
	prevInside, prevHost := IsInsideContainer, IsHostNetwork
	IsInsideContainer, IsHostNetwork = inside, hostNet
	t.Cleanup(func() { IsInsideContainer, IsHostNetwork = prevInside, prevHost })
}

func TestUnitIsLocalPeer(t *testing.T) {
	onBridge := func(ip string) bool { return ip == "172.17.0.7" || ip == "172.16.0.34" }

	tests := []struct {
		name    string
		ip      string
		inside  bool
		hostNet bool
		want    bool
	}{
		{"loopback v4", "127.0.0.1", false, false, true},
		{"loopback v6", "::1", false, false, true},
		{"default bridge", "172.17.0.7", false, false, true},
		{"user defined network", "172.16.0.34", false, false, true},
		{"a nebula peer is not local", "192.168.201.2", false, false, false},
		{"a public client is not local", "51.15.7.9", false, false, false},
		// A LAN address must NOT pass: it is private but it is not ours.
		{"lan address", "192.168.1.50", false, false, false},
		{"garbage", "not-an-ip", false, false, false},
		// Inside a bridge container docker's userland proxy can present remote clients as the bridge gateway.
		{"bridge container distrusts docker source", "172.17.0.7", true, false, false},
		{"bridge container still trusts loopback", "127.0.0.1", true, false, true},
		{"host network container trusts docker source", "172.17.0.7", true, true, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stubLocalDocker(t, onBridge)
			stubContainerMode(t, test.inside, test.hostNet)
			if got := IsLocalPeer(test.ip); got != test.want {
				t.Errorf("IsLocalPeer(%q) = %v, want %v", test.ip, got, test.want)
			}
		})
	}
}

// A restricted route must admit a container on its own host, while an API token restricted the same way must not.
func TestUnitCheckIPAccessLocalPeer(t *testing.T) {
	stubLocalDocker(t, func(ip string) bool { return ip == "172.17.0.7" })
	stubContainerMode(t, false, false)

	prev := IsConstellationIP
	IsConstellationIP = func(ip string) bool { return ip == "192.168.201.2" }
	t.Cleanup(func() { IsConstellationIP = prev })

	tests := []struct {
		name      string
		ip        string
		whitelist []string
		route     bool // CheckRouteIPAccess vs CheckIPAccess
		want      bool
	}{
		{"route: local container passes", "172.17.0.7", nil, true, true},
		{"route: loopback passes", "127.0.0.1", nil, true, true},
		{"route: constellation peer passes", "192.168.201.2", nil, true, true},
		{"route: outsider blocked", "51.15.7.9", nil, true, false},
		{"token: local container blocked", "172.17.0.7", nil, false, false},
		{"token: loopback blocked", "127.0.0.1", nil, false, false},
		{"token: constellation peer passes", "192.168.201.2", nil, false, true},
		// A whitelist adds to the constellation set rather than replacing it.
		{"route: local pass survives an added whitelist", "172.17.0.7", []string{"10.0.0.1"}, true, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got bool
			if test.route {
				got = CheckRouteIPAccess(test.ip, test.ip, true, test.whitelist)
			} else {
				got = CheckIPAccess(test.ip, test.ip, true, test.whitelist)
			}
			if got != test.want {
				t.Errorf("access(%q) = %v, want %v", test.ip, got, test.want)
			}
		})
	}
}

// A whitelist-only route stays strict: being local is not a substitute for being listed.
func TestUnitCheckRouteIPAccessWhitelistOnlyStaysStrict(t *testing.T) {
	stubLocalDocker(t, func(ip string) bool { return ip == "172.17.0.7" })
	stubContainerMode(t, false, false)

	if CheckRouteIPAccess("172.17.0.7", "172.17.0.7", false, []string{"10.0.0.1"}) {
		t.Error("a local peer satisfied a whitelist it is not on")
	}
	if !CheckRouteIPAccess("10.0.0.1", "10.0.0.1", false, []string{"10.0.0.1"}) {
		t.Error("a whitelisted peer was blocked")
	}
}
