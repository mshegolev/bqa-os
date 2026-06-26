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
		"droid":        {},
		"runtime":      {},
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
		sourcePath := strings.ToLower(entry.OriginalPath + " " + entry.NormalizedPath)
		text := cleanEvidenceText(body)

		if isETLSignal(lower, sourcePath) {
			profile.ETLSignals++
			findings["etl"] = append(findings["etl"], Finding{Name: "etl_validation", Domain: "etl", Evidence: evidence(text, etlNeedle(lower)), SourcePath: entry.NormalizedPath})
		}
		if isGraphQLSignal(lower, sourcePath) {
			profile.GraphQLSignals++
			findings["graphql"] = append(findings["graphql"], Finding{Name: "graphql_functional_testing", Domain: "graphql", Evidence: evidence(text, graphqlNeedle(lower)), SourcePath: entry.NormalizedPath})
		}
		if isAPISignal(lower) {
			profile.APISignals++
			findings["api"] = append(findings["api"], Finding{Name: "api_contract_testing", Domain: "api", Evidence: evidence(text, apiNeedle(lower)), SourcePath: entry.NormalizedPath})
		}
		if isDataQualitySignal(lower) {
			profile.DQSignals++
			findings["data_quality"] = append(findings["data_quality"], Finding{Name: "data_quality_validation", Domain: "data_quality", Evidence: evidence(text, dqNeedle(lower)), SourcePath: entry.NormalizedPath})
		}
		if isFailureSignal(lower) {
			findings["bugs"] = append(findings["bugs"], Finding{Name: "common_failure_signal", Domain: "bugs", Evidence: evidence(text, failureNeedle(lower)), SourcePath: entry.NormalizedPath})
		}
		if isPromptSignal(lower) {
			findings["prompts"] = append(findings["prompts"], Finding{Name: "successful_prompt_candidate", Domain: "prompts", Evidence: evidence(text, promptNeedle(lower)), SourcePath: entry.NormalizedPath})
		}
		if isDroidSignal(sourcePath) {
			profile.DroidSignals++
			findings["droid"] = append(findings["droid"], Finding{Name: "factory_droid_session", Domain: "droid", Evidence: evidence(text, droidNeedle(sourcePath)), SourcePath: entry.NormalizedPath})
		}
		if isRuntimeSignal(lower, sourcePath) {
			profile.RuntimeSignals++
			findings["runtime"] = append(findings["runtime"], Finding{Name: "runtime_execution_pattern", Domain: "runtime", Evidence: evidence(text, runtimeNeedle(lower)), SourcePath: entry.NormalizedPath})
		}
	}

	artifacts := []Artifact{
		{Filename: "etl_patterns.yaml", Content: renderFindings("etl_patterns", findings["etl"])},
		{Filename: "graphql_patterns.yaml", Content: renderFindings("graphql_patterns", findings["graphql"])},
		{Filename: "api_patterns.yaml", Content: renderFindings("api_patterns", findings["api"])},
		{Filename: "data_quality_patterns.yaml", Content: renderFindings("data_quality_patterns", findings["data_quality"])},
		{Filename: "common_bugs.yaml", Content: renderFindings("common_bugs", findings["bugs"])},
		{Filename: "successful_prompts.yaml", Content: renderFindings("successful_prompts", findings["prompts"])},
		{Filename: "droid_patterns.yaml", Content: renderFindings("droid_patterns", findings["droid"])},
		{Filename: "runtime_patterns.yaml", Content: renderFindings("runtime_patterns", findings["runtime"])},
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

func isETLSignal(text string, sourcePath string) bool {
	if isMetadataOnly(text) {
		return false
	}
	return hasAny(text, "airflow", "spark", "hive", "oozie", "dag run", "dag_id", "etl_logs", "reconciliation", "source table", "target table", "row count", "parquet", "data pipeline") || strings.Contains(sourcePath, "etl")
}

func isGraphQLSignal(text string, sourcePath string) bool {
	if hasAny(sourcePath, "normalized/droid") {
		return false
	}
	if hasAny(text, "github_graphql_url", "api/graphql", "github api url") {
		return false
	}
	if !strings.Contains(text, "graphql") {
		return false
	}
	return hasAny(
		text,
		"graphql query",
		"graphql mutation",
		"graphql schema",
		"graphql resolver",
		"graphql introspection",
		"schema and operations",
		"queries, mutations",
	)
}

func isAPISignal(text string) bool {
	if hasAny(text, "github_api_url", "github_server_url") {
		return false
	}
	return hasAny(text, "rest api", "http status", "status code", "endpoint", "contract test", "openapi", "swagger", "request payload", "response payload", "post /", "get /")
}

func isDataQualitySignal(text string) bool {
	return hasAny(text, "data quality", "schema drift", "null check", "duplicate check", "row count", "checksum", "not null", "unique constraint", "dq check", "data validation")
}

func isFailureSignal(text string) bool {
	return hasAny(text, "failed", "failure", "error:", "panic", "regression", "flaky", "stack trace", "exception", "traceback")
}

func isPromptSignal(text string) bool {
	return hasAny(text, "task:", "read .bqa", "act as", "please", "your task", "implement", "analyze this repository")
}

func isDroidSignal(sourcePath string) bool {
	return hasAny(sourcePath, "/.factory/", "normalized/droid")
}

func isRuntimeSignal(text string, sourcePath string) bool {
	if hasAny(sourcePath, "normalized/droid") {
		return true
	}
	return hasAny(sourcePath, "normalized/claude", "normalized/codex", "normalized/opencode") && hasAny(text, "tooluse", "tool call", "run command", "sandbox", "approval", "transcript", "agenttype")
}

func isMetadataOnly(text string) bool {
	return hasAny(text, "agenttype", "tooluseid") && !hasAny(text, "airflow", "spark", "hive", "oozie", "etl_logs", "reconciliation", "parquet")
}

func hasAny(text string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(text, value) {
			return true
		}
	}
	return false
}

func cleanEvidenceText(text string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}

func evidence(text string, needle string) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(needle))
	if idx < 0 {
		if len(text) > 220 {
			return text[:220]
		}
		return text
	}
	start := idx - 100
	if start < 0 {
		start = 0
	}
	end := idx + 220
	if end > len(text) {
		end = len(text)
	}
	return strings.TrimSpace(text[start:end])
}

func etlNeedle(text string) string     { return firstNeedle(text, "airflow", "spark", "hive", "oozie", "etl_logs", "reconciliation", "parquet", "row count", "etl") }
func graphqlNeedle(text string) string {
	return firstNeedle(text, "graphql query", "graphql mutation", "graphql schema", "graphql resolver", "graphql")
}
func apiNeedle(text string) string     { return firstNeedle(text, "rest api", "http status", "status code", "endpoint", "contract test", "openapi", "request payload") }
func dqNeedle(text string) string      { return firstNeedle(text, "data quality", "schema drift", "null check", "duplicate check", "row count", "checksum", "dq check") }
func failureNeedle(text string) string { return firstNeedle(text, "traceback", "exception", "failed", "failure", "error:", "panic", "regression", "flaky") }
func promptNeedle(text string) string  { return firstNeedle(text, "task:", "your task", "read .bqa", "act as", "please", "implement", "analyze this repository") }
func droidNeedle(sourcePath string) string {
	if strings.Contains(sourcePath, "/.factory/") {
		return ".factory"
	}
	return "droid"
}
func runtimeNeedle(text string) string { return firstNeedle(text, "tooluse", "tool call", "run command", "sandbox", "approval", "transcript", "agenttype") }

func firstNeedle(text string, values ...string) string {
	for _, value := range values {
		if strings.Contains(text, value) {
			return value
		}
	}
	if len(values) == 0 {
		return ""
	}
	return values[0]
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
	return fmt.Sprintf("project_profile:\n  sessions_analyzed: %d\n  signals:\n    etl: %d\n    graphql: %d\n    api: %d\n    data_quality: %d\n    droid: %d\n    runtime: %d\n  maturity: initial\n", p.Sessions, p.ETLSignals, p.GraphQLSignals, p.APISignals, p.DQSignals, p.DroidSignals, p.RuntimeSignals)
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
