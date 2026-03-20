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
	"github.com/gorilla/mux"
)

var validTokenName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// GenerateAPIToken creates a new API token with the given parameters and returns
// the raw token (to be shown once to the caller) and the config to store.
// The caller is responsible for persisting the config.
func GenerateAPIToken(name, description, owner string, permissions []utils.Permission) (string, utils.APITokenConfig, error) {
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", utils.APITokenConfig{}, err
	}
	rawToken := "cosmos_" + base64.RawURLEncoding.EncodeToString(rawBytes)
	tokenSuffix := rawToken[len(rawToken)-5:]

	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])

	tokenConfig := utils.APITokenConfig{
		Name:        name,
		Description: description,
		Owner:       owner,
		TokenHash:   tokenHash,
		TokenSuffix: tokenSuffix,
		Permissions: permissions,
		CreatedAt:   time.Now(),
	}

	return rawToken, tokenConfig, nil
}

type CreateAPITokenRequest struct {
	Name                    string             `json:"name" validate:"required"`
	Description             string             `json:"description,omitempty"`
	ReadOnly                bool               `json:"readOnly"`
	Permissions             []utils.Permission `json:"permissions,omitempty"`
	IPWhitelist             []string           `json:"ipWhitelist,omitempty"`
	RestrictToConstellation bool               `json:"restrictToConstellation"`
	ExpiryDays              int                `json:"expiryDays,omitempty"`
}

type DeleteAPITokenRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateAPITokenRequest struct {
	Description             *string            `json:"description,omitempty"`
	Permissions             []utils.Permission `json:"permissions,omitempty"`
	IPWhitelist             []string           `json:"ipWhitelist,omitempty"`
	RestrictToConstellation *bool              `json:"restrictToConstellation,omitempty"`
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

// listAPITokens godoc
// @Summary List all API tokens
// @Description Returns all configured API tokens (without the actual token hashes)
// @Tags api-tokens
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/api-tokens [get]
func listAPITokens(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN_READ) != nil {
		return
	}

	config := utils.GetMainConfig()

	tokens := make(map[string]interface{})
	for name, tok := range config.APITokens {
		entry := map[string]interface{}{
			"name":                    tok.Name,
			"description":             tok.Description,
			"owner":                   tok.Owner,
			"tokenSuffix":             tok.TokenSuffix,
			"permissions":             tok.Permissions,
			"ipWhitelist":             tok.IPWhitelist,
			"restrictToConstellation": tok.RestrictToConstellation,
			"createdAt":               tok.CreatedAt,
		}
		if !tok.ExpiresAt.IsZero() {
			entry["expiresAt"] = tok.ExpiresAt
		}
		tokens[name] = entry
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   tokens,
	})
}

// createAPIToken godoc
// @Summary Create a new API token
// @Description Generates a new API token with the specified permissions and returns the raw token (shown only once)
// @Tags api-tokens
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAPITokenRequest true "Token creation details"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/api-tokens [post]
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

	// Build permissions array
	var permissions []utils.Permission
	if len(request.Permissions) > 0 {
		permissions = request.Permissions
	} else if request.ReadOnly {
		permissions = []utils.Permission{
			utils.PERM_ADMIN_READ,
			utils.PERM_USERS_READ,
			utils.PERM_RESOURCES_READ,
			utils.PERM_CONFIGURATION_READ,
		}
	} else {
		permissions = utils.DefaultAdminTokenPermissions
	}

	owner := utils.GetAuthContext(req).Nickname

	rawToken, tokenConfig, err := GenerateAPIToken(request.Name, request.Description, owner, permissions)
	if err != nil {
		utils.Error("CreateAPIToken: Failed to generate token", err)
		utils.HTTPError(w, "Internal error", http.StatusInternalServerError, "AT012")
		return
	}

	// Set extra fields not handled by GenerateAPIToken
	tokenConfig.IPWhitelist = request.IPWhitelist
	tokenConfig.RestrictToConstellation = request.RestrictToConstellation
	if request.ExpiryDays > 0 {
		tokenConfig.ExpiresAt = time.Now().Add(time.Duration(request.ExpiryDays) * 24 * time.Hour)
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

	responseData := map[string]interface{}{
		"token":       rawToken,
		"name":        request.Name,
		"tokenSuffix": tokenConfig.TokenSuffix,
	}
	if !tokenConfig.ExpiresAt.IsZero() {
		responseData["expiresAt"] = tokenConfig.ExpiresAt
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   responseData,
	})
}

// deleteAPIToken godoc
// @Summary Delete an API token
// @Description Removes an API token by name
// @Tags api-tokens
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DeleteAPITokenRequest true "Token name to delete"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/api-tokens [delete]
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

func APITokenIdRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "PUT" {
		updateAPIToken(w, req)
	} else {
		utils.Error("APITokenIdRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

// updateAPIToken godoc
// @Summary Update an API token
// @Description Updates an existing API token's description, permissions, IP whitelist, or constellation restriction
// @Tags api-tokens
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Token name"
// @Param request body UpdateAPITokenRequest true "Fields to update"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Router /api/api-tokens/{name} [put]
func updateAPIToken(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_ADMIN) != nil {
		return
	}

	name := mux.Vars(req)["name"]
	if name == "" {
		utils.HTTPError(w, "Token name is required", http.StatusBadRequest, "AT030")
		return
	}

	var request UpdateAPITokenRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		utils.Error("UpdateAPIToken: Invalid request", err)
		utils.HTTPError(w, "Invalid request", http.StatusBadRequest, "AT031")
		return
	}

	utils.ConfigLock.Lock()
	defer utils.ConfigLock.Unlock()

	config := utils.ReadConfigFromFile()

	if config.APITokens == nil {
		utils.HTTPError(w, "Token not found", http.StatusNotFound, "AT032")
		return
	}

	token, exists := config.APITokens[name]
	if !exists {
		utils.HTTPError(w, "Token not found", http.StatusNotFound, "AT032")
		return
	}

	// Update only the fields that were provided — token hash stays the same
	if request.Description != nil {
		token.Description = *request.Description
	}
	if request.Permissions != nil {
		token.Permissions = request.Permissions
	}
	if request.IPWhitelist != nil {
		token.IPWhitelist = request.IPWhitelist
	}
	if request.RestrictToConstellation != nil {
		token.RestrictToConstellation = *request.RestrictToConstellation
	}

	config.APITokens[name] = token
	utils.SetBaseMainConfig(config)

	utils.TouchDatabase()
	go constellation.SendNewDBSyncMessage()

	utils.TriggerEvent(
		"cosmos.api.token.updated",
		"API Token updated",
		"important",
		"token@"+name,
		map[string]interface{}{
			"tokenName": name,
		},
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
	})
}
