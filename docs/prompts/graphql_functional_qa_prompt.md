# GraphQL Functional QA Prompt

Functional QA prompt for GraphQL APIs: schema coverage, query/mutation behavior,
error handling, and pagination. Synthetic and safe.

## When to use

- Testing a GraphQL endpoint's functional correctness.
- Generating a coverage-oriented test plan from a schema.

## Template

```text
You are doing functional QA on a GraphQL API.

Inputs:
- Schema (SDL or introspection): [paste or reference]
- Endpoint: [placeholder URL]
- Auth model: [none / token / [placeholder]]
- Focus area: [type / query / mutation under test]

Produce a functional test plan covering:
1. Happy path
   - For each target query/mutation: a valid request with the smallest selection
     set and a fully expanded selection set; state expected fields and types.
2. Input validation
   - Missing required args, wrong types, out-of-range values, null vs absent.
   - Expected: a typed GraphQL error, not a 500.
3. Nullability & partial data
   - Force a resolver error on one field; assert siblings still resolve and the
     `errors` array references the right path.
4. Pagination & ordering
   - First/after, last/before, empty page, beyond-last-page; stable ordering.
5. Authorization (if applicable)
   - Authenticated vs anonymous; field-level access where relevant.

For each case output: name, request (query + variables, with [placeholder]
values), expected response shape, and the assertion.

Rules:
- Never use real tokens, customer IDs, or PII. Use [placeholder].
- Assert on shape and error paths, not on volatile values.
- Call out any schema field with no test as a coverage gap.
```

## Example case (synthetic)

```graphql
query GetOrder($id: ID!) {
  order(id: $id) { id status total { amount currency } }
}
# variables: { "id": "[placeholder-order-id]" }
# expect: order.id == input id; status in enum; total.currency is 3-letter code
# negative: id = "" -> typed error, path ["order"], no 5xx
```

## Notes

- Reusable assertions (typed-error-not-500, stable pagination) belong in the
  knowledge base — capture via the
  [knowledge extraction prompt](knowledge_extraction_prompt.md).
