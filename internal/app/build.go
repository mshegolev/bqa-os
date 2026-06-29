package app

import (
	"context"
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coreartifacts "github.com/mshegolev/bqa-os/internal/core/artifacts"
	coreknowledge "github.com/mshegolev/bqa-os/internal/core/knowledge"
	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	var sessionBaseDir string
	var knowledgeDir string
	var includeSalesPackage bool

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build reusable QA knowledge artifacts from normalized sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store := fsadapter.KnowledgeStore{SessionBaseDir: sessionBaseDir, KnowledgeDir: knowledgeDir}

			knowledgeUC := coreknowledge.UseCase{Reader: store, Writer: store, OutputDir: knowledgeDir}
			knowledgeResult, err := knowledgeUC.Run(ctx)
			if err != nil {
				return err
			}

			artifactUC := coreartifacts.UseCase{Writer: store, IncludeSalesPackage: includeSalesPackage}
			artifactResult, err := artifactUC.Run(ctx)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Sessions processed: %d\n", knowledgeResult.SessionsProcessed)
			fmt.Fprintf(out, "Knowledge artifacts created: %d\n", knowledgeResult.ArtifactsCreated)
			fmt.Fprintf(out, "BQA artifacts created: %d\n", artifactResult.ArtifactsCreated)
			fmt.Fprintf(out, "Knowledge dir: %s\n", knowledgeDir)
			generatedDirs := ".bqa/skills .bqa/agents .bqa/workflows .bqa/registry"
			if includeSalesPackage {
				generatedDirs = ".bqa/skills .bqa/agents .bqa/workflows .bqa/sales .bqa/registry"
			}
			fmt.Fprintf(out, "Generated dirs: %s\n", generatedDirs)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionBaseDir, "sessions", ".bqa/input/sessions", "session input directory")
	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", ".bqa/knowledge", "knowledge output directory")
	cmd.Flags().BoolVar(&includeSalesPackage, "sales-package", false, "also generate internal pilot sales package artifacts")
	return cmd
}
