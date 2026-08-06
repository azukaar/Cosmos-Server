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

var openIDCmd = &cobra.Command{
	Use:   "openid",
	Short: "Manage OpenID Connect clients",
}

var openIDListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all OpenID clients",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiOpenid(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		var clients []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &clients); err != nil {
			return output.JSON(ar.Data)
		}
		rows := [][]string{}
		for _, oc := range clients {
			rows = append(rows, []string{
				fmt.Sprint(oc["id"]),
				fmt.Sprint(oc["clientName"]),
			})
		}
		output.Table([]string{"ID", "NAME"}, rows)
		return nil
	},
}

var openIDGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an OpenID client by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiOpenidId(context.Background(), args[0])
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var openIDDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an OpenID client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !openIDDeleteYes {
			fmt.Printf("Delete OpenID client %q? [y/N]: ", args[0])
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
		resp, err := c.DeleteApiOpenidId(context.Background(), args[0])
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("OpenID client %q deleted.", args[0])
		return nil
	},
}

var openIDDeleteYes bool

func init() {
	openIDDeleteCmd.Flags().BoolVar(&openIDDeleteYes, "yes", false, "skip confirmation")
	openIDCmd.AddCommand(openIDListCmd, openIDGetCmd, openIDDeleteCmd)
}
