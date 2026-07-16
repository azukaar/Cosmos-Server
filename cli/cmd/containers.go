package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/cosmos-server/cli/internal/client"
	"github.com/azukaar/cosmos-server/cli/internal/config"
	"github.com/azukaar/cosmos-server/cli/internal/output"
	"github.com/spf13/cobra"
)

var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Manage Docker containers",
}

// ── list ─────────────────────────────────────────────────────────────────────

var containersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all containers",
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
		resp, err := c.GetApiServapps(context.Background(), nil)
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		var containers []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &containers); err != nil {
			return output.JSON(ar.Data)
		}
		rows := make([][]string, 0, len(containers))
		for _, ctr := range containers {
			name := ""
			if names, ok := ctr["Names"].([]interface{}); ok && len(names) > 0 {
				name = strings.TrimPrefix(fmt.Sprint(names[0]), "/")
			}
			rows = append(rows, []string{
				name,
				fmt.Sprint(ctr["State"]),
				fmt.Sprint(ctr["Image"]),
			})
		}
		output.Table([]string{"NAME", "STATE", "IMAGE"}, rows)
		return nil
	},
}

// ── inspect ───────────────────────────────────────────────────────────────────

var containersInspectCmd = &cobra.Command{
	Use:   "inspect <id>",
	Short: "Inspect a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiServappsContainerId(context.Background(), args[0])
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

// ── manage actions (start / stop / restart / recreate) ────────────────────────

func containerManageAction(action cosmossdk.GetApiServappsContainerIdManageActionParamsAction) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiServappsContainerIdManageAction(context.Background(), args[0], action)
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("container %s: %s", args[0], string(action))
		return nil
	}
}

var containersStartCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "Start a container",
	Args:  cobra.ExactArgs(1),
	RunE:  containerManageAction(cosmossdk.Start),
}

var containersStopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "Stop a container",
	Args:  cobra.ExactArgs(1),
	RunE:  containerManageAction(cosmossdk.Stop),
}

var containersRestartCmd = &cobra.Command{
	Use:   "restart <id>",
	Short: "Restart a container",
	Args:  cobra.ExactArgs(1),
	RunE:  containerManageAction(cosmossdk.Restart),
}

var containersRecreateCmd = &cobra.Command{
	Use:   "recreate <id>",
	Short: "Recreate a container",
	Args:  cobra.ExactArgs(1),
	RunE:  containerManageAction(cosmossdk.Recreate),
}

// ── logs ─────────────────────────────────────────────────────────────────────

var containersTailFlag int

var containersLogsCmd = &cobra.Command{
	Use:   "logs <id>",
	Short: "Fetch container logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		params := &cosmossdk.GetApiServappsContainerIdLogsParams{
			Limit: &containersTailFlag,
		}
		resp, err := c.GetApiServappsContainerIdLogs(context.Background(), args[0], params)
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

// ── check-update ─────────────────────────────────────────────────────────────

var containersCheckUpdateCmd = &cobra.Command{
	Use:   "check-update <id>",
	Short: "Check if a container image has an update",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.GetApiServappsContainerIdCheckUpdate(context.Background(), args[0])
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

// ── networks ──────────────────────────────────────────────────────────────────

var containersNetworksCmd = &cobra.Command{
	Use:   "networks",
	Short: "List Docker networks",
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
		resp, err := c.GetApiNetworks(context.Background())
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		var networks []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &networks); err != nil {
			return output.JSON(ar.Data)
		}
		rows := make([][]string, 0, len(networks))
		for _, n := range networks {
			rows = append(rows, []string{
				fmt.Sprint(n["Name"]),
				fmt.Sprint(n["Driver"]),
				fmt.Sprint(n["Id"]),
			})
		}
		output.Table([]string{"NAME", "DRIVER", "ID"}, rows)
		return nil
	},
}

// ── volumes ───────────────────────────────────────────────────────────────────

var containersVolumesCmd = &cobra.Command{
	Use:   "volumes",
	Short: "List Docker volumes",
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
		resp, err := c.GetApiVolumes(context.Background())
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		var volumes []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &volumes); err != nil {
			return output.JSON(ar.Data)
		}
		rows := make([][]string, 0, len(volumes))
		for _, v := range volumes {
			rows = append(rows, []string{
				fmt.Sprint(v["Name"]),
				fmt.Sprint(v["Driver"]),
			})
		}
		output.Table([]string{"NAME", "DRIVER"}, rows)
		return nil
	},
}

// ── images ────────────────────────────────────────────────────────────────────

var containersImageNameFlag string

var containersImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Inspect a Docker image",
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
		params := &cosmossdk.GetApiImagesParams{
			ImageName: containersImageNameFlag,
		}
		resp, err := c.GetApiImages(context.Background(), params)
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

// ── pull ──────────────────────────────────────────────────────────────────────

var containersPullCmd = &cobra.Command{
	Use:   "pull <name>",
	Short: "Pull a Docker image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		params := &cosmossdk.GetApiImagesPullParams{
			ImageName: args[0],
		}
		resp, err := c.GetApiImagesPull(context.Background(), params)
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		output.Success("pulled image %s", args[0])
		_ = ar
		return nil
	},
}

func init() {
	containersLogsCmd.Flags().IntVar(&containersTailFlag, "tail", 100, "Number of log lines to fetch")
	containersImagesCmd.Flags().StringVar(&containersImageNameFlag, "name", "", "Image name to inspect")

	containersCmd.AddCommand(
		containersListCmd,
		containersInspectCmd,
		containersStartCmd,
		containersStopCmd,
		containersRestartCmd,
		containersRecreateCmd,
		containersLogsCmd,
		containersCheckUpdateCmd,
		containersNetworksCmd,
		containersVolumesCmd,
		containersImagesCmd,
		containersPullCmd,
	)
}
