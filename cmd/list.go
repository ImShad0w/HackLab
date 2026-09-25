package cmd

import (
	"fmt"
	"os"

	"hacklab/internal/lab"
	"hacklab/internal/store"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available labs",
	Long:  `List all labs installed locally with their name, difficulty, and container type.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		labsDir, err := store.LabsDir()
		if err != nil {
			return err
		}

		if _, err := os.Stat(labsDir); os.IsNotExist(err) {
			fmt.Println("\n  no labs found — add one with 'hacklab add <source>'")
			return nil
		}

		labs, err := lab.DiscoverLabs(labsDir)
		if err != nil {
			return err
		}

		if len(labs) == 0 {
			fmt.Println("\n  no labs found — add one with 'hacklab add <source>'")
			return nil
		}

		fmt.Println()
		fmt.Printf("  ⚡  hacklab: %d lab(s)\n\n", len(labs))

		for _, l := range labs {
			typeText := "docker-compose"
			if l.Manifest.ComposeFile == "" {
				typeText = "single container"
			}

			fmt.Printf("  %s (difficulty: %s, type: %s)\n",
				l.Name, l.Manifest.Difficulty, typeText)
		}

		fmt.Println()
		fmt.Println("  start a lab with: hacklab start <name>")
		fmt.Println()
		return nil
	},
}
