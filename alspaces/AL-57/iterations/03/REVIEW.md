# AL-57 iteration 03: Independent typed-boundary audit

Status: APPROVED. No open findings. This is an implementation/structural audit, separate from the parent integration and delivery gate.

## Authority, snapshot and limits

Reviewed against baseline ee7558749c7101ddc4134cbdd22961bcba0a5a09 and approved CONTRACT.md / iterations/03/PLAN.md. The user accepted the concrete typed DTO alternative with "$agentic-lens-dev fix it" and included migration/cost_decode.go in the correction. Cost semantics, historical numeric observed compatibility, v4 reads, field order, migration ownership and frontend behavior remain frozen.

Reviewed all production changes and new untracked files in this delta, the removed generic transform bodies, and relevant existing/new test bodies. Backend/web source SHA256: 3edc1375f8dd66983375d8b73d3b9cb3f73b2cf7d67abe08878d932f3cd591f1 across 219 existing source paths (deleted tracked files are excluded). Method: sorted unique paths from git ls-files -co --exclude-standard -- backend web, hashing each existing file's path, NUL, bytes, NUL. Ignored generated assets/dependencies are excluded. This matches the tester's final focused snapshot. No source or test edits by this reviewer.

## Axis 1: Structural and contract alignment

The correction removes the approach the user rejected. adapters/cost_json.go and ordered_cost_json.go are deleted. No generic cost-tree traversal, first-byte object/array dispatch, recursive byte rewriting, reflection, ordered-object framework or classification of arbitrary objects by tokens/model/role/task field names replaces them. Existing schema-validation probes inspect named fields at known canonical locations, which is the approved distinction.

The new adapters/cost_fact_dto.go invocation/task boundary has nullable amounts and narrowly scoped presence codecs. fact_ledger_decode.go selects concrete canonical-v4 versus canonical-v5 policy before conversion. Legacy numeric zero/nonzero is observed, absent/null is unavailable, and current v5 requires both named cost members and valid evidence. Missing required keys remain distinguishable from explicit null. Unrelated unknown fields remain strict at the runtime adapter boundary, with existing retired-field codecs retained.

Domain/migration/cost_decode.go now decodes concrete schema-v1-to-v4 or schema-v5 inputs and directly assembles known metadata, sessions, invocations and task links. There is no intermediate rewritten JSON document or adapter import. Its memberPresence probe handles named root/session collection presence, including explicit null mixed-shape guards and removed canonical invocation id members. Legacy row owner metadata and all repeated occurrences remain intact. Unknown/retired migration metadata keeps its prior tolerant behavior and is not mistaken for cost records. Existing source normalization, anomaly sanitation, dropped-tool diagnostics, chain recovery and invariants remain in engine/step owners.

Shared domain.ValidateCostEvidence is pure: finite/nonnegative checks and observed/partial/unavailable validation operate on nullable values, with no JSON, file, database, clock or outer-layer dependency. Both adapter and migration decoders call it before projecting wire values into internal accumulators. A partial subtotal enters only internal arithmetic; it does not become a complete wire amount.

## Every affected output boundary

| Prior generic projection caller | Concrete replacement checked |
| --- | --- |
| Fact DTO MarshalJSON / SaveFactLedger / migration runner / backfill persistence | AgenticFactLedgerDTOFromDomain with InvocationFactDTO and TaskLinkFactDTO; direct single typed marshal/indent, original field order/newline preserved |
| Metrics export MarshalJSON / SerializeLedger / SaveMetricsLedger / CLI JSON | AgenticMetricsLedgerDTO with concrete workflow/grouped/role/model DTOs; export remains compact and excludes HTTP-only facts/threads |
| HTTP workspace list | MetricsLedgerResponseDTOsFromDomain, including typed facts, sessions, invocation/task evidence, threads and all rollups |
| HTTP workspace detail | Handler constructs MetricsLedgerResponseDTO before WorkspaceDetailResponse; artifacts/anomalies keep their original fields |
| HTTP fleet | FleetSummaryDTO; nullable total and role/model maps have their corresponding summaries; unaffected metrics copied directly |
| HTTP comparisons | TicketComparisonRowDTO; nullable complete cost and coverage, original remaining fields/order retained |
| HTTP progression | TimeBucketMetricDTO; bucket and cumulative amounts/summaries remain independent, with incomplete cumulatives staying null |
| Error/build/config/events/live tree/live trace/live anomaly families | Closed ResponsePayload union still contains these concrete pass-through families; WriteJSON simply marshals its typed input and performs no cost projection on unrelated/source payloads |

Checked the declared JSON tags/order for GroupedMetric, RoleMetric, ModelMetric, WorkflowEfficiency, ThreadRoleMetric, ThreadExecutionRecord, FleetSummary, TicketComparisonRow and TimeBucketMetric against their explicit DTOs: all match. Checked separate persisted fact and HTTP fact/session shapes, including retained sparse events, anomaly facts, nil/empty collections and optional task fields. Real maps use deterministic standard encoding; preserving arbitrary input JSON-member encounter order is no longer a requirement of this refactor.

## Assertion transfer and removal audit

Inspected the deleted TestCostEvidenceBoundaryCompatibility arbitrary-normalizer subtest and the recipients below; no score/count quota was applied.

| Removed or moved assertion | Disposition / coherent existing recipient |
| --- | --- |
| Arbitrary zeta/rows/alpha/tail input order and arbitrary map encounter order | Intentionally removed with the unsupported generic API under PLAN.md; these are not real generated DTO structures |
| Partial complete amount null, known contribution retained, observed zero preserved | TestSaveMetricsLedger actual typed persistence plus HTTP output; TestCostEvidenceBoundaryCompatibility concrete invocation/task roundtrip; TestResponseHelpersAndSerialization real response families |
| Role-map null/observed-zero values and unavailable percentage nulls | TestResponseHelpersAndSerialization typed FleetSummary/WorkflowEfficiency response case, including original non-cost metrics and reason |
| Repeated byte stability/deterministic new evidence/generated field order | Existing TestSaveFactLedger and TestSaveMetricsLedger repeat actual persistence and compare bytes; existing field-order helpers assert real emitted DTO subsequences |
| Missing legacy cost becomes unavailable without inventing zero | Retained original canonical-v4 numeric/zero/null/omitted table; concrete task table and migration cost-presence/version/task cases |
| Malformed/trailing JSON rejected | Actual canonical-v4/v5 decoder negative cases and migration TestDetectVersion; arbitrary malformed generic-object transformation is intentionally removed |
| V5 required-member/contradictory evidence rejection | Retained negative table plus omitted cost, null evidence, concrete optional task and migration canonical/flat/task guards |
| HTTP facts/threads retained while export stays compact | TestSaveMetricsLedger and existing TestCLIJsonOutput; real workspace detail/events/baseline-preservation bodies retain their previous assertions |
| Unaffected source text with cost-like keys | Existing WriteEventsResponse case roundtrips the verbatim source string; static structural audit confirms no arbitrary source objects are walked |

Also reviewed the typed migration owner’s task/version/strict-v5 additions and the tester's removed-id-presence guard. Existing ownership/multiplicity, anomaly/event sanitation, v4 read/no-rewrite and cost summary cases remain. The new tests execute production serialization/decoding paths and independent expected values, rather than only assigning DTO literals or snapshotting implementation source.

## Axis 2: Engineering standards

Boundary conversion and presence handling remain in adapters or the existing pure migration decoder; no new dependency points from domain into adapters. Nullable amounts are concrete pointers, evidence/summary types are canonical, and HTTP payloads use the closed concrete union. No new any/interface{} or generic serializer framework is introduced. New cost validation has explicit errors with caller context. Frontend source is unchanged. docs/COST_SEMANTICS.md describes the typed order-preserving boundary implementation without changing the financial contract.

## Verification reuse

The tester reports both non-race and race-enabled focused adapters/HTTP/migration/CLI owners passed: TestCostEvidenceBoundaryCompatibility, TestSaveFactLedger, TestSaveMetricsLedger, TestResponseHelpersAndSerialization, TestCanonicalV4MigrationRejectsRemovedInvocationKey, TestCostEvidenceMigrationPreservesObservedHistoryAndMissingPresence and TestCLIJsonOutput. Adapter coder reports adapter/CLI tests and full Go vet passed. Migration coder reports domain/migration/domain/backfill race tests and targeted vet passed. No silent skips or production defects were reported. git diff --check passed during this review.

These valid checks were reused rather than duplicating them without a concrete concern. Parent still owns current-main integration, the required full make gate, standalone/embedded-asset verification, refreshed real session telemetry, commits and updating the existing PR. Previous iteration full gates are historical evidence, not proof of this refactor's final integrated result.

Final verdict: APPROVED.


## Parent integration and binary gate

Reviewed implementation commit 3c8f09331c925cf0c2ecc8c4d78ada64414ef099 integrated current origin/main 36bea1f with no new changes. Required full make passed: Go vet and all 12 packages under race, frontend typecheck/lint, 755 tests in 35 files, canonical assets and four platform builds. Standalone linux/amd64 smoke passed exact current HTML/JS/CSS, ordered pass-through version JSON and empty fleet null cost/empty summary with Node unavailable. Asset trees match. Remaining edits are final telemetry and delivery records; no application changes followed this gate.
