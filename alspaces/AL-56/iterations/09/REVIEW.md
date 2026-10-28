# AL-56 Iteration 09 Review

## Verdict: CHANGES_REQUESTED

### Findings

1. **Source-backed readiness can be claimed without validating completeness.** `ValidateRuntimeFactLedger` accepts any schema-v4 document with a `sessions` key when the `migration` object is absent, and otherwise only checks migration status and pending steps. It does not validate the v4 session-owned/atomic fields described by the contract (including session source and required `user_iterations` representation). Consequently an incomplete or shape-migrated ledger with readiness metadata omitted can pass ordinary ingestion. Add canonical v4 structural validation and a readiness invariant that cannot be bypassed by omitting migration metadata; cover missing source, missing atomic fields, and migrated-only ledgers.

2. **Required server/API regression coverage is missing.** The added focused tests exercise the validator for v1-v3 and incomplete v4, but there is no boundary test proving each v1, v2, and v3 ledger is rejected through the HTTP/API routes with a clear client-visible error and no successful partial response. The contract explicitly requires coverage for each legacy version at server/API boundaries. Add table-driven tests for list/scan and ticket detail/events behavior as applicable.

3. **Source-unavailable backfill outcomes are not tested.** The new happy-path test proves a matching OpenCode session can make a ledger ready, but does not exercise no matching source data or source read/extraction failure after schema migration. Since `dbSessionFound` is set only after successful extraction with invocations, these cases should retain `source_backfill` as pending and remain rejected by runtime ingestion. Add explicit tests asserting the report is incomplete and the persisted ledger remains non-ingestible.

### Review scope

Read-only review of the contract, iteration plan, technical guidelines, and working diff. No tests were run, per instruction. No implementation files were changed.
