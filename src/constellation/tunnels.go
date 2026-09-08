package constellation

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/pro"
	"github.com/azukaar/cosmos-server/src/utils"
)

// getNATSReplicas derives the bucket replication factor from the declared
// topology: R3 iff the constellation was created in HA mode, else R1.
func getNATSReplicas() int {
	if IsNATSHA() {
		return 3
	}
	return 1
}

// createKVAtTopology creates a KV bucket at the declared replication factor
// (R3 in HA, R1 otherwise); creation fails while an HA cluster is still
// forming and callers retry. A bucket at the wrong replica count or storage
// type (see nodesKVStorage) is dropped and recreated — these are ephemeral
// caches with short TTLs.
func createKVAtTopology(cfg nats.KeyValueConfig) (nats.KeyValue, error) {
	cfg.Replicas = getNATSReplicas()
	kv, err := js.CreateKeyValue(&cfg)
	if err == nil {
		return kv, nil
	}
	if si, errInfo := js.StreamInfo("KV_" + cfg.Bucket); errInfo == nil && kvTopologyOutdated(si.Config, cfg) {
		utils.Warn("[NATS] KV '" + cfg.Bucket + "' exists as " + describeKVTopology(si.Config.Storage, si.Config.Replicas) +
			" instead of " + describeKVTopology(cfg.Storage, cfg.Replicas) + ", recreating at the declared topology")
		if errDel := js.DeleteKeyValue(cfg.Bucket); errDel != nil {
			return nil, errDel
		}
		return js.CreateKeyValue(&cfg)
	}
	return nil, err
}

// kvTopologyOutdated compares only the fields that decide survival; TTL or
// description drift is not worth dropping a bucket over.
func kvTopologyOutdated(have nats.StreamConfig, want nats.KeyValueConfig) bool {
	return have.Storage != want.Storage || have.Replicas != want.Replicas
}

func describeKVTopology(storage nats.StorageType, replicas int) string {
	name := "memory"
	if storage == nats.FileStorage {
		name = "file"
	}
	return name + "/R" + strconv.Itoa(replicas)
}

// kvCreatorFallbackAfter: how long a non-designated manager waits on a missing
// bucket before assuming the designated creator is down and creating it itself.
const kvCreatorFallbackAfter = 60 * time.Second

// designatedKVCreator elects, deterministically from the local device cache, the
// manager (lowest device name, CosmosNode == 2) that owns creating the shared
// 'constellation-nodes' bucket. Best-effort: the liveness fallback and a stale/empty
// cache can still allow a second creator — the divergence sweep is the backstop.
func designatedKVCreator() (creator string, isSelf bool) {
	device, err := GetCurrentDevice()
	if err != nil {
		// cannot identify ourselves: claim creatorship rather than deadlock
		return "", true
	}
	self := device.DeviceName
	creator = self
	devices, _ := deviceCacheSnapshot()
	for _, d := range devices {
		if d.CosmosNode != 2 || d.DeviceName == "" {
			continue
		}
		if d.DeviceName < creator {
			creator = d.DeviceName
		}
	}
	return creator, creator == self
}

// nodesKVStorage: file-backed in HA, memory otherwise. Memory is wrong for a
// clustered R3 bucket: a restarted memory replica has nothing to elect around
// while the file-backed stream assignment survives, so a rolling update leaves
// a stream everyone can resolve and nobody can write — file replicas rejoin
// from their own logs and re-elect.
func nodesKVStorage() nats.StorageType {
	if IsNATSHA() {
		return nats.FileStorage
	}
	return nats.MemoryStorage
}

// nodesKVConfig: the 'constellation-nodes' bucket shape — an ephemeral 10s-TTL
// cache repopulated by every node's 2s heartbeat, so drop-and-recreate is safe.
func nodesKVConfig() nats.KeyValueConfig {
	return nats.KeyValueConfig{
		Bucket:  "constellation-nodes",
		TTL:     10 * time.Second,
		Storage: nodesKVStorage(),
	}
}

func GetAllTunneledRoutes() []utils.ProxyRouteConfig {
	// list routes with a tunnel property matching the device name
	routesList := utils.GetMainConfig().HTTPConfig.ProxyConfig.Routes
	tunnels := []utils.ProxyRouteConfig{}

	thisIp, err := GetCurrentDeviceIP()
	if err != nil {
		utils.Error("Error getting current device IP for tunneled routes", err)
		return tunnels
	}

	serverProtocol, _, configHostport := utils.GetServerRawAccess()

	for _, route := range routesList {
		if route.Tunnel != "" {
			// checked against the original target: route.Target is rewritten below
			if !isTunnelBackendHealthy(route) {
				continue
			}

			protocol := ""
			port := configHostport

			if route.UseHost {
				if tunnelHostOverridden(route) {
					route.OverwriteHostHeader = route.Host
				} else {
					route.OverwriteHostHeader = ""
				}

				// extract port from host if it's a number
				if strings.Contains(route.Host, ":") {
					_port := strings.Split(route.Host, ":")[1]
					if _, err := strconv.Atoi(_port); err == nil {
						port = _port
					}
				}

				// Extract ANY protocol from target, if empty leave empty
				if idx := strings.Index(route.Target, "://"); idx != -1 {
					protocol = route.Target[:idx+3]
				}
			} else {
				route.OverwriteHostHeader = ""
				// TODO: This wont work needs to copy setting
				protocol = "http://"
			}

			// if protocol is https, skip certificate check for tunnel VPN IP
			if protocol == "https://" || protocol == "http://" {
				protocol = serverProtocol
				if protocol == "https://" {
					route.AcceptInsecureHTTPSTarget = true
				}
			}

			route.Target = protocol + thisIp

			if port != "" {
				route.Target += ":" + port
			}

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
	// Scheduler must stop before heartbeat: it depends on the NATS client and
	// KV buckets that the heartbeat tears down.
	StopSchedulerInConstellation()
	// Sampler is independent of the scheduler (runs on every node regardless
	// of leader role) but we stop it here because the heartbeat loop is its
	// only consumer in production.
	pro.StopResourceSampler()

	StopTunnelProber()

	heartbeatLock.Lock()
	defer heartbeatLock.Unlock()

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

	// Only managers create buckets — an agent's unsynced defaults could win
	// the creation race with the wrong replica count; agents just wait.
	isAgent := isAgentNode()

	// Among managers, only the designated creator normally creates; the
	// others wait, with a liveness fallback after kvCreatorFallbackAfter of
	// the bucket being continuously missing. The loop is sized so a
	// non-creator can actually outwait that window (checks every 3s).
	var nodesKVMissingSince time.Time

	for i := 0; i < 40; i++ {
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

		if isAgent {
			clientConfigLock.RUnlock()
			utils.Debug("[NATS] Waiting for a manager to create the Key-Value store... " + err.Error())
			continue
		}

		if nodesKVMissingSince.IsZero() {
			nodesKVMissingSince = time.Now()
		}

		// Single-creator guard (see designatedKVCreator): non-designated
		// managers hold off unless the creator has left the bucket missing
		// past the liveness window.
		creator, isCreator := designatedKVCreator()
		if !isCreator && time.Since(nodesKVMissingSince) < kvCreatorFallbackAfter {
			clientConfigLock.RUnlock()
			utils.Debug("[NATS] Waiting for designated creator '" + creator + "' to create 'constellation-nodes'...")
			continue
		}
		if !isCreator {
			utils.Warn("[NATS] Designated creator '" + creator + "' appears down ('constellation-nodes' missing for " +
				time.Since(nodesKVMissingSince).Round(time.Second).String() + "), creating the bucket from this node instead")
		}

		_, err = createKVAtTopology(nodesKVConfig())
		if err == nil {
			// pre-create the sticky bucket too, so agents (which never
			// create buckets) can use it even when the manager is not a LB
			if _, errSticky := ensureStickyBucket(); errSticky != nil {
				utils.Warn("[NATS] Could not pre-create tunnel-sticky store: " + errSticky.Error())
			}
		}
		clientConfigLock.RUnlock()

		if err == nil {
			utils.Log("[NATS] Created Key-Value store 'constellation-nodes'")
			break
		}

		utils.Debug("[NATS] JetStream not ready, retrying... " + err.Error())
	}

	pro.ClientHeartbeatInit(&clientConfigLock, js, getNATSReplicas())

	utils.Debug("[NATS] Key-Value store 'constellation-nodes' ready")

	// Resource sampler: caches CPU/memory-pressure samples for the heartbeat builder, scheduler, and load-based LB.
	pro.StartResourceSampler()

	// Scheduler: runs the deployment reconciler on whichever node wins the
	// leader election in constellation-nodes. Must be started after
	// pro.ClientHeartbeatInit so the deployments KV exists.
	StartSchedulerInConstellation()

	// advertise a tunneled route only while its local backend answers
	StartTunnelProber()

	UpdateLocalTunnelCache()

	heartbeatLock.Lock()
	heartbeatStopChan = make(chan struct{})
	heartbeatTicker = time.NewTicker(2 * time.Second)

	// Capture in local variables to avoid race conditions
	stopChan := heartbeatStopChan
	ticker := heartbeatTicker
	heartbeatLock.Unlock()

	// Watch KV for changes and refresh tunnel cache. Outer loop re-establishes the
	// watcher when its stream goes away (e.g. the divergence cure deletes and
	// recreates the bucket) instead of dying silently.
	go func() {
		for {
			select {
			case <-stopChan:
				utils.Log("[NATS] KV watcher stopped")
				return
			default:
			}

			watcher := func() nats.KeyWatcher {
				if err := ClientConnectToJS(); err != nil {
					utils.Debug("[NATS] Error connecting to JetStream for KV watcher: " + err.Error())
					return nil
				}
				clientConfigLock.RLock()
				defer clientConfigLock.RUnlock()
				if js == nil {
					utils.Debug("[NATS] JetStream context is nil for KV watcher")
					return nil
				}
				kv, err := js.KeyValue("constellation-nodes")
				if err != nil {
					utils.Debug("[NATS] Error getting KV store for watcher: " + err.Error())
					return nil
				}
				w, err := kv.WatchAll()
				if err != nil {
					utils.Debug("[NATS] Error creating KV watcher: " + err.Error())
					return nil
				}
				return w
			}()

			if watcher != nil {
				utils.Log("[NATS] KV watcher started for tunnel cache updates")
			watch:
				for {
					select {
					case <-stopChan:
						watcher.Stop()
						utils.Log("[NATS] KV watcher stopped")
						return
					case entry, ok := <-watcher.Updates():
						if !ok {
							// stream deleted/recreated (divergence cure) or connection lost
							watcher.Stop()
							utils.Warn("[NATS] KV watcher channel closed, re-establishing...")
							break watch
						}
						if entry == nil {
							// nil marks end of initial values
							continue
						}
						GetLocalTunnelCache()
					}
				}
			}

			select {
			case <-stopChan:
				utils.Log("[NATS] KV watcher stopped")
				return
			case <-time.After(10 * time.Second):
			}
		}
	}()

	go func() {
		// selfHealMissingSince tracks how long this manager has continuously
		// seen the bucket missing, gating the non-creator recreation fallback
		// exactly like the init loop above.
		var selfHealMissingSince time.Time

		// Divergence detection (F8): after a partition merge JetStream can be
		// left with two same-name incarnations of the constellation-nodes
		// stream that never reconcile; the reliable signature is kv.Keys()
		// returning the same key more than once. Checked every 10th tick
		// (~20s); acted on only after 3 CONSECUTIVE positives (~1 minute) so
		// a transient mid-merge wobble never triggers the cure.
		tickCount := 0
		divergedChecks := 0

		// consecutive failed writes, for the unwritable-bucket alarm at the Put below
		writeFailures := 0

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

				key := device.DeviceName

				// Docker is authoritative for "what is running here" — query
				// containers labeled cosmos-deployment each tick rather than
				// tracking in-process state. Failures degrade gracefully to
				// an empty list; the scheduler's safety-net reconcile will
				// retry within 1 minute.
				running, rerr := docker.ListDeploymentNamesRunningHere()
				if rerr != nil {
					utils.Warn("[SCHED-NODE] failed to list cosmos-deployment containers for heartbeat: " + rerr.Error())
					running = nil
				}

				// Per-deployment spec version those containers were created from,
				// so the scheduler can spot a node running a stale spec. Degrades
				// to nil on error like the name list above.
				runningVersions, vrerr := docker.ListDeploymentVersionsRunningHere()
				if vrerr != nil {
					utils.Warn("[SCHED-NODE] failed to list cosmos-deployment versions for heartbeat: " + vrerr.Error())
					runningVersions = nil
				}

				// Separate field, never folded into `running`: the scheduler removes
				// entries there it has no matching deployment for.
				managedDBs, mdberr := pro.ListManagedDBsRunningHere()
				if mdberr != nil {
					utils.Warn("[MDB] failed to list cosmos-managed-db containers for heartbeat: " + mdberr.Error())
					managedDBs = nil
				}

				// Same separate-field rationale as ManagedDBs.
				seaweedFS, swfserr := pro.ListSeaweedFSMastersRunningHere()
				if swfserr != nil {
					utils.Warn("[SWFS] failed to list cosmos-seaweedfs containers for heartbeat: " + swfserr.Error())
					seaweedFS = nil
				}

				// Derived from access records and tags, not docker; read before
				// the config lock is taken below — the helper takes it too.
				registries := pro.RegistryAccessesServedHere(&clientConfigLock, js, device.Tags)

				// Read the latest cached resource sample. Cheap — the
				// sampler's own goroutine paid the cpu.Percent cost.
				res := pro.GetCurrentResources()

				heartbeat := NodeHeartbeat{
					DeviceName:                device.DeviceName,
					IP:                        device.IP,
					IsRelay:                   device.IsRelay,
					IsLighthouse:              device.IsLighthouse,
					IsExitNode:                device.IsExitNode,
					CosmosNode:                device.CosmosNode,
					Tunnels:                   GetAllTunneledRoutes(),
					Hostnames:                 localHostnames(),
					RunningDeployments:        running,
					RunningDeploymentVersions: runningVersions,
					ManagedDBs:                managedDBs,
					SeaweedFS:                 seaweedFS,
					Registries:                registries,
					CPUPercent:                res.CPUPercent,
					RAMPercent:                res.RAMPercent,
					MonitoringOn:              res.MonitoringOn,
					Tags:                      device.Tags,
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
				if err != nil && !isAgent {
					// self-heal: bucket missing (HA cluster still forming at
					// init, or memory stream wiped by a cluster restart).
					// Single-creator guard (F3): every-manager recreation is
					// how partitioned islands each grew their own diverged
					// copy of the stream — only the designated creator heals
					// immediately, other managers give it the liveness window
					// before assuming it's down.
					if selfHealMissingSince.IsZero() {
						selfHealMissingSince = time.Now()
					}
					creator, isCreator := designatedKVCreator()
					if isCreator || time.Since(selfHealMissingSince) >= kvCreatorFallbackAfter {
						if !isCreator {
							utils.Warn("[NATS] Designated creator '" + creator + "' appears down ('constellation-nodes' missing for " +
								time.Since(selfHealMissingSince).Round(time.Second).String() + "), recreating the bucket from this node instead")
						}
						kv, err = createKVAtTopology(nodesKVConfig())
						if err == nil {
							utils.Log("[NATS] Recreated Key-Value store 'constellation-nodes'")
						}
					} else {
						utils.Debug("[NATS] 'constellation-nodes' missing, waiting for designated creator '" + creator + "' to recreate it")
					}
				}
				if err != nil {
					clientConfigLock.RUnlock()
					utils.Warn("[NATS] JetStream KV store unavailable during heartbeat, skipping cycle: " + err.Error())
					continue
				}
				selfHealMissingSince = time.Time{}

				_, err = kv.Put(key, heartbeatData)
				if err != nil {
					utils.Error("[NATS] Error updating heartbeat in Key-Value store", err)

					// The bucket can resolve (off the file-backed meta layer) while its
					// leaderless stream refuses every write — invisible to
					// checkNodesKVDivergence. Name it once instead of a wall of per-tick errors.
					writeFailures++
					if writeFailures == nodesKVWriteFailureStreak {
						utils.MajorError("[NATS] 'constellation-nodes' has refused every write for "+
							(time.Duration(writeFailures)*2*time.Second).String()+
							" while still resolving: its stream has no leader. NO node reads as live cluster-wide while this lasts — "+
							"every deployment shows Degraded and the scheduler will not place anything. "+
							"Recreating the bucket, or restarting the managers together, clears it.", err)
					}
				} else {
					writeFailures = 0
				}

				// Divergence sweep after the Put so a just-applied cure never
				// invalidates the handle we are about to write with.
				tickCount++
				if tickCount%10 == 0 {
					checkNodesKVDivergence(kv, &divergedChecks)
				}

				clientConfigLock.RUnlock()
			}
		}
	}()
}

// nodesKVCureCooldown throttles the divergence cure so a divergence that survives
// (or immediately recurs after) a cure alarms instead of thrashing the bucket.
const nodesKVCureCooldown = 10 * time.Minute

// lastNodesKVCure marks the last cure attempt; only touched from the heartbeat goroutine.
var lastNodesKVCure time.Time

// nodesKVWriteFailureStreak: consecutive failed heartbeat writes before the
// bucket is called unwritable (~30s at the 2s tick). Deliberately an alarm,
// not a drop-and-recreate: the same symptom is genuine quorum loss, which
// heals by itself when the peers return.
const nodesKVWriteFailureStreak = 15

// checkNodesKVDivergence detects duplicate keys from kv.Keys() — the signature of two
// unreconciled same-name stream incarnations after a partition merge, which JetStream
// never repairs on its own — and, after 3 consecutive positives (~1 min), has the
// designated creator delete and recreate the bucket (safe: ephemeral heartbeat cache).
// Must be called with clientConfigLock read-held.
func checkNodesKVDivergence(kv nats.KeyValue, consecutive *int) {
	keys, err := kv.Keys()
	if err != nil {
		// nats.ErrNoKeysFound / transport errors say nothing about
		// divergence — count as a clean check
		*consecutive = 0
		return
	}

	seen := map[string]bool{}
	dup := ""
	for _, k := range keys {
		if seen[k] {
			dup = k
			break
		}
		seen[k] = true
	}
	if dup == "" {
		*consecutive = 0
		return
	}

	*consecutive++
	utils.Warn("[NATS] 'constellation-nodes' returned duplicate key '" + dup + "' (" + strconv.Itoa(*consecutive) + "/3 consecutive checks before cure)")
	if *consecutive < 3 {
		return
	}
	// reset so a failed/foreign cure re-arms a full 3-check window
	*consecutive = 0

	creator, isCreator := designatedKVCreator()
	if !isCreator {
		utils.Warn("[NATS] 'constellation-nodes' is DIVERGED (duplicate keys); waiting for designated creator '" + creator + "' to apply the cure")
		return
	}

	utils.MajorError("[NATS] 'constellation-nodes' is DIVERGED: duplicate keys across merged JetStream stream incarnations "+
		"(this is the two-scheduler-leaders / phantom-placements state). Attempting the automated cure.", nil)

	// cooldown: a divergence that survives a cure should alarm, not thrash the bucket
	if !lastNodesKVCure.IsZero() && time.Since(lastNodesKVCure) < nodesKVCureCooldown {
		utils.MajorError("[NATS] Divergence cure already attempted "+time.Since(lastNodesKVCure).Round(time.Second).String()+
			" ago and 'constellation-nodes' is diverged again — NOT retrying within cooldown, manual intervention likely needed", nil)
		return
	}
	lastNodesKVCure = time.Now()

	utils.Warn("[NATS] Attempting automated divergence cure: deleting and recreating 'constellation-nodes'")
	if errDel := js.DeleteKeyValue("constellation-nodes"); errDel != nil {
		utils.MajorError("[NATS] Divergence cure FAILED to delete 'constellation-nodes'", errDel)
		return
	}
	if _, errCreate := createKVAtTopology(nodesKVConfig()); errCreate != nil {
		utils.MajorError("[NATS] Divergence cure deleted 'constellation-nodes' but FAILED to recreate it (the heartbeat self-heal will retry)", errCreate)
		return
	}
	utils.Log("[NATS] Divergence cure applied: 'constellation-nodes' recreated from a single seed; heartbeats repopulate it within ~2s")
}

var localTunnelCache []utils.ConstellationTunnel
var localTunnelCacheMutex = &sync.RWMutex{}
var lastCacheUpdate time.Time
var heartbeatStopChan chan struct{}
var heartbeatTicker *time.Ticker
var heartbeatLock sync.Mutex

// mergeTunnelHeartbeats collapses every advertiser's tunnel routes into one
// entry per route name. The governing Route config is the one advertised by the
// lowest device name (same tiebreak as designatedKVCreator): picking the first
// seen made the effective config flap with KV iteration order.
func mergeTunnelHeartbeats(heartbeats []NodeHeartbeat, currentDeviceName string) []utils.ConstellationTunnel {
	byName := map[string]*utils.ConstellationTunnel{}
	governing := map[string]string{}

	for _, heartbeat := range heartbeats {
		advertiser := heartbeat.DeviceName

		for _, tunnelRoute := range heartbeat.Tunnels {
			if tunnelRoute.Tunnel != "_ANY_" && tunnelRoute.Tunnel != currentDeviceName {
				continue
			}

			target := utils.TunnelTarget{
				DeviceName:   heartbeat.DeviceName,
				TargetURL:    tunnelRoute.Target,
				CPUPercent:   heartbeat.CPUPercent,
				RAMPercent:   heartbeat.RAMPercent,
				MonitoringOn: heartbeat.MonitoringOn,
			}
			tunnelRoute.Const_IsTunneled = true

			existing, ok := byName[tunnelRoute.Name]
			if !ok {
				byName[tunnelRoute.Name] = &utils.ConstellationTunnel{
					Route:   tunnelRoute,
					Targets: []utils.TunnelTarget{target},
				}
				governing[tunnelRoute.Name] = advertiser
				continue
			}

			existing.Targets = append(existing.Targets, target)
			// an unnamed advertiser never governs, and never blocks a named one
			if advertiser != "" && (governing[tunnelRoute.Name] == "" || advertiser < governing[tunnelRoute.Name]) {
				existing.Route = tunnelRoute
				governing[tunnelRoute.Name] = advertiser
			}
		}
	}

	tunnels := make([]utils.ConstellationTunnel, 0, len(byName))
	for _, t := range byName {
		// Ensure local node is always first in targets
		for i, target := range t.Targets {
			if target.DeviceName == currentDeviceName && i != 0 {
				t.Targets[0], t.Targets[i] = t.Targets[i], t.Targets[0]
				break
			}
		}
		tunnels = append(tunnels, *t)
	}

	return tunnels
}

// sortTunnelsForComparison canonicalizes a tunnel list for change detection:
// stable order, per-target load samples zeroed (they feed the "load_based" LB
// at request time, and comparing them restarted every LB on every sample).
// MonitoringOn stays in the comparison: monitoring off is a real change.
func sortTunnelsForComparison(t []utils.ConstellationTunnel) []utils.ConstellationTunnel {
	copied := make([]utils.ConstellationTunnel, len(t))
	for i, tunnel := range t {
		copied[i] = tunnel
		copied[i].Targets = make([]utils.TunnelTarget, len(tunnel.Targets))
		copy(copied[i].Targets, tunnel.Targets)
		for j := range copied[i].Targets {
			copied[i].Targets[j].CPUPercent = 0
			copied[i].Targets[j].RAMPercent = 0
		}
		sort.Slice(copied[i].Targets, func(a, b int) bool {
			return copied[i].Targets[a].DeviceName < copied[i].Targets[b].DeviceName
		})
	}
	sort.Slice(copied, func(i, j int) bool {
		return copied[i].Route.Name < copied[j].Route.Name
	})
	return copied
}

// tunnelsForNode gates serving on load-balancer status: only LBs serve tunnels,
func tunnelsForNode(heartbeats []NodeHeartbeat, currentDeviceName string) []utils.ConstellationTunnel {
	isLB, err := GetCurrentDeviceIsLoadbalancer()
	if err != nil || !isLB {
		return []utils.ConstellationTunnel{}
	}
	return mergeTunnelHeartbeats(heartbeats, currentDeviceName)
}

func UpdateLocalTunnelCache() {
	if IsConstellationStandalone() {
		return
	}

	localTunnelCacheMutex.Lock()
	defer localTunnelCacheMutex.Unlock()

	currentDeviceName, err := GetCurrentDeviceName()
	if err != nil {
		utils.Warn("[constellation] Failed to get current device name for tunnel cache update: " + err.Error())
		return
	}

	err = ClientConnectToJS()
	if err != nil {
		utils.Warn("[NATS] Error connecting to JetStream during tunnel cache update: " + err.Error())
		return
	}

	// Hold read lock for KV operations
	clientConfigLock.RLock()
	if js == nil {
		clientConfigLock.RUnlock()
		utils.Warn("[NATS] JetStream context is nil during tunnel cache update")
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
		utils.Warn("[NATS] Error getting keys from Key-Value store during tunnel cache update " + err.Error())
		return
	}

	heartbeats := make([]NodeHeartbeat, 0, len(keys))

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

		heartbeats = append(heartbeats, heartbeat)
	}
	clientConfigLock.RUnlock() // Done with KV operations

	// every node keeps the cluster DNS map, even non load balancers
	setClusterDNS(buildClusterDNS(heartbeats, loadBalancerIPs()))

	tunnels := tunnelsForNode(heartbeats, currentDeviceName)

	changed := !utils.JSONEquals(sortTunnelsForComparison(localTunnelCache), sortTunnelsForComparison(tunnels))

	localTunnelCache = tunnels
	lastCacheUpdate = time.Now()

	if changed {
		utils.Log("[constellation] Tunnel cache changed, restarting HTTP server...")
		go utils.RestartHTTPServer()
	}
}

func GetLocalTunnelCache() []utils.ConstellationTunnel {
	if !utils.GetMainConfig().ConstellationConfig.Enabled {
		return []utils.ConstellationTunnel{}
	}

	if IsConstellationStandalone() {
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

func ensureStickyBucket() (nats.KeyValue, error) {
	if js == nil {
		return nil, nats.ErrConnectionClosed
	}
	kv, err := js.KeyValue("tunnel-sticky")
	if err == nil {
		return kv, nil
	}
	if utils.FBL.AgentMode {
		// managers own bucket creation; the manager pre-creates this in
		// ClientHeartbeatInit, so just report it's not there yet
		return nil, err
	}
	return createKVAtTopology(nats.KeyValueConfig{
		Bucket:  "tunnel-sticky",
		TTL:     120 * time.Second,
		Storage: nats.MemoryStorage,
	})
}

func GetStickyTarget(clientKey string) (string, bool) {
	clientConfigLock.RLock()
	defer clientConfigLock.RUnlock()

	kv, err := ensureStickyBucket()
	if err != nil {
		return "", false
	}

	entry, err := kv.Get(clientKey)
	if err != nil {
		return "", false
	}
	return string(entry.Value()), true
}

func SetStickyTarget(clientKey string, deviceName string) {
	clientConfigLock.RLock()
	defer clientConfigLock.RUnlock()

	kv, err := ensureStickyBucket()
	if err != nil {
		utils.Error("[NATS] Error accessing sticky KV bucket", err)
		return
	}

	_, err = kv.Put(clientKey, []byte(deviceName))
	if err != nil {
		utils.Error("[NATS] Error setting sticky target", err)
	}
}
