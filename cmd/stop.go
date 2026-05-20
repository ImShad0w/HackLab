package cmd

import (
	"hacklab/internal/docker"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <lab-name>",
	Short: "Stop a running lab",
	Long:  `Stop and remove a lab's containers.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return docker.Stop(args[0])
	},
}
