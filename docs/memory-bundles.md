# Memory Bundles: `bqa brain export` / `import`

A **memory bundle** is a portable, sanitized snapshot of a project's BQA-OS
memory — the knowledge, agents, skills, workflows, prompts, and registry that
`bqa build` produces under `.bqa/`. Bundles let you move project memory between
machines or teammates as a single `zip`/`tar` file, or publish it to a shared
GitHub brain, and restore it safely somewhere else.

## What is (and isn't) exported

Only an **allow-list** of `.bqa` subdirectories is ever packaged:

```
knowledge  agents  skills  workflows  prompts  registry
```

Everything else is excluded by construction. In particular, `input/sessions/`
(raw session transcripts) and any local caches are **never exported** — the
allow-list is the primary safety guard, so secrets or PII that live only in raw
sessions cannot leak into a bundle.

A second guard runs on top of the allow-list: before writing an archive, export
stages the payload and runs a non-destructive **sanitize audit** that flags
files containing likely secrets/PII (redaction candidates). The audit summary is
recorded in the bundle and, with `--strict`, aborts the export.

## Bundle layout

An exported `zip`/`tar` bundle is deterministic (fixed archive mod-times, no
wall-clock fields) and contains:

```
knowledge/...                 # allow-listed payload (only dirs that exist)
agents/...
registry/...
manifest.json                 # bundle_version, tool, generated_by, included dirs, file_count
metadata/checksums.json       # SHA-256 of every payload file
metadata/memory_audit.yaml    # machine-readable audit summary
README.md                     # what the bundle is
install.md                    # how to restore it
audit.md                      # human-readable audit summary
```

## Exporting

```bash
# Zip bundle (default target)
bqa brain export --source .bqa --target zip --out ./bqa-memory.zip

# Tar bundle
bqa brain export --source .bqa --target tar --out ./bqa-memory.tar
```

Flags:

- `--source` — source memory directory (default `.bqa`).
- `--target` — `zip` | `tar` | `github`.
- `--out` — output archive path (required for `zip`/`tar`).
- `--dry-run` — print the planned files and audit summary; write nothing.
- `--strict` — abort if the audit flags any redaction candidates.
- `--exclude` — glob pattern matched against bundle-relative paths; repeatable.

Examples:

```bash
# Preview what would be exported, without writing an archive
bqa brain export --source .bqa --target zip --out /tmp/bundle.zip --dry-run

# Fail loudly if anything looks like a secret
bqa brain export --source .bqa --target zip --out ./bundle.zip --strict

# Drop draft prompts from the bundle
bqa brain export --source .bqa --target zip --out ./bundle.zip --exclude 'prompts/drafts/*'
```

## Importing

```bash
bqa brain import --from ./bqa-memory.zip --target /path/to/project
```

Import first reads the archive and **verifies the manifest and every checksum
before writing anything** to the target. If `manifest.json` or
`metadata/checksums.json` is missing, the `bundle_version` is unsupported, or any
payload file fails its SHA-256 check, import aborts and the target is left
untouched.

Once verification passes, import stages the allow-listed payload and reuses the
existing non-destructive brain install to copy it into `<target>/.bqa/`.
Unrelated files in the target are never modified.

Flags:

- `--from` — bundle path, `.zip` or `.tar` (required).
- `--target` — target project directory (required).
- `--dry-run` — verify manifest + checksums only; install nothing.

```bash
# Verify a bundle's integrity without installing it
bqa brain import --from ./bqa-memory.zip --target /path/to/project --dry-run
```

## GitHub brain flow

Instead of a file, memory can be published to a connected GitHub brain:

```bash
# One-time: connect a brain repository
bqa brain connect https://example.invalid/org/bqa-brain.git

# Publish the allow-listed memory to the brain (sanitized on the way out)
bqa brain export --target github
```

The `github` target materializes the allow-listed payload into the connected
brain cache and commits/pushes it via the existing brain sync (which sanitizes
before committing). It requires a connected brain; otherwise the export reports
that the brain is not connected.

_Versioning: the patch version auto-bumps in `VERSION` on every merge to `main`._
