package cmd

import (
	"fmt"
	"os"

	"github.com/azukaar/cosmos-server/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cliVersion  string
	profileFlag string
	urlFlag     string
	tokenFlag   string
)

// Execute is the entrypoint called from main.
func Execute(v string) error {
	cliVersion = v
	rootCmd.Version = v
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "cosmos",
	Short: "CLI for Cosmos Server",
	Long:  "cosmos is the official command-line interface for Cosmos Server.\nManage containers, routes, users, backups, and more from your terminal.",
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profileFlag, "profile", "", "config profile to use (overrides current_profile)")
	rootCmd.PersistentFlags().StringVar(&urlFlag, "url", "", "server URL (overrides config and COSMOS_URL)")
	rootCmd.PersistentFlags().StringVar(&tokenFlag, "token", "", "API token (overrides keychain and COSMOS_TOKEN)")

	rootCmd.AddCommand(
		configureCmd,
		containersCmd,
		routesCmd,
		usersCmd,
		systemCmd,
		metricsCmd,
		backupsCmd,
		cronCmd,
		storageCmd,
		constellationCmd,
		apiTokensCmd,
		openIDCmd,
	)
}

func resolve() (*config.Resolved, error) {
	return config.Resolve(profileFlag, urlFlag, tokenFlag)
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
