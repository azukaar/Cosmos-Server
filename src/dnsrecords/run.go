package dnsrecords

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

// how often this node checks the address it advertises. With the default TTL of
// 60s this is min(30, TTL-30): a change of IP is out within a minute or so.
const addressCheckInterval = 30 * time.Second

const interfacePrefix = "iface:"

// IsWriter reports whether this node writes the DNS records. Standalone nodes
// always do, constellation overrides it with the leader check.
var IsWriter = func() bool { return true }

// OwnerID tells the records of this server or cluster from the ones another
// Cosmos wrote in the same zone. It comes from the auth key pair: made with
// the server, shared by every node of a cluster, and not lost when the
// constellation is created again. "" when there is none.
var OwnerID = func() string {
	key := strings.TrimSpace(utils.GetMainConfig().HTTPConfig.AuthPublicKey)
	if key == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:6])
}

// HostAddresses returns every hostname served with the addresses it must
// resolve to. Standalone: the hostnames of this node, at its own address.
// Constellation overrides it with the view of the whole cluster; ok is false
// when that view is not complete, and the pass is skipped.
var HostAddresses = func() (hosts map[string][]string, wildcard []string, ok bool) {
	address := CurrentAddress()
	if address == "" {
		return nil, nil, false
	}

	hosts = map[string][]string{}
	for _, host := range utils.GetAllHostnames(true) {
		hosts[strings.ToLower(host)] = []string{address}
	}
	return hosts, []string{address}, true
}

var currentAddress atomic.Value

// CurrentAddress is the address this node advertises for its hostnames, ""
// until it is known.
func CurrentAddress() string {
	address, _ := currentAddress.Load().(string)
	return address
}

// InterfaceAddress returns the first IPv4 of a network interface
func InterfaceAddress(name string) (string, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
			return ipnet.IP.String(), nil
		}
	}
	return "", errors.New("no IPv4 address on interface " + name)
}

// resolveAddress applies HTTPConfig.AdvertisedAddress: empty for the detected
// public IP, "iface:<name>" for the address of an interface (a home server
// published on its LAN), or a fixed IPv4.
func resolveAddress(setting string) (string, error) {
	setting = strings.TrimSpace(setting)

	if setting == "" {
		return DetectPublicIP()
	}
	if strings.HasPrefix(setting, interfacePrefix) {
		return InterfaceAddress(strings.TrimPrefix(setting, interfacePrefix))
	}
	if ip := net.ParseIP(setting); ip != nil && ip.To4() != nil {
		return ip.To4().String(), nil
	}
	return "", errors.New("invalid advertised address: " + setting)
}

// anyManagedZone reports whether there is any DynDNS work at all, so nodes
// without it never call out to find their public IP
func anyManagedZone() bool {
	for _, zone := range utils.GetMainConfig().HTTPConfig.DNSZones {
		if zone.ManageRecords {
			return true
		}
	}
	return false
}

func refreshAddress() {
	address, err := resolveAddress(utils.GetMainConfig().HTTPConfig.AdvertisedAddress)
	if err != nil {
		// keep advertising the last known address: a failed lookup is not a change of IP
		utils.Warn("[DynDNS] cannot find the address to advertise: " + err.Error())
		return
	}

	if previous := CurrentAddress(); previous != address {
		currentAddress.Store(address)
		if previous != "" {
			utils.Log("[DynDNS] advertised address changed from " + previous + " to " + address)
		}
	}
}

var trigger = make(chan struct{}, 1)
var startOnce sync.Once
var wasWriter bool

// Trigger asks for a reconcile pass now: the cluster view or the config moved
func Trigger() {
	select {
	case trigger <- struct{}{}:
	default:
	}
}

func pass() {
	if !anyManagedZone() {
		return
	}

	refreshAddress()

	if !IsWriter() {
		if wasWriter {
			ResetWriter()
			wasWriter = false
		}
		return
	}
	wasWriter = true

	hosts, wildcard, ok := HostAddresses()
	if !ok {
		return
	}

	zones := utils.GetMainConfig().HTTPConfig.DNSZones
	for i, zone := range zones {
		if !zone.ManageRecords || zone.DNSChallengeProvider == "" {
			continue
		}
		Reconcile(zones, i, BuildDesired(zones, i, hosts, wildcard))
	}
}

// Start runs the DynDNS loop: a pass on every trigger, and at least every
// addressCheckInterval to notice a change of IP.
func Start() {
	startOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(addressCheckInterval)
			defer ticker.Stop()

			for {
				pass()

				select {
				case <-trigger:
				case <-ticker.C:
				}
			}
		}()
	})
}
