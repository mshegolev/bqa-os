package app

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mshegolev/bqa-os/internal/core/knowledge"
	"github.com/mshegolev/bqa-os/internal/ports"
)

// knowledgeSummary holds the condensed view of a single knowledge artifact.
// It deliberately keeps only counts and a few short, sanitized evidence
// snippets so the generated Codex context never dumps raw YAML or raw session
// content.
type knowledgeSummary struct {
	rootKey string
	count   int
	names   []string
	samples []string
}

// codexKnowledgeSection reads the knowledge artifacts produced by `bqa build`
// via the supplied reader, condenses them, and renders a markdown section to be
// appended to the Codex master context. The second return value reports whether
// any knowledge artifacts were found; when none are present the caller appends a
// graceful "run bqa build" hint instead.
//
// This function only reads knowledge artifacts (it never writes them) and lives
// in the app layer so the core/knowledge package stays untouched.
func codexKnowledgeSection(ctx context.Context, reader ports.KnowledgeArtifactReader) (string, bool) {
	if reader == nil {
		return codexNoKnowledgeSection(), false
	}

	summaries := make(map[string]knowledgeSummary)
	profile := ""
	found := false

	for _, spec := range knowledge.ExpectedArtifacts() {
		content, err := reader.ReadKnowledgeArtifact(ctx, spec.Filename)
		if err != nil {
			continue
		}
		found = true
		if spec.RootKey == "project_profile" {
			profile = summarizeProfile(content)
			continue
		}
		summaries[spec.RootKey] = summarizeFindings(spec.RootKey, content)
	}

	if !found {
		return codexNoKnowledgeSection(), false
	}

	return renderKnowledgeSection(profile, summaries), true
}

// summarizeFindings parses the v1 findings YAML emitted by knowledge.UseCase
// (a `patterns:` list where each entry starts with "- id:" and carries
// "name:/domain:/evidence:/source:/reusable_check:/confidence:" fields) and
// returns a condensed summary: the finding count, distinct names, and a couple
// of short evidence snippets. No raw source paths are retained.
func summarizeFindings(rootKey, content string) knowledgeSummary {
	summary := knowledgeSummary{rootKey: rootKey}
	nameSeen := map[string]bool{}

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "- id:"):
			summary.count++
		case strings.HasPrefix(line, "name:"):
			name := unquoteYAMLScalar(strings.TrimSpace(strings.TrimPrefix(line, "name:")))
			if name != "" && !nameSeen[name] {
				nameSeen[name] = true
				summary.names = append(summary.names, name)
			}
		case strings.HasPrefix(line, "evidence:"):
			ev := unquoteYAMLScalar(strings.TrimSpace(strings.TrimPrefix(line, "evidence:")))
			ev = condenseEvidence(ev)
			if ev != "" && len(summary.samples) < 2 {
				summary.samples = append(summary.samples, ev)
			}
		}
	}

	sort.Strings(summary.names)
	return summary
}

// summarizeProfile extracts the human-readable signal lines from the
// project_profile artifact without echoing the whole YAML document.
func summarizeProfile(content string) string {
	keep := []string{
		"sessions_analyzed",
		"etl",
		"graphql",
		"api",
		"data_quality",
		"droid",
		"runtime",
		"maturity",
	}
	var lines []string
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		for _, key := range keep {
			if strings.HasPrefix(line, key+":") {
				lines = append(lines, line)
				break
			}
		}
	}
	return strings.Join(lines, "; ")
}

// condenseEvidence trims an evidence snippet to a short, single-line preview so
// the master context stays readable and never reproduces large raw blobs.
func condenseEvidence(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	const max = 160
	if len(text) > max {
		return strings.TrimSpace(text[:max]) + "…"
	}
	return text
}

// unquoteYAMLScalar reverses textutil.QuoteYAML's double-quoted output enough to
// recover a readable scalar. It only needs to handle the escapes the writer
// produces (\\, \", \n, \t) plus bare scalars.
func unquoteYAMLScalar(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		inner := value[1 : len(value)-1]
		replacer := strings.NewReplacer(`\\`, `\`, `\"`, `"`, `\n`, " ", `\t`, " ")
		return strings.TrimSpace(replacer.Replace(inner))
	}
	return value
}

// sectionTitles maps artifact root keys to friendly headings used in the
// generated Codex context.
var sectionTitles = []struct {
	rootKey string
	title   string
}{
	{"etl_patterns", "ETL patterns"},
	{"graphql_patterns", "GraphQL patterns"},
	{"api_patterns", "API patterns"},
	{"data_quality_patterns", "Data quality patterns"},
	{"common_bugs", "Common bugs"},
	{"successful_prompts", "Successful prompts"},
	{"droid_patterns", "Droid patterns"},
	{"runtime_patterns", "Runtime patterns"},
}

func renderKnowledgeSection(profile string, summaries map[string]knowledgeSummary) string {
	var b strings.Builder
	b.WriteString("\n## Project QA Knowledge (generated by `bqa build`)\n\n")
	b.WriteString("The following project-specific QA knowledge was extracted from local sessions. ")
	b.WriteString("Use it to ground your QA work in how this team actually tests.\n\n")

	if profile != "" {
		b.WriteString("### Project profile summary\n\n")
		b.WriteString("- " + profile + "\n\n")
	}

	for _, st := range sectionTitles {
		summary, ok := summaries[st.rootKey]
		if !ok {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s\n\n", st.title))
		if summary.count == 0 {
			b.WriteString("- No findings recorded yet.\n\n")
			continue
		}
		b.WriteString(fmt.Sprintf("- %d finding(s) recorded.\n", summary.count))
		if len(summary.names) > 0 {
			b.WriteString("- Patterns: " + strings.Join(summary.names, ", ") + "\n")
		}
		for _, sample := range summary.samples {
			b.WriteString("- Example signal: " + sample + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(codexUsageInstructions())
	b.WriteString(privacyNote())
	return b.String()
}

// codexNoKnowledgeSection is appended when no knowledge artifacts are present so
// the command still works pre-build and points the user at `bqa build`.
func codexNoKnowledgeSection() string {
	var b strings.Builder
	b.WriteString("\n## Project QA Knowledge (generated by `bqa build`)\n\n")
	b.WriteString("No `.bqa/knowledge/*.yaml` artifacts were found yet. ")
	b.WriteString("Run `bqa discover`, `bqa ingest`, then `bqa build` to extract project-specific QA knowledge, ")
	b.WriteString("then re-run `bqa codex` to embed it into this context.\n\n")
	b.WriteString(privacyNote())
	return b.String()
}

func codexUsageInstructions() string {
	return `### How Codex should use this knowledge

- Prefer the ETL, GraphQL, API, and data-quality patterns above when planning QA work.
- For ETL tasks, check partitioning, retries/idempotency, duplicates, schema drift, and data-quality checks.
- Reuse the recorded successful prompts as starting points for similar QA tasks.
- Treat common bugs as a regression checklist before signing off a change.
- When knowledge is thin, inspect the repository and propose a plan before changing code.

`
}

func privacyNote() string {
	return `### Privacy note

- Do not expose raw or private logs, secrets, customer data, or raw session content.
- The summaries above are condensed signals only; never reconstruct or paste raw transcripts into shared output.
`
}
