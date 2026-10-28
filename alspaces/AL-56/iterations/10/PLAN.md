# AL-56 Iteration 10 Plan: Close v4 Readiness and Regression Gaps

## Objective

Address only the three changes requested by the iteration 09 review: enforce complete canonical v4 structure and non-bypassable migration readiness, prove HTTP/API rejection of legacy versions without partial success, and verify that unavailable or failing source backfill remains pending and runtime-ineligible.

## Review Finding 1: Canonical v4 Structure and Readiness

- Replace the current metadata-only acceptance check in `ValidateRuntimeFactLedger` with validation of the complete canonical schema-v4 structure before any ordinary loader decodes or merges facts.
- Validate session-owned source metadata and the atomic `user_iterations` representation on every invocation. Distinguish an omitted JSON field from an explicit numeric zero; zero remains valid where the contract permits it.
- Define a durable source-backfill readiness invariant for migrated ledgers. Runtime readiness must not be inferred from an omitted `migration` object or omitted readiness metadata. Freshly generated ledgers and explicit migration/backfill must write the appropriate state, and shape-only migration must remain pending until source-backed backfill succeeds.
- Preserve actionable errors directing legacy/incomplete ledgers to explicit migration followed by source-backed backfill. Keep ordinary reads non-mutating and fail before returning or merging partial facts.
- Add validator and loader tests for missing session source, missing `user_iterations`, valid explicit zero, migrated-only ledgers, and omitted readiness metadata. Retain successful loading coverage for complete canonical v4 ledgers.
- In `ValidateRuntimeFactLedger`, bind the raw JSON bytes from `fields["sessions"]` to a local variable before decoding the sessions collection; do not reference an undefined `sessions` identifier.
- Repair the `missing source readiness` fixture by mutating `migration.steps_completed` to an empty array in the actual base JSON, then assert that validation returns the source-readiness error. Do not use a string replacement for `,"source_backfill"`, which is not a substring of the fixture and therefore leaves it unchanged.

## Review Finding 2: HTTP/API Legacy Rejection

- Add table-driven boundary tests using schema v1, v2, and v3 ledgers with recognizable legacy facts for workspace listing/scan and ticket detail/events routes.
- Assert each route returns a client-visible error with the detected schema version and migration plus source-backed backfill guidance.
- Assert rejected requests do not return a successful response containing an empty or partially interpreted ledger. Where workspace listing aggregates scan results, assert the invalid ledger causes the documented error response rather than being silently presented as a successful partial result.
- Keep endpoint behavior aligned with the shared runtime validator; do not add a legacy parser or endpoint-specific compatibility path.

## Review Finding 3: Source-Unavailable Backfill

- Add focused backfill cases at available code seams for unavailable/no matching source data and source failure. Add an extraction/parse-failure case only if the existing test infrastructure permits deterministic failure injection.
- For each covered failure case, assert the backfill report indicates incomplete/failure, `source_backfill` remains pending, and persisted migration state is not marked complete.
- Feed each resulting ledger through runtime validation/loading and assert runtime rejection. Verify the source ledger remains in its expected safe persisted state and no partial refresh is reported as ready.
- Retain the existing matching-source success case as the positive control.

## Acceptance Criteria

- Ordinary runtime validation accepts only structurally complete canonical v4 ledgers, including required session source and invocation `user_iterations` fields; explicit zero is accepted, omission is rejected.
- Migrated ledgers cannot become runtime-ready by omitting migration/readiness metadata. Shape-only migration stays pending until successful source-backed backfill records readiness.
- HTTP/API tests prove schema v1, v2, and v3 are each rejected on workspace listing/scan and applicable ticket detail/events routes with actionable client-visible errors and no partial-success payload.
- Backfill tests prove unavailable/no-match and source-failure cases leave `source_backfill` pending, report incomplete/failure, and produce ledgers rejected by runtime ingestion; extraction/parse failure is covered when deterministic failure injection is available.
- Complete canonical v4 and successful source-backed migration/backfill continue to pass their existing positive-path coverage; the iteration does not change unrelated extraction, Git, or OpenCode behavior.

## Scope Boundary

This is a planner-only delta for the three iteration 09 review findings. Implementation, test execution, unrelated contract changes, and broader cleanup are outside this plan.
