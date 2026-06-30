# Knowledge Extraction Prompt

Turn a raw QA session, chat transcript, or set of notes into reusable knowledge
findings that mirror what `bqa build` produces. Synthetic and safe.

## When to use

- After a testing session, to distill what was learned into durable findings.
- To seed `.bqa/knowledge/` artifacts by hand before enough sessions accumulate.

## Template

```text
You are extracting reusable QA knowledge from the material below.

Input:
[paste the session transcript, bug notes, or regression log]

Produce a list of findings. For each finding, output:
- name: a short snake_case identifier (e.g. partition_reconciliation_check)
- domain: one of [etl, graphql, api, data_quality, bugs, prompts, runtime]
- evidence: one sentence quoting or paraphrasing the concrete signal from the
  input that justifies this finding
- source: where it came from (ticket ID, session name, or [UNKNOWN])
- reusable_check: the generalized check, query, or step that would catch this
  issue again, written so it can be copied into a workflow

Rules:
- Only emit a finding if the input contains concrete evidence. No speculation.
- Drop generic, polite, or filler text (e.g. "thanks", "looks good").
- Deduplicate: merge findings with the same name and source.
- Never include client data, secrets, tokens, PII, or internal hostnames.
  Replace with [placeholder].

Output as YAML under a top-level key `successful_prompts:` (or the relevant
domain key) matching this shape:

successful_prompts:
  - name: "..."
    domain: "prompts"
    evidence: "..."
    source: "..."
    reusable_prompt: "..."
```

## Notes

- Output deliberately matches the `bqa build` artifact contract
  (`name` / `domain` / `evidence` / `source`) so extracted findings can be
  appended to `.bqa/knowledge/successful_prompts.yaml`. The `reusable_prompt` /
  `reusable_check` keys are additive and safe to ignore.
- See [`prompt_eval_checklist.md`](prompt_eval_checklist.md) before promoting
  extracted prompts into the shared library.
