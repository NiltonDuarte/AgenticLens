# AL-64 iteration 03 architecture discovery review

Outcome: CHANGES_REQUESTED. One canonical-output contract gap is substantiated below. The canonical domain and frontend are genuinely numeric; no hidden mandatory rich-duration graph remains.

Run: lean_arch_sol, independent first-pass architecture discovery. Candidate LEAN-01 on baseline c68fbc9079d2f24c624d5de0fe87d313ed095b19. Assigned model/effort: gpt-6.1-sol/high, inherited settings; independently queried runtime model metadata is unavailable. Requirements: approved iteration 03 PLAN and frozen review brief, including accepted source-state estimation and the supersession of reasoning/liveness proposals. Sibling review reports were not read.

Frozen input manifest: /tmp/al64-03-review/fingerprint.json, SHA256 b05c2065769609e7e29cbb7e4b8158d804fc2b38ca905cea9b6a2bdbd1f237b5. All 399 listed hashes matched before probes and at final verification. This is a risk-directed review of changed paths and consequential unchanged dependencies, not a whole-repository audit.

## Finding ARCH-SOL-01: completed-migration cleanup republishes retired duration objects

Severity: 5/10. Classification: introduced contract coverage gap in an existing output branch; blocking the approved canonical-write acceptance criterion. Numeric facts are preserved, so this finding does not assert duration data loss.

Primary location: backend/internal/adapters/backfill/current_ledger_preparation.go:49. Consequential consumers: backend/internal/adapters/backfill/engine.go:280 and backend/internal/adapters/backfill/migration_runner.go:239.

Trigger: a supported schema-v6 ledger contains completed migration metadata and a valid draft duration_evidence invocation. Run ordinary backfill without source/Git recomputation, or explicit migration to v6 to clean up the completed metadata.

Expected: a published target retains owner/source identity and total/provisional numbers, removes migration metadata, writes the positive flat provisional_seconds contribution and omits duration_evidence and its retired state/completed_seconds/cutoff/basis members. The approved iteration requires this normalization for owned retained facts and explicit migration outputs without source availability prerequisites.

Observed: prepareCurrentLedger obtains a normalized domain ledger through Engine.Migrate at line 30, but builds WithoutMigration separately from the original compatibility input. ValidateRuntimeFactLedger accepts that valid old representation, while FormatLedgerJSON only applies whitespace policy. Both cleanup publication branches use WithoutMigration instead of the normalized duration representation. They therefore remove migration metadata while republishing the rich draft and omitting its flat numeric replacement. Subsequent ordinary backfill can skip this now metadata-free ledger, leaving the retired representation in place.

Independent disposable probe: /tmp/al64-03-review/architecture-sol-probe/main.go imports the actual candidate adapters and invokes BackfillEngine.Execute and MigrationRunner.Run. The fixture has one valid unavailable-source owner, active_duration_seconds=10, old completed_seconds=3 and provisional_seconds=7, with valid cost/identity/required collections. ValidateRuntimeFactLedger accepted the original fixture. Both operations reported one successful update/migration, with no error. Both resulting files removed migration but retained the whole duration_evidence object; the invocation-level provisional_seconds member was absent. Re-reading through the DTO recovered 10/7, confirming preservation and distinguishing the output-shape defect from numeric corruption.

Probe output files:

- /tmp/al64-03-review/architecture-cleanup-2325183000/TEST-LEAN-CLEAN/agentic_metrics.json (backfill).
- /tmp/al64-03-review/architecture-cleanup-3363021798/TEST-LEAN-CLEAN/agentic_metrics.json (explicit migration).

Required correction: apply the shared duration-only normalization to the raw cleanup candidate before publication, preserving unrelated recorded facts and optional-category presence. Extend the existing completed-metadata cleanup tests for both workflows with a positive draft contribution, zero omission and numeric/owner preservation. The existing tests cover metadata removal and raw-fact preservation but their invocation fixtures lack the draft duration object, so they cannot detect this gap. Do not replace this with a fresh source requirement or a whole-ledger domain rewrite that changes unrelated recorded distinctions.

## Consequential hypotheses and disposition

- Hidden rich canonical model: refuted. InvocationFact, WorkflowEfficiency, FleetSummary, TicketComparisonRow and TimeBucketMetric contain flat int64 ProvisionalSeconds fields. Old state/basis/cutoff/completed graphs occur only in backend migration/input compatibility types. DTOs and frontend types consume scalars. The cleanup finding concerns an output branch preserving raw old wire data, not a parallel domain graph.
- Impure arithmetic or a hidden clock: refuted in inspected duration paths. domain/duration.go receives milliseconds and cutoff explicitly, rounds once at invocation scope, and assigns the rounding remainder to the included provisional portion. Codex and both OpenCode source adapters own source activity/terminal markers and transient accumulation. Their booleans remain necessary for completion timestamps; they do not recreate a persisted evidence graph. Domain aggregation uses no source lookup or clock.
- Double counting through derived metrics: refuted. Workflow/fleet/comparison/progression computations independently sum existing totals and their included provisional portions. No code adds the provisional subtotal to active_duration_seconds. HTTP/metric DTOs copy the scalar directly.
- Frontend legacy conversion or average mismatch: refuted. duration.ts validates safe nonnegative included integers and describes absent/zero values neutrally. TicketComparisons uses filteredRows and Math.round for both average total and average provisional portion, including the empty case. Progression carries only the numeric subtotal. No reasoning model or configuration dependency was found in the scoped production paths.
- Retained-history ownership or shared evidence-pointer mutation: no new defect established. The numeric component copies by value through invocation/session DTOs and ordinary output projections. Existing capture_history/repository responsibility keeps baseline DTOs separate from successful exact-owner source replacement. The supported native force-only normalization and source-refresh evidence also independently demonstrates ordinary-path identity/numeric preservation. The completed-migration raw-copy branch is the substantiated exception.
- Exact-key normalization versus Go case folding: a disposable direct Engine.Migrate probe showed that noncanonical collection names such as Invocations can bypass raw duration traversal. For actual CLI current-ledger publication, prepareCurrentLedger validates original lowercase-required sessions/invocations through ValidateRuntimeFactLedger and rejects that malformed shape. This direct-helper observation does not establish a supported publication regression and is not counted as a finding. The supplied duration-member alias fix itself was inspected; its discovery remains credited to the implementation owner.
- Accepted open-source-state estimate risk: respected as approved behavior. This review does not require process liveness, caller proofs or new source authenticity receipts. Missing/zero provisional values are not interpreted as completed-source coverage.

## Verification and limits

Executed independently: both 399-input manifest checks; source/diff/caller tracing; disposable Go migration alias probe; actual BackfillEngine and MigrationRunner completed-metadata publication probes. All probe writes were under /tmp, with only this assigned report written in the checkout. Product code, tests, staging and existing user/agent changes were preserved.

Reused verification from TEST_REPORT.md: frontend 868 tests across 38 files plus typecheck/lint; correction-aware focused Go race checks covering DTO/migration/retained history/capture cache/source behavior; prior unchanged source-family and optional-category/overlap evidence. Broad pre-alias go test -race ./... remains a failed full run because of the CLI installation context deadline; its unchanged isolated retry passed. Neither is represented here as the required integrated make gate.

Reused NATIVE_REFRESH_EVIDENCE.md with its stated intermediate-binary boundary: authentic 19-owner draft force-only normalization preserved all owner facts under the numeric projection, then real Codex recomputation retained those owners and legitimately discovered a twentieth without cumulative duplication. That experiment covers the ordinary force/source-refresh paths; it does not cover the completed-migration cleanup branch reproduced here. No new native OpenCode or liveness experiment was run.

Remaining work: parent validates/consolidates ARCH-SOL-01, assigns its fix and targeted regression verification, then performs integrated make and final standalone asset verification. This verdict covers the frozen architecture scope only; no integration, release, telemetry or PR approval is implied.
