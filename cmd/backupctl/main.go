package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"backupctl/internal/publish"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "backupctl",
		Short:         "Publish and restore snapshots",
		SilenceErrors: true,
		// Runs after argument validation, so bad arguments still print usage.
		PersistentPreRun: func(cmd *cobra.Command, _ []string) { cmd.SilenceUsage = true },
	}
	root.AddCommand(&cobra.Command{
		Use:   "backup",
		Short: "Publish snapshot.txt under an exclusive lock",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := publish.Publish("snapshot.txt", []byte("snapshot-version=4\n")); err != nil {
				return fmt.Errorf("publish snapshot: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Snapshot published under exclusive lock")
			return nil
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "restore <destination>",
		Short: "Copy snapshot.txt to a destination",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source, err := filepath.Abs("snapshot.txt")
			if err != nil {
				return fmt.Errorf("resolve snapshot path: %w", err)
			}
			target, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("resolve destination path: %w", err)
			}
			if source == target {
				return fmt.Errorf("source and destination must differ")
			}
			return runChild(cmd, exec.CommandContext(cmd.Context(), "cp", source, target))
		},
	})
	return root
}

func runChild(cmd *cobra.Command, child *exec.Cmd) error {
	child.Stdout = cmd.OutOrStdout()
	child.Stderr = cmd.ErrOrStderr()
	if err := child.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Command completed")
	return nil
}
