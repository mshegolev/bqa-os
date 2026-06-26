# BQA-OS

**BQA-OS** — Big Data QA Operating System: агентная система для тестирования ETL, Data Engineering и Big Data проектов.

Идея: вы вызываете одного **BQA Master Agent**, а он читает registry, memory, agents, skills, workflows и guardrails, после чего сам выбирает нужных специализированных агентов.

## Установка

```bash
pip install -e .
```

## Быстрый старт

```bash
cd /path/to/autotests-repo
bqa init
bqa ingest --sources claude,codex --global --local
bqa build
bqa doctor
bqa run "Протестируй DATA-12345"
```

## Режим внутри Codex/Claude Code

Основной UX должен быть таким:

```bash
cd /path/to/project
bqa codex
```

Дальше внутри Codex CLI:

```text
Протестируй DATA-12345
```

BQA-OS подготавливает master prompt/context bundle, чтобы Codex работал в режиме BQA Master Agent.

## Команды

```bash
bqa init
bqa discover --sources claude,codex
bqa export-sessions --sources claude,codex
bqa ingest --sources claude,codex --global --local
bqa analyze .bqa/input/sessions
bqa build
bqa run "Протестируй ETL"
bqa codex
bqa doctor
```

## Структура

```text
.bqa/
  input/sessions/
  output/
  registry/
  memory/
  agents/
  skills/
  workflows/
  rules/
  guardrails/
  prompts/
```

## Ограничение

BQA-OS может импортировать только те Claude Code / Codex сессии, которые реально доступны локально в файлах. Если история хранится только в облаке и не экспортирована локально, её нужно экспортировать отдельно.
