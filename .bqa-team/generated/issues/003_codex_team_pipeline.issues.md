---ISSUE---
TITLE: Codex Team Pipeline MVP
LABELS: bqa:arch-approved,bqa:ready-dev,bqa:codex-team
BODY:
## Context

Business task from `003_codex_team_pipeline.md`. This issue was routed through the technical architect stage.

## Goal

Codex Team Pipeline MVP

## Scope

### Create/change

- To be refined by Architect/Codex output before execution.

### Do not touch

- Private repo data
- Real session logs
- Secrets

## Architecture

Follow:

core use case
↓
port interface
↓
adapter implementation
↓
CLI wiring

Business logic must not live directly in Cobra commands.

## Behavior

# Business Task: Codex Team Pipeline MVP

Create the first automated workflow where business tasks are transformed into GitHub issues, routed through architecture review, implemented by role-specific Codex agents, checked by QA, and finally sent to business acceptance.

Goal:
- use GitHub Issues as the source of truth;
- use labels to represent workflow state;
- use Codex CLI for role execution;
- keep all task specs aligned with BQA-OS architecture and issue template;
- create bugs when QA rejects implementation.

Constraints:
- dry-run by default;
- no infinite uncontrolled loop unless explicitly enabled;
- do not store secrets;
- do not commit private data.


## Acceptance criteria

- [ ] Architecture boundaries are respected.
- [ ] Synthetic data only.
- [ ] Manual verification steps pass.

## Manual verification

```bash
go test ./...
```

## Role routing

- Business owner: validates value and scope.
- Technical architect: validates architecture before development.
- Developer: implements only after architecture approval.
- QA: verifies and creates bug issues if acceptance criteria fail.
- Business owner: performs final acceptance.

---END---
