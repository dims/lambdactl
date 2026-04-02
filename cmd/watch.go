package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/dims/lambdactl/api"
	"github.com/spf13/cobra"
)

func init() {
	var gpu, sshKey, name, region string
	var interval, timeout int
	var dryRun, waitSSH bool
	cmd := &cobra.Command{
		Use:          "watch",
		Short:        "Poll until a GPU type is available, then launch",
		SilenceUsage: true,
		Long: `Poll until a GPU type is available, then launch it.

If --gpu is specified, watch only that type. If omitted, watch for ANY
available GPU and launch the cheapest one found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if gpu != "" {
				statusf("Watching for %s availability (every %ds)...\n", gpu, interval)
			} else {
				statusf("Watching for any GPU availability (every %ds)...\n", interval)
			}
			pollInterval := time.Duration(interval) * time.Second

			var deadline <-chan time.Time
			if timeout > 0 {
				deadline = time.After(time.Duration(timeout) * time.Second)
				statusf("Timeout set to %ds.\n", timeout)
			}

			for {
				// Check timeout before each poll.
				if deadline != nil {
					select {
					case <-deadline:
						return fmt.Errorf("timed out after %ds waiting for GPU availability", timeout)
					default:
					}
				}

				types, err := client.ListInstanceTypes()
				if err != nil {
					statusf("  [%s] error: %v\n", time.Now().Format("15:04:05"), err)
					time.Sleep(pollInterval)
					continue
				}

				match, target, err := findAvailable(types, gpu, region)
				if err != nil {
					return err
				}
				if match == "" {
					if gpu != "" {
						statusf("  [%s] no availability\n", time.Now().Format("15:04:05"))
					} else {
						statusf("  [%s] nothing available\n", time.Now().Format("15:04:05"))
					}
					time.Sleep(pollInterval)
					continue
				}

				entry := types[match]
				statusf("  [%s] %s ($%.2f/hr) available in %s!",
					time.Now().Format("15:04:05"), match,
					float64(entry.Type.PriceCents)/100, target)

				if dryRun {
					statusf(" (dry-run, not launching)\n")
					return nil
				}

				statusf(" Launching...\n")
				id, err := client.Launch(match, sshKey, name, target)
				if err != nil {
					if !isRetryable(err) {
						return err
					}
					statusf("  launch failed (capacity): %v — continuing to watch...\n", err)
					time.Sleep(pollInterval)
					continue
				}
				statusf("Launched instance %s. Waiting for IP...\n", id)
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
		},
	}
	cmd.Flags().StringVarP(&gpu, "gpu", "g", "", "GPU instance type (omit to watch for any)")
	cmd.Flags().StringVarP(&sshKey, "ssh", "s", "", "SSH key name (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "instance name")
	cmd.Flags().StringVarP(&region, "region", "r", "", "only launch in this region")
	cmd.Flags().IntVar(&interval, "interval", 10, "poll interval in seconds")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "give up waiting for availability after this many seconds; does not cover launch/boot time (0 = no timeout)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report availability but do not launch")
	cmd.Flags().BoolVar(&waitSSH, "wait-ssh", false, "wait for SSH to become available after launch")
	cmd.MarkFlagRequired("ssh")
	rootCmd.AddCommand(cmd)
}

// findAvailable returns the instance type name and region to launch.
// If gpu is set, only that type is considered. Otherwise the cheapest
// available type is selected. Returns ("", "", nil) when nothing is available.
func findAvailable(types map[string]api.InstanceTypeEntry, gpu, region string) (string, string, error) {
	if gpu != "" {
		entry, ok := types[gpu]
		if !ok {
			return "", "", fmt.Errorf("unknown GPU type: %s", gpu)
		}
		target := matchRegion(entry, region)
		if target == "" {
			return "", "", nil
		}
		return gpu, target, nil
	}

	// Collect all types with availability, sorted by price (cheapest first).
	type candidate struct {
		name   string
		entry  api.InstanceTypeEntry
		region string
	}
	var candidates []candidate
	for name, entry := range types {
		if target := matchRegion(entry, region); target != "" {
			candidates = append(candidates, candidate{name, entry, target})
		}
	}
	if len(candidates) == 0 {
		return "", "", nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].entry.Type.PriceCents < candidates[j].entry.Type.PriceCents
	})
	c := candidates[0]
	return c.name, c.region, nil
}

// matchRegion returns the target region name if the entry has availability
// in the requested region (or any region if region is ""). Returns "" if none.
func matchRegion(entry api.InstanceTypeEntry, region string) string {
	if len(entry.Regions) == 0 {
		return ""
	}
	if region == "" {
		return entry.Regions[0].Name
	}
	for _, r := range entry.Regions {
		if r.Name == region {
			return region
		}
	}
	return ""
}
