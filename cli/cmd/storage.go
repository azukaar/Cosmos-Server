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

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage disks, mounts, and SnapRAID",
}

var storageDisksCmd = &cobra.Command{
	Use:   "disks",
	Short: "List all disks",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiDisks(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		var disks []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &disks); err != nil {
			return output.JSON(ar.Data)
		}
		rows := [][]string{}
		for _, d := range disks {
			rows = append(rows, []string{
				fmt.Sprint(d["name"]),
				fmt.Sprint(d["size"]),
				fmt.Sprint(d["type"]),
				fmt.Sprint(d["serial"]),
			})
		}
		output.Table([]string{"NAME", "SIZE", "TYPE", "SERIAL"}, rows)
		return nil
	},
}

var storageMountsCmd = &cobra.Command{
	Use:   "mounts",
	Short: "List mounted filesystems",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiMounts(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		var mounts []map[string]interface{}
		if err := json.Unmarshal(ar.Data, &mounts); err != nil {
			return output.JSON(ar.Data)
		}
		rows := [][]string{}
		for _, m := range mounts {
			rows = append(rows, []string{
				fmt.Sprint(m["path"]),
				fmt.Sprint(m["device"]),
				fmt.Sprint(m["type"]),
			})
		}
		output.Table([]string{"PATH", "DEVICE", "TYPE"}, rows)
		return nil
	},
}

var storageListDirCmd = &cobra.Command{
	Use:   "list-dir",
	Short: "List contents of a directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		params := &cosmossdk.GetApiListDirParams{Path: &storagePath}
		resp, err := c.GetApiListDir(context.Background(), params)
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var storageUnmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount a filesystem",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		// StorageMountRequest.Path is string (not pointer)
		body := cosmossdk.PostApiUnmountJSONRequestBody{Path: storagePath}
		resp, err := c.PostApiUnmount(context.Background(), body)
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Unmounted %q.", storagePath)
		return nil
	},
}

var storageNewDirCmd = &cobra.Command{
	Use:   "new-dir",
	Short: "Create a new directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		// PostApiNewDir takes params (query), not a body
		params := &cosmossdk.PostApiNewDirParams{Path: &storagePath}
		resp, err := c.PostApiNewDir(context.Background(), params)
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Directory %q created.", storagePath)
		return nil
	},
}

var storageSnapraidCmd = &cobra.Command{
	Use:   "snapraid",
	Short: "List SnapRAID configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		// GetApiSnapraid requires a body (UtilsSnapRAIDConfig)
		resp, err := c.GetApiSnapraid(context.Background(), cosmossdk.GetApiSnapraidJSONRequestBody{})
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var storageSnapraidRunCmd = &cobra.Command{
	Use:   "snapraid-run <name> <action>",
	Short: "Run a SnapRAID action (sync|scrub|fix|enable|disable|status)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		// action must be typed as GetApiSnapraidNameActionParamsAction
		action := cosmossdk.GetApiSnapraidNameActionParamsAction(args[1])
		resp, err := c.GetApiSnapraidNameAction(context.Background(), args[0], action)
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var storageSnapraidDeleteCmd = &cobra.Command{
	Use:   "snapraid-delete <name>",
	Short: "Delete a SnapRAID configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !storageSnapraidDeleteYes {
			fmt.Printf("Delete SnapRAID config %q? [y/N]: ", args[0])
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
		resp, err := c.DeleteApiSnapraidName(context.Background(), args[0], cosmossdk.DeleteApiSnapraidNameJSONRequestBody{})
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("SnapRAID config %q deleted.", args[0])
		return nil
	},
}

var (
	storagePath              string
	storageSnapraidDeleteYes bool
)

func init() {
	storageListDirCmd.Flags().StringVar(&storagePath, "path", "/", "directory path")
	storageUnmountCmd.Flags().StringVar(&storagePath, "path", "", "mount path to unmount (required)")
	_ = storageUnmountCmd.MarkFlagRequired("path")
	storageNewDirCmd.Flags().StringVar(&storagePath, "path", "", "directory path to create (required)")
	_ = storageNewDirCmd.MarkFlagRequired("path")
	storageSnapraidDeleteCmd.Flags().BoolVar(&storageSnapraidDeleteYes, "yes", false, "skip confirmation")
	storageCmd.AddCommand(
		storageDisksCmd,
		storageMountsCmd,
		storageListDirCmd,
		storageUnmountCmd,
		storageNewDirCmd,
		storageSnapraidCmd,
		storageSnapraidRunCmd,
		storageSnapraidDeleteCmd,
	)
}
