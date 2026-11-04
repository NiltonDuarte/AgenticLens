# AL-62 iteration 03 testing

Current acceptance is final lean schema 6, directly migrated from released schema 5. Schema 6's earlier rich form and schema 7 were unreleased ticket candidates, not compatibility inputs. The final schema-numbering correction and current verification appear at the end of this document; prior freeze commands/hashes remain historical evidence for their actual reviewed candidates.

## Authority and ownership

Approved lean scope in CONTRACT.md and this iteration's PLAN.md, with the final direct schema-5-to-6 correction governing current numbering. Initial baseline: 5e4eb34019370150786085d69c47987b328e4e79. Initial observed changes: parent-owned CONTRACT.md, SUMMARY.md and PLAN.md only. Parent explicitly transferred backend tests and runtime fixtures to tester after initial inventory. Later transfers assigned all HTTP tests to the frontend coder and all CLI tests to the backend coder; the latest correction returns all backend tests except HTTP to tester. Product code remained read-only to tester. Parent owns integration, review, telemetry and delivery.

## Existing regression recipients and assertion transfers

| Boundary | Existing recipient | Preserve or change |
| --- | --- | --- |
| Native Codex extraction | adapters/codex/rollout_test.go: native calls, inherited occurrences, distinct attachments, malformed recognized/outer payloads | Keep exact output pairing, owner, distinct equal-looking occurrences, replay exclusion, one token-plus-attachment selection, no catalogue/quoted/plain-read activation. Move native ID assertions to ephemeral source evidence/events; lean rows retain source start time and nullable role/type. |
| OpenCode extraction/capture | adapters/opencode/schema_v1_session_extraction_test.go and schema_v2_session_extraction_test.go | Keep both real schema-family fixtures, exact owner, explicit child identity, optional malformed-input behavior and capture-path coverage. Remove occurrence outcome or provenance serialization expectations only. |
| Actual generate portability | cmd/agentic-lens/owned_refresh_regression_test.go, capture_history_test.go, capture_test.go | Keep actual CLI save/reload, no private registration or pinned-file prerequisite, unavailable same/different harness owners, equal/older/fewer/empty complete snapshots and repeat stability. DATA-01 malformed arguments/numeric call_id/name must fail without changing accepted ledger bytes. |
| Actual backfill CC-02 | adapters/backfill/engine_test.go: TestBackfillCodexCompleteParentRefreshesWhileMissingChildFactsSurvive and recovery tables | Keep complete parent replacement independent of missing child; child lean facts, proven IDs, tree, metrics and aggregate survive two actual Execute/save/reload passes. Remove only retired child provenance oracle. |
| Lean JSON and migration | adapters/schemas_test.go, domain/migration/engine_test.go and lean_v6_test.go | Keep v5 ownerless deletion/audit and owned multiplicity in the direct final6 transition. Exact three/four-field shapes, required-null matrix, retired-key rejection, conversion audit, exact historical child proof, summaries/cost/invocation/tree preservation and source-free round trip remain. |
| Child identity | domain/skills_tasks_test.go; adapters/schemas_test.go namespace persistence regression | Runtime saved agent_path/reference join is retired. Transfer unique/ambiguous/wrong-parent exact namespace controls to source extraction and v6 migration. Persist proven native child ID; opaque handle alone yields no link. Retain unavailable proven child and no prompt/sidecar payload. |
| Source-only hydration | adapters/live/sqlite_test.go | Replace committed-event fallback/full-tree veto with complete, partial and unavailable coverage; successful metadata-only source is complete empty. Keep real SQLite/Codex detail, source selection, role, hierarchy, anomaly sequence and no-input-mutation/no-file-write assertions. |
| HTTP GET | adapters/http/server_test.go | Replace saved-event baselines/echo reconciliation with explicit transient availability and coverage. Keep detail metrics/lean facts/summaries, source failure and empty success, partial events, malformed-current sibling isolation, path guards and recursive byte snapshot no-write proof. |
| Anomalies/aggregation | domain/events_test.go, adapters/schemas_test.go sparse-events regression, cmd/agentic-lens/main_test.go aggregate regression | Transfer full-source generic failure/interleaved skill control to explicit ephemeral input. Preserve unavailable/partial owner summaries and whole-ticket aggregate across refreshed plus retained owners; do not synthesize sequences from lean facts. |
| Frontend | Frontend owner, all web tests | Counts/filters, nullable fields, unavailable child navigation, persisted lean panels before/after hydration, successful empty and partial traces, no skill outcomes/rates, no canonical events. Backend wire fixture handoff follows settled response. |

## Retired oracles

Skill execution success/unknown denominators, delegation execution outcomes/costs and native provenance fields no longer belong to persisted occurrence tests. Source tool event status remains meaningful transient evidence. Saved minimal occurrence timelines, includes_baseline, events_are_partial and runtime agent_path joins are deliberately removed; they are replaced by strict current-field rejection, migration audit and source-only availability assertions. Earlier BEH-03's ban on persisting a uniquely proven derived child UUID is superseded by the approved lean tradeoff: unique exact proof can persist the UUID without claiming it was returned raw. Ambiguous/unproven links remain absent. No old assertion is retained by introducing private authenticity gates, hidden native IDs or replacement persistence.

## Implemented missing regressions

- `schemas_test.go`: exact nullable current skill/delegation matrix, missing/empty/null/wrong-time boundaries and retired-key rejection. Session/root retired event, partial marker, task and agent-path members are rejected even when null/false/empty.
- `migration/lean_v6_test.go`: actual public Engine direct v5-to-final6 conversion preserves duplicate skills, nullable missing values, native ownership/tree, exact unique child proof, unavailable proven IDs, unrelated invocation costs/tokens and partial-source cached summaries. Ambiguous/wrong-parent links remain absent. Independent expected audit includes 3 skills, 4 delegations, 1 event, 1 partial marker and 2 unresolved references; current re-migration remains byte-stable. The reviewed fixture initially counted 19 rich fields; correction coverage removes the wrong-parent row's opaque reference to prove explicit-ID-only rejection/audit, giving 18 rich fields without changing occurrence counts.
- `live/sqlite_test.go`: actual SQLite/Codex failure table covers nil service/hook, missing session/source, closed database, missing tables/persona column, successful empty source and usable partial source. Coverage, diagnostics, source payload, stored summaries, input immutability, ledger bytes and absence of new workspace files are checked.
- `live/lean_hydration_test.go`: real mixed SQLite/Codex hydration retains tool input/output, native roles/call identity, complete/partial owner coverage, repeat stability and supported source rows after one source disappears. Lean stored rows never become synthetic traces.
- `capture/lean_snapshot_test.go`: the initial rejection-only test was superseded after ARCH-01 established that it blocked recovery. Current regression verifies metadata discovery, unrelated-ticket isolation, callback recovery and complete native replacement while preserving the old opaque body until success; it never treats old cached facts as current empty snapshots.
- `domain/events_test.go`: transient response validation distinguishes complete-empty, partial with supported coverage/rows, unavailable and invalid committed fallback/foreign/duplicate owner claims. Parent rework and root summaries survive incomplete group coverage; complete-empty whole-group coverage clears them. The owner-refresh helper preserves incomplete group summaries without mutating input.
- `backfill/engine_test.go`: existing actual CC-02 Execute/save/reload regression now also preserves parent rework and root cached summaries across two passes while the child's source is absent.
- `codex/rollout_test.go`: lean projection preserves distinct equal-looking requests; selection plus later attachment retains the initial source time. Missing source requested type/role/time retains the occurrence with null members. Opaque output/status variations cannot introduce occurrence outcomes or remove attempts.

## Assertion removal and transfer details

The two old NormalizeLedgerForRead event-copy tests are retired because the canonical ledger no longer accepts event transport. Chronological stable ordering remains in SortEvents tests; source identity, role and payload preservation are exercised through actual extraction and hydration, while input immutability and disk bytes are checked at those boundaries. The old aggregation event-preservation test retains its independent invocation-cost and empty-ledger defaults; event storage assertions are removed. Raw invocation role mapping, model/token metrics and immutability assertions remain in their existing aggregation test. Source-only rework rules now use explicit events, including the five-cycle critical boundary; metrics-only and unmapped-source cases reject invented ordering instead of expecting invocation-order fallback.

The old mixed-source committed-echo and hostile-fallback hydration tests are replaced by the new real mixed-source test and existing failure table. Nil service/hook, missing/closed/corrupt database, absent source, empty recovery, native payloads, role controls, repeat stability and no-write assertions retain concrete recipients. The three duplicate saved-ID/legacy/name-time persona matching variants are removed with saved-event reconciliation; all six actual native persona controls remain. Missing native persona stays unknown instead of borrowing a metrics-row role. Root/child availability recovery fixtures now retain supported partial source events and become complete after the missing source is restored.

Older migration step/invariant/cost tests use migration-only input aliases and MigrateInput, preserving their intermediate 4-to-5 cost and 5-to-6 ownerless deletion assertions. Intermediate event pruning is superseded by final schema-7 removal; old summary preservation remains asserted. Current delegation costs/outcomes are retired, while invocation partial-cost round trip remains in the owning adapter cost test. Runtime testdata schemas were upgraded and event-derived summaries were carried forward using actual Engine.Migrate on the baseline fixture bytes before dropping events. Temporary fixture exporters were removed; tests do not depend on ticket artifacts or /tmp outputs.

## Defects exposed while adapting tests

Tester retained failing contract oracles and reported fixes to the backend owner: empty nonnull caller_role/agent_type accepted by current validation; tolerant OpenCode extraction losing otherwise valid facts on malformed optional skill data; supported partial source prompt payload lost on early parse failure; and migration-input invalid target classification. The backend owner corrected production. Existing frontend/source fixture checks separately exposed missing transient native tool arguments; frontend owns that discovery and final wire assertions. These development corrections are not an independent review approval.

## Verification evidence and limits

Read contract, plan, technical/regression policy, tester persona, predecessor TESTING.md and REVIEW_SUMMARY.json. Inventory uses bounded declarations and direct assertion inspection; this is not an exhaustive suite count or runtime proof. No runtime suites executed during initial unstable production edits. All fixtures remain self-contained; synthetic source-shaped records do not establish real-source compatibility. Actual ParseRollout, SaveFactLedger, WorkspaceDetailHandler and real source hydration/WorkspaceEventsHandler generated the temporary wire fixtures handed to frontend; frontend reran export after native payload fixes.

Passed on the evolving implementation diff after backend build compiled:

```text
go -C backend test ./internal/domain ./internal/domain/migration ./internal/adapters ./internal/adapters/opencode ./internal/adapters/codex ./internal/adapters/backfill ./internal/adapters/live ./internal/adapters/capture -count=1
```

All eight packages passed: domain 0.079s, migration 0.045s, adapters 1.752s, OpenCode 0.601s, Codex 0.125s, backfill 5.450s, live 3.015s and capture 42.941s. Subsequent test-oracle cleanup reran domain 0.036s and Codex 0.086s. This run is working-diff evidence and predates final summary hardening; it is not a frozen-revision or race/make verdict.

After final summary hardening, passed affected domain/backfill regressions (0.008s/0.023s):

```text
go -C backend test ./internal/domain ./internal/adapters/backfill -run 'TestSourceSummaryRefreshRequiresCompleteChildGroup|TestBackfillCodexCompleteParentRefreshesWhileMissingChildFactsSurvive|TestPersistAnomalySummariesRefreshesNewSessionsAndAggregates' -count=1
```

The exact rich-field migration audit assertion subsequently passed (migration 0.004s). `git diff --check` passed. Frontend owner reports HTTP package passed after its current fixture and response-contract adaptation; backend owner retains CLI verification. Parent retains final frozen-revision integration/make/race, embedded assets/standalone verification, independent review and telemetry. No new silent skip or external source dependency was added.

## Final tester freeze

After backend production froze, complete affected domain/OpenCode/live packages passed (0.035s/0.243s/1.348s). The existing metric-only context-ceiling test still asserts detection from saved PeakContext without events. Persistence now additionally proves source-less normalization leaves the recorded owner snapshot unchanged, while an explicitly complete owner snapshot permits writing the metric-derived summary. This transfers the write-authority oracle without discarding fact-only health detection.

Focused current adapter validation/migration/source-proof/partial-anomaly checks and actual backfill CC-02 passed (0.030s/0.033s). Current test names clarify the approved behavior: TestOwnedFactsSurviveV5ToV7MigrationSaveReloadAndRejectRetiredRuntimeShapes, TestNativeNamespaceChildProofSurvivesLeanPersistenceAndRejectsAmbiguity and TestFullSourceAnomaliesSurviveLeanPersistenceAndPartialRefresh. The renamed cases plus native OpenCode extraction/failure-detail recipients passed again (adapters 0.030s/OpenCode 0.019s).

Native tool-input preservation now retains filePath/path variants and additional offset/limit arguments byte-for-byte; the former path-collapsing test was updated to protect the complete transient payload. Recorded OpenCode skill failures retain error status and exact output details in transient events, with no occurrence outcome field. Serialization fixtures now carry valid availability/coverage and invoke response validation before their payload round trip.

Handoff diff SHA256: `edda2bf56fba7fda24a24fcf82c820021e9c2d7894cbbfff3052c6382b0c5175`, defined as `git diff --binary 5e4eb34019370150786085d69c47987b328e4e79 -- backend testdata` at tester freeze; generated ignored asset trees and ticket documents are excluded. This identifies the final handed-off diff, not a claim that the earlier eight-package run tested every later correction. Latest affected reruns above cover changed boundaries; parent still owns the integrated gate. Tester is paused and relinquishes writable test ownership for review freeze.

## Review correction verification

Parent lifted the review freeze for the three confirmed findings on the reviewed candidate `86056b01ea7de69f2d4dec91fde1de9fd5d9c451d4de34260f1aa8a77d68ed7f`. Tester again owns all backend tests and fixtures except HTTP; the backend coder owns production only. Read the consolidated REVIEW/FINDINGS, all four sol reports and isolated source/cache probes. These are review corrections within iteration 03; no new scope or compatibility policy was introduced.

| Finding | Discovery credit | Permanent regression and correction verification |
| --- | --- | --- |
| BEH-01, severity 5 | Behavior reviewer, data-integrity reviewer and test-effectiveness reviewer; each gpt-6.1-sol/high, coordinator-acknowledged identity | Existing schema-v2 native fixture now asserts exact independent skill `.288000` and delegation `.301000` times instead of enclosing message `.238000`, through extraction, strict extraction, exact capture and save/reload. Missing native call times are null at each boundary. Existing v1 fixture asserts native part times plus SQL-null/zero controls. The earlier nonnull-only assertion was inadequate. Backend coder carries invocation time separately as transient json-excluded data before lean projection. |
| DATA-01, severity 5 | Data-integrity reviewer, gpt-6.1-sol/high, coordinator-acknowledged identity | TestOpenCodeDelegationChildProofUsesKnownNativeParentOutsideSelection exercises both native schema families, known wrong parent selected and outside selection, a referenced root, valid direct child and unavailable explicit ID. Strict extraction, exact capture, actual Store.Reconcile and save/reload agree. Source trace navigation metadata omits contradictory IDs while retaining raw returned output. The existing migration wrong-parent row now has only an explicit bad child ID; omission and unresolved audit remain asserted independently of opaque handles. |
| ARCH-01, severity 8 | Architecture reviewer, gpt-6.1-sol/high, coordinator-acknowledged identity | TestRegistryRetiredCachedSnapshotsRecoverFromUnchangedNativeSource replaces the rejection-only oracle: retired event/partial/task/path and rich-skill cache bodies expose valid binding/identity/checkpoint metadata with no current Snapshot; List does not write; another ticket proceeds; Register preserves opaque raw snapshot bytes; unchanged complete native source replaces the body and clears recovery state. Actual registered-only CLI tests cover unchanged available, missing and malformed native source, accepted current owners/root summaries/costs and same-record recovery after restoring the exact source. |

TestRegisteredGenerateRecoversRetiredCacheAndPreservesUnavailableCurrentOwners asserts accepted owner and root-summary JSON bytes, known facts, explicit unavailable child and costs survive failed source refresh, and the opaque registry body remains until successful extraction. Binding/request/native checkpoint remain unchanged. The existing observation finalization may advance root updated_at/completed_at; parent explicitly preserved that established behavior rather than adding an unrelated abort policy. Incomplete coverage is checked through the documented `--append` rendered summary, where failed owners cannot be presented as fresh. Failed explicit native-source generation retains its existing whole-ledger-byte preservation regression.

Final combined correction command passed on stable production:

```text
go -C backend test ./internal/adapters/opencode ./internal/adapters/capture ./internal/domain/migration ./cmd/agentic-lens -run 'TestSchemaV1ExtractsLeanSkillsAndDelegationsFromNativeParts|TestSchemaV2ExtractsLeanOccurrencesWithExplicitChildSession|TestOpenCodeDelegationChildProofUsesKnownNativeParentOutsideSelection|TestRegistryRetiredCachedSnapshotsRecoverFromUnchangedNativeSource|TestRegisteredGenerateRecoversRetiredCacheAndPreservesUnavailableCurrentOwners|TestV7MigrationPreservesOwnersMultiplicityProvenChildrenAndSummaries|TestGenerateMalformedSkillArgumentsPreservePreviouslyOwnedSnapshot|TestGenerateExactSourceWithoutPrivateRegistrationReplacesPortableOwner|TestCaptureGenerationRejectsExplicitSelectionOfAnotherTicketOwner|TestReconcileReplacesCompleteRestoredSnapshotAndPreservesMissingSource|TestOpenCodeRegistrationNeverFallsBackToMatchingTitle' -count=1
```

Passed OpenCode 1.101s, capture 1.133s, migration 0.008s and CLI 3.064s. Earlier setup runs corrected test-fixture registry path and selected the existing rendered coverage output; no production oracle was weakened. Backend owner independently reports unchanged original review probes now passing: architecture recovery 0.067s, data extraction/publication 0.019s and exact/missing-time effectiveness probes 0.013s. Those are separately attributed owner results, not tester executions or a paired re-review approval.

Correction freeze content SHA256: `c44382d238f2f07b98e650caed71c2a449af5b032fd97a7319a3f8c7646281f2`, 293 files. Inputs are sorted repository-relative paths for every backend Go file, backend/go.mod, backend/go.sum and root testdata JSON/JSONL files, including untracked new tests; hash input is UTF-8 path, NUL, file bytes, NUL for each file. Generated embedded assets and ticket documents are excluded. `git diff --check` passed. Tester stops writing for the corrected review freeze; integrated make/race, frontend/embedded assets, paired independent re-review and consolidated approval remain parent-owned.

## Final schema-6 correction and freeze

User steering challenged the extra schema bump because main releases schema 5 and this ticket's rich schema 6 was never released. Parent resumed execution from `7cec253d6e62f0440df1f1c906263f645d075c9d` with final lean schema 6 and direct 5-to-6 migration; no interim compatibility or schema-7 downgrade was authorized. Tester owned backend tests/fixtures except HTTP and root runtime fixtures; backend coder owned production/docs, frontend retained HTTP/web, and parent retained live ledger regeneration/integration.

Assertion transfers and retained guarantees:

- Renamed lean_v7_test.go to lean_v6_test.go and its regression to TestV5ToV6LeanMigrationPreservesOwnersMultiplicityProvenChildrenAndSummaries. Its source is genuine supported schema-5 rich input, with identical independent ownership, multiplicity, nullability, child-proof, cost/token, summary and removal-count oracles. It also requires exactly the single schema_v6_lean_occurrences completion marker. No 6-to-7 step is exercised or simulated.
- TestV6MigrationDropsOnlyOwnerlessFactsAndPreservesNativeMultiplicity now calls public Engine.Migrate to final6, rather than inspecting a rich migration-only current6 pseudo-output. Ownerless drop counts, native overlap handling, cross-owner equal values, unavailable child ID, row counts, null unknowns and repeat stability remain. Final occurrence outcome/delegation-cost assertions were already retired by the lean contract; historical observed-zero/unavailable task-cost conversions remain covered at the genuine 4-to-5 intermediate boundary.
- Legacy invocation ownership fixtures iterate only supported legacy schemas 1 through 5 and verify final6 through Engine.Migrate. MigrateInput is restricted to genuine historical intermediate targets through 5. The missing-step regression uses a custom historical registry gap at 2-to-3, preserving the actual gap oracle without inventing a future released version.
- TestFinalV6MigrationBoundaryRejectsRichCurrentAndUnsupportedVersions accepts valid lean6, rejects rich current6 fields instead of converting them, rejects source/target7 as unsupported and rejects rich current6 pseudo-output from MigrateInput. Adapter save/reload tests independently reject rich6 and unsupported7; the repository rejection test preserves current-rich6 bytes without falsely requiring a released older-schema migration.
- Added actual CLI TestCLIMigrationUsesDirectV5ToFinalV6AndRejectsUnreleasedSchemas: released5 produces final6 with owned rows, ownerless drop audit, original backup bytes and unchanged unrelated costs/tokens; rich6 and schema7 remain unchanged and report an error rather than up-to-date. The existing summary-idempotence migration fixture now starts at released5, retaining its byte-identical second summary assertion.
- Current lean runtime fixtures and root testdata JSON discriminator/migration metadata changed from the ticket's schema7 candidate to final6. Fact values and recorded event-derived summaries were retained. No live ledger was rewritten by tester. BEH-01 exact/null native times, DATA-01 contradictory-parent proof and ARCH-01 metadata/cache/registered-only recovery tests and their production paths are unchanged by this correction; their established evidence and shared reviewer credit remain applicable.

Passed affected suites on stable direct-migration production:

```text
go -C backend test ./internal/domain/migration ./internal/adapters ./internal/adapters/backfill -count=1
```

Migration 0.025s, adapters 0.699s, backfill 2.823s.

```text
go -C backend test ./cmd/agentic-lens -run 'TestCLIMigrationUsesDirectV5ToFinalV6AndRejectsUnreleasedSchemas|TestCLIMigrateCommand|TestCLIBackfillMigrateSchemaFlag|TestCLIMigrateReportsDroppedUnscopedToolRows' -count=1
```

CLI passed 0.946s, including the adversarial migration cases matched by TestCLIMigrateCommand. Earlier fixture compilation and idempotence failures were corrected in tests without changing the required oracle. `git diff --check` passed. No duplicate full make/race was run; parent owns the integrated gate and focused independent schema correction rechecks.

Final source/test freeze hash: `df8a2c16a3998bf315876327bb756d2f5abdb06bb509883d09808216d0de655f`, 293 files, using the same sorted path/NUL/bytes/NUL definition above and including new/untracked tests. Tester relinquishes writable paths and is frozen for parent-directed review.
