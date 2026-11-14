package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

func testCertPEM(t *testing.T, names ...string) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: names[0]},
		DNSNames:     names,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
}

func servedName(t *testing.T, host string) string {
	t.Helper()
	cert, err := GetCertificate(&tls.ClientHelloInfo{ServerName: host})
	if err != nil || cert == nil {
		t.Fatalf("%s: no certificate (%v)", host, err)
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	return leaf.Subject.CommonName
}

// the handshake picks the zone certificate when it covers the host, the node's
// own one otherwise, and the self-signed one for self-signed zones and local names
func TestGetCertificatePerZone(t *testing.T) {
	load := func(names ...string) *tls.Certificate {
		pub, priv := testCertPEM(t, names...)
		cert, err := tls.X509KeyPair([]byte(pub), []byte(priv))
		if err != nil {
			t.Fatal(err)
		}
		return &cert
	}
	primaryCert = load("node-bundle")
	secondaryCert = load("self-signed", "reg.10-0-0-1.sslip.io")
	t.Cleanup(func() {
		primaryCert, secondaryCert = nil, nil
		zoneCertStore.Store(map[string]loadedZoneCert{})
		utils.LoadBaseMainConfig(utils.Config{})
	})

	zonePub, zonePriv := testCertPEM(t, "zone-domain.com", "domain.com", "*.domain.com")

	config := utils.Config{}
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	config.HTTPConfig.DNSZones = []utils.DNSZoneConfig{
		{Zone: "domain.com", DNSChallengeProvider: "cloudflare", HTTPSCertificateMode: "LETSENCRYPT"},
		{Zone: "lan.domain.com", HTTPSCertificateMode: "SELFSIGNED"},
		{Zone: "pending.org", DNSChallengeProvider: "cloudflare", HTTPSCertificateMode: "LETSENCRYPT"},
	}
	config.HTTPConfig.ZoneCerts = map[string]utils.ZoneCert{
		"domain.com": {TLSCert: zonePub, TLSKey: zonePriv, Hosts: []string{"domain.com", "*.domain.com"}},
	}
	utils.LoadBaseMainConfig(config)
	loadZoneCertStore()

	cases := map[string]string{
		"app.domain.com":        "zone-domain.com",
		"domain.com":            "zone-domain.com",
		"x.deep.domain.com":     "node-bundle", // not covered by the zone certificate yet
		"nas.lan.domain.com":    "self-signed",
		"app.pending.org":       "node-bundle", // zone certificate not issued yet
		"other.net":             "node-bundle",
		"reg.10-0-0-1.sslip.io": "self-signed", // Let's Encrypt could not certify it, only the self-signed one names it
		"192.168.1.10":          "self-signed",
		"cosmos.local":          "self-signed",
	}
	for host, want := range cases {
		if got := servedName(t, host); got != want {
			t.Errorf("%s served with %q, want %q", host, got, want)
		}
	}

	// a server-wide mode from before zones never comes before the zone: a
	// self-signed leader serves the certificate of its cluster like everyone else
	config.HTTPConfig.HTTPSCertificateMode = "SELFSIGNED"
	utils.LoadBaseMainConfig(config)
	if got := servedName(t, "app.domain.com"); got != "zone-domain.com" {
		t.Errorf("server-wide SELFSIGNED: app.domain.com served with %q, want the zone certificate", got)
	}
	if got := servedName(t, "other.net"); got != "self-signed" {
		t.Errorf("server-wide SELFSIGNED: other.net has no zone, served with %q, want self-signed", got)
	}
}

// a zone can carry its own certificate, or none at all
func TestGetCertificateProvidedAndHTTPOnly(t *testing.T) {
	load := func(names ...string) *tls.Certificate {
		pub, priv := testCertPEM(t, names...)
		cert, err := tls.X509KeyPair([]byte(pub), []byte(priv))
		if err != nil {
			t.Fatal(err)
		}
		return &cert
	}
	primaryCert = load("node-bundle")
	secondaryCert = load("self-signed")
	t.Cleanup(func() {
		primaryCert, secondaryCert = nil, nil
		zoneCertStore.Store(map[string]loadedZoneCert{})
		providedCertStore.Store(map[string]*tls.Certificate{})
		utils.LoadBaseMainConfig(utils.Config{})
	})

	ownPub, ownPriv := testCertPEM(t, "provided-own.org", "own.org", "*.own.org")
	_, otherPriv := testCertPEM(t, "someone-else")

	config := utils.Config{}
	config.HTTPConfig.HTTPSCertificateMode = "LETSENCRYPT"
	config.HTTPConfig.Hostname = "cosmos.own.org"
	config.HTTPConfig.DNSZones = []utils.DNSZoneConfig{
		{Zone: "own.org", HTTPSCertificateMode: "PROVIDED", TLSCert: ownPub, TLSKey: ownPriv},
		// the key of another certificate: unusable
		{Zone: "broken.org", HTTPSCertificateMode: "PROVIDED", TLSCert: ownPub, TLSKey: otherPriv},
		{Zone: "plain.net", HTTPSCertificateMode: "DISABLED"},
	}
	utils.LoadBaseMainConfig(config)
	loadZoneCertStore()

	cases := map[string]string{
		"www.own.org":    "provided-own.org",
		"own.org":        "provided-own.org",
		"www.broken.org": "self-signed",
		"other.net":      "node-bundle",
		// nobody should come here over TLS, and whoever does gets no real certificate
		"app.plain.net": "self-signed",
	}
	for host, want := range cases {
		if got := servedName(t, host); got != want {
			t.Errorf("%s served with %q, want %q", host, got, want)
		}
	}

	// The HTTP port serves the HTTP-only zones, and redirects everything else to
	// HTTPS. A bookmark, an app hard-coded to port 80 or a LAN health check that
	// receives 404 has no way to discover that the hostname lives on HTTPS.
	// This block asserted 404 before this PR restored the redirect (removed in
	// 748302d), so the expectation moves with the behaviour it tests.
	config.HTTPConfig.AllowHTTPLocalIPAccess = true
	utils.LoadBaseMainConfig(config)
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	for _, tt := range []struct {
		host   string
		remote string // client address; empty leaves httptest's default 1.2.3.4
		want   int
	}{
		// an HTTP-only zone is served over plain HTTP, whatever the client address
		{host: "app.plain.net", want: http.StatusTeapot},
		{host: "app.plain.net:8080", want: http.StatusTeapot},
		// a hostname of an HTTPS zone arriving on the HTTP port is redirected
		{host: "www.own.org", remote: "1.2.3.4:5678", want: http.StatusMovedPermanently},
		{host: "other.net", remote: "1.2.3.4:5678", want: http.StatusMovedPermanently},
		// a bare IP redirects only for a remote client. From a local address it is
		// served directly (AllowHTTPLocalIPAccess), which is what keeps LAN access
		// working without a redirect loop.
		//
		// The old fixture never set RemoteAddr, so httptest's default 1.2.3.4 was
		// used for every case: not loopback, not private and not CGNAT, so
		// IsLocalIP was false and the passthrough branch below was never reached.
		// Setting it explicitly is what makes the redirect cases and the
		// passthrough cases both reachable in the test.
		{host: "192.168.1.10", remote: "1.2.3.4:5678", want: http.StatusMovedPermanently},
		{host: "192.168.1.10", remote: "192.168.1.10:5678", want: http.StatusTeapot},
		// 100.64/10 (CGNAT, what a Tailscale address sits in) is served directly
		// too: IsLocalIP covers that range explicitly, beyond the private ranges
		// net.IP.IsPrivate knows. Constellation peers are covered here as well as
		// by its own IsConstellationIP check — its default range 192.168.201.0/24
		// is private, so URL tunneling keeps working over plain HTTP.
		{host: "100.64.5.6", remote: "100.64.5.6:5678", want: http.StatusTeapot},
	} {
		request := httptest.NewRequest("GET", "http://"+tt.host+"/", nil)
		if tt.remote != "" {
			request.RemoteAddr = tt.remote
		}
		response := httptest.NewRecorder()
		plainHTTPHandler(router).ServeHTTP(response, request)
		if response.Code != tt.want {
			t.Errorf("http://%s from %s answered %d, want %d", tt.host, tt.remote, response.Code, tt.want)
		}
	}

	// cookies and forwarded scheme follow the hostname, the server's own URL its own hostname
	utils.HTTPSListening = true
	t.Cleanup(func() { utils.HTTPSListening = false })
	if !utils.HostServedOverHTTPS("www.own.org") || utils.HostServedOverHTTPS("app.plain.net") {
		t.Error("www.own.org is served over HTTPS, app.plain.net is not")
	}
}
