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
	var family, region string
	cmd := &cobra.Command{
		Use:   "images",
		Short: "List available OS images",
		RunE: func(cmd *cobra.Command, args []string) error {
			images, err := client.ListImages()
			if err != nil {
				return err
			}
			// filter
			filtered := images[:0]
			for _, img := range images {
				if family != "" && img.Family != family {
					continue
				}
				if region != "" && img.Region.Name != region {
					continue
				}
				filtered = append(filtered, img)
			}
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(filtered)
			}
			// dedupe by family -- show latest version per family with region count
			type summary struct {
				family  string
				name    string
				version string
				regions map[string]bool
			}
			byFamily := map[string]*summary{}
			for _, img := range filtered {
				s, ok := byFamily[img.Family]
				if !ok {
					s = &summary{family: img.Family, name: img.Name, version: img.Version, regions: map[string]bool{}}
					byFamily[img.Family] = s
				}
				if img.Version > s.version {
					s.version = img.Version
				}
				s.regions[img.Region.Name] = true
			}
			rows := make([]*summary, 0, len(byFamily))
			for _, s := range byFamily {
				rows = append(rows, s)
			}
			sort.Slice(rows, func(i, j int) bool { return rows[i].family < rows[j].family })
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "FAMILY\tNAME\tVERSION\tREGIONS")
			for _, r := range rows {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", r.family, r.name, r.version, len(r.regions))
			}
			return w.Flush()
		},
	}
	cmd.Flags().StringVar(&family, "family", "", "filter by image family")
	cmd.Flags().StringVar(&region, "region", "", "filter by region")
	rootCmd.AddCommand(cmd)
}

// imagesByRegion is used by other commands to resolve image IDs
func imagesByRegion(images []api.Image, family, region string) []api.Image {
	var out []api.Image
	for _, img := range images {
		if img.Family == family && img.Region.Name == region {
			out = append(out, img)
		}
	}
	return out
}
