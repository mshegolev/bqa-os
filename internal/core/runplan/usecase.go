package runplan

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

const defaultTask = "Протестируй ETL в текущем проекте"

type UseCase struct {
	RegistryReader ports.BQARegistryReader
}

type Options struct {
	Task string
}

type Plan struct {
	Task      string
	Registry  ports.BQARegistry
	Agents    []ports.BQARegistryItem
	Skills    []ports.BQARegistryItem
	Workflows []ports.BQARegistryItem
	Steps     []string
	Report    string
}

func (u UseCase) Run(ctx context.Context, opts Options) (Plan, error) {
	if u.RegistryReader == nil {
		return Plan{}, errors.New("registry reader is required")
	}

	registry, err := u.RegistryReader.LoadBQARegistry(ctx)
	if err != nil {
		return Plan{}, err
	}
	if len(registry.Agents) == 0 {
		return Plan{}, errors.New("registry has no agents")
	}

	task := strings.TrimSpace(opts.Task)
	if task == "" {
		task = defaultTask
	}

	domains := inferDomains(task)
	agents := selectItems(registry.Agents, domains)
	skills := selectItems(registry.Skills, domains)
	workflows := selectItems(registry.Workflows, domains)

	plan := Plan{
		Task:      task,
		Registry:  registry,
		Agents:    agents,
		Skills:    skills,
		Workflows: workflows,
		Steps: []string{
			"Load the project BQA registry.",
			"Use the selected agents to inspect the task and project context.",
			"Follow the selected workflows to gather evidence and checks.",
			"Produce a human-reviewed QA report with findings, risks, and next steps.",
		},
	}
	plan.Report = fmt.Sprintf(
		"Plan created from %d selected agent(s), %d skill(s), and %d workflow(s). Execution remains human-in-the-loop; no external systems were modified.",
		len(plan.Agents),
		len(plan.Skills),
		len(plan.Workflows),
	)
	return plan, nil
}

func inferDomains(task string) map[string]bool {
	text := strings.ToLower(task)
	domains := map[string]bool{}
	for domain, needles := range map[string][]string{
		"etl": {
			"etl",
			"big data",
			"pipeline",
			"hadoop",
			"hive",
			"spark",
			"oozie",
			"stage",
			"uat",
		},
		"runtime": {
			"runtime",
			"codex",
			"claude",
			"opencode",
		},
		"memory": {
			"memory",
			"session",
			"sessions",
			"knowledge",
		},
		"graphql": {
			"graphql",
		},
		"api": {
			"api",
			"contract",
		},
	} {
		for _, needle := range needles {
			if strings.Contains(text, needle) {
				domains[domain] = true
				break
			}
		}
	}
	return domains
}

func selectItems(items []ports.BQARegistryItem, domains map[string]bool) []ports.BQARegistryItem {
	if len(items) == 0 {
		return nil
	}
	if len(domains) == 0 {
		return append([]ports.BQARegistryItem(nil), items...)
	}

	var selected []ports.BQARegistryItem
	for _, item := range items {
		domain := strings.ToLower(strings.TrimSpace(item.Domain))
		if domains[domain] {
			selected = append(selected, item)
		}
	}
	if len(selected) == 0 {
		return append([]ports.BQARegistryItem(nil), items...)
	}
	return selected
}
