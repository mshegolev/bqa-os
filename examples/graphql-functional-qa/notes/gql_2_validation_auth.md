# GraphQL session 2 — Validation errors + role-based access (synthetic)

> Synthetic example only. No real schema, tokens, customers, or PII.

- **Service / schema:** `commerce-gateway` GraphQL schema, types `Order`, `Customer`, `AdminReport`.
- **Domain:** GraphQL functional testing — input validation and auth / role-based access.

## Case C — Validation error case

We sent a GraphQL mutation with an invalid input to confirm the schema rejects it before
any resolver runs. `first` is constrained to `1..100`; we passed `0`.

```graphql
query OrdersByCustomer($customerId: ID!, $first: Int!) {
  orders(customerId: $customerId, first: $first) {
    edges { node { id } }
  }
}
```

Variables: `{ "customerId": "cust_synthetic_001", "first": 0 }`.

- **Expected behavior:** the GraphQL server returns a top-level `errors[0]` with
  `extensions.code = "BAD_USER_INPUT"`, HTTP 200 (GraphQL convention), `data = null`, and a
  message naming the `first` argument. No order resolver is invoked.
- **Observed error shape:**

```json
{
  "data": null,
  "errors": [{
    "message": "Variable \"$first\" got invalid value 0; expected value in range 1..100",
    "extensions": { "code": "BAD_USER_INPUT", "argument": "first" }
  }]
}
```

- **Common failure mode:** validation is implemented inside the resolver instead of the schema,
  so the bad value triggers a 500 / `INTERNAL_SERVER_ERROR` instead of `BAD_USER_INPUT`.
  The error contract changes and clients can no longer distinguish "your input was wrong" from
  "the server broke".

## Case D — Auth / role-based access case

We tested role-based access on a GraphQL query that requires the `admin` role.

```graphql
query RevenueReport($month: String!) {
  adminReport(month: $month) { grossRevenueCents refundsCents }
}
```

- **As `viewer` role:** expected top-level `errors[0].extensions.code = "FORBIDDEN"`,
  `data.adminReport = null`, HTTP 200. The field must be nulled, not the whole response dropped.
- **As `admin` role:** expected populated `adminReport`, no errors.
- **Observed (regression we are guarding against):** as `viewer`, the resolver still returned
  `grossRevenueCents` — a **field-level authorization bypass**. Auth was checked at the route but
  not at the resolver, so a deep query reached the protected field.

## Common bugs seen here

- Auth enforced at HTTP route only; nested GraphQL resolvers leak protected fields.
- `FORBIDDEN` vs `UNAUTHENTICATED` confused — a missing token returns `FORBIDDEN`, a wrong role
  returns `UNAUTHENTICATED`, so clients can't decide whether to re-login or show "no access".
- Error message leaks the protected value in `extensions` even when the field is nulled.

## Useful QA prompt

> "Run the `adminReport` GraphQL query as each role; verify the viewer role gets
> data.adminReport = null with extensions.code = FORBIDDEN and no leaked revenue numbers,
> and only the admin role sees populated figures."
