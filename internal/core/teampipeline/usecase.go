package teampipeline

import (
	"context"
	"errors"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

const (
	LabelBusinessApproved = "bqa:business-approved"
	LabelArchApproved     = "bqa:arch-approved"
	LabelBlocked          = "bqa:blocked"
	LabelReadyDev         = "bqa:ready-dev"
	LabelInDev            = "bqa:in-dev"
	LabelReadyQA          = "bqa:ready-qa"
	LabelQAFailed         = "bqa:qa-failed"
	LabelBug              = "bqa:bug"
	LabelReadyBusiness    = "bqa:ready-business"
	LabelDone             = "bqa:done"
	LabelNeedsArch        = "bqa:needs-arch"
	LabelBusiness         = "bqa:business"
	LabelCodexTeam        = "bqa:codex-team"
)

type UseCase struct {
	IssueSource ports.TeamIssueSource
}

type Options struct {
	SelectedSubagent string
	Execute          bool
	AllowLoop        bool
	MaxIterations    int
}

type Plan struct {
	Repo             string
	Issue            ports.TeamIssue
	DryRun           bool
	SourceOfTruth    string
	CurrentState     string
	SelectedSubagent string
	LoopEnabled      bool
	MaxIterations    int
	Stages           []Stage
	Actions          []Action
	Warnings         []string
}

type Stage struct {
	ID     string
	Name   string
	Role   string
	Status string
	Reason string
	Labels []string
}

type Action struct {
	Kind                 string
	Role                 string
	Runtime              string
	Subagent             string
	Description          string
	AddLabels            []string
	RemoveLabels         []string
	VerificationCommands []string
}

func (u UseCase) Run(ctx context.Context, ref ports.TeamIssueRef, opts Options) (Plan, error) {
	if u.IssueSource == nil {
		return Plan{}, errors.New("team issue source is required")
	}

	issue, err := u.IssueSource.LoadTeamIssue(ctx, ref)
	if err != nil {
		return Plan{}, err
	}
	if issue.Number == 0 {
		issue.Number = ref.Number
	}

	subagent := strings.TrimSpace(opts.SelectedSubagent)
	if subagent == "" {
		subagent = "go-cli-implementer"
	}

	maxIterations := opts.MaxIterations
	if maxIterations < 1 {
		maxIterations = 1
	}

	state := currentState(issue.Labels)
	verification := manualVerificationCommands(issue.Body)
	plan := Plan{
		Repo:             ref.Repo,
		Issue:            issue,
		DryRun:           !opts.Execute,
		SourceOfTruth:    "github-issue",
		CurrentState:     state,
		SelectedSubagent: subagent,
		LoopEnabled:      opts.AllowLoop,
		MaxIterations:    maxIterations,
		Stages:           stagesForState(issue, state),
	}
	if !opts.AllowLoop && plan.MaxIterations != 1 {
		plan.Warnings = append(plan.Warnings, "looping is disabled; max iterations forced to 1")
		plan.MaxIterations = 1
	}

	switch state {
	case "blocked":
		plan.Warnings = append(plan.Warnings, "GitHub issue has bqa:blocked; resolve the blocker before continuing the team pipeline.")
		plan.Actions = append(plan.Actions, Action{
			Kind:         "resolve-blocker",
			Role:         "Technical architect",
			Description:  "Resolve the blocker and remove bqa:blocked before development, QA, or business acceptance runs.",
			RemoveLabels: []string{LabelBlocked},
		})
	case "ready-business":
		plan.Actions = append(plan.Actions, businessAcceptanceActions()...)
	case "qa-failed":
		plan.Actions = append(plan.Actions, developmentActions(subagent, verification, true)...)
		plan.Actions = append(plan.Actions, qaActions(verification)...)
		plan.Actions = append(plan.Actions, businessHandoffActions()...)
	case "ready-qa":
		plan.Actions = append(plan.Actions, qaActions(verification)...)
		plan.Actions = append(plan.Actions, businessHandoffActions()...)
	case "in-dev":
		plan.Actions = append(plan.Actions, developerRunAction(subagent, verification))
		plan.Actions = append(plan.Actions, qaHandoffAction())
		plan.Actions = append(plan.Actions, qaActions(verification)...)
		plan.Actions = append(plan.Actions, businessHandoffActions()...)
	case "ready-dev":
		plan.Actions = append(plan.Actions, developmentActions(subagent, verification, false)...)
		plan.Actions = append(plan.Actions, qaActions(verification)...)
		plan.Actions = append(plan.Actions, businessHandoffActions()...)
	case "arch-approved":
		plan.Actions = append(plan.Actions, Action{
			Kind:        "wait-for-ready-dev",
			Role:        "Technical architect",
			Description: "Architecture is approved, but the issue is not labeled ready for development.",
			AddLabels:   []string{LabelReadyDev},
		})
	case "business-approved", "done":
		// Terminal or accepted states need no further planned role action.
	default:
		plan.Actions = append(plan.Actions, Action{
			Kind:        "request-architecture-review",
			Role:        "Technical architect",
			Description: "Route the issue through architecture review before any development role runs.",
			AddLabels:   []string{LabelNeedsArch},
		})
	}

	return plan, nil
}

func developmentActions(subagent string, verification []string, fromQAFailed bool) []Action {
	removeLabels := []string{LabelReadyDev}
	description := "Move the GitHub issue into development before role execution."
	if fromQAFailed {
		removeLabels = append(removeLabels, LabelQAFailed)
		description = "Move the QA rejection back into development for a fix."
	}

	return []Action{
		{
			Kind:         "transition-labels",
			Role:         "Developer",
			Description:  description,
			AddLabels:    []string{LabelInDev},
			RemoveLabels: removeLabels,
		},
		developerRunAction(subagent, verification),
		qaHandoffAction(),
	}
}

func developerRunAction(subagent string, verification []string) Action {
	return Action{
		Kind:                 "run-role",
		Role:                 "Developer",
		Runtime:              "codex-cli",
		Subagent:             subagent,
		Description:          "Run the selected Codex subagent against the architecture-approved issue.",
		VerificationCommands: verification,
	}
}

func qaHandoffAction() Action {
	return Action{
		Kind:         "transition-labels",
		Role:         "QA",
		Description:  "Move the issue to QA after development completes.",
		AddLabels:    []string{LabelReadyQA},
		RemoveLabels: []string{LabelInDev},
	}
}

func qaActions(verification []string) []Action {
	return []Action{
		{
			Kind:                 "run-role",
			Role:                 "QA",
			Runtime:              "codex-cli",
			Description:          "Verify acceptance criteria and manual checks from the GitHub issue.",
			VerificationCommands: verification,
		},
		{
			Kind:        "create-bug-issue",
			Role:        "QA",
			Description: "QA rejection creates bug issue with synthetic or minimized evidence and a link back to the source issue.",
			AddLabels:   []string{LabelBug, LabelQAFailed, LabelReadyDev, LabelCodexTeam},
		},
	}
}

func businessHandoffActions() []Action {
	return []Action{
		{
			Kind:         "transition-labels",
			Role:         "Business owner",
			Description:  "On QA pass, send the issue to business acceptance.",
			AddLabels:    []string{LabelReadyBusiness},
			RemoveLabels: []string{LabelReadyQA},
		},
		{
			Kind:        "run-role",
			Role:        "Business owner",
			Runtime:     "codex-cli",
			Description: "Perform final business acceptance before done.",
		},
	}
}

func businessAcceptanceActions() []Action {
	return []Action{
		{
			Kind:        "run-role",
			Role:        "Business owner",
			Runtime:     "codex-cli",
			Description: "Perform final business acceptance before done.",
		},
		{
			Kind:         "transition-labels",
			Role:         "Business owner",
			Description:  "Mark the issue accepted and done after business approval.",
			AddLabels:    []string{LabelBusinessApproved, LabelDone},
			RemoveLabels: []string{LabelReadyBusiness},
		},
	}
}

func stagesForState(issue ports.TeamIssue, state string) []Stage {
	labelSet := issueLabelSet(issue.Labels)
	businessStatus := "complete"
	businessReason := "GitHub issue exists as workflow source of truth."
	if issue.Number == 0 || strings.TrimSpace(issue.Title) == "" {
		businessStatus = "blocked"
		businessReason = "GitHub issue number and title are required."
	}

	if state == "blocked" {
		archStatus := "blocked"
		archReason := "Issue has bqa:blocked; architecture review cannot proceed."
		if labelSet[LabelArchApproved] {
			archStatus = "complete"
			archReason = "Architecture approval is satisfied, but the issue is blocked."
		}

		blockedReason := "Issue has bqa:blocked; resolve the blocker before continuing."
		return []Stage{
			{ID: "business-intake", Name: "Business intake", Role: "Business owner", Status: businessStatus, Reason: businessReason, Labels: []string{LabelBusiness}},
			{ID: "architecture-review", Name: "Architecture review", Role: "Technical architect", Status: archStatus, Reason: archReason, Labels: []string{LabelArchApproved, LabelNeedsArch, LabelBlocked}},
			{ID: "development", Name: "Development", Role: "Developer", Status: "blocked", Reason: blockedReason, Labels: []string{LabelReadyDev, LabelInDev, LabelBlocked}},
			{ID: "qa", Name: "QA", Role: "QA", Status: "blocked", Reason: blockedReason, Labels: []string{LabelReadyQA, LabelQAFailed, LabelBug, LabelBlocked}},
			{ID: "business-acceptance", Name: "Business acceptance", Role: "Business owner", Status: "blocked", Reason: blockedReason, Labels: []string{LabelReadyBusiness, LabelBusinessApproved, LabelDone, LabelBlocked}},
		}
	}

	archStatus := "blocked"
	archReason := "Missing bqa:arch-approved."
	if stateAtLeast(state, "arch-approved") {
		archStatus = "complete"
		archReason = "Architecture approval is satisfied by the current workflow state."
	}

	devStatus := "blocked"
	devReason := "Development waits for architecture approval."
	switch state {
	case "arch-approved":
		devStatus = "pending"
		devReason = "Architecture is approved; bqa:ready-dev is not present."
	case "ready-dev", "qa-failed":
		devStatus = "ready"
		devReason = "Development is the next role for the current workflow state."
	case "in-dev":
		devStatus = "active"
		devReason = "Development is in progress."
	case "ready-qa", "ready-business", "business-approved", "done":
		devStatus = "complete"
		devReason = "Development is complete for the current workflow state."
	}

	qaStatus := "pending"
	qaReason := "Runs after development and creates bug issues on rejection."
	switch state {
	case "qa-failed":
		qaStatus = "blocked"
		qaReason = "QA failed; development fixes are required before verification resumes."
	case "ready-qa":
		qaStatus = "ready"
		qaReason = "bqa:ready-qa is the highest-priority current workflow label."
	case "ready-business", "business-approved", "done":
		qaStatus = "complete"
		qaReason = "QA is complete for the current workflow state."
	}

	businessAcceptanceStatus := "pending"
	businessAcceptanceReason := "Runs after QA pass."
	switch state {
	case "ready-business":
		businessAcceptanceStatus = "ready"
		businessAcceptanceReason = "bqa:ready-business is the highest-priority current workflow label."
	case "business-approved", "done":
		businessAcceptanceStatus = "complete"
		businessAcceptanceReason = "Business acceptance is complete."
	}

	return []Stage{
		{ID: "business-intake", Name: "Business intake", Role: "Business owner", Status: businessStatus, Reason: businessReason, Labels: []string{LabelBusiness}},
		{ID: "architecture-review", Name: "Architecture review", Role: "Technical architect", Status: archStatus, Reason: archReason, Labels: []string{LabelArchApproved, LabelNeedsArch}},
		{ID: "development", Name: "Development", Role: "Developer", Status: devStatus, Reason: devReason, Labels: []string{LabelReadyDev, LabelInDev}},
		{ID: "qa", Name: "QA", Role: "QA", Status: qaStatus, Reason: qaReason, Labels: []string{LabelReadyQA, LabelQAFailed, LabelBug}},
		{ID: "business-acceptance", Name: "Business acceptance", Role: "Business owner", Status: businessAcceptanceStatus, Reason: businessAcceptanceReason, Labels: []string{LabelReadyBusiness, LabelBusinessApproved, LabelDone}},
	}
}

func stateAtLeast(actual string, minimum string) bool {
	return stateRank(actual) >= stateRank(minimum)
}

func stateRank(state string) int {
	switch state {
	case "business":
		return 1
	case "needs-arch":
		return 2
	case "arch-approved":
		return 3
	case "ready-dev":
		return 4
	case "in-dev":
		return 5
	case "ready-qa":
		return 6
	case "qa-failed":
		return 7
	case "ready-business":
		return 8
	case "business-approved":
		return 9
	case "done":
		return 10
	default:
		return 0
	}
}

func issueLabelSet(labels []ports.TeamIssueLabel) map[string]bool {
	out := map[string]bool{}
	for _, label := range labels {
		name := strings.TrimSpace(label.Name)
		if name != "" {
			out[name] = true
		}
	}
	return out
}

func currentState(labels []ports.TeamIssueLabel) string {
	labelSet := issueLabelSet(labels)
	ordered := []struct {
		label string
		state string
	}{
		{LabelBlocked, "blocked"},
		{LabelDone, "done"},
		{LabelBusinessApproved, "business-approved"},
		{LabelReadyBusiness, "ready-business"},
		{LabelQAFailed, "qa-failed"},
		{LabelReadyQA, "ready-qa"},
		{LabelInDev, "in-dev"},
		{LabelReadyDev, "ready-dev"},
		{LabelArchApproved, "arch-approved"},
		{LabelNeedsArch, "needs-arch"},
		{LabelBusiness, "business"},
	}
	for _, item := range ordered {
		if labelSet[item.label] {
			return item.state
		}
	}
	return "unknown"
}

func manualVerificationCommands(body string) []string {
	lines := strings.Split(body, "\n")
	inManualSection := false
	inFence := false
	var commands []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "## manual verification") {
			inManualSection = true
			continue
		}
		if inManualSection && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if !inManualSection {
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence && trimmed != "" {
			commands = append(commands, trimmed)
		}
	}

	if len(commands) == 0 && strings.Contains(body, "go test ./...") {
		return []string{"go test ./..."}
	}
	return commands
}
