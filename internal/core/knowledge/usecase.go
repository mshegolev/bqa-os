package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type UseCase struct {
	Reader    ports.NormalizedSessionReader
	Writer    ports.KnowledgeArtifactWriter
	OutputDir string
}

func (u UseCase) Run(ctx context.Context) (Result, error) {
	index, err := u.Reader.LoadSessionIndex(ctx)
	if err != nil {
		return Result{}, err
	}

	findings := map[string][]Finding{
		"etl":          {},
		"graphql":      {},
		"api":          {},
		"data_quality": {},
		"bugs":         {},
		"prompts":      {},
	}
	profile := ProjectProfile{Sessions: len(index.Entries)}
	processed := 0

	for _, entry := range index.Entries {
		body, err := u.Reader.ReadNormalizedSession(ctx, entry.NormalizedPath)
		if err != nil {
			continue
		}
		processed++
		lower := strings.ToLower(body)
		if hasAny(lower, "etl", "spark", "hive", "airflow", "dag", "reconciliation", "source table", "target table") {
			profile.ETLSignals++
			findings["etl"] = append(findings["etl"], Finding{Name: "etl_validation", Domain: "etl", Evidence: evidence(lower, "etl"), SourcePath: entry.NormalizedPath})
		}
		if hasAny(lower, "graphql", "schema", "query", "mutation", "resolver") {
			profile.GraphQLSignals++
			findings["graphql"] = append(findings["graphql"], Finding{Name: "graphql_functional_testing", Domain: "graphql", Evidence: evidence(lower, "graphql"), SourcePath: entry.NormalizedPath})
		}
		if hasAny(lower, "api", "http", "rest", "endpoint", "status code", "contract") {
			profile.APISignals++
			findings["api"] = append(findings["api"], Finding{Name: "api_contract_testing", Domain: "api", Evidence: evidence(lower, "api"), SourcePath: entry.NormalizedPath})
		}
		if hasAny(lower, "data quality", "dq", "null", "duplicate", "schema drift", "row count", "checksum") {
			profile.DQSignals++
			findings["data_quality"] = append(findings["data_quality"], Finding{Name: "data_quality_validation", Domain: "data_quality", Evidence: evidence(lower, "data"), SourcePath: entry.NormalizedPath})
		}
		if hasAny(lower, "bug", "failed", "failure", "error", "panic", "regression", "flaky") {
			findings["bugs"] = append(findings["bugs"], Finding{Name: "common_failure_signal", Domain: "bugs", Evidence: evidence(lower, "error"), SourcePath: entry.NormalizedPath})
		}
		if hasAny(lower, "task:", "prompt", "read .bqa", "act as", "please") {
			findings["prompts"] = append(findings["prompts"], Finding{Name: "successful_prompt_candidate", Domain: "prompts", Evidence: evidence(body, "Task:"), SourcePath: entry.NormalizedPath})
		}
	}

	artifacts := []Artifact{
		{Filename: "etl_patterns.yaml", Content: renderFindings("etl_patterns", findings["etl"])},
		{Filename: "graphql_patterns.yaml", Content: renderFindings("graphql_patterns", findings["graphql"])},
		{Filename: "api_patterns.yaml", Content: renderFindings("api_patterns", findings["api"])},
		{Filename: "data_quality_patterns.yaml", Content: renderFindings("data_quality_patterns", findings["data_quality"])},
		{Filename: "common_bugs.yaml", Content: renderFindings("common_bugs", findings["bugs"])},
		{Filename: "successful_prompts.yaml", Content: renderFindings("successful_prompts", findings["prompts"])},
		{Filename: "project_profile.yaml", Content: renderProfile(profile)},
	}

	created := 0
	for _, artifact := range artifacts {
		if err := u.Writer.WriteKnowledgeArtifact(ctx, artifact.Filename, artifact.Content); err != nil {
			return Result{}, err
		}
		created++
	}

	return Result{SessionsProcessed: processed, ArtifactsCreated: created, OutputDir: u.OutputDir}, nil
}

func hasAny(text string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(text, value) {
			return true
		}
	}
	return false
}

func evidence(text string, needle string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	idx := strings.Index(strings.ToLower(text), strings.ToLower(needle))
	if idx < 0 {
		if len(text) > 160 {
			return text[:160]
		}
		return text
	}
	start := idx - 80
	if start < 0 {
		start = 0
	}
	end := idx + 160
	if end > len(text) {
		end = len(text)
	}
	return strings.TrimSpace(text[start:end])
}

func renderFindings(root string, items []Finding) string {
	var b strings.Builder
	b.WriteString(root + ":\n")
	if len(items) == 0 {
		b.WriteString("  []\n")
		return b.String()
	}
	items = uniqueFindings(items)
	for _, item := range items {
		b.WriteString("  - name: " + yamlString(item.Name) + "\n")
		b.WriteString("    domain: " + yamlString(item.Domain) + "\n")
		b.WriteString("    evidence: " + yamlString(item.Evidence) + "\n")
		b.WriteString("    source: " + yamlString(item.SourcePath) + "\n")
	}
	return b.String()
}

func renderProfile(p ProjectProfile) string {
	return fmt.Sprintf("project_profile:\n  sessions_analyzed: %d\n  signals:\n    etl: %d\n    graphql: %d\n    api: %d\n    data_quality: %d\n  maturity: initial\n", p.Sessions, p.ETLSignals, p.GraphQLSignals, p.APISignals, p.DQSignals)
}

func uniqueFindings(items []Finding) []Finding {
	seen := map[string]bool{}
	var out []Finding
	for _, item := range items {
		key := item.Name + "|" + item.SourcePath
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
		if len(out) >= 50 {
			break
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SourcePath < out[j].SourcePath })
	return out
}

func yamlString(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return "\"" + value + "\""
}
