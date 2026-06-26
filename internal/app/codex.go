package app

import (
	"github.com/mshegolev/bqa-os/internal/runtime"
	"github.com/spf13/cobra"
)

func codexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "codex",
		Short: "Prepare BQA Master Agent context for Codex CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runtime.Prepare("codex")
		},
	}
}
