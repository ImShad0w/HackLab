package cmd

import (
	"fmt"

	"hacklab/internal/docker"
	"hacklab/internal/store"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <lab-name>",
	Short: "Remove a lab from your collection",
	Long:  `Remove a lab from ~/.hacklab/labs/. Running containers are also stopped.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		exists, err := store.LabExists(name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("lab '%s' not found", name)
		}

		// Stop any running containers first
		_ = docker.Stop(name)

		if err := store.RemoveLabDir(name); err != nil {
			return fmt.Errorf("removing lab: %w", err)
		}

		fmt.Printf("  ✅ lab '%s' removed\n\n", name)
		return nil
	},
}
