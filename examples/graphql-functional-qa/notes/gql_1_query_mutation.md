# GraphQL session 1 — Orders query + checkout mutation (synthetic)

> Synthetic example only. No real schema, tokens, customers, or PII. Sanitize real
> client data before committing.

- **Service / schema:** `commerce-gateway` GraphQL schema, types `Order`, `Cart`, `Customer`.
- **Domain:** GraphQL functional testing — query with variables and a happy-path mutation.

## Case A — GraphQL query with variables

We tested the `orders` GraphQL query with variables to make sure paginated reads
return a stable, typed shape.

```graphql
query OrdersByCustomer($customerId: ID!, $status: OrderStatus, $first: Int!) {
  orders(customerId: $customerId, status: $status, first: $first) {
    edges {
      node { id status totalCents currency placedAt }
    }
    pageInfo { hasNextPage endCursor }
  }
}
```

Variables: `{ "customerId": "cust_synthetic_001", "status": "PAID", "first": 20 }`.

- **Expected behavior:** returns up to 20 `Order` nodes, each non-null `id`/`status`,
  `totalCents` as an `Int`, `pageInfo.hasNextPage` reflecting whether more pages exist.
- **Observed:** 20 nodes, `hasNextPage = true`, `endCursor` is an opaque base64 cursor.

## Case B — GraphQL mutation happy path

We exercised the `checkout` GraphQL mutation happy path against a synthetic cart.

```graphql
mutation Checkout($cartId: ID!, $idempotencyKey: String!) {
  checkout(cartId: $cartId, idempotencyKey: $idempotencyKey) {
    order { id status totalCents }
    userErrors { field message }
  }
}
```

Variables: `{ "cartId": "cart_synthetic_77", "idempotencyKey": "idem-abc-123" }`.

- **Expected behavior:** `order.status = "PAID"`, empty `userErrors`, HTTP 200, no top-level
  `errors` array. Re-sending the same `idempotencyKey` returns the same order (no double charge).
- **Observed:** order created once, second call with same key returned the same `order.id`.

## Common bugs seen here

- Mutation returns HTTP 200 with a populated top-level `errors` array while `order` is null —
  clients that only check the HTTP status treat the failed checkout as success.
- `totalCents` silently changes from `Int` to `Float` after a refactor, breaking integer clients.
- Idempotency key ignored on retry, creating duplicate orders.

## Useful QA prompt

> "Run the `checkout` GraphQL mutation happy path twice with the same idempotencyKey and
> verify the second call returns the same order id, empty userErrors, and no top-level errors."
