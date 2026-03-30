package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	sshKeysCmd := &cobra.Command{
		Use:   "ssh-keys",
		Short: "Manage SSH keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := client.ListSSHKeys()
			if err != nil {
				return err
			}
			if jsonOutput {
				return outputJSON(keys)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME")
			for _, k := range keys {
				fmt.Fprintf(w, "%s\t%s\n", k.ID, k.Name)
			}
			return w.Flush()
		},
	}

	addCmd := &cobra.Command{
		Use:   "add <name> [public-key-file]",
		Short: "Add an SSH key (omit file to generate a new key pair)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pubKey := ""
			if len(args) == 2 {
				b, err := os.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("reading public key file: %w", err)
				}
				pubKey = string(b)
			}
			key, err := client.AddSSHKey(args[0], pubKey)
			if err != nil {
				return err
			}
			if jsonOutput {
				return outputJSON(key)
			}
			fmt.Printf("Added SSH key %q (%s)\n", key.Name, key.ID)
			if pubKey == "" {
				fmt.Println("Generated private key (save this, it won't be shown again):")
				fmt.Println(key.PrivateKey)
			}
			return nil
		},
	}

	rmCmd := &cobra.Command{
		Use:   "rm <key-id>",
		Short: "Delete an SSH key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.DeleteSSHKey(args[0]); err != nil {
				return err
			}
			fmt.Printf("SSH key %s deleted.\n", args[0])
			return nil
		},
	}

	sshKeysCmd.AddCommand(addCmd, rmCmd)
	rootCmd.AddCommand(sshKeysCmd)
}
