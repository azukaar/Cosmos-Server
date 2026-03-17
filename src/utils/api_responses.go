package utils

// APIResponse is the standard success response wrapper used by all endpoints.
type APIResponse struct {
	Status string      `json:"status" example:"OK"`
	Data   interface{} `json:"data,omitempty"`
}

// APIResponseMessage is used when only status and message are returned.
type APIResponseMessage struct {
	Status  string `json:"status" example:"OK"`
	Message string `json:"message,omitempty"`
}

// StatusData represents the data returned by the /api/status endpoint.
type StatusData struct {
	HomepageConfig HomepageConfig `json:"homepage"`
	ThemeConfig    ThemeConfig    `json:"theme"`
	CPU            string         `json:"CPU"`
	AVX            bool           `json:"AVX"`
	LetsEncryptErrors []string    `json:"LetsEncryptErrors"`
	IsAdmin        bool           `json:"isAdmin"`
	MonitoringDisabled bool       `json:"MonitoringDisabled"`
	Hostname       string         `json:"hostname"`
	Domain         bool           `json:"domain"`
	HTTPSCertificateMode string   `json:"HTTPSCertificateMode"`
	NewVersionAvailable string    `json:"newVersionAvailable"`
	NeedsRestart   bool           `json:"needRestart"`
	Database       bool           `json:"database"`
	Docker         bool           `json:"docker"`
	BackupStatus   string         `json:"backup_status"`
	Constellation  bool           `json:"constellation"`
}
