package constellation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/azukaar/cosmos-server/src/utils"
)

// Everything the op-log replicates that isn't a SQLite row lives here, one entry
// per domain. Adding a future domain is one registry entry — that is the point
// of the indirection.
//
// Apply runs with utils.ConfigLock already held by the caller; React runs after
// the sequence is committed, and must not block the apply loop.
type Domain struct {
	Name     string
	Apply    func(state json.RawMessage) error
	React    func(old json.RawMessage, new json.RawMessage)
	Snapshot func() (json.RawMessage, error)
}

const (
	DomainAuthKeys      = "auth_keys"
	DomainDNS           = "dns"
	DomainAPITokens     = "api_tokens"
	DomainRoles         = "roles"
	DomainOpenIDClients = "openid_clients"
	DomainFileCACrt     = "file:ca.crt"
	DomainFileCAKey     = "file:ca.key"
	DomainFileRclone    = "file:rclone.conf"
	DomainDatabase      = "database"
	DomainDNSZones      = "dns_zones"
	DomainZoneCerts     = "zone_certs"
	DomainHTTPSSettings = "https_settings"
	DomainOIDCSigningKey = "oidc_signing_key"
)

// HTTPSSettingsPayload is what the servers of a cluster share about HTTPS
// besides the zones: the contact their certificates are requested with.
// Everything else about HTTPS belongs to a zone, or to one server (HTTP only).
type HTTPSSettingsPayload struct {
	SSLEmail string `json:"sslEmail"`
}

type AuthKeysPayload struct {
	AuthPrivateKey string `json:"authPrivateKey"`
	AuthPublicKey  string `json:"authPublicKey"`
}

type OIDCSigningKeyPayload struct {
	Key string `json:"key"`
}

type DNSPayload struct {
	DNSPort                 string                        `json:"dnsPort"`
	DNSFallback             string                        `json:"dnsFallback"`
	DNSBlockBlacklist       bool                          `json:"dnsBlockBlacklist"`
	DNSAdditionalBlocklists []string                      `json:"dnsAdditionalBlocklists"`
	CustomDNSEntries        []utils.ConstellationDNSEntry `json:"customDNSEntries"`
}

// DatabasePayload carries the Postgres connection only. NodeName stays
// node-local: it is this node's metrics identity.
type DatabasePayload struct {
	PostgresHost     string `json:"postgresHost"`
	PostgresDatabase string `json:"postgresDatabase"`
	PostgresUsername string `json:"postgresUsername"`
	PostgresPassword string `json:"postgresPassword"`
}

// FilePayload carries a loose config file as base64. Empty Data means "no such
// file here" and is never applied, so a node that lacks a CA can't wipe everyone's.
type FilePayload struct {
	Data string `json:"data"`
}

var oplogDomains = map[string]Domain{}

func init() {
	register := func(d Domain) { oplogDomains[d.Name] = d }

	register(Domain{
		Name: DomainAuthKeys,
		Apply: func(state json.RawMessage) error {
			var p AuthKeysPayload
			if err := json.Unmarshal(state, &p); err != nil {
				return err
			}
			config := utils.ReadConfigFromFile()
			config.HTTPConfig.AuthPrivateKey = p.AuthPrivateKey
			config.HTTPConfig.AuthPublicKey = p.AuthPublicKey
			utils.SetBaseMainConfig(config)
			return nil
		},
		// rotating the JWT keypair invalidates live sessions, so only bounce
		// the server when the keys actually moved
		React: func(old json.RawMessage, new json.RawMessage) {
			if !bytes.Equal(old, new) {
				go utils.RestartHTTPServer()
			}
		},
		Snapshot: func() (json.RawMessage, error) {
			config := utils.GetMainConfig()
			return json.Marshal(AuthKeysPayload{
				AuthPrivateKey: config.HTTPConfig.AuthPrivateKey,
				AuthPublicKey:  config.HTTPConfig.AuthPublicKey,
			})
		},
	})

	// an empty state is never applied: a node without a key can't wipe everyone's
	register(Domain{
		Name: DomainOIDCSigningKey,
		Apply: func(state json.RawMessage) error {
			var p OIDCSigningKeyPayload
			if err := json.Unmarshal(state, &p); err != nil {
				return err
			}
			if p.Key == "" {
				return nil
			}
			config := utils.ReadConfigFromFile()
			config.HTTPConfig.OIDCSigningKey = p.Key
			config.HTTPConfig.OIDCSigningKeyShared = true
			utils.SetBaseMainConfig(config)
			return nil
		},
		// the key is loaded when the OpenID server starts
		React: func(old json.RawMessage, new json.RawMessage) {
			if !bytes.Equal(old, new) {
				go utils.RestartHTTPServer()
			}
		},
		Snapshot: func() (json.RawMessage, error) {
			return json.Marshal(OIDCSigningKeyPayload{Key: utils.GetMainConfig().HTTPConfig.OIDCSigningKey})
		},
	})

	register(Domain{
		Name: DomainDNS,
		// Assigns the five replicated fields individually, and must keep doing so:
		// ConstellationConfig also holds per-node state (Enabled, IPRange,
		// ThisDeviceName, ConstellationHostname, NATSReplicas, DNSDisabled,
		// FirewallBlockedClients) that api_nebula_connect.go writes locally by design.
		// Replacing the struct wholesale here would push one node's identity to every
		// other node.
		Apply: func(state json.RawMessage) error {
			var p DNSPayload
			if err := json.Unmarshal(state, &p); err != nil {
				return err
			}
			config := utils.ReadConfigFromFile()
			config.ConstellationConfig.DNSPort = p.DNSPort
			config.ConstellationConfig.DNSFallback = p.DNSFallback
			config.ConstellationConfig.DNSBlockBlacklist = p.DNSBlockBlacklist
			config.ConstellationConfig.DNSAdditionalBlocklists = p.DNSAdditionalBlocklists
			config.ConstellationConfig.CustomDNSEntries = p.CustomDNSEntries
			utils.SetBaseMainConfig(config)
			return nil
		},
		React: func(old json.RawMessage, new json.RawMessage) {
			if !bytes.Equal(old, new) {
				go ReloadDNS()
			}
		},
		Snapshot: func() (json.RawMessage, error) {
			c := utils.GetMainConfig().ConstellationConfig
			return json.Marshal(DNSPayload{
				DNSPort:                 c.DNSPort,
				DNSFallback:             c.DNSFallback,
				DNSBlockBlacklist:       c.DNSBlockBlacklist,
				DNSAdditionalBlocklists: c.DNSAdditionalBlocklists,
				CustomDNSEntries:        c.CustomDNSEntries,
			})
		},
	})

	register(Domain{
		Name: DomainDNSZones,
		Apply: func(state json.RawMessage) error {
			var zones []utils.DNSZoneConfig
			if err := json.Unmarshal(state, &zones); err != nil {
				return err
			}
			config := utils.ReadConfigFromFile()
			config.HTTPConfig.DNSZones = zones
			// the certificate of a deleted zone goes with it, on every node
			for name := range config.HTTPConfig.ZoneCerts {
				if i := utils.FindZone(zones, name); i == -1 || zones[i].Zone != name {
					delete(config.HTTPConfig.ZoneCerts, name)
				}
			}
			utils.SetBaseMainConfig(config)
			return nil
		},
		// a zone decides which certificate every hostname is served with and who
		// issues it, so the server re-evaluates all of it
		React: func(old json.RawMessage, new json.RawMessage) {
			if !bytes.Equal(old, new) {
				go utils.RestartHTTPServer()
			}
		},
		Snapshot: func() (json.RawMessage, error) {
			return json.Marshal(utils.GetMainConfig().HTTPConfig.DNSZones)
		},
	})

	// read when a certificate is requested, so there is nothing to react to
	register(Domain{
		Name: DomainHTTPSSettings,
		Apply: func(state json.RawMessage) error {
			var p HTTPSSettingsPayload
			if err := json.Unmarshal(state, &p); err != nil {
				return err
			}
			config := utils.ReadConfigFromFile()
			config.HTTPConfig.SSLEmail = p.SSLEmail
			config.HTTPConfig.SSLEmailShared = true
			utils.SetBaseMainConfig(config)
			return nil
		},
		Snapshot: func() (json.RawMessage, error) {
			return json.Marshal(HTTPSSettingsPayload{SSLEmail: utils.GetMainConfig().HTTPConfig.SSLEmail})
		},
	})

	// written by the zone issuer only. An empty state is never applied: a node
	// that has no certificate yet can't wipe everyone's.
	register(Domain{
		Name: DomainZoneCerts,
		Apply: func(state json.RawMessage) error {
			var certs map[string]utils.ZoneCert
			if err := json.Unmarshal(state, &certs); err != nil {
				return err
			}
			if len(certs) == 0 {
				return nil
			}
			config := utils.ReadConfigFromFile()
			config.HTTPConfig.ZoneCerts = certs
			utils.SetBaseMainConfig(config)
			return nil
		},
		// served straight from the store, no restart
		React: func(old json.RawMessage, new json.RawMessage) {
			if !bytes.Equal(old, new) {
				go utils.ReloadZoneCerts()
			}
		},
		Snapshot: func() (json.RawMessage, error) {
			return json.Marshal(utils.GetMainConfig().HTTPConfig.ZoneCerts)
		},
	})

	// api_tokens / roles / openid_clients need no reaction: every one of them is
	// read per request straight off the config, never cached in a running server
	register(Domain{
		Name: DomainAPITokens,
		Apply: func(state json.RawMessage) error {
			var tokens map[string]utils.APITokenConfig
			if err := json.Unmarshal(state, &tokens); err != nil {
				return err
			}
			config := utils.ReadConfigFromFile()
			config.APITokens = tokens
			utils.SetBaseMainConfig(config)
			return nil
		},
		Snapshot: func() (json.RawMessage, error) {
			return json.Marshal(utils.GetMainConfig().APITokens)
		},
	})

	register(Domain{
		Name: DomainRoles,
		Apply: func(state json.RawMessage) error {
			var roles map[utils.Role]utils.RoleConfig
			if err := json.Unmarshal(state, &roles); err != nil {
				return err
			}
			config := utils.ReadConfigFromFile()
			config.Roles = roles
			utils.SetBaseMainConfig(config)
			return nil
		},
		Snapshot: func() (json.RawMessage, error) {
			return json.Marshal(utils.GetMainConfig().Roles)
		},
	})

	register(Domain{
		Name: DomainOpenIDClients,
		Apply: func(state json.RawMessage) error {
			var clients []utils.OpenIDClient
			if err := json.Unmarshal(state, &clients); err != nil {
				return err
			}
			config := utils.ReadConfigFromFile()
			config.OpenIDClients = clients
			utils.SetBaseMainConfig(config)
			return nil
		},
		Snapshot: func() (json.RawMessage, error) {
			return json.Marshal(utils.GetMainConfig().OpenIDClients)
		},
	})

	register(Domain{
		Name: DomainDatabase,
		// assign fields individually: DatabaseConfig.NodeName stays node-local
		Apply: func(state json.RawMessage) error {
			var p DatabasePayload
			if err := json.Unmarshal(state, &p); err != nil {
				return err
			}
			config := utils.ReadConfigFromFile()
			config.Database.PostgresHost = p.PostgresHost
			config.Database.PostgresDatabase = p.PostgresDatabase
			config.Database.PostgresUsername = p.PostgresUsername
			config.Database.PostgresPassword = p.PostgresPassword
			utils.SetBaseMainConfig(config)
			return nil
		},
		// reconnect only when the connection actually moved
		React: func(old json.RawMessage, new json.RawMessage) {
			if bytes.Equal(old, new) {
				return
			}
			go func() {
				if err := utils.InitMetricsDatabase(); err != nil {
					utils.Error("[OPLOG] monitoring database reconnect failed", err)
				}
			}()
		},
		Snapshot: func() (json.RawMessage, error) {
			d := utils.GetMainConfig().Database
			return json.Marshal(DatabasePayload{
				PostgresHost:     d.PostgresHost,
				PostgresDatabase: d.PostgresDatabase,
				PostgresUsername: d.PostgresUsername,
				PostgresPassword: d.PostgresPassword,
			})
		},
	})

	register(fileDomain(DomainFileCACrt, "ca.crt", nil))
	register(fileDomain(DomainFileCAKey, "ca.key", nil))
	register(fileDomain(DomainFileRclone, "rclone.conf", func(old json.RawMessage, new json.RawMessage) {
		if bytes.Equal(old, new) || utils.InitRemoteStorage == nil {
			return
		}
		go utils.InitRemoteStorage()
	}))
}

func fileDomain(name string, filename string, react func(old json.RawMessage, new json.RawMessage)) Domain {
	path := func() string { return utils.CONFIGFOLDER + filename }
	return Domain{
		Name:  name,
		React: react,
		Apply: func(state json.RawMessage) error {
			var p FilePayload
			if err := json.Unmarshal(state, &p); err != nil {
				return err
			}
			if p.Data == "" {
				return nil
			}
			data, err := base64.StdEncoding.DecodeString(p.Data)
			if err != nil {
				return err
			}
			return writeFileAtomic(path(), data, 0600)
		},
		Snapshot: func() (json.RawMessage, error) {
			data, err := ioutil.ReadFile(path())
			if err != nil {
				return json.Marshal(FilePayload{})
			}
			return json.Marshal(FilePayload{Data: base64.StdEncoding.EncodeToString(data)})
		},
	}
}

// PublishFileDomain publishes a loose config file's current contents as the
// full state of its domain.
func PublishFileDomain(domain string, filename string) error {
	data, err := ioutil.ReadFile(utils.CONFIGFOLDER + filename)
	if err != nil {
		return err
	}
	return PublishDomainOp(domain, FilePayload{Data: base64.StdEncoding.EncodeToString(data)})
}

// writeFileAtomic keeps a crash from leaving a half-written cert on disk.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := ioutil.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// applyDomainLocal installs a domain's state under ConfigLock and fires its
// reaction. Shared by the apply loop and the direct write path.
func applyDomainLocal(name string, state json.RawMessage) error {
	d, ok := oplogDomains[name]
	if !ok {
		return errors.New("oplog: unknown domain " + name)
	}

	old, _ := d.Snapshot()

	utils.ConfigLock.Lock()
	err := d.Apply(state)
	utils.ConfigLock.Unlock()

	if err != nil {
		return err
	}
	if d.React != nil {
		d.React(old, state)
	}
	return nil
}

// snapshotDomains captures every domain's current state for a fast-forward payload.
func snapshotDomains() map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	for name, d := range oplogDomains {
		state, err := d.Snapshot()
		if err != nil {
			utils.Warn("[OPLOG] snapshot of domain " + name + " failed: " + err.Error())
			continue
		}
		out[name] = state
	}
	return out
}
