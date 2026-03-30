package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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
