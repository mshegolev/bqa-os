# ADR-0002: Hexagonal Architecture

## Status

Accepted.

## Context

BQA-OS is evolving from a CLI prototype into a reusable QA operating system. It must support multiple domains and multiple external runtimes:

- Big Data and ETL testing
- GraphQL functional testing
- API and contract testing
- Codex
- Claude Code
- OpenCode
- local file-system storage
- private BQA Brain git repositories

If domain logic depends directly on CLI commands, file paths, git commands, or runtime-specific formats, the system will become hard to test, hard to extend, and hard to keep stable.

## Decision

BQA-OS will use hexagonal architecture, also known as ports and adapters.

The core business logic must not depend directly on Cobra, git, the local file system, Codex, Claude Code, OpenCode, GitHub, or any specific storage implementation.

## Target package layout

```text
internal/
  core/
    discovery/
    ingest/
    brain/
    runtime/
    sanitize/
    knowledge/
    agents/
    workflows/

  ports/
    session_source.go
    session_store.go
    brain_store.go
    sanitizer.go
    runtime_adapter.go
    clock.go
    logger.go

  adapters/
    fs/
    git/
    codex/
    claude/
    opencode/
    github/

  app/
    cli/
```

## Layer responsibilities

### Core

Core packages contain BQA business rules:

- discover sessions through ports
- ingest sessions through ports
- normalize sessions
- extract reusable knowledge
- generate agents, skills, workflows, and guardrails
- build project profiles
- prepare BQA Master Agent context

Core code must be deterministic, testable, and independent of external systems.

### Ports

Ports define interfaces owned by the core.

Examples:

```go
type SessionSource interface {
    Discover(ctx context.Context) ([]SessionRef, error)
    Read(ctx context.Context, ref SessionRef) ([]byte, error)
}

type SessionStore interface {
    SaveRaw(ctx context.Context, session RawSession) error
    SaveNormalized(ctx context.Context, session NormalizedSession) error
    SaveIndex(ctx context.Context, index SessionIndex) error
}
```

### Adapters

Adapters implement ports for specific systems:

- file-system session store
- Codex session source
- Claude Code session source
- OpenCode session source
- git-backed BQA Brain store
- local BQA Brain cache

Adapters may depend on external systems. Core must not.

### App / CLI

The CLI wires dependencies together.

Cobra commands should be thin orchestration wrappers:

```text
parse flags
construct adapters
call core use case
print result
```

## Migration strategy

The current prototype may keep existing packages temporarily. Refactoring should be incremental:

1. introduce ports;
2. move pure logic into `internal/core`;
3. move file-system and git logic into `internal/adapters`;
4. keep CLI stable;
5. add tests around core use cases;
6. gradually retire direct infrastructure calls from existing packages.

## Consequences

Benefits:

- easier testing
- cleaner domain model
- easier runtime support
- easier future storage backends
- safer refactoring
- clearer public engine versus private brain separation

Tradeoffs:

- more files and interfaces
- slightly slower initial development
- requires discipline to avoid putting business logic into adapters or CLI commands

## Rule

When adding new functionality, prefer this flow:

```text
core use case
↓
port interface
↓
adapter implementation
↓
CLI wiring
```

Do not add new business logic directly into Cobra command handlers.
