package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestE2EDiscoverIngestBuildProducesKnowledge is an end-to-end synthetic
// regression test for the canonical local workflow:
//
//	bqa discover   (manifest of local AI coding session artifacts)
//	bqa ingest     (normalize synthetic ETL notes into .bqa/input/sessions)
//	bqa build      (extract reusable QA knowledge into .bqa/knowledge/*.yaml)
//
// It runs entirely from t.TempDir() with synthetic, non-private input and
// drives the real cobra commands via their constructors, so it exercises the
// same code path a user gets on the CLI. The test proves that the workflow
// produces the expected .bqa/knowledge/*.yaml artifacts and that each domain
// signal (etl / graphql / api / data quality / failure / reusable prompt) is
// reflected in the generated knowledge.
func TestE2EDiscoverIngestBuildProducesKnowledge(t *testing.T) {
	tmp := t.TempDir()

	notesDir := filepath.Join(tmp, "etl-notes")
	bqaDir := filepath.Join(tmp, ".bqa")
	sessionsDir := filepath.Join(bqaDir, "input", "sessions")
	knowledgeDir := filepath.Join(bqaDir, "knowledge")
	manifestPath := filepath.Join(bqaDir, "input", "sessions", "manifest.json")

	writeSyntheticNotes(t, notesDir)

	// Step 1: discover. Run hermetically (no global/local scan) so the test
	// never touches private user data; this still exercises the discover
	// command end-to-end and writes a manifest under the temp .bqa dir.
	executeCmd(t, discoverCmd(), []string{
		"--global=false",
		"--local=false",
		"--manifest", manifestPath,
	})
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("discover did not produce manifest %q: %v", manifestPath, err)
	}

	// Step 2: ingest synthetic ETL notes into normalized sessions + index.json.
	executeCmd(t, ingestCmd(), []string{
		"--from", notesDir,
		"--base-dir", sessionsDir,
	})

	indexPath := filepath.Join(sessionsDir, "index.json")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("ingest did not produce session index %q: %v", indexPath, err)
	}

	// Step 3: build knowledge artifacts from the normalized sessions.
	executeCmd(t, buildCmd(), []string{
		"--sessions", sessionsDir,
		"--knowledge-dir", knowledgeDir,
	})

	// All expected knowledge artifacts must exist and be non-empty.
	expected := []string{
		"etl_patterns.yaml",
		"graphql_patterns.yaml",
		"api_patterns.yaml",
		"data_quality_patterns.yaml",
		"common_bugs.yaml",
		"successful_prompts.yaml",
		"droid_patterns.yaml",
		"runtime_patterns.yaml",
		"project_profile.yaml",
	}
	for _, name := range expected {
		path := filepath.Join(knowledgeDir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected knowledge artifact %q missing: %v", name, err)
		}
		if info.Size() == 0 {
			t.Fatalf("expected knowledge artifact %q is empty", name)
		}
	}

	// Each domain signal from the synthetic notes must surface in its artifact.
	// We assert the v1 kind plus a populated finding (the artifact is
	// not the empty "patterns: []" placeholder).
	domainChecks := []struct {
		file    string
		rootKey string
	}{
		{"etl_patterns.yaml", "kind: etl_patterns"},
		{"graphql_patterns.yaml", "kind: graphql_patterns"},
		{"api_patterns.yaml", "kind: api_patterns"},
		{"data_quality_patterns.yaml", "kind: data_quality_patterns"},
		{"common_bugs.yaml", "kind: common_bugs"},
		{"successful_prompts.yaml", "kind: successful_prompts"},
	}
	for _, dc := range domainChecks {
		body := readFile(t, filepath.Join(knowledgeDir, dc.file))
		if !strings.Contains(body, dc.rootKey) {
			t.Fatalf("%s missing kind %q, got:\n%s", dc.file, dc.rootKey, body)
		}
		if !strings.Contains(body, "- id:") {
			t.Fatalf("%s has no findings (expected non-empty domain signal), got:\n%s", dc.file, body)
		}
	}

	// project_profile must record at least one analyzed session.
	profile := readFile(t, filepath.Join(knowledgeDir, "project_profile.yaml"))
	if !strings.Contains(profile, "kind: project_profile") {
		t.Fatalf("project_profile.yaml missing kind, got:\n%s", profile)
	}
	if strings.Contains(profile, "sessions_analyzed: 0") || !strings.Contains(profile, "sessions_analyzed:") {
		t.Fatalf("project_profile.yaml should report a non-zero sessions_analyzed, got:\n%s", profile)
	}
}

// executeCmd executes a cobra command with the given args, routing stdout/stderr
// into a buffer, and fails the test on any error. The buffered output is
// returned so callers can assert on it if needed.
func executeCmd(t *testing.T, cmd *cobra.Command, args []string) string {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("%q execute returned error: %v\noutput: %s", cmd.Name(), err, out.String())
	}
	return out.String()
}

// writeSyntheticNotes creates synthetic, non-private ETL note files whose
// content covers every QA domain signal the knowledge extractor looks for:
// ETL reconciliation / row count, a GraphQL query/schema, an API endpoint /
// status code, a data quality null/duplicate check, a common failure signal,
// and a reusable task-prompt candidate (explicit Task:/acceptance structure so
// successful_prompts.yaml is non-empty).
func writeSyntheticNotes(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll notes dir returned error: %v", err)
	}

	notes := map[string]string{
		"etl_reconciliation.md": "" +
			"Task: validate the ETL reconciliation for the nightly Airflow dag run.\n" +
			"The Spark job reads the source table and writes the target table in parquet.\n" +
			"Reconciliation step compares the row count between source and target;\n" +
			"acceptance criteria: row count must match and etl_logs show no failure.\n" +
			"Verify that the pipeline reconciliation passes so that the data is complete.\n",
		"graphql_api.md": "" +
			"Task: add a GraphQL query and update the GraphQL schema for the orders resolver.\n" +
			"We tested the REST API endpoint POST /orders and asserted the http status code 201.\n" +
			"The contract test checks the response payload against the openapi schema.\n" +
			"Acceptance criteria: the GraphQL query returns the order and the endpoint\n" +
			"should return status code 200 for a valid request so that clients can rely on it.\n",
		"data_quality_failure.md": "" +
			"Task: implement a data quality check for the warehouse table.\n" +
			"Run a null check and a duplicate check on the primary key; the dq check must\n" +
			"flag any duplicate rows. During the run the job failed with an exception and\n" +
			"the traceback showed a regression in the data validation step.\n" +
			"Acceptance criteria: data quality validation must pass with no nulls and no\n" +
			"duplicates so that downstream consumers can trust the dataset.\n",
	}

	for name, body := range notes {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatalf("WriteFile %q returned error: %v", name, err)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %q returned error: %v", path, err)
	}
	return string(data)
}
