package constellation

import (
	"crypto/tls"
	"encoding/pem"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"sync"
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
	// CPUPercent and RAMPercent are the node's latest resource-usage sample,
	// populated from pro.GetCurrentResources() on each heartbeat tick. Used by
	// the LeastBusyPlacement strategy. Zero when MonitoringOn is false.
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

var ns *server.Server

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

func sanitizeNATSUsername(username string) string {
	username = strings.ReplaceAll(username, " ", "_")
	username = strings.ReplaceAll(username, ".", "_")
	username = strings.ReplaceAll(username, "-", "_")
	username = strings.ReplaceAll(username, ":", "_")
	username = strings.ReplaceAll(username, "/", "_")
	username = strings.ReplaceAll(username, "\\", "_")
	return username
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
	for _, device := range CachedDevices {
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

func StartNATS() {
	if ns != nil {
		return
	}

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

	for _, devices := range CachedDevices {
		utils.Debug("[NATS] Adding NATS user for device: " + devices.DeviceName + " With API Key: " + redactSecret(devices.APIKey))
		username := sanitizeNATSUsername(devices.DeviceName)

		// TODO: Agent / users with less permissions

		users = append(users, &server.User{
			Username: username,
			Password: devices.APIKey,
			Permissions: &server.Permissions{
				Publish: &server.SubjectPermission{
					Allow: []string{
						"cosmos." + username + ".>", "_INBOX.>",
						"cosmos._global_.>",
						"_INBOX.>",
						// Scheduler: leader publishes per-target deployment commands
						// to cosmos.<target>.deployments.command. Scoped so non-leaders
						// can't fabricate arbitrary cross-node traffic.
						"cosmos.*.deployments.>",
						"$KV.constellation-nodes.>",
						"$KV.constellation-deployments.>",
						"$JS.API.STREAM.INFO.>",
						"$JS.API.>",
					},
				},
				Subscribe: &server.SubjectPermission{
					Allow: []string{
						"cosmos." + username + ".>", "_INBOX.>",
						"cosmos._global_.>",
						"_INBOX.>",
						"$KV.constellation-nodes.>",
						"$KV.constellation-deployments.>",
						"$JS.API.STREAM.INFO.>",
						"$JS.API.>",
					},
				},
			},
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

	standalone := IsConstellationStandalone()

	cips, err := GetClusterIPs()
	if err != nil && !standalone {
		utils.Error("[NATS] Failed to get cluster IPs", err)
	}

	utils.Debug("[NATS] Cluster IPs: ")
	for _, cip := range cips {
		utils.Debug("[NATS] Cluster IP: " + cip.String())
	}

	// Agents never run their own JetStream: a JetStream Raft group needs a
	// majority to make progress, and with exactly 2 nodes that majority is
	// both of them — worse availability than a single node, since either one
	// going down stalls it. Agents instead just cluster (core NATS routing)
	// and reach the Manager's JetStream transparently over that route (JS
	// API subjects are exported via the system account and proxy across
	// cluster connections), so cluster_size stays 1 regardless of how many
	// Agents join.
	isAgent := utils.FBL.AgentMode

	opts := &server.Options{
		Host: natsHost,
		Port: 4222,

		ServerName: natsName,

		JetStream: !isAgent,
		StoreDir:  utils.CONFIGFOLDER + "/jetstream",

		TLSConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			ClientAuth:         tls.NoClientCert,
			InsecureSkipVerify: true,
		},

		Users: users,
	}

	// Always open the cluster listener so a peer that already knows about us
	// (e.g. via its own nebula lighthouse.hosts) can connect in, even before
	// we know of anyone ourselves.
	//
	// nats-server refuses to start JetStream at all if Cluster is configured
	// with zero configured Routes and no leafnode ("JetStream cluster
	// requires configured routes or solicited leafnode for the system
	// account") — that check is static (just len(opts.Routes) > 0), it does
	// not require a route to actually be connected. So when standalone, we
	// list our own cluster address as the sole route: nats-server always
	// excludes self-addressed routes from actual connection attempts
	// (route.go's routesToSelf/excludedAddresses), so this never dials
	// anything — it only satisfies the startup check while we wait to
	// accept a real incoming route.
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

	if standalone {
		utils.Log("[NATS] Constellation has no known peers yet, listening for cluster routes without dialing out")
		selfRoute, err := url.Parse("nats-route://" + device.IP + ":6222")
		if err != nil {
			utils.Error("[NATS] Failed to build self route", err)
			return
		}
		opts.Routes = []*url.URL{selfRoute}
	} else {
		opts.Routes = cips
	}

	// Create and start the embedded NATS server
	retries := 0
	err = errors.New("")

	for err != nil && retries < 5 {
		// drop any instance left over from a failed attempt
		if ns != nil {
			ns.Shutdown()
			ns.WaitForShutdown()
			ns = nil
		}

		ns, err = server.NewServer(opts)
		if err != nil {
			retries++
			continue
		}

		if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
			ns.SetLogger(natsLogAdapter{}, true, true)
		}

		if !NebulaStarted {
			utils.Error("[NATS] Nebula not started, aborting NATS server setup", nil)
			ns.Shutdown()
			ns = nil
			return
		}

		go ns.Start()

		// Wait for the server to be ready
		if !ns.ReadyForConnections(time.Duration(2*(retries+1)) * time.Second) {
			retries++
			utils.Debug("[NATS] NATS server not ready...")
			err = errors.New("NATS server not ready")
			continue
		}

		utils.Debug("[NATS] Retrying to start NATS server")
	}

	if err != nil {
		if ns != nil {
			ns.Shutdown()
			ns.WaitForShutdown()
			ns = nil
		}
		utils.MajorError("[NATS] Error starting NATS server", err)
		return
	}

	NATSStarted = true

	utils.Log("[NATS] Started NATS server on host " + opts.Host + ":" + strconv.Itoa(opts.Port))
	InitNATSClient()
}

func StopNATS() {
	utils.Log("[NATS] Stopping NATS server...")

	if ns != nil {
		ns.Shutdown()
		ns.WaitForShutdown()
		ns = nil
	}
	NATSStarted = false
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

		nats.NoEcho(),
	)
}

func InitNATSClient() error {
	if !NATSStarted {
		utils.Warn("[NATS] NATS server not started, cannot initialize client")
		return errors.New("NATS server not started, cannot initialize client")
	}

	clientConfigLock.Lock()
	defer clientConfigLock.Unlock()

	if nc != nil {
		return errors.New("NATS client already initialized")
	}

	var err error
	retries := 0

	if !NebulaStarted {
		utils.Error("[NATS] Nebula not started, aborting NATS client connection", nil)
		return errors.New("Nebula not started, aborting NATS client connection")
	}

	utils.Log("[NATS] Connecting to NATS server...")

	time.Sleep(2 * time.Second)

	user, pwd, err := GetNATSCredentials()
	user = sanitizeNATSUsername(user)

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

		if !NebulaStarted {
			utils.Error("[NATS] Nebula not started, aborting NATS client connection retry", nil)
			nc = nil
			return errors.New("Nebula not started, aborting NATS client connection retry")
		}

		clientConfigLock.Unlock()
		time.Sleep(time.Duration(2*(retries+1)) * time.Second)
		clientConfigLock.Lock()

		if !NebulaStarted {
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

	// standalone has no peers to heartbeat with or sync from
	if IsConstellationStandalone() {
		utils.Debug("[NATS] Standalone constellation, skipping heartbeat and sync request")
	} else {
		go ClientHeartbeatInit()

		go SendRequestSyncMessage()
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
	for _, device := range CachedDevices {
		if device.IP == myIP {
			continue
		}
		if device.CosmosNode > 0 {
			return false
		}
	}
	// Before the initial sync, CachedDevices only knows about this server.
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

func CloseNATSClient() {
	utils.Log("Closing NATS client...")

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

	SyncNATSClientRouter(nc)

	// Scheduler: subscribe this node to its own per-target deployment command
	// subject so the leader can dispatch apply/remove here.
	if device, err := GetCurrentDevice(); err == nil {
		self := sanitizeNATSUsername(device.DeviceName)
		if subErr := pro.RegisterNodeDispatchHandler(nc, self); subErr != nil {
			utils.Warn("[SCHED-NODE] failed to register dispatch handler: " + subErr.Error())
		}
	} else {
		utils.Warn("[SCHED-NODE] cannot register dispatch handler: GetCurrentDevice failed: " + err.Error())
	}
}

func PingNATSClient() bool {
	if IsConstellationStandalone() {
		return true
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
