package constellation

import (
	"math/rand"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/azukaar/cosmos-server/src/utils"
)

// isZoneIssuer reports whether this node issues the zone certificates of the
// cluster: the leader, or the only manager where there is no leader lease.
func isZoneIssuer() bool {
	config := utils.GetMainConfig()
	if !config.ConstellationConfig.Enabled {
		return true
	}

	device, err := GetCurrentDevice()
	if err != nil {
		return false
	}

	if !utils.IsPro() {
		return device.CosmosNode == 2
	}

	leader := GetCurrentLeaderName()
	return leader != "" && leader == device.DeviceName
}

// zonesPending: a server that has not materialized the cluster state yet does
// not know which of its hostnames belong to a zone of the cluster
func zonesPending() bool {
	return utils.GetMainConfig().ConstellationConfig.Enabled && !IsClientNode() && !utils.IsOplogBootstrapped()
}

// clusterHostnames lists every hostname served in the cluster, tunneled or not
func clusterHostnames() []string {
	clusterDNSMutex.RLock()
	defer clusterDNSMutex.RUnlock()

	names := make([]string, 0, len(clusterDNS))
	for name := range clusterDNS {
		names = append(names, name)
	}
	return names
}

func publishZoneCerts() {
	certs := utils.GetMainConfig().HTTPConfig.ZoneCerts
	if len(certs) == 0 || !utils.GetMainConfig().ConstellationConfig.Enabled {
		return
	}

	// the op-log may not be attached yet when the certificate comes out of a boot-time issuance
	go func() {
		for attempt := 0; attempt < 10; attempt++ {
			err := PublishDomainOp(DomainZoneCerts, utils.GetMainConfig().HTTPConfig.ZoneCerts)
			if err == nil {
				return
			}
			utils.Warn("[Zones] cannot publish the zone certificates yet: " + err.Error())
			time.Sleep(30 * time.Second)
		}
		utils.MajorError("[Zones] gave up publishing the zone certificates to the cluster", nil)
	}()
}

// shareLetsEncryptEmail offers the Let's Encrypt email of this server to a
// cluster that has none in common yet: before it was shared, each server kept
// its own, and the leader may well be one that never had any. Once any value
// was received the cluster's is the only one, so a server coming back with an
// old email of its own never overrides it.
func shareLetsEncryptEmail() {
	config := utils.GetMainConfig()
	if !config.ConstellationConfig.Enabled || config.HTTPConfig.SSLEmailShared || config.HTTPConfig.SSLEmail == "" {
		return
	}

	go func() {
		for attempt := 0; attempt < 20; attempt++ {
			time.Sleep(30 * time.Second)

			config := utils.GetMainConfig()
			if config.HTTPConfig.SSLEmailShared || config.HTTPConfig.SSLEmail == "" {
				return
			}
			// not caught up with the cluster yet: its email may be on the way
			if zonesPending() {
				continue
			}
			if err := PublishDomainOp(DomainHTTPSSettings, HTTPSSettingsPayload{SSLEmail: config.HTTPConfig.SSLEmail}); err == nil {
				return
			}
		}
	}()
}

// shareOIDCSigningKey gives a cluster without a common signing key the issuer's
func shareOIDCSigningKey() {
	config := utils.GetMainConfig()
	if !config.ConstellationConfig.Enabled || config.HTTPConfig.OIDCSigningKeyShared {
		return
	}

	go func() {
		for attempt := 0; attempt < 20; attempt++ {
			time.Sleep(30 * time.Second)

			http := utils.GetMainConfig().HTTPConfig
			if http.OIDCSigningKeyShared {
				return
			}
			if zonesPending() || !isZoneIssuer() || http.OIDCSigningKey == "" {
				continue
			}
			if err := PublishDomainOp(DomainOIDCSigningKey, OIDCSigningKeyPayload{Key: http.OIDCSigningKey}); err == nil {
				return
			}
		}
	}()
}

// publishMigratedZones adds zones a local migration created to the cluster's.
// A domain op carries the full state, and every node may migrate at the same
// time after an upgrade: each one re-checks that its zones made it, and merges
// them again if a concurrent publish dropped them.
func publishMigratedZones(own []utils.DNSZoneConfig) {
	if len(own) == 0 || !utils.GetMainConfig().ConstellationConfig.Enabled {
		return
	}

	go func() {
		for attempt := 0; attempt < 10; attempt++ {
			time.Sleep(time.Duration(15+rand.Intn(30)) * time.Second)

			merged := append([]utils.DNSZoneConfig{}, utils.GetMainConfig().HTTPConfig.DNSZones...)
			missing := false
			for _, zone := range own {
				if i := utils.FindZone(merged, zone.Zone); i == -1 || merged[i].Zone != zone.Zone {
					merged = append(merged, zone)
					missing = true
				}
			}

			// attempt 0 always publishes: the zones exist locally but nowhere else yet
			if !missing && attempt > 0 {
				return
			}
			if err := PublishDomainOp(DomainDNSZones, merged); err != nil {
				utils.Warn("[Zones] cannot publish the migrated zones yet: " + err.Error())
			}
		}
	}()
}

// checkZoneCerts runs on every cluster view refresh: the issuer gets the
// certificates the cluster is now missing, e.g. after gaining leadership or
// when another node starts serving a new hostname.
func checkZoneCerts() {
	if !isZoneIssuer() {
		return
	}
	if len(utils.ZoneCertsToIssue(utils.GetMainConfig(), false)) > 0 {
		go utils.RefreshZoneCerts()
	}
}

// ClusterDomain is the domain joining servers get their hostname under: the
// configured one, or the zone this server's own hostname lives in.
func ClusterDomain() string {
	config := utils.GetMainConfig()
	if config.ConstellationConfig.ClusterDomain != "" {
		return strings.ToLower(config.ConstellationConfig.ClusterDomain)
	}

	hostname := strings.ToLower(strings.Split(config.HTTPConfig.Hostname, ":")[0])
	if !utils.IsDomain(hostname) || utils.IsLocalDomain(hostname) {
		return ""
	}
	if zone, ok := utils.GetZoneForHost(hostname); ok && !zone.Derived {
		return zone.Zone
	}
	if domain, err := publicsuffix.EffectiveTLDPlusOne(hostname); err == nil {
		return domain
	}
	return ""
}

// DeviceHostname is the hostname a server joining as deviceName gets by default
func DeviceHostname(deviceName string, clusterDomain string) string {
	label := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		}
		return '-'
	}, deviceName)
	label = strings.Trim(label, "-")

	if label == "" || clusterDomain == "" {
		return ""
	}
	return label + "." + strings.ToLower(clusterDomain)
}
