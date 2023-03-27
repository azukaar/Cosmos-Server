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
}

type Config struct {
	LoggingLevel LoggingLevel `validate:"oneof=DEBUG INFO WARNING ERROR"`
	MongoDB string
	HTTPConfig HTTPConfig
}

type HTTPConfig struct {
	TLSCert string
	TLSKey string
	AuthPrivateKey string
	AuthPublicKey string
	GenerateMissingAuthCert bool
	HTTPSCertificateMode string
	HTTPPort string
	HTTPSPort string
	ProxyConfig ProxyConfig
	Hostname string
	SSLEmail string
} 

type ProxyConfig struct {
	Routes []ProxyRouteConfig
}

type ProxyRouteConfig struct {
	Name string
	Description string
	UseHost bool
  Host string
	UsePathPrefix bool
	PathPrefix string
	Timeout time.Duration
	ThrottlePerMinute int
	CORSOrigin string
	StripPathPrefix bool
	AuthEnabled bool
	Target  string `validate:"required"`
	Mode ProxyMode
}
