package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func ingestCmd() *cobra.Command {
	var sources string
	var global bool
	var local bool

	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Discover, export, normalize, and analyze sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Ingesting sessions: sources=%s global=%v local=%v\n", sources, global, local)
			fmt.Println("TODO: create .bqa/input/sessions/manifest.json, normalize sessions, update memory and registry")
			return nil
		},
	}

	cmd.Flags().StringVar(&sources, "sources", "claude,codex", "comma-separated sources: claude,codex")
	cmd.Flags().BoolVar(&global, "global", true, "scan global user directories")
	cmd.Flags().BoolVar(&local, "local", true, "scan current repository")
	return cmd
}
