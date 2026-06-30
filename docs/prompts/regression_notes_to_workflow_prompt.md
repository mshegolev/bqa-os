# Regression Notes → Checklist / Workflow Prompt

Turn ad-hoc regression notes (the scratch list you make during a release) into a
repeatable checklist or workflow. Synthetic and safe.

## When to use

- After a manual regression pass, to make it repeatable.
- Converting a teammate's freeform notes into a shared checklist.

## Template

```text
You are converting freeform regression notes into a repeatable checklist.

Regression notes:
[paste the raw notes — areas touched, things you clicked/checked, anything flaky]

Produce a checklist where each item has:
- area: the feature or surface (e.g. "login", "partition load", "search API")
- check: a single, verifiable action with an expected result
- data: the input shape needed (use [placeholder] for any sensitive value)
- risk: [high / medium / low] — how likely this area is to regress
- automatable: [yes / no] — and if yes, name the test or validation rule

Then group the checklist by risk (high first) and output it as a Markdown table.

Rules:
- One assertion per check. Split compound notes into separate items.
- Drop vague items ("looked fine") unless you can state a concrete expectation.
- Replace client data, secrets, tokens, PII, hostnames with [placeholder].
- End with a "Promote to automation" section listing the high-risk,
  automatable checks that should become tests or validation rules.
```

## Example shape (synthetic)

| area | check | risk | automatable |
| --- | --- | --- | --- |
| partition load | Reload `[date]`; row count equals source extract count | high | yes — reconciliation rule |
| search API | Query `[term]`; results sorted by relevance, no 5xx | medium | yes — contract test |
| login | Valid creds → session; invalid → 401, no lockout bypass | high | yes — auth regression |

## Notes

- The "Promote to automation" section is the reusable artifact — feed it through
  the [knowledge extraction prompt](knowledge_extraction_prompt.md) so it reaches
  `.bqa/knowledge/` (see [`README.md`](README.md)).
