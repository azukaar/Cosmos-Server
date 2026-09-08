package docker

import (
	"net"
	"sync"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/docker/docker/api/types"
)

// localSubnetsTTL bounds staleness; the cache is also dropped on network create/destroy events.
const localSubnetsTTL = 30 * time.Second

var (
	localSubnetsMu         sync.Mutex
	localSubnets           []*net.IPNet
	localSubnetsExpires    time.Time
	localSubnetsRefreshing bool
)

// ForgetLocalNetworkSubnets expires the cache (called from the docker event loop); the
// stale list keeps answering until the next lookup triggers a refresh.
func ForgetLocalNetworkSubnets() {
	localSubnetsMu.Lock()
	defer localSubnetsMu.Unlock()
	localSubnetsExpires = time.Time{}
}

// localNetworkSubnets returns the cached CIDRs of host-local bridge networks. Never blocks on
// docker: an expired cache is served as-is and refreshed by one background goroutine.
func localNetworkSubnets() []*net.IPNet {
	localSubnetsMu.Lock()
	defer localSubnetsMu.Unlock()

	if !time.Now().Before(localSubnetsExpires) && !localSubnetsRefreshing {
		localSubnetsRefreshing = true
		go refreshLocalNetworkSubnets()
	}
	return localSubnets
}

// refreshLocalNetworkSubnets lists docker networks and replaces the cache; on error the previous answer is kept.
func refreshLocalNetworkSubnets() {
	defer func() {
		localSubnetsMu.Lock()
		localSubnetsRefreshing = false
		localSubnetsMu.Unlock()
	}()

	if err := Connect(); err != nil {
		return
	}

	networks, err := DockerClient.NetworkList(DockerContext, types.NetworkListOptions{})
	if err != nil {
		utils.Error("Docker - Cannot list networks for the local subnet cache", err)
		return
	}

	subnets := []*net.IPNet{}
	for _, network := range networks {
		// only a host-local bridge is all-local containers; macvlan/overlay subnets include remote hosts
		if network.Driver != "bridge" || network.Scope != "local" {
			continue
		}
		for _, conf := range network.IPAM.Config {
			if conf.Subnet == "" {
				continue
			}
			if _, cidr, err := net.ParseCIDR(conf.Subnet); err == nil {
				subnets = append(subnets, cidr)
			}
		}
	}

	localSubnetsMu.Lock()
	localSubnets = subnets
	localSubnetsExpires = time.Now().Add(localSubnetsTTL)
	localSubnetsMu.Unlock()
}

// IsLocalDockerIP reports whether ip sits on a network of this docker host; wired into utils.IsLocalDockerIP at startup.
func IsLocalDockerIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}

	for _, subnet := range localNetworkSubnets() {
		if subnet.Contains(parsed) {
			return true
		}
	}
	return false
}
