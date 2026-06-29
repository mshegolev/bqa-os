package app

import (
	"context"
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coreknowledge "github.com/mshegolev/bqa-os/internal/core/knowledge"
	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	var sessionBaseDir string
	var knowledgeDir string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build reusable QA knowledge artifacts from normalized sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			reader := fsadapter.KnowledgeSessionReader{SessionBaseDir: sessionBaseDir}
			writer := fsadapter.KnowledgeWriter{KnowledgeDir: knowledgeDir}

			knowledgeUC := coreknowledge.UseCase{Reader: reader, Writer: writer, OutputDir: knowledgeDir}
			knowledgeResult, err := knowledgeUC.Run(ctx)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "BQA knowledge build completed")
			fmt.Fprintf(out, "Sessions processed: %d\n", knowledgeResult.SessionsProcessed)
			fmt.Fprintf(out, "Artifacts created: %d\n", knowledgeResult.ArtifactsCreated)
			fmt.Fprintf(out, "Output directory: %s\n", knowledgeResult.OutputDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionBaseDir, "sessions", ".bqa/input/sessions", "session input directory")
	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", ".bqa/knowledge", "knowledge output directory")
	return cmd
}
