package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"reflect"
	"testing"
	"time"
)

func TestNeedsLetsEncrypt(t *testing.T) {
	le, off := HTTPSCertModeList["LETSENCRYPT"], HTTPSCertModeList["DISABLED"]
	dnsZone := []DNSZoneConfig{{Zone: "example.com", HTTPSCertificateMode: le, DNSChallengeProvider: "cloudflare"}}
	selfSigned := []DNSZoneConfig{{Zone: "example.com", HTTPSCertificateMode: HTTPSCertModeList["SELFSIGNED"]}}

	for name, tc := range map[string]struct {
		hostname string
		mode     string
		zones    []DNSZoneConfig
		issuer   bool
		want     bool
	}{
		"ip only":              {"192.168.1.10", le, nil, true, false},
		"local name":           {"nas.local", le, nil, true, false},
		"public name":          {"cosmos.example.com", le, nil, true, true},
		"https off":            {"cosmos.example.com", off, nil, true, false},
		"self-signed zone":     {"cosmos.example.com", le, selfSigned, true, false},
		"dns zone":             {"cosmos.example.com", le, dnsZone, true, true},
		"dns zone, not issuer": {"cosmos.example.com", le, dnsZone, false, false},
		"dns zone, https off":  {"cosmos.example.com", off, dnsZone, true, false},
	} {
		t.Run(name, func(t *testing.T) {
			withHostnamesConfig(t, tc.hostname, nil, nil)
			previous := IsZoneIssuer
			t.Cleanup(func() { IsZoneIssuer = previous })
			issuer := tc.issuer
			IsZoneIssuer = func() bool { return issuer }

			config := GetMainConfig()
			config.HTTPConfig.HTTPSCertificateMode = tc.mode
			config.HTTPConfig.DNSZones = tc.zones
			if got := NeedsLetsEncrypt(config); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHostCertificatePEM(t *testing.T) {
	config := Config{}
	config.HTTPConfig.HTTPSCertificateMode = HTTPSCertModeList["LETSENCRYPT"]
	config.HTTPConfig.SelfTLSCert = "self"
	config.HTTPConfig.TLSCert = "shared"
	config.HTTPConfig.DNSZones = []DNSZoneConfig{{Zone: "example.com", HTTPSCertificateMode: HTTPSCertModeList["LETSENCRYPT"], DNSChallengeProvider: "cloudflare"}}
	config.HTTPConfig.ZoneCerts = map[string]ZoneCert{"example.com": {TLSCert: "zone", Hosts: []string{"*.example.com"}}}
	config.HTTPConfig.LocalCerts = map[string]ZoneCert{"other.org": {TLSCert: "local", Hosts: []string{"other.org"}}}

	for host, want := range map[string]string{"app.example.com": "zone", "other.org": "local", "192.168.1.10": "self", "new.net": "shared"} {
		if got := GetHostCertificate(config, host).PEM; got != want {
			t.Errorf("%s: got %q, want %q", host, got, want)
		}
	}
}

func testCertificate(t *testing.T, names ...string) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: names[0]},
		DNSNames:     names,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func zoneNames(zones []ResolvedZone) []string {
	names := []string{}
	for _, z := range zones {
		name := z.Zone
		if z.Derived {
			name += "(derived)"
		}
		names = append(names, name)
	}
	return names
}

func TestFindZoneLongestSuffix(t *testing.T) {
	zones := []DNSZoneConfig{{Zone: "domain.com"}, {Zone: "lab.domain.com"}, {Zone: "other.co.uk"}}

	cases := map[string]int{
		"domain.com":          0,
		"app.domain.com":      0,
		"app.lab.domain.com":  1,
		"lab.domain.com":      1,
		"*.lab.domain.com":    1,
		"APP.Domain.com:443":  0,
		"x.other.co.uk":       2,
		"notdomain.com":       -1,
		"domain.com.evil.net": -1,
		"192.168.1.10":        -1,
	}
	for host, want := range cases {
		if got := FindZone(zones, host); got != want {
			t.Errorf("FindZone(%q) = %d, want %d", host, got, want)
		}
	}
}

func TestResolveZonesDerivedPerHost(t *testing.T) {
	config := Config{}
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	hosts := []string{"app1.domain.com", "app2.domain.com", "app1.domain.com:443", "localhost", "10.0.0.1", "nas"}

	got := ResolveZones(config, hosts)
	want := []string{"app1.domain.com(derived)", "app2.domain.com(derived)"}
	if !reflect.DeepEqual(zoneNames(got), want) {
		t.Fatalf("zones = %v, want %v", zoneNames(got), want)
	}
	if got[0].HTTPSCertificateMode != "LETSENCRYPT" {
		t.Errorf("a public hostname gets Let's Encrypt on its own, got %q", got[0].HTTPSCertificateMode)
	}

	// what Let's Encrypt cannot certify is self-signed, and a server without HTTPS has none
	got = ResolveZones(config, []string{"nas.local"})
	if got[0].HTTPSCertificateMode != "SELFSIGNED" {
		t.Errorf("nas.local: got %q, want SELFSIGNED", got[0].HTTPSCertificateMode)
	}
	config.HTTPConfig.HTTPSCertificateMode = "DISABLED"
	got = ResolveZones(config, []string{"app1.domain.com"})
	if got[0].HTTPSCertificateMode != "DISABLED" {
		t.Errorf("HTTP-only server: got %q, want DISABLED", got[0].HTTPSCertificateMode)
	}
}

// the zone of a hostname decides how it is served, whatever the server was set to
func TestHostHTTPSMode(t *testing.T) {
	config := Config{}
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	config.HTTPConfig.DNSZones = []DNSZoneConfig{
		{Zone: "domain.com", HTTPSCertificateMode: "LETSENCRYPT", DNSChallengeProvider: "cloudflare"},
		{Zone: "lan.domain.com", HTTPSCertificateMode: "SELFSIGNED"},
		{Zone: "own.org", HTTPSCertificateMode: "PROVIDED"},
		{Zone: "plain.net", HTTPSCertificateMode: "DISABLED"},
		{Zone: "unset.io"},
	}

	cases := map[string]string{
		"app.domain.com":        "LETSENCRYPT",
		"nas.lan.domain.com":    "SELFSIGNED",
		"www.own.org":           "PROVIDED",
		"app.plain.net:8080":    "DISABLED",
		"x.unset.io":            "LETSENCRYPT",
		"automatic.example.com": "LETSENCRYPT",
		"nas.local":             "SELFSIGNED",
		"192.168.1.10":          "SELFSIGNED",
		"localhost":             "SELFSIGNED",
	}
	for host, want := range cases {
		if got, _ := HostHTTPSMode(config, host); got != want {
			t.Errorf("%s: got %s, want %s", host, got, want)
		}
	}
	if !HostIsHTTPOnly(config, "app.plain.net") || HostIsHTTPOnly(config, "app.domain.com") {
		t.Error("only the hostnames of plain.net are HTTP-only")
	}

	// a server-wide mode from before zones never overrides a zone: this is what
	// kept a self-signed leader from ever serving, or issuing, its cluster's certificate
	for _, legacy := range []string{"SELFSIGNED", "PROVIDED"} {
		config.HTTPConfig.HTTPSCertificateMode = legacy
		if got, _ := HostHTTPSMode(config, "app.domain.com"); got != "LETSENCRYPT" {
			t.Errorf("server %s: the zone must decide, got %s", legacy, got)
		}
		if got, _ := HostHTTPSMode(config, "automatic.example.com"); got != legacy {
			t.Errorf("server %s: a hostname without a zone keeps it until migrated, got %s", legacy, got)
		}
	}

	// except for not serving HTTPS at all
	for _, off := range []string{"DISABLED", ""} {
		config.HTTPConfig.HTTPSCertificateMode = off
		for host := range cases {
			if !HostIsHTTPOnly(config, host) {
				t.Errorf("server %q: %s must be HTTP-only", off, host)
			}
		}
	}
}

func hostNames(config Config) map[string]string {
	out := map[string]string{}
	for _, zone := range config.HTTPConfig.DNSZones {
		out[zone.Zone] = zone.HTTPSCertificateMode
	}
	return out
}

// a server-wide self-signed mode becomes zones: the registrable domains of the
// hostnames it was serving, so the ones added later are served the same way
func TestMigrateServerModeSelfSigned(t *testing.T) {
	config := Config{}
	config.HTTPConfig.HTTPSCertificateMode = "SELFSIGNED"
	config.HTTPConfig.Hostname = "cosmos.example.com:8443"
	config.HTTPConfig.TLSCert = "kept"
	config.HTTPConfig.DNSZones = []DNSZoneConfig{{Zone: "lab.example.com", HTTPSCertificateMode: "LETSENCRYPT", DNSChallengeProvider: "cloudflare"}}
	for _, host := range []string{"app.example.com", "x.lab.example.com", "nas.local", "10.0.0.5", "other.org", "a.com, b.com"} {
		config.HTTPConfig.ProxyConfig.Routes = append(config.HTTPConfig.ProxyConfig.Routes, ProxyRouteConfig{UseHost: true, Host: host})
	}

	if !MigrateToZones(&config, false) {
		t.Fatal("a self-signed server has something to migrate")
	}
	want := map[string]string{
		"lab.example.com": "LETSENCRYPT", // already a zone: it keeps deciding for its hostnames
		"example.com":     "SELFSIGNED",
		"other.org":       "SELFSIGNED",
		// nas.local and 10.0.0.5 are self-signed on their own
	}
	if got := hostNames(config); !reflect.DeepEqual(got, want) {
		t.Errorf("zones = %v, want %v", got, want)
	}
	if config.HTTPConfig.HTTPSCertificateMode != "LETSENCRYPT" || config.HTTPConfig.TLSCert != "kept" {
		t.Errorf("the server is left with HTTPS on and its own certificate, got %q / %q", config.HTTPConfig.HTTPSCertificateMode, config.HTTPConfig.TLSCert)
	}

	// every hostname is served as before
	for host, mode := range map[string]string{
		"cosmos.example.com": "SELFSIGNED", "app.example.com": "SELFSIGNED", "other.org": "SELFSIGNED",
		"nas.local": "SELFSIGNED", "10.0.0.5": "SELFSIGNED", "x.lab.example.com": "LETSENCRYPT",
		// and so is one added later
		"later.example.com": "SELFSIGNED",
	} {
		if got, _ := HostHTTPSMode(config, host); got != mode {
			t.Errorf("after migration %s: got %s, want %s", host, got, mode)
		}
	}

	// nothing left to do the second time
	if MigrateToZones(&config, false) {
		t.Error("the migration must not run twice")
	}
}

// in a cluster the zones are everyone's: a server only speaks for its own
// hostnames, and waits for the zones of the cluster before it does
func TestMigrateServerModeInACluster(t *testing.T) {
	newConfig := func() Config {
		config := Config{}
		config.ConstellationConfig.Enabled = true
		config.HTTPConfig.HTTPSCertificateMode = "SELFSIGNED"
		config.HTTPConfig.Hostname = "node-2.example.com"
		return config
	}

	config := newConfig()
	MigrateToZones(&config, false)
	if got, want := hostNames(config), (map[string]string{"node-2.example.com": "SELFSIGNED"}); !reflect.DeepEqual(got, want) {
		t.Errorf("zones = %v, want %v", got, want)
	}
	// the hostname of another server under the same domain is not touched
	if got, _ := HostHTTPSMode(config, "node-1.example.com"); got != "LETSENCRYPT" {
		t.Errorf("node-1.example.com: got %s, want LETSENCRYPT", got)
	}

	// a zone of the cluster already covers the hostname: nothing to add, it decides
	config = newConfig()
	config.HTTPConfig.DNSZones = []DNSZoneConfig{{Zone: "example.com", HTTPSCertificateMode: "LETSENCRYPT", DNSChallengeProvider: "cloudflare"}}
	if !MigrateToZones(&config, false) || len(config.HTTPConfig.DNSZones) != 1 || config.HTTPConfig.HTTPSCertificateMode != "LETSENCRYPT" {
		t.Errorf("covered by the cluster: got mode %q and zones %v", config.HTTPConfig.HTTPSCertificateMode, hostNames(config))
	}

	// not caught up with the cluster yet: nothing is decided
	previous := ZonesPending
	ZonesPending = func() bool { return true }
	defer func() { ZonesPending = previous }()
	config = newConfig()
	if MigrateToZones(&config, false) || len(config.HTTPConfig.DNSZones) != 0 || config.HTTPConfig.HTTPSCertificateMode != "SELFSIGNED" {
		t.Errorf("zones pending: got mode %q and zones %v, want nothing touched", config.HTTPConfig.HTTPSCertificateMode, hostNames(config))
	}
}

// a server-wide provided certificate goes along into the zones
func TestMigrateServerModeProvided(t *testing.T) {
	config := Config{}
	config.HTTPConfig.HTTPSCertificateMode = "PROVIDED"
	config.HTTPConfig.Hostname = "cosmos.example.com"
	config.HTTPConfig.TLSCert = "CERT"
	config.HTTPConfig.TLSKey = "KEY"
	config.HTTPConfig.TLSKeyHostsCached = []string{"cosmos.example.com"}
	config.HTTPConfig.ProxyConfig.Routes = []ProxyRouteConfig{{UseHost: true, Host: "nas.local"}}

	if !MigrateToZones(&config, false) {
		t.Fatal("a provided certificate has to move into zones")
	}
	if got, want := hostNames(config), (map[string]string{"example.com": "PROVIDED", "nas.local": "PROVIDED"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("zones = %v, want %v", got, want)
	}
	for _, zone := range config.HTTPConfig.DNSZones {
		if zone.HTTPSCertificateMode != "PROVIDED" || zone.TLSCert != "CERT" || zone.TLSKey != "KEY" {
			t.Errorf("%s = %+v, want the provided certificate", zone.Zone, zone)
		}
	}
	http := config.HTTPConfig
	if http.HTTPSCertificateMode != "LETSENCRYPT" || http.TLSCert != "" || http.TLSKey != "" || len(http.TLSKeyHostsCached) != 0 {
		t.Errorf("the server's own certificate must start over, got %q %q %v", http.HTTPSCertificateMode, http.TLSCert, http.TLSKeyHostsCached)
	}
}

// HTTPS on or off is the one thing a server keeps for itself
func TestMigrateServerModeLeavesTheRest(t *testing.T) {
	for _, mode := range []string{"LETSENCRYPT", "DISABLED", ""} {
		config := Config{}
		config.HTTPConfig.HTTPSCertificateMode = mode
		config.HTTPConfig.Hostname = "cosmos.example.com"
		if MigrateToZones(&config, false) || len(config.HTTPConfig.DNSZones) != 0 || config.HTTPConfig.HTTPSCertificateMode != mode {
			t.Errorf("mode %q must be left alone, got %q and %v", mode, config.HTTPConfig.HTTPSCertificateMode, config.HTTPConfig.DNSZones)
		}
	}
}

func TestResolveZonesExplicitAbsorbsDerived(t *testing.T) {
	config := Config{}
	config.HTTPConfig.DNSZones = []DNSZoneConfig{{Zone: "domain.com"}, {Zone: "empty.net"}}
	hosts := []string{"app1.domain.com", "app2.domain.com", "solo.other.org"}

	got := ResolveZones(config, hosts)
	want := []string{"domain.com", "empty.net", "solo.other.org(derived)"}
	if !reflect.DeepEqual(zoneNames(got), want) {
		t.Fatalf("zones = %v, want %v", zoneNames(got), want)
	}
	if !reflect.DeepEqual(got[0].Hosts, []string{"app1.domain.com", "app2.domain.com"}) {
		t.Errorf("domain.com hosts = %v", got[0].Hosts)
	}
	if len(got[1].Hosts) != 0 {
		t.Errorf("empty explicit zone should stay listed without hosts, got %v", got[1].Hosts)
	}
}

func TestMigrateToZones(t *testing.T) {
	config := Config{}
	config.HTTPConfig.Hostname = "cosmos.domain.com"
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	config.HTTPConfig.ProxyConfig.Routes = []ProxyRouteConfig{
		{UseHost: true, Host: "app.domain.com"},
		{UseHost: true, Host: "blog.other.co.uk"},
		{UseHost: true, Host: "a.com, b.com"},
		{UseHost: false, Host: "ignored.net"},
	}

	if MigrateToZones(&config, false) {
		t.Fatal("nothing to migrate without a legacy DNS provider")
	}

	config.HTTPConfig.DNSChallengeProvider = "cloudflare"
	config.HTTPConfig.DNSChallengeConfig = map[string]string{"CF_DNS_API_TOKEN": "t"}
	config.HTTPConfig.UseWildcardCertificate = true

	if !MigrateToZones(&config, false) {
		t.Fatal("expected a migration")
	}

	zones := config.HTTPConfig.DNSZones
	if len(zones) != 2 || zones[0].Zone != "domain.com" || zones[1].Zone != "other.co.uk" {
		t.Fatalf("zones = %+v", zones)
	}
	for _, z := range zones {
		if z.ManageRecords || z.WildcardRecord {
			t.Errorf("%s: a migrated zone must never manage records", z.Zone)
		}
		if z.DNSChallengeProvider != "cloudflare" || z.DNSChallengeConfig["CF_DNS_API_TOKEN"] != "t" || !z.UseWildcardCertificate {
			t.Errorf("%s: legacy HTTPS setup not carried over: %+v", z.Zone, z)
		}
	}

	if config.HTTPConfig.DNSChallengeProvider != "" || config.HTTPConfig.DNSChallengeConfig != nil || config.HTTPConfig.UseWildcardCertificate {
		t.Error("the legacy setup must be cleared once the zones own it")
	}
	if MigrateToZones(&config, false) {
		t.Error("migration must be one-shot")
	}
}

func TestMigrateToZonesDepositsCertAndKeepsExistingZones(t *testing.T) {
	config := Config{}
	config.HTTPConfig.Hostname = "cosmos.domain.com"
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	config.HTTPConfig.DNSChallengeProvider = "cloudflare"
	config.HTTPConfig.UseWildcardCertificate = true
	config.HTTPConfig.OverrideWildcardDomains = "*.lab.domain.com,lab.domain.com"
	config.HTTPConfig.TLSCert = "cert"
	config.HTTPConfig.TLSKey = "key"
	config.HTTPConfig.TLSKeyHostsCached = []string{"*.lab.domain.com", "lab.domain.com", "cosmos.domain.com"}
	config.HTTPConfig.DNSZones = []DNSZoneConfig{{Zone: "domain.com", DNSChallengeProvider: "route53"}}

	if !MigrateToZones(&config, false) {
		t.Fatal("expected a migration")
	}

	zones := config.HTTPConfig.DNSZones
	if len(zones) != 2 || zones[0].Zone != "domain.com" || zones[1].Zone != "lab.domain.com" {
		t.Fatalf("zones = %+v", zones)
	}
	if zones[0].DNSChallengeProvider != "route53" {
		t.Error("an existing zone must not be overwritten")
	}
	if !zones[1].UseWildcardCertificate {
		t.Error("a wildcard override must become a wildcard zone")
	}
	if cert := config.HTTPConfig.ZoneCerts["lab.domain.com"]; cert.TLSCert != "cert" || cert.TLSKey != "key" {
		t.Errorf("current certificate not deposited: %+v", cert)
	}
	if _, ok := config.HTTPConfig.ZoneCerts["domain.com"]; ok {
		t.Error("no deposit for a zone that already existed")
	}
}

func TestMigrateToZonesFromWizardManagesRecords(t *testing.T) {
	config := Config{}
	config.HTTPConfig.Hostname = "cosmos.domain.com"
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	config.HTTPConfig.DNSChallengeProvider = "cloudflare"
	config.HTTPConfig.UseWildcardCertificate = true

	prev := RecordsSupported
	RecordsSupported = func(provider string) bool { return provider == "cloudflare" }
	t.Cleanup(func() { RecordsSupported = prev })

	MigrateToZones(&config, true)

	zone := config.HTTPConfig.DNSZones[0]
	if !zone.ManageRecords || !zone.WildcardRecord {
		t.Errorf("a zone created by the wizard manages its records: %+v", zone)
	}

	other := Config{}
	other.HTTPConfig.Hostname = "cosmos.domain.com"
	other.HTTPConfig.DNSChallengeProvider = "godaddy"
	MigrateToZones(&other, true)
	if other.HTTPConfig.DNSZones[0].ManageRecords {
		t.Error("records cannot be managed through a provider without support")
	}
}

func TestZoneCertHostnames(t *testing.T) {
	hosts := []string{"domain.com", "b.domain.com", "a.domain.com", "x.deep.domain.com"}

	got := ZoneCertHostnames(DNSZoneConfig{Zone: "domain.com", UseWildcardCertificate: true}, hosts)
	want := []string{"domain.com", "*.domain.com", "x.deep.domain.com"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wildcard zone = %v, want %v", got, want)
	}

	got = ZoneCertHostnames(DNSZoneConfig{Zone: "domain.com"}, hosts)
	want = []string{"a.domain.com", "b.domain.com", "domain.com", "x.deep.domain.com"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plain zone = %v, want %v", got, want)
	}
}

func TestZoneCertCovers(t *testing.T) {
	cert := ZoneCert{Hosts: []string{"domain.com", "*.domain.com"}}

	if !ZoneCertCovers(cert, []string{"domain.com", "app.domain.com", "*.domain.com"}) {
		t.Error("wildcard pair should cover the apex and one level")
	}
	if ZoneCertCovers(cert, []string{"x.deep.domain.com"}) {
		t.Error("a wildcard covers one level only")
	}
	if ZoneCertCovers(cert, []string{"other.com"}) {
		t.Error("other.com is not covered")
	}
}

func TestLocalCertHostnames(t *testing.T) {
	config := Config{}
	config.HTTPConfig.DNSZones = []DNSZoneConfig{
		{Zone: "domain.com", DNSChallengeProvider: "cloudflare", HTTPSCertificateMode: "LETSENCRYPT"},
		{Zone: "noprovider.org", HTTPSCertificateMode: "LETSENCRYPT"},
		{Zone: "internal.lan.net", HTTPSCertificateMode: "SELFSIGNED"},
		{Zone: "own.org", HTTPSCertificateMode: "PROVIDED"},
		{Zone: "plain.net", HTTPSCertificateMode: "DISABLED"},
		{Zone: "unset.io"},
	}

	got := LocalCertHostnames(config, []string{"app.domain.com", "app.noprovider.org", "solo.net", "nas.internal.lan.net", "www.own.org", "app.plain.net", "x.unset.io"})
	want := []string{"app.noprovider.org", "solo.net", "x.unset.io"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("local hosts = %v, want %v", got, want)
	}
}

func TestGetHostCertificate(t *testing.T) {
	config := Config{}
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	config.HTTPConfig.TLSKeyHostsCached = []string{"solo.net"}
	config.HTTPConfig.SelfTLSKeyHostsCached = []string{"192.168.1.10"}
	config.HTTPConfig.DNSZones = []DNSZoneConfig{
		{Zone: "domain.com", DNSChallengeProvider: "cloudflare", HTTPSCertificateMode: "LETSENCRYPT"},
		{Zone: "lan.domain.com", HTTPSCertificateMode: "SELFSIGNED"},
	}
	config.HTTPConfig.ZoneCerts = map[string]ZoneCert{"domain.com": {Hosts: []string{"domain.com", "*.domain.com"}}}

	cases := map[string]struct {
		source  string
		covered bool
	}{
		"app.domain.com":     {"zone", true},
		"x.deep.domain.com":  {"node", false}, // zone certificate doesn't cover it yet
		"solo.net":           {"node", true},
		"nas.lan.domain.com": {"selfsigned", false},
		"new.org":            {"node", false},
		"192.168.1.10":       {"selfsigned", true},
		"nas.local":          {"selfsigned", false},
	}
	for host, want := range cases {
		got := GetHostCertificate(config, host)
		if got.Source != want.source || got.Covered != want.covered {
			t.Errorf("%s: got %s/%v, want %s/%v", host, got.Source, got.Covered, want.source, want.covered)
		}
	}

	// a server-wide mode from before zones does not come first anymore
	config.HTTPConfig.HTTPSCertificateMode = "SELFSIGNED"
	if got := GetHostCertificate(config, "app.domain.com"); got.Source != "zone" {
		t.Errorf("the zone must decide over a server-wide SELFSIGNED, got %s", got.Source)
	}
	if got := GetHostCertificate(config, "new.org"); got.Source != "selfsigned" {
		t.Errorf("without a zone the server-wide SELFSIGNED still applies, got %s", got.Source)
	}

	// provided and HTTP-only zones
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	config.HTTPConfig.DNSZones = append(config.HTTPConfig.DNSZones,
		DNSZoneConfig{Zone: "own.org", HTTPSCertificateMode: "PROVIDED", TLSCert: testCertificate(t, "own.org", "*.own.org")},
		DNSZoneConfig{Zone: "broken.org", HTTPSCertificateMode: "PROVIDED", TLSCert: "not a certificate"},
		DNSZoneConfig{Zone: "plain.net", HTTPSCertificateMode: "DISABLED"},
	)
	if got := GetHostCertificate(config, "www.own.org"); got.Source != "provided" || !got.Covered || got.Zone != "own.org" || got.ValidUntil.IsZero() {
		t.Errorf("www.own.org: got %+v, want the provided certificate of own.org", got)
	}
	if got := GetHostCertificate(config, "a.b.own.org"); got.Source != "provided" || got.Covered {
		t.Errorf("a.b.own.org: got %+v, want the provided certificate, not covering it", got)
	}
	if got := GetHostCertificate(config, "www.broken.org"); got.Source != "selfsigned" {
		t.Errorf("www.broken.org: got %s, want the self-signed fallback", got.Source)
	}
	if got := GetHostCertificate(config, "app.plain.net"); got.Source != "none" {
		t.Errorf("app.plain.net: got %s, want none", got.Source)
	}

	// a server that does not serve HTTPS has no certificate for anything
	config.HTTPConfig.HTTPSCertificateMode = "DISABLED"
	if got := GetHostCertificate(config, "app.domain.com"); got.Source != "none" {
		t.Errorf("HTTP-only server: got %s, want none", got.Source)
	}
}

// A name under a TLD that does not exist can never be certified by a public
// CA; it is self-signed by rule, like .local, instead of being ordered and
// refused on every restart.
func TestPublicTLD(t *testing.T) {
	for host, want := range map[string]bool{
		"app.example.com":      true,
		"node-1.cluster.io":    true,
		"app.example.com:8443": true,
		"app.example.com.":     true,
		"registry.test":        false,
		"node-1.constellation": false,
		"nas.lan":              false,
		"git.internal":         false,
		"box.home.arpa":        false,
		"intranet":             false,
		"":                     false,
	} {
		if got := PublicTLD(host); got != want {
			t.Errorf("PublicTLD(%q) = %v, want %v", host, got, want)
		}
	}
	if letsEncryptable("registry.test") || !letsEncryptable("registry.example.com") {
		t.Fatal("letsEncryptable must follow PublicTLD")
	}
	if got := LetsEncryptValidOnly([]string{"a.test", "b.example.org", "10.0.0.1"}, false); len(got) != 1 || got[0] != "b.example.org" {
		t.Fatalf("LetsEncryptValidOnly = %v", got)
	}
}
