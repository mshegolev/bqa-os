package knowledge

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeReader struct {
	index ports.SessionIndex
	files map[string]string
}

func (f fakeReader) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	return f.index, nil
}

func (f fakeReader) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	return f.files[path], nil
}

type fakeWriter struct {
	files map[string]string
}

func (f fakeWriter) WriteKnowledgeArtifact(ctx context.Context, filename string, content string) error {
	f.files[filename] = content
	return nil
}

func TestUseCaseBuildsKnowledgeArtifacts(t *testing.T) {
	reader := fakeReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "/tmp/session.jsonl", NormalizedPath: "s1.md"},
			{OriginalPath: "/Users/me/.factory/logs/droid-log-single.log", NormalizedPath: "droid.md"},
		}},
		files: map[string]string{
			"s1.md":    "Task: test GraphQL query, REST API endpoint, ETL reconciliation, and null check data validation",
			"droid.md": "Droid mission worker-transcripts run command failed with traceback",
		},
	}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.SessionsProcessed != 2 {
		t.Fatalf("expected 2 processed sessions, got %d", result.SessionsProcessed)
	}
	if result.ArtifactsCreated != 9 {
		t.Fatalf("expected 9 artifacts, got %d", result.ArtifactsCreated)
	}
	if !strings.Contains(writer.files["graphql_patterns.yaml"], "graphql_functional_testing") {
		t.Fatalf("expected graphql finding, got %s", writer.files["graphql_patterns.yaml"])
	}
	if !strings.Contains(writer.files["etl_patterns.yaml"], "etl_validation") {
		t.Fatalf("expected etl finding, got %s", writer.files["etl_patterns.yaml"])
	}
	if !strings.Contains(writer.files["droid_patterns.yaml"], "factory_droid_session") {
		t.Fatalf("expected droid finding, got %s", writer.files["droid_patterns.yaml"])
	}
	if strings.Contains(writer.files["graphql_patterns.yaml"], "github_graphql_url") {
		t.Fatalf("graphql artifact should not be driven by github_graphql_url noise")
	}
}

func TestSuccessfulPromptCapturesReusableCandidate(t *testing.T) {
	strong := "Task: implement a REST API endpoint that validates ETL reconciliation row counts. " +
		"Acceptance criteria: the endpoint should return 200 with a JSON summary and the tests pass " +
		"against the staging pipeline schema."
	reader := fakeReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "/tmp/strong.md", NormalizedPath: "strong.md"},
		}},
		files: map[string]string{"strong.md": strong},
	}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	if _, err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	out := writer.files["successful_prompts.yaml"]
	if !strings.Contains(out, "successful_prompt_candidate") {
		t.Fatalf("expected strong prompt to be captured, got %s", out)
	}
	if !strings.Contains(out, "Acceptance criteria") {
		t.Fatalf("expected captured prompt to retain reusable acceptance text, got %s", out)
	}
	if !strings.Contains(out, "evidence:") || !strings.Contains(out, "source:") {
		t.Fatalf("expected record to include evidence and source, got %s", out)
	}
}

func TestSuccessfulPromptRejectsGenericPoliteText(t *testing.T) {
	reader := fakeReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "/tmp/polite.md", NormalizedPath: "polite.md"},
		}},
		files: map[string]string{
			"polite.md": "Hi there, please could you help me out? Thanks so much, I really appreciate it!",
		},
	}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	if _, err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	out := writer.files["successful_prompts.yaml"]
	if strings.Contains(out, "successful_prompt_candidate") {
		t.Fatalf("generic polite text must not be captured as a prompt, got %s", out)
	}
	if !strings.Contains(out, "[]") {
		t.Fatalf("expected empty successful_prompts list, got %s", out)
	}
}

func TestReusablePromptHeuristics(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "task with domain and acceptance",
			text: "Task: refactor the auth module so that all existing tests pass and coverage is preserved.",
			want: true,
		},
		{
			name: "imperative with domain and output cue",
			text: "Implement a GraphQL resolver for the orders schema; the result should return paginated edges.",
			want: true,
		},
		{
			name: "only pleasantries",
			text: "please could you help, thanks a lot, really appreciate it",
			want: false,
		},
		{
			name: "too short",
			text: "fix it please",
			want: false,
		},
		{
			name: "task intent but no domain or acceptance",
			text: "Task: write something nice for me whenever you have a free moment today, thank you kindly.",
			want: false,
		},
		{
			name: "raw transcript without task structure is bounded",
			text: "user opened the chat and the assistant replied with a long greeting and some unrelated chatter " +
				strings.Repeat("blah ", 200),
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prompt, ok := reusablePrompt(cleanEvidenceText(tc.text))
			if ok != tc.want {
				t.Fatalf("reusablePrompt(%q) ok = %v, want %v (prompt=%q)", tc.text, ok, tc.want, prompt)
			}
			if ok && len(prompt) > promptMaxLen {
				t.Fatalf("captured prompt exceeds max length: %d", len(prompt))
			}
		})
	}
}

func TestUseCaseReturnsErrorWhenIndexedSessionCannotBeRead(t *testing.T) {
	reader := failingReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "synthetic-session.md", NormalizedPath: "normalized/codex/missing.md"},
		}},
		err: errors.New("open normalized/codex/missing.md: no such file or directory"),
	}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	_, err := uc.Run(context.Background())
	if err == nil {
		t.Fatalf("expected unreadable indexed session error")
	}
	if !strings.Contains(err.Error(), "normalized/codex/missing.md") {
		t.Fatalf("expected error to include normalized path, got %q", err.Error())
	}
	if len(writer.files) != 0 {
		t.Fatalf("expected no artifacts when indexed session cannot be read, got %d", len(writer.files))
	}
}

type failingReader struct {
	index ports.SessionIndex
	err   error
}

func (f failingReader) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	return f.index, nil
}

func (f failingReader) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	return "", f.err
}
