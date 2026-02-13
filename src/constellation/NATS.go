package constellation

import (
	"time"
	"errors"
	"strconv"
	"sync"
	"strings"
	"crypto/tls"
	"encoding/pem"
	"io/ioutil"
	"fmt"
	"gopkg.in/yaml.v2"
    "net/url"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/azukaar/cosmos-server/src/utils"
	
	natsClient "github.com/nats-io/nats.go"
)

type NodeHeartbeat struct {
	DeviceName string
	IP string
	IsRelay bool
	IsLighthouse bool
	IsExitNode bool
	CosmosNode int
	Tunnels []utils.ProxyRouteConfig
}

var ns *server.Server

func truncateLog(s string) string {
	if len(s) > 100 {
		return s[:100] + "..."
	}
	return s
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

	// read IPs from cached devices
	for _, device := range CachedDevices {
		if device.IsLighthouse {
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

	utils.Debug("GetNATSCredentials: currentDevice.APIKey=" + currentDevice.APIKey + " currentDevice.DeviceName=" + currentDevice.DeviceName)

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

		utils.Debug("GetNATSCredentials: found credentials in nebula.yml: deviceName=" + deviceName + " apiKey=" + apiKey)

		return deviceName, apiKey, nil
	}

	return "", "", errors.New("NATS credentials not found")
}

func StartNATS() {
	if ns != nil {
		return
	}
	
	ip,err := GetCurrentDeviceIP()
	if err != nil {
		utils.Error("[NATS] Failed to get current device IP", err)
		return
	}

	utils.Log("[NATS] Starting NATS server on " + ip + ":4222")

	time.Sleep(2 * time.Second)
	
	config := utils.GetMainConfig()
	HTTPConfig := config.HTTPConfig

	var tlsCert = HTTPConfig.TLSCert
	var tlsKey= HTTPConfig.TLSKey

	// Ensure the PEM data is correctly formatted
	certPEMBlock := []byte(tlsCert)
	keyPEMBlock := []byte(tlsKey)

	// Decode PEM encoded certificate
	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil {
			utils.MajorError("[NATS] Failed to start NATS: parse certificate PEM", nil)
	}

	// Decode PEM encoded private key
	keyDERBlock, _ := pem.Decode(keyPEMBlock)
	if keyDERBlock == nil {
			utils.MajorError("[NATS] Failed to start NATS: parse key PEM", nil)
	}

	// Create tls.Certificate using the original PEM data
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
			utils.MajorError("[NATS] Failed to start NATS: create TLS certificate", err)
	}

	// Configure the NATS server options
	// Make users
	
	users := []*server.User{}

	for _, devices := range CachedDevices {
		utils.Debug("[NATS] Adding NATS user for device: " + devices.DeviceName + " With API Key: " + devices.APIKey)
		username := sanitizeNATSUsername(devices.DeviceName)
		
		// TODO: Agent / users with less permissions 

		users = append(users, &server.User{
			Username: username,
			Password: devices.APIKey,
			Permissions: &server.Permissions{
				Publish: &server.SubjectPermission{
						Allow: []string{
							"cosmos."+username+".>", "_INBOX.>",
							"cosmos._global_.>", 
							"_INBOX.>",
							"$KV.constellation-nodes.>",
		                    "$JS.API.STREAM.INFO.>",
							"$JS.API.>", 
						},
				},
				Subscribe: &server.SubjectPermission{
						Allow: []string{
							"cosmos."+username+".>", "_INBOX.>",
							"cosmos._global_.>",
							"_INBOX.>",
		                    "$KV.constellation-nodes.>",
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
	natsName:= device.DeviceName

	// if debug, add debug user 
	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		users = append(users, &server.User{
			Username: "DEBUG",
			Password: "DEBUG",
			Permissions: nil,
		})

		natsHost = "0.0.0.0"
	}

	cips, err := GetClusterIPs()
	if err != nil {
		utils.Error("[NATS] Failed to get cluster IPs", err)
	}

	// if only one, abort
	if (len(cips) == 0) {
		utils.Warn("[NATS] No cluster IPs found, NATS server will not start")
		utils.Debug("[NATS] Current cached devices: ")
		for _, d := range cips {
			utils.Debug("[NATS] Device: " + d.String())
		}
		return
	}

	opts := &server.Options{
		Host: natsHost,
		Port: 4222,

		ServerName: natsName,

	    JetStream: true,
    	StoreDir:  utils.CONFIGFOLDER + "/jetstream",

		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.NoClientCert,
			InsecureSkipVerify: true,
		},

		Cluster: server.ClusterOpts{
        	Name: "Constellation",
			Host: device.IP,
			Port: 6222,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientAuth:   tls.NoClientCert,
				InsecureSkipVerify: true,
			},
		},

		Routes: cips,

		Users: users,
	}

	// Create and start the embedded NATS server
	retries := 0
	err = errors.New("")

	for err != nil && retries < 5 {
		ns, err = server.NewServer(opts)
		if err != nil {
			retries++
			continue
		}

		if !NebulaStarted {
			utils.Error("[NATS] Nebula not started, aborting NATS server setup", nil)
			return
		}

		go ns.Start()

		// Wait for the server to be ready
		if !ns.ReadyForConnections(time.Duration(2 * (retries + 1)) * time.Second) {
			retries++
			utils.Debug("[NATS] NATS server not ready...")
			err = errors.New("NATS server not ready")
			continue
		}

		utils.Debug("[NATS] Retrying to start NATS server")
	}

	if err != nil {
		utils.MajorError("[NATS] Error starting NATS server", err)
	} else {
		utils.Log("[NATS] Started NATS server on host " + opts.Host + ":" + strconv.Itoa(opts.Port))
		InitNATSClient()
	}
}

func StopNATS() {
	utils.Log("[NATS] Stopping NATS server...")

	if ns != nil {
		ns.Shutdown()
		ns.WaitForShutdown()
		ns = nil
	}
}

// sync lock - RWMutex allows multiple readers, single writer
var clientConfigLock = sync.RWMutex{}
var NATSClientTopic = ""
var nc *nats.Conn
var js nats.JetStreamContext
func InitNATSClient() error {
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

	nc, err = natsClient.Connect("nats://"+deviceIp+":4222",

		// nats.DisconnectHandler(func(nc *nats.Conn) {
		// 		utils.Log("Disconnected from NATS server - trying to reconnect")
		// }),
	
		nats.Secure(&tls.Config{
			InsecureSkipVerify: true,
		}),

		nats.UserInfo(user, pwd),
		
		// timeout
		nats.Timeout(2*time.Second),

    	nats.NoEcho(), 
	)

	for err != nil {
		if retries == 10 {
			utils.MajorError("[NATS] Error connecting to Constellation NATS server after 10 tries", err)
		}
		
		if !NebulaStarted {
			utils.Error("[NATS] Nebula not started, aborting NATS client connection retry", nil)
			nc = nil
			return errors.New("Nebula not started, aborting NATS client connection retry")
		}

		time.Sleep(time.Duration(2 * (retries + 1)) * time.Second)

		if !NebulaStarted {
			retries++
			utils.Warn("[NATS] Nebula not started yet, delaying NATS client connection retry")
			continue
		}

		nc, err = natsClient.Connect("nats://localhost:4222",
			nats.Secure(&tls.Config{
				InsecureSkipVerify: true,
			}),

			nats.UserInfo(user, pwd),

			// timeout
			nats.Timeout(2*time.Second),
		
    		nats.NoEcho(), 
		)
		
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

	go ClientHeartbeatInit()

	go SendRequestSyncMessage()

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
    js, err = nc.JetStream(nats.MaxWait(10 * time.Second))
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

func CloseNATSClient() {
	utils.Log("Closing NATS client...")

	StopHeartbeat()

	clientConfigLock.Lock()
	defer clientConfigLock.Unlock()

	if nc != nil {
		nc.Close()
		nc = nil
	}
	js = nil
}

func SendNATSMessage(topic string, payload string) (string, error) {
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
}

func PingNATSClient() bool {
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