# AL-57 iteration 05 frontend validation

Baseline: 412bc08e3199af04c03009ed60ca1bd3b592dd85. Parent planning artifacts and concurrent backend/domain edits preserved. This agent owns only frontend API production/tests and these notes; no commits, push, telemetry or PR action performed.

## Implemented boundary

Removed recursive validateCostPayload and generic readCostResponse. The client now explicitly selects readFleetSummaryResponse, readMetricsLedgerArrayResponse, readWorkspaceDetailResponse, readComparisonRowsResponse or readTimeBucketsResponse. Named schema helpers validate only declared metric, workflow, thread/model, fact-ledger/session/invocation/task and map locations. No arbitrary object walk or property-name financial classification remains.

Dynamic model, role and role-effort map keys are opaque names. Paired summary values are validated by matching own keys; cost-like names do not select a different schema. Source events, artifact content and unrelated metadata are not traversed or rewritten. Root response families, return objects, diagnostics and existing error propagation remain intact.

Fact-ledger version chooses strictness. Schema v5 requires own cost_usd and cost_evidence on both invocation and task rows and validates null/value/evidence combinations. Canonical v4 and older numeric responses without an explicit version retain optional cost presence: numeric zero remains observed compatibility, absent/null is retained without browser coercion. Explicit unsupported fact-ledger versions are rejected rather than treated as an implicit v4 policy. Numeric aggregate responses still need no newly added summary; null aggregate/map amounts require the accepted coverage summary.

## Verification

- npm run typecheck: passed.
- npm run lint: passed.
- npm test -- src/api/client.test.ts: 100 tests passed, no skips. These exercise actual client entry points with isolated mock HTTP responses, not direct guard calls.
- Existing API happy/error/diagnostic tests remain intact. New probes cover opaque model/role keys, strict v5 task fields, invocation omissions, numeric/null/omitted v4 presence, opaque source event/artifact payloads, all ledger-array callers, comparison-array coverage and cumulative time-bucket coverage.
- npm run build: passed. diff -qr web/dist backend/internal/adapters/web/dist: passed; current embed assets match and remain ignored under the existing policy.
- git diff --check: passed.

Production frontend diff is confined to api/client.ts and api/costValidation.ts. Iteration04 card and average behavior is unchanged. All frontend test paths have been released to the independent tester after the 100-test pass. Full integrated make and standalone-binary verification remain parent-owned. No browser or geometry claim is made.
