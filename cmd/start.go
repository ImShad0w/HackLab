package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"hacklab/internal/docker"
	"hacklab/internal/lab"
	"hacklab/internal/progress"
	"hacklab/internal/store"
	"hacklab/tui"

	"github.com/spf13/cobra"
)

var openBrowser bool

var startCmd = &cobra.Command{
	Use:   "start <lab-name>",
	Short: "Start a hacking lab",
	Long:  `Spin up a lab environment and launch the interactive challenge session.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		labName := args[0]

		labPath, err := store.LabPath(labName)
		if err != nil {
			return err
		}

		if _, err := os.Stat(labPath); os.IsNotExist(err) {
			return fmt.Errorf("lab '%s' not found — add it with 'hacklab add <source>'", labName)
		}

		// Load lab manifest
		l, err := lab.LoadLab(labPath)
		if err != nil {
			return fmt.Errorf("loading lab: %w", err)
		}

		// Load progress
		p, err := progress.Load()
		if err != nil {
			return err
		}

		// Initialize Docker
		mgr, err := docker.NewManager(labPath, labName)
		if err != nil {
			return err
		}

		// Start containers
		targetURL := ""

		if l.Manifest.Image != "" {
			_, _, err = mgr.StartSingle(l.Manifest)
			if err != nil {
				return err
			}
			targetURL = fmt.Sprintf("http://localhost:%d", l.Manifest.Port)
		} else if l.Manifest.ComposeFile != "" {
			if err := mgr.StartCompose(l.Manifest); err != nil {
				return err
			}
		}

		fmt.Println()
		fmt.Printf("  ✅ Lab '%s' is running\n", labName)
		if targetURL != "" {
			fmt.Printf("  📡 Target: %s\n", targetURL)
		}
		fmt.Println()

		// Wait for readiness if configured
		if err := mgr.WaitForReady(l.Manifest.WaitFor, l.Manifest.WaitSecs); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠️  %v\n", err)
		} else if l.Manifest.WaitFor != "" {
			fmt.Printf("  ✅ Lab is ready\n\n")
		}

		// Handle browser opening
		if l.Manifest.OpenBrowser || openBrowser {
			openInBrowser(targetURL)
		}

		// Launch TUI
		return tui.RunLab(l, p, targetURL)
	},
}

func init() {
	startCmd.Flags().BoolVarP(&openBrowser, "browser", "b", false, "Open lab in browser")
}

func openInBrowser(url string) {
	var cmd *exec.Cmd
	switch {
	case os.PathSeparator == '\\':
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()
}
