package app

import (
	"context"
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coreingest "github.com/mshegolev/bqa-os/internal/core/ingest"
	"github.com/spf13/cobra"
)

func ingest2Cmd() *cobra.Command {
	var sources string
	var global bool
	var local bool
	var baseDir string

	cmd := &cobra.Command{
		Use:   "ingest2",
		Short: "Ingest through hexagonal core",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := &fsadapter.SessionStore{BaseDir: baseDir}
			uc := coreingest.UseCase{
				Source: fsadapter.SessionSource{Roots: sessionRoots(sources, global, local)},
				Store:  store,
			}
			result, err := uc.Run(context.Background())
			if err != nil {
				return err
			}
			fmt.Printf("Discovered: %d\n", result.Discovered)
			fmt.Printf("Ingested: %d\n", result.Ingested)
			fmt.Printf("Index: %s/index.json\n", baseDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&sources, "sources", "claude,codex,opencode", "sources")
	cmd.Flags().BoolVar(&global, "global", true, "global")
	cmd.Flags().BoolVar(&local, "local", true, "local")
	cmd.Flags().StringVar(&baseDir, "base-dir", ".bqa/input/sessions", "base dir")
	return cmd
}
