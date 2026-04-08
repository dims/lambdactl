package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dims/lambdactl/api"
	"github.com/spf13/cobra"
)

func outputJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func statusWriter() io.Writer {
	if jsonOutput {
		return os.Stderr
	}
	return os.Stdout
}

func statusf(format string, args ...any) {
	fmt.Fprintf(statusWriter(), format, args...)
}

func shouldInitClient(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "completion", "help":
			return false
		}
	}
	return true
}

func printInstanceSummary(inst *api.Instance) {
	statusf("\n")
	statusf("  Instance:  %s", inst.ID)
	if inst.Name != "" {
		statusf(" (%s)", inst.Name)
	}
	statusf("\n")
	if inst.Type != nil {
		statusf("  GPU:       %s  [%s]\n", inst.Type.Description, inst.Type.Name)
		if inst.Type.GPUDescription != "" {
			statusf("  GPU Info:  %s\n", inst.Type.GPUDescription)
		}
		statusf("  Specs:     %d vCPUs, %d GiB RAM, %d GiB storage\n",
			inst.Type.Specs.VCPUs, inst.Type.Specs.MemoryGiB, inst.Type.Specs.StorageGiB)
	}
	if inst.Region != nil {
		statusf("  Region:    %s (%s)\n", inst.Region.Name, inst.Region.Description)
	}
	statusf("  IP:        %s\n\n", inst.IP)
}
