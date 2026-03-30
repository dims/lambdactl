package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/dims/lambdactl/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "types",
		Short: "List available GPU instance types",
		RunE: func(cmd *cobra.Command, args []string) error {
			types, err := client.ListInstanceTypes()
			if err != nil {
				return err
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(types)
			}
			type row struct {
				name  string
				entry api.InstanceTypeEntry
			}
			rows := make([]row, 0, len(types))
			for name, entry := range types {
				rows = append(rows, row{name, entry})
			}
			sort.Slice(rows, func(i, j int) bool {
				return rows[i].entry.Type.PriceCents < rows[j].entry.Type.PriceCents
			})
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDESCRIPTION\t$/HR\tVCPUS\tRAM\tDISK\tREGIONS")
			for _, r := range rows {
				t := r.entry.Type
				fmt.Fprintf(w, "%s\t%s\t$%.2f\t%d\t%dGB\t%dGB\t%d available\n",
					r.name, t.Description, float64(t.PriceCents)/100,
					t.Specs.VCPUs, t.Specs.MemoryGiB, t.Specs.StorageGiB, len(r.entry.Regions))
			}
			return w.Flush()
		},
	})
}
