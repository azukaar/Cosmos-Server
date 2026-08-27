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

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Manage scheduled CRON jobs",
}

var cronListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CRON configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiCron(context.Background(), cosmossdk.GetApiCronJSONRequestBody{})
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var cronJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "List all scheduled jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiJobs(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		var jobs map[string]map[string]interface{}
		if err := json.Unmarshal(ar.Data, &jobs); err != nil {
			return output.JSON(ar.Data)
		}
		rows := [][]string{}
		for group, groupJobs := range jobs {
			for name := range groupJobs {
				rows = append(rows, []string{group, name})
			}
		}
		output.Table([]string{"GROUP", "NAME"}, rows)
		return nil
	},
}

var cronJobsRunningCmd = &cobra.Command{
	Use:   "jobs-running",
	Short: "List currently running jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		resp, err := c.GetApiJobsRunning(context.Background())
		if err != nil { return err }
		ar, err := client.Parse(resp)
		if err != nil { return err }
		return output.JSON(ar.Data)
	},
}

var cronRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Trigger a job manually",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		body := cosmossdk.PostApiJobsRunJSONRequestBody{Name: args[0]}
		resp, err := c.PostApiJobsRun(context.Background(), body)
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Job %q triggered.", args[0])
		return nil
	},
}

var cronStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := config.Resolve(profileFlag, urlFlag, tokenFlag)
		if err != nil { return err }
		c, err := client.New(r)
		if err != nil { return err }
		body := cosmossdk.PostApiJobsStopJSONRequestBody{Name: args[0]}
		resp, err := c.PostApiJobsStop(context.Background(), body)
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Job %q stopped.", args[0])
		return nil
	},
}

var cronDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a CRON configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cronDeleteYes {
			fmt.Printf("Delete cron %q? [y/N]: ", args[0])
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
		resp, err := c.DeleteApiCronName(context.Background(), args[0], cosmossdk.DeleteApiCronNameJSONRequestBody{Name: args[0]})
		if err != nil { return err }
		_, err = client.Parse(resp)
		if err != nil { return err }
		output.Success("Cron %q deleted.", args[0])
		return nil
	},
}

var cronDeleteYes bool

func init() {
	cronDeleteCmd.Flags().BoolVar(&cronDeleteYes, "yes", false, "skip confirmation")
	cronCmd.AddCommand(cronListCmd, cronJobsCmd, cronJobsRunningCmd, cronRunCmd, cronStopCmd, cronDeleteCmd)
}
