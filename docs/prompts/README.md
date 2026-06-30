# Prompt Library (MVP)

Reusable, privacy-safe prompt templates for BQA-OS QA workflows. Copy any of
these into a pilot workflow, an LLM chat, or an agent system prompt and adapt the
bracketed placeholders to your context.

Everything here is **synthetic and safe** — no client data, no internal
credentials, no `bqa-brain` content. The records are illustrative shapes, not
real captured sessions.

## What's in here

| File | Purpose |
| --- | --- |
| [`qa_assistant_system_prompt.md`](qa_assistant_system_prompt.md) | Baseline system prompt for an AI QA assistant. |
| [`knowledge_extraction_prompt.md`](knowledge_extraction_prompt.md) | Turn raw QA sessions/notes into reusable knowledge findings. |
| [`bug_report_to_workflow_prompt.md`](bug_report_to_workflow_prompt.md) | Convert a one-off bug report into a standardized, reusable workflow. |
| [`regression_notes_to_workflow_prompt.md`](regression_notes_to_workflow_prompt.md) | Turn ad-hoc regression notes into a repeatable checklist/workflow. |
| [`graphql_functional_qa_prompt.md`](graphql_functional_qa_prompt.md) | Functional QA prompt for GraphQL APIs. |
| [`etl_data_quality_qa_prompt.md`](etl_data_quality_qa_prompt.md) | Data-quality QA prompt for ETL / Big Data pipelines. |
| [`prompt_eval_checklist.md`](prompt_eval_checklist.md) | Checklist to evaluate a prompt before adding it to the library, plus a failed-prompt notes template. |

Synthetic example artifact:

- [`../../examples/knowledge/successful_prompts.yaml`](../../examples/knowledge/successful_prompts.yaml)

## Relation to `bqa build` and `.bqa/knowledge/successful_prompts.yaml`

`bqa build` analyzes normalized QA sessions and renders knowledge artifacts into
`.bqa/knowledge/`. One of those artifacts is `successful_prompts.yaml`: prompts
that proved reusable across sessions are surfaced as `successful_prompt_candidate`
findings under a top-level `successful_prompts:` key.

The artifact shape produced by `bqa build` (see
`internal/core/knowledge/usecase.go`, `renderFindings`) is a list of records with
`name`, `domain`, `evidence`, and `source`:

```yaml
successful_prompts:
  - name: "successful_prompt_candidate"
    domain: "prompts"
    evidence: "task: validate etl reconciliation for the daily partition"
    source: "normalized/etl/session-4.txt"
```

This Prompt Library is the **human-curated counterpart** to that
machine-generated artifact:

- `bqa build` **discovers** prompts that worked, bottom-up, from real sessions and
  writes them to `.bqa/knowledge/successful_prompts.yaml`.
- This library **publishes** vetted, generalized templates, top-down, that QA can
  copy into pilots without waiting for enough sessions to accumulate.

`examples/knowledge/successful_prompts.yaml` mirrors the `bqa build` artifact
shape so you can see what a populated knowledge file looks like, and so tooling
that reads the real artifact can be exercised against a safe synthetic fixture.
The example extends the core four fields with optional `reusable_prompt` and
`reusable_check` text — additive keys that downstream consumers can ignore while
still parsing the base `name`/`domain`/`evidence`/`source` contract.

## How to use

1. Pick the prompt that matches your QA task.
2. Copy the template block into your tool of choice.
3. Replace every `[placeholder]` with concrete, **non-sensitive** values.
4. Run it through [`prompt_eval_checklist.md`](prompt_eval_checklist.md) before
   promoting it into a shared workflow or the knowledge base.
