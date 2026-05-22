// Package config manages CLI configuration and credential storage.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const (
	keyringService = "cosmos-cli"
	configFileName = "config.yaml"
)

// Profile holds the configuration for a single server.
type Profile struct {
	URL string `yaml:"url"`
}

// Config holds all CLI configuration.
type Config struct {
	CurrentProfile string             `yaml:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
}

// Resolved holds the final URL, host header, and token for a request.
type Resolved struct {
	URL   string
	Host  string
	Token string
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".cosmos"), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Load reads the config file. Returns an empty config if the file does not exist.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		CurrentProfile: "default",
		Profiles:       make(map[string]Profile),
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	return cfg, nil
}

// Save writes the config file, creating ~/.cosmos if needed.
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("serialising config: %w", err)
	}

	path := filepath.Join(dir, configFileName)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// SetToken stores a token in the OS keychain for the given profile.
func SetToken(profile, token string) error {
	return keyring.Set(keyringService, profile, token)
}

// GetToken retrieves a token from the OS keychain for the given profile.
func GetToken(profile string) (string, error) {
	token, err := keyring.Get(keyringService, profile)
	if err != nil {
		return "", fmt.Errorf("token not found in keychain for profile %q — run 'cosmos configure'", profile)
	}
	return token, nil
}

// DeleteToken removes a token from the OS keychain for the given profile.
func DeleteToken(profile string) error {
	return keyring.Delete(keyringService, profile)
}

// hostFromURL extracts the hostname from a URL string.
func hostFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Hostname()
}

// Resolve returns the URL, host header, and token to use for a request.
// Priority: flags > env vars > config file + keychain.
func Resolve(profileFlag, urlFlag, tokenFlag string) (*Resolved, error) {
	// 1. Direct flag overrides
	if urlFlag != "" && tokenFlag != "" {
		host := os.Getenv("COSMOS_HOST")
		if host == "" {
			host = hostFromURL(urlFlag)
		}
		return &Resolved{URL: urlFlag, Host: host, Token: tokenFlag}, nil
	}

	// 2. Environment variable overrides — CI / headless servers
	envURL := os.Getenv("COSMOS_URL")
	envToken := os.Getenv("COSMOS_TOKEN")
	envHost := os.Getenv("COSMOS_HOST")

	if envURL != "" && envToken != "" {
		host := envHost
		if host == "" {
			host = hostFromURL(envURL)
		}
		return &Resolved{URL: envURL, Host: host, Token: envToken}, nil
	}

	// 3. Config file + keychain
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	profile := profileFlag
	if profile == "" {
		profile = os.Getenv("COSMOS_PROFILE")
	}
	if profile == "" {
		profile = cfg.CurrentProfile
	}
	if profile == "" {
		profile = "default"
	}

	p, ok := cfg.Profiles[profile]
	if !ok && envURL == "" && urlFlag == "" {
		return nil, fmt.Errorf("profile %q not found — run 'cosmos configure'", profile)
	}

	rawURL := urlFlag
	if rawURL == "" { rawURL = envURL }
	if rawURL == "" { rawURL = p.URL }
	if rawURL == "" {
		return nil, fmt.Errorf("no URL configured — run 'cosmos configure' or set COSMOS_URL")
	}

	token := tokenFlag
	if token == "" { token = envToken }
	if token == "" {
		token, err = GetToken(profile)
		if err != nil {
			return nil, err
		}
	}

	host := envHost
	if host == "" {
		host = hostFromURL(rawURL)
	}

	return &Resolved{URL: rawURL, Host: host, Token: token}, nil
}
