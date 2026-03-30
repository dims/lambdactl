package cmd

import (
	"fmt"
	"os"

	"github.com/dims/lambdactl/api"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	client     *api.Client
	version    = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "lambdactl",
	Short:   "CLI for Lambda Labs cloud GPU instances",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !shouldInitClient(cmd) {
			return nil
		}
		var err error
		client, err = api.NewClient()
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := client.ListInstances()
		if err != nil {
			return fmt.Errorf("API key invalid: %w", err)
		}
		if jsonOutput {
			return outputJSON(struct {
				Valid bool `json:"valid"`
			}{Valid: true})
		}
		fmt.Println("API key is valid.")
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
