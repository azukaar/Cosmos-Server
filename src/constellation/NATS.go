package constellation

import (
	"crypto/tls"
	"encoding/pem"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/azukaar/cosmos-server/src/pro"
	"github.com/azukaar/cosmos-server/src/utils"

	natsClient "github.com/nats-io/nats.go"
)

type NodeHeartbeat struct {
	DeviceName   string
	IP           string
	IsRelay      bool
	IsLighthouse bool
	IsExitNode   bool
	CosmosNode   int
	Tunnels      []utils.ProxyRouteConfig
	// Hostnames are the names this node serves itself (main hostname and
	// non-tunneled routes) so every constellation DNS can answer for them.
	Hostnames []string `json:"hostnames,omitempty"`
	// RunningDeployments is the list of scheduler-managed deployment names
	// currently running on this node, derived from docker containers carrying
	// the `cosmos-deployment` label. Populated from docker at heartbeat time;
	// see UpdateLocalTunnelCache / heartbeat goroutine in tunnels.go.
	RunningDeployments []string `json:"runningDeployments"`
	// RunningDeploymentVersions maps each running deployment name to the spec
	// version its containers were created from (the cosmos-deployment-version
	// label). The scheduler diffs this against the desired Deployment.Version to
	// detect a node running a stale spec and trigger a rolling re-apply. Built
	// from docker alongside RunningDeployments each heartbeat.
	RunningDeploymentVersions map[string]int `json:"runningDeploymentVersions,omitempty"`
	// Liveness only, from the `cosmos-managed-db` label; never folded into
	// RunningDeployments, which the scheduler prunes of unrecognized entries.
	ManagedDBs []string `json:"managedDBs,omitempty"`
	// Same contract as ManagedDBs, from the `cosmos-seaweedfs` label (masters
	// only; the volume/filer tiers are deployments and report normally).
	SeaweedFS []string `json:"seaweedfs,omitempty"`
	// Registry endpoints this node serves. Same contract as ManagedDBs, but
	// derived from the access records and this node's tags, not from docker.
	Registries []string `json:"registries,omitempty"`
	// Load sample from pro.GetCurrentResources(); RAMPercent is memory
	// pressure (PSI + exhaustion guard), not occupancy. Zero when MonitoringOn is false.
	CPUPercent float64 `json:"cpuPercent,omitempty"`
	RAMPercent float64 `json:"ramPercent,omitempty"`
	// MonitoringOn signals whether CPU/RAM numbers are trustworthy. False when
	// the operator disabled monitoring (MonitoringDisabled config flag) or
	// when the sampler hasn't produced a reading yet.
	MonitoringOn bool `json:"monitoringOn"`
	// Tags mirror ConstellationDevice.Tags so the leader can filter eligible
	// placement targets by deployment affinity without an extra DB round-trip.
	Tags []string `json:"tags,omitempty"`
}

// atomic: written under natsStartMutex, but read lock-free by status/ping paths
var ns atomic.Pointer[server.Server]

func truncateLog(s string) string {
	if len(s) > 100 {
		return s[:100] + "..."
	}
	return s
}

// redactSecret returns a short, non-reversible hint of a secret for logs
func redactSecret(s string) string {
	if s == "" {
		return "<empty>"
	}
	if len(s) <= 4 {
		return "****(len " + strconv.Itoa(len(s)) + ")"
	}
	return s[:4] + "****(len " + strconv.Itoa(len(s)) + ")"
}

// the mobile app's only subject (see MasterNATSClientRouter)
const publicDevicesSubject = "cosmos._global_.constellation.public-devices"

// natsUserPermissions returns a device's NATS subject rules.
//
// Cosmos servers get none: their traffic already crosses a leafnode link whose
// users carry no permissions, so the allow-list only ever handicapped
// manager-local callers relative to agents — which is how the missing $JS.FC.>
// publish hid, stalling ordered consumers and KV watches under backlog.
//
// CosmosNode 0 is the mobile app: one request, one inbox. No publish on
// _INBOX.> — the server checks the published subject, never the reply-to — and
// no cosmos._global_.>, which would let a phone write the op-log stream.
func natsUserPermissions(device utils.ConstellationDevice) *server.Permissions {
	if device.CosmosNode > 0 {
		return nil
	}

	return &server.Permissions{
		Publish: &server.SubjectPermission{
			Allow: []string{publicDevicesSubject},
		},
		Subscribe: &server.SubjectPermission{
			// the app subscribes to _INBOX.<nuid>.>, not a literal inbox
			Allow: []string{"_INBOX.>"},
		},
	}
}

func GetClusterIPs() ([]*url.URL, error) {
	ipsMap := make(map[string]bool)

	// add lighthouse IPs from nebula config
	lips, _ := GetAllLighthouseIPFromTempConfig()

	for _, ip := range lips {
		ipsMap[ip] = true
	}

	// read IPs from cached devices — any device running its own NATS server
	// (CosmosNode > 0), not just the Nebula lighthouse: a Manager/Agent peer
	// that isn't the Nebula lighthouse would otherwise never be routed to.
	cachedDevices, _ := deviceCacheSnapshot()
	for _, device := range cachedDevices {
		if device.CosmosNode > 0 {
			ipsMap[device.IP] = true
		}
	}

	ips := []*url.URL{}
	for ip := range ipsMap {
		parsedIP, err := url.Parse("nats-route://" + ip + ":6222")
		if err == nil {
			ips = append(ips, parsedIP)
		}
	}

	if len(ips) == 0 {
		return ips, errors.New("No cluster IPs found")
	}

	return ips, nil
}

// natsLeafPort is the port managers listen on for agent leafnode connections.
const natsLeafPort = 7422

// jetstreamDir is THE JetStream store directory — one expression, used by the
// server option and by every path that deletes it.
//
// CONFIGFOLDER is copied verbatim out of COSMOS_CONFIG_FOLDER with no
// trailing-slash normalization (index.go), so string concatenation gives two
// different answers depending on the spelling: with COSMOS_CONFIG_FOLDER=/config,
// `CONFIGFOLDER + "/jetstream"` is /config/jetstream while `CONFIGFOLDER +
// "jetstream"` is /configjetstream. The second silently RemoveAll's a path that
// does not exist and reports success, which would leave the abandoned cluster's
// state in place through a force-reform — the one thing the reform must not do.
func jetstreamDir() string {
	return filepath.Join(utils.CONFIGFOLDER, "jetstream")
}

// readCstlnConfigField returns a cstln_* field from this node's nebula.yml;
// the creator's own nebula.yml lacks these (its values live in the main config).
func readCstlnConfigField(key string) interface{} {
	nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")
	if err != nil {
		return nil
	}
	configMap := make(map[string]interface{})
	if yaml.Unmarshal(nebulaFile, &configMap) != nil {
		return nil
	}
	return configMap[key]
}

// IsNATSHA reports whether this constellation was created with the HA
// (clustered JetStream) option. Immutable: read from the creator's main
// config (NATSReplicas >= 3) or from this node's device config (cstln_nats_ha).
func IsNATSHA() bool {
	if utils.GetMainConfig().ConstellationConfig.NATSReplicas >= 3 {
		return true
	}
	if ha, ok := readCstlnConfigField("cstln_nats_ha").(bool); ok {
		return ha
	}
	return false
}

// getSeedManagerIPs returns the manager IPs written into this node's device
// config at enrollment (cstln_nats_managers); used only as bootstrap dial hints.
func getSeedManagerIPs() []string {
	seeds := []string{}
	if list, ok := readCstlnConfigField("cstln_nats_managers").([]interface{}); ok {
		for _, v := range list {
			if s, ok := v.(string); ok && s != "" {
				seeds = append(seeds, cleanIp(s))
			}
		}
	}
	return seeds
}

// blockedDeviceIPs returns the nebula IPs that belong ONLY to blocked devices —
// read from the store rather than the cache, since refreshDeviceCache drops
// blocked devices and so can say who is live but never who was removed.
//
// Subtracting the live set is not a refinement, it is the whole correctness of
// this function. An IP is not retired with its device: GetNextAvailableIP
// allocates out of the unblocked devices only, so the replacement manager
// enrolled after a force-reform normally takes the dead manager's address back.
// Keyed on IP alone, this would then remove the LIVE replacement from the
// cluster route list and JetStream would never elect — turning the reform into
// the outage it was meant to repair. The schema allows the overlap on purpose
// (UNIQUE(ip) WHERE blocked=0).
func blockedDeviceIPs() map[string]bool {
	blocked := map[string]bool{}
	devices, err := utils.FindDevices(map[string]interface{}{"Blocked": true})
	if err != nil {
		utils.Debug("[NATS] Could not read blocked devices: " + err.Error())
		return blocked
	}
	for _, d := range devices {
		if d.IP != "" {
			blocked[cleanIp(d.IP)] = true
		}
	}
	if len(blocked) == 0 {
		return blocked
	}

	live, err := utils.ListDevices(false)
	if err != nil {
		// Fail open: subtracting nothing leaves a stale route, which the epoch fence
		// still makes harmless. Subtracting wrongly drops a live manager out of the
		// cluster, which nothing else protects against.
		utils.Debug("[NATS] Could not read live devices, keeping every route: " + err.Error())
		return map[string]bool{}
	}
	for _, d := range live {
		delete(blocked, cleanIp(d.IP))
	}
	return blocked
}

// isAgentNode derives the role from nebula.yml (cstln_cosmos_node) via
// GetCurrentDevice, the source of truth that survives enrollment without a
// ProcessLicence refresh; FBL.AgentMode is only the fallback when it's unreadable.
func isAgentNode() bool {
	if d, err := GetCurrentDevice(); err == nil && d.CosmosNode > 0 {
		return d.CosmosNode == 1
	}
	return utils.FBL.AgentMode
}

// getManagerIPs returns the nebula IPs of every known Manager node
// (CosmosNode == 2), excluding excludeIP: device cache union enrollment seeds, minus
// anything blocked, falling back to the nebula lighthouses when both are empty.
func getManagerIPs(excludeIP string) []string {
	ipsMap := map[string]bool{}
	cachedDevices, _ := deviceCacheSnapshot()
	for _, d := range cachedDevices {
		if d.CosmosNode == 2 && d.IP != "" {
			ipsMap[cleanIp(d.IP)] = true
		}
	}
	for _, ip := range getSeedManagerIPs() {
		ipsMap[ip] = true
	}
	if len(ipsMap) == 0 {
		lips, _ := GetAllLighthouseIPFromTempConfig()
		for _, ip := range lips {
			ipsMap[cleanIp(ip)] = true
		}
	}
	delete(ipsMap, cleanIp(excludeIP))

	// The seeds above are a bootstrap hint frozen into nebula.yml at enrollment
	// and never rewritten, so on their own they can only ever ADD a manager and
	// never retire one. Subtracting the blocked set is what lets this list shrink
	// — which force-reform depends on, since these IPs become the cluster route
	// list and "routes must never point at non-voters" (see StartNATS): a reform
	// that wiped the store dir and then rebuilt the meta group around the two
	// managers it just fenced would have isolated nothing.
	for ip := range blockedDeviceIPs() {
		delete(ipsMap, ip)
	}

	ips := make([]string, 0, len(ipsMap))
	for ip := range ipsMap {
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	return ips
}

func GetNATSCredentials() (string, string, error) {
	currentDevice, _ := GetCurrentDevice()

	utils.Debug("GetNATSCredentials: currentDevice.APIKey=" + redactSecret(currentDevice.APIKey) + " currentDevice.DeviceName=" + currentDevice.DeviceName)

	if currentDevice.APIKey != "" && currentDevice.DeviceName != "" {
		return currentDevice.DeviceName, currentDevice.APIKey, nil
	} else {
		nebulaFile, err := ioutil.ReadFile(utils.CONFIGFOLDER + "nebula.yml")
		if err != nil {
			utils.Error("GetNATSCredentials: error while reading nebula.yml", err)
			return "", "", err
		}

		configMap := make(map[string]interface{})
		err = yaml.Unmarshal(nebulaFile, &configMap)
		if err != nil {
			utils.Error("GetNATSCredentials: Invalid slave config file for resync", err)
			return "", "", err
		}

		if configMap["cstln_api_key"] == nil || configMap["cstln_device_name"] == nil {
			utils.Error("GetNATSCredentials: Invalid slave config file for resync", nil)
			return "", "", errors.New("Invalid slave config file for resync")
		}

		apiKey := configMap["cstln_api_key"].(string)
		deviceName := configMap["cstln_device_name"].(string)

		utils.Debug("GetNATSCredentials: found credentials in nebula.yml: deviceName=" + deviceName + " apiKey=" + redactSecret(apiKey))

		return deviceName, apiKey, nil
	}
}

// natsLogAdapter forwards the embedded nats-server's own internal logging
// (route/cluster listener bind errors, JetStream state, etc.) into cosmos's
// logger. nats-server otherwise logs these to its own discarded/unconfigured
// logger, so failures like a route listener bind error never reach cosmos.log.
type natsLogAdapter struct{}

func (natsLogAdapter) Noticef(format string, v ...interface{}) {
	utils.Debug("[NATS internal] " + fmt.Sprintf(format, v...))
}
func (natsLogAdapter) Warnf(format string, v ...interface{}) {
	utils.Warn("[NATS internal] " + fmt.Sprintf(format, v...))
}
func (natsLogAdapter) Fatalf(format string, v ...interface{}) {
	utils.Error("[NATS internal] "+fmt.Sprintf(format, v...), nil)
}
func (natsLogAdapter) Errorf(format string, v ...interface{}) {
	utils.Error("[NATS internal] "+fmt.Sprintf(format, v...), nil)
}
func (natsLogAdapter) Debugf(format string, v ...interface{}) {
	utils.Debug("[NATS internal] " + fmt.Sprintf(format, v...))
}
func (natsLogAdapter) Tracef(format string, v ...interface{}) {
	utils.Debug("[NATS internal] " + fmt.Sprintf(format, v...))
}

// natsStartMutex serializes all `ns` writers (StartNATS from Init/restart/watchdog, StopNATS);
// waiters re-check ns/NATSStarted after acquiring rather than no-oping.
var natsStartMutex sync.Mutex

func StartNATS() {
	natsStartMutex.Lock()
	defer natsStartMutex.Unlock()

	if ns.Load() != nil {
		return
	}

	setNATSStartTime(time.Now())

	ip, err := GetCurrentDeviceIP()
	if err != nil {
		utils.Error("[NATS] Failed to get current device IP", err)
		return
	}

	utils.Log("[NATS] Starting NATS server on " + ip + ":4222")

	time.Sleep(2 * time.Second)

	config := utils.GetMainConfig()
	HTTPConfig := config.HTTPConfig

	var tlsCert = HTTPConfig.TLSCert
	var tlsKey = HTTPConfig.TLSKey

	// Ensure the PEM data is correctly formatted
	certPEMBlock := []byte(tlsCert)
	keyPEMBlock := []byte(tlsKey)

	// Decode PEM encoded certificate
	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil {
		utils.MajorError("[NATS] Failed to start NATS: parse certificate PEM", nil)
		return
	}

	// Decode PEM encoded private key
	keyDERBlock, _ := pem.Decode(keyPEMBlock)
	if keyDERBlock == nil {
		utils.MajorError("[NATS] Failed to start NATS: parse key PEM", nil)
		return
	}

	// Create tls.Certificate using the original PEM data
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		utils.MajorError("[NATS] Failed to start NATS: create TLS certificate", err)
		return
	}

	// Configure the NATS server options
	// Make users

	users := []*server.User{}
	// The cache indexes a device under its name AND every public hostname;
	// leafnode auth rejects duplicate users, which kills the server start.
	seenUsers := map[string]bool{}

	cachedDevices, _ := deviceCacheSnapshot()
	for _, devices := range cachedDevices {
		username := devices.DeviceName
		if seenUsers[username] {
			continue
		}
		seenUsers[username] = true
		utils.Debug("[NATS] Adding NATS user for device: " + devices.DeviceName + " With API Key: " + redactSecret(devices.APIKey))

		users = append(users, &server.User{
			Username:    username,
			Password:    devices.APIKey,
			Permissions: natsUserPermissions(devices),
		})
	}

	device, err := GetCurrentDevice()
	if err != nil {
		utils.Error("[NATS] Failed to get current device IP", err)
		return
	}

	natsHost := device.IP
	natsName := device.DeviceName

	// if debug, add debug user
	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		users = append(users, &server.User{
			Username:    "DEBUG",
			Password:    "DEBUG",
			Permissions: nil,
		})

		natsHost = "0.0.0.0"
	}

	// Topology is declared, never inferred: IsNATSHA says whether a JS
	// cluster exists; managers (CosmosNode == 2) are the only cluster
	// members, agents attach as leafnodes. nats-server sizes and persists
	// the JS meta-Raft group from the initial route list, so routes must
	// never point at non-voters.
	isAgent := isAgentNode()
	haMode := IsNATSHA()

	opts := &server.Options{
		Host: natsHost,
		Port: 4222,

		ServerName: natsName,

		JetStream: !isAgent,
		StoreDir:  jetstreamDir(),

		// Lets agents reach the manager's JetStream over leafnodes: nats-server
		// otherwise denies $JS.API.> across account leaf links. Enforced
		// per-server, so BOTH manager and agents need it.
		JsAccDefaultDomain: map[string]string{"$G": ""},

		TLSConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			ClientAuth:         tls.NoClientCert,
			InsecureSkipVerify: true,
		},

		Users: users,
	}

	if isAgent {
		// Agent: solicit leafnode connections to every known manager; only
		// affects failover redundancy — one reachable manager is enough.
		user, pwd, errCreds := GetNATSCredentials()
		if errCreds != nil {
			utils.Error("[NATS] Failed to get credentials for leafnode connection", errCreds)
			return
		}

		managerIPs := getManagerIPs(device.IP)
		if len(managerIPs) == 0 {
			utils.Warn("[NATS] Agent knows no managers yet, NATS will stay isolated until the device DB syncs")
		}

		leafURLs := []*url.URL{}
		for _, mip := range managerIPs {
			leafURLs = append(leafURLs, &url.URL{
				Scheme: "nats-leaf",
				User:   url.UserPassword(user, pwd),
				Host:   mip + ":" + strconv.Itoa(natsLeafPort),
			})
		}

		opts.LeafNode = server.LeafNodeOpts{
			Remotes: []*server.RemoteLeafOpts{{
				URLs: leafURLs,
				TLS:  true,
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}},
		}
	} else {
		// Manager: accept agents as leafnodes, same credentials as clients.
		leafUsers := make([]*server.User, 0, len(users))
		for _, u := range users {
			leafUsers = append(leafUsers, &server.User{Username: u.Username, Password: u.Password})
		}

		opts.LeafNode = server.LeafNodeOpts{
			Host:        natsHost,
			Port:        natsLeafPort,
			NoAdvertise: true,
			TLSConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				ClientAuth:         tls.NoClientCert,
				InsecureSkipVerify: true,
			},
			Users: leafUsers,
		}

		if haMode {
			// HA: clustered JetStream among managers ONLY. Routes are just
			// dial hints — gossip meshes the rest.
			opts.Cluster = server.ClusterOpts{
				Name: "Constellation",
				Host: device.IP,
				Port: 6222,
				TLSConfig: &tls.Config{
					Certificates:       []tls.Certificate{cert},
					ClientAuth:         tls.NoClientCert,
					InsecureSkipVerify: true,
				},
			}

			routes := []*url.URL{}
			for _, mip := range getManagerIPs(device.IP) {
				if r, errR := url.Parse("nats-route://" + mip + ":6222"); errR == nil {
					routes = append(routes, r)
				}
			}
			if len(routes) == 0 {
				// nats-server refuses clustered JS with zero configured routes;
				// a self-route is never dialed, it only passes that check. JS
				// stays down until a second manager enrolls — formation, not failure.
				selfRoute, errR := url.Parse("nats-route://" + device.IP + ":6222")
				if errR != nil {
					utils.Error("[NATS] Failed to build self route", errR)
					return
				}
				routes = []*url.URL{selfRoute}
				utils.Log("[NATS] HA cluster forming: JetStream waits for a second manager to enroll before it can elect a leader")
			}
			opts.Routes = routes
		} else {
			// Clear clustered meta state ($SYS raft groups) left by older
			// versions so it cannot confuse a non-clustered start.
			os.RemoveAll(filepath.Join(jetstreamDir(), "$SYS"))
		}
	}

	// Create and start the embedded NATS server
	retries := 0
	err = errors.New("")

	for err != nil && retries < 5 {
		// drop any instance left over from a failed attempt
		if old := ns.Load(); old != nil {
			old.Shutdown()
			old.WaitForShutdown()
			ns.Store(nil)
		}

		var srv *server.Server
		srv, err = server.NewServer(opts)
		if err != nil {
			retries++
			continue
		}
		ns.Store(srv)

		if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
			srv.SetLogger(natsLogAdapter{}, true, true)
		}

		if !NebulaStarted.Load() {
			utils.Error("[NATS] Nebula not started, aborting NATS server setup", nil)
			srv.Shutdown()
			ns.Store(nil)
			return
		}

		go srv.Start()

		// Wait for the server to be ready
		if !srv.ReadyForConnections(time.Duration(2*(retries+1)) * time.Second) {
			retries++
			utils.Debug("[NATS] NATS server not ready...")
			err = errors.New("NATS server not ready")
			continue
		}

		utils.Debug("[NATS] Retrying to start NATS server")
	}

	if err != nil {
		if old := ns.Load(); old != nil {
			old.Shutdown()
			old.WaitForShutdown()
			ns.Store(nil)
		}
		utils.MajorError("[NATS] Error starting NATS server", err)
		// hand off to the watchdog so a failed bring-up gets retried (standalone nodes don't need a server)
		if NebulaStarted.Load() && !IsConstellationStandalone() {
			go natsServerWatchdog()
		}
		return
	}

	NATSStarted.Store(true)

	utils.Log("[NATS] Started NATS server on host " + opts.Host + ":" + strconv.Itoa(opts.Port))

	// Retry client init with capped backoff for as long as our server is up —
	// InitNATSClient's own retries are bounded, so a slow nebula bring-up
	// would otherwise strand the process with a server but no client.
	go func() {
		if !atomic.CompareAndSwapInt32(&natsInitSupervisorRunning, 0, 1) {
			return
		}
		defer atomic.StoreInt32(&natsInitSupervisorRunning, 0)

		backoff := 2 * time.Second
		for NATSStarted.Load() {
			if err := InitNATSClient(); err == nil {
				return
			}
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
		}
	}()
}

// natsInitSupervisorRunning keeps repeated StartNATS calls (constellation
// restarts) from stacking retry goroutines.
var natsInitSupervisorRunning int32

// natsServerWatchdogRunning single-flights the watchdog goroutine (CAS on entry, store-0 on exit)
var natsServerWatchdogRunning int32

// natsServerWatchdog retries StartNATS with capped backoff while nebula is up but the server isn't
func natsServerWatchdog() {
	if !atomic.CompareAndSwapInt32(&natsServerWatchdogRunning, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&natsServerWatchdogRunning, 0)

	backoff := 5 * time.Second
	for {
		// re-check on both sides of the sleep so we never fight an in-progress restart
		if !NebulaStarted.Load() || NATSStarted.Load() {
			return
		}
		time.Sleep(backoff)
		if !NebulaStarted.Load() || NATSStarted.Load() {
			return
		}

		utils.Warn("[NATS] Watchdog retrying NATS server start (backoff " + backoff.String() + ")")
		StartNATS()

		if backoff < 60*time.Second {
			backoff *= 2
			if backoff > 60*time.Second {
				backoff = 60 * time.Second
			}
		}
	}
}

// natsStartTimeNano marks the last constellation NATS bring-up (UnixNano, 0 =
// never); within natsStartGracePeriod the UI reports failing steps as still
// starting.
//
// Atomic rather than a time.Time: StartNATS writes it while GetNATSStatus reads
// it from HTTP handlers, and a time.Time is multi-word (wall, ext, loc), so that
// pairing is an unconditional data race, not a narrow interleaving. Polling
// status during a nebula bounce makes the overlap structural.
var natsStartTimeNano atomic.Int64

func setNATSStartTime(t time.Time) {
	natsStartTimeNano.Store(t.UnixNano())
}

// natsStarting reports whether we are inside the post-start grace period.
func natsStarting() bool {
	nano := natsStartTimeNano.Load()
	if nano == 0 {
		return false
	}
	return time.Since(time.Unix(0, nano)) < natsStartGracePeriod
}

const natsStartGracePeriod = 2 * time.Minute

func StopNATS() {
	utils.Log("[NATS] Stopping NATS server...")

	StopOplogApply()

	// serialized with StartNATS so we never shut down a server mid-bring-up
	natsStartMutex.Lock()
	defer natsStartMutex.Unlock()

	if old := ns.Load(); old != nil {
		old.Shutdown()
		old.WaitForShutdown()
		ns.Store(nil)
	}
	NATSStarted.Store(false)
}

// sync lock - RWMutex allows multiple readers, single writer
var clientConfigLock = sync.RWMutex{}
var NATSClientTopic = ""
var nc *nats.Conn
var js nats.JetStreamContext

func connectNATSClient(url string, user string, pwd string) (*nats.Conn, error) {
	return natsClient.Connect(url,
		nats.Secure(&tls.Config{
			InsecureSkipVerify: true,
		}),

		nats.UserInfo(user, pwd),

		// timeout
		nats.Timeout(2*time.Second),

		// the default (60) gives up permanently after ~2min of server
		// downtime, stranding the process until a full restart
		nats.MaxReconnects(-1),

		nats.NoEcho(),

		nats.ErrorHandler(natsAsyncError),
	)
}

// natsAsyncError surfaces the errors nats.go reports out of band instead of
// returning them from a call.
//
// Nothing registered a handler before this, so every one of those was discarded
// — which is why a survivor whose ordered consumer died during a quorum outage
// produced no log line at all, only silence and 503s. The library knew
// (ErrConsumerNotActive, and the consumer-recreation failures behind it); we
// simply were not listening.
//
// This is diagnostics plus a hint, NOT the recovery mechanism. Recovery stays
// with the progress probe (oplog_apply.go), which owes nothing to nats.go's
// internals and so cannot be broken by a client upgrade.
func natsAsyncError(_ *nats.Conn, sub *nats.Subscription, err error) {
	if err == nil {
		return
	}

	where := ""
	if sub != nil {
		where = " on " + sub.Subject
	}
	utils.Warn("[NATS] async error" + where + ": " + err.Error())

	// The library spots an inactive consumer well before three probes can prove
	// it. Arming the detector rather than detaching here keeps the evidence
	// requirement intact: the next probe still has to see the log ahead of us AND
	// our position frozen, so a spurious signal costs nothing.
	if sub != nil && oplogOwnsSubscription(sub) {
		oplogArmStallDetector()
	}
}

func InitNATSClient() error {
	if !NATSStarted.Load() {
		utils.Warn("[NATS] NATS server not started, cannot initialize client")
		return errors.New("NATS server not started, cannot initialize client")
	}

	clientConfigLock.Lock()
	defer clientConfigLock.Unlock()

	if nc != nil {
		// already connected — success for callers using this as "ensure client"
		return nil
	}

	var err error
	retries := 0

	if !NebulaStarted.Load() {
		utils.Error("[NATS] Nebula not started, aborting NATS client connection", nil)
		return errors.New("Nebula not started, aborting NATS client connection")
	}

	utils.Log("[NATS] Connecting to NATS server...")

	time.Sleep(2 * time.Second)

	user, pwd, err := GetNATSCredentials()

	if err != nil {
		utils.MajorError("[NATS] Error getting constellation credentials", err)
		return err
	}

	deviceIp, err := GetCurrentDeviceIP()
	if err != nil {
		utils.MajorError("[NATS] Error getting current device IP", err)
		return err
	}

	natsUrl := "nats://" + deviceIp + ":4222"

	nc, err = connectNATSClient(natsUrl, user, pwd)

	for err != nil {
		if retries >= 10 {
			utils.MajorError("[NATS] Error connecting to Constellation NATS server after 10 tries", err)
			nc = nil
			return err
		}

		if !NebulaStarted.Load() {
			utils.Error("[NATS] Nebula not started, aborting NATS client connection retry", nil)
			nc = nil
			return errors.New("Nebula not started, aborting NATS client connection retry")
		}

		clientConfigLock.Unlock()
		time.Sleep(time.Duration(2*(retries+1)) * time.Second)
		clientConfigLock.Lock()

		if !NebulaStarted.Load() {
			retries++
			utils.Warn("[NATS] Nebula not started yet, delaying NATS client connection retry")
			continue
		}

		nc, err = connectNATSClient(natsUrl, user, pwd)

		if err != nil {
			retries++
			utils.Debug("[NATS] Retrying to start NATS Client: " + err.Error())
		}
	}

	if err != nil {
		utils.MajorError("[NATS] Error connecting to Constellation NATS server", err)
		nc = nil
		return err
	} else {
		utils.Log("[NATS] Connected to NATS server as " + user)
		NATSClientTopic = "cosmos." + user
	}

	utils.Debug("[NATS] NATS client connected")

	MasterNATSClientRouter()

	// Initialize JetStream directly (holding write lock)
	js, err = nc.JetStream(nats.MaxWait(10 * time.Second))
	if err != nil {
		utils.Error("[NATS] Failed to get JetStream context", err)
	}

	// standalone has no peers to heartbeat with and writes straight to SQLite
	if IsConstellationStandalone() {
		utils.Debug("[NATS] Standalone constellation, skipping heartbeat and op-log")
	} else {
		go ClientHeartbeatInit()

		// the apply loop materializes this node's config from the op-log; it
		// waits for JetStream itself, so starting it here is safe even if JS
		// isn't elected yet
		StartOplogApply()
	}

	// POST CLIENT CONNECTION HOOK

	return nil
}

var lastCheck time.Time

func ClientConnectToJS() error {
	clientConfigLock.Lock()
	defer clientConfigLock.Unlock()

	if nc == nil {
		return errors.New("NATS client not connected")
	}

	if js != nil && time.Since(lastCheck) < 5*time.Second {
		return nil
	}

	if js != nil {
		if _, err := js.AccountInfo(); err == nil {
			lastCheck = time.Now()
			return nil
		}
	}

	var err error
	js, err = nc.JetStream(nats.MaxWait(6 * time.Second))
	if err != nil {
		return fmt.Errorf("error getting JetStream context: %w", err)
	}

	// nc.JetStream only builds a local context — a real round-trip is needed
	// so a nil error means "JS reachable".
	if _, err := js.AccountInfo(); err != nil {
		return fmt.Errorf("JetStream API not reachable: %w", err)
	}

	lastCheck = time.Now()
	return nil
}

func IsClientConnected() bool {
	clientConfigLock.RLock()
	defer clientConfigLock.RUnlock()

	if nc == nil {
		return false
	}

	return nc.IsConnected()
}

// IsConstellationStandalone reports whether this server has no Cosmos peers
// to talk to over NATS — neither peer lighthouses to cluster with nor non-
// lighthouse Cosmos servers that would connect to this node as clients. Plain
// Nebula client devices (CosmosNode == 0) don't run NATS so they don't count.
// When true, all NATS-adjacent activity must be skipped.
func IsConstellationStandalone() bool {
	if !utils.GetMainConfig().ConstellationConfig.Enabled {
		return true
	}
	myIP, err := GetCurrentDeviceIP()
	if err != nil {
		utils.Error("[NATS] Failed to get current device IP", err)
		return true
	}
	cachedDevices, _ := deviceCacheSnapshot()
	for _, device := range cachedDevices {
		if device.IP == myIP {
			continue
		}
		if device.CosmosNode > 0 {
			return false
		}
	}
	// Before the initial sync, CachedDevices only knows about this server.
	// The enrollment seed managers exist for exactly this window: an agent that
	// is itself the lighthouse would otherwise read as standalone and never
	// start the heartbeat/op-log that populates the cache.
	for _, ip := range getSeedManagerIPs() {
		if ip != cleanIp(myIP) {
			return false
		}
	}
	// GetClusterIPs also pulls peer lighthouses from the nebula config file,
	// so fall back to it to detect bootstrap-time peers we can cluster with.
	cips, _ := GetClusterIPs()
	for _, u := range cips {
		if u.Hostname() != myIP {
			return false
		}
	}
	return true
}

// NATSStatus is a debugging snapshot of this node's declared topology and
// the live state of the embedded NATS stack.
type NATSStatus struct {
	// declared topology
	Role   string `json:"role"`   // "manager" | "agent"
	HAMode bool   `json:"haMode"` // clustered JetStream constellation

	// live state
	NebulaStarted   bool `json:"nebulaStarted"`
	ServerRunning   bool `json:"serverRunning"`
	ClientConnected bool `json:"clientConnected"`

	// Starting: within the post-start grace period — failing steps are
	// likely still coming up rather than broken
	Starting bool `json:"starting"`

	// "disabled" (agents, proxied to managers) | "standalone" | "clustered"
	JetStreamMode  string `json:"jetstreamMode"`
	JetStreamReady bool   `json:"jetstreamReady"` // JS API answers
	KVNodesReady   bool   `json:"kvNodesReady"`   // constellation-nodes bucket reachable

	KnownManagers []string `json:"knownManagers"`

	// connection counts from the embedded server
	ConnectedClients int `json:"connectedClients"`
	ConnectedLeafs   int `json:"connectedLeafs"` // manager: attached agents; agent: 1 when attached to a manager
	ClusterRoutes    int `json:"clusterRoutes"`  // HA managers: connected manager routes

	// ManagerLinkUp: can this node reach a manager's NATS — matters on
	// agents, whose local client is always connected to its own server.
	ManagerLinkUp bool `json:"managerLinkUp"`

	// op-log health. OplogAttached is also "config is writable here": writes
	// are published, not applied locally, so a detached node is read-only.
	OplogAttached   bool   `json:"oplogAttached"`
	OplogHalted     bool   `json:"oplogHalted"`
	OplogHaltReason string `json:"oplogHaltReason,omitempty"`
	OplogEpoch      uint64 `json:"oplogEpoch"`
	OplogSeq        uint64 `json:"oplogSeq"`
	ConfigWritable  bool   `json:"configWritable"`

	// OplogReplicas is the log's actual replication factor, 0 when the stream is
	// not reachable. It is the only externally visible evidence that the R1→R3
	// scale-up happened: writes succeed perfectly well against an R1 stream, so a
	// cluster that silently never scaled looks identical to a healthy one.
	OplogReplicas int `json:"oplogReplicas"`
}

func GetNATSStatus() NATSStatus {
	isAgent := isAgentNode()

	srv := ns.Load()
	status := NATSStatus{
		Role:          "manager",
		HAMode:        IsNATSHA(),
		NebulaStarted: NebulaStarted.Load(),
		ServerRunning: NATSStarted.Load() && srv != nil,
		Starting:      natsStarting(),
	}
	if isAgent {
		status.Role = "agent"
		status.JetStreamMode = "disabled"
	} else if status.HAMode {
		status.JetStreamMode = "clustered"
	} else {
		status.JetStreamMode = "standalone"
	}

	myIP := ""
	if device, err := GetCurrentDevice(); err == nil {
		myIP = device.IP
	}
	status.KnownManagers = getManagerIPs(myIP)

	if srv != nil {
		status.ConnectedClients = srv.NumClients()
		status.ConnectedLeafs = srv.NumLeafNodes()
		status.ClusterRoutes = srv.NumRoutes()
	}

	if isAgent {
		// on an agent NumLeafNodes counts its solicited links to managers
		status.ManagerLinkUp = status.ConnectedLeafs > 0
	} else {
		status.ManagerLinkUp = status.ServerRunning
	}

	status.ClientConnected = IsClientConnected()

	status.OplogAttached = oplogAttached.Load()
	status.OplogHalted = oplogHalted.Load()
	if reason, ok := oplogHaltReason.Load().(string); ok && status.OplogHalted {
		status.OplogHaltReason = reason
	}
	status.OplogEpoch = utils.GetOplogEpoch()
	status.OplogSeq = utils.GetLastAppliedSeq()
	status.ConfigWritable = oplogWriteMode() != oplogReadOnly

	clientConfigLock.RLock()
	if js != nil {
		if _, err := js.AccountInfo(); err == nil {
			status.JetStreamReady = true
		}
		if _, err := js.KeyValue("constellation-nodes"); err == nil {
			status.KVNodesReady = true
		}
		if si, err := js.StreamInfo(oplogStreamName(status.OplogEpoch)); err == nil {
			status.OplogReplicas = si.Config.Replicas
		}
	}
	clientConfigLock.RUnlock()

	return status
}

func CloseNATSClient() {
	utils.Log("Closing NATS client...")

	// before taking clientConfigLock: the loop's own goroutine reads it
	StopOplogApply()

	StopHeartbeat()

	utils.Debug("[NATS] Closing NATS client connection")

	clientConfigLock.Lock()
	defer clientConfigLock.Unlock()

	utils.Debug("[NATS] NATS client connection closed")

	if nc != nil {
		nc.Close()
		nc = nil
	}

	js = nil
}

func SendNATSMessage(topic string, payload string) (string, error) {
	if IsConstellationStandalone() {
		utils.Debug("[MQ] Skipping send on standalone constellation: " + topic)
		return "", nil
	}

	if !IsClientConnected() {
		utils.Warn("NATS client not connected")
		err := InitNATSClient()
		if err != nil {
			return "", err
		}
	}

	utils.Debug("[MQ] Sending message to topic: " + topic)

	// Send a request and wait for a response
	msg, err := nc.Request(topic, []byte(payload), 2*time.Second)
	if err != nil {
		utils.Error("[MQ] Error sending request", err)
		return "", err
	}

	utils.Debug("[MQ] Received response: " + truncateLog(string(msg.Data)))

	return string(msg.Data), nil
}

func SendNATSMessageAllReply(topic string, payload string, timeout time.Duration, callback func(response string)) error {
	if IsConstellationStandalone() {
		utils.Debug("[MQ] Skipping send-all on standalone constellation: " + topic)
		return nil
	}

	if !IsClientConnected() {
		utils.Warn("NATS client not connected")
		err := InitNATSClient()
		if err != nil {
			return err
		}
	}

	utils.Debug("[MQ] Sending message to topic: " + topic)

	inbox := nats.NewInbox()
	sub, err := nc.SubscribeSync(inbox)
	if err != nil {
		utils.Error("[MQ] Error creating subscription", err)
		return err
	}
	defer sub.Unsubscribe()

	err = nc.PublishRequest(topic, inbox, []byte(payload))
	if err != nil {
		utils.Error("[MQ] Error publishing request", err)
		return err
	}

	deadline := time.Now().Add(timeout)
	for {
		msg, err := sub.NextMsg(time.Until(deadline))
		if err != nil {
			break // timeout or connection closed
		}
		utils.Debug("[MQ] Received response: " + truncateLog(string(msg.Data)))
		callback(string(msg.Data))
	}

	return nil
}

func PublishNATSMessage(topic string, payload string) error {
	if IsConstellationStandalone() {
		utils.Debug("[MQ] Skipping publish on standalone constellation: " + topic)
		return nil
	}

	if !IsClientConnected() {
		utils.Warn("NATS client not connected")
		err := InitNATSClient()
		if err != nil {
			return err
		}
	}

	utils.Debug("[MQ] Publishing message to topic: " + topic)

	// Send a request and wait for a response
	err := nc.Publish(topic, []byte(payload))
	if err != nil {
		utils.Error("[NATS] Error sending request", err)
		return err
	}

	return nil
}

func MasterNATSClientRouter() {
	utils.Log("[NATS] Starting NATS Master client router.")

	nc.Subscribe("cosmos._global_.ping", func(m *nats.Msg) {
		utils.Debug("[MQ] Received: " + truncateLog(string(m.Data)) + " from " + m.Subject)
		m.Respond([]byte("Pong"))
	})

	// public device list survives the legacy sync retirement — it is a read-only
	// query, not a replication path
	nc.Subscribe("cosmos._global_.constellation.public-devices", PublicDeviceListNATS)

	oplogSnapshotRouter(nc)

	// Scheduler: subscribe this node to its own per-target deployment command
	// subject so the leader can dispatch apply/remove here.
	if device, err := GetCurrentDevice(); err == nil {
		self := device.DeviceName
		if subErr := pro.RegisterNodeDispatchHandler(nc, self); subErr != nil {
			utils.Warn("[SCHED-NODE] failed to register dispatch handler: " + subErr.Error())
		}

		if _, subErr := pro.RegisterManagedDBResponder(nc, self); subErr != nil {
			utils.Warn("[MDB] failed to register op responder: " + subErr.Error())
		}
		pro.StartManagedDBBackupReconciler()

		if _, subErr := pro.RegisterSeaweedFSResponder(nc, self); subErr != nil {
			utils.Warn("[SWFS] failed to register op responder: " + subErr.Error())
		}
		pro.SetSeaweedFSClusterHandles(SeaweedFSClusterHandles)
		pro.StartSeaweedFSReconciler()

		pro.SetRegistryClusterHandles(RegistryClusterHandles)
		pro.SetRegistryNodeProvider(RegistryNodeInfo)
		pro.StartRegistryServing()
		// GC and cron triggers are leadership-bound so each runs once per cluster.
		pro.StartRegistryGC()
		pro.StartFunctionTriggers()

		pro.StartFeatureMetrics()
	} else {
		utils.Warn("[SCHED-NODE] cannot register dispatch handler: GetCurrentDevice failed: " + err.Error())
	}
}

func PingNATSClient() bool {
	if IsConstellationStandalone() {
		return true
	}

	// An isolated agent's local client still looks healthy — without a
	// leafnode link to a manager it isn't connected to the constellation.
	if srv := ns.Load(); isAgentNode() && (srv == nil || srv.NumLeafNodes() == 0) {
		utils.Warn("[NATS] Agent has no leafnode link to a manager")
		return false
	}

	response, err := SendNATSMessage("cosmos._global_.ping", "Ping")
	if err != nil {
		utils.Error("[NATS] Error pinging NATS client", err)
		return false
	}

	if response != "" {
		utils.Debug("[NATS] NATS client response: " + response)
		return true
	}

	return false
}
