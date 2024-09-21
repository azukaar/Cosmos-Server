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
}

func MakeSyncPayload() string {
	utils.Log("Constellation: MakeSyncPayload: Making sync payload")
	
	// Read database file
	dbPath := utils.CONFIGFOLDER + "database"
	dbData, err := ioutil.ReadFile(dbPath)
	if err != nil {
		utils.Error("Constellation: MakeSyncPayload: Failed to read database file", err)
		return ""
	}

	// Encode database content to base64
	dbBase64 := base64.StdEncoding.EncodeToString(dbData)

	// Read auth keys
	AuthPrivateKey := utils.GetMainConfig().HTTPConfig.AuthPrivateKey
	AuthPublicKey := utils.GetMainConfig().HTTPConfig.AuthPublicKey

	payload := SyncPayload{
		Database:       dbBase64,
		AuthPrivateKey: AuthPrivateKey,
		AuthPublicKey:  AuthPublicKey,
	}

	// JSON encode the payload
	payloadBytes, err := json.Marshal(payload)
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
		utils.Error("Constellation: ReceiveSyncPayload: Failed to unmarshal payload", err)
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


func RequestSyncPayload() {
	user, _, err := GetNATSCredentials(!utils.GetMainConfig().ConstellationConfig.SlaveMode)
	if err != nil {
		utils.Error("Error getting constellation credentials", err)
		return
	}

	user = sanitizeNATSUsername(user)

	response, err := SendNATSMessage("cosmos."+user+".constellation.data.sync-request", "")

	if err != nil {
		utils.Error("Constellation: RequestSyncPayload: Failed to send request", err)
		return
	}

	ReceiveSyncPayload(string(response))
}

func SendSyncPayload(username string) {
	payload := MakeSyncPayload()

	response, err := SendNATSMessage("cosmos."+username+".constellation.data.sync-receive", payload)

	if err != nil {
		utils.Error("Constellation: SendSyncPayload: Failed to send payload", err)
		return
	}

	if string(response) != "OK" {
		utils.Error("Constellation: SendSyncPayload: Failed to send payload", err)
		return
	}
}