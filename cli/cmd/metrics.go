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
	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Alerts, events, and metrics data",
}

var metricsAlertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "List monitoring alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiAlerts(context.Background(), cosmossdk.GetApiAlertsJSONRequestBody{})
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		var alerts map[string]map[string]interface{}
		if err := json.Unmarshal(ar.Data, &alerts); err != nil {
			return output.JSON(ar.Data)
		}
		rows := [][]string{}
		for name, a := range alerts {
			rows = append(rows, []string{name, fmt.Sprint(a["Enabled"])})
		}
		output.Table([]string{"NAME", "ENABLED"}, rows)
		return nil
	},
}

var metricsEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "List system events",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		var params *cosmossdk.GetApiEventsParams
		if metricsFrom != "" || metricsTo != "" {
			params = &cosmossdk.GetApiEventsParams{}
			if metricsFrom != "" { params.From = &metricsFrom }
			if metricsTo != "" { params.To = &metricsTo }
		}
		resp, err := c.GetApiEvents(context.Background(), params)
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var metricsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available metric keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiListMetrics(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var metricsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show metrics data",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		var params *cosmossdk.GetApiMetricsParams
		if metricsKey != "" {
			params = &cosmossdk.GetApiMetricsParams{Metrics: &metricsKey}
		}
		resp, err := c.GetApiMetrics(context.Background(), params)
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var metricsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all metrics and events data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !metricsResetYes {
			fmt.Print("This will delete all metrics and events. Are you sure? [y/N]: ")
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
		resp, err := c.GetApiResetMetrics(context.Background())
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Metrics reset.")
		return nil
	},
}

var (
	metricsFrom     string
	metricsTo       string
	metricsKey      string
	metricsResetYes bool
)

func init() {
	metricsEventsCmd.Flags().StringVar(&metricsFrom, "from", "", "start date RFC3339 (e.g. 2026-01-01T00:00:00Z)")
	metricsEventsCmd.Flags().StringVar(&metricsTo, "to", "", "end date RFC3339")
	metricsShowCmd.Flags().StringVar(&metricsKey, "key", "", "filter by metric key")
	metricsResetCmd.Flags().BoolVar(&metricsResetYes, "yes", false, "skip confirmation")
	metricsCmd.AddCommand(metricsAlertsCmd, metricsEventsCmd, metricsListCmd, metricsShowCmd, metricsResetCmd)
}
