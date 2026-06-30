# BQA Team Smoke Test

This page records a docs-only smoke test for the BQA Team pipeline.

The pipeline is expected to complete the safe-change loop:

1. Create a working branch.
2. Make a small documentation change with synthetic content only.
3. Run the repository checks, including `go test ./...`.
4. Open a pull request for review.
5. Pass review without product logic changes.

The smoke test is intentionally narrow. It must not add Go application code,
core use cases, command wiring, generated product docs, private data, real
session logs, or secrets. Product work belongs in its own issue and branch.

## How to run and verify

1. Pick a `bqa:ready-dev` smoke-test issue (such as this one) and branch off
   `origin/main` with a `feature/issue-<n>-...` name.
2. Apply a safe docs-only edit to this file. Touch no Go, JS, or generated
   product files.
3. Run the repository checks and confirm they pass:

   ```sh
   go build ./...
   go test ./...
   ```

4. Commit, push, and open a pull request that references the issue.
5. The change is verified when the PR shows only this docs file changed,
   `go build ./...` and `go test ./...` pass, and review approves without
   requesting product logic changes.
