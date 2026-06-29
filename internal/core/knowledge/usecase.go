package knowledge

import (
	"context"
	"fmt"
	"regexp"
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

	sessions := make([]NormalizedSession, 0, len(index.Entries))
	processed := 0

	for _, entry := range index.Entries {
		body, err := u.Reader.ReadNormalizedSession(ctx, entry.NormalizedPath)
		if err != nil {
			continue
		}
		processed++
		sessions = append(sessions, NormalizedSession{
			SourcePath:         entry.OriginalPath,
			NormalizedPath:     entry.NormalizedPath,
			NormalizedMarkdown: body,
		})
	}

	extraction := Extractor{}.Extract(sessions)
	artifacts := []Artifact{
		{Filename: "etl_patterns.yaml", Content: renderFindings("etl_patterns", extraction.ETLPatterns)},
		{Filename: "graphql_patterns.yaml", Content: renderFindings("graphql_patterns", extraction.GraphQLPatterns)},
		{Filename: "api_patterns.yaml", Content: renderFindings("api_patterns", extraction.APIPatterns)},
		{Filename: "data_quality_patterns.yaml", Content: renderFindings("data_quality_patterns", extraction.DataQualityPatterns)},
		{Filename: "common_bugs.yaml", Content: renderFindings("common_bugs", extraction.CommonBugs)},
		{Filename: "successful_prompts.yaml", Content: renderFindings("successful_prompts", extraction.SuccessfulPrompts)},
		{Filename: "project_profile.yaml", Content: renderProfile(extraction.Profile)},
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
	return hasAny(text, "data quality", "duplicate", "schema drift", "null check", "duplicate check", "row count", "checksum", "not null", "unique constraint", "dq check", "data validation")
}

func isFailureSignal(text string) bool {
	return hasAny(text, "failed", "failure", "error:", "panic", "regression", "flaky", "stack trace", "exception", "traceback")
}

func isPromptSignal(text string) bool {
	return hasAny(text, "task:", "read .bqa", "act as", "please", "your task", "implement", "analyze this repository")
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
	text = strings.TrimSpace(text)
	for _, redaction := range evidenceRedactions {
		text = redaction.re.ReplaceAllString(text, redaction.replacement)
	}
	return text
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

func etlNeedle(text string) string {
	return firstNeedle(text, "airflow", "spark", "hive", "oozie", "etl_logs", "reconciliation", "parquet", "row count", "etl")
}
func graphqlNeedle(text string) string {
	return firstNeedle(text, "graphql query", "graphql mutation", "graphql schema", "graphql resolver", "graphql")
}
func apiNeedle(text string) string {
	return firstNeedle(text, "rest api", "http status", "status code", "endpoint", "contract test", "openapi", "request payload")
}
func dqNeedle(text string) string {
	return firstNeedle(text, "duplicate", "data quality", "schema drift", "null check", "duplicate check", "row count", "checksum", "dq check")
}
func failureNeedle(text string) string {
	return firstNeedle(text, "traceback", "exception", "failed", "failure", "error:", "panic", "regression", "flaky")
}
func promptNeedle(text string) string {
	return firstNeedle(text, "task:", "your task", "read .bqa", "act as", "please", "implement", "analyze this repository")
}
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

type evidenceRedaction struct {
	re          *regexp.Regexp
	replacement string
}

var evidenceRedactions = []evidenceRedaction{
	{re: regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`), replacement: "[REDACTED_EMAIL]"},
	{re: regexp.MustCompile(`(?i)\b(password|token|secret|api[_-]?key)\s*[:=]\s*[^,\s;]+`), replacement: "${1}=[REDACTED]"},
	{re: regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----.*-----END [A-Z ]*PRIVATE KEY-----`), replacement: "[REDACTED_PRIVATE_KEY]"},
}
