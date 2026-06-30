# Runtime Adapters

BQA-OS is designed to work with multiple AI coding runtimes.

Supported targets:

- Codex
- Claude Code
- OpenCode

## Design principle

BQA-OS should not depend on private or unstable internal session formats unless a runtime provides a stable contract.

The safe integration layer is:

1. detect runtime CLI in PATH;
2. generate BQA Master Agent context;
3. prepare project-local `.bqa` artifacts;
4. install project-local command helpers where the runtime has a stable file-based convention;
5. emit runtime-native agents, skills, workflows, guardrails, and memory indexes from a BQA registry;
6. instruct the user how to start or feed the context into the selected runtime;
7. ingest exported sessions only from known local files or explicit user-provided paths.

## Commands

```bash
bqa runtime detect
bqa runtime install-commands
bqa codex
bqa claude
bqa opencode
bqa emit --registry /path/to/bqa-team/team/brain/registry.json --target /path/to/project
```

`bqa runtime install-commands` writes:

```text
.bqa/prompts/bqa-master-context.md
.bqa/runtime-commands/codex/bqa-master.md
.bqa/runtime-commands/claude/bqa-master.md
.bqa/runtime-commands/opencode/bqa-master.md
.claude/commands/bqa-master.md
```

Claude Code can use the project command as `/bqa-master`. Codex and OpenCode
use the `.bqa/runtime-commands/` helper files until a stable native project
slash-command contract is confirmed for those runtimes.

`bqa emit` turns the unified BQA registry into runtime-native files:

- Claude Code: `.claude/agents/`, `.claude/skills/`, `.claude/commands/`
- Codex: `.codex/AGENTS.md`, `.codex/prompts/`, `.codex/bqa/`
- OpenCode: `.opencode/agent/`, `.opencode/command/`, `.opencode/bqa/`

When the command is run from outside the `bqa-team` repository, pass
`--registry /path/to/bqa-team/team/brain/registry.json` explicitly.

## OpenCode support

OpenCode is treated as a first-class runtime adapter.

Initial behavior:

```bash
bqa opencode
```

This generates:

```text
.bqa/prompts/bqa-master-context.md
```

The user can then launch OpenCode and provide this context as the initial BQA Master Agent instruction.

Future behavior may include direct launch or context injection if OpenCode exposes a stable CLI/API contract for doing so.
