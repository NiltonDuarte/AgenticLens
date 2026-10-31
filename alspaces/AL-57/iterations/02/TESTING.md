# AL-57 iteration 02: Adversarial verification

Status: PASS after four production corrections. This is testing evidence, not the independent reviewer verdict or final integration gate.

## Tested source

Baseline commit: 45cb679f21475cd14b7dd4cbc25996a39d734403, with the backend/frontend implementation and tester changes in the shared working tree. The first approved CONTRACT.md and iterations/02/PLAN.md govern; iteration 01's proposed historical special classification is superseded. Valid legacy numeric costs, including zero, are observed.

Backend/web source-tree SHA256: 4ea3cf6ce5f8e522c79c860e089054011150a494de222f7a5baae63e30ef2ab8. Fingerprint covers 217 unique paths from git ls-files -co --exclude-standard -- backend web, sorted by path, hashing each path, NUL, file bytes, NUL. Ignored generated assets and dependencies are excluded. The make test run includes corrections 1 and 2 below. A later reviewer precision correction changes backend/domain code and tests; focused and full affected-package race checks passed on the newer source snapshot.

Latest backend/web source-tree SHA256 using the same 217-path method: 066f35392987dd9dd28a1deda735be1e851a013c1c2c2b9e4faf978140470391. The parent still owns the final integrated full make gate.

## Coverage inspected and extended

Existing frontend guards already covered subtotal/count formatting, cumulative incompleteness and date filter recomputation, both cost sort directions, chart gaps, known-zero scatter coordinates and non-cost axes. Existing backend additions covered coalesced mixed source rows, aggregate summary algebra, v4 numeric/null/omitted decoding, strict v5 contradiction rejection, and v4-v5 migration. Tests were extended at these coherent owners:

- web/src/api/client.test.ts: actual workspace API client accepts v5 observed zero, partial and unavailable facts without coercion; rejects missing/contradictory fact evidence; rejects null role-map values without coverage metadata.
- web/src/components/Components.test.tsx: a partial known-zero subtotal remains visible while an incomplete fleet denominator makes all shares unavailable and hides percentage bars; complete known values precede missing/partial values.
- backend/internal/adapters/http/server_test.go, TestGetFleet: a self-contained custom workspace containing canonical-v4 observed zero, canonical-v4 missing amount, and v5 partial source cost exercises real server routing, repository reads, derivation and JSON serialization. Complete fleet/role amounts are null, known subtotal is 2, invocation counts are 1 observed/1 partial/1 unavailable, both cost order directions keep complete zero first with stable incomplete ties, and GET leaves all input bytes unchanged.
- backend/internal/domain/migration/engine_test.go: v1-v3 flattened inputs preserve numeric zero/nonzero and null/omitted presence through the full v5 chain, including four legitimate rows with one owning session.
- backend/internal/domain/aggregations_test.go: existing role/model owner now verifies 0.00001 versus 0.00002 ordering across model and thread projections, and complete costs rank ahead of incomplete known contributions. Existing ledger sort owner verifies tiny distinct costs in both directions; the old timestamp-reversal tie expectation is replaced with the approved stable-identity rule.
- Both OpenCode schema-family extraction suites: existing mixed-row cases now also exercise explicit observed zero, zero-plus-null partial evidence, all-null unavailable evidence, and rejection of negative source amounts through real SQLite extraction.

No production code, domain mocks, live ticket workspaces, or actual source databases were written by the tester. No tests were removed or skipped. Synthetic source records establish deterministic behavior, not real-session spending evidence.

## Failed-then-fixed findings

1. Frontend boundary accepted role_costs containing null without its role_cost_summaries coverage map. A focused existing API-client regression failed (85 passed, 1 failed). Frontend coder corrected validation for null map entries and mismatched/orphan summary maps; the unchanged regression now passes. Numeric legacy maps remain supported.
2. The real HTTP descending cost test returned partial/missing workspaces before observed zero. The cause was ParseLedgerSortField rejecting the UI/API cost_usd key and falling back to updated_at; the cost comparator itself already handled availability. Backend coder added cost_usd as an alias. The unchanged HTTP regression passes in both directions.

3. Independent reviewer found model/thread/ledger sort comparisons using rounded internal CostUSD despite distinct complete wire amounts. Backend coder corrected comparisons to use unrounded CompleteAmount and stable identity ties; tester added the coherent precision cases described above and updated the conflicting old timestamp tie expectation. Full domain and HTTP race packages passed afterward.

## Commands and results

- npm test -- --run src/api/client.test.ts src/components/Components.test.tsx: 121 tests passed after the frontend correction.
- npm run typecheck and npm run lint: passed after tester frontend additions and the correction.
- go test ./internal/adapters/opencode -run 'TestSchemaV[12]CostEvidenceCoalescesReportedAndMissingRecords' -count=1: passed both schema families and added source-presence cases.
- go test ./internal/adapters/http ./internal/domain/migration -run 'TestGetFleet|TestCostEvidenceMigrationPreservesObservedHistoryAndMissingPresence' -count=1: migration additions passed; the first HTTP run exposed finding 2, and backend coder's focused rerun passed after its production fix. The complete race run below independently passed the final HTTP test.
- make test: passed. Canonical frontend build completed; go test -race ./... passed all 12 Go packages; frontend Vitest passed 35 files and all 754 tests, zero skipped.
- diff -qr web/dist backend/internal/adapters/web/dist: passed after the current frontend build.
- go test -race ./internal/domain ./internal/adapters/http -run 'TestComputeRoleAndModelMetrics|TestComputeThreadSummaries|TestSortLedgers|TestTimeBoundaryAndSortingEdgeCases|TestGetFleet' -count=1: passed after precision corrections (the named time-boundary pattern matches no test; the other named owners executed).
- go test -race ./internal/domain ./internal/adapters/http: both full affected packages passed after the precision correction, including the actual TestAdversarialDomainAggregations tie owner.
- git diff --check: passed.

## Limits

Existing MetricProgression tests print React act warnings; no stage failed or skipped. No browser geometry/accessibility beyond tested DOM interactions is claimed. Full make with static checks, final upstream integration, standalone binary smoke checks, source-grounded final telemetry, commit and PR delivery remain parent-owned. Re-run affected verification if subsequent review changes source; the recorded source fingerprint allows unchanged results to be reused.


## Final reviewer task-cost correction

Reviewer found that a valid partial task-link cost was formatted as unavailable even though its reported subtotal was present. Frontend coder changed the owning formatter/panel to display the known subtotal with a partial label and no invocation-coverage claim. Tester extended the existing TicketDeepDive skills/subagents suite with that exact auxiliary-cost case and extended the existing DeriveMetricsLedger invariant owner to prove a partial task amount does not enter invocation totals or coverage. No production edits were made by the tester.

Rechecks:

- Six selected Skills/subagents cases passed; the filtering run explicitly deselected 58 other tests. The full owning file then ran: all 64 TicketDeepDive tests passed, zero skipped.
- npm run typecheck: passed after the new panel case.
- go test -race ./internal/domain -run TestDeriveMetricsLedger_EventsPreservationAndInvariants -count=1: passed after correcting the test's initial omitted ProjectConfig argument.
- git diff --check: passed.

Latest 217-path backend/web source fingerprint after this correction: 82d6d7261258a2b33dd8350b64a5215185c2e24929a5f5301e0f25d9f6ac5c81. This supersedes earlier source fingerprints for current affected verification. The parent runs the complete integrated make gate, including the now-755 frontend tests, after integration; earlier full-suite evidence is kept distinct above.
