# ADR-0001: Language and Distribution Strategy

## Status

Accepted.

## Decision

BQA-OS core CLI will be implemented in **Go** and distributed as a single native binary named `bqa`.

## Why Go

- one static-ish binary for macOS, Linux, and Windows;
- no Python/Node runtime required on user machines;
- simple cross-compilation;
- strong standard library for file scanning, JSON/YAML processing, subprocess execution, and CLI tooling;
- easier private/internal distribution;
- good fit for local-first developer tooling.

## What stays outside the binary

Business/domain value should live in user-local data, not hardcoded source code:

- `.bqa/memory/`
- `.bqa/agents/`
- `.bqa/skills/`
- `.bqa/workflows/`
- `.bqa/rules/`
- `.bqa/guardrails/`
- `.bqa/registry/`

The public binary provides the engine. The private value is generated locally from user sessions and repositories.

## Security and IP posture

BQA-OS should be **local-first**:

- no automatic upload of project code;
- no automatic upload of sessions;
- explicit user-controlled export/import;
- generated memory can be gitignored;
- sensitive project knowledge stays in `.bqa/` inside the user's workspace or private storage.

## Target UX

```bash
cd project
bqa init
bqa ingest --sources claude,codex --global --local
bqa build
bqa codex
```

Then inside Codex/Claude Code:

```text
Протестируй DATA-12345
```

## Future packaging

- GitHub Releases with binaries:
  - `bqa-darwin-arm64`
  - `bqa-darwin-amd64`
  - `bqa-linux-amd64`
  - `bqa-linux-arm64`
  - `bqa-windows-amd64.exe`
- checksums file;
- optional Homebrew tap later.
