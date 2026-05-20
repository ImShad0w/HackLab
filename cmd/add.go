package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"hacklab/internal/store"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Add a lab from git repo or local path",
	Long: `Add a lab to your collection.

Examples:
  hacklab add https://github.com/user/hacklab-juice-shop
  hacklab add ./my-labs/sqli-lab
  hacklab add /home/user/repos/lab-web-attacks
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]

		if err := store.Ensure(); err != nil {
			return err
		}

		var name string
		var destPath string
		isGit := strings.HasPrefix(source, "http://") ||
			strings.HasPrefix(source, "https://") ||
			strings.HasPrefix(source, "git@")

		if isGit {
			parts := strings.Split(strings.TrimSuffix(source, "/"), "/")
			rawName := parts[len(parts)-1]
			name = strings.TrimSuffix(rawName, ".git")

			labsDir, err := store.LabsDir()
			if err != nil {
				return err
			}
			destPath = filepath.Join(labsDir, name)

			if _, err := os.Stat(destPath); err == nil {
				return fmt.Errorf("lab '%s' already exists — remove it first", name)
			}

			fmt.Printf("  📦 cloning %s ...\n", source)
			cloneCmd := exec.Command("git", "clone", "--depth", "1", source, destPath)
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			if err := cloneCmd.Run(); err != nil {
				return fmt.Errorf("git clone failed: %w", err)
			}
		} else {
			absPath, err := filepath.Abs(source)
			if err != nil {
				return fmt.Errorf("resolving path: %w", err)
			}
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", absPath)
			}

			name = filepath.Base(absPath)

			labsDir, err := store.LabsDir()
			if err != nil {
				return err
			}
			destPath = filepath.Join(labsDir, name)

			if _, err := os.Stat(destPath); err == nil {
				return fmt.Errorf("lab '%s' already exists — remove it first", name)
			}

			if err := copyDir(absPath, destPath); err != nil {
				return fmt.Errorf("copying lab: %w", err)
			}
		}

		fmt.Println()
		fmt.Printf("  ⚡  hacklab: lab added\n")
		fmt.Printf("  🎯 name:     %s\n", name)
		fmt.Printf("  📁 location: %s\n", destPath)
		fmt.Println()
		fmt.Printf("  start it with: hacklab start %s\n", name)
		fmt.Println()
		return nil
	},
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
