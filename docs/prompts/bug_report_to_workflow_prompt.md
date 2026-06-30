# Bug Report → Reusable Workflow Prompt

Convert a one-off bug report into a standardized, reusable QA workflow so the
same class of defect is caught automatically next time. Synthetic and safe.

## When to use

- A bug slipped through; you want a repeatable check, not just a fix.
- Standardizing inconsistent bug reports into a common shape.

## Template

```text
You are turning a single bug report into a reusable QA workflow.

Bug report:
[paste the bug report — symptom, steps, environment, expected vs actual]

Produce:
1. Standardized summary
   - title: imperative, specific (e.g. "Duplicate rows on daily partition reload")
   - severity: [blocker / critical / major / minor]
   - component: [system / service / pipeline]
   - root-cause hypothesis: one sentence, marked [HYPOTHESIS] if unconfirmed

2. Minimal reproduction
   - Numbered steps with concrete (non-sensitive) inputs.
   - Expected result vs actual result.

3. Reusable workflow
   - precondition: data/state required before the check
   - check: the assertion, query, or validation rule that detects this bug
   - frequency: [per-release / nightly / on-change]
   - owner: [role]

4. Regression guard
   - The single automated test or validation rule that would have caught this.

Rules:
- Replace any client data, secrets, tokens, PII, or hostnames with [placeholder].
- If a field is unknown, write [UNKNOWN] rather than guessing.
- The "check" must be copy-pasteable into a workflow with minimal edits.
```

## Example shape (synthetic)

```text
title: Duplicate rows on daily partition reload
severity: major
component: [etl-pipeline]
check: |
  SELECT partition_date, COUNT(*) c, COUNT(DISTINCT business_key) d
  FROM [table]
  WHERE partition_date = '[date]'
  GROUP BY partition_date
  HAVING c <> d;  -- expect zero rows
frequency: per-release
```

## Notes

- The "Reusable workflow" and "Regression guard" blocks are the reusable
  artifact — capture them via the [knowledge extraction prompt](knowledge_extraction_prompt.md)
  so they land in `.bqa/knowledge/` (see [`README.md`](README.md)).
