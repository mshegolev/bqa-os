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

	findings := map[Domain][]Finding{
		DomainETL:         {},
		DomainGraphQL:     {},
		DomainAPI:         {},
		DomainDataQuality: {},
		DomainBugs:        {},
		DomainPrompts:     {},
	}

	processed := 0
	for _, entry := range index.Entries {
		body, err := u.Reader.ReadNormalizedSession(ctx, entry.NormalizedPath)
		if err != nil {
			continue
		}
		processed++

		for _, line := range splitLines(body) {
			line = cleanEvidenceText(line)
			if line == "" {
				continue
			}
			for _, match := range matchLine(line, entry.NormalizedPath) {
				findings[match.Domain] = append(findings[match.Domain], match)
			}
		}
	}

	findings = uniqueAndLimit(findings, 100)
	profile := ProjectProfile{
		Sessions:            processed,
		ETLFindings:         len(findings[DomainETL]),
		GraphQLFindings:     len(findings[DomainGraphQL]),
		APIFindings:         len(findings[DomainAPI]),
		DataQualityFindings: len(findings[DomainDataQuality]),
		BugFindings:         len(findings[DomainBugs]),
		PromptFindings:      len(findings[DomainPrompts]),
	}

	artifacts := []Artifact{
		{Filename: "etl_patterns.yaml", Content: renderFindings("etl_patterns", findings[DomainETL])},
		{Filename: "graphql_patterns.yaml", Content: renderFindings("graphql_patterns", findings[DomainGraphQL])},
		{Filename: "api_patterns.yaml", Content: renderFindings("api_patterns", findings[DomainAPI])},
		{Filename: "data_quality_patterns.yaml", Content: renderFindings("data_quality_patterns", findings[DomainDataQuality])},
		{Filename: "common_bugs.yaml", Content: renderFindings("common_bugs", findings[DomainBugs])},
		{Filename: "successful_prompts.yaml", Content: renderFindings("successful_prompts", findings[DomainPrompts])},
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

var domainKeywords = map[Domain][]string{
	DomainETL: {
		"partition", "row count", "duplicate", "schema drift", "retry", "late-arriving", "aggregation", "null", "timestamp", "timezone",
	},
	DomainGraphQL: {
		"query", "mutation", "resolver", "schema", "nullable", "fragment", "pagination", "authorization",
	},
	DomainAPI: {
		"status code", "contract", "payload", "response", "auth", "timeout", "idempotency",
	},
	DomainDataQuality: {
		"null", "duplicate", "freshness", "completeness", "accuracy", "consistency", "reconciliation",
	},
	DomainBugs: {
		"bug", "defect", "failed", "failure", "error", "exception", "panic", "regression", "flaky", "timeout",
	},
	DomainPrompts: {
		"task:", "prompt", "act as", "please", "your task", "implement", "analyze", "create",
	},
}

var domainNames = map[Domain]string{
	DomainETL:         "etl_pattern",
	DomainGraphQL:     "graphql_pattern",
	DomainAPI:         "api_pattern",
	DomainDataQuality: "data_quality_pattern",
	DomainBugs:        "common_bug_candidate",
	DomainPrompts:     "successful_prompt_candidate",
}

func matchLine(line string, sourcePath string) []Finding {
	lower := strings.ToLower(line)
	matches := make([]Finding, 0)

	for _, domain := range []Domain{DomainETL, DomainGraphQL, DomainAPI, DomainDataQuality, DomainBugs, DomainPrompts} {
		keywords := matchingKeywords(lower, domainKeywords[domain])
		if len(keywords) == 0 {
			continue
		}
		matches = append(matches, Finding{
			Name:       domainNames[domain],
			Domain:     domain,
			Keywords:   keywords,
			Evidence:   trimEvidence(line, 260),
			SourcePath: sourcePath,
		})
	}

	return matches
}

func matchingKeywords(lowerLine string, keywords []string) []string {
	matched := make([]string, 0)
	for _, keyword := range keywords {
		if strings.Contains(lowerLine, keyword) {
			matched = append(matched, keyword)
		}
	}
	return matched
}

func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}

func cleanEvidenceText(text string) string {
	text = strings.ReplaceAll(text, "\t", " ")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}

func trimEvidence(text string, limit int) string {
	text = cleanEvidenceText(text)
	if len(text) <= limit {
		return text
	}
	return strings.TrimSpace(text[:limit]) + "..."
}

func uniqueAndLimit(findings map[Domain][]Finding, limit int) map[Domain][]Finding {
	out := make(map[Domain][]Finding, len(findings))
	for domain, items := range findings {
		seen := map[string]bool{}
		for _, item := range items {
			key := string(item.Domain) + "|" + item.SourcePath + "|" + item.Evidence
			if seen[key] {
				continue
			}
			seen[key] = true
			out[domain] = append(out[domain], item)
			if len(out[domain]) >= limit {
				break
			}
		}
		sort.Slice(out[domain], func(i, j int) bool {
			if out[domain][i].SourcePath == out[domain][j].SourcePath {
				return out[domain][i].Evidence < out[domain][j].Evidence
			}
			return out[domain][i].SourcePath < out[domain][j].SourcePath
		})
	}
	return out
}

func renderFindings(root string, items []Finding) string {
	var b strings.Builder
	b.WriteString(root + ":\n")
	if len(items) == 0 {
		b.WriteString("  []\n")
		return b.String()
	}
	for i, item := range items {
		b.WriteString("  - id: " + yamlString(fmt.Sprintf("%s_%03d", item.Domain, i+1)) + "\n")
		b.WriteString("    name: " + yamlString(item.Name) + "\n")
		b.WriteString("    domain: " + yamlString(string(item.Domain)) + "\n")
		b.WriteString("    keywords:\n")
		for _, keyword := range item.Keywords {
			b.WriteString("      - " + yamlString(keyword) + "\n")
		}
		b.WriteString("    evidence: " + yamlString(item.Evidence) + "\n")
		b.WriteString("    source: " + yamlString(item.SourcePath) + "\n")
	}
	return b.String()
}

func renderProfile(p ProjectProfile) string {
	return fmt.Sprintf("project_profile:\n  sessions_analyzed: %d\n  artifacts:\n    - etl_patterns.yaml\n    - graphql_patterns.yaml\n    - api_patterns.yaml\n    - data_quality_patterns.yaml\n    - common_bugs.yaml\n    - successful_prompts.yaml\n    - project_profile.yaml\n  findings:\n    etl: %d\n    graphql: %d\n    api: %d\n    data_quality: %d\n    common_bugs: %d\n    successful_prompts: %d\n  maturity: initial\n", p.Sessions, p.ETLFindings, p.GraphQLFindings, p.APIFindings, p.DataQualityFindings, p.BugFindings, p.PromptFindings)
}

func yamlString(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return "\"" + value + "\""
}
