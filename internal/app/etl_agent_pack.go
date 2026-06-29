package app

import (
	"context"
	"fmt"
	"path/filepath"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coreetlpack "github.com/mshegolev/bqa-os/internal/core/etlpack"
	"github.com/spf13/cobra"
)

func etlAgentPackCmd() *cobra.Command {
	var sessionBaseDir string
	var knowledgeDir string

	cmd := &cobra.Command{
		Use:   "etl-agent-pack",
		Short: "Generate a local ETL QA Agent Pack for Codex and Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store := fsadapter.KnowledgeStore{SessionBaseDir: sessionBaseDir, KnowledgeDir: knowledgeDir}

			uc := coreetlpack.UseCase{
				Sessions:  store,
				Knowledge: store,
				Writer:    store,
			}
			result, err := uc.Run(ctx)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "ETL agent pack files created: %d\n", result.ArtifactsCreated)
			fmt.Fprintf(out, "Sessions processed: %d\n", result.SessionsProcessed)
			fmt.Fprintf(out, "Knowledge artifacts found: %d\n", result.KnowledgeArtifactsFound)
			fmt.Fprintf(out, "Synthetic fallback: %t\n", result.UsedSyntheticDemo)
			fmt.Fprintf(out, "Pack dir: %s\n", filepath.Join(filepath.Dir(knowledgeDir), result.OutputDir))
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionBaseDir, "sessions", ".bqa/input/sessions", "session input directory")
	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", ".bqa/knowledge", "knowledge input directory and .bqa root anchor")
	return cmd
}
