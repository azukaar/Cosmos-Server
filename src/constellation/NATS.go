package constellation

import (
	"time"
	"errors"
	"strconv"
	"sync"
	"strings"
	"crypto/tls"

	"encoding/pem"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/azukaar/cosmos-server/src/utils"
	
	natsClient "github.com/nats-io/nats.go"
)

var ns *server.Server
var MASTERUSER = "SERVERUSER"
var MASTERPWD = utils.GenerateRandomString(24)

func sanitizeNATSUsername(username string) string {
	username = strings.ReplaceAll(username, " ", "_")
	username = strings.ReplaceAll(username, ".", "_")
	username = strings.ReplaceAll(username, "-", "_")
	username = strings.ReplaceAll(username, ":", "_")
	username = strings.ReplaceAll(username, "/", "_")
	username = strings.ReplaceAll(username, "\\", "_")
	return username
}

func StartNATS() {
	if ns != nil {
		return
	}
	
	utils.Log("Starting NATS server...")

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
			utils.MajorError("Failed to start NATS: parse certificate PEM", nil)
	}

	// Decode PEM encoded private key
	keyDERBlock, _ := pem.Decode(keyPEMBlock)
	if keyDERBlock == nil {
			utils.MajorError("Failed to start NATS: parse key PEM", nil)
	}

	// Create tls.Certificate using the original PEM data
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
			utils.MajorError("Failed to start NATS: create TLS certificate", err)
	}

	// Configure the NATS server options
	// Make users
	
	users := []*server.User{
		&server.User{
			Username: MASTERUSER,
			Password: MASTERPWD,
			Permissions: nil,
		},
	}

	for _, devices := range CachedDevices {
		username := sanitizeNATSUsername(devices.DeviceName)
		
		users = append(users, &server.User{
			Username: username,
			Password: devices.APIKey,
			Permissions: &server.Permissions{
				Publish: &server.SubjectPermission{
						Allow: []string{"cosmos."+username+".>", "_INBOX.>"},
				},
				Subscribe: &server.SubjectPermission{
						Allow: []string{"cosmos."+username+".>", "_INBOX.>"},
				},
			},
		})
	}

	opts := &server.Options{
		Host: "192.168.201.1",
		Port: 4222,

		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.NoClientCert,
			InsecureSkipVerify: true,
		},

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
			utils.Error("Nebula failed to start, aborting NATS server setup", nil)
			return
		}

		go ns.Start()

		// Wait for the server to be ready
		if !ns.ReadyForConnections(time.Duration(2 * (retries + 1)) * time.Second) {
			retries++
			utils.Debug("NATS server not ready...")
			err = errors.New("NATS server not ready")
			continue
		}

		utils.Debug("Retrying to start NATS server")
	}

	if err != nil {
		utils.MajorError("Error starting NATS server", err)
	} else {
		utils.Log("Started NATS server on host " + opts.Host + ":" + strconv.Itoa(opts.Port))
		if !utils.GetMainConfig().ConstellationConfig.SlaveMode {
			InitNATSClient()
		}
	}
}

func StopNATS() {
	utils.Log("Stopping NATS server...")

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
func InitNATSClient() {
	if nc != nil {
		return
	}

	clientConfigLock.Lock()
	defer clientConfigLock.Unlock()

	var err error
	retries := 0

	if NebulaFailedStarting {
		utils.Error("Nebula failed to start, aborting NATS client connection", nil)
		return
	}

	utils.Log("Connecting to NATS server...")
	
	time.Sleep(2 * time.Second)
	
	user, pwd, err := GetNATSCredentials(!utils.GetMainConfig().ConstellationConfig.SlaveMode)
	user = sanitizeNATSUsername(user)
	
	if err != nil {
		utils.MajorError("Error getting constellation credentials", err)
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
			utils.MajorError("Error connecting to Constellation NATS server (timeout) - will continue trying", err)
		}
		
		if retries >= 11 {
			retries = 11
		}

		if NebulaFailedStarting {
			utils.Error("Nebula failed to start, aborting NATS client connection", nil)
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
			utils.Debug("Retrying to start NATS Client: " + err.Error())
		}
	}

	if err != nil {
		utils.MajorError("Error connecting to Constellation NATS server", err)
		return
	} else {
		utils.Log("Connected to NATS server as " + user)
		NATSClientTopic = "cosmos." + user
	}

	utils.Debug("NATS client connected")

	if !utils.GetMainConfig().ConstellationConfig.SlaveMode {
		go MasterNATSClientRouter()
	} else {
		clientConfigLock.Unlock()
		go SlaveNATSClientRouter()
		SendNATSMessage("cosmos."+user+".debug", "NATS Client connected as " + user)
		RequestSyncPayload()
		clientConfigLock.Lock()
	}
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
		utils.Error("Error sending request", err)
		return err
	}

	return nil
}

func MasterNATSClientRouter() {
	utils.Log("Starting NATS Master client router.")

	nc.Subscribe("cosmos."+MASTERUSER+".ping", func(m *nats.Msg) {
		utils.Debug("[MQ] Received: " + string(m.Data) + " from " + m.Subject)
		m.Respond([]byte("Pong"))
	})

	for _, devices := range CachedDevices {
		localDevice := devices
		username := sanitizeNATSUsername(localDevice.DeviceName)
		
		nc.Subscribe("cosmos."+username+".debug", func(m *nats.Msg) {
			utils.Debug("[MQ] Received: " + string(m.Data))
			m.Respond([]byte("Received: " + string(m.Data)))
		})

		nc.Subscribe("cosmos."+username+".ping", func(m *nats.Msg) {
			utils.Debug("[MQ] Received: " + string(m.Data) + " from " + m.Subject)
			m.Respond([]byte("Pong"))
		})

		nc.Subscribe("cosmos."+username+".constellation.config", func(m *nats.Msg) {
			utils.Debug("[MQ] Received: " + string(m.Data) + " from " + m.Subject)

			res, err := GetDeviceConfigForSync(localDevice.Nickname, localDevice.DeviceName)
			if err != nil {
				utils.Error("Error getting device config for sync", err)
			} else {
				m.Respond([]byte(res))
			}
		})

		if localDevice.IsCosmosNode {
			nc.Subscribe("cosmos."+username+".constellation.data.sync-request", func(m *nats.Msg) {
				if !utils.GetMainConfig().ConstellationConfig.SlaveMode && !utils.GetMainConfig().ConstellationConfig.DoNotSyncNodes {
					utils.Debug("[MQ] Received: " + string(m.Data) + " from " + m.Subject)
					m.Respond([]byte(MakeSyncPayload()))
				}
			})
		}
	}
}

func SlaveNATSClientRouter() {
	utils.Log("Starting NATS Slave client router.")

	username := sanitizeNATSUsername(DeviceName)

	nc.Subscribe("cosmos."+username+".constellation.config.resync", func(m *nats.Msg) {
		utils.Log("Constellation config changed, resyncing...")
		
		config := m.Data

		needRestart, err := SlaveConfigSync((string)(config))

		if err != nil {
			utils.MajorError("Error re-syncing Constellation config, please manually sync", err)
		} else {
			if needRestart {
				utils.Warn("Slave config has changed, restarting Nebula...")
				RestartNebula()
				utils.RestartHTTPServer()
			}
		}
	})
	
	nc.Subscribe("cosmos."+username+".constellation.data.sync-receive", func(m *nats.Msg) {
		utils.Log("Constellation data sync received")
		
		payload := m.Data

		ReceiveSyncPayload((string)(payload))
	})
}

func PingNATSClient() bool {
	user, _, err := GetNATSCredentials(!utils.GetMainConfig().ConstellationConfig.SlaveMode)
	if err != nil {
		utils.Error("Error getting constellation credentials", err)
		return false
	}

	user = sanitizeNATSUsername(user)

	response, err := SendNATSMessage("cosmos."+user+".ping", "Ping")
	if err != nil {
		utils.Error("Error pinging NATS client", err)
		return false
	}

	if response != "" {
		utils.Debug("NATS client response: " + response)
		return true
	}

	return false
}