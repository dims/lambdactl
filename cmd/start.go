package cmd

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/dims/lambdactl/api"
	"github.com/spf13/cobra"
)

func init() {
	var gpu, sshKey, name, region string
	var retries, retryDelay int
	var waitSSH bool
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Launch a GPU instance",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var lastErr error
			for attempt := 0; attempt <= retries; attempt++ {
				if attempt > 0 {
					statusf("  attempt %d/%d failed: %v\n", attempt, retries+1, lastErr)
					statusf("  retrying in %ds...\n", retryDelay)
					time.Sleep(time.Duration(retryDelay) * time.Second)
				}

				r := region
				if r == "" {
					var err error
					r, err = pickRegion(client, gpu)
					if err != nil {
						lastErr = err
						continue
					}
				}

				id, err := client.Launch(gpu, sshKey, name, r)
				if err != nil {
					lastErr = err
					continue
				}

				statusf("Launched instance %s in %s. Waiting for IP...\n", id, r)
				inst, err := waitForIP(client, id)
				if err != nil {
					return err
				}

				if waitSSH {
					statusf("Waiting for SSH on %s...\n", inst.IP)
					if err := waitForSSH(inst.IP); err != nil {
						return err
					}
				}

				if jsonOutput {
					return outputJSON(inst)
				}
				fmt.Printf("Ready! ssh ubuntu@%s\n", inst.IP)
				return nil
			}
			return fmt.Errorf("failed after %d attempt(s): %w", retries+1, lastErr)
		},
	}
	cmd.Flags().StringVarP(&gpu, "gpu", "g", "", "GPU instance type (required)")
	cmd.Flags().StringVarP(&sshKey, "ssh", "s", "", "SSH key name (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "instance name")
	cmd.Flags().StringVarP(&region, "region", "r", "", "region (auto-selects if omitted)")
	cmd.Flags().IntVar(&retries, "retries", 0, "number of retry attempts if GPU unavailable")
	cmd.Flags().IntVar(&retryDelay, "retry-delay", 60, "seconds to wait between retries")
	cmd.Flags().BoolVar(&waitSSH, "wait-ssh", false, "wait for SSH to become available after launch")
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

func waitForIP(c *api.Client, id string) (*api.Instance, error) {
	deadline := time.After(5 * time.Minute)
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-deadline:
			return nil, fmt.Errorf("timed out waiting for IP. Check status with: lambdactl instances")
		case <-tick.C:
			inst, err := c.GetInstance(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "poll error: %v\n", err)
				continue
			}
			switch inst.Status {
			case "terminated", "unhealthy":
				return nil, fmt.Errorf("instance %s is %s", id, inst.Status)
			}
			if inst.IP != "" {
				return inst, nil
			}
			statusf("  status: %s\n", inst.Status)
		}
	}
}

func waitForSSH(ip string) error {
	deadline := time.After(5 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-deadline:
			return fmt.Errorf("timed out waiting for SSH on %s", ip)
		case <-tick.C:
			conn, err := net.DialTimeout("tcp", ip+":22", 5*time.Second)
			if err == nil {
				conn.Close()
				statusf("SSH is ready.\n")
				return nil
			}
			statusf("  waiting for SSH...\n")
		}
	}
}
