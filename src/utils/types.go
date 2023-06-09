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
	ServerCountry string
	RequireMFA bool
	AutoUpdate bool
	OpenIDClients []OpenIDClient
	MarketConfig MarketConfig
}

type HTTPConfig struct {
	TLSCert string `validate:"omitempty,contains=\n`
	TLSKey string
	TLSKeyHostsCached []string
	AuthPrivateKey string
	AuthPublicKey string
	GenerateMissingAuthCert bool
	HTTPSCertificateMode string
	DNSChallengeProvider string
	HTTPPort string `validate:"required,containsany=0123456789,min=1,max=6"`
	HTTPSPort string `validate:"required,containsany=0123456789,min=1,max=6"`
	ProxyConfig ProxyConfig
	Hostname string `validate:"required,excludesall=0x2C/ "`
	SSLEmail string `validate:"omitempty,email"`
	UseWildcardCertificate bool
	AcceptAllInsecureHostname bool
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
}

type ProxyConfig struct {
	Routes []ProxyRouteConfig
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