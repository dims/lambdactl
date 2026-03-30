package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "ssh <instance-id-or-name>",
		Short: "SSH into an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inst, err := resolveInstanceRef(client, args[0])
			if err != nil {
				return err
			}
			if inst.IP == "" {
				return fmt.Errorf("instance %s has no IP (status: %s)", describeInstance(inst), inst.Status)
			}
			sshBin, err := exec.LookPath("ssh")
			if err != nil {
				return fmt.Errorf("ssh not found in PATH")
			}
			fmt.Fprintf(os.Stderr, "Connecting to ubuntu@%s...\n", inst.IP)
			return syscall.Exec(sshBin, []string{"ssh", "ubuntu@" + inst.IP}, os.Environ())
		},
	})
}
