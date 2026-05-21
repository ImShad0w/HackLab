package cmd

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"hacklab/internal/lab"
	"hacklab/internal/store"

	"github.com/spf13/cobra"
)

const (
	slugCol = 22
	nameCol = 38
	diffCol = 14
	objCol  = 15
	typeCol = 17
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

		// Top border
		printTop()

		// Header
		printRow("  SLUG", "LAB NAME", "DIFFICULTY", "OBJECTIVES", "TYPE")
		printSep()

		// Data rows
		for i, l := range labs {
			mf := l.Manifest
			objText := fmt.Sprintf("%d objectives", len(mf.Objectives))
			if len(mf.Objectives) == 1 {
				objText = "1 objective"
			}

			typeText := "docker-compose"
			if mf.ComposeFile == "" {
				typeText = "single container"
			}

			// Truncate long values
			slug := truncate(l.Name, slugCol)
			name := truncate(mf.Name, nameCol)
			diff := truncate(mf.Difficulty, diffCol)

			printRow(slug, name, diff, objText, typeText)

			// Description as sub-line spanning full row
			if mf.Description != "" {
				printDesc("  " + truncate(mf.Description, fullDescW()))
			}

			// Add separator between rows, not after last
			if i < len(labs)-1 {
				printSep()
			}
		}

		// Bottom border
		printBottom()

		fmt.Println()
		fmt.Println("  start a lab with: hacklab start <name>")
		fmt.Println()
		return nil
	},
}

func printTop() {
	fmt.Printf("  ┌%s┬%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", slugCol),
		strings.Repeat("─", nameCol),
		strings.Repeat("─", diffCol),
		strings.Repeat("─", objCol),
		strings.Repeat("─", typeCol),
	)
}

func printSep() {
	fmt.Printf("  ├%s┼%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", slugCol),
		strings.Repeat("─", nameCol),
		strings.Repeat("─", diffCol),
		strings.Repeat("─", objCol),
		strings.Repeat("─", typeCol),
	)
}

func printBottom() {
	fmt.Printf("  └%s┴%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", slugCol),
		strings.Repeat("─", nameCol),
		strings.Repeat("─", diffCol),
		strings.Repeat("─", objCol),
		strings.Repeat("─", typeCol),
	)
}

func printRow(slug, name, diff, obj, typ string) {
	fmt.Printf("  │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
		slugCol, slug,
		nameCol, name,
		diffCol, diff,
		objCol, obj,
		typeCol, typ,
	)
}

func printDesc(desc string) {
	fmt.Printf("  │ %-*s │\n", fullDescW(), desc)
}

func fullDescW() int {
	return slugCol + nameCol + diffCol + objCol + typeCol + 8 // 4 inner borders (│ │ │ │)
}

func truncate(s string, maxW int) string {
	if s == "" {
		return "—"
	}
	rw := utf8.RuneCountInString(s)
	if rw <= maxW {
		return s
	}
	// Build truncated string with ellipsis, staying within max rune width
	var b strings.Builder
	count := 0
	for _, r := range s {
		if count+1 >= maxW {
			break
		}
		b.WriteRune(r)
		count++
	}
	return b.String() + "…"
}
