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

var backupsCmd = &cobra.Command{
	Use:   "backups",
	Short: "Manage Restic backup configurations and operations",
}

var backupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backup configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiBackupsConfig(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		var configs map[string]map[string]interface{}
		if err := json.Unmarshal(ar.Data, &configs); err != nil {
			return output.JSON(ar.Data)
		}
		rows := [][]string{}
		for name := range configs {
			rows = append(rows, []string{name})
		}
		output.Table([]string{"NAME"}, rows)
		return nil
	},
}

var backupsReposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List backup repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiBackupsRepository(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var backupsSnapshotsCmd = &cobra.Command{
	Use:   "snapshots <name>",
	Short: "List snapshots for a backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiBackupsNameSnapshots(context.Background(), args[0])
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var backupsRestoreCmd = &cobra.Command{
	Use:   "restore <name>",
	Short: "Restore files from a backup snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		body := cosmossdk.PostApiBackupsNameRestoreJSONRequestBody{
			"snapshot":   backupsSnapshot,
			"targetPath": backupsTarget,
		}
		resp, err := c.PostApiBackupsNameRestore(context.Background(), args[0], body)
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Restore started.")
		return nil
	},
}

var backupsUnlockCmd = &cobra.Command{
	Use:   "unlock <name>",
	Short: "Unlock a backup repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.PostApiBackupsNameUnlock(context.Background(), args[0])
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Repository unlocked.")
		return nil
	},
}

var backupsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a backup configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !backupsDeleteYes {
			fmt.Printf("Delete backup %q? [y/N]: ", args[0])
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
		resp, err := c.DeleteApiBackupsName(context.Background(), args[0])
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Backup %q deleted.", args[0])
		return nil
	},
}

var (
	backupsSnapshot  string
	backupsTarget    string
	backupsDeleteYes bool
)

func init() {
	backupsRestoreCmd.Flags().StringVar(&backupsSnapshot, "snapshot", "", "snapshot ID (required)")
	backupsRestoreCmd.Flags().StringVar(&backupsTarget, "target", "", "target restore path (required)")
	_ = backupsRestoreCmd.MarkFlagRequired("snapshot")
	_ = backupsRestoreCmd.MarkFlagRequired("target")
	backupsDeleteCmd.Flags().BoolVar(&backupsDeleteYes, "yes", false, "skip confirmation")
	backupsCmd.AddCommand(
		backupsListCmd,
		backupsReposCmd,
		backupsSnapshotsCmd,
		backupsRestoreCmd,
		backupsUnlockCmd,
		backupsDeleteCmd,
	)
}
