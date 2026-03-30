package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	var yes bool
	cmd := &cobra.Command{
		Use:   "stop <instance-id-or-name>",
		Short: "Terminate an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inst, err := resolveInstanceRef(client, args[0])
			if err != nil {
				return err
			}
			if !yes {
				fmt.Printf("Terminate instance %s? [y/N] ", describeInstance(inst))
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(answer)) != "y" {
					fmt.Println("Aborted.")
					return nil
				}
			}
			if err := client.Terminate(inst.ID); err != nil {
				return err
			}
			fmt.Printf("Instance %s terminated.\n", describeInstance(inst))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	rootCmd.AddCommand(cmd)
}
