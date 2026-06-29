package app

import (
	"context"
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coreetlpack "github.com/mshegolev/bqa-os/internal/core/etlpack"
	"github.com/spf13/cobra"
)

func etlAgentPackCmd() *cobra.Command {
	var sessionBaseDir string
	var knowledgeDir string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "etl-agent-pack",
		Short: "Generate a local ETL QA agent pack",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store := fsadapter.KnowledgeStore{
				SessionBaseDir: sessionBaseDir,
				KnowledgeDir:   knowledgeDir,
				ETLPackDir:     outputDir,
			}
			uc := coreetlpack.UseCase{Reader: store, Writer: store, OutputDir: outputDir}
			result, err := uc.Run(ctx)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "ETL agent pack artifacts created: %d\n", result.ArtifactsCreated)
			fmt.Fprintf(out, "Sessions processed: %d\n", result.SessionsProcessed)
			fmt.Fprintf(out, "Knowledge artifacts found: %d\n", result.KnowledgeArtifactsFound)
			fmt.Fprintf(out, "Synthetic examples used: %t\n", result.SyntheticExamplesUsed)
			fmt.Fprintf(out, "Output dir: %s\n", result.OutputDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionBaseDir, "sessions", ".bqa/input/sessions", "session input directory")
	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", ".bqa/knowledge", "knowledge input directory")
	cmd.Flags().StringVar(&outputDir, "output-dir", ".bqa/output/etl-agent-pack", "ETL agent pack output directory")
	return cmd
}
