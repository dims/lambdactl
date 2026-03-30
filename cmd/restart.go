package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "restart <instance-id-or-name>",
		Short: "Restart an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inst, err := resolveInstanceRef(client, args[0])
			if err != nil {
				return err
			}
			if err := client.Restart(inst.ID); err != nil {
				return err
			}
			fmt.Printf("Instance %s restarting.\n", describeInstance(inst))
			return nil
		},
	})
}
