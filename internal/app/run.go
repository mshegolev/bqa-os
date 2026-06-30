package app

import (
	"fmt"
	"io"
	"strings"

	bqarun "github.com/mshegolev/bqa-os/internal/core/run"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var baseDir string
	cmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run BQA Master Agent task plan locally",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			task := strings.TrimSpace(strings.Join(args, " "))
			if task == "" {
				task = "Протестируй ETL в текущем проекте"
			}
			plan := bqarun.Build(baseDir, task)
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "BQA Master task: %s\n", plan.Task)
			renderArtifacts(out, "Agents", plan.Agents)
			renderArtifacts(out, "Skills", plan.Skills)
			renderArtifacts(out, "Workflows", plan.Workflows)
			if plan.Empty() {
				fmt.Fprintf(out, "No BQA artifacts found under %q. Run `bqa init` and `bqa build` first.\n", baseDir)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&baseDir, "base-dir", ".bqa", "workspace base directory")
	return cmd
}

func renderArtifacts(out io.Writer, label string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(out, "%s (%d):\n", label, len(items))
	for _, item := range items {
		fmt.Fprintf(out, "  - %s\n", item)
	}
}
