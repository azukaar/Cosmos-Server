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

var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Manage proxy routes",
}

// ── list ─────────────────────────────────────────────────────────────────────

var routesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all routes",
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
		resp, err := c.GetApiRoutes(context.Background())
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		var routes []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &routes); err != nil {
			return output.JSON(ar.Data)
		}
		rows := make([][]string, 0, len(routes))
		for _, rt := range routes {
			rows = append(rows, []string{
				fmt.Sprint(rt["Name"]),
				fmt.Sprint(rt["Mode"]),
				fmt.Sprint(rt["Target"]),
			})
		}
		output.Table([]string{"NAME", "MODE", "TARGET"}, rows)
		return nil
	},
}

// ── get ───────────────────────────────────────────────────────────────────────

var routesGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a route by name",
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
		resp, err := c.GetApiRoutesName(context.Background(), args[0])
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

// ── create ────────────────────────────────────────────────────────────────────

var routesDataFlag string

var routesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new route",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if routesDataFlag == "" {
			return fmt.Errorf("--data is required")
		}
		var body cosmossdk.PostApiRoutesJSONRequestBody
		if err := json.Unmarshal([]byte(routesDataFlag), &body); err != nil {
			return fmt.Errorf("invalid JSON in --data: %w", err)
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.PostApiRoutes(context.Background(), body)
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("route created")
		return nil
	},
}

// ── update ────────────────────────────────────────────────────────────────────

var routesUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update an existing route",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if routesDataFlag == "" {
			return fmt.Errorf("--data is required")
		}
		var body cosmossdk.PutApiRoutesNameJSONRequestBody
		if err := json.Unmarshal([]byte(routesDataFlag), &body); err != nil {
			return fmt.Errorf("invalid JSON in --data: %w", err)
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.PutApiRoutesName(context.Background(), args[0], body)
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("route %q updated", args[0])
		return nil
	},
}

// ── delete ────────────────────────────────────────────────────────────────────

var routesYesFlag bool

var routesDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a route",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !routesYesFlag {
			fmt.Printf("Are you sure you want to delete route %q? [y/N]: ", args[0])
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
		resp, err := c.DeleteApiRoutesName(context.Background(), args[0])
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("route %q deleted", args[0])
		return nil
	},
}

func init() {
	routesCreateCmd.Flags().StringVar(&routesDataFlag, "data", "", "Route config as JSON string")
	routesUpdateCmd.Flags().StringVar(&routesDataFlag, "data", "", "Route config as JSON string")
	routesDeleteCmd.Flags().BoolVar(&routesYesFlag, "yes", false, "Skip confirmation prompt")

	routesCmd.AddCommand(
		routesListCmd,
		routesGetCmd,
		routesCreateCmd,
		routesUpdateCmd,
		routesDeleteCmd,
	)
}
