package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func discoverCmd() *cobra.Command {
	var sources string
	var global bool
	var local bool

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover local Claude Code and Codex session artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Discovering sessions: sources=%s global=%v local=%v\n", sources, global, local)
			fmt.Println("TODO: scan ~/.claude, ~/.codex, .claude, .codex and repo-local traces")
			return nil
		},
	}

	cmd.Flags().StringVar(&sources, "sources", "claude,codex", "comma-separated sources: claude,codex")
	cmd.Flags().BoolVar(&global, "global", true, "scan global user directories")
	cmd.Flags().BoolVar(&local, "local", true, "scan current repository")
	return cmd
}
