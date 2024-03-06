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
	Nickname      string     `validate:"required" json:"nickname" bson:"Nickname"`
	Password       string    `validate:"required" json:"-" bson:"Password"`
	RegisterKey       string  `json:"registerKey" bson:"RegisterKey"`
	RegisterKeyExp       time.Time  `json:"registerKeyExp" bson:"RegisterKeyExp"`
	Role       Role    `validate:"required" json:"role" bson:"Role"`
	PasswordCycle			 int    `json:"-" bson:"PasswordCycle"`
	Link 		 string    `json:"link" bson:"-"`
	Email string `validate:"email" json:"email" bson:"Email"`
	RegisteredAt time.Time   `json:"registeredAt" bson:"RegisteredAt"`
	LastPasswordChangedAt time.Time   `json:"lastPasswordChangedAt" bson:"LastPasswordChangedAt"`
	CreatedAt time.Time   `json:"createdAt" bson:"CreatedAt"`
	LastLogin time.Time   `json:"lastLogin" bson:"LastLogin"`
	MFAKey string `json:"-" bson:"MFAKey"`
	Was2FAVerified bool `json:"-" bson:"Was2FAVerified"`
	MFAState int `json:"-" bson:"-"` 
	// 0 = done, 1 = needed, 2 = not set
}

type Config struct {
	LoggingLevel LoggingLevel `required,validate:"oneof=DEBUG INFO WARNING ERROR"`
	MongoDB string
	Database DatabaseConfig `validate:"dive"`
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
	MonitoringDisabled bool
	MonitoringAlerts map[string]Alert
	BackupOutputDir string
	DisableHostModeWarning bool
	AdminWhitelistIPs []string
	AdminConstellationOnly bool
	Storage StorageConfig
	CRON map[string]CRONConfig
}

type CRONConfig struct {
	Enabled bool
	Name string
	Crontab string
	Command string
}

type StorageConfig struct {
	SnapRAIDs []SnapRAIDConfig
}

type SnapRAIDConfig struct {
	Name string
	Enabled bool
	Data []string
	Parity []string
	SyncCrontab string
	ScrubCrontab string
	CheckOnFix bool
}

type DatabaseConfig struct {
	PuppetMode bool
	Hostname string
	DbVolume string
	ConfigVolume string
	Version string
	Username string
	Password string
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
	UseForwardedFor bool
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
	SkipPruneImages bool
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
	Disabled bool
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
	RestrictToConstellation bool
	OverwriteHostHeader string
	WhitelistInboundIPs []string
	Icon string
}

type EmailConfig struct {
	Enabled		 bool
	Host       string
	Port       string
	Username   string
	Password   string
	From       string
	UseTLS		 bool
	AllowInsecureTLS		 bool
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
	SlaveMode bool
	PrivateNode bool
	DNSDisabled bool
	DNSPort string
	DNSFallback string
	DNSBlockBlacklist bool
	DNSAdditionalBlocklists []string
	CustomDNSEntries []ConstellationDNSEntry
	NebulaConfig NebulaConfig
	ConstellationHostname string
}

type ConstellationDNSEntry struct {
	Type string
	Key string
	Value string
}
type ConstellationDevice struct {
	Nickname string `json:"nickname" bson:"Nickname"`
	DeviceName string `json:"deviceName" bson:"DeviceName"`
	PublicKey string `json:"publicKey" bson:"PublicKey"`
	IP string `json:"ip" bson:"IP"`
	IsLighthouse bool `json:"isLighthouse" bson:"IsLighthouse"`
	IsRelay bool `json:"isRelay" bson:"IsRelay"`
	PublicHostname string `json:"publicHostname" bson:"PublicHostname"`
	Port string `json:"port" bson:"Port"`
	Blocked bool `json:"blocked" bson:"Blocked"`
	Fingerprint string `json:"fingerprint" 	bson:"Fingerprint"`
	APIKey string `json:"-" bson:"APIKey"`
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
		Blocklist []string `yaml:"blocklist"`
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
	DeviceName string `json:"deviceName"validate:"required,min=3,max=32,alphanum",bson:"DeviceName"`
	Nickname string `json:"nickname",validate:"required,min=3,max=32,alphanum",bson:"Nickname"`
	PublicKey string `json:"publicKey",omitempty,bson:"PublicKey"`
	PrivateKey string `json:"privateKey",omitempty,bson:"PrivateKey"`
	IP string `json:"ip",validate:"required,ipv4",bson:"IP"`
}

type Alert struct {
	Name string
	Enabled bool
	Period string
	TrackingMetric string
	Condition AlertCondition
	Actions []AlertAction
	LastTriggered time.Time
	Throttled bool
	Severity string
}

type AlertCondition struct {
	Operator string
	Value int
	Percent bool
}

type AlertAction struct {
	Type string
	Target string
}

type AlertMetricTrack struct {
	Key string
	Object string
	Max uint64
}