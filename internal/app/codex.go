package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func codexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "codex",
		Short: "Prepare BQA Master Agent context for Codex CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("TODO: generate .bqa/prompts/bqa-master-context.md and launch or instruct Codex CLI")
			return nil
		},
	}
}
