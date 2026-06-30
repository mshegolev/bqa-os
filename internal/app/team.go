package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	githubadapter "github.com/mshegolev/bqa-os/internal/adapters/github"
	"github.com/mshegolev/bqa-os/internal/core/teampipeline"
	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/spf13/cobra"
)

func teamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Plan BQA team workflows",
	}
	cmd.AddCommand(teamPipelineCmd())
	return cmd
}

func teamPipelineCmd() *cobra.Command {
	var issueJSON string
	var issueNumber int
	var repo string
	var subagent string
	var allowLoop bool
	var maxIterations int

	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Plan a dry-run Codex team pipeline from a GitHub issue snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(issueJSON) == "" {
				return errors.New("--issue-json is required")
			}

			source := githubadapter.IssueJSONSource{Path: issueJSON}
			uc := teampipeline.UseCase{IssueSource: source}
			plan, err := uc.Run(context.Background(), ports.TeamIssueRef{
				Repo:     repo,
				Number:   issueNumber,
				JSONPath: issueJSON,
			}, teampipeline.Options{
				SelectedSubagent: subagent,
				AllowLoop:        allowLoop,
				MaxIterations:    maxIterations,
			})
			if err != nil {
				return err
			}

			renderTeamPipelinePlan(cmd.OutOrStdout(), plan)
			return nil
		},
	}

	cmd.Flags().StringVar(&issueJSON, "issue-json", "", "path to a GitHub issue JSON snapshot")
	cmd.Flags().IntVar(&issueNumber, "issue-number", 0, "GitHub issue number used when the JSON snapshot omits number")
	cmd.Flags().StringVar(&repo, "repo", "mshegolev/bqa-os", "GitHub repository in owner/name format")
	cmd.Flags().StringVar(&subagent, "subagent", "go-cli-implementer", "Codex subagent key to plan for development")
	cmd.Flags().BoolVar(&allowLoop, "allow-loop", false, "allow more than one planned pipeline iteration")
	cmd.Flags().IntVar(&maxIterations, "max-iterations", 1, "maximum planned pipeline iterations")
	return cmd
}

func renderTeamPipelinePlan(out io.Writer, plan teampipeline.Plan) {
	mode := "execute"
	if plan.DryRun {
		mode = "dry-run"
	}

	fmt.Fprintf(out, "Mode: %s\n", mode)
	fmt.Fprintln(out, "Source of truth: GitHub issue")
	fmt.Fprintf(out, "Repository: %s\n", plan.Repo)
	fmt.Fprintf(out, "Issue: #%d %s\n", plan.Issue.Number, plan.Issue.Title)
	fmt.Fprintf(out, "Current state: %s\n", plan.CurrentState)
	fmt.Fprintf(out, "Selected subagent: %s\n", plan.SelectedSubagent)
	fmt.Fprintf(out, "Loop enabled: %t\n", plan.LoopEnabled)
	fmt.Fprintf(out, "Max iterations: %d\n", plan.MaxIterations)

	if len(plan.Warnings) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Warnings:")
		for _, warning := range plan.Warnings {
			fmt.Fprintf(out, "- %s\n", warning)
		}
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Stages:")
	for _, stage := range plan.Stages {
		fmt.Fprintf(out, "- %s: %s (%s)\n", stage.Name, stage.Status, stage.Role)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Actions:")
	for _, action := range plan.Actions {
		fmt.Fprintf(out, "- %s: %s\n", action.Role, action.Description)
		if action.Runtime != "" {
			fmt.Fprintf(out, "  runtime: %s\n", action.Runtime)
		}
		if action.Subagent != "" {
			fmt.Fprintf(out, "  subagent: %s\n", action.Subagent)
		}
		if len(action.AddLabels) > 0 {
			fmt.Fprintf(out, "  add labels: %s\n", strings.Join(action.AddLabels, ", "))
		}
		if len(action.RemoveLabels) > 0 {
			fmt.Fprintf(out, "  remove labels: %s\n", strings.Join(action.RemoveLabels, ", "))
		}
		for _, command := range action.VerificationCommands {
			fmt.Fprintf(out, "  verify: %s\n", command)
		}
		if action.BugSpec != nil {
			fmt.Fprintf(out, "  bug title: %s\n", action.BugSpec.Title)
			if len(action.BugSpec.Labels) > 0 {
				fmt.Fprintf(out, "  bug labels: %s\n", strings.Join(action.BugSpec.Labels, ", "))
			}
			fmt.Fprintln(out, "  bug body:")
			for _, line := range strings.Split(strings.TrimRight(action.BugSpec.Body, "\n"), "\n") {
				fmt.Fprintf(out, "    %s\n", line)
			}
		}
	}
}
