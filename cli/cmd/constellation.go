package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/azukaar/cosmos-server/cli/internal/client"
	"github.com/azukaar/cosmos-server/cli/internal/config"
	"github.com/azukaar/cosmos-server/cli/internal/output"
	"github.com/spf13/cobra"
)

var constellationCmd = &cobra.Command{
	Use:   "constellation",
	Short: "Manage the Nebula VPN mesh network",
}

var constellationConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show the current Nebula configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationConfig(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var constellationDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List Constellation devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationDevices(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		var devices []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &devices); err != nil {
			return output.JSON(ar.Data)
		}
		rows := [][]string{}
		for _, d := range devices {
			rows = append(rows, []string{
				fmt.Sprint(d["deviceName"]),
				fmt.Sprint(d["ip"]),
				fmt.Sprint(d["isLighthouse"]),
				fmt.Sprint(d["isRelay"]),
			})
		}
		output.Table([]string{"NAME", "IP", "LIGHTHOUSE", "RELAY"}, rows)
		return nil
	},
}

var constellationPingDeviceCmd = &cobra.Command{
	Use:   "ping <id>",
	Short: "Ping a Constellation device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationDevicesIdPing(context.Background(), args[0])
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var constellationNextIPCmd = &cobra.Command{
	Use:   "next-ip",
	Short: "Get the next available IP in the Constellation network",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationGetNextIp(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var constellationLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show Nebula VPN service logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationLogs(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var constellationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check NATS client connection status",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationPing(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var constellationRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Nebula VPN service",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationRestart(context.Background())
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Constellation restarted.")
		return nil
	},
}

var constellationTunnelsCmd = &cobra.Command{
	Use:   "tunnels",
	Short: "List active Constellation tunnels",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationTunnels(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var constellationDNSCmd = &cobra.Command{
	Use:   "dns",
	Short: "List custom Constellation DNS entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationDns(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var constellationDNSGetCmd = &cobra.Command{
	Use:   "dns-get <key>",
	Short: "Get a DNS entry by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiConstellationDnsKey(context.Background(), args[0])
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var constellationDNSDeleteCmd = &cobra.Command{
	Use:   "dns-delete <key>",
	Short: "Delete a DNS entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !constellationDNSDeleteYes {
			fmt.Printf("Delete DNS entry %q? [y/N]: ", args[0])
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.DeleteApiConstellationDnsKey(context.Background(), args[0])
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("DNS entry %q deleted.", args[0])
		return nil
	},
}

var constellationDNSDeleteYes bool

func init() {
	constellationDNSDeleteCmd.Flags().BoolVar(&constellationDNSDeleteYes, "yes", false, "skip confirmation")
	constellationCmd.AddCommand(
		constellationConfigCmd,
		constellationDevicesCmd,
		constellationPingDeviceCmd,
		constellationNextIPCmd,
		constellationLogsCmd,
		constellationStatusCmd,
		constellationRestartCmd,
		constellationTunnelsCmd,
		constellationDNSCmd,
		constellationDNSGetCmd,
		constellationDNSDeleteCmd,
	)
}
