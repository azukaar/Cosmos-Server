package constellation

import (
	"strings"
	"strconv"
	"time"
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


func ClientHeartbeatInit() {
	for i := 0; i < 20; i++ {
		time.Sleep(3 * time.Second)
		
		_, err := js.KeyValue("constellation-nodes")
		if err == nil {
			utils.Log("[NATS] Connected to existing Key-Value store 'constellation-nodes'")
			break
		}
		
		_, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket:   "constellation-nodes",
			TTL:      10 * time.Second,
			Storage:  nats.MemoryStorage,
			Replicas: 1,
		})

		if err == nil {
			utils.Log("[NATS] Created Key-Value store 'constellation-nodes'")
			break
		}
		
		utils.Debug("[NATS] JetStream not ready, retrying... " + err.Error())
	}
	
	utils.Debug("[NATS] Key-Value store 'constellation-nodes' ready")

	go UpdateLocalTunnelCache();

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			ClientConnectToJS()

			kv, err := js.KeyValue("constellation-nodes")
			if err != nil {
				utils.Warn("[NATS] NATS client not connected getting Key-Value store during heartbeat, store is offline will skip this cycle " + err.Error())
				continue
			}

			device, err := GetCurrentDevice()
			if err != nil {
				utils.Warn("[NATS] NATS client not connected getting current device for heartbeat " + err.Error())
				continue
			}

			key := sanitizeNATSUsername(device.DeviceName)

			if !IsClientConnected() {
				utils.Warn("[NATS] NATS client not connected during heartbeat")
				InitNATSClient()
			}

			// insert JSON encoded heartbeat
			heartbeat := NodeHeartbeat{
				DeviceName: device.DeviceName,
				IP: device.IP,
				IsRelay: device.IsRelay,
				IsLighthouse: device.IsLighthouse,
				IsExitNode: device.IsExitNode,
				IsCosmosNode: device.IsCosmosNode,
				Tunnels: GetAllTunneledRoutes(),
			}

			heartbeatData, err := json.Marshal(heartbeat)
			if err != nil {
				utils.Error("[NATS] Error marshalling heartbeat JSON", err)
				continue
			}

			_, err = kv.Put(key, heartbeatData)
			if err != nil {
				utils.Error("[NATS] Error updating heartbeat in Key-Value store", err)
			}
		}
	}()
}

var localTunnelCache []utils.ConstellationTunnel
var localTunnelCacheMutex = &sync.RWMutex{}
var lastCacheUpdate time.Time

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

	kv, err := js.KeyValue("constellation-nodes")
	if err != nil {
		utils.Error("[NATS] Error getting Key-Value store during tunnel cache update, store is offline will skip this cycle", err)
		return
	}


	keys, err := kv.Keys()
	if err != nil {
		utils.Error("[NATS] Error getting keys from Key-Value store during tunnel cache update", err)
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
			if tunnelRoute.Tunnel == "_ANY_" || tunnelRoute.Tunnel == currentDeviceName {
				if existing, ok := byName[tunnelRoute.Name]; ok {
					existing.From = append(existing.From, heartbeat.DeviceName)
				} else {
					tunnelRoute._IsTunneled = true
					byName[tunnelRoute.Name] = &utils.ConstellationTunnel{
						Route: tunnelRoute,
						From:  []string{heartbeat.DeviceName},
					}
				}
			}
		}
	}

	tunnels := make([]utils.ConstellationTunnel, 0, len(byName))
	for _, t := range byName {
		tunnels = append(tunnels, *t)
	}

	localTunnelCache = tunnels	
	lastCacheUpdate = time.Now()
}

func GetLocalTunnelCache() []utils.ConstellationTunnel {
	localTunnelCacheMutex.RLock()
	defer localTunnelCacheMutex.RUnlock()

	result := make([]utils.ConstellationTunnel, len(localTunnelCache))
	copy(result, localTunnelCache)

	if time.Since(lastCacheUpdate) > 5*time.Second {
		go UpdateLocalTunnelCache()
	}

	return result
}