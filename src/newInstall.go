package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/azukaar/cosmos-server/src/configapi"
	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
)

func waitForDB() {
	time.Sleep(1 * time.Second)
	err := utils.DB()
	if err != nil {
		utils.Debug("DB error: " + err.Error())
		utils.Warn("DB Not ready yet")
		waitForDB()
	}
}

type NewInstallJSON struct {
	MongoDBMode string `json:"mongodbMode"`
	MongoDB string `json:"mongodb"`
	HTTPSCertificateMode string `json:"httpsCertificateMode"`
	TLSCert string `json:"tlsCert"`
	TLSKey string `json:"tlsKey"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
	Email string `json:"omitempty,email"`
	Hostname string `json:"hostname"`
	Step string `json:"step"`
	SSLEmail string `json:"sslEmail",validate:"omitempty,email"`
	UseWildcardCertificate bool `json:"useWildcardCertificate",validate:"omitempty"`
	DNSChallengeProvider string `json:"dnsChallengeProvider",validate:"omitempty"`
	DNSChallengeConfig map[string]string
	AllowHTTPLocalIPAccess bool `json:"allowHTTPLocalIPAccess",validate:"omitempty"`
}

type AdminJSON struct {
	Nickname string `validate:"required,min=3,max=32,alphanum"`
	Password string `validate:"required,min=9,max=128,containsany=~!@#$%^&*()_+=-{[}]:;\"'<>.?/,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
}

// NewInstallRoute godoc
// @Summary Process new installation steps
// @Description Handles multi-step server installation including database setup, HTTPS configuration, and admin user creation
// @Tags system
// @Accept json
// @Produce json
// @Param request body NewInstallJSON true "Installation step data"
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/newInstall [post]
func NewInstallRoute(w http.ResponseWriter, req *http.Request) {
	if !utils.GetMainConfig().NewInstall {
		utils.Error("Status: not a new New install", nil)
		utils.HTTPError(w, "New install", http.StatusForbidden, "NI001")
		return
	}

	if(req.Method == "POST") {
		var request NewInstallJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("NewInstall: Invalid User Request", err1)
			utils.HTTPError(w, "New Install: Invalid User Request" + err1.Error(),
				http.StatusInternalServerError, "NI001")
			return
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("NewInstall: Invalid User Request", errV)
			utils.HTTPError(w, "New Install: Invalid User Request " + errV.Error(),
				http.StatusInternalServerError, "NI001")
			return
		}

		newConfig := utils.GetBaseMainConfig()

		if(request.Step == "-1") {
			// remove everythin in /config
			utils.Log("NewInstall: Step cleanup")
			utils.Log("NewInstall: Removing config file")
			configFile := utils.GetConfigFileName()
			os.Remove(configFile)
			if(utils.IsInsideContainer) {
				utils.Log("NewInstall: Emptying /config")
				os.RemoveAll("/config")
				os.Mkdir("/config", 0700)
			}
			utils.DisconnectDB()
			LoadConfig()
		}
		if(request.Step == "2") {
			utils.Log("NewInstall: Step Database")
			// User Management & Mongo DB
			if(request.MongoDBMode == "DisableUserManagement") {
				utils.Log("NewInstall: Disable User Management")
				newConfig.DisableUserManagement = true
				utils.SaveConfigTofile(newConfig)
				utils.LoadBaseMainConfig(newConfig)
			} else if (request.MongoDBMode == "Provided") {
				utils.Log("NewInstall: DB Provided")
				newConfig.DisableUserManagement = false
				newConfig.MongoDB = request.MongoDB
				newConfig.Database.PuppetMode = false
				utils.SaveConfigTofile(newConfig)
				utils.LoadBaseMainConfig(newConfig)
			} else if (request.MongoDBMode == "Create") {
				utils.Log("NewInstall: Create DB")
				newConfig.DisableUserManagement = false
				newConfig.MongoDB = ""

				strco, err := docker.NewDB(w, req)
				if err != nil {
					utils.Error("NewInstall: Error creating MongoDB", err)
					return
				}

				newConfig.Database = strco
				utils.SaveConfigTofile(newConfig)
				utils.LoadBaseMainConfig(newConfig)
				utils.Log("NewInstall: MongoDB created, waiting for it to be ready")
				waitForDB()
				w.WriteHeader(http.StatusOK)
				return
			} else {
				utils.Log("NewInstall: Invalid MongoDBMode")
				utils.Error("NewInstall: Invalid MongoDBMode", nil)
				utils.HTTPError(w, "New Install: Invalid MongoDBMode",
					http.StatusInternalServerError, "NI001")
				return
			}
		} else if (request.Step == "3") {
			// HTTPS Certificate Mode & Certs & Let's Encrypt
			newConfig.HTTPConfig.HTTPSCertificateMode = request.HTTPSCertificateMode
			newConfig.HTTPConfig.SSLEmail = request.SSLEmail
			newConfig.HTTPConfig.UseWildcardCertificate = request.UseWildcardCertificate
			newConfig.HTTPConfig.DNSChallengeProvider = request.DNSChallengeProvider
			newConfig.HTTPConfig.DNSChallengeConfig = request.DNSChallengeConfig
			newConfig.HTTPConfig.TLSCert = request.TLSCert
			newConfig.HTTPConfig.TLSKey = request.TLSKey
			newConfig.HTTPConfig.AllowHTTPLocalIPAccess = request.AllowHTTPLocalIPAccess

			// Hostname
			newConfig.HTTPConfig.Hostname = request.Hostname

			utils.SaveConfigTofile(newConfig)
			utils.LoadBaseMainConfig(newConfig)
		} else if (request.Step == "4") {
			adminObj := AdminJSON{
				Nickname: request.Nickname,
				Password: request.Password,
			}

			errV2 := utils.Validate.Struct(adminObj)
			if errV2 != nil {
				utils.Error("NewInstall: Invalid User Request", errV2)
				utils.HTTPError(w, errV2.Error(), http.StatusInternalServerError, "UL001")
				return
			}

			// Admin User
			c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
  defer closeDb()
			if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
			}

			nickname := utils.Sanitize(request.Nickname)
			hashedPassword, err2 := bcrypt.GenerateFromPassword([]byte(request.Password), 14)

			if err2 != nil {
				utils.Error("NewInstall: Error hashing password", err2)
				utils.HTTPError(w, "New Install: Error hashing password " + err2.Error(),
					http.StatusInternalServerError, "NI001")
				return
			}

			// pre-remove every users
			_, err4 := c.DeleteMany(nil, map[string]interface{}{})
			if err4 != nil {
				utils.Error("NewInstall: Error deleting users", err4)
				utils.HTTPError(w, "New Install: Error deleting users " + err4.Error(),
					http.StatusInternalServerError, "NI001")
				return
			}

			_, err3 := c.InsertOne(nil, map[string]interface{}{
				"Nickname": nickname,
				"Email": request.Email,
				"Password": hashedPassword,
				"Role": utils.ADMIN,
				"PasswordCycle": 0,
				"CreatedAt": time.Now(),
				"RegisteredAt": time.Now(),
			})

			if err3 != nil {
				utils.Error("NewInstall: Error creating admin user", err3)
				utils.HTTPError(w, "New Install: Error creating admin user " + err3.Error(),
					http.StatusInternalServerError, "NI001")
				return
			}
		} else if (request.Step == "5") {
			newConfig.NewInstall = false
			utils.SaveConfigTofile(newConfig)
			os.Exit(0)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

// noopResponseWriter discards all output — used to call docker.NewDB without streaming.
type noopResponseWriter struct {
	header http.Header
}

func (n *noopResponseWriter) Header() http.Header        { return n.header }
func (n *noopResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (n *noopResponseWriter) WriteHeader(int)             {}
func (n *noopResponseWriter) Flush()                      {}

type SetupJSON struct {
	// Database
	MongoDBMode string `json:"mongodbMode"`
	MongoDB     string `json:"mongodb"`

	// HTTPS
	Hostname               string            `json:"hostname"`
	HTTPSCertificateMode   string            `json:"httpsCertificateMode"`
	SSLEmail               string            `json:"sslEmail"`
	UseWildcardCertificate bool              `json:"useWildcardCertificate"`
	DNSChallengeProvider   string            `json:"dnsChallengeProvider"`
	DNSChallengeConfig     map[string]string `json:"DNSChallengeConfig"`
	TLSCert                string            `json:"tlsCert"`
	TLSKey                 string            `json:"tlsKey"`
	AllowHTTPLocalIPAccess bool              `json:"allowHTTPLocalIPAccess"`

	// Admin
	Nickname string `json:"nickname"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`

	// Optional: clear config before setup
	ClearConfig bool `json:"clearConfig,omitempty"`

	// Optional: Constellation connect
	ConstellationConfig string `json:"constellationConfig,omitempty"`

	// Optional: create admin API token
	CreateAdminToken bool `json:"createAdminToken,omitempty"`

	// Optional: Pro licence key — stored on the lighthouse config and
	// validated via utils.ProcessLicence at the end of setup.
	Licence string `json:"licence,omitempty"`
}

// SetupRoute godoc
// @Summary One-shot server provisioning
// @Description Performs the entire initial setup in a single call: database, HTTPS, admin user, optional Constellation connect, and optional admin API token creation. Only available when the server is in NewInstall mode.
// @Tags system
// @Accept json
// @Produce json
// @Param request body SetupJSON true "Setup configuration"
// @Success 200 {object} utils.APIResponse
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/setup [post]
func SetupRoute(w http.ResponseWriter, req *http.Request) {
	if !utils.GetMainConfig().NewInstall {
		utils.Error("Setup: not a new install", nil)
		utils.HTTPError(w, "Server is already configured", http.StatusForbidden, "SU001")
		return
	}

	if req.Method != "POST" {
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	var request SetupJSON
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		utils.Error("Setup: Invalid request", err)
		utils.HTTPError(w, "Invalid request: "+err.Error(), http.StatusBadRequest, "SU002")
		return
	}

	// ── Step 0: Clear config (optional) ──────────────────────────────
	if request.ClearConfig {
		utils.Log("Setup: Clearing config")
		configFile := utils.GetConfigFileName()
		os.Remove(configFile)
		if utils.IsInsideContainer {
			os.RemoveAll("/config")
			os.Mkdir("/config", 0700)
		}
		utils.DisconnectDB()
		LoadConfig()
	}

	config := utils.GetBaseMainConfig()

	// ── Step 1: Database ──────────────────────────────────────────────
	if request.MongoDBMode == "" {
		utils.HTTPError(w, "mongodbMode is required", http.StatusBadRequest, "SU003")
		return
	}

	switch request.MongoDBMode {
	case "DisableUserManagement":
		utils.Log("Setup: Disable User Management")
		config.DisableUserManagement = true
	case "Provided":
		utils.Log("Setup: DB Provided")
		config.DisableUserManagement = false
		config.MongoDB = request.MongoDB
		config.Database.PuppetMode = false
	case "Create":
		utils.Log("Setup: Create DB")
		config.DisableUserManagement = false
		config.MongoDB = ""
		noop := &noopResponseWriter{header: make(http.Header)}
		dbConf, err := docker.NewDB(noop, req)
		if err != nil {
			utils.Error("Setup: Error creating MongoDB", err)
			utils.HTTPError(w, "Error creating MongoDB: "+err.Error(), http.StatusInternalServerError, "SU004")
			return
		}
		config.Database = dbConf
	default:
		utils.HTTPError(w, "Invalid mongodbMode", http.StatusBadRequest, "SU003")
		return
	}

	utils.SaveConfigTofile(config)
	utils.LoadBaseMainConfig(config)

	if !config.DisableUserManagement {
		waitForDB()
	}

	// ── Step 1b: Licence (optional) ───────────────────────────────────
	// Persist the licence before constellation join so the joining flow
	// can pick it up (constellation_config may rely on Pro features).
	if request.Licence != "" {
		config.Licence = request.Licence
		utils.SaveConfigTofile(config)
		utils.LoadBaseMainConfig(config)
		utils.ProcessLicence()
	}

	// ── Step 2: HTTPS ─────────────────────────────────────────────────
	config.HTTPConfig.Hostname = request.Hostname
	config.HTTPConfig.HTTPSCertificateMode = request.HTTPSCertificateMode
	config.HTTPConfig.SSLEmail = request.SSLEmail
	config.HTTPConfig.UseWildcardCertificate = request.UseWildcardCertificate
	config.HTTPConfig.DNSChallengeProvider = request.DNSChallengeProvider
	config.HTTPConfig.DNSChallengeConfig = request.DNSChallengeConfig
	config.HTTPConfig.TLSCert = request.TLSCert
	config.HTTPConfig.TLSKey = request.TLSKey
	config.HTTPConfig.AllowHTTPLocalIPAccess = request.AllowHTTPLocalIPAccess

	utils.SaveConfigTofile(config)
	utils.LoadBaseMainConfig(config)

	// ── Step 3: Admin User ────────────────────────────────────────────
	if !config.DisableUserManagement {
		adminObj := AdminJSON{
			Nickname: request.Nickname,
			Password: request.Password,
		}
		if errV := utils.Validate.Struct(adminObj); errV != nil {
			utils.Error("Setup: Invalid admin credentials", errV)
			utils.HTTPError(w, "Invalid admin credentials: "+errV.Error(), http.StatusBadRequest, "SU005")
			return
		}

		c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
		defer closeDb()
		if errCo != nil {
			utils.Error("Setup: Database Connect", errCo)
			utils.HTTPError(w, "Database error", http.StatusInternalServerError, "SU006")
			return
		}

		nickname := utils.Sanitize(request.Nickname)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), 14)
		if err != nil {
			utils.Error("Setup: Error hashing password", err)
			utils.HTTPError(w, "Error hashing password", http.StatusInternalServerError, "SU007")
			return
		}

		c.DeleteMany(nil, map[string]interface{}{})
		_, err = c.InsertOne(nil, map[string]interface{}{
			"Nickname":      nickname,
			"Email":         request.Email,
			"Password":      hashedPassword,
			"Role":          utils.ADMIN,
			"PasswordCycle": 0,
			"CreatedAt":     time.Now(),
			"RegisteredAt":  time.Now(),
		})
		if err != nil {
			utils.Error("Setup: Error creating admin user", err)
			utils.HTTPError(w, "Error creating admin user: "+err.Error(), http.StatusInternalServerError, "SU008")
			return
		}
	}

	// ── Step 4: Constellation (optional) ──────────────────────────────
	if request.ConstellationConfig != "" {
		var err error
		config, err = constellation.ConnectToExisting([]byte(request.ConstellationConfig), config)
		if err != nil {
			utils.Error("Setup: Constellation connect error", err)
			utils.HTTPError(w, "Constellation error: "+err.Error(), http.StatusInternalServerError, "SU009")
			return
		}
		utils.SaveConfigTofile(config)
		utils.LoadBaseMainConfig(config)
	}

	// ── Step 5: Admin Token (optional) ────────────────────────────────
	responseData := map[string]interface{}{}

	if request.CreateAdminToken {
		tokenName := "provisioning-token"
		owner := utils.Sanitize(request.Nickname)
		rawToken, tokenConfig, err := configapi.GenerateAPIToken(tokenName, "Auto-generated during setup", owner, utils.DefaultAdminTokenPermissions)
		if err != nil {
			utils.Error("Setup: Error generating token", err)
			utils.HTTPError(w, "Error generating token", http.StatusInternalServerError, "SU010")
			return
		}

		if config.APITokens == nil {
			config.APITokens = make(map[string]utils.APITokenConfig)
		}
		config.APITokens[tokenName] = tokenConfig

		responseData["adminToken"] = rawToken
		responseData["adminTokenName"] = tokenName
	}

	// ── Finalize ──────────────────────────────────────────────────────
	config.NewInstall = false
	utils.SaveConfigTofile(config)
	utils.LoadBaseMainConfig(config)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   responseData,
	})

	// Flush so the client receives the response (including adminToken)
	// before we exit, then restart so the freshly-saved config is reloaded
	// from scratch by the launcher / supervisor.
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}
