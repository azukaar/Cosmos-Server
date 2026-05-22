// Package config manages CLI configuration and credential storage.
//
// Configuration is stored in ~/.cosmos/config.yaml.
// Tokens are stored in the OS keychain (macOS Keychain, Linux libsecret, Windows Credential Manager).
// Environment variables COSMOS_URL and COSMOS_TOKEN override config and keychain respectively.
package config

import (
	"fmt"
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

// Resolved holds the final resolved URL and token for a request.
type Resolved struct {
	URL   string
	Token string
}

// configDir returns ~/.cosmos
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

// Resolve returns the URL and token to use for a request, applying the
// following priority: flags > env vars > config file + keychain.
func Resolve(profileFlag, urlFlag, tokenFlag string) (*Resolved, error) {
	// 1. Direct flag overrides — used in scripts / one-off commands
	if urlFlag != "" && tokenFlag != "" {
		return &Resolved{URL: urlFlag, Token: tokenFlag}, nil
	}

	// 2. Environment variable overrides — CI / headless servers
	envURL := os.Getenv("COSMOS_URL")
	envToken := os.Getenv("COSMOS_TOKEN")
	if envURL != "" && envToken != "" {
		return &Resolved{URL: envURL, Token: envToken}, nil
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

	url := urlFlag
	if url == "" {
		url = envURL
	}
	if url == "" {
		url = p.URL
	}
	if url == "" {
		return nil, fmt.Errorf("no URL configured — run 'cosmos configure' or set COSMOS_URL")
	}

	token := tokenFlag
	if token == "" {
		token = envToken
	}
	if token == "" {
		token, err = GetToken(profile)
		if err != nil {
			return nil, err
		}
	}

	return &Resolved{URL: url, Token: token}, nil
}
