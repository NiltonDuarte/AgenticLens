# AL-63 data integrity evidence

Run `AL-63/02/data_sol`, discovery, 2026-11-08. Requested profile: inherited gpt-6.1-sol/high; runtime model identity, usage and cost were not independently exposed. Reviewed `f4152225af06e2594477574c0e9f175a81218053`; `git diff --name-only f4152225 -- backend web` returned no paths. Initial checkout contained the parent's ticket restoration and untracked iteration02 artifacts. Product files were read-only; this report is the only assigned repository write. No sibling reports were read.

**Scoped outcome: APPROVED for discovery evidence.** The supported clean-family recovery paths preserve unavailable history in the exercised cases. This does not certify native OpenCode recovery, source parity, every historical deployment, or whole-repository correctness. The matrix must retain the native evidence gaps and known mixed-family limitation below. No implementation is proposed as part of AL-63.

## Matrix-ready conclusions

Here, “verified” means current production code executed with synthetic SQLite/rollout fixtures, not native harness activity. Native Codex checks belong to the parent. OpenCode v1 means `session/message/part`; v2 means `session_v2/session_message`. No native OpenCode version was measured in this pass.

| Capability | Codex | OpenCode v1 | OpenCode v2 |
| --- | --- | --- | --- |
| Exact stored-owner recovery, source and parent identity | Supported; fixture verified | Supported; fixture verified | Supported; fixture verified |
| Missing same-harness owner and unrelated-harness preservation | Supported; fixture verified | Supported; fixture verified | Supported; fixture verified |
| Missing child retains child facts while complete parent refreshes | Supported; real parser on synthetic rollout verified | Supported; disposable CLI verified | Supported; disposable CLI verified |
| Malformed source retains previous owner skill/task snapshot | Supported; warning-bearing rollout fixture verified | Supported; disposable CLI verified | Supported; disposable CLI verified |
| Missing/malformed records disclose incomplete extraction | Supported; fixture verified | Supported; fixture and CLI verified | Supported; fixture and CLI verified |
| Explicit mixed-age disclosure | Per-owner diagnostics and retained-summary behavior verified; no OpenCode-style collections counter asserted | CLI warns previous snapshots retained and trees may be newer | Same CLI warning verified |
| Growing snapshot replaces cumulative totals; unchanged repeat | Supported; fixed-cutoff fixture verified | Supported; fixed-cutoff fixture verified | Supported; fixed-cutoff fixture verified |
| Heuristic recovery | No title/workspace fallback in Codex recovery | Explicit opt-in only; fixture verified | Same opt-in gate declared; enabled v2 fallback not separately executed |
| Native source-backed backfill this pass | Not executed by this reviewer | Not executed | Not executed |

`completed` migration status and exit 0 do not establish full recovery: the disposable missing-owner CLI cases reported `Completed: 1` together with `Incomplete: 1`. Inspect refresh diagnostics and per-ticket errors. A valid stored ledger remains runtime-readable while an optional refresh fails; pending mandatory migration obligations remain pending in the global-recovery tests.

## Production path and counterchecks

`backfill/engine.go:365` partitions owners by source; `:497` recombines all partitions. `recovery.go:42` recovers Codex exact roots and remaining stored owners, rejects changed parents at `:71`, and filters events to accepted owners. `:115` rejects warning-bearing parsed owners, retains failed reads and groups complete transcripts before replacement. `recovery.go:229` discovers exact OpenCode descendants, captures each owner transactionally, checks returned identity and rejects changed parent links at `:270`. `opencode/capture_snapshot.go:29` uses strict extraction within a SQLite read transaction; `skills_tasks.go:86` rejects malformed skill/task arguments in strict mode. `engine.go:646` replaces successful owner snapshots by ID and retains unavailable owners; `domain/events.go:659` preserves group summaries whose child coverage is incomplete.

Thus unavailable reads cannot authorize empty replacement; successful complete-empty snapshots can clear stale occurrences. Genuine equal-valued skill/task occurrences remain separate by source occurrence identity. `engine.go:491` distinguishes refresh incompleteness from migration completion. Default heuristic recovery is disabled (`options.go`, CLI `main.go:903`); `recovery.go:301` permits explicit fallback only when exact recovery updated nothing, rejects overlap with known IDs and keeps recovery incomplete. These guards and the passing fallback controls refute silent title-based owner reassignment.

**DATA-SOL-01, severity 3/10, known conditional capability limit; owner REV-7.** Trigger: a v1-only exact owner in a database that also contains complete empty v2 tables. `recovery.go:170` selects `session_v2` by table existence before exact capture, so backfill reports the available v1 tree unavailable and preserves its stale snapshot. A disposable CLI probe reproduced this twice; removing only the empty v2 tables allowed the same owner to refresh to model `fresh`. Consequence is missed refresh, not lost history. Refuting/limiting evidence: clean v1/v2 controls passed; generic `schema_family_detection.go` intentionally rejects mixed layouts; exact capture has separate membership selection at `capture_snapshot.go:102`. The existing [REV-7 specification](../../../plan/to-do/REV-7%20review%20source%20read%20guarantees.md) explicitly owns this policy divergence and says mixed recovery was not promised. No native mixed-family store was established. This is not an AL-63 required-contract defect or authorization to broaden generic ingestion.

## Executed verification

All commands below used unchanged source at the reviewed revision, Go `go1.26.0 linux/amd64`, `-count=1`, no race detector. Existing test cases use disposable files/SQLite stores; Codex discovery/parser injection cases do not establish native source observation.

```sh
cd backend
go test ./internal/adapters/backfill -run 'TestBackfillEngine_(SkillTaskRecoverySnapshots|GlobalRecoveryRequiresEveryKnownTree|SchemaV2RecoversKnownGrandchildren|StrictSessionTrustRejectsUntrustedTitleFallback|ExplicitFallbackResolvesWhenAllowed|CodexPreservesChildParentIDAndUser|CodexRefreshesRootTreeOnce|CodexParseFailurePreservesTree|WarningBearingRolloutsPreserveTree|MultiMachineLifecycle|SourceOpenCodeSkipsCodexSessions|SourceCodexSkipsOpenCodeSessions)|TestBackfillCodexCompleteParentRefreshesWhileMissingChildFactsSurvive|TestBackfillActiveDurationReplacesSnapshotsAcrossSourceFamilies|TestRecoverOpenCodeRetainsUnavailableOrphanSiblingTree|TestReplaceRecoveredOwnersPreservesUnavailableChildWithoutDuplication' -count=1 -v
go test ./internal/adapters/opencode -run 'TestSchemaV1StrictRecoveryRejectsSkippedRecords|TestSchemaV[12]MalformedOrIncompleteSkillTaskInputsPreserveOtherFacts|TestSchemaV2SessionMalformedJSONAndUnknownMessageVariantFailWithContext|TestOpenCodeDelegationChildProofUsesKnownNativeParentOutsideSelection' -count=1 -v
go test ./cmd/agentic-lens -run 'TestGenerateExactSourceWithoutPrivateRegistrationReplacesPortableOwner|TestGenerateCopiedLedgerPreservesUnavailableOwnedFactsWithoutPinnedDestination|TestGenerateMalformedSkillArgumentsPreservePreviouslyOwnedSnapshot|TestGenerateCompleteParentRetainsGroupSummaryForUnavailableChild|TestCLIGenerate_SessionRefreshMatchesFreshExtraction' -count=1 -v
```

Results: PASS, respectively `3.142s`, `1.045s`, `6.037s`. The backfill duration test executes all three families with root, child, unavailable same-harness root and unrelated harness, growing cutoffs `10,20,20,30,30`, unchanged repeats, full-ledger totals and unchanged parent identity. The v1 snapshot test checks repeated equal-valued occurrences, missing root/child, late malformed part, complete-empty clearing, write failure and dry-run preservation. The v2 test reaches a malformed grandchild while complete parent owners refresh. Global recovery tests retain mandatory source-backfill obligations after unavailable siblings, missing children, parse warnings/errors or grouping failure. CLI portable-history tests verify exact local-source replacement without private registration and copied-ledger preservation.

Additional disposable CLI evidence used `go build -o /tmp/al63-data-sol.G8FqPw/agentic-lens ./cmd/agentic-lens`, then `python3 /tmp/al63-data-sol.G8FqPw/probe.py` and a follow-on Python SQLite mutation probe. The binary reports `v0.0.0-dev (commit: unknown, built: unknown)` because build metadata was not injected; revision is established by the clean source comparison above. Successful invocation shape:

```sh
/tmp/al63-data-sol.G8FqPw/agentic-lens backfill --source opencode --workspaces /tmp/al63-data-sol.G8FqPw/v2-final/workspaces --ticket-id AL-63-PROBE --recompute-db --db-path /tmp/al63-data-sol.G8FqPw/v2-final/source.db
```

The v1 equivalent substitutes `v1-final`. Both have exact IDs `stored-root`, unavailable `stored-child` (parent `stored-root`), unavailable `remote-root` and Codex `other-source`. Fresh root becomes model `fresh`, clears stale skill, retains all four IDs and exact child/other-source records on both runs, and emits `Incomplete: 1` plus mixed-age warnings. Changing the exact source message `m` JSON to `{` preserves the entire prior session array, including `keep-on-malformed` skill and `keep-task` delegation, in both families. Errors identify family, message `m` and `stored-root`; both return exit 0 with explicit incomplete diagnostics. Earlier probe attempts rejected unsupported wrapper flags, incomplete configuration or missing occurrence members before extraction; they are not counted as preservation evidence. Private CLI probes set `XDG_CACHE_HOME`, `XDG_CONFIG_HOME` and capture state inside their disposable directory and never modify native source databases.

## Limits and retained history

AL-62 `SUMMARY.md:210-212` records historical native Codex generation with 26 source/parent identities and source-unavailable API preservation; `:233-235` records historical regression coverage. Those are historical evidence only; they do not establish a new successful AL-55 backfill or native OpenCode parity. This pass did not rerun full make, use native OpenCode recovery, simulate concurrent source writers/power loss, or execute changed-parent conflict probes; parent guards were traced statically and unchanged parent preservation executed. Optional hydration/API visibility is outside this assigned data pass and belongs to the parent/behavior evidence.

Original AL-55 ledger was read only. Current and `git show HEAD:alspaces/AL-55/agentic_metrics.json` SHA256 both equal `77625de54c0125576bbf48240d9e41e16a2db1497e289283e52e99cf71d4b3b2`. No new product tests, runtime edits, native-store writes, commits, PRs or subagents were made.
