# GraphQL session 3 — Pagination/filtering, resolver failure, schema-change risk (synthetic)

> Synthetic example only. No real schema, tokens, customers, or PII.

- **Service / schema:** `commerce-gateway` GraphQL schema, types `Order`, `Product`.
- **Domain:** GraphQL functional testing — pagination/filtering, resolver failure shape,
  and schema-change risk.

## Case E — Pagination / filtering case

We tested cursor pagination and filtering on the `products` GraphQL query.

```graphql
query Products($first: Int!, $after: String, $filter: ProductFilter) {
  products(first: $first, after: $after, filter: $filter) {
    edges { cursor node { id name priceCents inStock } }
    pageInfo { hasNextPage endCursor }
    totalCount
  }
}
```

Variables page 1: `{ "first": 2, "filter": { "inStock": true, "minPriceCents": 1000 } }`.
Page 2 reuses `after = pageInfo.endCursor` from page 1.

- **Expected behavior:** exactly `first` nodes per page; every node satisfies the filter
  (`inStock = true`, `priceCents >= 1000`); cursors are opaque and stable; walking pages with
  `after` visits each node once with no gaps or duplicates; `totalCount` is consistent across pages.
- **Observed:** page 1 + page 2 = 4 distinct ids, no overlap, `totalCount = 4`, last page
  `hasNextPage = false`.

## Common pagination failure modes

- **Offset/cursor drift:** an insert between page fetches shifts an offset-based cursor and a
  row is skipped or seen twice.
- **Filter not pushed down:** `totalCount` counts all products while `edges` are filtered, so
  the count and the page disagree.
- **Cursor leaks state:** the cursor encodes a raw DB primary key, exposing row ordering/volume.

## Case F — Resolver failure / error shape

We forced a downstream dependency (the inventory service) to time out and inspected the
GraphQL error shape for partial failure.

```graphql
query ProductWithStock($id: ID!) {
  product(id: $id) { id name inStock }   # inStock resolver calls inventory svc
}
```

- **Expected behavior:** GraphQL returns `data.product` with `id`/`name` populated and
  `inStock = null`, plus a top-level `errors` entry whose `path = ["product","inStock"]` and
  `extensions.code = "DOWNSTREAM_UNAVAILABLE"`. Partial data + scoped error, not a blanket 500.
- **Observed error shape:**

```json
{
  "data": { "product": { "id": "prod_1", "name": "Widget", "inStock": null } },
  "errors": [{
    "message": "inventory service timeout",
    "path": ["product", "inStock"],
    "extensions": { "code": "DOWNSTREAM_UNAVAILABLE" }
  }]
}
```

- **Common failure mode:** an unhandled resolver exception bubbles up as
  `INTERNAL_SERVER_ERROR` with a stack trace in `message`, nulling the whole `data` and leaking
  internals. The error contract (code + path) is the thing under test.

## Case G — Schema-change risk

PR under review renames `Order.totalCents` to `Order.totalAmountCents` and changes
`Order.status` from a free `String` to an `OrderStatus` enum.

- **Schema-change risk:** this is a **breaking GraphQL schema change**. Existing clients querying
  `totalCents` get a validation error (`Cannot query field "totalCents"`), and clients sending
  `status: "paid"` (lowercase string) break against the new enum.
- **Expected QA gate:** run an introspection diff (old vs new GraphQL schema), flag removed/renamed
  fields and type narrowings as breaking, and require a deprecation (`@deprecated`) + alias window
  before removal.
- **Common failure mode:** the field is removed in the same release it is renamed, with no
  `@deprecated` period — production clients 400 immediately.

## Useful QA prompt

> "Diff the GraphQL schema introspection before and after this PR; list every removed or renamed
> field and every enum/type narrowing as a breaking change, and confirm each has an @deprecated
> alias window before it can ship."
