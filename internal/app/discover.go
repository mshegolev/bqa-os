package app

import (
	"fmt"
	"strings"

	"github.com/mshegolev/bqa-os/internal/discovery"
	"github.com/spf13/cobra"
)

func discoverCmd() *cobra.Command {
	var sources string
	var global bool
	var local bool
	var manifestPath string

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover local AI coding session artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Discovering sessions: sources=%s global=%v local=%v\n", sources, global, local)
			manifest, err := discovery.Discover(discovery.Options{
				Sources: strings.Split(sources, ","),
				Global:  global,
				Local:   local,
			})
			if err != nil {
				return err
			}
			discovery.PrintSummary(manifest)
			if err := discovery.WriteManifest(manifest, manifestPath); err != nil {
				return err
			}
			fmt.Printf("Manifest: %s\n", manifestPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&sources, "sources", "claude,codex,opencode", "comma-separated sources: claude,codex,opencode")
	cmd.Flags().BoolVar(&global, "global", true, "scan global user directories")
	cmd.Flags().BoolVar(&local, "local", true, "scan current repository")
	cmd.Flags().StringVar(&manifestPath, "manifest", ".bqa/input/sessions/manifest.json", "manifest output path")
	return cmd
}
