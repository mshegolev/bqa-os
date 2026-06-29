package teampipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeIssueSource struct {
	issue ports.TeamIssue
	err   error
}

func (f fakeIssueSource) LoadTeamIssue(ctx context.Context, ref ports.TeamIssueRef) (ports.TeamIssue, error) {
	return f.issue, f.err
}

func TestUseCasePlansDryRunForArchApprovedIssue(t *testing.T) {
	issue := ports.TeamIssue{
		Number: 27,
		Title:  "Codex Team Pipeline MVP",
		Body: strings.Join([]string{
			"## Acceptance criteria",
			"- [ ] Architecture boundaries are respected.",
			"",
			"## Manual verification",
			"",
			"```bash",
			"go test ./...",
			"```",
		}, "\n"),
		Labels: []ports.TeamIssueLabel{
			{Name: "bqa:arch-approved"},
			{Name: "bqa:ready-dev"},
			{Name: "bqa:codex-team"},
		},
	}
	uc := UseCase{IssueSource: fakeIssueSource{issue: issue}}

	plan, err := uc.Run(context.Background(), ports.TeamIssueRef{Repo: "mshegolev/bqa-os", Number: 27}, Options{
		SelectedSubagent: "senior-go-ai-engineer",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !plan.DryRun {
		t.Fatalf("pipeline MVP must be dry-run by default")
	}
	if plan.SourceOfTruth != "github-issue" {
		t.Fatalf("expected GitHub issue source of truth, got %q", plan.SourceOfTruth)
	}
	if plan.CurrentState != "ready-dev" {
		t.Fatalf("expected ready-dev state, got %q", plan.CurrentState)
	}
	if plan.SelectedSubagent != "senior-go-ai-engineer" {
		t.Fatalf("expected selected subagent to be preserved, got %q", plan.SelectedSubagent)
	}
	if plan.MaxIterations != 1 || plan.LoopEnabled {
		t.Fatalf("expected bounded one-shot dry-run, got max=%d loop=%v", plan.MaxIterations, plan.LoopEnabled)
	}

	assertStage(t, plan, "architecture-review", "complete")
	assertStage(t, plan, "development", "ready")
	assertStage(t, plan, "qa", "pending")
	assertStage(t, plan, "business-acceptance", "pending")

	if !hasAction(plan, "run-role", "Developer") {
		t.Fatalf("expected developer role action, got %#v", plan.Actions)
	}
	if !hasVerification(plan, "go test ./...") {
		t.Fatalf("expected manual verification command to be preserved, got %#v", plan.Actions)
	}
	bugAction := findAction(plan, "create-bug-issue")
	if bugAction == nil {
		t.Fatalf("expected QA rejection bug action, got %#v", plan.Actions)
	}
	for _, label := range []string{"bqa:bug", "bqa:qa-failed", "bqa:ready-dev", "bqa:codex-team"} {
		if !contains(bugAction.AddLabels, label) {
			t.Fatalf("expected bug action to add label %q, got %#v", label, bugAction.AddLabels)
		}
	}
}

func TestUseCaseBlocksDevelopmentWithoutArchitectureApproval(t *testing.T) {
	issue := ports.TeamIssue{
		Number: 28,
		Title:  "Synthetic task without architecture",
		Body:   "## Manual verification\n\n```bash\ngo test ./...\n```",
		Labels: []ports.TeamIssueLabel{{Name: "bqa:ready-dev"}},
	}
	uc := UseCase{IssueSource: fakeIssueSource{issue: issue}}

	plan, err := uc.Run(context.Background(), ports.TeamIssueRef{Repo: "mshegolev/bqa-os", Number: 28}, Options{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	assertStage(t, plan, "architecture-review", "blocked")
	assertStage(t, plan, "development", "blocked")
	if hasAction(plan, "run-role", "Developer") {
		t.Fatalf("developer action must not be planned before architecture approval")
	}
	archAction := findAction(plan, "request-architecture-review")
	if archAction == nil {
		t.Fatalf("expected architecture review request action, got %#v", plan.Actions)
	}
	if !contains(archAction.AddLabels, "bqa:needs-arch") {
		t.Fatalf("expected architecture action to add bqa:needs-arch, got %#v", archAction.AddLabels)
	}
}

func assertStage(t *testing.T, plan Plan, id string, status string) {
	t.Helper()
	for _, stage := range plan.Stages {
		if stage.ID == id {
			if stage.Status != status {
				t.Fatalf("stage %s status = %q, expected %q", id, stage.Status, status)
			}
			return
		}
	}
	t.Fatalf("stage %s not found in %#v", id, plan.Stages)
}

func hasAction(plan Plan, kind string, role string) bool {
	return findActionByKindAndRole(plan, kind, role) != nil
}

func findActionByKindAndRole(plan Plan, kind string, role string) *Action {
	for i := range plan.Actions {
		if plan.Actions[i].Kind == kind && plan.Actions[i].Role == role {
			return &plan.Actions[i]
		}
	}
	return nil
}

func findAction(plan Plan, kind string) *Action {
	for i := range plan.Actions {
		if plan.Actions[i].Kind == kind {
			return &plan.Actions[i]
		}
	}
	return nil
}

func hasVerification(plan Plan, command string) bool {
	for _, action := range plan.Actions {
		for _, actual := range action.VerificationCommands {
			if actual == command {
				return true
			}
		}
	}
	return false
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
