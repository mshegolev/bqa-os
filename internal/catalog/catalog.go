// Package catalog is the single source of truth for bqa-os's built-in
// agents, skills, and workflows, held as data and rendered on demand.
package catalog

import "strings"

// Entry is a built-in skill or workflow artifact.
type Entry struct {
	ID      string
	Type    string // "skill" | "workflow"
	Domain  string
	Title   string
	Content string
}

// AgentCore is the de-duplicated base persona for an agent.
type AgentCore struct {
	ID      string
	Title   string
	Domain  string
	Mission string
	Rules   []string
}

// RuntimeFlavor is the runtime-specific delta applied on top of an AgentCore.
type RuntimeFlavor struct {
	TitlePrefix string
	Intro       string
	ExtraRules  []string
}

// RenderAgent renders an agent persona from a core and an optional runtime flavor.
func RenderAgent(core AgentCore, flavor *RuntimeFlavor) string {
	title := core.Title
	role := core.Mission
	rules := core.Rules
	if flavor != nil {
		if flavor.TitlePrefix != "" {
			title = flavor.TitlePrefix + " " + core.Title
		}
		if flavor.Intro != "" {
			role = flavor.Intro + " " + core.Mission
		}
		rules = append(append([]string{}, core.Rules...), flavor.ExtraRules...)
	}

	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString("## Role\n\n")
	b.WriteString(role + "\n\n")
	b.WriteString("## Operating Rules\n\n")
	for _, rule := range rules {
		b.WriteString("- " + rule + "\n")
	}
	return b.String()
}

var etlQACore = AgentCore{
	ID:      "etl-qa-agent",
	Title:   "ETL QA Agent",
	Domain:  "etl",
	Mission: "Validate Big Data / ETL changes (Airflow, Spark, Hive, reconciliation, data quality) with reproducible evidence.",
	Rules: []string{
		"Build concise investigation plans before changing anything.",
		"Prefer commands, logs, row counts, and schema evidence over guesses.",
		"Keep private data, raw logs, secrets, and customer records out of output.",
		"Report scope, checks, evidence, result, risks, and next action.",
	},
}

var codexFlavor = RuntimeFlavor{
	TitlePrefix: "Codex",
	Intro:       "You are the ETL QA Agent working in Codex.",
	ExtraRules: []string{
		"Prefer existing repository test commands before inventing new helpers.",
		"When inputs are missing, use the synthetic example in this pack to demonstrate the workflow.",
	},
}

var claudeFlavor = RuntimeFlavor{
	TitlePrefix: "Claude Code",
	Intro:       "You are the ETL QA Agent working in Claude Code.",
	ExtraRules: []string{
		"Follow the project's existing tooling and test framework.",
		"Ask a blocker question when environment, dataset, or acceptance criteria are unclear.",
	},
}

// ETLQA returns the de-duplicated ETL QA agent core persona.
func ETLQA() AgentCore { return etlQACore }

// CodexFlavor returns the Codex runtime flavor.
func CodexFlavor() *RuntimeFlavor { return &codexFlavor }

// ClaudeFlavor returns the Claude Code runtime flavor.
func ClaudeFlavor() *RuntimeFlavor { return &claudeFlavor }
