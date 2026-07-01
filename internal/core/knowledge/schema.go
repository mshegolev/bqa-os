package knowledge

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
	"github.com/mshegolev/bqa-os/internal/version"
)

// generatedBy is the provenance stamp written as `generated_by`. It is "bqa dev"
// in dev/test builds (deterministic) and "bqa vX.Y.Z" in a release build.
func generatedBy() string { return "bqa " + version.Version }

// findingID returns a stable, content-derived id "<domain>-<8 hex>". It is stable
// when other findings are inserted or removed, so artifacts diff and merge cleanly.
func findingID(f Finding) string {
	sum := sha256.Sum256([]byte(f.Domain + "|" + f.Name + "|" + f.SourcePath))
	return f.Domain + "-" + hex.EncodeToString(sum[:])[:8]
}

// domainKeywords are the per-domain signal words used to gauge how many
// corroborating signals appear in a finding's evidence.
var domainKeywords = map[string][]string{
	"etl":          {"airflow", "spark", "hive", "oozie", "dag", "reconciliation", "source table", "target table", "row count", "parquet", "pipeline", "partition", "schedule"},
	"graphql":      {"graphql", "query", "mutation", "schema", "resolver", "variables", "pagination", "auth", "error"},
	"api":          {"rest api", "http status", "status code", "endpoint", "contract", "openapi", "swagger", "request", "response"},
	"data_quality": {"data quality", "schema drift", "null", "duplicate", "row count", "checksum", "unique", "validation"},
	"bugs":         {"failed", "failure", "error", "panic", "regression", "flaky", "stack trace", "exception", "traceback"},
	"runtime":      {"runtime", "execution", "exit code", "stdout", "stderr", "command"},
	"droid":        {"droid", "factory", "automation", "agent"},
	"prompts":      {"task", "goal", "acceptance", "implement", "verify", "context"},
}

// findingConfidence returns low/medium/high by counting distinct domain keywords
// present in the finding's evidence. It is a heuristic, not a probability.
func findingConfidence(f Finding) string {
	lower := strings.ToLower(f.Evidence)
	n := 0
	for _, kw := range domainKeywords[f.Domain] {
		if strings.Contains(lower, kw) {
			n++
		}
	}
	switch {
	case n >= 3:
		return "high"
	case n == 2:
		return "medium"
	default:
		return "low"
	}
}

// reusableCheck returns a per-domain check candidate — a suggestion for a human
// to review, not an extracted command.
func reusableCheck(f Finding) string {
	lower := strings.ToLower(f.Evidence)
	switch f.Domain {
	case "etl":
		if textutil.HasAny(lower, "null", "duplicate") {
			return "assert no unexpected nulls or duplicate keys"
		}
		return "compare source vs target row counts for the window"
	case "graphql":
		return "assert query/mutation response shape and error handling"
	case "api":
		return "assert endpoint status code and response contract"
	case "data_quality":
		return "assert null / duplicate / schema-drift rules pass"
	case "bugs":
		return "add a regression check reproducing the failure signal"
	case "prompts":
		return f.Evidence // the reusable prompt text itself
	case "runtime":
		return "assert the runtime command exits cleanly and emits expected output"
	case "droid":
		return "capture the automation step as a repeatable check"
	default:
		return "add a check that reproduces this signal"
	}
}

// domainSignal pairs a profile domain with its signal count.
type domainSignal struct {
	name  string
	count int
}

// detectedSignals returns the domains with signals > 0, ordered by count
// descending then name ascending (deterministic).
func detectedSignals(p ProjectProfile) []domainSignal {
	all := []domainSignal{
		{"etl", p.ETLSignals}, {"graphql", p.GraphQLSignals}, {"api", p.APISignals},
		{"data_quality", p.DQSignals}, {"droid", p.DroidSignals}, {"runtime", p.RuntimeSignals},
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].count != all[j].count {
			return all[i].count > all[j].count
		}
		return all[i].name < all[j].name
	})
	out := make([]domainSignal, 0, len(all))
	for _, d := range all {
		if d.count > 0 {
			out = append(out, d)
		}
	}
	return out
}
