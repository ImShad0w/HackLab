package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var openBrowser bool

var startCmd = &cobra.Command{
	Use:   "start <lab-name>",
	Short: "Start a hacking lab",
	Long:  `Spin up a lab environment and launch the interactive challenge session.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLabSession(args[0], false)
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
