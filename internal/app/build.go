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
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build reusable QA knowledge artifacts from normalized sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store := fsadapter.KnowledgeStore{SessionBaseDir: sessionBaseDir, KnowledgeDir: knowledgeDir}

			if checkOnly {
				return runBuildCheck(ctx, cmd, store, knowledgeDir)
			}

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
	cmd.Flags().BoolVar(&checkOnly, "check", false, "validate existing knowledge artifacts instead of building; exits non-zero on invalid output")
	return cmd
}

// runBuildCheck validates already-generated knowledge artifacts without
// building. It runs locally with no external services and returns an error
// (non-zero exit) when any expected artifact is missing, empty, or malformed.
func runBuildCheck(ctx context.Context, cmd *cobra.Command, store fsadapter.KnowledgeStore, knowledgeDir string) error {
	out := cmd.OutOrStdout()
	report := coreknowledge.Validate(ctx, store)

	fmt.Fprintf(out, "Validating knowledge artifacts in: %s\n", knowledgeDir)
	fmt.Fprintf(out, "Artifacts valid: %d of %d expected\n", report.Valid, report.Expected)

	if report.OK() {
		fmt.Fprintln(out, "All knowledge artifacts are present and valid.")
		return nil
	}

	fmt.Fprintln(out, "Invalid build output:")
	for _, issue := range report.Issues {
		if issue.Filename == "" {
			fmt.Fprintf(out, "  - %s\n", issue.Detail)
			continue
		}
		fmt.Fprintf(out, "  - %s: %s\n", issue.Filename, issue.Detail)
	}
	return fmt.Errorf("knowledge artifact validation failed: %d issue(s); run `bqa build` to regenerate", len(report.Issues))
}
