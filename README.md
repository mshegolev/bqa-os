# BQA-OS

**BQA-OS (Better QA Operating System)** is an AI-native operating system for quality engineering.

BQA-OS is designed to connect QA knowledge, agents, skills, workflows, guardrails, memory, and AI coding runtimes into one reusable system.

## Live demo

**BQA-OS Agent Citadel** is a GitHub Pages demo placeholder for the upcoming visual flow:

```text
raw QA sessions → sanitized knowledge → skills → agents → workflows
```

Live URL, after GitHub Pages is enabled for `/docs` on the `main` branch:

```text
https://mshegolev.github.io/bqa-os/
```

Local preview:

```bash
python3 -m http.server 8080 -d docs
```

Then open:

```text
http://localhost:8080
```

## Vision

A user should be able to open a repository and say:

```text
Test DATA-12345
```

or:

```text
Create GraphQL functional tests
```

or:

```text
Validate this ETL pipeline
```

BQA-OS should then help the selected AI coding runtime act as a BQA Master Agent.

## Supported domains

- Big Data & ETL Testing
- GraphQL Functional Testing
- API Testing
- Contract Testing
- Data Quality Validation
- Test Automation Engineering

## Supported AI coding runtimes

- Codex
- Claude Code
- OpenCode

## Repository split

```text
mshegolev/bqa-os      public engine / binary
mshegolev/bqa-brain   private knowledge / agents / memory / workflows
```

The public repository contains the runtime engine. Private project value should live in BQA Brain or local `.bqa` workspaces.

## Install

Early installer requires Go:

```bash
brew install go
curl -fsSL https://raw.githubusercontent.com/mshegolev/bqa-os/main/install.sh | bash
export PATH="$HOME/.local/bin:$PATH"
```

Check installation:

```bash
bqa --help
bqa runtime detect
```

## Project usage

```bash
cd /path/to/project
bqa init
bqa runtime detect
bqa runtime install-commands
bqa codex
```

This creates the BQA master context and project-local command helpers:

```text
.bqa/prompts/bqa-master-context.md
.bqa/runtime-commands/codex/bqa-master.md
.bqa/runtime-commands/claude/bqa-master.md
.bqa/runtime-commands/opencode/bqa-master.md
.claude/commands/bqa-master.md
```

Claude Code can then use:

```text
/bqa-master
```

For Codex and OpenCode, reference the helper file from `.bqa/runtime-commands/`
or start your AI coding runtime and use:

```text
Read .bqa/prompts/bqa-master-context.md and act as BQA Master Agent.

Task:
Test DATA-12345.
```

To install the unified BQA Team agents, skills, workflows, guardrails, and
memory indexes into a project, emit runtime-native files from the BQA registry:

```bash
bqa emit --registry /path/to/bqa-team/team/brain/registry.json --target .
```

This writes runtime-native assets under `.claude/`, `.codex/`, and `.opencode/`
without overwriting user-owned root instruction files such as `CLAUDE.md` or
`AGENTS.md`.

## Commands available now

```bash
bqa init
bqa discover
bqa ingest
bqa build
bqa build --sales-package
bqa emit --registry /path/to/bqa-team/team/brain/registry.json --target .
bqa etl-agent-pack
bqa run "Test DATA-12345"
bqa team pipeline --issue-json issue.json --issue-number 123
bqa runtime detect
bqa runtime install-commands
bqa codex
bqa claude
bqa opencode
bqa doctor
```

## 2-week QA Memory Pilot package

For internal pilot validation, generate the Monday sales package alongside the
starter QA artifacts:

```bash
bqa build --sales-package
```

The command writes the normal `.bqa/knowledge`, `.bqa/skills`, `.bqa/agents`,
`.bqa/workflows`, and `.bqa/registry` outputs, plus `.bqa/sales/` materials:

- pilot offer one-pager
- internal demo script
- discovery call script
- onboarding checklist
- sample Slack, LinkedIn, and email outreach
- pricing hypothesis
- internal stakeholder FAQ

Use synthetic artifacts for public demos and sanitized artifacts only for pilot
customers. Do not place private repo data, real session logs, customer records,
or secrets in public artifacts.

## ETL QA Agent Pack

Generate a copy-paste-ready local ETL QA Agent Pack for Codex and Claude Code:

```bash
bqa etl-agent-pack
```

The command reads `.bqa/knowledge/` and `.bqa/input/sessions/` when available,
uses only aggregate statistics from local inputs, and writes synthetic-safe pack
files under:

```text
.bqa/output/etl-agent-pack/
```

Generated pack directories:

- `statistics/`
- `agents/`
- `workflows/`
- `specs/`
- `prompts/`
- `examples/`
- `README_NEXT_STEPS.md`

## Current implementation status

Implemented:

- Go single-binary CLI foundation
- project-local `.bqa` workspace initialization
- runtime detection for Codex, Claude Code, and OpenCode
- BQA Master Agent context generation for runtime adapters
- early one-line installer through `install.sh`
- GitHub Pages placeholder for BQA-OS Agent Citadel
- optional Monday sales package generation for internal pilot validation
- local ETL QA Agent Pack generation for Codex and Claude Code

Planned:

- `bqa brain connect`
- `bqa brain pull`
- `bqa brain sync`
- `bqa sanitize`
- real session analyzer
- agent generator
- skill generator
- workflow generator
- project profile builder
- GitHub Releases with prebuilt binaries
- `bqa self-update`

## Security posture

BQA-OS should not hardcode private business value. Generated knowledge, project profiles, prompts, agents, skills, workflows, and guardrails should be stored in a private BQA Brain repository or local encrypted cache after sanitization.
