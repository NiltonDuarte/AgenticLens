# AL-62 tester verification

Status: tester-owned targeted regressions pass; required integrated verification and review pending.

Tester owns only `backend/cmd/agentic-lens/main_codex_test.go`, `backend/internal/domain/skills_tasks_test.go` and this report following explicit transfer from backend owner. Other source/test paths remain with their owners.

## Regression selection and recipients

- Extended `TestGenerateCodexPreservesChildIdentityAndExcludesInheritedUsage`: source-format root read/spawn calls replayed in a child with identical IDs must remain only root-owned; child-owned calls remain separate. Saved ledger has canonical nested collections and ticket projections count every occurrence once.
- Added `TestGenerateCodexPreservesUnavailableOwnersAndReplacesLocalFactSnapshot`: actual CLI generate with legitimately registered and committed remote Codex owner in a separate worktree and unresolved legacy history; save/reload after initial, growing, unchanged and authoritative-empty local snapshots. Assert remote facts and child references survive, local facts replace rather than accumulate, unknown skill execution outcomes remain unavailable, namespace references never become invented UUIDs.
- Added `TestOwnedFactProjectionsPreserveOccurrenceAndLegacyMultiplicity`: same call IDs across sessions, explicit transitional projection echo, repeated equal unresolved legacy values, unavailable child reference and read nonmutation.
- Existing backfill `TestBackfillEngine_SkillTaskRecoverySnapshots` is a coherent recipient for partial/missing roots, malformed descendants, empty snapshots and repeated recovery; backend owner notified that its former top-level replacement oracle must change to preserve unresolved legacy history.
- Frontend owner notified to extend existing events projection and comparison recovery tests for granular ownership and unknown outcome denominators.

Fixtures are synthetic source-format envelopes and self-contained temporary projects. No live ticket workspace is read by tests.

## Executed checks

- `cd backend && go test ./cmd/agentic-lens -run 'TestGenerateCodex(PreservesUnavailableOwners|PreservesChildIdentity)' -count=1`: PASS on current in-progress diff, 2026-11-03. Actual generate/save/reload, committed remote provenance, canonical owned facts, inherited replay exclusion, growing/repeated/empty authoritative snapshots.
- `cd backend && go test ./internal/domain -run 'Test(OwnedFactProjections|SkillAndTaskFact)' -count=1`: PASS on current in-progress diff, 2026-11-03. Full revision not yet frozen; this proves the owned projection contracts only.

## Pending and limits

- Refuted fixture finding: an initial growing-source case appended calls without advancing the source invocation completion checkpoint. `rejectOlderExplicitOwners`/`filterCapturePrecedence` in `capture.go` and `CaptureOwnerCheckpoint` in `domain/capture.go` correctly preserve differing equal/unknown snapshots. Existing `TestCaptureGenerationPreservesDifferingSnapshotsWithoutProvenSourceOrder` documents that accepted AL-76 boundary. The fixture now advances source token-usage completion evidence for replacement and keeps the unchanged stage identical; all stages pass. No chronology guard was weakened.
- Initial arbitrary baseline seed was also corrected to reuse actual cross-worktree committed-history provenance helpers after authorization correctly rejected it. That rejection was a fixture setup error, not a product defect.
- Missing roots/partial malformed source save-reload in generate and backfill, actual API persistence with unavailable hydration, schema migration, capture reconciliation and embedded standalone assets remain part of integrated verification; this report does not certify those paths.
- Potential extraction false-positive hypothesis supplied to backend owner (not confirmed): failed shell result containing a quoted successful-exit phrase must not establish completed skill loading; verification pending.
- Required race/full make gate and independent focused review remain pending.

## Review correction pass

The independent test-effectiveness reviewer (`reviewer-tests`, gpt-6.1-sol/high) identified incomplete granular persistence oracles in the existing backfill, capture, HTTP and migration tests. These are review-supplied corrections, not new tester discovery. The backend owner also implemented the review-confirmed extractor cache coverage version fix; this tester supplied its persistence/precedence regressions. Exact model identity here is inherited dispatch identity, not independent source-session metadata.

Writable scope for this pass: `backfill/engine_test.go`, `capture/reconcile_test.go`, `http/server_test.go`, `cmd/agentic-lens/main_codex_test.go`, `domain/migration/engine_test.go` and this report. Existing AL-76 source-order and writer-lock gates and AL-60 shared capture ownership were preserved.

- Backfill recovery matrix now starts with distinct populated skill/task facts on refreshed roots, unavailable roots/children and the other harness. Save/reload asserts exact preserved provenance/outcome/error/duration/child references on unaffected owners; refreshed roots replace their own facts, and authoritative empty recovery removes prior populated facts. Growing and repeated snapshots still preserve ambiguous legacy history and other-source facts.
- Capture resume/truncation/missing-source and partial-parse recipients now exercise actual synthetic native `exec_command` skill reads and native spawn calls, then assert source identifiers, unknown execution outcome, explicit unavailable child reference and cost evidence after reconciliation and persisted registry reload. Their existing failure cases compare populated snapshots.
- `TestReconcileUpgradesAuxiliaryCoverageOfUnchangedCachedSource` simulates an old successful cache whose fingerprint and checkpoint have not changed. Available source must repopulate canonical auxiliary facts and persist both extraction coverage revisions; unavailable source must retain its old snapshot/revision without claiming current coverage.
- Actual workspace-detail HTTP responses now retain rich nested facts while optional event hydration is unavailable, fails, is empty or partial. The existing filesystem snapshot oracle verifies GET leaves ledger bytes and workspace contents unchanged.
- Version 5-to-6 migration now checks complete skill/task field equality, including source event/message/call identity, original timestamp/positions, evidence/outcome/error, child reference, duration and observed cost evidence.
- Actual CLI generate now aborts on an otherwise populated but partially malformed source. The test requires the specific parse-preservation diagnostic, exact accepted ledger bytes and previously accepted granular facts after reload; a generic unrelated command failure cannot satisfy it.
- `TestGenerateUpgradesAuxiliaryCoverageWithoutChangingEqualCheckpointMetrics` covers registered and explicit generation. An old zero-coverage canonical owner gains missing facts from unchanged raw source at the same source checkpoint. A differing accepted metric fact at that checkpoint prevents the upgrade and preserves the entire prior owner. No timestamp or writer-lock requirement was weakened.

Executed focused checks on the current correction diff (2026-11-03):

- `go test ./internal/adapters/backfill ./internal/adapters/capture ./internal/adapters/http ./internal/domain/migration -run 'Test(BackfillEngine_SkillTaskRecoverySnapshots|ReconcilePreservesAcceptedSnapshotsThroughResumeTruncationAndMissingSource|ReconcilePreservesFactsWhenSourceParsingIsPartial|ReconcileUpgradesAuxiliaryCoverage|RecordedV4FactsSurviveDetailEventsAndOptionalHydration|V6MigrationPreservesAmbiguousHistoryAndExplicitOwners)' -count=1`: PASS in all four packages.
- `go test ./cmd/agentic-lens -run 'Test(GenerateUpgradesAuxiliaryCoverage|GenerateCodexPreservesUnavailableOwners|GenerateCodexPreservesChildIdentity)' -count=1`: PASS, including both cache upgrade modes and conflicting metric guards.
- After updating positive direct-shell fixtures to the native `exec_command` source envelope, `go test ./internal/adapters/capture -run 'TestReconcile(PreservesAcceptedSnapshots|PreservesFactsWhenSourceParsingIsPartial|UpgradesAuxiliaryCoverage)' -count=1`: PASS.
- `git diff --check`: PASS.

The wrapper false-positive absence cases are confirmed reviewer findings and are implemented/tested by the backend owner in `codex/rollout_test.go`; this tester does not claim independent discovery or verification of those tests. Full `make`, race checks, integrated predecessor validation, standalone embedded assets and renewed focused review remain parent-owned and pending. The earlier Pending and limits section is historical to the initial implementation pass; the concrete granular persistence gaps listed there are closed by the correction tests above.
