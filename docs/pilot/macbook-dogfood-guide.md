# MacBook dogfood guide: run BQA-OS on 5 local ETL checks

A short, copy-pasteable guide for running BQA-OS locally on a small,
**synthetic** ETL corpus. By the end you will have turned a handful of
sanitized ETL notes into reusable QA knowledge and a ready-to-use Codex master
context — all on your own MacBook, with no data leaving the machine.

This guide uses only **synthetic** examples. Never use real client logs, names,
or secrets while dogfooding (see [Privacy](#privacy-never-commit-this)).

Related reading:

- [Knowledge Extractor](../knowledge-extractor.md) — what `bqa build` produces.
- [Knowledge Review Checklist](../knowledge-review-checklist.md) — how to review
  generated `.bqa/knowledge/*.yaml`.
- [Runtime Adapters](../runtimes.md) — Codex / Claude Code / OpenCode targets.

---

## 1. Prerequisites

- **macOS** with a terminal.
- **Go** (used to build the `bqa` CLI). The installer can check the minimum
  version before doing any build work.

- **git** and a local checkout of this repository:

  ```bash
  git clone git@github.com:mshegolev/bqa-os.git
  cd bqa-os
  bash install.sh --check-go
  ```

---

## 2. Build (or install) the `bqa` CLI

From a clean checkout, install the binary:

```bash
bash install.sh
export PATH="$HOME/.local/bin:$PATH"
bqa --help
```

Expected: the CLI help listing the available commands (`init`, `discover`,
`ingest`, `build`, `codex`, `doctor`, `brain`, …).

> If `bash install.sh --check-go` fails, install or upgrade Go with Homebrew:
> `brew install go` for a missing toolchain, or `brew upgrade go` for an old
> version.

Confirm the binary works:

```bash
bqa version
```

---

## 3. Work inside a throwaway project

Dogfood in a scratch directory so you never touch a real repo:

```bash
mkdir -p ~/bqa-dogfood && cd ~/bqa-dogfood
```

Initialize the `.bqa/` workspace:

```bash
bqa init
```

Expected output:

```text
BQA workspace initialized in .bqa/
```

This creates the `.bqa/` directory tree (`input/sessions/`, `output/`,
`registry/`, `agents/`, `skills/`, `workflows/`, `prompts/`, …).

---

## 4. Prepare 5 sanitized ETL sessions

Create a folder of small, **synthetic** ETL notes / log snippets. BQA-OS
imports `*.md`, `*.log`, and `*.txt` files from a directory.

```bash
mkdir -p etl-notes
```

Add five files. Below are synthetic starters — adapt them to mirror the *shape*
of your checks, but keep them fake:

`etl-notes/etl_1_netsuite.md`

```markdown
# ETL 1 — NetSuite transactions (synthetic)
- Pipeline: netsuite_extract -> transform -> warehouse.fct_transactions
- Common bugs: duplicate (transaction_id, line_no) on retried batches; null account_id rows.
- Data quality: target row_count = source row_count minus rows dropped for null account_id.
- Useful prompt: "verify target row_count = source row_count minus null-account drops".
> Synthetic example only. Sanitize real client data before committing.
```

`etl-notes/etl_2_orders.md`

```markdown
# ETL 2 — Orders incremental load (synthetic)
- Common bugs: late-arriving updates overwritten by a stale partition.
- Data quality: order_total must equal sum(line_total) per order.
- Useful prompt: "check no order has order_total != sum(line_total)".
```

`etl-notes/etl_3_users.md`

```markdown
# ETL 3 — Users SCD2 dimension (synthetic)
- Common bugs: overlapping valid_from / valid_to windows; more than one current row per user.
- Data quality: exactly one row per user with is_current = true.
```

`etl-notes/etl_4_events.md`

```markdown
# ETL 4 — Events partitioned load (synthetic)
- Common bugs: duplicates on retry because the load is not idempotent.
- Data quality: no duplicate event_id within a single load partition.
```

`etl-notes/etl_5_pipeline.log`

```text
WARN dedup applied on batch 7 (transaction_id, line_no)
INFO row_count source=10012 target=10000 diff explained by null account_id
ERROR schema drift: new column "currency" appeared in source, not mapped to target
```

> Keep every file synthetic. The importer redacts obvious secrets, but it is not
> a substitute for not putting real data here in the first place.

---

## 5. Discover local sessions (optional)

`bqa discover` scans for AI-coding session artifacts. For a clean MacBook
dogfood with hand-written notes you can skip this, but it is harmless to run:

```bash
bqa discover --global=false --local=true
```

Expected output (no prior AI sessions in a fresh scratch dir):

```text
Discovering sessions: sources=claude,codex,opencode,droid global=false local=true
Discovered session-like files: 0
claude   0
codex    0
opencode 0
droid    0
Manifest: .bqa/input/sessions/manifest.json
```

The ETL notes you wrote by hand are imported in the next step via `--from`, not
by `discover`.

---

## 6. Ingest the ETL notes

Import your synthetic notes directory into normalized sessions plus a valid
`index.json`:

```bash
bqa ingest --from etl-notes
```

Expected output (counts match your file count):

```text
Discovered: 5
Imported: 5
Redactions: 0
Index: .bqa/input/sessions/index.json
```

Expected files:

```text
.bqa/input/sessions/index.json                  # session index
.bqa/input/sessions/raw/local-etl/*.md|*.log     # verbatim copies
.bqa/input/sessions/normalized/local-etl/*.md    # normalized sessions
```

> `Redactions: N` (N > 0) means the importer masked something that looked like a
> secret. Investigate why a secret was in a "synthetic" note before continuing.

---

## 7. Build QA knowledge artifacts

Turn normalized sessions into reusable knowledge plus starter runtime artifacts:

```bash
bqa build
```

Expected output:

```text
Sessions processed: 5
Knowledge artifacts created: 9
BQA artifacts created: 10
Knowledge dir: .bqa/knowledge
Generated dirs: .bqa/skills .bqa/agents .bqa/workflows .bqa/registry
```

Expected files in `.bqa/knowledge/`:

```text
etl_patterns.yaml
graphql_patterns.yaml
api_patterns.yaml
data_quality_patterns.yaml
common_bugs.yaml
successful_prompts.yaml
droid_patterns.yaml
runtime_patterns.yaml
project_profile.yaml
```

`project_profile.yaml` summarizes what was found, for example:

```yaml
project_profile:
  sessions_analyzed: 5
  signals:
    etl: 5
    data_quality: 4
    ...
  maturity: initial
```

Generated knowledge is **heuristic and keyword-driven**. Review it against the
[Knowledge Review Checklist](../knowledge-review-checklist.md) before trusting it
— domains with no signal (e.g. `graphql_patterns.yaml`) will be near-empty, and
that is expected for an ETL-only corpus.

---

## 8. Generate the Codex master context

Embed the project knowledge into a Codex-ready master context:

```bash
bqa codex
```

Expected output (the "Detected" line appears only if the Codex CLI is on your
`PATH`):

```text
BQA context generated: .bqa/prompts/bqa-master-context.md
Detected codex CLI: /Users/you/.local/bin/codex
Next: start codex and paste or reference .bqa/prompts/bqa-master-context.md as the initial project instruction.
```

Expected file: `.bqa/prompts/bqa-master-context.md`, which now contains a
`## Project QA Knowledge (generated by bqa build)` section condensed from your
ETL knowledge, e.g.:

```text
### ETL patterns
- 5 finding(s) recorded.
- Patterns: etl_validation
...
### Data quality patterns
- 4 finding(s) recorded.
```

The same workspace also works for other runtimes: run `bqa claude` or
`bqa opencode` to write the master context for Claude Code / OpenCode.

---

## 9. Use the master context in Codex

1. Start Codex in this project directory.
2. Reference the generated context as the initial project instruction:

   ```text
   Read .bqa/prompts/bqa-master-context.md and act as the BQA Master Agent.
   ```

3. Ask a QA task grounded in your corpus, e.g.:

   > Plan QA for the orders incremental load. Use the ETL and data-quality
   > patterns in the project knowledge.

Codex now has your project's ETL patterns, common bugs, and data-quality checks
as grounding instead of generic advice.

---

## 10. Validate the workspace

Two complementary checks:

```bash
bqa build --check
```

Expected:

```text
Validating knowledge artifacts in: .bqa/knowledge
Artifacts valid: 9 of 9 expected
All knowledge artifacts are present and valid.
```

```bash
bqa doctor
```

Expected (abridged): every workspace directory reports `[ok]`, the session index
is found with the right entry count, and the run ends with:

```text
All checks passed.
```

`bqa build --check` exits non-zero if any knowledge artifact is missing, empty,
or malformed; `bqa doctor` exits non-zero if the workspace layout is broken.
Both are safe local checks that touch no external services.

---

## 11. (Optional) Install the brain into another project

Once `.bqa/registry/` exists (created by `bqa build`), you can install the safe
artifacts into a separate target project:

```bash
mkdir -p ~/another-project
bqa brain install --from .bqa --target ~/another-project
```

This copies registry, agents, skills, workflows, prompts, and knowledge into
`~/another-project/.bqa/`. Raw sessions and secrets are **never** copied, and
unrelated files in the target are left untouched.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| `bqa: command not found` | CLI not built / not on `PATH` | Run `bash install.sh`, then add `export PATH="$HOME/.local/bin:$PATH"` to the shell. |
| `Imported: 0` from `bqa ingest --from` | Wrong directory, or no `*.md` / `*.log` / `*.txt` files in it | Check the path passed to `--from` and the file extensions. |
| Empty / near-empty `*.yaml` after `bqa build` | No signal for that domain in your notes (e.g. GraphQL files for an ETL corpus) | Expected. Add notes that mention the missing domain, or ignore that artifact. |
| `bqa build --check` fails | Knowledge dir missing or artifacts malformed | Re-run `bqa build` to regenerate `.bqa/knowledge/`. |
| Missing `index.json` / "0 entries" in `bqa doctor` | `bqa ingest` never ran, or ran against an empty dir | Re-run `bqa ingest --from <dir>` and confirm the discovered/imported counts. |
| Missing knowledge dir referenced by `bqa codex` | `bqa build` not run yet | The codex context still generates with a "run `bqa build`" hint; run `bqa build`, then `bqa codex` again. |
| `bqa doctor` reports a `FAIL` directory | Workspace partially created | Run `bqa init` to (re)create missing directories. |

---

## Privacy: never commit this

- **Never commit real logs, secrets, client names, tokens, or internal URLs.**
- The `.bqa/` directory can contain raw and normalized copies of whatever you
  import — keep it out of any shared/public repo unless you have reviewed it.
- Redaction during ingest is a safety net, **not** a guarantee. Sanitize before
  you import, not after.
- Before sharing anything generated here, run a scan: `bqa sanitize .bqa` (add
  `--write` only after reviewing the dry-run output).

When in doubt, keep it local. Local-first is the point of dogfooding.
