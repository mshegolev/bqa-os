package app

import (
	"github.com/mshegolev/bqa-os/internal/runtime"
	"github.com/spf13/cobra"
)

func claudeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claude",
		Short: "Prepare BQA Master Agent context for Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runtime.Prepare("claude")
		},
	}
}
