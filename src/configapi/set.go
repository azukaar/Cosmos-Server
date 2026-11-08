package configapi

import (
	"net/http"
	"encoding/json"
	"reflect"
	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/storage"
	"github.com/azukaar/cosmos-server/src/utils"
)

// publishReplicatedDomains publishes only the domains this PUT changed, so a
// node without a writable op-log can still save node-local settings.
func publishReplicatedDomains(request utils.Config, current utils.Config) error {
	newDNS := dnsPayloadOf(request)
	if !reflect.DeepEqual(newDNS, dnsPayloadOf(current)) {
		if err := constellation.PublishDomainOp(constellation.DomainDNS, newDNS); err != nil {
			return err
		}
	}
	if !sameRoles(request.Roles, current.Roles) {
		if err := constellation.PublishDomainOp(constellation.DomainRoles, request.Roles); err != nil {
			return err
		}
	}
	if !sameOpenIDClients(request.OpenIDClients, current.OpenIDClients) {
		if err := constellation.PublishDomainOp(constellation.DomainOpenIDClients, request.OpenIDClients); err != nil {
			return err
		}
	}
	// caller restored the masked credentials; publishing earlier would replicate them empty
	if !reflect.DeepEqual(zonesPayloadOf(request), zonesPayloadOf(current)) {
		if err := constellation.PublishDomainOp(constellation.DomainDNSZones, zonesPayloadOf(request)); err != nil {
			return err
		}
	}
	if request.HTTPConfig.SSLEmail != current.HTTPConfig.SSLEmail {
		if err := constellation.PublishDomainOp(constellation.DomainHTTPSSettings, constellation.HTTPSSettingsPayload{SSLEmail: request.HTTPConfig.SSLEmail}); err != nil {
			return err
		}
	}
	// caller restored the real password; publishing earlier would replicate "***"
	newDB := dbPayloadOf(request)
	if !reflect.DeepEqual(newDB, dbPayloadOf(current)) {
		return constellation.PublishDomainOp(constellation.DomainDatabase, newDB)
	}
	return nil
}

// slices are normalized to nil so a UI []-for-null round-trip never looks like a change
func dnsPayloadOf(c utils.Config) constellation.DNSPayload {
	cc := c.ConstellationConfig
	blocklists := cc.DNSAdditionalBlocklists
	if len(blocklists) == 0 {
		blocklists = nil
	}
	entries := cc.CustomDNSEntries
	if len(entries) == 0 {
		entries = nil
	}
	return constellation.DNSPayload{
		DNSPort:                 cc.DNSPort,
		DNSFallback:             cc.DNSFallback,
		DNSBlockBlacklist:       cc.DNSBlockBlacklist,
		DNSAdditionalBlocklists: blocklists,
		CustomDNSEntries:        entries,
	}
}

func zonesPayloadOf(c utils.Config) []utils.DNSZoneConfig {
	if len(c.HTTPConfig.DNSZones) == 0 {
		return nil
	}
	return c.HTTPConfig.DNSZones
}

func dbPayloadOf(c utils.Config) constellation.DatabasePayload {
	return constellation.DatabasePayload{
		PostgresHost:     c.Database.PostgresHost,
		PostgresDatabase: c.Database.PostgresDatabase,
		PostgresUsername: c.Database.PostgresUsername,
		PostgresPassword: c.Database.PostgresPassword,
	}
}

// nil and empty are the same thing here: JSON round-trips blur them
func sameRoles(a, b map[utils.Role]utils.RoleConfig) bool {
	if len(a) != len(b) {
		return false
	}
	for k, ra := range a {
		rb, ok := b[k]
		if !ok || ra.Name != rb.Name || !samePermissions(ra.Permissions, rb.Permissions) {
			return false
		}
	}
	return true
}

func samePermissions(a, b []utils.Permission) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameOpenIDClients(a, b []utils.OpenIDClient) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// restoreReplicatedDomains takes the replicated fields from the freshly applied
// config rather than from the request, so the local write can never roll back
// what the apply loop just installed. Same treatment the auth keypair and API
// tokens already get. Every OTHER ConstellationConfig field (Enabled, IPRange,
// ThisDeviceName, DNSDisabled...) stays node-local from the request by design.
func restoreReplicatedDomains(request *utils.Config, config utils.Config) {
	request.ConstellationConfig.DNSPort = config.ConstellationConfig.DNSPort
	request.ConstellationConfig.DNSFallback = config.ConstellationConfig.DNSFallback
	request.ConstellationConfig.DNSBlockBlacklist = config.ConstellationConfig.DNSBlockBlacklist
	request.ConstellationConfig.DNSAdditionalBlocklists = config.ConstellationConfig.DNSAdditionalBlocklists
	request.ConstellationConfig.CustomDNSEntries = config.ConstellationConfig.CustomDNSEntries
	request.Roles = config.Roles
	request.OpenIDClients = config.OpenIDClients
	request.HTTPConfig.SSLEmail = config.HTTPConfig.SSLEmail
	request.HTTPConfig.SSLEmailShared = config.HTTPConfig.SSLEmailShared
	// Database.NodeName stays node-local from the request, like the ConstellationConfig fields above
	request.Database.PostgresHost = config.Database.PostgresHost
	request.Database.PostgresDatabase = config.Database.PostgresDatabase
	request.Database.PostgresUsername = config.Database.PostgresUsername
	request.Database.PostgresPassword = config.Database.PostgresPassword
}

// ConfigApiSet godoc
// @Summary Update server configuration
// @Description Replaces the entire server configuration (masked credential fields are preserved from existing config)
// @Tags config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body utils.Config true "Full configuration object"
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 405 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/config [put]
func ConfigApiSet(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_CONFIGURATION) != nil {
		return
	} 

	if(req.Method == "PUT") {
		var request utils.Config
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("SettingsUpdate: Invalid User Request", err1)
			utils.HTTPError(w, "User Creation Error", 
				http.StatusInternalServerError, "UC001")
			return 
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("SettingsUpdate: Invalid User Request", errV)
			utils.HTTPError(w, "User Creation Error: " + errV.Error(),
				http.StatusInternalServerError, "UC003")
			return 
		}

		// this PUT writes routes wholesale: validate before any write, including the op-log publish
		for _, route := range request.HTTPConfig.ProxyConfig.Routes {
			if msg := validateRoute(route); msg != "" {
				utils.Error("SettingsUpdate: "+msg, nil)
				utils.HTTPError(w, msg, http.StatusBadRequest, "UC005")
				return
			}
		}

		// an unchanged fallback is left alone so that an older config still saves
		if request.ConstellationConfig.DNSFallback != utils.GetMainConfig().ConstellationConfig.DNSFallback {
			if err := utils.ValidateDNSFallback(request.ConstellationConfig.DNSFallback); err != nil {
				utils.Error("SettingsUpdate: Invalid DNS fallback", err)
				utils.HTTPError(w, err.Error(), http.StatusBadRequest, "UC007")
				return
			}
		}

		// restore fields that are never sent to the client or are masked with ***
		config := utils.ReadConfigFromFile()
		request.HTTPConfig.AuthPrivateKey = config.HTTPConfig.AuthPrivateKey
		request.HTTPConfig.AuthPublicKey = config.HTTPConfig.AuthPublicKey
		request.HTTPConfig.OIDCSigningKey = config.HTTPConfig.OIDCSigningKey
		request.HTTPConfig.OIDCSigningKeyShared = config.HTTPConfig.OIDCSigningKeyShared
		request.HTTPConfig.TLSCert = config.HTTPConfig.TLSCert
		request.HTTPConfig.TLSKey = config.HTTPConfig.TLSKey
		// zone certificates are never sent to the client; the ones of deleted zones go with them
		request.HTTPConfig.ZoneCerts = config.HTTPConfig.ZoneCerts
		for name := range request.HTTPConfig.ZoneCerts {
			if i := utils.FindZone(request.HTTPConfig.DNSZones, name); i == -1 || request.HTTPConfig.DNSZones[i].Zone != name {
				delete(request.HTTPConfig.ZoneCerts, name)
			}
		}
		// the server's own certificates are never sent to the client either
		request.HTTPConfig.LocalCerts = config.HTTPConfig.LocalCerts
		request.NewInstall = config.NewInstall

		// restore API token as we cannot edit it here
		request.APITokens = config.APITokens

		if !sameRoles(request.Roles, config.Roles) {
			if err := utils.ValidateRolesChange(req, utils.GetRoles(), request.Roles); err != nil {
				utils.Error("SettingsUpdate: "+err.Error(), nil)
				utils.HTTPError(w, err.Error(), http.StatusForbidden, "UC006")
				return
			}
		}

		// restore DNS if user cannot read credentials, as they are sent as empty and we don't want to override them
		canReadCredentials := utils.HasPermission(req, utils.PERM_CREDENTIALS_READ)
		if !canReadCredentials {
			request.HTTPConfig.DNSChallengeConfig = config.HTTPConfig.DNSChallengeConfig
			for i, zone := range request.HTTPConfig.DNSZones {
				if j := utils.FindZone(config.HTTPConfig.DNSZones, zone.Zone); j != -1 && config.HTTPConfig.DNSZones[j].Zone == zone.Zone {
					request.HTTPConfig.DNSZones[i].DNSChallengeConfig = config.HTTPConfig.DNSZones[j].DNSChallengeConfig
				}
			}
		}

		// the private key of a provided certificate is never sent to the client either
		for i, zone := range request.HTTPConfig.DNSZones {
			if zone.TLSKey != "" {
				continue
			}
			if j := utils.FindZone(config.HTTPConfig.DNSZones, zone.Zone); j != -1 && config.HTTPConfig.DNSZones[j].Zone == zone.Zone {
				request.HTTPConfig.DNSZones[i].TLSKey = config.HTTPConfig.DNSZones[j].TLSKey
			}
		}

		// restore credential fields if they were masked (sent as "***")
		if request.EmailConfig.Password == "***" {
			request.EmailConfig.Password = config.EmailConfig.Password
		}
		if request.EmailConfig.Username == "***" {
			request.EmailConfig.Username = config.EmailConfig.Username
		}
		if request.EmailConfig.Host == "***" {
			request.EmailConfig.Host = config.EmailConfig.Host
		}
		if request.Database.PostgresPassword == "***" {
			request.Database.PostgresPassword = config.Database.PostgresPassword
		}
		if request.Licence == "***" {
			request.Licence = config.Licence
		}
		if request.ServerToken == "***" {
			request.ServerToken = config.ServerToken
		}

		// restore backup passwords. Users without PERM_CREDENTIALS_READ can never
		// set a password — always restore from existing config. Otherwise only
		// restore when the value is missing or masked, so the settings UI
		// round-trip doesn't wipe it.
		for name, b := range request.Backup.Backups {
			shouldRestore := !canReadCredentials || b.Password == "" || b.Password == "***"
			if shouldRestore {
				if existing, ok := config.Backup.Backups[name]; ok && existing.Password != "" {
					b.Password = existing.Password
					request.Backup.Backups[name] = b
				}
			}
		}

		// Replicated domains go through the op-log, never straight to disk, and are
		// published BEFORE the local write so a read-only node fails with nothing
		// written. ConfigLock must NOT be held — the apply loop takes it.
		if err := publishReplicatedDomains(request, config); err != nil {
			utils.HTTPStoreError(w, err, "UC004")
			return
		}

		// The local write is the node-local half. Under the lock so it cannot
		// interleave with an apply-loop domain install, and re-reading here picks up
		// the state that loop just applied.
		utils.ConfigLock.Lock()
		config = utils.ReadConfigFromFile()
		restoreReplicatedDomains(&request, config)
		// Managed routes are owned by their materializer: keep the stored copies whatever the payload says.
		request.HTTPConfig.ProxyConfig.Routes = utils.PreserveManagedRoutes(
			request.HTTPConfig.ProxyConfig.Routes, config.HTTPConfig.ProxyConfig.Routes)
		utils.SetBaseMainConfig(request)
		utils.ConfigLock.Unlock()

		// Storage shares carry their own route; derive the ProxyConfig.Routes copies server-side.
		storage.ReconcileShareRoutes()
		go constellation.ReapplyHostDNS()

		utils.TriggerEvent(
			"cosmos.settings",
			"Settings updated",
			"success",
			"",
			map[string]interface{}{
		})

		go utils.SoftRestartServer()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("SettingsUpdate: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}
