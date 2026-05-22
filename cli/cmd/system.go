package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/cosmos-server/cli/internal/client"
	"github.com/azukaar/cosmos-server/cli/internal/config"
	"github.com/azukaar/cosmos-server/cli/internal/output"
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System status and operations",
}

// ── status ────────────────────────────────────────────────────────────────────

var systemStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show server status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiStatus(context.Background())
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		// Pretty key/value output
		var kv map[string]interface{}
		if err := json.Unmarshal(ar.Data, &kv); err != nil {
			return output.JSON(ar.Data)
		}
		maxLen := 0
		keys := make([]string, 0, len(kv))
		for k := range kv {
			keys = append(keys, k)
			if len(k) > maxLen {
				maxLen = len(k)
			}
		}
		for _, k := range keys {
			fmt.Printf("  %-*s  %v\n", maxLen, k, kv[k])
		}
		return nil
	},
}

// ── ping ──────────────────────────────────────────────────────────────────────

var systemPingURLFlag string

var systemPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping a URL",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if systemPingURLFlag == "" {
			return fmt.Errorf("--url is required")
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		params := &cosmossdk.GetApiPingParams{
			Q: systemPingURLFlag,
		}
		resp, err := c.GetApiPing(context.Background(), params)
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		return output.JSON(ar.Data)
	},
}

// ── dns ───────────────────────────────────────────────────────────────────────

var systemDNSURLFlag string

var systemDNSCmd = &cobra.Command{
	Use:   "dns",
	Short: "Resolve DNS for a URL",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if systemDNSURLFlag == "" {
			return fmt.Errorf("--url is required")
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		params := &cosmossdk.GetApiDnsParams{
			Url: systemDNSURLFlag,
		}
		resp, err := c.GetApiDns(context.Background(), params)
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		return output.JSON(ar.Data)
	},
}

// ── dns-check ─────────────────────────────────────────────────────────────────

var systemDNSCheckURLFlag string

var systemDNSCheckCmd = &cobra.Command{
	Use:   "dns-check",
	Short: "Check DNS resolution for a URL",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if systemDNSCheckURLFlag == "" {
			return fmt.Errorf("--url is required")
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		params := &cosmossdk.GetApiDnsCheckParams{
			Url: systemDNSCheckURLFlag,
		}
		resp, err := c.GetApiDnsCheck(context.Background(), params)
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		return output.JSON(ar.Data)
	},
}

// ── can-send-email ────────────────────────────────────────────────────────────

var systemCanSendEmailCmd = &cobra.Command{
	Use:   "can-send-email",
	Short: "Check whether the server can send email",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiCanSendEmail(context.Background())
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		return output.JSON(ar.Data)
	},
}

// ── force-update ──────────────────────────────────────────────────────────────

var systemForceUpdateCmd = &cobra.Command{
	Use:   "force-update",
	Short: "Force a server update check",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.PostApiForceServerUpdate(context.Background())
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("server update check triggered")
		return nil
	},
}

// ── memory ────────────────────────────────────────────────────────────────────

var systemMemoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Show server memory stats",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiMemory(context.Background())
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		return output.JSON(ar.Data)
	},
}

// ── restart (Cosmos service) ──────────────────────────────────────────────────

var systemRestartYesFlag bool

var systemRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Cosmos service",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !systemRestartYesFlag {
			fmt.Print("Are you sure you want to restart the Cosmos service? [y/N]: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiRestart(context.Background())
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("Cosmos service restarting")
		return nil
	},
}

// ── restart-host (HOST machine) ───────────────────────────────────────────────

var systemRestartHostYesFlag bool

var systemRestartHostCmd = &cobra.Command{
	Use:   "restart-host",
	Short: "Restart the HOST machine (dangerous)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !systemRestartHostYesFlag {
			fmt.Print("WARNING: This will restart the HOST machine. Are you sure? [y/N]: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiRestartServer(context.Background())
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("host machine restarting")
		return nil
	},
}

// ── notifications ─────────────────────────────────────────────────────────────

var systemNotificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "List notifications",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiNotifications(context.Background(), nil)
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		return output.JSON(ar.Data)
	},
}

// ── notifications-read ────────────────────────────────────────────────────────

var systemNotificationsIDsFlag string

var systemNotificationsReadCmd = &cobra.Command{
	Use:   "notifications-read",
	Short: "Mark notifications as read",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if systemNotificationsIDsFlag == "" {
			return fmt.Errorf("--ids is required (comma-separated notification IDs)")
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		params := &cosmossdk.GetApiNotificationsReadParams{
			Ids: systemNotificationsIDsFlag,
		}
		resp, err := c.GetApiNotificationsRead(context.Background(), params)
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("notifications marked as read")
		return nil
	},
}

func init() {
	systemPingCmd.Flags().StringVar(&systemPingURLFlag, "url", "", "URL to ping")
	systemDNSCmd.Flags().StringVar(&systemDNSURLFlag, "url", "", "URL to resolve DNS for")
	systemDNSCheckCmd.Flags().StringVar(&systemDNSCheckURLFlag, "url", "", "URL to check DNS for")
	systemRestartCmd.Flags().BoolVar(&systemRestartYesFlag, "yes", false, "Skip confirmation prompt")
	systemRestartHostCmd.Flags().BoolVar(&systemRestartHostYesFlag, "yes", false, "Skip confirmation prompt")
	systemNotificationsReadCmd.Flags().StringVar(&systemNotificationsIDsFlag, "ids", "", "Comma-separated notification IDs to mark as read")

	systemCmd.AddCommand(
		systemStatusCmd,
		systemPingCmd,
		systemDNSCmd,
		systemDNSCheckCmd,
		systemCanSendEmailCmd,
		systemForceUpdateCmd,
		systemMemoryCmd,
		systemRestartCmd,
		systemRestartHostCmd,
		systemNotificationsCmd,
		systemNotificationsReadCmd,
	)
}
