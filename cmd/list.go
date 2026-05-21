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
	Long:  `List all labs installed locally with difficulty and objective counts.`,
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

		for _, l := range labs {
			mf := l.Manifest
			objCount := len(mf.Objectives)
			objWord := "objectives"
			if objCount == 1 {
				objWord = "objective"
			}

			container := "single container"
			if mf.ComposeFile != "" {
				container = "docker-compose"
			}

			// Slug (directory name) is what you actually type with 'hacklab start'
			fmt.Printf("  🎯 %-18s  %-34s  %s  ·  %d %s  ·  %s\n",
				l.Name,
				mf.Name,
				mf.Difficulty,
				objCount, objWord,
				container,
			)
			if mf.Description != "" {
				fmt.Printf("     %s\n", mf.Description)
			}
		}

		fmt.Println()
		fmt.Println("  start a lab with: hacklab start <name>")
		fmt.Println()
		return nil
	},
}
