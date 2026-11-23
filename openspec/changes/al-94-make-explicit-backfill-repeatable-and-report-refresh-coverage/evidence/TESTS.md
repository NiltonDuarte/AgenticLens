# AL-94 verification

The focused checks cover the worktree based on `db4477fa` with backend content fingerprint `2599892658fedd668c2b5bf29ce308b8b80e71bf28470a2eb839e4a0ff19519e`. The fingerprint includes changed tracked and untracked backend paths and their SHA-256 content hashes.

From `backend/`, the tester ran:

```sh
go test ./internal/adapters/backfill ./cmd/agentic-lens -run '^(TestExplicitBackfillPartialRepeatAndIndependentSummary|TestCLIBackfill.*|TestBackfillFreshGitFailurePreservesPreparedMigrationBytes|TestBackfillEngine_ExplicitRecomputeDBDoesNotSkipCompletedLedger|TestBackfillEngine_MigrationThenSourceBackfillMakesLedgerReady|TestBackfillEngine_SchemaV2RecoversKnownGrandchildren|TestFIX19BackfillRetainsAmbiguousAndUnavailableOwnersWhileRefreshingHealthy)$' -count=1
```

The backfill package passed in 27.656 seconds and the CLI package passed in 34.504 seconds. The cases cover native exact-source recovery, a retained child and unrelated harness, verified-empty collections, byte and modification-time stability, summary-only repair, store preflight, combined Git recomputation, and a batch budget limit followed by an exact-ticket retry.

`make web-assets` built the assets required by the CLI tests. Strict OpenSpec validation and `git diff --check` passed. The integrated `make` gate and independent review are pending.

The tester reproduced a false incomplete refresh in a disposable native OpenCode v1 CLI fixture. The configured two-owner budget covered both recorded owners, but a redundant descendant lookup reported exhaustion after both were refreshed. The correction visits recorded roots first and avoids rediscovering subtrees already enumerated in full.

On corrected backend fingerprint `c76f6a0e3ed5598cba3879eda0fb69c0332ccbbffb0a901335f3a7e4d06d28f3`, `go test ./internal/adapters/backfill -run '^TestExplicitBackfillOpenCodeTreeReadBudget$' -count=1` passed in 1.228 seconds. Both lexical parent/child orders verify complete coverage at the exact limit. A third independent recorded owner beyond the same limit remains retained and is identified with the budget reason. The coder also reran the focused backfill tests for skill/task snapshots, schema v2 grandchildren, required global recovery, unavailable orphan siblings and partial-repeat stability. They passed in 19.817 seconds.

`go test ./internal/adapters/backfill ./cmd/agentic-lens -run '^(TestExplicitBackfillOpenCodeTreeReadBudget|TestCLIBackfillOpenCodeBatchBudgetReportsRetry|TestRecoverOpenCodeRetainsUnavailableOrphanSiblingTree|TestBackfillEngine_SchemaV2RecoversKnownGrandchildren)$' -count=1` passed on the same fingerprint (25.701 seconds for backfill and 26.336 seconds for the CLI).

## Review corrections

The tester verified backend fingerprint `d59398aa4cf916f8aa9196cc226bc20305a3a7e6010dc710d8c43a392ae075a4` (16 changed paths, baseline `db4477fa`):

```sh
go test ./internal/adapters/backfill ./cmd/agentic-lens -run '^(TestExplicitBackfill.*|TestBackfillLedgerStagingPreservesConfiguredSummary|TestBackfillEngine_SkillTaskRecoverySnapshots|TestBackfillEngine_ExplicitRecomputeDBDoesNotSkipCompletedLedger|TestBackfillEngine_MigrationThenSourceBackfillMakesLedgerReady|TestBackfillEngine_SchemaV2RecoversKnownGrandchildren|TestRecoverOpenCodeRetainsUnavailableOrphanSiblingTree|TestFIX19BackfillRetainsAmbiguousAndUnavailableOwnersWhileRefreshingHealthy|TestCLIBackfill.*|TestBackfillFreshGitFailurePreservesPreparedMigrationBytes)$' -count=1
```

Backfill passed in 25.337 seconds and the CLI passed in 19.991 seconds. These cases cover the six product findings and the positive CLI summary-write assertion described in [independent review](REVIEW_SUMMARY.json). The coder also reran the shared publication and summary tests (24.123 seconds) and reproduced the four data-integrity probes through an isolated Go overlay (2.735 seconds). Those probes cover rejected summary validation, matching incomplete counts, the staging filename collision and late summary write rollback.

The tester verified the Git/source preflight and interrupted-observation corrections on backend fingerprint `fa8ed58750c884760a28386671278f0e6caed25a74ec5d8df110774c5a61bca4` (17 changed paths, baseline `db4477fa`):

```sh
go test ./internal/adapters/backfill ./cmd/agentic-lens -run '^(TestBackfillCancellation.*|TestBackfillOpenCodeInheritsCallerDeadlineDuringNativeRead|TestCLIBackfillSelectedStoreAndGitPreflight|TestBackfillFreshGitFailurePreservesPreparedMigrationBytes|TestCLIBackfillCommand|TestCLIBackfillOpenCodeBatchBudgetReportsRetry|TestCLIBackfillRepairsSummaryWithoutRecomputingFacts)$' -count=1
```

Backfill passed in 4.855 seconds and the CLI passed in 7.544 seconds. The coder also ran `go test ./internal/adapters/backfill ./cmd/agentic-lens -run 'TestBackfillCancellation|TestCLIBackfill|TestBackfillFreshGitFailure' -count=1`, which passed, and all four independent Sol re-check probes passed through isolated overlays.

The tester verified the candidate-preparation failure correction on backend fingerprint `bacfd5d065e07a9ef552a9871cadba78ba695bb7dac9c65f128935c74c985388`, with the same 17 changed paths and baseline:

```sh
go test ./internal/adapters/backfill ./cmd/agentic-lens -run '^(TestBackfillReportsVerifiedReadsWhenAnomalyPolicyPreventsComparison|TestBackfillCancellation.*|TestBackfillOpenCodeInheritsCallerDeadlineDuringNativeRead|TestCLIBackfillSelectedStoreAndGitPreflight|TestCLIBackfillOpenCodeBatchBudgetReportsRetry|TestCLIBackfillRepairsSummaryWithoutRecomputingFacts)$' -count=1
```

Backfill passed in 4.625 seconds and the CLI passed in 4.955 seconds. The native regression verifies that a late anomaly-policy error preserves successful owner and collection observations, reports the unavailable comparison, and writes neither file. It also exercises cancellation after Codex consumes real source bytes. The coder's independent Go-overlay replay of the prior architecture probe passed both cases.

## Integrated verification

Candidate `da90e8e0` includes `origin/main` at `db4477fa`. The integration merge reported that the branch was already up to date. The candidate matches the final reviewed backend fingerprint.

- `GORACE=atexit_sleep_ms=0 GOFLAGS='-p=3 -timeout=30m' make -j1` passed. The CLI race suite took 391.329 seconds and the backfill adapter race suite took 24.547 seconds. All 231 frontend unit tests and 819 frontend integration tests passed. The gate also covered vet, lint, installation, distribution and workflow tests, and produced manager and engine binaries for four platforms.
- `python3 scripts/tests/openspec_workflow_test.py` passed all three tests. Strict validation of the AL-94 change and `git diff --check` passed.
