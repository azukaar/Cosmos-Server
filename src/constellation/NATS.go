package constellation

import (
	"time"
	"errors"
	"strconv"
	"sync"
	"strings"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"gopkg.in/yaml.v2"

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
	IsCosmosNode bool
}

var ns *server.Server

func sanitizeNATSUsername(username string) string {
	username = strings.ReplaceAll(username, " ", "_")
	username = strings.ReplaceAll(username, ".", "_")
	username = strings.ReplaceAll(username, "-", "_")
	username = strings.ReplaceAll(username, ":", "_")
	username = strings.ReplaceAll(username, "/", "_")
	username = strings.ReplaceAll(username, "\\", "_")
	return username
}

func GetNATSCredentials(isMaster bool) (string, string, error) {
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

	// if debug, add debug user 
	for _, devices := range CachedDevices {
		utils.Debug("[NATS] Adding NATS user for device: " + devices.DeviceName + " With API Key: " + devices.APIKey)
		username := sanitizeNATSUsername(devices.DeviceName)
		
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
						},
				},
				Subscribe: &server.SubjectPermission{
						Allow: []string{
							"cosmos."+username+".>", "_INBOX.>",
							"cosmos._global_.>",
							"_INBOX.>",
		                    "$KV.constellation-nodes.>",
		                    "$JS.API.STREAM.INFO.>",
						},
				},
			},
		})
	}

	natsHost, err := GetCurrentDeviceIP()
	if err != nil {
		utils.Error("[NATS] Failed to get current device IP", err)
		return
	}

	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		users = append(users, &server.User{
			Username: "DEBUG",
			Password: "DEBUG",
			Permissions: nil,
		})

		natsHost = "0.0.0.0"
	}


	opts := &server.Options{
		Host: natsHost,
		Port: 4222,

	    JetStream: true,
    	StoreDir:  utils.CONFIGFOLDER + "/jetstream",

		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.NoClientCert,
			InsecureSkipVerify: true,
		},

		// Cluster: server.ClusterOpts{
		// 	Host: GetCurrentDeviceIP(),
		// 	Port: 6222,
		// 	TLSConfig: &tls.Config{
		// 		Certificates: []tls.Certificate{cert},
		// 		ClientAuth:   tls.NoClientCert,
		// 		InsecureSkipVerify: true,
		// 	},
		// },

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
		if NebulaFailedStarting {
			utils.Error("[NATS] Nebula failed to start, aborting NATS server setup", nil)
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

// sync lock
var clientConfigLock = sync.Mutex{}
var NATSClientTopic = ""
var nc *nats.Conn
var js nats.JetStreamContext
func InitNATSClient() {
	if nc != nil {
		return
	}

	clientConfigLock.Lock()
	defer clientConfigLock.Unlock()

	var err error
	retries := 0

	if NebulaFailedStarting {
		utils.Error("[NATS] Nebula failed to start, aborting NATS client connection", nil)
		return
	}

	utils.Log("[NATS] Connecting to NATS server...")
	
	time.Sleep(2 * time.Second)
	
	user, pwd, err := GetNATSCredentials(!utils.GetMainConfig().ConstellationConfig.SlaveMode)
	user = sanitizeNATSUsername(user)
	
	if err != nil {
		utils.MajorError("[NATS] Error getting constellation credentials", err)
		return
	}

	nc, err = natsClient.Connect("nats://192.168.201.1:4222",

		// nats.DisconnectHandler(func(nc *nats.Conn) {
		// 		utils.Log("Disconnected from NATS server - trying to reconnect")
		// }),
	
		nats.Secure(&tls.Config{
			InsecureSkipVerify: true,
		}),

		nats.UserInfo(user, pwd),
		
	)

	for err != nil {
		if retries == 10 {
			utils.MajorError("[NATS] Error connecting to Constellation NATS server (timeout) - will continue trying", err)
		}
		
		if retries >= 11 {
			retries = 11
		}

		if NebulaFailedStarting {
			utils.Error("[NATS] Nebula failed to start, aborting NATS client connection", nil)
			return
		}

		time.Sleep(time.Duration(2 * (retries + 1)) * time.Second)

		nc, err = natsClient.Connect("nats://192.168.201.1:4222",
			nats.Secure(&tls.Config{
				InsecureSkipVerify: true,
			}),

			nats.UserInfo(user, pwd),

			// timeout
			nats.Timeout(2*time.Second),
		)
		
		if err != nil {
			retries++
			utils.Debug("[NATS] Retrying to start NATS Client: " + err.Error())
		}
	}

	if err != nil {
		utils.MajorError("[NATS] Error connecting to Constellation NATS server", err)
		return
	} else {
		utils.Log("[NATS] Connected to NATS server as " + user)
		NATSClientTopic = "cosmos." + user
	}

	utils.Debug("[NATS] NATS client connected")

	js, err = nc.JetStream(nats.MaxWait(10 * time.Second)) 
	if err != nil {
		utils.MajorError("[NATS] Error getting JetStream context", err)
	}

	utils.Debug("[NATS] JetStream context obtained")

	go MasterNATSClientRouter()

	ClientHeartbeatInit()

	go SendRequestSyncMessage()

	// POST CLIENT CONNECTION HOOK
}

func ClientHeartbeatInit() {
	var kv nats.KeyValue
	var err error

	kv, err = js.KeyValue("constellation-nodes")
	if err != nil {
		time.Sleep(2 * time.Second)
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket: "constellation-nodes",
			TTL:    10 * time.Second,
		})
		if err != nil {
			utils.MajorError("[NATS] Error creating Key-Value store", err)
			return
		}
	}
	
	utils.Debug("[NATS] Key-Value store 'constellation-nodes' ready")

	go func() {
		
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			device, err := GetCurrentDevice()
			if err != nil {
				utils.Error("[NATS] Error getting current device for heartbeat", err)
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

func IsClientConnected() bool {
	clientConfigLock.Lock()
	defer clientConfigLock.Unlock()

	if nc == nil {
		return false
	}

	isc := nc.IsConnected()

	if !isc {
		nc = nil
	}

	return isc
}

func CloseNATSClient() {
	utils.Log("Closing NATS client...")

	nc.Close()
	nc = nil
}

func SendNATSMessage(topic string, payload string) (string, error) {
	if !IsClientConnected() {
		utils.Warn("NATS client not connected")
		InitNATSClient()
	}

	utils.Debug("[MQ] Sending message to topic: " + topic)

	// Send a request and wait for a response
	msg, err := nc.Request(topic, []byte(payload), 2*time.Second)
	if err != nil {
		utils.Error("[MQ] Error sending request", err)
		return "", err
	}

	utils.Debug("[MQ] Received response: " +  string(msg.Data))

	return string(msg.Data), nil
}

func PublishNATSMessage(topic string, payload string) error {
	if !IsClientConnected() {
		utils.Warn("NATS client not connected")
		InitNATSClient()
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
		utils.Debug("[MQ] Received: " + string(m.Data) + " from " + m.Subject)
		m.Respond([]byte("Pong"))
	})

	nc.Subscribe("cosmos._global_.constellation.data.sync-request", func(m *nats.Msg) {
		utils.Log("[NATS] Constellation data sync request received")

		payload := m.Data
		response := MakeSyncPayload((string)(payload))
		if response != "" {
			m.Respond([]byte(response))
		}
	})
	
	nc.Subscribe("cosmos._global_.constellation.data.sync-receive", func(m *nats.Msg) {
		utils.Log("[NATS] Constellation data sync received")
		
		payload := m.Data
		ReceiveSyncPayload((string)(payload))
	})
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