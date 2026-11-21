# FIX-20 cancellation regression tests

Focused checks ran in the uncommitted FIX-20 backend against baseline `83f6f6c1` on Go 1.26.0, Linux amd64, and modernc SQLite v1.60.1. The [delivery record](../delivery.md) identifies the final verified revision.

| Boundary | Fixture and assertions |
| --- | --- |
| Live SQL and cache | A real v1 trace enters a SQLite VM loop through a test-local view. Cases exercise the initial step and a row that an indexed query can evaluate after yielding its first row. A second lookup registers its cache wait and cancels while the native query remains held. Cancelling the first lookup interrupts SQL before its completion marker. An independent lookup reuses the accepted entry and observes a genuine revision change. The pre-cancellation and observer-failure cases verify stable cache identity and recovery after a real observer connection closes. |
| Live listing order | An older listing pauses before staging its cache admission. A newer listing observes a written session and accepts its trace. Resuming the older listing leaves the newer trace cached at the same source revision. |
| Historical events | The HTTP GET hydrates a native Codex owner before cancellation during either SQLite connection acquisition or an already-running SQLite row scan for the second owner. The native scan case gates a real SQLite comparison and checks that the query did not finish scanning. Partial coverage keeps the earlier event and reports both required owners. An already-cancelled request reports unavailable coverage. Fresh reads return both events and the ledger bytes remain unchanged. |
| OpenCode exact capture | Real v1 and v2 files contain one source owner and message. A native collation gate synchronizes cancellation after byte preflight during exact body extraction. The cancelled snapshot returns no owner, leaves accepted evidence intact, and a fresh read succeeds on the same one-connection pool. |
| Codex source | Native plain JSONL and gzip discovery cancel after actual source bytes were counted. Already-cancelled exact selection and path parsing report cancellation, and independent reads recover. A controlled reader cancels while returning the next parser record and verifies that the interrupted record is excluded. |
| Backfill publication | A previous workspace commits before the next workspace reads one owner and interrupts the second after native Codex discovery consumes bytes. The report records cancellation and incomplete refresh, both files in the interrupted and later workspaces retain their exact bytes, and the earlier ledger and summary keep their committed updates. A portable OpenCode owner is present in the interrupted ledger. |

Run from `backend/`. The following checks passed:

- `go test -count=5 ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill -run 'Test(CodexDiscoveryCancellation|CodexParserCancellation|ExactSnapshotCancellation|BackfillCancellation)'`
- `go test -count=5 ./internal/adapters/opencode -run '^TestExactSnapshotCancellationReleasesTransaction$'` after moving the collation gate to body extraction.
- `go test -count=3 ./internal/adapters/live -run 'TestLiveTraceCancellation|TestHistoricalEventsCancellation|TestSourceRevisionReopens|TestLiveTraceLegacyRevision' -v`
- `go test -race -count=1 ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/live ./internal/adapters/backfill ./internal/adapters/http`. This also ran existing uncancelled v1, v2, Codex, ambiguity, ordinary partial refresh and mixed-owner controls.

A disposable `git archive 83f6f6c1` probe used the same live SQLite fixture with the pre-cancellation and cache-wait assertions omitted to isolate the native read. Its targeted test failed at the expected guard with “cancelled request kept running native SQLite work”. The test deadlines are failure guards. The fixtures cover the SQLite and regular-file boundaries exercised above. They establish no latency guarantee for arbitrary system calls or drivers. The cancellation scenario in backfill uses caller cancellation. The production per-workspace timeout was checked by context propagation and did not run to wall-clock expiry in this suite.

The correction for BEH-01 adds a real rollout fixture to the historical GET tests. It cancels at the next parser record after a completed tool call, checks that the event remains visible, leaves the owner outside complete coverage, and verifies that an incompatible native parent still prevents attribution. The ledger stays byte-for-byte unchanged. The orchestrator reran all five affected adapter suites with `-race -count=1` on the corrected fingerprint `9a7308d5…d3ddf39`. The separate full CLI test passed without race instrumentation on the initial fingerprint. The tester's duplicate long CLI run was stopped before verification of its result.

On `9a7308d5…d3ddf39`, the isolated Go overlay for the later indexed SQLite step exposed its native completion marker despite a cancellation error. The older listing overlay exposed eviction of the newer accepted trace. The already-running historical SQLite read passed on that backend. A separate overlay that severed context propagation for the native query while leaving connection-acquisition coverage intact failed because SQLite scanned all the fixture rows.

After the SQL and cache corrections, this command passed five race-instrumented runs of both packages on the backend diff and untracked source fingerprint `cb761ce76f5750e7dbb553a678fead64dae256769bc83229d96672d472e5eeb6`:

```bash
go test -race -count=5 ./internal/adapters/live ./internal/adapters/opencode \
  -run 'TestOlderLiveTreeCannotEvictNewerCache|TestLiveTraceCancellationInterruptsSQLiteAndPreservesCache|TestHistoricalEventsCancellationKeepsEarlierCodexOwner|TestExactSnapshotCancellationReleasesTransaction|TestSourceRevisionReopensAfterActualObserverFailure|TestLiveTraceLegacyRevision'
```

## Additional regression coverage after review

Native SQLite probes now assert that the underlying query stopped. The active-listing case synchronizes inside a SQLite VM loop and checks its completion marker. The exact v1 and v2 snapshot cases interrupt body extraction after byte preflight and ensure the remaining fixture rows are not scanned. OpenCode backfill uses a three-second caller deadline that expires after the native read has started. The interrupted scan reports the inherited deadline, leaves the ledger and summary unchanged, and releases its resources for a fresh backfill. A separate publication test cancels immediately after a successful Codex recovery check and verifies that the last publication guard preserves both files.

The historical spawn case uses two native rollouts. The interrupted owner's completed spawn event remains visible with partial coverage. Its child link is retained when the already-read child's native metadata agrees and cleared when that metadata names another parent.

Each disposable mutation failed the corresponding assertion. Detaching the OpenCode backfill context, the exact body query context, or the live listing context let native work finish despite cancellation. Removing the final publication guard changed the file bytes. Bypassing shared relationship resolution preserved the contradictory child link.

On backend diff and untracked source fingerprint `f9e212ba932ca58316b37db7ea18d4eb0c9a09733f813e3bb43aeaa82c78b34a`, all three packages passed three race-instrumented runs:

```bash
go test -race -count=3 ./internal/adapters/backfill ./internal/adapters/opencode ./internal/adapters/live \
  -run 'TestBackfillOpenCodeInheritsCallerDeadlineDuringNativeRead|TestBackfillCancellationAfterSourceSuccessPreservesPublicationBytes|TestExactSnapshotCancellationReleasesTransaction|TestLiveSessionListingCancellationInterruptsSQLite|TestCancelledHistoricalCodexSpawnRespectsNativeChildParent|TestOlderLiveTreeCannotEvictNewerCache|TestLiveTraceCancellationInterruptsSQLiteAndPreservesCache|TestHistoricalEventsCancellationKeepsEarlierCodexOwner'
```

## Stable ledger ordering

On base `00289db5` and backend diff and untracked source fingerprint `f9248f17…e152e16`, the persistence regression alternated which session was refreshed and checked the same ID order after both writes, correct ownership of changed metrics, and the exact nested arrays and evidence of a retained session. A second test checked chronological ordering across timezone offsets, stable ties and unknown timestamps, repeated occurrences, anomaly evidence, migration steps, nullable collections and repeat serialization. The migration tests exercised v5-to-v6 conversion and metadata-only cleanup of a current ledger with unsorted sessions.

The complete adapters, backfill and domain suites passed with `go test -race -count=1 ./internal/adapters ./internal/adapters/backfill ./internal/domain/...`. Focused CLI race tests passed for committed-history preservation, generation policy and automatic capture publication. Read-only overlays of the preceding writer and migration runner reproduced the ordering failures while the current-schema metadata cleanup control passed. `make PLATFORMS=linux/amd64 verify-build` and strict OpenSpec validation passed. The configured manager selected the freshly built host engine.

On the corrected backend fingerprint `b07c6b8c…052d4d5d`, an unavailable Codex source drove two actual backfill publications. Both retained the ordered evidence of the unavailable session, a session from the other source, and the ticket-level summaries. Completing the Git step removed it from pending and appended it to completed, while existing steps kept their relative order. Five focused adapter and backfill regressions passed three repetitions with `-race`. Independent disposable overlays restoring the preceding backfill writer or migration evaluator failed the corresponding evidence or step-order assertion.

The final diff `93a81a79…bd9438d` covers a native Codex parent refresh with a known missing child. The parent model changed and stale owner-local evidence was cleared. The unrecovered child, unchanged group summary and accepted ticket summary retained their evidence and order. Adding the child source let a subsequent complete refresh replace the stale group and ticket analysis. Both backfill evidence tests passed three repetitions with `-race`. Isolated changes that skipped group evidence restoration or ordered the retained references failed the exact sequence assertion. The domain tests verified that a partial read preserves the entire ticket summary and complete coverage replaces it.
