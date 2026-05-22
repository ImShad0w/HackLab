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
	Short: "Add a lab from git repo, subdirectory, or local path",
	Long: `Add a lab to your collection.

Examples:
  # Full repo as a lab
  hacklab add https://github.com/user/my-hacklab

  # Specific subdirectory (repo#path)
  hacklab add https://github.com/HackLab-cli/lab-examples#labs/juice-shop

  # Local folder
  hacklab add ./my-labs/sqli-lab
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]

		if err := store.Ensure(); err != nil {
			return err
		}

		// Parse #subdir syntax early, before deciding git vs local
		sourcePath := source
		var subdir string
		if idx := strings.Index(source, "#"); idx >= 0 {
			sourcePath = source[:idx]
			subdir = strings.TrimPrefix(source[idx+1:], "/")
		}

		isGit := strings.HasPrefix(sourcePath, "http://") ||
			strings.HasPrefix(sourcePath, "https://") ||
			strings.HasPrefix(sourcePath, "git@")

		if isGit {
			return addFromGit(sourcePath, subdir)
		}
		return addFromLocal(sourcePath, subdir)
	},
}

// addFromGit handles git URLs, with optional subdir
func addFromGit(repoURL, subdir string) error {

	// Determine lab name
	var name string
	if subdir != "" {
		// Name comes from the subdirectory (last segment)
		parts := strings.Split(strings.TrimSuffix(subdir, "/"), "/")
		name = parts[len(parts)-1]
	} else {
		// Name comes from the repo URL
		parts := strings.Split(strings.TrimSuffix(repoURL, "/"), "/")
		rawName := parts[len(parts)-1]
		name = strings.TrimSuffix(rawName, ".git")
	}

	labsDir, err := store.LabsDir()
	if err != nil {
		return err
	}
	destPath := filepath.Join(labsDir, name)

	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("lab '%s' already exists — remove it first", name)
	}

	if subdir == "" {
		// Simple case: clone entire repo as the lab
		fmt.Printf("  📦 cloning %s ...\n", repoURL)
		cloneCmd := exec.Command("git", "clone", "--depth", "1", repoURL, destPath)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	} else {
		// Subdirectory case: clone to temp, extract subdir, clean up
		tmpDir, err := os.MkdirTemp("", "hacklab-clone-")
		if err != nil {
			return fmt.Errorf("creating temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		fmt.Printf("  📦 cloning %s ...\n", repoURL)
		cloneCmd := exec.Command("git", "clone", "--depth", "1", repoURL, tmpDir)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}

		subPath := filepath.Join(tmpDir, subdir)
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			return fmt.Errorf("subdirectory '%s' not found in repo", subdir)
		}

		fmt.Printf("  📁 extracting lab from %s#%s ...\n", repoURL, subdir)
		if err := copyDir(subPath, destPath); err != nil {
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
}

func addFromLocal(path, subdir string) error {
	src := path
	if subdir != "" {
		src = filepath.Join(path, subdir)
	}

	absPath, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	name := filepath.Base(absPath)

	labsDir, err := store.LabsDir()
	if err != nil {
		return err
	}
	destPath := filepath.Join(labsDir, name)

	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("lab '%s' already exists — remove it first", name)
	}

	if err := copyDir(absPath, destPath); err != nil {
		return fmt.Errorf("copying lab: %w", err)
	}

	fmt.Println()
	fmt.Printf("  ⚡  hacklab: lab added\n")
	fmt.Printf("  🎯 name:     %s\n", name)
	fmt.Printf("  📁 location: %s\n", destPath)
	fmt.Println()
	fmt.Printf("  start it with: hacklab start %s\n", name)
	fmt.Println()
	return nil
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
