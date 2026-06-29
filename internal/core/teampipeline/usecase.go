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

	plan := Plan{
		Repo:             ref.Repo,
		Issue:            issue,
		DryRun:           !opts.Execute,
		SourceOfTruth:    "github-issue",
		CurrentState:     currentState(issue.Labels),
		SelectedSubagent: subagent,
		LoopEnabled:      opts.AllowLoop,
		MaxIterations:    maxIterations,
	}
	if !opts.AllowLoop && plan.MaxIterations != 1 {
		plan.Warnings = append(plan.Warnings, "looping is disabled; max iterations forced to 1")
		plan.MaxIterations = 1
	}

	labelSet := issueLabelSet(issue.Labels)
	archApproved := labelSet[LabelArchApproved]
	readyDev := labelSet[LabelReadyDev]

	plan.Stages = stagesForIssue(issue, archApproved, readyDev)
	verification := manualVerificationCommands(issue.Body)

	if !archApproved {
		plan.Actions = append(plan.Actions, Action{
			Kind:        "request-architecture-review",
			Role:        "Technical architect",
			Description: "Route the issue through architecture review before any development role runs.",
			AddLabels:   []string{LabelNeedsArch},
		})
		return plan, nil
	}

	if !readyDev {
		plan.Actions = append(plan.Actions, Action{
			Kind:        "wait-for-ready-dev",
			Role:        "Technical architect",
			Description: "Architecture is approved, but the issue is not labeled ready for development.",
			AddLabels:   []string{LabelReadyDev},
		})
		return plan, nil
	}

	plan.Actions = append(plan.Actions,
		Action{
			Kind:         "transition-labels",
			Role:         "Developer",
			Description:  "Move the GitHub issue into development before role execution.",
			AddLabels:    []string{LabelInDev},
			RemoveLabels: []string{LabelReadyDev},
		},
		Action{
			Kind:                 "run-role",
			Role:                 "Developer",
			Runtime:              "codex-cli",
			Subagent:             subagent,
			Description:          "Run the selected Codex subagent against the architecture-approved issue.",
			VerificationCommands: verification,
		},
		Action{
			Kind:         "transition-labels",
			Role:         "QA",
			Description:  "Move the issue to QA after development completes.",
			AddLabels:    []string{LabelReadyQA},
			RemoveLabels: []string{LabelInDev},
		},
		Action{
			Kind:                 "run-role",
			Role:                 "QA",
			Runtime:              "codex-cli",
			Description:          "Verify acceptance criteria and manual checks from the GitHub issue.",
			VerificationCommands: verification,
		},
		Action{
			Kind:        "create-bug-issue",
			Role:        "QA",
			Description: "QA rejection creates bug issue with synthetic or minimized evidence and a link back to the source issue.",
			AddLabels:   []string{LabelBug, LabelQAFailed, LabelReadyDev, LabelCodexTeam},
		},
		Action{
			Kind:         "transition-labels",
			Role:         "Business owner",
			Description:  "On QA pass, send the issue to business acceptance.",
			AddLabels:    []string{LabelReadyBusiness},
			RemoveLabels: []string{LabelReadyQA},
		},
		Action{
			Kind:        "run-role",
			Role:        "Business owner",
			Runtime:     "codex-cli",
			Description: "Perform final business acceptance before done.",
		},
	)

	return plan, nil
}

func stagesForIssue(issue ports.TeamIssue, archApproved bool, readyDev bool) []Stage {
	businessStatus := "complete"
	businessReason := "GitHub issue exists as workflow source of truth."
	if issue.Number == 0 || strings.TrimSpace(issue.Title) == "" {
		businessStatus = "blocked"
		businessReason = "GitHub issue number and title are required."
	}

	archStatus := "blocked"
	archReason := "Missing bqa:arch-approved."
	if archApproved {
		archStatus = "complete"
		archReason = "bqa:arch-approved is present."
	}

	devStatus := "blocked"
	devReason := "Development waits for architecture approval."
	if archApproved && readyDev {
		devStatus = "ready"
		devReason = "bqa:ready-dev is present."
	} else if archApproved {
		devStatus = "pending"
		devReason = "Architecture is approved; bqa:ready-dev is not present."
	}

	return []Stage{
		{ID: "business-intake", Name: "Business intake", Role: "Business owner", Status: businessStatus, Reason: businessReason, Labels: []string{LabelBusiness}},
		{ID: "architecture-review", Name: "Architecture review", Role: "Technical architect", Status: archStatus, Reason: archReason, Labels: []string{LabelArchApproved, LabelNeedsArch}},
		{ID: "development", Name: "Development", Role: "Developer", Status: devStatus, Reason: devReason, Labels: []string{LabelReadyDev, LabelInDev}},
		{ID: "qa", Name: "QA", Role: "QA", Status: "pending", Reason: "Runs after development and creates bug issues on rejection.", Labels: []string{LabelReadyQA, LabelQAFailed, LabelBug}},
		{ID: "business-acceptance", Name: "Business acceptance", Role: "Business owner", Status: "pending", Reason: "Runs after QA pass.", Labels: []string{LabelReadyBusiness, LabelBusinessApproved, LabelDone}},
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
