package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/dims/lambdactl/api"
	"github.com/spf13/cobra"
)

func init() {
	var gpu, sshKey, name, region string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch a GPU instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if region == "" {
				var err error
				region, err = pickRegion(client, gpu)
				if err != nil {
					return err
				}
			}
			id, err := client.Launch(gpu, sshKey, name, region)
			if err != nil {
				return err
			}
			statusf("Launched instance %s in %s. Waiting for IP...\n", id, region)
			return waitForIP(client, id)
		},
	}
	cmd.Flags().StringVarP(&gpu, "gpu", "g", "", "GPU instance type (required)")
	cmd.Flags().StringVarP(&sshKey, "ssh", "s", "", "SSH key name (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "instance name")
	cmd.Flags().StringVarP(&region, "region", "r", "", "region (auto-selects if omitted)")
	cmd.MarkFlagRequired("gpu")
	cmd.MarkFlagRequired("ssh")
	rootCmd.AddCommand(cmd)
}

func pickRegion(c *api.Client, gpu string) (string, error) {
	types, err := c.ListInstanceTypes()
	if err != nil {
		return "", err
	}
	entry, ok := types[gpu]
	if !ok {
		return "", fmt.Errorf("unknown GPU type: %s", gpu)
	}
	if len(entry.Regions) == 0 {
		return "", fmt.Errorf("no regions available for %s", gpu)
	}
	return entry.Regions[0].Name, nil
}

func waitForIP(c *api.Client, id string) error {
	deadline := time.After(5 * time.Minute)
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-deadline:
			return fmt.Errorf("timed out waiting for IP. Check status with: lambdactl instances")
		case <-tick.C:
			inst, err := c.GetInstance(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "poll error: %v\n", err)
				continue
			}
			switch inst.Status {
			case "terminated", "unhealthy":
				return fmt.Errorf("instance %s is %s", id, inst.Status)
			}
			if inst.IP != "" {
				if jsonOutput {
					return outputJSON(inst)
				}
				fmt.Printf("Ready! ssh ubuntu@%s\n", inst.IP)
				return nil
			}
			statusf("  status: %s\n", inst.Status)
		}
	}
}
