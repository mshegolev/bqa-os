# BQA-OS Issue Spec Template

Используй этот шаблон для задач разработчикам.

---

## Title

Короткое название задачи.

## Context

Почему задача нужна и частью какого workflow она является.

Пример:

This task is part of the Knowledge Extractor vertical slice. BQA-OS should read normalized sessions from `.bqa/input/sessions/` and write reusable QA knowledge artifacts to `.bqa/knowledge/`.

## Goal

Что должно заработать после задачи.

## Scope

### Create/change

- `internal/...`

### Do not touch

- `internal/...`
- private repo data
- real session logs
- secrets

## Architecture

Follow:

core use case
↓
port interface
↓
adapter implementation
↓
CLI wiring

Cobra must stay thin.

## Behavior

Опиши expected behavior.

## Acceptance criteria

- [ ] Works on synthetic test data.
- [ ] Does not require private data.
- [ ] Does not break existing commands.
- [ ] Has unit tests where reasonable.
- [ ] `go test ./...` passes.
- [ ] Errors are clear and actionable.

## Manual verification

```bash
go test ./...
go run ./cmd/bqa <command>
```

## Expected output

Опиши stdout и файлы, которые должны появиться.

## Notes for developer

Дополнительные ограничения или подсказки.
