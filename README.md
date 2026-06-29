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
bqa codex
```

This creates:

```text
.bqa/prompts/bqa-master-context.md
```

Then start your AI coding runtime and use:

```text
Read .bqa/prompts/bqa-master-context.md and act as BQA Master Agent.

Task:
Test DATA-12345.
```

## First local knowledge workflow

BQA-OS is local-first: it helps a human QA owner and an AI coding runtime reuse project context from local artifacts. It is not a fully autonomous QA agent, and generated knowledge should be reviewed before use.

The first end-to-end workflow is:

```bash
bqa discover
bqa ingest2
bqa build
```

`bqa discover` finds supported local session sources. `bqa ingest2` normalizes discovered sessions into `.bqa/input/sessions/`. `bqa build` reads the normalized session index and writes reusable QA knowledge into `.bqa/knowledge/`.

Generated knowledge files:

```text
.bqa/knowledge/etl_patterns.yaml
.bqa/knowledge/graphql_patterns.yaml
.bqa/knowledge/api_patterns.yaml
.bqa/knowledge/data_quality_patterns.yaml
.bqa/knowledge/common_bugs.yaml
.bqa/knowledge/successful_prompts.yaml
.bqa/knowledge/project_profile.yaml
```

These files give Codex a compact QA context pack: domain patterns, recurring failures, useful prompt candidates, and a project profile. Use synthetic or sanitized sessions for public demos, and do not commit real logs, secrets, customer data, or private `.bqa/` outputs.

See [docs/knowledge-extractor.md](docs/knowledge-extractor.md) for a synthetic ETL walkthrough and manual verification commands.

## Commands available now

```bash
bqa init
bqa discover
bqa ingest
bqa ingest2
bqa build
bqa run "Test DATA-12345"
bqa runtime detect
bqa codex
bqa claude
bqa opencode
bqa doctor
```

## Current implementation status

Implemented:

- Go single-binary CLI foundation
- project-local `.bqa` workspace initialization
- runtime detection for Codex, Claude Code, and OpenCode
- BQA Master Agent context generation for runtime adapters
- local knowledge extraction into seven `.bqa/knowledge/*.yaml` artifacts
- early one-line installer through `install.sh`
- GitHub Pages placeholder for BQA-OS Agent Citadel

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
