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
4. instruct the user how to start or feed the context into the selected runtime;
5. ingest exported sessions only from known local files or explicit user-provided paths.

## Commands

```bash
bqa runtime detect
bqa codex
bqa claude
bqa opencode
```

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
