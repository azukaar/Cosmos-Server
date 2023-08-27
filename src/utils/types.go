package utils

import (
	"os"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role int
type ProxyMode string
type LoggingLevel string

const (
	GUEST 			 = 0
	USER         = 1
	ADMIN        = 2
)

const (
	DEBUG = 0
	INFO = 1
	WARNING = 2
	ERROR = 3
)

var LoggingLevelLabels = map[LoggingLevel]int{
	"DEBUG": DEBUG,
	"INFO": INFO,
	"WARNING": WARNING,
	"ERROR": ERROR,
}

var ProxyModeList = map[string]string{
	"PROXY": "PROXY",
	"SPA": "SPA",
	"STATIC": "STATIC",
	"SERVAPP": "SERVAPP",
	"REDIRECT": "REDIRECT",
}

var HTTPSCertModeList = map[string]string{
	"DISABLED": "DISABLED",
	"PROVIDED": "PROVIDED",
	"SELFSIGNED": "SELFSIGNED",
	"LETSENCRYPT": "LETSENCRYPT",
}

type FileStats struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64 `json:"size"`
	Mode os.FileMode `json:"mode"`
	ModTime time.Time `json:"modTime"`
	IsDir bool `json:"isDir"`
}

type User struct {
	ID       primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Nickname      string     `validate:"required" json:"nickname"`
	Password       string    `validate:"required" json:"-"`
	RegisterKey       string  `json:"registerKey"`
	RegisterKeyExp       time.Time  `json:"registerKeyExp"`
	Role       Role    `validate:"required" json:"role"`
	PasswordCycle			 int    `json:"-"`
	Link 		 string    `json:"link"`
	Email string `validate:"email" json:"email"`
	RegisteredAt time.Time   `json:"registeredAt"`
	LastPasswordChangedAt time.Time   `json:"lastPasswordChangedAt"`
	CreatedAt time.Time   `json:"createdAt"`
	LastLogin time.Time   `json:"lastLogin"`
	MFAKey string `json:"-"`
	Was2FAVerified bool `json:"-"`
	MFAState int `json:"-"` // 0 = done, 1 = needed, 2 = not set
}

type Config struct {
	LoggingLevel LoggingLevel `required,validate:"oneof=DEBUG INFO WARNING ERROR"`
	MongoDB string
	DisableUserManagement bool
	NewInstall bool `validate:"boolean"`
	HTTPConfig HTTPConfig `validate:"required,dive,required"`
	EmailConfig EmailConfig `validate:"required,dive,required"`
	DockerConfig DockerConfig
	BlockedCountries []string
	CountryBlacklistIsWhitelist bool
	ServerCountry string
	RequireMFA bool
	AutoUpdate bool
	OpenIDClients []OpenIDClient
	MarketConfig MarketConfig
	HomepageConfig HomepageConfig
	ThemeConfig ThemeConfig
	ConstellationConfig ConstellationConfig
}

type HomepageConfig struct {
	Background string
	Widgets []string
	Expanded bool
}

type ThemeConfig struct {
	PrimaryColor string
	SecondaryColor string
}

type HTTPConfig struct {
	TLSCert string `validate:"omitempty,contains=\n`
	TLSKey string
	TLSKeyHostsCached []string
	TLSValidUntil time.Time
	AuthPrivateKey string
	AuthPublicKey string
	GenerateMissingAuthCert bool
	HTTPSCertificateMode string
	DNSChallengeProvider string
	ForceHTTPSCertificateRenewal bool
	HTTPPort string `validate:"required,containsany=0123456789,min=1,max=6"`
	HTTPSPort string `validate:"required,containsany=0123456789,min=1,max=6"`
	ProxyConfig ProxyConfig
	Hostname string `validate:"required,excludesall=0x2C/ "`
	SSLEmail string `validate:"omitempty,email"`
	UseWildcardCertificate bool
	OverrideWildcardDomains string `validate:"omitempty,excludesall=/ "`
	AcceptAllInsecureHostname bool
	DNSChallengeConfig map[string]string `json:"DNSChallengeConfig,omitempty"`
} 

const (
	STRICT = 1
	NORMAL = 2
	LENIENT = 3
)
type SmartShieldPolicy struct {
	Enabled bool
	PolicyStrictness int
	PerUserTimeBudget float64
	PerUserRequestLimit int
	PerUserByteLimit int64
	PerUserSimultaneous int
	MaxGlobalSimultaneous int
	PrivilegedGroups int
}

type DockerConfig struct {
	SkipPruneNetwork bool
	DefaultDataPath string
}

type ProxyConfig struct {
	Routes []ProxyRouteConfig
}

type AddionalFiltersConfig struct {
	Type string
	Name string
	Value string
}

type ProxyRouteConfig struct {
	Name string `validate:"required"`
	Description string
	UseHost bool
  Host string
	UsePathPrefix bool
	PathPrefix string
	Timeout time.Duration
	ThrottlePerMinute int
	CORSOrigin string
	StripPathPrefix bool
	MaxBandwith int64
	AuthEnabled bool
	AdminOnly bool
	Target  string `validate:"required"`
	SmartShield SmartShieldPolicy
	Mode ProxyMode
	BlockCommonBots bool
	BlockAPIAbuse bool
	AcceptInsecureHTTPSTarget bool
	HideFromDashboard bool
	DisableHeaderHardening bool
	VerboseForwardHeader bool
	AddionalFilters []AddionalFiltersConfig
}

type EmailConfig struct {
	Enabled		 bool
	Host       string
	Port       string
	Username   string
	Password   string
	From       string
	UseTLS		 bool
}

type OpenIDClient struct {
	ID       string `json:"id"`
	Secret 	 string `json:"secret"`
	Redirect string `json:"redirect"`
}

type MarketConfig struct {
	Sources []MarketSource
}

type MarketSource struct {
	Name string
	Url string
}

type ConstellationConfig struct {
	Enabled bool
	DNS bool
	DNSPort string
	DNSFallback string
	DNSBlockBlacklist bool
	CustomDNSEntries map[string]string
	NebulaConfig NebulaConfig
}

type NebulaFirewallRule struct {
	Port   string   `yaml:"port"`
	Proto  string   `yaml:"proto"`
	Host   string   `yaml:"host"`
	Groups []string `yaml:"groups,omitempty"omitempty"`
}

type NebulaConntrackConfig struct {
	TCPTimeout     string `yaml:"tcp_timeout"`
	UDPTimeout     string `yaml:"udp_timeout"`
	DefaultTimeout string `yaml:"default_timeout"`
}

type NebulaConfig struct {
	PKI struct {
		CA   string `yaml:"ca"`
		Cert string `yaml:"cert"`
		Key  string `yaml:"key"`
	} `yaml:"pki"`

	StaticHostMap map[string][]string `yaml:"static_host_map"`

	Lighthouse struct {
		AMLighthouse bool     `yaml:"am_lighthouse"`
		Interval     int      `yaml:"interval"`
		Hosts        []string `yaml:"hosts"`
	} `yaml:"lighthouse"`

	Listen struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"listen"`

	Punchy struct {
		Punch bool `yaml:"punch"`
		Respond bool `yaml:"respond"`
	} `yaml:"punchy"`

	Relay struct {
		AMRelay   bool `yaml:"am_relay"`
		UseRelays bool `yaml:"use_relays"`
		Relays		[]string `yaml:"relays"`
	} `yaml:"relay"`

	TUN struct {
		Disabled            bool     `yaml:"disabled"`
		Dev                 string   `yaml:"dev"`
		DropLocalBroadcast bool     `yaml:"drop_local_broadcast"`
		DropMulticast       bool     `yaml:"drop_multicast"`
		TxQueue             int      `yaml:"tx_queue"`
		MTU                 int      `yaml:"mtu"`
		Routes              []string `yaml:"routes"`
		UnsafeRoutes        []string `yaml:"unsafe_routes"`
	} `yaml:"tun"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`

	Firewall struct {
		OutboundAction string                    `yaml:"outbound_action"`
		InboundAction  string                    `yaml:"inbound_action"`
		Conntrack      NebulaConntrackConfig `yaml:"conntrack"`
		Outbound       []NebulaFirewallRule  `yaml:"outbound"`
		Inbound        []NebulaFirewallRule  `yaml:"inbound"`
	} `yaml:"firewall"`
}

type Device struct {
	DeviceName string `json:"deviceName",validate:"required,min=3,max=32,alphanum"`
	Nickname string `json:"nickname",validate:"required,min=3,max=32,alphanum"`
	PublicKey string `json:"publicKey",omitempty`
	PrivateKey string `json:"privateKey",omitempty`
	IP string `json:"ip",validate:"required,ipv4"`
}
