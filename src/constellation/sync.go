package constellation

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"time"
	"github.com/nats-io/nats.go"

	"github.com/azukaar/cosmos-server/src/utils"
)

type SyncPayload struct {
	Database       string `json:"database"`
	AuthPrivateKey string `json:"authPrivateKey"`
	AuthPublicKey  string `json:"authPublicKey"`
	CART           string `json:"caCrt"`
	CAKey 	       string `json:"caKey"`
	DNS 		   SyncDNSPayload `json:"dns"`
	LastEdited     int64  `json:"lastEdited"`
}

type SyncDNSPayload struct {
	DNSPort string `json:"dnsPort"`
	DNSFallback string `json:"dnsFallback"`
	DNSBlockBlacklist bool `json:"dnsBlockBlacklist"`
	DNSAdditionalBlocklists []string `json:"dnsAdditionalBlocklists"`
	CustomDNSEntries []utils.ConstellationDNSEntry `json:"customDNSEntries"`
}

type SyncRequestPayload struct {
	EditedAt int64 `json:"editedAt"`
}

func MakeSyncPayload(rawPayload string) string {
	utils.Log("Constellation: MakeSyncPayload: Making sync payload")
	
	// Read database file
	dbPath := utils.CONFIGFOLDER + "database"
	dbData, err := ioutil.ReadFile(dbPath)
	if err != nil {
		utils.Error("Constellation: MakeSyncPayload: Failed to read database file", err)
		return ""
	}
	
	if (rawPayload != "") {
		var payload SyncPayload
		err := json.Unmarshal([]byte(rawPayload), &payload)
		if err != nil {
			utils.Error("Constellation: ReceiveSyncPayload: Failed to unmarshal payload", err)
			return ""
		}

		editedAt := payload.LastEdited

		// If our local is older or same, skip sending
		if utils.GetFileLastModifiedTime(dbPath).Unix() <= editedAt {
			utils.Warn("Constellation: MakeSyncPayload: Local database is older or same as requested one, skipping sending")
			return ""
		}
	}

	// Encode database content to base64
	dbBase64 := base64.StdEncoding.EncodeToString(dbData)

	// Read auth keys
	AuthPrivateKey := utils.GetMainConfig().HTTPConfig.AuthPrivateKey
	AuthPublicKey := utils.GetMainConfig().HTTPConfig.AuthPublicKey

	// Read CA
	CART := ""
	CAKey := ""

	caCrtData, err := ioutil.ReadFile(utils.CONFIGFOLDER + "ca.crt")
	if err == nil {
		CART = base64.StdEncoding.EncodeToString(caCrtData)
	}

	caKeyData, err := ioutil.ReadFile(utils.CONFIGFOLDER + "ca.key")
	if err == nil {
		CAKey = base64.StdEncoding.EncodeToString(caKeyData)
	}

	// Create payload

	constellationConfig := utils.GetMainConfig().ConstellationConfig

	sendPayload := SyncPayload{
		Database:       dbBase64,
		AuthPrivateKey: AuthPrivateKey,
		AuthPublicKey:  AuthPublicKey,
		CART:           CART,
		CAKey:          CAKey,
		DNS: SyncDNSPayload{
			DNSPort:                constellationConfig.DNSPort,
			DNSFallback:            constellationConfig.DNSFallback,
			DNSBlockBlacklist:      constellationConfig.DNSBlockBlacklist,
			DNSAdditionalBlocklists: constellationConfig.DNSAdditionalBlocklists,
			CustomDNSEntries:       constellationConfig.CustomDNSEntries,
		},
		LastEdited:     utils.GetFileLastModifiedTime(dbPath).Unix(),
	}

	// JSON encode the payload
	payloadBytes, err := json.Marshal(sendPayload)
	if err != nil {
		utils.Error("Constellation: MakeSyncPayload: Failed to marshal payload", err)
		return ""
	}

	return string(payloadBytes)
}

var lastDBSyncTime time.Time

func ReceiveSyncPayload(rawPayload string) bool {
	utils.Log("Constellation: ReceiveSyncPayload: Received sync payload")
	
	var payload SyncPayload
	err := json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		utils.Error("Constellation: ReceiveSyncPayload: Failed to unmarshal payload " + rawPayload, err)
		return false
	}

	// Decode base64 database content
	dbData, err := base64.StdEncoding.DecodeString(payload.Database)
	if err != nil {
		utils.Error("Constellation: ReceiveSyncPayload: Failed to decode database content", err)
		return false
	}

	// Write database file
	dbPath := utils.CONFIGFOLDER + "database"

	// Check if local database has 0 constellation devices - if so, always accept sync
	skipDateCheck := false
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "devices")
	if errCo == nil {
		count, errCount := c.CountDocuments(nil, map[string]interface{}{
			"Blocked": false,
		})
		closeDb()
		if errCount == nil && count == 0 {
			utils.Log("Constellation: ReceiveSyncPayload: Local database has 0 devices, skipping date check")
			skipDateCheck = true
		}
	}

	// if local database is newer or same, skip update (unless we have 0 devices)
	if !skipDateCheck && utils.FileExists(dbPath) && utils.GetFileLastModifiedTime(dbPath).Unix() >= payload.LastEdited {
		utils.Warn("Constellation: ReceiveSyncPayload: Local database is newer or same as received one, skipping update -  local time:" + strconv.FormatInt(utils.GetFileLastModifiedTime(dbPath).Unix(), 10) + " received time:" + strconv.FormatInt(payload.LastEdited, 10))
		return false
	}

	err = ioutil.WriteFile(dbPath, dbData, 0644)
	if err != nil {
		utils.Error("Constellation: ReceiveSyncPayload: Failed to write database file", err)
		return false
	}

	// set the last modified time to the one received
	err = utils.SetFileLastModifiedTime(dbPath, payload.LastEdited)
	if err != nil {
		utils.Error("Constellation: ReceiveSyncPayload: Failed to set database file last modified time", err)
		return false
	}

	utils.Warn("Constellation: ReceiveSyncPayload: Database file updated -  local time:" + strconv.FormatInt(utils.GetFileLastModifiedTime(dbPath).Unix(), 10) + " received time:" + strconv.FormatInt(payload.LastEdited, 10))

	// set the CA filses if present
	if payload.CART != "" && payload.CAKey != "" {
		caCrtData, err := base64.StdEncoding.DecodeString(payload.CART)
		if err != nil {
			utils.Error("Constellation: ReceiveSyncPayload: Failed to decode CA crt", err)
		} else {
			err = ioutil.WriteFile(utils.CONFIGFOLDER + "ca.crt", caCrtData, 0644)
			if err != nil {
				utils.Error("Constellation: ReceiveSyncPayload: Failed to write CA crt file", err)
			} else {
				caKeyData, err := base64.StdEncoding.DecodeString(payload.CAKey)
				if err != nil {
					utils.Error("Constellation: ReceiveSyncPayload: Failed to decode CA key", err)
				} else {
					err = ioutil.WriteFile(utils.CONFIGFOLDER + "ca.key", caKeyData, 0644)
					if err != nil {
						utils.Error("Constellation: ReceiveSyncPayload: Failed to write CA key file", err)
					}
				}
			}
		}
	}

	// Update auth keys and DNS config
	config := utils.ReadConfigFromFile()
	config.HTTPConfig.AuthPrivateKey = payload.AuthPrivateKey
	config.HTTPConfig.AuthPublicKey = payload.AuthPublicKey
	config.ConstellationConfig.DNSPort = payload.DNS.DNSPort
	config.ConstellationConfig.DNSFallback = payload.DNS.DNSFallback
	config.ConstellationConfig.DNSBlockBlacklist = payload.DNS.DNSBlockBlacklist
	config.ConstellationConfig.DNSAdditionalBlocklists = payload.DNS.DNSAdditionalBlocklists
	config.ConstellationConfig.CustomDNSEntries = payload.DNS.CustomDNSEntries
	utils.SetBaseMainConfig(config)

	utils.CloseEmbeddedDB()

	lastDBSyncTime = time.Now()

	utils.TriggerEvent(
		"cosmos.settings",
		"Settings updated",
		"success",
		"",
		map[string]interface{}{},
	)

	return true
}

func SendRequestSyncMessage() {
	if !NebulaStarted {
		utils.Warn("Constellation: SendRequestSyncMessage: Nebula not started, skipping sync request")
		return
	}

	if !lastDBSyncTime.IsZero() && time.Since(lastDBSyncTime) < 10*time.Second {
		utils.Warn("Constellation: SendRequestSyncMessage: Last DB sync was less than 10s ago, skipping")
		return
	}

	utils.Log("Constellation: SendRequestSyncMessage: Requesting sync payload")
	
	payload := SyncRequestPayload{
		EditedAt: utils.GetFileLastModifiedTime(utils.CONFIGFOLDER + "database").Unix(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		utils.Error("Constellation: SendRequestSyncMessage: Failed to marshal request payload", err)
		return
	}

	payloadStr := string(payloadBytes)

	needRestart := false

	SendNATSMessageAllReply("cosmos._global_.constellation.data.sync-request", payloadStr, 2*time.Second, func(response string) {
		utils.Log("Constellation: SendRequestSyncMessage: Received sync response")
		needRestart = ReceiveSyncPayload(response) || needRestart
	})

	if needRestart {
		go func() {
			utils.RestartHTTPServer()
			RestartNebula()
		}()
	}
}

func SendNewDBSyncMessage() {
	if !NebulaStarted {
		return
	}

	utils.Log("Constellation: SendNewDBSyncMessage: sending sync payload")
	
	payload := MakeSyncPayload("")
	if payload == "" {
		utils.Warn("Constellation: SendNewDBSyncMessage: No sync payload to send")
		return
	}

	err := PublishNATSMessage("cosmos._global_.constellation.data.sync-receive", payload)

	if err != nil {
		utils.Error("Constellation: SendNewDBSyncMessage: Failed to send request", err)
		return
	}
}

func SyncNATSClientRouter(nc *nats.Conn) {
	nc.Subscribe("cosmos._global_.constellation.data.sync-request", func(m *nats.Msg) {
		utils.Log("[NATS] Constellation data sync request received")

		payload := m.Data
		response := MakeSyncPayload((string)(payload))
		if response != "" {
			m.Respond([]byte(response))
		}
	})

	nc.Subscribe("cosmos._global_.constellation.public-devices", PublicDeviceListNATS)
	
	nc.Subscribe("cosmos._global_.constellation.data.sync-receive", func(m *nats.Msg) {
		utils.Log("[NATS] Constellation data sync received")

		payload := m.Data
		if ReceiveSyncPayload((string)(payload)) {
			go RestartNebula()
		}

		m.Respond([]byte("ack"))
	})
}