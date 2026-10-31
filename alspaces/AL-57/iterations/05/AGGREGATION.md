# AL-57 iteration 05 domain aggregation cleanup

Approved scope: concern 4 of iterations/05/PLAN.md, baseline 412bc08e3199af04c03009ed60ca1bd3b592dd85. This implementation owns backend/internal/domain/aggregations.go, explanatory comments in cost.go, and affected domain aggregation tests. No migration production code changed. Initial Git status contained parent-owned SUMMARY.md and iteration 05 PLAN.md/REVIEW.md changes; all unrelated concurrent adapter, CLI, frontend, and workspace edits were preserved.

## Delivered changes

- accumulateInvocations merges SummaryForCost during its existing token loop. It no longer independently sums cost and then calls SummarizeCost for a second traversal. The rounded scalar is projected once from the resulting raw subtotal; counts, evidence and unknown/empty states are retained.
- ComputeTotals uses the accumulator's row count instead of grouping by role and running extra accumulations solely to count rows.
- Thread per-model metrics reuse accumulateInvocations. Session/role/model grouping, ordering by unrounded complete amounts, configured health ceiling, token/cache fields and four-decimal scalar projection are preserved.
- bucketAccumulator no longer stores a redundant costUSD subtotal. Bucket scalar amounts derive from their summary; cumulative scalar amounts derive from the same chronological summary merges. Bucket filtering, ordering and four-decimal rounding are unchanged.
- Comments clarify that EffectiveCost and EffectiveCostSummary are pure in-memory compatibility policies. Source/JSON presence is established separately before domain construction.

## Justified retentions

Fleet totalCostUSD, roleCosts and modelCosts still sum the already-rounded per-ledger scalar projections. CalculateRoleEffortRollup retains that same scalar summation alongside raw summaries. Two source ledgers reporting 0.00004 each have scalar ledger projections 0, fleet/role scalar sums 0, and an evidence subtotal 0.00008. Recomputing their scalar from the merged raw summary would produce 0.0001 and change accepted compatibility precision. Complete wire values remain the boundary projection's summary responsibility.

Public CostUSD fields, EffectiveCost, EffectiveCostSummary, and their independent call-site policies remain intact. A present legacy numeric zero is observed, an empty zero-count projection is empty, and source/JSON missing/null is explicitly unknown. No public type, function signature, wire contract or cost-evidence policy changed.

## Regression evidence

Existing TestComputeTotals's evidence table now also asserts rounded scalars and row counts, including complete and partial tiny contributions that merge before rounding. Existing TestComputeRoleAndModelMetrics's tiny-cost case retains unrounded ordering assertions and now checks scalar/evidence plus model/thread token/cache/health projections. TestConfigurableHealthyCeilingAggregations adds the missing thread per-model ceiling guard.

TestCalculateFleetSummary adds a real DeriveMetricsLedger-to-fleet/role-rollup precision oracle for the two 0.00004 source ledgers. It asserts rounded total/role/model scalar preservation, unrounded complete summaries, nonempty role token/cache/count/health rollup behavior, and legacy numeric-zero versus empty compatibility fallback.

TestAggregateMetricsOverTime_Hourly adds unordered source rows with complete tiny cost, partial tiny contributions, unavailable cost, observed zero and a code-only bucket. It verifies chronological cumulative precision, irrecoverable cumulative incompleteness, code-only empty evidence and coverage counts without copying the production calculation.

All original meaningful assertions remain. Tests use self-contained source values and do not read live workspaces.

## Verification

From backend, against the current domain diff:

- go test ./internal/domain: passed during implementation.
- go test ./internal/domain -run 'TestComputeTotals|TestComputeRoleAndModelMetrics|TestCalculateFleetSummary|TestAggregateMetricsOverTime_Hourly|TestAggregateMetricsOverTime_AdversarialHourly|TestConfigurableHealthyCeilingAggregations' -count=1: passed.
- go test -race ./internal/domain/...: passed after final test additions, including the unchanged migration suite.
- go vet ./internal/domain/...: passed after final test additions.
- git diff --check -- backend/internal/domain: passed.

These focused checks do not claim integrated adapter/frontend or full make verification. Parent owns the independent audit, final integrated make gate, telemetry, commits and PR delivery. No subagents, commits, push or telemetry were invoked by this implementation agent.
