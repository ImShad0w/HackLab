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
	Long:  `List all labs installed locally.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		labsDir, err := store.LabsDir()
		if err != nil {
			return err
		}

		if _, err := os.Stat(labsDir); os.IsNotExist(err) {
			fmt.Println("\n  no labs found — add one with 'hacklab add <source>'\n")
			return nil
		}

		labs, err := lab.DiscoverLabs(labsDir)
		if err != nil {
			return err
		}

		if len(labs) == 0 {
			fmt.Println("\n  no labs found — add one with 'hacklab add <source>'\n")
			return nil
		}

		fmt.Println()
		fmt.Printf("  ⚡  hacklab: %d lab(s)\n\n", len(labs))

		for i, l := range labs {
			mf := l.Manifest

			// Determine lab type
			typeLabel := "single container"
			if mf.ComposeFile != "" {
				typeLabel = "multi container"
			}

			// Slug and type header
			fmt.Printf("  %s", l.Name)
			fmt.Printf("  [%s]\n", typeLabel)

			// Description
			if mf.Description != "" {
				fmt.Printf("  %s\n", mf.Description)
			} else {
				fmt.Printf("  no description\n")
			}

			// Separator between labs
			if i < len(labs)-1 {
				fmt.Println()
			}
		}

		fmt.Println()
		fmt.Println("  start a lab with: hacklab start <name>")
		fmt.Println()
		return nil
	},
}

