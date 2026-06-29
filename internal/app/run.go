package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/core/runplan"
	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [task]",
		Short: "Run BQA Master Agent task plan locally",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			task := strings.TrimSpace(strings.Join(args, " "))
			store := fsadapter.BQAWorkspaceStore{}
			uc := runplan.UseCase{RegistryReader: store}
			plan, err := uc.Run(context.Background(), runplan.Options{Task: task})
			if err != nil {
				return err
			}

			renderRunPlan(cmd.OutOrStdout(), plan)
			return nil
		},
	}
}

func renderRunPlan(out io.Writer, plan runplan.Plan) {
	fmt.Fprintf(out, "BQA Master task: %s\n", plan.Task)
	fmt.Fprintf(out, "Registry loaded: agents=%d skills=%d workflows=%d\n", len(plan.Registry.Agents), len(plan.Registry.Skills), len(plan.Registry.Workflows))

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Selected agents:")
	renderRegistryItems(out, plan.Agents)

	if len(plan.Skills) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Selected skills:")
		renderRegistryItems(out, plan.Skills)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Selected workflows:")
	renderRegistryItems(out, plan.Workflows)

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Execution plan:")
	for i, step := range plan.Steps {
		fmt.Fprintf(out, "%d. %s\n", i+1, step)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Report:")
	fmt.Fprintln(out, plan.Report)
}

func renderRegistryItems(out io.Writer, items []ports.BQARegistryItem) {
	for _, item := range items {
		if item.Domain == "" {
			fmt.Fprintf(out, "- %s: %s\n", item.ID, item.Path)
			continue
		}
		fmt.Fprintf(out, "- %s (%s): %s\n", item.ID, item.Domain, item.Path)
	}
}
