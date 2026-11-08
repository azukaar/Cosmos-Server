package utils

import (
	"crypto/x509"
	"encoding/pem"
	osnet "net"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// DNSZoneConfig is an explicit zone: a domain with its HTTPS setup and its
// DynDNS setup. Hostnames no explicit zone covers get a derived zone of their
// own at runtime (see ResolveZones), which is never stored.
type DNSZoneConfig struct {
	Zone string `validate:"required,excludesall=0x2C/ "`

	// HTTPS setup. The mode is LETSENCRYPT, SELFSIGNED, PROVIDED (the
	// certificate below) or DISABLED (the hostnames are served over plain HTTP).
	// The zone decides alone: no server has an HTTPS mode of its own to
	// override it, only the choice of not serving HTTPS at all.
	HTTPSCertificateMode        string
	TLSCert                     string `json:"TLSCert,omitempty"`
	TLSKey                      string `json:"TLSKey,omitempty"`
	UseWildcardCertificate      bool
	DNSChallengeProvider        string
	DNSChallengeConfig          map[string]string `json:"DNSChallengeConfig,omitempty"`
	DNSChallengeResolvers       string
	DNSChallengePropagationWait int
	DisablePropagationChecks    bool

	// DynDNS setup, needs a DNSChallengeProvider
	ManageRecords  bool
	WildcardRecord bool
}

// ResolvedZone is a zone as the server sees it: explicit or derived, with the
// hostnames currently filed under it.
type ResolvedZone struct {
	DNSZoneConfig
	Derived bool
	Hosts   []string
}

// ZoneCert is the certificate of an explicit zone. The zone issuer (the node
// itself when standalone, the cluster leader otherwise) is the only one
// writing it; every node serves from it.
type ZoneCert struct {
	TLSCert    string
	TLSKey     string
	Hosts      []string
	ValidUntil time.Time
	IssuedBy   string `json:"IssuedBy,omitempty"`
}

// IsZoneIssuer reports whether this node issues zone certificates. Standalone
// nodes always do, constellation overrides it with the leader check.
var IsZoneIssuer = func() bool { return true }

// PublishZoneCerts hands the zone certificates of the issuer to the rest of the
// cluster, constellation overrides it.
var PublishZoneCerts = func() {}

// PublishMigratedZones hands zones created by a local migration to the rest of
// the cluster, constellation overrides it.
var PublishMigratedZones = func(zones []DNSZoneConfig) {}

// ZonesPending reports that this server has joined a cluster whose zones it has
// not received yet, constellation overrides it.
var ZonesPending = func() bool { return false }

// RecordsSupported reports whether the records of a zone can be managed
// through a DNS challenge provider, wired by the dnsrecords package.
var RecordsSupported = func(provider string) bool { return false }

// RefreshZoneCerts issues whatever zone certificate is missing and serves it,
// without restarting the server. ReloadZoneCerts only re-reads them from the
// config. Both are wired by the HTTP server.
var RefreshZoneCerts = func() {}
var ReloadZoneCerts = func() {}

// ClusterHostnames lists the hostnames served by the other nodes of the
// cluster, constellation overrides it.
var ClusterHostnames = func() []string { return nil }

// zoneHost normalises a hostname for zone matching: no port, lowercase, no trailing dot
func zoneHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, err := osnet.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.TrimSuffix(host, ".")
}

// zoneableHost reports whether a hostname can belong to a zone at all
func zoneableHost(host string) bool {
	if host == "" || host == "localhost" || strings.ContainsAny(host, ", ") {
		return false
	}
	return osnet.ParseIP(host) == nil && strings.Contains(host, ".")
}

// FindZone returns the index of the explicit zone owning the hostname
// (longest suffix wins), -1 if none.
func FindZone(zones []DNSZoneConfig, host string) int {
	host = strings.TrimPrefix(zoneHost(host), "*.")
	best := -1

	for i, zone := range zones {
		name := zoneHost(zone.Zone)
		if name == "" {
			continue
		}
		if host == name || strings.HasSuffix(host, "."+name) {
			if best == -1 || len(name) > len(zoneHost(zones[best].Zone)) {
				best = i
			}
		}
	}

	return best
}

// ServerHTTPSDisabled reports that this server serves plain HTTP only. It is the
// one HTTPS setting a server keeps for itself (HTTPConfig.HTTPSCertificateMode
// DISABLED, or never set up): it is the only one a hostname without a domain,
// such as an IP, can carry.
func ServerHTTPSDisabled(http HTTPConfig) bool {
	return http.HTTPSCertificateMode == HTTPSCertModeList["DISABLED"] || http.HTTPSCertificateMode == ""
}

var notLetsEncryptable = regexp.MustCompile(`^(localhost|(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})|.*\.local)$`)

// letsEncryptable reports whether a public CA can certify a hostname at all:
// not an IP, not localhost, not .local, no wildcard or list — and under a
// top-level domain that exists. A name under a made-up or special-use TLD
// (.test, .lan, .internal, .home.arpa) is rejected by Let's Encrypt as an
// invalid identifier, after a round trip, on every single order; it belongs
// on the self-signed certificate like an IP does. Mirrors LetsEncryptValidOnly
// for one hostname, without its log.
func letsEncryptable(host string) bool {
	return host != "" && !notLetsEncryptable.MatchString(host) && !strings.ContainsAny(host, "* ,") &&
		!strings.Contains(host, "::") && PublicTLD(host)
}

// PublicTLD reports whether the last label of a hostname is a top-level domain
// a public CA will issue under: one in the ICANN section of the public suffix
// list. Ports are ignored; a trailing dot is tolerated.
func PublicTLD(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if i := strings.LastIndex(host, ":"); i != -1 && !strings.Contains(host[i+1:], ".") {
		host = host[:i]
	}
	tld := host
	if i := strings.LastIndex(host, "."); i != -1 {
		tld = host[i+1:]
	}
	if tld == "" || tld == host && !strings.Contains(host, ".") && len(host) > 0 {
		// a bare label ("intranet") is a TLD by itself and never a public name
		return false
	}
	// .arpa is ICANN's but infrastructure only (home.arpa, reverse DNS)
	if tld == "arpa" {
		return false
	}
	_, icann := publicsuffix.PublicSuffix(tld)
	return icann
}

// automaticMode is the HTTPS mode of a hostname no explicit zone covers. It is
// a rule, not a setting: Let's Encrypt for a public name (with the self-signed
// certificate as the fallback when it cannot be obtained), self-signed for what
// Let's Encrypt cannot certify. A server-wide self-signed or provided mode from
// before zones only shows here until MigrateToZones has turned it into zones.
func automaticMode(http HTTPConfig, host string) string {
	if ServerHTTPSDisabled(http) {
		return HTTPSCertModeList["DISABLED"]
	}
	if http.HTTPSCertificateMode == HTTPSCertModeList["SELFSIGNED"] || http.HTTPSCertificateMode == HTTPSCertModeList["PROVIDED"] {
		return http.HTTPSCertificateMode
	}
	if !letsEncryptable(host) {
		return HTTPSCertModeList["SELFSIGNED"]
	}
	return HTTPSCertModeList["LETSENCRYPT"]
}

// HostHTTPSMode is how a hostname is served, with the index of the explicit
// zone deciding it (-1 for an automatic one). The zone decides alone, unless
// the server does not serve HTTPS at all.
func HostHTTPSMode(config Config, host string) (string, int) {
	host = zoneHost(host)
	i := FindZone(config.HTTPConfig.DNSZones, host)
	if ServerHTTPSDisabled(config.HTTPConfig) {
		return HTTPSCertModeList["DISABLED"], i
	}
	if i == -1 {
		return automaticMode(config.HTTPConfig, host), i
	}
	if mode := config.HTTPConfig.DNSZones[i].HTTPSCertificateMode; mode != "" {
		return mode, i
	}
	return HTTPSCertModeList["LETSENCRYPT"], i
}

// HostIsHTTPOnly reports that a hostname is served over plain HTTP
func HostIsHTTPOnly(config Config, host string) bool {
	mode, _ := HostHTTPSMode(config, host)
	return mode == HTTPSCertModeList["DISABLED"]
}

// HostServedOverHTTPS reports whether the requests for a hostname come in over
// HTTPS on this server. Same answer as IsHTTPS for every hostname, until a
// zone is HTTP-only on a server that otherwise serves HTTPS.
func HostServedOverHTTPS(host string) bool {
	return HTTPSListening && !HostIsHTTPOnly(GetMainConfig(), host)
}

// ServesSelfSigned reports whether the server's own hostname is served with a
// certificate no client trusts on its own
func ServesSelfSigned() bool {
	config := GetMainConfig()
	return GetHostCertificate(config, config.HTTPConfig.Hostname).Source == "selfsigned"
}

// CertificateInfo reads what a PEM certificate covers and until when
func CertificateInfo(certPEM string) ([]string, time.Time, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, time.Time{}, errInvalidCertificate
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, time.Time{}, err
	}
	hosts := append([]string{}, cert.DNSNames...)
	for _, ip := range cert.IPAddresses {
		hosts = append(hosts, ip.String())
	}
	if len(hosts) == 0 && cert.Subject.CommonName != "" {
		hosts = append(hosts, cert.Subject.CommonName)
	}
	return hosts, cert.NotAfter, nil
}

type zoneError string

func (e zoneError) Error() string { return string(e) }

const errInvalidCertificate = zoneError("not a PEM certificate")

// derivedZone is the zone of a hostname nothing explicit covers: the hostname
// itself, on the automatic HTTPS mode, without DynDNS.
func derivedZone(host string, http HTTPConfig) DNSZoneConfig {
	return DNSZoneConfig{
		Zone:                 host,
		HTTPSCertificateMode: automaticMode(http, host),
	}
}

// ConfigHostnames lists every hostname the config serves: main hostname and
// host-based routes. Multi-host routes are skipped, like everywhere else.
func ConfigHostnames(config Config) []string {
	hosts := []string{config.HTTPConfig.Hostname}
	for _, route := range config.HTTPConfig.ProxyConfig.Routes {
		if route.UseHost && route.Host != "" && !strings.ContainsAny(route.Host, ", ") {
			hosts = append(hosts, route.Host)
		}
	}
	return hosts
}

// ResolveZones files the hostnames under their zones: every explicit zone
// (even empty), then one derived zone per hostname left over. Sorted by name.
func ResolveZones(config Config, hosts []string) []ResolvedZone {
	explicit := config.HTTPConfig.DNSZones
	resolved := make([]ResolvedZone, len(explicit))
	for i, zone := range explicit {
		resolved[i] = ResolvedZone{DNSZoneConfig: zone}
	}

	derived := map[string]bool{}
	for _, host := range hosts {
		host = zoneHost(host)
		if !zoneableHost(host) {
			continue
		}

		if i := FindZone(explicit, host); i != -1 {
			resolved[i].Hosts = appendUnique(resolved[i].Hosts, host)
			continue
		}

		if !derived[host] {
			derived[host] = true
			resolved = append(resolved, ResolvedZone{
				DNSZoneConfig: derivedZone(host, config.HTTPConfig),
				Derived:       true,
				Hosts:         []string{host},
			})
		}
	}

	sort.SliceStable(resolved, func(i, j int) bool {
		return resolved[i].Zone < resolved[j].Zone
	})

	return resolved
}

// GetZoneForHost returns the zone a hostname resolves to in the live config.
func GetZoneForHost(host string) (ResolvedZone, bool) {
	config := GetMainConfig()
	host = zoneHost(host)
	if !zoneableHost(host) {
		return ResolvedZone{}, false
	}

	if i := FindZone(config.HTTPConfig.DNSZones, host); i != -1 {
		return ResolvedZone{DNSZoneConfig: config.HTTPConfig.DNSZones[i]}, true
	}
	return ResolvedZone{DNSZoneConfig: derivedZone(host, config.HTTPConfig), Derived: true}, true
}

func appendUnique(list []string, s string) []string {
	for _, v := range list {
		if v == s {
			return list
		}
	}
	return append(list, s)
}

// IssuesZoneCert reports whether the zone gets its own certificate from the
// zone issuer. That takes a DNS provider: without one the hosts of the zone
// stay in the node-local HTTP-01 certificate, like a derived zone.
func IssuesZoneCert(zone DNSZoneConfig) bool {
	return zone.DNSChallengeProvider != "" &&
		zone.HTTPSCertificateMode == HTTPSCertModeList["LETSENCRYPT"]
}

// ZoneCertHostnames is what the certificate of a zone must cover given the
// hostnames filed under it: the wildcard pair plus anything deeper than one
// level, or just the hostnames.
func ZoneCertHostnames(zone DNSZoneConfig, hosts []string) []string {
	name := zoneHost(zone.Zone)
	names := []string{}

	if zone.UseWildcardCertificate {
		names = append(names, name, "*."+name)
		for _, host := range hosts {
			if !CertCovers(names, host) {
				names = append(names, host)
			}
		}
		return names
	}

	for _, host := range hosts {
		names = appendUnique(names, host)
	}
	sort.Strings(names)
	return names
}

// CertCovers reports whether a certificate issued for certHosts serves host
func CertCovers(certHosts []string, host string) bool {
	for _, certHost := range certHosts {
		if certHost == host {
			return true
		}
		if strings.HasPrefix(certHost, "*.") {
			rest := strings.TrimSuffix(host, certHost[1:])
			if rest != host && rest != "" && !strings.Contains(rest, ".") {
				return true
			}
		}
	}
	return false
}

// ZoneCertCovers reports whether the certificate serves every wanted hostname
func ZoneCertCovers(cert ZoneCert, wanted []string) bool {
	for _, host := range wanted {
		if !CertCovers(cert.Hosts, host) {
			return false
		}
	}
	return true
}

// ZoneCertsToIssue lists the explicit zones whose certificate is missing,
// expiring or not covering their hostnames anymore, with what it must cover.
func ZoneCertsToIssue(config Config, force bool) map[string][]string {
	toIssue := map[string][]string{}
	hosts := append(GetAllHostnames(true), ClusterHostnames()...)

	for _, zone := range ResolveZones(config, hosts) {
		if zone.Derived || !IssuesZoneCert(zone.DNSZoneConfig) {
			continue
		}
		wanted := ZoneCertHostnames(zone.DNSZoneConfig, zone.Hosts)
		if len(wanted) == 0 {
			continue
		}
		cert, ok := config.HTTPConfig.ZoneCerts[zone.Zone]
		if ok && !force && ZoneCertCovers(cert, wanted) && time.Now().Add(45*24*time.Hour).Before(cert.ValidUntil) {
			continue
		}
		toIssue[zone.Zone] = wanted
	}

	return toIssue
}

// LocalCertHostnames filters the hostnames this node must put in its own
// Let's Encrypt certificate: the ones of automatic zones, and of the Let's
// Encrypt zones without a DNS provider. Every other zone has a certificate of
// its own, serves another one, or none.
func LocalCertHostnames(config Config, hosts []string) []string {
	local := []string{}
	for _, host := range hosts {
		if i := FindZone(config.HTTPConfig.DNSZones, host); i != -1 {
			zone := config.HTTPConfig.DNSZones[i]
			if IssuesZoneCert(zone) || (zone.HTTPSCertificateMode != "" && zone.HTTPSCertificateMode != HTTPSCertModeList["LETSENCRYPT"]) {
				continue
			}
		}
		local = append(local, host)
	}
	return local
}

// LocalCertZone names the certificate a hostname is served from among the
// ones this server gets by itself: its explicit zone, or the hostname itself
// for an automatic one.
func LocalCertZone(config Config, host string) string {
	host = zoneHost(host)
	if i := FindZone(config.HTTPConfig.DNSZones, host); i != -1 {
		return config.HTTPConfig.DNSZones[i].Zone
	}
	return host
}

// LocalCertGroups groups the hostnames of the server's own certificates by
// domain: one order per domain, so a hostname that fails validation only
// costs its own domain.
func LocalCertGroups(config Config, hosts []string) map[string][]string {
	groups := map[string][]string{}
	for _, host := range hosts {
		host = zoneHost(host)
		if !zoneableHost(host) {
			continue
		}
		zone := LocalCertZone(config, host)
		groups[zone] = appendUnique(groups[zone], host)
	}
	return groups
}

// LocalCertsToIssue lists the domains whose certificate this server has to
// order: none yet, a hostname it does not cover, or expiring within 45 days.
// The single certificate of older versions keeps serving the domains it
// covers until then.
func LocalCertsToIssue(config Config, hosts []string, force bool) map[string][]string {
	toIssue := map[string][]string{}
	http := config.HTTPConfig
	for zone, wanted := range LocalCertGroups(config, hosts) {
		if !force {
			if cert, ok := http.LocalCerts[zone]; ok && certServes(cert.Hosts, cert.ValidUntil, wanted) {
				continue
			}
			if http.TLSCert != "" && certServes(http.TLSKeyHostsCached, http.TLSValidUntil, wanted) {
				continue
			}
		}
		toIssue[zone] = wanted
	}
	return toIssue
}

// LocalCertHostnamesNow is what this server has to certify by itself right now
func LocalCertHostnamesNow(config Config) []string {
	return LetsEncryptValidOnly(LocalCertHostnames(config, GetAllHostnames(true)), false)
}

// NeedsLetsEncrypt reports whether this server orders certificates from Let's Encrypt
func NeedsLetsEncrypt(config Config) bool {
	disabled := ServerHTTPSDisabled(config.HTTPConfig)
	if !disabled {
		for _, host := range LocalCertHostnames(config, GetAllHostnames(true)) {
			if letsEncryptable(zoneHost(host)) {
				return true
			}
		}
	}

	// an HTTP-only leader still issues for its cluster, see issueZoneCerts
	if (disabled && !config.ConstellationConfig.Enabled) || !IsZoneIssuer() {
		return false
	}
	for _, zone := range ResolveZones(config, append(GetAllHostnames(true), ClusterHostnames()...)) {
		if !zone.Derived && IssuesZoneCert(zone.DNSZoneConfig) && len(ZoneCertHostnames(zone.DNSZoneConfig, zone.Hosts)) > 0 {
			return true
		}
	}
	return false
}

// certServes reports whether a certificate covers every wanted hostname and
// is good for 45 more days
func certServes(certHosts []string, validUntil time.Time, wanted []string) bool {
	if !time.Now().Add(45 * 24 * time.Hour).Before(validUntil) {
		return false
	}
	for _, host := range wanted {
		if !CertCovers(certHosts, host) {
			return false
		}
	}
	return true
}

// MigrateToZones moves the HTTPS setup a server used to hold for itself into
// zones: its DNS challenge provider, then its self-signed or provided mode. The
// install wizard and the older clients still fill those legacy fields, so this
// is also what turns their form into zones. Returns true when the config was
// changed.
func MigrateToZones(config *Config, manageRecords bool) bool {
	challenge := migrateDNSChallenge(config, manageRecords)
	mode := migrateServerMode(config)
	return challenge || mode
}

// migrateServerMode moves a server-wide self-signed or provided mode into
// zones, so that nothing changes for the hostnames it was serving (the provided
// certificate goes along). The server itself is left with a single choice,
// HTTPS or not.
//
// Alone, the server gets one zone per registrable domain: everything was
// served that way, and the hostnames added later keep being. In a cluster the
// zones are everyone's, so each hostname gets a zone of its own and no other
// server sees its hostnames change.
func migrateServerMode(config *Config) bool {
	http := config.HTTPConfig
	mode := http.HTTPSCertificateMode
	if mode != HTTPSCertModeList["SELFSIGNED"] && mode != HTTPSCertModeList["PROVIDED"] {
		return false
	}
	// the zones of the cluster may already cover these hostnames: wait for them
	if ZonesPending() {
		return false
	}

	for _, host := range ConfigHostnames(*config) {
		host = zoneHost(host)
		if !zoneableHost(host) || strings.Contains(host, "*") || FindZone(config.HTTPConfig.DNSZones, host) != -1 {
			continue
		}
		// already what such a name gets on its own
		if mode == HTTPSCertModeList["SELFSIGNED"] && !letsEncryptable(host) {
			continue
		}

		name := host
		if !config.ConstellationConfig.Enabled {
			if registrable, err := publicsuffix.EffectiveTLDPlusOne(host); err == nil {
				name = registrable
			}
		}

		zone := DNSZoneConfig{Zone: name, HTTPSCertificateMode: mode}
		if mode == HTTPSCertModeList["PROVIDED"] {
			zone.TLSCert = http.TLSCert
			zone.TLSKey = http.TLSKey
		}
		config.HTTPConfig.DNSZones = append(config.HTTPConfig.DNSZones, zone)
	}

	sort.SliceStable(config.HTTPConfig.DNSZones, func(i, j int) bool {
		return config.HTTPConfig.DNSZones[i].Zone < config.HTTPConfig.DNSZones[j].Zone
	})

	config.HTTPConfig.HTTPSCertificateMode = HTTPSCertModeList["LETSENCRYPT"]
	if mode == HTTPSCertModeList["PROVIDED"] {
		// the zones carry it now: the node's own certificate starts over
		config.HTTPConfig.TLSCert = ""
		config.HTTPConfig.TLSKey = ""
		config.HTTPConfig.TLSKeyHostsCached = nil
		config.HTTPConfig.TLSValidUntil = time.Time{}
	}

	return true
}

// migrateDNSChallenge moves the legacy global DNS challenge setup into explicit
// zones, one per registrable domain in use (the legacy credentials were
// serving all of them) or per wildcard override. The install wizard still
// fills the legacy fields, so this is also what turns its form into the first
// zone, records included (manageRecords). An upgrade passes false: it must not
// touch anyone's DNS. The current certificate is deposited under every new
// zone so nothing is re-issued. Returns true when the config was changed.
func migrateDNSChallenge(config *Config, manageRecords bool) bool {
	http := config.HTTPConfig
	if http.DNSChallengeProvider == "" {
		return false
	}

	manageRecords = manageRecords && RecordsSupported(http.DNSChallengeProvider)

	newZone := func(name string, wildcard bool) {
		name = strings.TrimPrefix(zoneHost(name), "*.")
		if !zoneableHost(name) {
			return
		}
		for _, zone := range config.HTTPConfig.DNSZones {
			if zoneHost(zone.Zone) == name {
				return
			}
		}

		zone := DNSZoneConfig{
			Zone:                        name,
			HTTPSCertificateMode:        http.HTTPSCertificateMode,
			UseWildcardCertificate:      wildcard,
			DNSChallengeProvider:        http.DNSChallengeProvider,
			DNSChallengeResolvers:       http.DNSChallengeResolvers,
			DNSChallengePropagationWait: http.DNSChallengePropagationWait,
			DisablePropagationChecks:    http.DisablePropagationChecks,
			ManageRecords:               manageRecords,
			WildcardRecord:              manageRecords && wildcard,
		}
		if len(http.DNSChallengeConfig) > 0 {
			zone.DNSChallengeConfig = map[string]string{}
			for k, v := range http.DNSChallengeConfig {
				zone.DNSChallengeConfig[k] = v
			}
		}
		config.HTTPConfig.DNSZones = append(config.HTTPConfig.DNSZones, zone)

		if http.HTTPSCertificateMode == HTTPSCertModeList["LETSENCRYPT"] && http.TLSCert != "" && http.TLSKey != "" {
			if config.HTTPConfig.ZoneCerts == nil {
				config.HTTPConfig.ZoneCerts = map[string]ZoneCert{}
			}
			config.HTTPConfig.ZoneCerts[name] = ZoneCert{
				TLSCert:    http.TLSCert,
				TLSKey:     http.TLSKey,
				Hosts:      http.TLSKeyHostsCached,
				ValidUntil: http.TLSValidUntil,
			}
		}
	}

	if http.UseWildcardCertificate && http.OverrideWildcardDomains != "" {
		for _, override := range strings.Split(http.OverrideWildcardDomains, ",") {
			newZone(override, true)
		}
	}

	for _, host := range ConfigHostnames(*config) {
		host = zoneHost(host)
		if !zoneableHost(host) || FindZone(config.HTTPConfig.DNSZones, host) != -1 {
			continue
		}
		name, err := publicsuffix.EffectiveTLDPlusOne(host)
		if err != nil {
			name = host
		}
		newZone(name, http.UseWildcardCertificate && http.OverrideWildcardDomains == "" && canDomainWildcard(name))
	}

	sort.SliceStable(config.HTTPConfig.DNSZones, func(i, j int) bool {
		return config.HTTPConfig.DNSZones[i].Zone < config.HTTPConfig.DNSZones[j].Zone
	})

	// the zones own the setup from now on
	config.HTTPConfig.DNSChallengeProvider = ""
	config.HTTPConfig.DNSChallengeConfig = nil
	config.HTTPConfig.UseWildcardCertificate = false
	config.HTTPConfig.OverrideWildcardDomains = ""

	return true
}

// HostCertificate describes the certificate a hostname is served with, for display.
type HostCertificate struct {
	// Source is "zone", "node" (this node's own Let's Encrypt certificate),
	// "selfsigned", "provided" or "none" (HTTPS disabled)
	Source     string    `json:"source"`
	Zone       string    `json:"zone,omitempty"`
	Hosts      []string  `json:"hosts"`
	ValidUntil time.Time `json:"validUntil"`
	// Covered is false when the certificate served does not list the hostname
	Covered bool `json:"covered"`
	// PEM is the certificate itself, kept out of the zones listing
	PEM string `json:"-"`
}

// GetHostCertificate mirrors the choice the TLS handshake makes for a hostname
func GetHostCertificate(config Config, host string) HostCertificate {
	http := config.HTTPConfig
	host = zoneHost(host)

	mode, i := HostHTTPSMode(config, host)
	zoneName := ""
	if i != -1 {
		zoneName = http.DNSZones[i].Zone
	}

	selfSigned := HostCertificate{
		Source:     "selfsigned",
		Zone:       zoneName,
		Hosts:      http.SelfTLSKeyHostsCached,
		ValidUntil: http.SelfTLSValidUntil,
		Covered:    CertCovers(http.SelfTLSKeyHostsCached, host),
		PEM:        http.SelfTLSCert,
	}

	switch mode {
	case HTTPSCertModeList["DISABLED"]:
		return HostCertificate{Source: "none", Zone: zoneName, Hosts: []string{}}
	case HTTPSCertModeList["SELFSIGNED"]:
		return selfSigned
	case HTTPSCertModeList["PROVIDED"]:
		if i == -1 {
			return HostCertificate{Source: "provided", Hosts: []string{}, ValidUntil: http.TLSValidUntil, Covered: true, PEM: http.TLSCert}
		}
		hosts, validUntil, err := CertificateInfo(http.DNSZones[i].TLSCert)
		if err != nil {
			// nothing usable was provided: the self-signed certificate serves meanwhile
			return selfSigned
		}
		return HostCertificate{Source: "provided", Zone: zoneName, Hosts: hosts, ValidUntil: validUntil, Covered: CertCovers(hosts, host), PEM: http.DNSZones[i].TLSCert}
	}

	if !zoneableHost(host) || !letsEncryptable(host) {
		return selfSigned
	}

	if i != -1 {
		if cert, ok := http.ZoneCerts[zoneName]; ok && CertCovers(cert.Hosts, host) {
			return HostCertificate{
				Source:     "zone",
				Zone:       zoneName,
				Hosts:      cert.Hosts,
				ValidUntil: cert.ValidUntil,
				Covered:    true,
				PEM:        cert.TLSCert,
			}
		}
	}

	if cert, ok := http.LocalCerts[LocalCertZone(config, host)]; ok && CertCovers(cert.Hosts, host) {
		return HostCertificate{
			Source:     "node",
			Zone:       zoneName,
			Hosts:      cert.Hosts,
			ValidUntil: cert.ValidUntil,
			Covered:    true,
			PEM:        cert.TLSCert,
		}
	}

	return HostCertificate{
		Source:     "node",
		Hosts:      http.TLSKeyHostsCached,
		ValidUntil: http.TLSValidUntil,
		Covered:    CertCovers(http.TLSKeyHostsCached, host),
		PEM:        http.TLSCert,
	}
}
