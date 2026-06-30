# Prompt Evaluation Checklist

Run a prompt through this checklist before adding it to the library or promoting
it into `.bqa/knowledge/successful_prompts.yaml`. Synthetic and safe.

## Quality

- [ ] **Single clear goal.** The prompt asks for one well-defined outcome.
- [ ] **Concrete output contract.** It specifies the exact output shape
      (fields, format, YAML/Markdown/table).
- [ ] **Reusable.** It generalizes beyond one ticket — produces a check,
      workflow, or rule that applies next time.
- [ ] **Evidence-driven.** It requires the model to cite the signal (ticket,
      log, data sample) behind each claim.
- [ ] **Deterministic-enough.** Re-running on the same input yields a stable
      shape (values may vary, structure should not).

## Safety & privacy

- [ ] **No client data.** No customer records, PII, or business-confidential
      values — uses `[placeholder]` instead.
- [ ] **No secrets.** No tokens, passwords, connection strings, API keys.
- [ ] **No internal hostnames / URLs.** Internal infra is parameterized.
- [ ] **No `bqa-brain` / private content.** Only synthetic or public material.
- [ ] **Refusal-safe.** Asking the model to fill placeholders won't pressure it
      to invent sensitive data.

## Fit with BQA-OS

- [ ] **Knowledge-loop aware.** Reusable output maps to a `name` / `domain` /
      `evidence` / `source` finding that `bqa build` could surface.
- [ ] **Copy-paste ready.** A QA engineer can drop it into a pilot with only
      placeholder edits.

## Scoring

Rate each section 0–2 (0 = fails, 1 = partial, 2 = solid). A prompt is
**library-ready at 14/16+** with no zeros in the Safety & privacy section.

---

## Failed Prompt Notes Template

When a prompt does **not** pass, record why so the next attempt is better. Keep
these notes synthetic and safe.

```text
prompt_name: [snake_case id]
date: [YYYY-MM-DD]
intended_goal: [what it was supposed to produce]
what_happened: [actual behavior — vague output / hallucinated data / unsafe / etc.]
failure_mode: [one of: ambiguous-goal, no-output-contract, not-reusable,
               leaked-sensitive, unstable-output, hallucination, other]
evidence: [paste the problematic output snippet, redacted]
fix_hypothesis: [the change you believe will fix it]
status: [retry / abandoned / superseded-by: <prompt_name>]
```

Failed-prompt notes are intentionally **not** promoted to
`successful_prompts.yaml`; they live alongside the library as learning records.
