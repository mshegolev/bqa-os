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
