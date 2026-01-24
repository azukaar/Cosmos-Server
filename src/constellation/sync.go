package constellation

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	"github.com/azukaar/cosmos-server/src/utils"
)

type SyncPayload struct {
	Database       string `json:"database"`
	AuthPrivateKey string `json:"authPrivateKey"`
	AuthPublicKey  string `json:"authPublicKey"`
	LastEdited     int64  `json:"lastEdited"`
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

	sendPayload := SyncPayload{
		Database:       dbBase64,
		AuthPrivateKey: AuthPrivateKey,
		AuthPublicKey:  AuthPublicKey,
		LastEdited:    utils.GetFileLastModifiedTime(dbPath).Unix(),
	}

	// JSON encode the payload
	payloadBytes, err := json.Marshal(sendPayload)
	if err != nil {
		utils.Error("Constellation: MakeSyncPayload: Failed to marshal payload", err)
		return ""
	}

	return string(payloadBytes)
}

func ReceiveSyncPayload(rawPayload string) {
	utils.Log("Constellation: ReceiveSyncPayload: Received sync payload")
	
	var payload SyncPayload
	err := json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		utils.Error("Constellation: ReceiveSyncPayload: Failed to unmarshal payload " + rawPayload, err)
		return
	}

	// Decode base64 database content
	dbData, err := base64.StdEncoding.DecodeString(payload.Database)
	if err != nil {
		utils.Error("Constellation: ReceiveSyncPayload: Failed to decode database content", err)
		return
	}

	// Write database file
	dbPath := utils.CONFIGFOLDER + "database"

	// if local database is newer or same, skip update
	if utils.FileExists(dbPath) && utils.GetFileLastModifiedTime(dbPath).Unix() >= payload.LastEdited {
		utils.Warn("Constellation: ReceiveSyncPayload: Local database is newer or same as received one, skipping update")
		return
	}

	err = ioutil.WriteFile(dbPath, dbData, 0644)
	if err != nil {
		utils.Error("Constellation: ReceiveSyncPayload: Failed to write database file", err)
		return
	}

	utils.Warn("Constellation: ReceiveSyncPayload: Database file updated")

	// Update auth keys
	config := utils.ReadConfigFromFile()
	config.HTTPConfig.AuthPrivateKey = payload.AuthPrivateKey
	config.HTTPConfig.AuthPublicKey = payload.AuthPublicKey
	utils.SetBaseMainConfig(config)

	utils.CloseEmbeddedDB()

	utils.TriggerEvent(
		"cosmos.settings",
		"Settings updated",
		"success",
		"",
		map[string]interface{}{},
	)

	go func() {
		utils.RestartHTTPServer()
	}()
}

func SendRequestSyncMessage() {
	if !NebulaStarted {
		utils.Warn("Constellation: SendRequestSyncMessage: Nebula not started, skipping sync request")
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

	response, err := SendNATSMessage("cosmos._global_.constellation.data.sync-request", payloadStr)

	if err != nil {
		utils.Error("Constellation: SendRequestSyncMessage: Failed to send request", err)
		return
	}

	utils.Log("Constellation: SendRequestSyncMessage: Received sync response")

	ReceiveSyncPayload(string(response))
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

	response, err := SendNATSMessage("cosmos._global_.constellation.data.sync-receive", payload)

	if err != nil {
		utils.Error("Constellation: SendNewDBSyncMessage: Failed to send request", err)
		return
	}

	ReceiveSyncPayload(string(response))
}
