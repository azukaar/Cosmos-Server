package constellation

import (
	"strings"
	"strconv"
	"time"
	"sort"
	"github.com/nats-io/nats.go"
	"encoding/json"
	"sync"

	"github.com/azukaar/cosmos-server/src/utils"
)

func GetAllTunneledRoutes() []utils.ProxyRouteConfig {
	// list routes with a tunnel property matching the device name
	routesList := utils.GetMainConfig().HTTPConfig.ProxyConfig.Routes
	tunnels := []utils.ProxyRouteConfig{}

	thisIp, err := GetCurrentDeviceIP()
	if err != nil {
		utils.Error("Error getting current device IP for tunneled routes", err)
		return tunnels
	}
	
	_, configHostport := utils.GetServerRawAccess()

	for _, route := range routesList {
		if route.Tunnel != "" {
			protocol := ""
			port := configHostport

			if route.UseHost {
				route.OverwriteHostHeader = route.Host

				// extract port from host if it's a number
				if strings.Contains(route.Host, ":") {
					_port := strings.Split(route.Host, ":")[1]
					if _, err := strconv.Atoi(_port); err == nil {
						port = _port
					}
				}

				// Extract ANY protocol from host, if empty leave empty
				if idx := strings.Index(route.Host, "://"); idx != -1 {
					protocol = route.Host[:idx+3]
					route.Host = route.Host[idx+3:]
				}
			} else {
				protocol = "http://"
			}

			route.Target = protocol + thisIp + ":" + port
			route.Mode = "PROXY"

			if route.TunneledHost != "" {
				route.Host = route.TunneledHost
			}

			tunnels = append(tunnels, route)
		}
	}

	return tunnels
}


func StopHeartbeat() {
	if heartbeatStopChan != nil {
		close(heartbeatStopChan)
		heartbeatStopChan = nil
	}
	if heartbeatTicker != nil {
		heartbeatTicker.Stop()
		heartbeatTicker = nil
	}
}

func ClientHeartbeatInit() {
	// Stop any existing heartbeat
	StopHeartbeat()

	for i := 0; i < 20; i++ {
		time.Sleep(3 * time.Second)

		// Reconnect JetStream if needed
		err := ClientConnectToJS()
		if err != nil {
			utils.Debug("[NATS] JetStream not connected, retrying... " + err.Error())
			continue
		}

		// Hold read lock for entire KV operation
		clientConfigLock.RLock()
		if js == nil {
			clientConfigLock.RUnlock()
			utils.Debug("[NATS] JetStream context is nil, retrying...")
			continue
		}

		_, err = js.KeyValue("constellation-nodes")
		if err == nil {
			clientConfigLock.RUnlock()
			utils.Log("[NATS] Connected to existing Key-Value store 'constellation-nodes'")
			break
		}

		_, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket:   "constellation-nodes",
			TTL:      10 * time.Second,
			Storage:  nats.MemoryStorage,
			Replicas: 1,
		})
		clientConfigLock.RUnlock()

		if err == nil {
			utils.Log("[NATS] Created Key-Value store 'constellation-nodes'")
			break
		}

		utils.Debug("[NATS] JetStream not ready, retrying... " + err.Error())
	}

	utils.Debug("[NATS] Key-Value store 'constellation-nodes' ready")

	go UpdateLocalTunnelCache();

	heartbeatStopChan = make(chan struct{})
	heartbeatTicker = time.NewTicker(2 * time.Second)

	// Capture in local variables to avoid race conditions
	stopChan := heartbeatStopChan
	ticker := heartbeatTicker

	go func() {
		for {
			select {
			case <-stopChan:
				utils.Log("[NATS] Heartbeat stopped")
				return
			case <-ticker.C:
				err := ClientConnectToJS()
				if err != nil {
					utils.Warn("[NATS] Error connecting to JetStream during heartbeat: " + err.Error())
					continue
				}

				// check connected status first
				if !ConstellationConnected() {
					utils.Warn("[NATS] Constellation not connected during heartbeat, stopping heartbeat")
					ticker.Stop()
					return
				}

				// Prepare heartbeat data outside the lock
				device, err := GetCurrentDevice()
				if err != nil {
					utils.Warn("[NATS] NATS client not connected getting current device for heartbeat " + err.Error())
					continue
				}

				key := sanitizeNATSUsername(device.DeviceName)

				heartbeat := NodeHeartbeat{
					DeviceName: device.DeviceName,
					IP: device.IP,
					IsRelay: device.IsRelay,
					IsLighthouse: device.IsLighthouse,
					IsExitNode: device.IsExitNode,
					CosmosNode: device.CosmosNode,
					Tunnels: GetAllTunneledRoutes(),
				}

				heartbeatData, err := json.Marshal(heartbeat)
				if err != nil {
					utils.Error("[NATS] Error marshalling heartbeat JSON", err)
					continue
				}

				// Hold read lock for entire KV operation
				clientConfigLock.RLock()
				if js == nil {
					clientConfigLock.RUnlock()
					utils.Warn("[NATS] JetStream context is nil during heartbeat, skipping cycle")
					continue
				}

				kv, err := js.KeyValue("constellation-nodes")
				if err != nil {
					clientConfigLock.RUnlock()
					utils.Warn("[NATS] NATS client not connected getting Key-Value store during heartbeat, store is offline will skip this cycle " + err.Error())
					continue
				}

				_, err = kv.Put(key, heartbeatData)
				clientConfigLock.RUnlock()

				if err != nil {
					utils.Error("[NATS] Error updating heartbeat in Key-Value store", err)
				}
			}
		}
	}()
}

var localTunnelCache []utils.ConstellationTunnel
var localTunnelCacheMutex = &sync.RWMutex{}
var lastCacheUpdate time.Time
var heartbeatStopChan chan struct{}
var heartbeatTicker *time.Ticker

func UpdateLocalTunnelCache() {
	localTunnelCacheMutex.Lock()
	defer localTunnelCacheMutex.Unlock()
	
	currentDeviceName, err := GetCurrentDeviceName()
	if err != nil {
		utils.Error("[constellation] Failed to get current device name for tunnel cache update", err)
		return
	}

	err = ClientConnectToJS()
	if err != nil {
		utils.Error("[NATS] Error connecting to JetStream during tunnel cache update", err)
		return
	}

	// Hold read lock for KV operations
	clientConfigLock.RLock()
	if js == nil {
		clientConfigLock.RUnlock()
		utils.Error("[NATS] JetStream context is nil during tunnel cache update", nil)
		return
	}

	kv, err := js.KeyValue("constellation-nodes")
	if err != nil {
		clientConfigLock.RUnlock()
		utils.Error("[NATS] Error getting Key-Value store during tunnel cache update, store is offline will skip this cycle", err)
		return
	}

	keys, err := kv.Keys()
	if err != nil {
		clientConfigLock.RUnlock()
		utils.Warn("[NATS] Error getting keys from Key-Value store during tunnel cache update "+ err.Error())
		return
	}

	byName := map[string]*utils.ConstellationTunnel{}

	for _, key := range keys {
		entry, err := kv.Get(key)
		if err != nil {
			utils.Error("[NATS] Error getting entry from Key-Value store during tunnel cache update", err)
			continue
		}

		var heartbeat NodeHeartbeat
		err = json.Unmarshal(entry.Value(), &heartbeat)
		if err != nil {
			utils.Error("[NATS] Error unmarshalling heartbeat JSON during tunnel cache update", err)
			continue
		}

		for _, tunnelRoute := range heartbeat.Tunnels {
			// Skip tunnels from ourselves
			if heartbeat.DeviceName == currentDeviceName {
				continue
			}
			if tunnelRoute.Tunnel == "_ANY_" || tunnelRoute.Tunnel == currentDeviceName {
				if existing, ok := byName[tunnelRoute.Name]; ok {
					existing.From = append(existing.From, heartbeat.DeviceName)
				} else {
					tunnelRoute.Const_IsTunneled = true
					byName[tunnelRoute.Name] = &utils.ConstellationTunnel{
						Route: tunnelRoute,
						From:  []string{heartbeat.DeviceName},
					}
				}
			}
		}
	}
	clientConfigLock.RUnlock() // Done with KV operations

	tunnels := make([]utils.ConstellationTunnel, 0, len(byName))
	for _, t := range byName {
		tunnels = append(tunnels, *t)
	}

	// Compare old and new cache using sorted copies for consistent comparison
	sortTunnelsForComparison := func(t []utils.ConstellationTunnel) []utils.ConstellationTunnel {
		copied := make([]utils.ConstellationTunnel, len(t))
		for i, tunnel := range t {
			copied[i] = tunnel
			copied[i].From = make([]string, len(tunnel.From))
			copy(copied[i].From, tunnel.From)
			sort.Strings(copied[i].From)
		}
		sort.Slice(copied, func(i, j int) bool {
			return copied[i].Route.Name < copied[j].Route.Name
		})
		return copied
	}

	oldJSON, _ := json.Marshal(sortTunnelsForComparison(localTunnelCache))
	newJSON, _ := json.Marshal(sortTunnelsForComparison(tunnels))

	localTunnelCache = tunnels
	lastCacheUpdate = time.Now()

	if string(oldJSON) != string(newJSON) {
		utils.Log("[constellation] Tunnel cache changed, restarting HTTP server...")
		go utils.RestartHTTPServer()
	}
}

func GetLocalTunnelCache() []utils.ConstellationTunnel {
	isLB, err := GetCurrentDeviceIsLoadbalancer()
	if err != nil {
		utils.Debug("[constellation] Failed to get current device load balancer status for tunnel cache retrieval " + err.Error())
		return []utils.ConstellationTunnel{}
	}
	
	if !isLB {
		return []utils.ConstellationTunnel{}
	}

	localTunnelCacheMutex.RLock()
	defer localTunnelCacheMutex.RUnlock()

	result := make([]utils.ConstellationTunnel, len(localTunnelCache))
	copy(result, localTunnelCache)

	if time.Since(lastCacheUpdate) > 5*time.Second {
		go UpdateLocalTunnelCache()
	}

	return result
}

func IsTunneled(route utils.ProxyRouteConfig) bool {
	return route.Const_IsTunneled
}