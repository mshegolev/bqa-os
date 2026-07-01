package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/mshegolev/bqa-os/internal/textutil"
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
			return Result{}, fmt.Errorf("read normalized session %q: %w", entry.NormalizedPath, err)
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
		if prompt, ok := reusablePrompt(text); ok {
			findings["prompts"] = append(findings["prompts"], Finding{Name: "successful_prompt_candidate", Domain: "prompts", Evidence: prompt, SourcePath: entry.NormalizedPath})
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

	// content is keyed by the artifact root key (filename without ".yaml").
	content := map[string]string{
		"etl_patterns":          renderFindings("etl_patterns", findings["etl"]),
		"graphql_patterns":      renderFindings("graphql_patterns", findings["graphql"]),
		"api_patterns":          renderFindings("api_patterns", findings["api"]),
		"data_quality_patterns": renderFindings("data_quality_patterns", findings["data_quality"]),
		"common_bugs":           renderFindings("common_bugs", findings["bugs"]),
		"successful_prompts":    renderFindings("successful_prompts", findings["prompts"]),
		"droid_patterns":        renderFindings("droid_patterns", findings["droid"]),
		"runtime_patterns":      renderFindings("runtime_patterns", findings["runtime"]),
		"project_profile":       renderProfile(profile),
	}

	artifacts := make([]Artifact, 0, len(ExpectedArtifacts()))
	for _, spec := range ExpectedArtifacts() {
		artifacts = append(artifacts, Artifact{Filename: spec.Filename, Content: content[spec.RootKey]})
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
	return textutil.HasAny(text, "airflow", "spark", "hive", "oozie", "dag run", "dag_id", "etl_logs", "reconciliation", "source table", "target table", "row count", "parquet", "data pipeline") || strings.Contains(sourcePath, "etl")
}

func isGraphQLSignal(text string, sourcePath string) bool {
	if textutil.HasAny(sourcePath, "normalized/droid") {
		return false
	}
	if textutil.HasAny(text, "github_graphql_url", "api/graphql", "github api url") {
		return false
	}
	if !strings.Contains(text, "graphql") {
		return false
	}
	return textutil.HasAny(
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
	if textutil.HasAny(text, "github_api_url", "github_server_url") {
		return false
	}
	return textutil.HasAny(text, "rest api", "http status", "status code", "endpoint", "contract test", "openapi", "swagger", "request payload", "response payload", "post /", "get /")
}

func isDataQualitySignal(text string) bool {
	return textutil.HasAny(text, "data quality", "schema drift", "null check", "duplicate check", "row count", "checksum", "not null", "unique constraint", "dq check", "data validation")
}

func isFailureSignal(text string) bool {
	return textutil.HasAny(text, "failed", "failure", "error:", "panic", "regression", "flaky", "stack trace", "exception", "traceback")
}

// reusablePrompt extracts the most reusable, actionable prompt candidate from a
// normalized session. It returns the trimmed candidate and true only when the
// text carries enough task structure to be worth re-using later.
//
// A candidate must clear three independent guards:
//   - minimum length (a one-liner like "please help" is too thin to reuse), and
//     a sane maximum so we never copy a long raw transcript verbatim;
//   - at least two task-structure signals out of {task intent / imperative verb,
//     domain or system context, expected output or acceptance cue};
//   - it must not be only pleasantries (polite filler with no task content).
func reusablePrompt(text string) (string, bool) {
	candidate := bestPromptCandidate(text)
	if candidate == "" {
		return "", false
	}
	lower := strings.ToLower(candidate)

	if isOnlyPleasantries(lower) {
		return "", false
	}

	signals := 0
	if hasTaskIntent(lower) {
		signals++
	}
	if hasDomainContext(lower) {
		signals++
	}
	if hasAcceptanceCue(lower) {
		signals++
	}
	if signals < 2 {
		return "", false
	}
	return candidate, true
}

// promptMinLen / promptMaxLen bound a reusable prompt: long enough to carry task
// structure, short enough that we never copy a raw transcript body wholesale.
const (
	promptMinLen = 40
	promptMaxLen = 400
)

// bestPromptCandidate picks the strongest task-shaped segment of the cleaned
// evidence text. It prefers an explicit task marker ("Task:", "Your task",
// "Acceptance criteria") and otherwise falls back to the whole bounded text.
// The returned candidate is always length-bounded so no raw session body leaks
// through verbatim.
func bestPromptCandidate(text string) string {
	text = strings.TrimSpace(text)
	if len(text) < promptMinLen {
		return ""
	}

	lower := strings.ToLower(text)
	for _, marker := range []string{"task:", "your task", "acceptance criteria", "goal:", "objective:"} {
		if idx := strings.Index(lower, marker); idx >= 0 {
			segment := strings.TrimSpace(text[idx:])
			return boundPrompt(segment)
		}
	}
	if !hasTaskIntent(lower) {
		return ""
	}
	return boundPrompt(text)
}

func boundPrompt(segment string) string {
	if len(segment) < promptMinLen {
		return ""
	}
	if r := []rune(segment); len(r) > promptMaxLen {
		// Slice on rune boundaries so a multi-byte character is never split.
		segment = strings.TrimSpace(string(r[:promptMaxLen]))
	}
	return segment
}

// hasTaskIntent detects an explicit task marker or an imperative action verb,
// i.e. the prompt actually asks for something concrete to be done.
func hasTaskIntent(lower string) bool {
	return textutil.HasAny(lower,
		"task:", "your task", "act as", "goal:", "objective:", "read .bqa",
		"implement", "generate", "write", "create", "build", "add ", "refactor",
		"fix ", "design", "analyze", "review", "test ", "validate", "extract",
		"migrate", "update", "verify",
	)
}

// hasDomainContext detects domain or system context that grounds the task in a
// concrete area, so the captured prompt is reusable rather than abstract.
func hasDomainContext(lower string) bool {
	// Keep domain-specific terms only. Generic prose words (test, function, table,
	// module, component, bare "api"/"repo") were too broad and let polite chatter
	// clear the domain gate.
	return textutil.HasAny(lower,
		"etl", "airflow", "spark", "hive", "oozie", "graphql", "rest api",
		"endpoint", "schema", "pipeline", "data quality", "reconciliation",
		"repository", "microservice", "database", "snowflake", "sql",
		"resolver", "mutation", "kafka", "postgres", "warehouse",
	)
}

// hasAcceptanceCue detects an expected-output or acceptance signal: the prompt
// states how to tell the work is done, which is what makes it reusable.
func hasAcceptanceCue(lower string) bool {
	return textutil.HasAny(lower,
		"acceptance criteria", "expected output", "expected result", "should return",
		"must pass", "should pass", "ensure that", "so that", "output format",
		"return format", "make sure", "verify that", "the result should",
		"tests pass", "passing tests", "definition of done", "criteria",
	)
}

// isOnlyPleasantries reports whether the text is polite filler with no task
// content. We strip common courtesy tokens and, if almost nothing meaningful
// remains, reject the candidate.
func isOnlyPleasantries(lower string) bool {
	stripped := lower
	for _, polite := range []string{
		"please", "thanks", "thank you", "could you", "would you", "can you",
		"kindly", "appreciate", "hello", "hi ", "hey", "if you don't mind",
		"i was wondering", "help me", "help ", "sorry",
	} {
		stripped = strings.ReplaceAll(stripped, polite, " ")
	}
	stripped = cleanEvidenceText(stripped)
	// Remove punctuation-only residue.
	stripped = strings.Trim(stripped, " .,!?;:-")
	return len(stripped) < promptMinLen
}

func isDroidSignal(sourcePath string) bool {
	return textutil.HasAny(sourcePath, "/.factory/", "normalized/droid")
}

func isRuntimeSignal(text string, sourcePath string) bool {
	if textutil.HasAny(sourcePath, "normalized/droid") {
		return true
	}
	return textutil.HasAny(sourcePath, "normalized/claude", "normalized/codex", "normalized/opencode") && textutil.HasAny(text, "tooluse", "tool call", "run command", "sandbox", "approval", "transcript", "agenttype")
}

func isMetadataOnly(text string) bool {
	return textutil.HasAny(text, "agenttype", "tooluseid") && !textutil.HasAny(text, "airflow", "spark", "hive", "oozie", "etl_logs", "reconciliation", "parquet")
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

// evidenceLeadIn / evidenceWindow bound the snippet captured around a matched
// signal. The window is wide enough to carry a full QA context sentence
// (reconciliation + row count + schema + partition + scheduler, or the
// null/duplicate/schema-drift/checksum chain) past the first keyword, but is
// hard-capped well below a transcript body so no raw session leaks verbatim.
const (
	evidenceLeadIn = 120
	evidenceWindow = 480
)

func evidence(text string, needle string) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(needle))
	if idx < 0 {
		return boundEvidence(text)
	}
	start := idx - evidenceLeadIn
	if start < 0 {
		start = 0
	}
	end := start + evidenceWindow
	if end > len(text) {
		end = len(text)
	}
	// Snap both offsets to valid rune boundaries so a multi-byte character is
	// never split (real session content contains non-ASCII).
	for start > 0 && !utf8.RuneStart(text[start]) {
		start--
	}
	for end < len(text) && !utf8.RuneStart(text[end]) {
		end++
	}
	return strings.TrimSpace(text[start:end])
}

// boundEvidence caps a snippet at the evidence window on a rune boundary so a
// multi-byte character is never split and no raw transcript is copied whole.
func boundEvidence(text string) string {
	if r := []rune(text); len(r) > evidenceWindow {
		return strings.TrimSpace(string(r[:evidenceWindow]))
	}
	return text
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
	return firstNeedle(text, "data quality", "schema drift", "null check", "duplicate check", "row count", "checksum", "dq check")
}
func failureNeedle(text string) string {
	return firstNeedle(text, "traceback", "exception", "failed", "failure", "error:", "panic", "regression", "flaky")
}
func droidNeedle(sourcePath string) string {
	if strings.Contains(sourcePath, "/.factory/") {
		return ".factory"
	}
	return "droid"
}
func runtimeNeedle(text string) string {
	return firstNeedle(text, "tooluse", "tool call", "run command", "sandbox", "approval", "transcript", "agenttype")
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

func renderFindings(kind string, items []Finding) string {
	var b strings.Builder
	b.WriteString(artifactHeader(kind))
	if len(items) == 0 {
		b.WriteString("patterns: []\n")
		return b.String()
	}
	items = uniqueFindings(items)
	b.WriteString("patterns:\n")
	for _, item := range items {
		b.WriteString("  - id: " + textutil.QuoteYAML(findingID(item)) + "\n")
		b.WriteString("    name: " + textutil.QuoteYAML(item.Name) + "\n")
		b.WriteString("    domain: " + textutil.QuoteYAML(item.Domain) + "\n")
		b.WriteString("    evidence: " + textutil.QuoteYAML(item.Evidence) + "\n")
		b.WriteString("    source: " + textutil.QuoteYAML(item.SourcePath) + "\n")
		b.WriteString("    reusable_check: " + textutil.QuoteYAML(reusableCheck(item)) + "\n")
		b.WriteString("    confidence: " + findingConfidence(item) + "\n")
	}
	return b.String()
}

// artifactHeader renders the common v1 envelope shared by every artifact.
func artifactHeader(kind string) string {
	return fmt.Sprintf("schema_version: %d\nkind: %s\ngenerated_by: %s\n", SchemaVersion, kind, generatedBy())
}

func renderProfile(p ProjectProfile) string {
	var b strings.Builder
	b.WriteString(artifactHeader("project_profile"))
	b.WriteString("profile:\n")
	b.WriteString(fmt.Sprintf("  sessions_analyzed: %d\n", p.Sessions))

	sigs := detectedSignals(p)
	names := make([]string, 0, len(sigs))
	for _, s := range sigs {
		names = append(names, s.name)
	}
	if len(names) == 0 {
		b.WriteString("  domains_detected: []\n")
	} else {
		b.WriteString("  domains_detected: [" + strings.Join(names, ", ") + "]\n")
	}

	b.WriteString("  signals:\n")
	b.WriteString(fmt.Sprintf("    etl: %d\n", p.ETLSignals))
	b.WriteString(fmt.Sprintf("    graphql: %d\n", p.GraphQLSignals))
	b.WriteString(fmt.Sprintf("    api: %d\n", p.APISignals))
	b.WriteString(fmt.Sprintf("    data_quality: %d\n", p.DQSignals))
	b.WriteString(fmt.Sprintf("    droid: %d\n", p.DroidSignals))
	b.WriteString(fmt.Sprintf("    runtime: %d\n", p.RuntimeSignals))

	b.WriteString("  suggested_next_reviews:\n")
	if len(sigs) == 0 {
		b.WriteString("    []\n")
	} else {
		for _, s := range sigs {
			b.WriteString("    - " + textutil.QuoteYAML(fmt.Sprintf("Review %s coverage (%d signals).", s.name, s.count)) + "\n")
		}
	}
	b.WriteString("  maturity: initial\n")
	return b.String()
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
