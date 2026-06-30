# ETL / Data Quality QA Prompt

Data-quality QA prompt for ETL / Big Data pipelines: completeness, uniqueness,
reconciliation, schema drift, and freshness. Synthetic and safe.

## When to use

- Validating a pipeline run before sign-off.
- Generating data-quality rules from a table contract.

## Template

```text
You are doing data-quality QA on an ETL pipeline.

Inputs:
- Pipeline / job: [name]
- Source -> destination: [source table] -> [destination table]
- Grain & business key: [columns]
- Partitioning: [e.g. partition_date]
- Run / partition under test: [date or run id]

Produce a data-quality plan covering these dimensions. For each, give the check
as a SQL-shaped assertion with [placeholder] values and the expected result:

1. Completeness — destination row count reconciles to source extract count.
2. Uniqueness — no duplicate business keys within the partition.
3. Not-null — required columns have no nulls.
4. Referential — foreign keys resolve to a parent.
5. Range / domain — numeric and enum columns within allowed bounds.
6. Schema drift — column set and types match the contract; flag added/removed.
7. Freshness — max(load_ts) within the expected SLA window.

For each check output: name, dimension, query (with [placeholder] values),
expected result, and severity if it fails.

Rules:
- Never include real data values, secrets, connection strings, or hostnames.
- Each check must be copy-pasteable into a validation_rules-style config.
- Prefer set-based assertions that return zero rows on success.
```

## Example checks (synthetic)

```sql
-- completeness
SELECT (SELECT COUNT(*) FROM [src] WHERE partition_date = '[date]')
     - (SELECT COUNT(*) FROM [dst] WHERE partition_date = '[date]') AS diff;
-- expect: diff = 0

-- uniqueness
SELECT business_key, COUNT(*) c
FROM [dst] WHERE partition_date = '[date]'
GROUP BY business_key HAVING c > 1;  -- expect zero rows
```

## Notes

- Output is shaped to drop into a `validation_rules.yml`-style framework.
- Reusable rules (reconciliation, dup-key, freshness) are exactly the kind of
  reusable artifact the [knowledge extraction prompt](knowledge_extraction_prompt.md)
  captures into `.bqa/knowledge/` (see [`README.md`](README.md)).
