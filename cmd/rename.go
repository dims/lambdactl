package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "rename <instance-id-or-name> <new-name>",
		Short: "Rename an instance",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inst, err := resolveInstanceRef(client, args[0])
			if err != nil {
				return err
			}

			updated, err := client.RenameInstance(inst.ID, args[1])
			if err != nil {
				return err
			}

			if jsonOutput {
				return outputJSON(updated)
			}

			fmt.Printf("Renamed instance %s to %q.\n", describeInstance(inst), updated.Name)
			return nil
		},
	})
}
