package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	var gpu, sshKey, name, region string
	var interval int
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Poll until a GPU type is available, then launch",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusf("Watching for %s availability (every %ds)...\n", gpu, interval)
			pollInterval := time.Duration(interval) * time.Second
			for {
				types, err := client.ListInstanceTypes()
				if err != nil {
					statusf("  [%s] error: %v\n", time.Now().Format("15:04:05"), err)
					time.Sleep(pollInterval)
					continue
				}
				entry, ok := types[gpu]
				if !ok {
					return fmt.Errorf("unknown GPU type: %s", gpu)
				}
				if len(entry.Regions) == 0 {
					statusf("  [%s] no availability\n", time.Now().Format("15:04:05"))
					time.Sleep(pollInterval)
					continue
				}
				target := entry.Regions[0].Name
				if region != "" {
					found := false
					for _, r := range entry.Regions {
						if r.Name == region {
							target = region
							found = true
							break
						}
					}
					if !found {
						statusf("  [%s] available but not in %s\n", time.Now().Format("15:04:05"), region)
						time.Sleep(pollInterval)
						continue
					}
				}
				statusf("  [%s] found in %s! Launching...\n", time.Now().Format("15:04:05"), target)
				id, err := client.Launch(gpu, sshKey, name, target)
				if err != nil {
					return err
				}
				statusf("Launched instance %s. Waiting for IP...\n", id)
				return waitForIP(client, id)
			}
		},
	}
	cmd.Flags().StringVarP(&gpu, "gpu", "g", "", "GPU instance type (required)")
	cmd.Flags().StringVarP(&sshKey, "ssh", "s", "", "SSH key name (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "instance name")
	cmd.Flags().StringVarP(&region, "region", "r", "", "only launch in this region")
	cmd.Flags().IntVar(&interval, "interval", 10, "poll interval in seconds")
	cmd.MarkFlagRequired("gpu")
	cmd.MarkFlagRequired("ssh")
	rootCmd.AddCommand(cmd)
}
