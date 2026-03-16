package configapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/utils"
)

var validTokenName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type CreateAPITokenRequest struct {
	Name                    string   `json:"name"`
	Description             string   `json:"description,omitempty"`
	ReadOnly                bool     `json:"readOnly"`
	IPWhitelist             []string `json:"ipWhitelist,omitempty"`
	RestrictToConstellation bool     `json:"restrictToConstellation"`
}

type DeleteAPITokenRequest struct {
	Name string `json:"name"`
}

func APITokenRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		listAPITokens(w, req)
	} else if req.Method == "POST" {
		createAPIToken(w, req)
	} else if req.Method == "DELETE" {
		deleteAPIToken(w, req)
	} else {
		utils.Error("APITokenRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

func listAPITokens(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN_READ) != nil {
		return
	}

	config := utils.GetMainConfig()

	tokens := make(map[string]interface{})
	for name, tok := range config.APITokens {
		tokens[name] = map[string]interface{}{
			"name":                    tok.Name,
			"description":             tok.Description,
			"owner":                   tok.Owner,
			"permissions":             tok.Permissions,
			"ipWhitelist":             tok.IPWhitelist,
			"restrictToConstellation": tok.RestrictToConstellation,
			"createdAt":               tok.CreatedAt,
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   tokens,
	})
}

func createAPIToken(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN) != nil {
		return
	}

	var request CreateAPITokenRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		utils.Error("CreateAPIToken: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "AT010")
		return
	}

	if request.Name == "" {
		utils.HTTPError(w, "Token name is required", http.StatusBadRequest, "AT011")
		return
	}

	if len(request.Name) > 64 || !validTokenName.MatchString(request.Name) {
		utils.HTTPError(w, "Token name must be 1-64 characters, alphanumeric, dash, or underscore", http.StatusBadRequest, "AT014")
		return
	}

	// Generate raw token: "cosmos_" + base64url(32 random bytes)
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		utils.Error("CreateAPIToken: Failed to generate random bytes", err)
		utils.HTTPError(w, "Internal error", http.StatusInternalServerError, "AT012")
		return
	}
	rawToken := "cosmos_" + base64.RawURLEncoding.EncodeToString(rawBytes)

	// Hash for storage
	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])

	// Build permissions array
	var permissions []utils.Permission
	if request.ReadOnly {
		permissions = []utils.Permission{
			utils.PERM_ADMIN_READ,
			utils.PERM_USERS_READ,
			utils.PERM_RESOURCES_READ,
			utils.PERM_CONFIGURATION_READ,
		}
	} else {
		permissions = []utils.Permission{
			utils.PERM_ADMIN_READ,
			utils.PERM_ADMIN,
			utils.PERM_USERS_READ,
			utils.PERM_USERS,
			utils.PERM_RESOURCES_READ,
			utils.PERM_RESOURCES,
			utils.PERM_CONFIGURATION_READ,
			utils.PERM_CONFIGURATION,
			utils.PERM_CREDENTIALS_READ,
		}
	}

	owner := utils.GetAuthContext(req).Nickname

	tokenConfig := utils.APITokenConfig{
		Name:                    request.Name,
		Description:             request.Description,
		Owner:                   owner,
		TokenHash:               tokenHash,
		Permissions:             permissions,
		IPWhitelist:             request.IPWhitelist,
		RestrictToConstellation: request.RestrictToConstellation,
		CreatedAt:               time.Now(),
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.APITokens == nil {
		config.APITokens = make(map[string]utils.APITokenConfig)
	}

	if _, exists := config.APITokens[request.Name]; exists {
		utils.HTTPError(w, "Token with this name already exists", http.StatusConflict, "AT013")
		return
	}

	config.APITokens[request.Name] = tokenConfig
	utils.SetBaseMainConfig(config)

	utils.TouchDatabase()
	go constellation.SendNewDBSyncMessage()

	utils.TriggerEvent(
		"cosmos.api.token.created",
		"API Token created",
		"important",
		"token@"+request.Name,
		map[string]interface{}{
			"tokenName": request.Name,
			"owner":     owner,
			"readOnly":  request.ReadOnly,
		},
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data": map[string]interface{}{
			"token": rawToken,
			"name":  request.Name,
		},
	})
}

func deleteAPIToken(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN) != nil {
		return
	}

	var request DeleteAPITokenRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		utils.Error("DeleteAPIToken: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "AT020")
		return
	}

	if request.Name == "" {
		utils.HTTPError(w, "Token name is required", http.StatusBadRequest, "AT021")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.APITokens == nil {
		utils.HTTPError(w, "Token not found", http.StatusNotFound, "AT022")
		return
	}

	if _, exists := config.APITokens[request.Name]; !exists {
		utils.HTTPError(w, "Token not found", http.StatusNotFound, "AT022")
		return
	}

	delete(config.APITokens, request.Name)
	utils.SetBaseMainConfig(config)

	utils.TouchDatabase()
	go constellation.SendNewDBSyncMessage()

	utils.TriggerEvent(
		"cosmos.api.token.deleted",
		"API Token deleted",
		"important",
		"token@"+request.Name,
		map[string]interface{}{
			"tokenName": request.Name,
		},
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}
