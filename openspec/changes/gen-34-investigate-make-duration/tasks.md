# GEN-34: Execution tasks

Reference design.md for approved scope. This change has no delta specs because product behavior does not change.

## Investigation
- [x] Time every `make` recipe with full `make` runs under concurrent load and at low load.
- [x] Profile the Go integration lane per package and per test, with and without `-race`, in parallel and in isolation, with CPU accounting.
- [x] Profile the frontend unit and integration suites per file.
- [x] Reproduce timing-sensitive failures under controlled load and identify their causes.
- [x] Rank the improvement proposals by estimated gain and risk, stating the coverage each one keeps and loses.

## Delivery
- [x] Write the report and delivery summary; validate the change.
- [x] Generate final telemetry, commit, push and open the `chore(GEN-34)` PR without merging.
