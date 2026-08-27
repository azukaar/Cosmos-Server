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

var apiTokensCmd = &cobra.Command{
	Use:   "api-tokens",
	Short: "Manage Cosmos Server API tokens",
}

var apiTokensListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiApiTokens(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		var tokens map[string]map[string]interface{}
		if err := json.Unmarshal(ar.Data, &tokens); err != nil {
			return output.JSON(ar.Data)
		}
		rows := [][]string{}
		for name, t := range tokens {
			rows = append(rows, []string{
				name,
				fmt.Sprint(t["owner"]),
				fmt.Sprint(t["tokenSuffix"]),
			})
		}
		output.Table([]string{"NAME", "OWNER", "SUFFIX"}, rows)
		return nil
	},
}

var apiTokensCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API token",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		body := cosmossdk.PostApiApiTokensJSONRequestBody{
			Name:        apiTokensName,
			Description: &apiTokensDesc,
		}
		resp, err := c.PostApiApiTokens(context.Background(), body)
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var apiTokensDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !apiTokensDeleteYes {
			fmt.Printf("Delete token %q? [y/N]: ", args[0])
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
		body := cosmossdk.DeleteApiApiTokensJSONRequestBody{Name: args[0]}
		resp, err := c.DeleteApiApiTokens(context.Background(), body)
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Token %q deleted.", args[0])
		return nil
	},
}

var (
	apiTokensName      string
	apiTokensDesc      string
	apiTokensDeleteYes bool
)

func init() {
	apiTokensCreateCmd.Flags().StringVar(&apiTokensName, "name", "", "token name (required)")
	apiTokensCreateCmd.Flags().StringVar(&apiTokensDesc, "description", "", "token description")
	_ = apiTokensCreateCmd.MarkFlagRequired("name")
	apiTokensDeleteCmd.Flags().BoolVar(&apiTokensDeleteYes, "yes", false, "skip confirmation")
	apiTokensCmd.AddCommand(apiTokensListCmd, apiTokensCreateCmd, apiTokensDeleteCmd)
}
