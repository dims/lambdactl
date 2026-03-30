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
		Use:   "instances",
		Short: "List instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			instances, err := client.ListInstances()
			if err != nil {
				return err
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(instances)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSTATUS\tIP\tTYPE\tREGION")
			for _, inst := range instances {
				typeName, regionName := "", ""
				if inst.Type != nil {
					typeName = inst.Type.Name
				}
				if inst.Region != nil {
					regionName = inst.Region.Name
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					inst.ID, inst.Name, inst.Status, inst.IP, typeName, regionName)
			}
			return w.Flush()
		},
	})
}
