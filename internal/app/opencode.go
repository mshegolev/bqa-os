package app

import (
	"github.com/mshegolev/bqa-os/internal/runtime"
	"github.com/spf13/cobra"
)

func opencodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "opencode",
		Short: "Prepare BQA Master Agent context for OpenCode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runtime.Prepare("opencode")
		},
	}
}
