# AL-55 iteration 01 adversarial testing

Test owner: tester. Baseline integrated commit 858a7eb2885d674b11c672abf7c1621e687bde9b (origin/main aa88cb78c93daa4a2dd6d7ba9b1e512603c48a0c). Tests exercise the shared AL-55 implementation diff, not an isolated committed implementation head. Parent owns final integrated make and delivery. All fixtures are temporary files, SQLite databases or existing typed adapter callbacks; no live ticket workspace or developer source history is used as an oracle.

## Observed regressions before the corresponding fixes

- Actual `LoadExistingFactLedger` omitted/null collection reads returned nil while the DTO-to-domain workspace path already initialized collections. `TestLedgerFileReadsNormalizeOptionalSkillTaskCollections` failed omitted/null variants; actual reader subsequently fixed. Populated/repeated and malformed-value variants already behaved correctly.
- Strict OpenCode refresh excluded malformed unclassified part text through the pre-parser SQL LIKE filter. `TestBackfillEngine_SkillTaskRecoverySnapshots/late_malformed_part` observed a completed refresh with no diagnostic and replacement collections despite the malformed record. Strict query subsequently changed to validate all relevant source parts; tolerant normal ingestion remains unchanged.
- Existing tests exposed expectations deliberately superseded by accepted recompute/completeness policy. Those assertions were reconciled as described below; they are not presented as regressions in accepted behavior.

## New boundary coverage

- `TestLedgerFileReadsNormalizeOptionalSkillTaskCollections`: real LoadExistingFactLedger and LoadTicketWorkspace reads for omitted, null, empty, repeated populated and malformed collections. Internal arrays initialized, original rows/multiplicity preserved, malformed data rejected, original optional wire bytes unchanged.
- `TestBackfillEngine_ExplicitRecomputeDBDoesNotSkipCompletedLedger`: completed metadata present/absent crossed with available/unavailable source. Assert actual cumulative invocation refresh without Force, preserved eligible baseline on optional failure, explicit diagnostic and ordinary subsequent skip without rewriting bytes.
- `TestBackfillEngine_SkillTaskRecoverySnapshots`: real SQLite complete, confirmed-empty, missing-root, missing-child, late-malformed, unknown-owner, unrooted-cycle, incomplete dry-run and temporary-write-failure paths. Preserve whole old arrays and incomplete trees, update independent eligible tree, retain other harness, assert ticket/aggregate diagnostics and mixed age, grow source once then repeat without duplication. Repeated legitimate skill/task rows remain repeated. No eligible owner is not evidence of empty history.
- `TestBackfillEngine_GlobalRecoveryRequiresEveryKnownTree`: unavailable Codex sibling, unselected OpenCode obligation, missing child, late parse and grouping failures. Keep global source-backfill and unrelated pending steps; retain nonselected and failed owner facts. Existing pending baseline remains runtime ineligible; one local success never certifies global recovery.
- `TestBackfillEngine_SchemaV2RecoversKnownGrandchildren`: real schema-family v2 root/child/grandchild extraction and late malformed descendant. Full valid tree refreshes cumulative facts; failure retains the entire old tree/collections.
- `TestSchemaV1StrictRecoveryRejectsSkippedRecords`: strict malformed message/unclassified part rejection, tolerant standard ingestion compatibility, and valid whitespace-formatted producer data recovered by strict refresh.
- `TestBackfillEngine_InvalidCandidatePreservesFiles`: invalid extracted tool counts reach final DTO validation in execution and dry-run; original ledger/summary bytes survive and no temporary write begins.
- `TestCLIBackfillCommand/optional_failed_refresh_reports_retention_in_ticket_and_command_summary`: actual built subprocess reports previous-snapshot/mixed-age details in ticket row and aggregate incomplete/retained counts, preserving runtime-eligible baseline.
- `TestRecoverOpenCodeRetainsUnavailableOrphanSiblingTree`: reviewer-requested defensive source-recovery boundary with real SQLite. Stored orphan siblings A/B share absent parent R; source contains R/A only; B and A remain while independent tree refreshes and recovery stays incomplete. This is a direct helper test, not a canonical-file demonstrated loss: current file validation independently rejects missing parent references. The test first ran after the coder fix, so no pre-fix execution claim is made.

## Reconciled existing coverage

- Completed-ledger explicit recompute now preserves canonical eligibility with a command diagnostic on optional source failure; it no longer asserts a newly mandatory migration marker. Its metadata-present case remains, alongside canonical metadata-absent cases.
- Migration-to-ready fixture now carries the exact stored source owner before shape migration. This retains source-backed completion coverage without claiming heuristic discovery from an empty history proves full recovery.
- Database-only zero-steering test now separately protects unchanged Git SDLC facts and source-confirmed zero invocation user iterations; explicit database refresh does not authorize recomputing Git facts.
- Multi-machine second explicit refresh processes both tickets: unavailable already-canonical A stays eligible with incomplete command diagnostics, while B genuinely completes. Routine skip is asserted separately in canonical recompute coverage.
- Codex root-tree test still asserts one discovery, identities, child facts, unrelated facts and no duplication. Any unavailable sibling now prevents positive all-owner DBFound even while DBUpdated records the independently successful tree.
- Explicit heuristic fallback still proves opt-in enrichment, but additionally preserves the unavailable exact owner and refuses to certify complete history. No title match may substitute for missing historical identity.

## Verification and limits

The focused race suite covers adapters, backfill, OpenCode and CLI. No test silently skips. Existing migration-runner backup/idempotence tests and AL-43 warning/content suites remain in their owning packages; no redundant backup interface or frontend tests were added. Parent must execute full make on the actual integrated delivery diff and record that evidence separately. No production history was migrated or claimed exhaustively recovered; source completeness is bounded by known stored owners and discovered records.

Commands executed include:

- `cd backend && go test ./internal/adapters -run TestLedgerFileReadsNormalizeOptionalSkillTaskCollections -count=1`: initially failed omitted/null variants, later passed in package/race verification after actual reader fix.
- `cd backend && go test ./internal/adapters/backfill -run 'TestBackfillEngine_(SkillTaskRecoverySnapshots|ExplicitRecomputeDBDoesNotSkipCompletedLedger)' -count=1`: initially failed malformed-part completeness, fixed strict parsing subsequently passed.
- `cd backend && go test ./internal/adapters/opencode -run TestSchemaV1StrictRecoveryRejectsSkippedRecords -count=1`: passed.
- `cd backend && go test ./internal/adapters/backfill -run TestBackfillEngine_InvalidCandidatePreservesFiles -count=1`: passed.
- `cd backend && go test ./internal/adapters/backfill -run TestRecoverOpenCodeRetainsUnavailableOrphanSiblingTree -count=1`: passed after defensive fix.
- `cd backend && go test ./cmd/agentic-lens -run TestCLIBackfillCommand -count=1`: passed after fixture supplied required title/user metadata; those fixture corrections are not application regressions.
- `cd backend && go test -race ./internal/adapters ./internal/adapters/backfill ./internal/adapters/opencode ./cmd/agentic-lens -count=1`: passed before final CLI/final-validation/orphan additions; final expanded run recorded below.

Final expanded focused race run passed: adapters 1.686s, backfill 6.039s, OpenCode 2.180s, CLI 4.962s. After the coder strengthened the same defensive replacement closure to include newly discovered intermediary parents, the orphan-helper test was extended with direct-parent and intermediary-parent variants. The affected backfill package race run passed again (5.605s). This latter coverage is still explicitly a defensive helper boundary, not an accepted malformed canonical-file claim. `git diff --check` passed. No tester production edits, commits, telemetry generation or PR actions were performed.
