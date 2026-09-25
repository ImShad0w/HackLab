package cmd

import (
	"fmt"
	"os"

	"hacklab/internal/docker"
	"hacklab/internal/lab"
	"hacklab/internal/progress"
	"hacklab/internal/session"
	"hacklab/internal/store"
	"hacklab/tui"

	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume your last lab session",
	Long: `Resume the most recent lab session.

If you quit the TUI mid-challenge — or the terminal was closed — 'hacklab resume'
re-launches the last lab, boots its containers again (if needed), and picks up
exactly where you left off.

Once a session has been wiped (via 'x' inside the TUI), there is nothing left
to resume.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := session.Load()
		if err != nil {
			return err
		}

		if sess == nil || sess.LabName == "" {
			fmt.Println()
			fmt.Println("  no session to resume")
			fmt.Println("  start a lab with: hacklab start <name>")
			fmt.Println()
			return nil
		}

		return runLabSession(sess.LabName, true)
	},
}

// runLabSession loads a lab, boots its containers (if needed), and launches
// the interactive TUI session. It is shared by `start` and `resume`.
//
// When resuming, already-running containers are reused instead of erroring.
// If the user wipes the session inside the TUI, containers are torn down and
// the saved session is cleared so it can no longer be resumed.
func runLabSession(labName string, resuming bool) error {
	labPath, err := store.LabPath(labName)
	if err != nil {
		return err
	}

	if _, err := os.Stat(labPath); os.IsNotExist(err) {
		if resuming {
			return fmt.Errorf("lab '%s' from your last session no longer exists — nothing to resume", labName)
		}
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

	// Start containers (reuse them when resuming)
	targetURL := ""
	if l.Manifest.Image != "" {
		if resuming {
			_, _, err = mgr.StartSingleIfNeeded(l.Manifest)
		} else {
			_, _, err = mgr.StartSingle(l.Manifest)
		}
		if err != nil {
			return err
		}
		targetURL = fmt.Sprintf("http://localhost:%d", l.Manifest.Port)
	} else if l.Manifest.ComposeFile != "" {
		if resuming {
			err = mgr.StartComposeIfNeeded(l.Manifest)
		} else {
			err = mgr.StartCompose(l.Manifest)
		}
		if err != nil {
			return err
		}
	}

	fmt.Println()
	if resuming {
		fmt.Printf("  ▶️  resuming session for '%s'\n", labName)
	} else {
		fmt.Printf("  ✅ Lab '%s' is running\n", labName)
	}
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
	result, err := tui.RunLab(l, p, targetURL)
	if err != nil {
		return err
	}

	// Session was wiped inside the TUI: tear down containers so an
	// abandoned target isn't left running for a session we deleted.
	if result.Wiped {
		_ = docker.Stop(labName)
		fmt.Printf("  🧹 session wiped — lab '%s' progress erased, can't be resumed\n\n", labName)
	}
	return nil
}
