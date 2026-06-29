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

func TestUseCaseRoutesReadyQAStateToQAVerification(t *testing.T) {
	issue := teamIssueWithLabels(27, "Ready QA task", "bqa:arch-approved", "bqa:ready-qa", "bqa:codex-team")
	uc := UseCase{IssueSource: fakeIssueSource{issue: issue}}

	plan, err := uc.Run(context.Background(), ports.TeamIssueRef{Repo: "mshegolev/bqa-os", Number: 27}, Options{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if plan.CurrentState != "ready-qa" {
		t.Fatalf("expected ready-qa state, got %q", plan.CurrentState)
	}
	assertStage(t, plan, "development", "complete")
	assertStage(t, plan, "qa", "ready")
	if hasAction(plan, "wait-for-ready-dev", "Technical architect") {
		t.Fatalf("ready-qa must not be routed back to ready-dev, got %#v", plan.Actions)
	}
	if hasAction(plan, "run-role", "Developer") {
		t.Fatalf("ready-qa must not plan development, got %#v", plan.Actions)
	}
	if !hasAction(plan, "run-role", "QA") {
		t.Fatalf("expected QA verification action, got %#v", plan.Actions)
	}
	if !hasVerification(plan, "QA", "go test ./...") {
		t.Fatalf("expected QA action to preserve manual verification, got %#v", plan.Actions)
	}
	if findAction(plan, "create-bug-issue") == nil {
		t.Fatalf("expected QA rejection bug issue action, got %#v", plan.Actions)
	}
}

func TestUseCaseTreatsBlockedLabelAsHighestPriorityState(t *testing.T) {
	issue := teamIssueWithLabels(27, "Blocked workflow task", "bqa:blocked", "bqa:arch-approved", "bqa:ready-dev", "bqa:codex-team")
	uc := UseCase{IssueSource: fakeIssueSource{issue: issue}}

	plan, err := uc.Run(context.Background(), ports.TeamIssueRef{Repo: "mshegolev/bqa-os", Number: 27}, Options{
		SelectedSubagent: "go-cli-implementer",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if plan.CurrentState != "blocked" {
		t.Fatalf("expected blocked state, got %q", plan.CurrentState)
	}
	assertStage(t, plan, "development", "blocked")
	assertStage(t, plan, "qa", "blocked")
	assertStage(t, plan, "business-acceptance", "blocked")
	if !hasWarning(plan, "bqa:blocked") {
		t.Fatalf("expected blocked warning, got %#v", plan.Warnings)
	}
	if hasAction(plan, "run-role", "Developer") || hasAction(plan, "run-role", "QA") {
		t.Fatalf("blocked issue must not plan developer or QA role actions, got %#v", plan.Actions)
	}
	action := findAction(plan, "resolve-blocker")
	if action == nil {
		t.Fatalf("expected blocker resolution action, got %#v", plan.Actions)
	}
	if !contains(action.RemoveLabels, "bqa:blocked") {
		t.Fatalf("expected blocker resolution to remove bqa:blocked after resolution, got %#v", action.RemoveLabels)
	}
}

func TestUseCaseRoutesQAFailedStateBackToDevelopment(t *testing.T) {
	issue := teamIssueWithLabels(76, "QA failed bug", "bqa:ready-dev", "bqa:qa-failed", "bqa:bug")
	uc := UseCase{IssueSource: fakeIssueSource{issue: issue}}

	plan, err := uc.Run(context.Background(), ports.TeamIssueRef{Repo: "mshegolev/bqa-os", Number: 76}, Options{
		SelectedSubagent: "go-cli-implementer",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if plan.CurrentState != "qa-failed" {
		t.Fatalf("expected qa-failed state, got %q", plan.CurrentState)
	}
	assertStage(t, plan, "development", "ready")
	assertStage(t, plan, "qa", "blocked")
	if hasAction(plan, "wait-for-ready-dev", "Technical architect") {
		t.Fatalf("qa-failed must not be routed back to ready-dev, got %#v", plan.Actions)
	}
	if !hasAction(plan, "run-role", "Developer") {
		t.Fatalf("expected development fix action, got %#v", plan.Actions)
	}
	fixAction := findActionByKindAndRole(plan, "transition-labels", "Developer")
	if fixAction == nil {
		t.Fatalf("expected development transition action, got %#v", plan.Actions)
	}
	if !contains(fixAction.RemoveLabels, "bqa:qa-failed") {
		t.Fatalf("expected development transition to remove bqa:qa-failed, got %#v", fixAction.RemoveLabels)
	}
	if !hasVerification(plan, "Developer", "go test ./...") {
		t.Fatalf("expected developer action to preserve manual verification, got %#v", plan.Actions)
	}
}

func TestUseCaseRoutesReadyBusinessStateToBusinessAcceptance(t *testing.T) {
	issue := teamIssueWithLabels(77, "Ready business task", "bqa:ready-business")
	uc := UseCase{IssueSource: fakeIssueSource{issue: issue}}

	plan, err := uc.Run(context.Background(), ports.TeamIssueRef{Repo: "mshegolev/bqa-os", Number: 77}, Options{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if plan.CurrentState != "ready-business" {
		t.Fatalf("expected ready-business state, got %q", plan.CurrentState)
	}
	assertStage(t, plan, "development", "complete")
	assertStage(t, plan, "qa", "complete")
	assertStage(t, plan, "business-acceptance", "ready")
	if hasAction(plan, "run-role", "Developer") || hasAction(plan, "run-role", "QA") {
		t.Fatalf("ready-business must not plan developer or QA role actions, got %#v", plan.Actions)
	}
	if !hasAction(plan, "run-role", "Business owner") {
		t.Fatalf("expected business acceptance action, got %#v", plan.Actions)
	}
}

func teamIssueWithLabels(number int, title string, labels ...string) ports.TeamIssue {
	issueLabels := make([]ports.TeamIssueLabel, 0, len(labels))
	for _, label := range labels {
		issueLabels = append(issueLabels, ports.TeamIssueLabel{Name: label})
	}

	return ports.TeamIssue{
		Number: number,
		Title:  title,
		Body: strings.Join([]string{
			"## Acceptance criteria",
			"- [ ] Synthetic acceptance criteria.",
			"",
			"## Manual verification",
			"",
			"```bash",
			"go test ./...",
			"```",
		}, "\n"),
		Labels: issueLabels,
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

func hasVerification(plan Plan, role string, command string) bool {
	action := findActionByKindAndRole(plan, "run-role", role)
	if action == nil {
		return false
	}
	for _, actual := range action.VerificationCommands {
		if actual == command {
			return true
		}
	}
	return false
}

func hasWarning(plan Plan, expected string) bool {
	for _, warning := range plan.Warnings {
		if strings.Contains(warning, expected) {
			return true
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
