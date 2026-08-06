package cmd

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/azukaar/cosmos-server/cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Cosmos CLI credentials and server URL",
	Long: `Interactive wizard to set up a server profile.

Saves the server URL to ~/.cosmos/config.yaml and stores the API token
securely in the OS keychain (macOS Keychain, Linux libsecret, Windows Credential Manager).

Examples:
  cosmos configure                        # configure default profile
  cosmos configure --profile homelab      # configure a named profile
  cosmos configure list                   # list all profiles
  cosmos configure delete --profile vps   # remove a profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigure(profileFlag)
	},
}

var configureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Profiles) == 0 {
			fmt.Println("No profiles configured. Run 'cosmos configure' to get started.")
			return nil
		}
		fmt.Printf("  %-20s  %s\n", "PROFILE", "URL")
		fmt.Printf("  %-20s  %s\n", strings.Repeat("-", 20), strings.Repeat("-", 40))
		for name, p := range cfg.Profiles {
			marker := " "
			if name == cfg.CurrentProfile {
				marker = "*"
			}
			fmt.Printf("%s %-20s  %s\n", marker, name, p.URL)
		}
		return nil
	},
}

var configureDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := profileFlag
		if profile == "" {
			return fmt.Errorf("--profile is required for delete")
		}
		if profile == "default" {
			return fmt.Errorf("cannot delete the default profile")
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if _, ok := cfg.Profiles[profile]; !ok {
			return fmt.Errorf("profile %q not found", profile)
		}
		delete(cfg.Profiles, profile)
		if cfg.CurrentProfile == profile {
			cfg.CurrentProfile = "default"
		}
		_ = config.DeleteToken(profile)
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("  ✓ Profile %q deleted\n", profile)
		return nil
	},
}

var configureUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Switch the active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if _, ok := cfg.Profiles[profile]; !ok {
			return fmt.Errorf("profile %q not found — run 'cosmos configure --profile %s'", profile, profile)
		}
		cfg.CurrentProfile = profile
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("  ✓ Switched to profile %q\n", profile)
		return nil
	},
}

func init() {
	configureCmd.AddCommand(configureListCmd, configureDeleteCmd, configureUseCmd)
}

func runConfigure(profile string) error {
	if profile == "" {
		profile = "default"
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	existing := cfg.Profiles[profile]
	urlHint := existing.URL
	if urlHint == "" {
		urlHint = "http://localhost:80"
	}

	fmt.Printf("Configuring profile %q\n\n", profile)
	fmt.Printf("Server URL [%s]: ", urlHint)
	var inputURL string
	fmt.Scanln(&inputURL)
	if inputURL == "" {
		inputURL = urlHint
	}
	if inputURL == "" {
		return fmt.Errorf("URL is required")
	}

	fmt.Print("API Token: ")
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("token is required")
	}

	cfg.Profiles[profile] = config.Profile{URL: inputURL}
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = profile
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if err := config.SetToken(profile, token); err != nil {
		return fmt.Errorf("saving token to keychain: %w", err)
	}

	fmt.Printf("\n  ✓ URL saved to ~/.cosmos/config.yaml\n")
	fmt.Printf("  ✓ Token saved to OS keychain (profile: %q)\n", profile)
	if cfg.CurrentProfile == profile {
		fmt.Printf("  ✓ Active profile: %q\n", profile)
	}
	return nil
}
