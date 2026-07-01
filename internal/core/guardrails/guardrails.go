// Package guardrails holds the canonical critical-thinking and memory-safety
// rules that generated/installed BQA agents must follow. It is the single source
// of truth so the bqa build artifact and the bqa codex master context never drift.
package guardrails

// CriticalThinking returns the canonical guardrails markdown. It is static and
// deterministic (no configuration), so callers can embed it verbatim.
func CriticalThinking() string {
	return `## Critical thinking & memory guardrails

Every BQA agent must follow these rules.

### Critical thinking
1. Do not invent information for critical decisions.
2. First use available local project memory and generated BQA artifacts.
3. If local memory is insufficient, use configured trusted sources where available.
4. For current or external facts, require source checking before making recommendations.
5. If evidence is still insufficient, ask a human for a decision instead of guessing.
6. Mark assumptions explicitly.
7. Separate facts, inference, and recommendations.
8. Keep human-in-the-loop review for high-impact QA, product, release, security, compliance, or customer decisions.

### Memory safety
- Never promote private, raw, or unsanitized session data into reusable artifacts.
- Do not leak secrets, credentials, customer data, or private logs into shared output.
- Attribute important knowledge to its source; flag anything unverified.
`
}
