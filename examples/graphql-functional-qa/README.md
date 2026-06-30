# GraphQL Functional QA — synthetic demo pack

A self-contained, **100% synthetic** demo of how BQA-OS turns GraphQL QA sessions into
reusable knowledge. Nothing here is real: no real schema, endpoints, tokens, customers,
or PII. It is written in the language of a QA lead working with GraphQL APIs and is safe
to use for docs, local checks, and sales demos.

> See also the smaller ETL example in [`../etl-weekend/`](../etl-weekend/).

## What's in the pack

```
graphql-functional-qa/
├── README.md                       # this file
├── workflow.md                     # reusable GraphQL regression workflow
├── notes/                          # synthetic QA session notes (the INPUTS)
│   ├── gql_1_query_mutation.md     # query w/ variables + mutation happy path
│   ├── gql_2_validation_auth.md    # validation error + role-based access
│   └── gql_3_pagination_resolver_schema.md  # pagination/filter + resolver failure + schema risk
├── raw/
│   └── gql_session.log             # synthetic functional-suite run log
└── expected/
    └── graphql_patterns.yaml       # illustrative `bqa build` OUTPUT
```

## The eight demo cases

| # | Case | Where | One-line expected behavior |
|---|------|-------|----------------------------|
| A | Query with variables | `notes/gql_1` | typed nodes, `pageInfo` correct, required fields non-null |
| B | Mutation happy path | `notes/gql_1` | `userErrors` empty, no top-level `errors`, idempotent retry |
| C | Validation error | `notes/gql_2` | `extensions.code = BAD_USER_INPUT`, `data = null`, resolver not run |
| D | Auth / role-based access | `notes/gql_2` | non-admin → `FORBIDDEN`, field nulled, no leaked values |
| E | Pagination / filtering | `notes/gql_3` | exactly `first` nodes, filter holds, no gaps/dupes, stable cursors |
| F | Resolver failure / error shape | `notes/gql_3` | partial `data`, scoped `errors[].path`, stable `extensions.code` |
| G | Schema-change risk | `notes/gql_3` | introspection diff flags breaking removal/rename without `@deprecated` |
| H | QA investigation prompt | below | reusable prompt to drive the investigation loop |

Each note states **expected behavior** and the **common failure modes** for its cases.

## H — QA investigation prompt

Drop this into your GraphQL endpoint runner (or an agent) when a check fails:

> **You are a GraphQL functional QA engineer.** Given the operation, variables, and
> observed response below, do not trust the HTTP status alone. Inspect the GraphQL error
> contract: top-level `errors`, each `errors[].extensions.code`, and `errors[].path`.
> Decide whether this is (a) input validation (`BAD_USER_INPUT`), (b) authz —
> `FORBIDDEN`/`UNAUTHENTICATED` or a field-level bypass, (c) a downstream/resolver failure
> (`DOWNSTREAM_UNAVAILABLE`, partial data expected), or (d) a breaking schema change. State
> the expected response, the diff vs observed, the most likely root cause, and the minimal
> reproduction (operation + variables). Flag any leaked internal value or stack trace.

## How `bqa build` surfaces these signals

BQA-OS reads QA sessions and emits domain-grouped pattern files under
`.bqa/knowledge/`. For this pack the relevant output is `graphql_patterns.yaml`.

Pipeline:

```bash
bqa discover   # locate local AI coding session artifacts
bqa ingest     # normalize them into .bqa/input/sessions/normalized/
bqa build      # emit .bqa/knowledge/*_patterns.yaml (incl. graphql_patterns.yaml)
```

**The mapping (inputs → `graphql_patterns.yaml`):**

- `bqa build` scans every normalized session. A session is tagged as a GraphQL signal when
  it contains the literal `graphql` **and** at least one trigger phrase —
  `graphql query`, `graphql mutation`, `graphql schema`, or `graphql resolver`. Each note
  in `notes/` is written to contain these phrases verbatim, so all three fire.
- Each fired signal becomes **one finding**:
  `name: graphql_functional_testing`, `domain: graphql`, an `evidence` snippet sliced
  around the first matched phrase, and the `source` path of the normalized session.
- Findings are de-duplicated by `(name, source)` and sorted by source path.
- `project_profile.yaml` increments `signals.graphql` by the number of GraphQL notes
  ingested (3 from this pack).

GitHub-API noise is intentionally excluded: sessions whose only GraphQL mention is the
GitHub GraphQL API (`api/graphql`, `github_graphql_url`) are **not** counted, so these
demo signals stay clean. See [`expected/graphql_patterns.yaml`](expected/graphql_patterns.yaml)
for the resulting shape.

## Try it locally

```bash
# from a repo where BQA-OS is set up, copy the synthetic notes into the session inputs
mkdir -p .bqa/input/sessions/normalized/codex
cp examples/graphql-functional-qa/notes/*.md .bqa/input/sessions/normalized/codex/

bqa build
cat .bqa/knowledge/graphql_patterns.yaml      # three graphql_functional_testing findings
grep -A2 'graphql:' .bqa/knowledge/project_profile.yaml   # signals.graphql incremented
```

## Reusable regression workflow

The per-PR loop (capture → build → 7-row regression matrix → schema diff gate →
investigation loop → done) lives in [`workflow.md`](workflow.md). Run it on every PR that
touches a GraphQL schema or resolvers; failed checks feed new notes back into `bqa build`,
so the GraphQL pattern set strengthens over time.
