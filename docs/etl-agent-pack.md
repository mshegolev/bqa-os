# ETL QA Agent Pack

The ETL QA Agent Pack creates a local, copy-pasteable pack for ETL testing in Codex and Claude Code.

## Command

```bash
bqa etl-agent-pack
```

Optional inputs:

```bash
bqa etl-agent-pack \
  --sessions .bqa/input/sessions \
  --knowledge-dir .bqa/knowledge
```

## Inputs

The command reads counts from local generated artifacts when they exist:

```text
.bqa/input/sessions/index.json
.bqa/knowledge/etl_patterns.yaml
.bqa/knowledge/data_quality_patterns.yaml
.bqa/knowledge/project_profile.yaml
```

If inputs are missing, the pack still renders with synthetic demo statistics.

## Outputs

Generated files are written under:

```text
.bqa/output/etl-agent-pack/
```

Structure:

```text
statistics/
agents/
workflows/
specs/
prompts/
examples/
README_NEXT_STEPS.md
```

## Safety

The pack uses synthetic examples and copies only counts from local indexes or knowledge files. It does not copy normalized session text, knowledge evidence, private logs, customer data, or secrets.
