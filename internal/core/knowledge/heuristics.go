package knowledge

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var promptLinePattern = regexp.MustCompile(`(?i)^\s*(prompt|user|request|ask)\s*:\s*(.+)$`)

func detectDomains(content string) []Domain {
	lower := strings.ToLower(content)
	var domains []Domain
	if containsAny(lower, "etl", "spark", "airflow", "pipeline", "dag", "warehouse", "bigquery", "snowflake", "databricks", "source-to-target", "source to target") {
		domains = append(domains, DomainETL)
	}
	if containsAny(lower, "graphql", "resolver", "schema", "fragment", "mutation") {
		domains = append(domains, DomainGraphQL)
	}
	if containsAny(lower, "rest api", "api contract", "http", "endpoint", "status code", "postman", "openapi", "swagger") {
		domains = append(domains, DomainAPI)
	}
	if containsAny(lower, "data quality", "dq", "null", "duplicate", "freshness", "reconciliation", "schema drift", "row count") {
		domains = append(domains, DomainDataQuality)
	}
	return domains
}

func extractPatterns(session Session, domain Domain) []KnowledgeArtifact {
	lower := strings.ToLower(session.Content)
	var artifacts []KnowledgeArtifact
	switch domain {
	case DomainETL:
		if containsAny(lower, "row count", "reconciliation", "source", "target") {
			artifacts = append(artifacts, newArtifact(ArtifactKindPattern, DomainETL, "Validate ETL source-to-target reconciliation", "Compare source and target row counts, key totals, and transformation-sensitive fields after pipeline runs.", session, []string{"etl", "reconciliation", "source-to-target"}))
		}
		if containsAny(lower, "partition", "incremental", "backfill") {
			artifacts = append(artifacts, newArtifact(ArtifactKindPattern, DomainETL, "Check partition and incremental load boundaries", "Validate partition filters, incremental windows, and backfill behavior to avoid missing or duplicated data.", session, []string{"etl", "incremental-load", "partitioning"}))
		}
	case DomainGraphQL:
		if containsAny(lower, "schema", "field", "type", "resolver", "nullability") {
			artifacts = append(artifacts, newArtifact(ArtifactKindPattern, DomainGraphQL, "Validate GraphQL schema and resolver behavior together", "Check field availability, nullability, resolver output, and error behavior for realistic GraphQL queries.", session, []string{"graphql", "schema", "resolver"}))
		}
		if containsAny(lower, "mutation", "side effect", "state") {
			artifacts = append(artifacts, newArtifact(ArtifactKindPattern, DomainGraphQL, "Verify GraphQL mutation side effects", "Validate mutation response payloads together with persisted state and follow-up query behavior.", session, []string{"graphql", "mutation", "state"}))
		}
	case DomainAPI:
		if containsAny(lower, "status code", "headers", "payload", "contract", "response") {
			artifacts = append(artifacts, newArtifact(ArtifactKindPattern, DomainAPI, "Validate API contract at response boundary", "Check status codes, headers, response shape, required fields, and error payload consistency.", session, []string{"api", "contract", "response-validation"}))
		}
	case DomainDataQuality:
		if containsAny(lower, "null", "duplicate", "freshness", "schema drift", "row count", "reconciliation") {
			artifacts = append(artifacts, newArtifact(ArtifactKindPattern, DomainDataQuality, "Run core data quality checks before sign-off", "Validate nulls, duplicates, freshness, schema drift, and business-critical consistency rules.", session, []string{"data-quality", "freshness", "schema-drift"}))
		}
	}
	return artifacts
}

func extractCommonBugs(session Session, domain Domain) []KnowledgeArtifact {
	lower := strings.ToLower(session.Content)
	if !containsAny(lower, "bug", "failed", "failure", "error", "regression", "flake", "flaky", "timeout", "mismatch") {
		return nil
	}
	title := "Session contains recurring QA failure signals"
	tags := []string{"bug", "failure-signal"}
	switch domain {
	case DomainETL:
		title = "ETL jobs may fail because of mismatched counts or incremental windows"
		tags = []string{"etl", "bug", "reconciliation"}
	case DomainGraphQL:
		title = "GraphQL failures often come from schema, resolver, or nullability mismatches"
		tags = []string{"graphql", "bug", "schema"}
	case DomainAPI:
		title = "API regressions often appear at contract boundaries"
		tags = []string{"api", "bug", "contract"}
	case DomainDataQuality:
		title = "Data quality bugs often involve nulls, duplicates, freshness, or schema drift"
		tags = []string{"data-quality", "bug", "dq"}
	}
	return []KnowledgeArtifact{newArtifact(ArtifactKindCommonBug, domain, title, "Review this session for reusable failure patterns and add regression coverage where the same signal appears repeatedly.", session, tags)}
}

func extractSuccessfulPrompts(session Session, domain Domain) []KnowledgeArtifact {
	var artifacts []KnowledgeArtifact
	for _, line := range strings.Split(session.Content, "\n") {
		matches := promptLinePattern.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}
		prompt := strings.TrimSpace(matches[2])
		if len(prompt) < 20 {
			continue
		}
		artifacts = append(artifacts, KnowledgeArtifact{Kind: ArtifactKindSuccessfulPrompt, Domain: domain, Title: "Reusable QA prompt", Summary: prompt, Evidence: []string{evidenceLine(session)}, SessionIDs: []string{session.ID}, Tags: []string{"prompt", string(domain)}})
	}
	return artifacts
}

func buildProjectProfile(sessions []Session) KnowledgeArtifact {
	domainCounts := map[Domain]int{}
	for _, session := range sessions {
		for _, domain := range detectDomains(session.Content) {
			domainCounts[domain]++
		}
	}
	var domains []Domain
	for domain := range domainCounts {
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i] < domains[j] })
	var tags []string
	var evidence []string
	for _, domain := range domains {
		tags = append(tags, string(domain))
		evidence = append(evidence, fmt.Sprintf("%s sessions: %d", domain, domainCounts[domain]))
	}
	if len(tags) == 0 {
		tags = []string{"general"}
		evidence = []string{"No dominant QA domain detected yet"}
	}
	return KnowledgeArtifact{Kind: ArtifactKindProjectProfile, Domain: DomainGeneral, Title: "Initial project QA profile", Summary: "Initial profile inferred from normalized QA sessions. This artifact is heuristic and should improve as more sessions are ingested.", Evidence: evidence, Tags: tags}
}

func newArtifact(kind ArtifactKind, domain Domain, title, summary string, session Session, tags []string) KnowledgeArtifact {
	return KnowledgeArtifact{Kind: kind, Domain: domain, Title: title, Summary: summary, Evidence: []string{evidenceLine(session)}, SessionIDs: []string{session.ID}, Tags: tags}
}

func evidenceLine(session Session) string {
	if strings.TrimSpace(session.Title) != "" {
		return fmt.Sprintf("%s (%s)", strings.TrimSpace(session.Title), session.ID)
	}
	if strings.TrimSpace(session.Path) != "" {
		return fmt.Sprintf("%s (%s)", strings.TrimSpace(session.Path), session.ID)
	}
	return session.ID
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
