package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "firewall",
		Short: "List inbound firewall rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			rules, err := client.ListFirewallRules()
			if err != nil {
				return err
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(rules)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "PROTOCOL\tPORTS\tSOURCE\tDESCRIPTION")
			for _, r := range rules {
				ports := "all"
				if r.PortRange[0] > 0 {
					if r.PortRange[0] == r.PortRange[1] {
						ports = fmt.Sprintf("%d", r.PortRange[0])
					} else {
						ports = fmt.Sprintf("%d-%d", r.PortRange[0], r.PortRange[1])
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Protocol, ports, r.SourceNetwork, r.Description)
			}
			return w.Flush()
		},
	})
}
