# Reusable GraphQL regression workflow

A repeatable loop a QA lead can run on every PR that touches a GraphQL schema or
resolvers. It is framed in BQA-OS terms: feed sessions in, let `bqa build` surface
GraphQL signals into `graphql_patterns.yaml`, then drive regression checks from those
patterns.

All inputs here are 100% synthetic.

## 0. Preconditions

- A GraphQL endpoint (or mock) reachable from the test runner.
- The eight canonical cases captured as session notes (see `../notes/`): query with
  variables, mutation happy path, validation error, auth/role-based access,
  pagination/filtering, resolver failure shape, schema-change risk, and an
  investigation prompt.

## 1. Capture → Build (turn QA sessions into patterns)

```bash
bqa discover                 # find local AI coding session artifacts
bqa ingest                   # normalize them into .bqa/input/sessions/normalized/
bqa build                    # emit .bqa/knowledge/*_patterns.yaml
```

`bqa build` scans each normalized session; any that mentions `graphql` together with a
trigger phrase (`graphql query` / `graphql mutation` / `graphql schema` /
`graphql resolver`) becomes a `graphql_functional_testing` finding in
`.bqa/knowledge/graphql_patterns.yaml`, and bumps `signals.graphql` in
`project_profile.yaml`. See `../expected/graphql_patterns.yaml` for the shape.

## 2. Regression matrix (run on every schema/resolver PR)

| # | Case | What to assert | Breaking-change signal |
|---|------|----------------|------------------------|
| 1 | Query with variables | typed, non-null required fields; stable `pageInfo` | required field removed / type narrowed |
| 2 | Mutation happy path | `userErrors` empty, no top-level `errors`, idempotent retry | mutation payload field renamed |
| 3 | Validation error | `extensions.code = BAD_USER_INPUT`, `data = null`, resolver not run | validation moved into resolver → 500 |
| 4 | Auth / role access | non-admin → `FORBIDDEN`, field nulled, no leaked values | field-level auth check dropped |
| 5 | Pagination / filtering | exactly `first` nodes, filter holds, no gaps/dupes, stable cursors | cursor encoding changed / filter not pushed down |
| 6 | Resolver failure shape | partial `data`, scoped `errors[].path`, stable `extensions.code` | exception leaks as `INTERNAL_SERVER_ERROR` + stack trace |
| 7 | Schema-change risk | introspection diff has no breaking removal without `@deprecated` | renamed/removed field, enum narrowing, non-null tightening |

## 3. Schema diff gate

```bash
# pseudo-steps — wire to your introspection tooling
graphql-introspect $BASELINE_URL > schema.base.json
graphql-introspect $PR_URL       > schema.pr.json
graphql-schema-diff schema.base.json schema.pr.json   # fail CI on BREAKING
```

Breaking = removed/renamed field or type, enum value removed, nullable → non-null on
input, or output non-null → nullable. Each breaking change must ship behind an
`@deprecated` alias window, not in the removal release.

## 4. Investigation loop (when a check fails)

1. Pull the matching `graphql_functional_testing` evidence from `graphql_patterns.yaml`.
2. Reproduce with the exact query/mutation + variables from the relevant note.
3. Inspect the **error contract**, not just HTTP status: top-level `errors`,
   `extensions.code`, and `path`.
4. Classify: input bug (`BAD_USER_INPUT`), authz bug (`FORBIDDEN`/bypass), downstream
   (`DOWNSTREAM_UNAVAILABLE`), or schema break.
5. Add the reproduction back as a session note so the next `bqa build` strengthens the
   pattern (the regression becomes part of the knowledge base).

## 5. Definition of done

- All seven matrix rows green against the PR endpoint.
- Schema diff gate reports no un-deprecated breaking change.
- `bqa build` re-run; `signals.graphql` reflects the new/updated notes.
