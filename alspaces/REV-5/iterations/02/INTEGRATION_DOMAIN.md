# REV-5 domain upstream integration

Premerge aa86f4b; incoming main 47af923. Owned backend/internal/domain/** only. Preserved other agents' changes and the staged upstream merge; no staging, commit or publication performed.

## Conflict decisions and incoming-code review

- aggregations.go: retained removal of CalculateRoleEffortRollup. Its only new references were two incoming test cases; no production consumer was introduced. All incoming cost evidence and summary accumulation in active totals, role/model/thread, fleet, comparison and time aggregation paths remains.
- migration/engine_test.go: retained both REV-5's marshalTestLedger serialization-only fixture and the entire incoming TestCostEvidenceMigrationPreservesObservedHistoryAndMissingPresence coverage. Removed conflict delimiters without restoring MigrateLedger.
- Reviewed automatically merged cost.go, models.go, aggregation logic, migration/engine.go, cost_decode.go, step_v4_v5.go and validator.go. Preserved explicit observed/partial/unavailable evidence, complete/partial/empty/unavailable summaries, source-precision subtotals separate from rounded internal scalars, v4 reads, v4-to-v5 migration without new source-recovery requirements, and pre-v4 recovery requirements. Typed JSON presence decoding remains pure; implicit memberPresence.UnmarshalJSON is a runtime codec hook, not dead code.
- Found incoming SummarizeCost has no production callers and duplicates accumulation already performed by ComputeTotals. Removed it and moved its three test uses to ComputeTotals(...).CostSummary, coordinating the two HTTP fixture changes with the backend owner. No lost assertion or test-only copy of the removed implementation.

## Assertion recipients

- TestComputeRoleAndModelMetrics's incoming tiny-cost model-order case now uses ComputeThreadSummariesWithCeiling with the prior default ceiling. Retains complete-before-incomplete ranking, unrounded tiny amounts, rounded internal scalar behavior, cache/tokens/health and matching thread projections.
- TestCalculateFleetSummary's rounded-ledger-versus-source-precision case retains fleet, fleet-role and fleet-model scalar/subtotal assertions. The deleted role-rollup assertions now exercise ComputeRoleMetricsWithCeiling over the original invocation facts: counts, tokens, cache rate, peak, configured health and precise 0.00008 summary remain asserted. This active API rounds the combined scalar to 0.0001; the dead rollup's separate sum-of-already-rounded scalar contract was deliberately not restored. Fleet scalar remains zero, protecting its distinct documented compatibility behavior.
- The incoming legacy-zero/empty role-rollup subtest now checks CalculateFleetSummary.RoleCostSummaries directly using the same legacy RoleMetric fixtures. Observed numeric zero remains a counted complete amount; empty projections retain empty/null summaries.
- TestCostEvidenceMigrationPreservesObservedHistoryAndMissingPresence still checks v1-v4 missing/null/zero/positive costs, repeated row ownership, task auxiliary cost exclusion, v5 required members/contradictions, arbitrary retired metadata, ambiguous null collections, and no new recovery marker for v4-to-v5. Its task-exclusion summary assertion now uses ComputeTotals.

## Verification

Final go test -race ./internal/domain/... passed after all changes (domain and migration). Earlier compile checks caught three automatically merged calls to removed APIs; corrected by retargeting as above, not restoring unused production functions. Final domain git diff --check passed and no merge markers remain in owned files.

Reran the existing Go AST identifier inventory across 127 backend files after coordination. Only implicit UnmarshalJSON/UnmarshalYAML hooks remain with no explicit production identifier references. No other new domain test-only named function found. This is a conservative identifier inventory rather than a whole-program call graph. No new domain I/O, clocks, schema, financial inference or public API introduced. Root owns independent tester/reviewer assessment and the integrated gate.
