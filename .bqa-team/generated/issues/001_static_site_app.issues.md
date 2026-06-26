---ISSUE---
TITLE: Static BQA Web App MVP
LABELS: bqa:arch-approved,bqa:ready-dev,bqa:static-site
BODY:
## Context

Business task from `001_static_site_app.md`. This issue was routed through the technical architect stage.

## Goal

Static BQA Web App MVP

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

# Business Task: Static BQA Web App MVP

Create a static HTML/JS application for BQA-OS where a user can upload a specially marked session archive and receive a generated output archive containing agents, workflows, specs, instructions, and recommendations.

Constraints:
- static site only: HTML/CSS/JavaScript;
- local-first: processing should happen in browser when feasible;
- no private data uploaded to external services by default;
- output should be downloadable as a zip archive;
- UX should explain the expected archive structure;
- should include sample synthetic input data only.

Expected outputs from the generated archive:
- agents/
- workflows/
- specs/
- knowledge/
- README_NEXT_STEPS.md
- recommendations.md


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
