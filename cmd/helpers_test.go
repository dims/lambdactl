package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestShouldInitClient(t *testing.T) {
	root := &cobra.Command{Use: "lambdactl"}
	completion := &cobra.Command{Use: "completion"}
	bash := &cobra.Command{Use: "bash"}
	help := &cobra.Command{Use: "help"}
	instances := &cobra.Command{Use: "instances"}

	root.AddCommand(completion, help, instances)
	completion.AddCommand(bash)

	if shouldInitClient(bash) {
		t.Fatalf("completion subcommands should not require client initialization")
	}
	if shouldInitClient(help) {
		t.Fatalf("help should not require client initialization")
	}
	if !shouldInitClient(instances) {
		t.Fatalf("normal commands should require client initialization")
	}
}
