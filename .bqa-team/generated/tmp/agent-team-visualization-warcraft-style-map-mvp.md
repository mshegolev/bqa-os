## Context

Business task from `002_agent_game_visualization.md`. This issue was routed through the technical architect stage.

## Goal

Agent Team Visualization / Warcraft-style Map MVP

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

# Business Task: Agent Team Visualization / Warcraft-style Map MVP

Create a lightweight visual dashboard idea for BQA-OS that represents agents as units on a map/board.

Goal:
- make team workflow visible and motivating;
- show roles: Business Owner, Architect, Developer, QA, Reviewer;
- show task states: idea, architecture review, ready for dev, in dev, QA, bug found, business acceptance, done;
- implement as simple static HTML/JS visualization first, not a real game engine.

Constraints:
- no heavy framework for MVP;
- no backend required;
- use synthetic demo data;
- should not distract from core BQA workflow.


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