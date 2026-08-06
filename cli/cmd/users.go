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

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage Cosmos users",
}

// ── list ─────────────────────────────────────────────────────────────────────

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
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
		resp, err := c.GetApiUsers(context.Background(), nil)
		if err != nil {
			return err
		}
		ar, err := client.Parse(resp)
		if err != nil {
			return err
		}
		var users []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &users); err != nil {
			return output.JSON(ar.Data)
		}
		rows := make([][]string, 0, len(users))
		for _, u := range users {
			role := ""
			if rv, ok := u["role"]; ok {
				role = fmt.Sprint(rv)
			}
			email := ""
			if ev, ok := u["email"]; ok {
				email = fmt.Sprint(ev)
			}
			rows = append(rows, []string{
				fmt.Sprint(u["nickname"]),
				role,
				email,
			})
		}
		output.Table([]string{"NICKNAME", "ROLE", "EMAIL"}, rows)
		return nil
	},
}

// ── get ───────────────────────────────────────────────────────────────────────

var usersGetCmd = &cobra.Command{
	Use:   "get <nickname>",
	Short: "Get a user by nickname",
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
		resp, err := c.GetApiUsersNickname(context.Background(), args[0])
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

var (
	usersCreateNickname string
	usersCreatePassword string
	usersCreateEmail    string
)

var usersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if usersCreateNickname == "" {
			return fmt.Errorf("--nickname is required")
		}
		if usersCreatePassword == "" {
			return fmt.Errorf("--password is required")
		}
		body := cosmossdk.PostApiUsersJSONRequestBody{
			Nickname: usersCreateNickname,
		}
		if usersCreateEmail != "" {
			body.Email = &usersCreateEmail
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.PostApiUsers(context.Background(), body)
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("user %q created", usersCreateNickname)
		return nil
	},
}

// ── delete ────────────────────────────────────────────────────────────────────

var usersYesFlag bool

var usersDeleteCmd = &cobra.Command{
	Use:   "delete <nickname>",
	Short: "Delete a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !usersYesFlag {
			fmt.Printf("Are you sure you want to delete user %q? [y/N]: ", args[0])
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
		resp, err := c.DeleteApiUsersNickname(context.Background(), args[0])
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("user %q deleted", args[0])
		return nil
	},
}

// ── invite ────────────────────────────────────────────────────────────────────

var usersInviteCmd = &cobra.Command{
	Use:   "invite <nickname>",
	Short: "Resend invite to a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := cosmossdk.PostApiInviteJSONRequestBody{
			Nickname: args[0],
		}
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil {
			return err
		}
		c, err := client.New(r)
		if err != nil {
			return err
		}
		resp, err := c.PostApiInvite(context.Background(), body)
		if err != nil {
			return err
		}
		if _, err := client.Parse(resp); err != nil {
			return err
		}
		output.Success("invite sent to %q", args[0])
		return nil
	},
}

func init() {
	usersCreateCmd.Flags().StringVar(&usersCreateNickname, "nickname", "", "User nickname")
	usersCreateCmd.Flags().StringVar(&usersCreatePassword, "password", "", "User password")
	usersCreateCmd.Flags().StringVar(&usersCreateEmail, "email", "", "User email address")
	usersDeleteCmd.Flags().BoolVar(&usersYesFlag, "yes", false, "Skip confirmation prompt")

	usersCmd.AddCommand(
		usersListCmd,
		usersGetCmd,
		usersCreateCmd,
		usersDeleteCmd,
		usersInviteCmd,
	)
}
