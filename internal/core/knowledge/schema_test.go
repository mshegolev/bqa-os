package knowledge

import (
	"strings"
	"testing"
)

func TestFindingIDIsStableAndDomainPrefixed(t *testing.T) {
	f := Finding{Name: "etl_validation", Domain: "etl", SourcePath: "normalized/etl/s1.md"}
	id1 := findingID(f)
	id2 := findingID(f)
	if id1 != id2 {
		t.Fatalf("findingID not deterministic: %q vs %q", id1, id2)
	}
	if !strings.HasPrefix(id1, "etl-") || len(id1) != len("etl-")+8 {
		t.Fatalf("unexpected id shape: %q", id1)
	}
	// Different source => different id.
	if findingID(Finding{Name: "etl_validation", Domain: "etl", SourcePath: "other.md"}) == id1 {
		t.Fatalf("id should change with source")
	}
}

func TestFindingConfidenceBandsBySignalCount(t *testing.T) {
	high := Finding{Domain: "etl", Evidence: "reconciliation row count between source table and target table"}
	if got := findingConfidence(high); got != "high" {
		t.Fatalf("expected high, got %q", got)
	}
	med := Finding{Domain: "api", Evidence: "the endpoint returned a 500 status code"}
	if got := findingConfidence(med); got != "medium" {
		t.Fatalf("expected medium, got %q", got)
	}
	low := Finding{Domain: "graphql", Evidence: "a graphql thing happened"}
	if got := findingConfidence(low); got != "low" {
		t.Fatalf("expected low, got %q", got)
	}
}

func TestReusableCheckIsDomainSpecific(t *testing.T) {
	if got := reusableCheck(Finding{Domain: "etl", Evidence: "row count reconciliation"}); !strings.Contains(got, "row counts") {
		t.Fatalf("etl reusable_check unexpected: %q", got)
	}
	if got := reusableCheck(Finding{Domain: "etl", Evidence: "null check and duplicate keys"}); !strings.Contains(got, "nulls") {
		t.Fatalf("etl null/dup reusable_check unexpected: %q", got)
	}
	prompt := Finding{Domain: "prompts", Evidence: "Task: verify X"}
	if got := reusableCheck(prompt); got != "Task: verify X" {
		t.Fatalf("prompts reusable_check should echo the prompt, got %q", got)
	}
}

func TestDetectedSignalsSortedByCountThenName(t *testing.T) {
	p := ProjectProfile{ETLSignals: 8, GraphQLSignals: 3, APISignals: 3}
	sigs := detectedSignals(p)
	if len(sigs) != 3 {
		t.Fatalf("expected 3 detected domains, got %d", len(sigs))
	}
	if sigs[0].name != "etl" || sigs[1].name != "api" || sigs[2].name != "graphql" {
		t.Fatalf("unexpected order: %+v", sigs)
	}
}
