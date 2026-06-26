package app

import (
	"github.com/mshegolev/bqa-os/internal/version"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print BQA-OS version information",
		Run: func(cmd *cobra.Command, args []string) {
			version.Print()
		},
	}
}
