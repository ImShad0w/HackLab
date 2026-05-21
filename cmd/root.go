package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "hacklab",
	Short: "Your terminal hacking playground",
	Long: `
██╗  ██╗ █████╗  ██████╗██╗  ██╗██╗      █████╗ ██████╗
██║  ██║██╔══██╗██╔════╝██║ ██╔╝██║     ██╔══██╗██╔══██╗
███████║███████║██║     █████╔╝ ██║     ███████║██████╔╝
██╔══██║██╔══██║██║     ██╔═██╗ ██║     ██╔══██║██╔══██╗
██║  ██║██║  ██║╚██████╗██║  ██╗███████╗██║  ██║██████╔╝
╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═════╝

 Your terminal hacking playground.
 Spin up vulnerable labs, exploit them, level up.
`,
	Version: version,
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(statusCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
