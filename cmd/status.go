package cmd

import (
	"fmt"

	"hacklab/internal/docker"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show running labs",
	Long:  `List all currently running hacklab containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		labs, err := docker.ListRunning()
		if err != nil {
			return err
		}

		if len(labs) == 0 {
			fmt.Println("\n  no labs running\n")
			return nil
		}

		fmt.Println()
		fmt.Printf("  ⚡  %d lab(s) running\n\n", len(labs))

		for _, name := range labs {
			fmt.Printf("  🟢 %s\n", name)
		}

		fmt.Println()
		fmt.Println("  stop a lab with: hacklab stop <name>")
		fmt.Println()
		return nil
	},
}
