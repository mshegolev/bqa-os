package app

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	var sessionBaseDir string
	var knowledgeDir string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build reusable QA knowledge artifacts from normalized sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, err := RunBuild(cmd.Context(), BuildOptions{
				SessionBaseDir: sessionBaseDir,
				KnowledgeDir:   knowledgeDir,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Sessions processed: %d\n", summary.SessionsProcessed)
			fmt.Printf("Knowledge artifacts created: %d\n", summary.KnowledgeArtifactsCreated)
			fmt.Printf("BQA artifacts created: %d\n", summary.BQAArtifactsCreated)
			fmt.Printf("Knowledge dir: %s\n", summary.KnowledgeDir)
			fmt.Printf("Generated dirs: %s\n", strings.Join(summary.GeneratedDirs, " "))
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionBaseDir, "sessions", defaultSessionBaseDir, "session input directory")
	cmd.Flags().StringVar(&knowledgeDir, "knowledge-dir", defaultKnowledgeDir, "knowledge output directory")
	return cmd
}
