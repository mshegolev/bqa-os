package knowledge

import (
	"context"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// The fixtures in this file are 100% synthetic. They contain no real customer
// data, session logs, credentials, or PII. They exist to prove that the
// Knowledge Extractor produces QA-credible artifacts: each generated finding
// must carry enough domain context to answer "what is being tested, what risk
// is reduced, and what evidence was found" -- not merely echo a keyword.

// runExtractor wires a set of synthetic normalized sessions through the
// UseCase and returns the rendered artifact map keyed by filename.
func runExtractor(t *testing.T, sessions map[string]string) map[string]string {
	t.Helper()

	entries := make([]ports.SessionIndexEntry, 0, len(sessions))
	for path := range sessions {
		entries = append(entries, ports.SessionIndexEntry{
			OriginalPath:   "synthetic/" + path,
			NormalizedPath: path,
		})
	}

	reader := fakeReader{
		index: ports.SessionIndex{Entries: entries},
		files: sessions,
	}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	if _, err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	return writer.files
}

// containsAll fails the test if any expected substring is missing from the
// rendered artifact. The check is case-insensitive because evidence text is
// preserved verbatim from the (mixed-case) source.
func assertArtifactContains(t *testing.T, artifact, body string, expected ...string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, want := range expected {
		if !strings.Contains(lower, strings.ToLower(want)) {
			t.Fatalf("%s artifact missing QA context %q\n--- artifact ---\n%s", artifact, want, body)
		}
	}
}

func assertArtifactExcludes(t *testing.T, artifact, body string, banned ...string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, bad := range banned {
		if strings.Contains(lower, strings.ToLower(bad)) {
			t.Fatalf("%s artifact unexpectedly contains %q\n--- artifact ---\n%s", artifact, bad, body)
		}
	}
}

// TestETLArtifactCarriesReconciliationContext covers the Big Data / ETL QA
// scenario. A credible ETL finding must reach a QA lead with reconciliation,
// row-count, schema, partition and scheduler context, not just the word
// "airflow".
func TestETLArtifactCarriesReconciliationContext(t *testing.T) {
	fixture := "Validated the nightly Airflow DAG run for the billing pipeline. " +
		"Ran source-to-target reconciliation: compared row count between the " +
		"source table and the target table, checked the partition for 2026-01-05, " +
		"verified the schema matches, and confirmed the scheduler triggered on time."

	files := runExtractor(t, map[string]string{"etl_session.md": fixture})
	etl := files["etl_patterns.yaml"]

	assertArtifactContains(t, "etl", etl, "etl_validation", "domain: \"etl\"")
	// What evidence was found: reconciliation + row count + schema + partition +
	// scheduler context all survive into the rendered evidence window.
	assertArtifactContains(t, "etl", etl,
		"reconciliation", "row count", "source table", "target table",
		"partition", "schema", "scheduler",
	)
}

// TestGraphQLArtifactCarriesOperationContext covers the GraphQL functional QA
// scenario. The finding must surface query/mutation/schema/resolver/variables/
// auth/pagination and error-shape context.
func TestGraphQLArtifactCarriesOperationContext(t *testing.T) {
	fixture := "Functional testing of the GraphQL API. Exercised a GraphQL query " +
		"and a GraphQL mutation against the orders graphql schema and its resolver. " +
		"Passed variables for pagination (first/after cursor), asserted the auth " +
		"header is required, and validated the error shape returned for an " +
		"unauthorized request."

	files := runExtractor(t, map[string]string{"graphql_session.md": fixture})
	gql := files["graphql_patterns.yaml"]

	assertArtifactContains(t, "graphql", gql, "graphql_functional_testing", "domain: \"graphql\"")
	assertArtifactContains(t, "graphql", gql,
		"graphql query", "graphql mutation", "schema", "resolver",
		"variables", "pagination", "auth", "error shape",
	)
	// Guard: github graphql URL noise must never masquerade as a GraphQL finding.
	assertArtifactExcludes(t, "graphql", gql, "github_graphql_url")
}

// TestAPIArtifactCarriesContractContext covers the API regression scenario. The
// finding must include endpoint/request/response/status-code/contract and
// regression context.
func TestAPIArtifactCarriesContractContext(t *testing.T) {
	fixture := "API regression suite for the payments service. Sent a request " +
		"payload to the POST /v1/charges endpoint and inspected the response " +
		"payload. Asserted the HTTP status code is 201, validated the response " +
		"against the OpenAPI contract test, and re-ran the regression to catch " +
		"any contract drift."

	files := runExtractor(t, map[string]string{"api_session.md": fixture})
	api := files["api_patterns.yaml"]

	assertArtifactContains(t, "api", api, "api_contract_testing", "domain: \"api\"")
	assertArtifactContains(t, "api", api,
		"endpoint", "request", "response", "status code", "contract", "regression",
	)
}

// TestDataQualityArtifactCarriesValidationContext covers the data quality
// validation scenario. The finding must include nulls/duplicates/schema-drift/
// row-count/uniqueness/checksum context.
func TestDataQualityArtifactCarriesValidationContext(t *testing.T) {
	fixture := "Ran the data quality validation suite on the customer dimension. " +
		"not null check on the email column, duplicate check on the natural key " +
		"(unique constraint), compared row count against the prior load, detected " +
		"schema drift in a renamed column, and recomputed the checksum to confirm " +
		"the payload was unchanged."

	files := runExtractor(t, map[string]string{"dq_session.md": fixture})
	dq := files["data_quality_patterns.yaml"]

	assertArtifactContains(t, "data_quality", dq, "data_quality_validation", "domain: \"data_quality\"")
	assertArtifactContains(t, "data_quality", dq,
		"not null", "duplicate check", "unique constraint", "row count",
		"schema drift", "checksum",
	)
}

// TestCommonBugsArtifactCapturesFailureWithContext checks that a bug finding
// captures both a failure signal and the surrounding source context, so a QA
// lead can recognise the failure pattern rather than reading a bare "failed".
func TestCommonBugsArtifactCapturesFailureWithContext(t *testing.T) {
	fixture := "The reconciliation job failed with a NullPointerException in the " +
		"aggregation step. Stack trace points at the partition writer; the spark " +
		"task threw an exception while writing the parquet output for the orders " +
		"table."

	files := runExtractor(t, map[string]string{"bug_session.md": fixture})
	bugs := files["common_bugs.yaml"]

	assertArtifactContains(t, "common_bugs", bugs, "common_failure_signal", "domain: \"bugs\"")
	// Failure signal plus enough source context to identify the pattern.
	assertArtifactContains(t, "common_bugs", bugs, "failed", "exception", "partition")
}

// TestSuccessfulPromptsRejectPolitePassThroughDomain re-asserts the no-noise
// guarantee at the QA-pass level: random polite text with no task structure
// must never be promoted to a reusable prompt, even when it is processed
// alongside legitimate domain sessions.
func TestSuccessfulPromptsRejectPolitePassThroughDomain(t *testing.T) {
	files := runExtractor(t, map[string]string{
		"polite.md": "Hi team, thanks so much for everything today, you are all " +
			"wonderful and I really appreciate the hard work, have a great evening!",
		"strong.md": "Task: validate the ETL reconciliation row count for the " +
			"billing pipeline; the result should return a matching count and the " +
			"tests pass against the staging schema.",
	})
	prompts := files["successful_prompts.yaml"]

	// The real, task-structured prompt is captured...
	assertArtifactContains(t, "successful_prompts", prompts, "successful_prompt_candidate", "strong.md")
	// ...but the polite filler never is.
	assertArtifactExcludes(t, "successful_prompts", prompts, "polite.md", "wonderful")
}

// TestArtifactsNeverLeakRawTranscriptVerbatim is the no-raw-data guardrail: a
// long synthetic transcript must be bounded in every artifact it feeds, so the
// extractor never copies a whole session body into the knowledge base. The
// guarantee is a hard length cap on each evidence snippet, not a keyword count.
func TestArtifactsNeverLeakRawTranscriptVerbatim(t *testing.T) {
	const phrase = "verbose internal chatter about the run that should never be copied wholesale. "
	huge := "Task: review the ETL reconciliation pipeline. " + strings.Repeat(phrase, 200)
	// Sanity-check the fixture really is far larger than any single snippet.
	if len(huge) <= evidenceWindow*4 {
		t.Fatalf("fixture too small to exercise the bound: %d", len(huge))
	}

	files := runExtractor(t, map[string]string{"huge.md": huge})

	for name, body := range files {
		for _, line := range strings.Split(body, "\n") {
			value, ok := strings.CutPrefix(strings.TrimSpace(line), "evidence: ")
			if !ok {
				continue
			}
			// Each evidence value is a single bounded snippet, never the whole
			// multi-KB transcript. Allow generous YAML-quoting overhead.
			if len([]rune(value)) > evidenceWindow+8 {
				t.Fatalf("%s evidence snippet exceeds the bound (%d runes): possible raw-transcript leak", name, len([]rune(value)))
			}
		}
	}
}
