package app

import (
	"context"
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coreingest "github.com/mshegolev/bqa-os/internal/core/ingest"
	"github.com/spf13/cobra"
)

func ingestCmd() *cobra.Command {
	var sources string
	var global bool
	var local bool
	var baseDir string
	var from string

	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest sessions through BQA core",
		Long: "Ingest sessions through BQA core.\n\n" +
			"With --from <dir>, import local ETL notes/log snippets (*.md, *.log, *.txt)\n" +
			"from a directory into normalized session markdown plus a valid index.json,\n" +
			"ready for `bqa build`. Secrets are redacted before they reach the normalized\n" +
			"artifacts; sanitize client data before committing the .bqa/ directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := &fsadapter.SessionStore{BaseDir: baseDir}

			if from != "" {
				uc := coreingest.ImportLocal{
					Source: fsadapter.LocalNotesSource{Dir: from, Source: "local-etl"},
					Store:  store,
				}
				result, err := uc.Run(context.Background())
				if err != nil {
					return err
				}
				fmt.Printf("Discovered: %d\n", result.Discovered)
				fmt.Printf("Imported: %d\n", result.Imported)
				fmt.Printf("Redactions: %d\n", result.Redactions)
				fmt.Printf("Index: %s/index.json\n", baseDir)
				return nil
			}

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
	cmd.Flags().StringVar(&from, "from", "", "import local ETL notes/logs (*.md,*.log,*.txt) from this directory")
	return cmd
}
