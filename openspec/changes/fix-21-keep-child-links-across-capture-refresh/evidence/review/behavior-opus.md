# FIX-21 behavior and integration review (run beh-opus)

- Role: reviewer-behavior (behavior and integration focus), first pass, independent.
- Requested identity: Claude Opus 5.5 / medium. Runtime metadata reports model `claude-opus-5-5`; effort is as requested, not independently observable.
- Reviewed revision: candidate `0f73f759` against base `133100bb` (`git diff 133100bb 0f73f759`). Working tree HEAD matched `0f73f759`; the only local change was the orchestrator's `tasks.md` edit, which I did not touch.
- Verdict: **APPROVED** for the behavior and integration scope. There are no blocking findings, but there are two low-severity observations.

## Path traced

1. **Worker.** `captureWorker` runs the unassigned `Reconcile`, then `generate --capture-auto` per ticket.
2. **Reconcile.** `ReconcileAt` runs these steps in order:
   - `newBudgetSplit` and the rotating rounds;
   - `releaseSplit`, then the final leftover `refreshTicketOwners`;
   - `prepareReadOwner`, then `selectCodexRollout`, then `DiscoverExactCaptureRolloutPath`;
   - `finishCaptureAnnotations` and `regroupArchiveFamily`, each running `RetainAcceptedChildLinks` against `previousSnapshots`;
   - `pendingChildWork` and `pendingChildLinks`, then `AddRetainedChildLinks`;
   - the deferred diagnostic, saved as `result.reconciliation`.
3. **CLI (`runGeneration`).** It runs these steps in order:
   - `retainCaptureChildLinks(baseline.Current, captureResult)`, which runs before lifecycle check, pending sync, merge and publication;
   - `fillMissingCaptureSessions`;
   - `mergeSessions(Restored, Refreshed)`;
   - `retainUnrefreshedGenerationOwners`;
   - `CoverageSummary(*captureResult)`, which sees the updated `PendingChildren` through the pointer;
   - `recordGenerationPublication`.
4. **Codex header and discovery.** `readRolloutHeader` and `decodeSessionMetadata` still apply the `source.subagent.thread_spawn.parent_thread_id` fallback, so child sweeps keep finding children whose only parent spelling is the nested one. Conflicting parent spellings are still rejected through `ConflictingParent`.

## Hypotheses investigated

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | Manual `generate` and backfill changed. | Refuted. `retainCaptureChildLinks` runs only under `*autoFlag`. `DiscoverRolloutPaths`, `listRolloutPaths` and `inspectRollout` are unchanged. `discoverExactRolloutPath` (live hydration, `DiscoverExactRolloutPath*`) changed only its ambiguous error type, and the message is identical. `discovery_budget.go` is purely additive. |
| H2 | Retention keys never match across a ledger round trip, so machine B never retains. | Refuted. `DelegationFactDTO` converts to the domain type directly, with no timestamp normalization. Codex `InvokedAt` is the raw source timestamp. The CLI test proves all four links are kept on machine B. |
| H3 | Retained links are double counted between the adapter and the CLI. | Refuted. `countedChildLinks` survives the `ReconcileAt` return, `AddRetainedChildLinks` deduplicates, and links the adapter already retained produce no CLI decision. |
| H4 | Contradiction evidence wrongly clears a legitimately read link. | Refuted for current data. `ParentID` is intrinsic to `session_meta` (the immediate parent), and fresh evidence takes precedence over the accepted record. A ledger with a missing parent counts as unobserved, not contradicted. |
| H5 | `PendingChildren > 0` from retained links keeps the worker alive or looping. | Refuted. Worker continuation depends only on `CodexChildDiscoveryCursors`, inbox progress and restore rotation, not on `PendingChildren`. |
| H6 | An isolated unassigned row still fails the pass or extends the lifetime. | Refuted. `IsolatedSources` is excluded from the success/`lastFailure` loop, and the row is `CapturePending` with its snapshot preserved. Budget errors keep their retry, and other I/O errors are still not isolated (`unassignedRelocationIsolation`). |
| H7 | Unassigned relocation adopts a file without proof. | Refuted. It requires `s.project != ""`, `ValidateProject` on the recovered cwd, an exact unique `session_meta` (ambiguity reported) and `keepsAcceptedCaptureParent`. Recovery that escapes the store is classified as a conflict. `TicketID` stays empty. |
| H8 | The window pre-filter drops a real child, or a resumed cursor breaks. | Refuted within approved scope. The file rule skips only conventional UUIDv7 names created more than 5 s before the parent. The directory rule applies only under `sessions/` with a 14 h + 5 s margin. The cursor comparison runs before the window checks. `visited` counts only candidates that are not pruned. |
| H9 | Publication bookkeeping breaks when CLI-retained links differ from the registry snapshot. | Partly confirmed, informational only. See O2. |

## Findings

No finding requests changes. The observations below are low severity and are noted for the parent.

### O1. A registry-contention failure in the new diagnostic write exits 1, not with the busy code

- **Classification:** introduced, minor.
- **Location:** `backend/cmd/agentic-lens/main.go:601-604`, which calls `retainCaptureChildLinks`. That function in turn calls `RecordDiagnostic` through `OpenProject` (`capture.go:823-831`).
- **Expectation:** the adjacent reconcile error path maps `capture.IsRegistryContention` to `capturePublicationBusyExit`, and the worker gives busy-only failures a short retry (`busyOnly`).
- **Trigger:** `--capture-auto` retains at least one new link, and SQLite is busy or locked during the diagnostic write (2 s timeout).
- **Consequence:**
  - The generation exits 1.
  - The worker records a `publication failed` diagnostic and doubles its backoff instead of taking the short busy retry.
  - The ledger is not published that pass. No accepted facts are lost.
- **Evidence:** code inspection only. I did not reproduce it.
- **Severity:** 2/10.

### O2. CLI-retained parents never match the registry snapshot, and coverage stays partial on a machine whose registry never sees the children

- **Classification:** introduced by design (D5) and informational.
- **Location:** `capture/diagnostics.go:155-193` (`RecordPublication`) together with `retainCaptureChildLinks`.
- **Behavior on machine B:**
  - The adapter saves the parent snapshot without the links, because there is no previous snapshot.
  - The CLI then adds the links in memory and in the ledger only.
  - The publication diagnostic's `association` for that parent is therefore `unknown`, with no fingerprint or store.
  - Each pass that re-reads the parent re-retains the links, records the reconciliation as partial, and raises the health issue "Recorded child evidence remains unresolved; continue capture reconcile."
- **Why it is not a defect:** `association` and `refresh_state` have no reader beyond the diagnostic record (repo grep). The persistent partial result is the approved D5 semantics: B genuinely cannot re-prove the links.
- **Suggestion:** the remediation hint "continue capture reconcile" cannot resolve this on machine B, so consider adjusting it or recording the limit in delivery.
- **Severity:** 1/10.
- **Related over-count, same class:** within one pass, a child refreshed in one `refreshTicketOwners` call and its parent in another (a rotating round, then the final leftover) leaves the parent's link retained and counted as pending, even though the child was read. This is conservative over-reporting of partial coverage, not a fact error.

### Test robustness note, for the tests reviewer

- `TestCaptureWorkerIsolatesUnrecoverableUnassignedRollout` failed once at host load average of about 18. Its fixture's first worker pass was killed by `runCaptureNative`'s 8 s harness timeout (`capture_corrections_test.go:26`), before any FIX-21 behavior was exercised.
- An immediate rerun passed in 7.17 s.
- The test also asserts that the worker finishes in under 8 s against a 10 s policy lifetime, which is load sensitive.
- This is not a product defect.

## Approved behavior changes

These changes are visible to the user. They are not findings.

- OpenCode child-page discovery under capture now draws on the identification quarter (`identificationRemaining() < 256`), and OpenCode owner validation draws on the extraction share. Link resolution for OpenCode and Claude Code is untouched: `RetainAcceptedChildLinks` filters to `SourceCodex` owners.
- The coverage summary's "Source attempts" no longer includes Codex header reads.
- An empty store (`errNoRolloutFiles`) found during exact capture lookup now wraps `ErrOwnerUnavailable`. An unassigned row that is relocating is therefore isolated as unavailable rather than failed. A bound row is unaffected, because it is still a failure through the existing path.

## Verification performed

- I read the diff for the domain, capture, codex and CLI files, plus the unchanged consumers: `mergeSessions`, `fillMissingCaptureSessions`, `filterCapturePrecedence`, `RecordPublication`, `resolveSourceChildren`, `decodeSessionMetadata`, the worker success and continuation logic, and `HealthIssues`.
- `go test ./internal/domain/ -run 'ChildLink|RetainAccepted' -count=1`: ok.
- `go test ./internal/adapters/capture/ -run 'TestReconcileRetainsAcceptedChildLinksThroughPartialRefresh|TestReconcileFreshContradictingChildClearsItsRetainedLink|TestReconcileArchivedUnassignedCodexChild|TestBindArchivedUnboundCodex' -count=1`: ok.
- `go test ./cmd/agentic-lens/ -run 'TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState|TestCaptureWorkerIsolatesUnrecoverableUnassignedRollout' -count=1`:
  - the CLI retention test passed;
  - the isolation test failed once under load (fixture timeout), then passed on rerun with `-run ...Isolates... -v` in 7.17 s.
- I reused `evidence/TESTS.md` for the full-suite, race, large-store and mutation results, and did not rerun them.

## Gaps

- I did not run a disposable probe for O1 (simulated SQLite contention during the retention diagnostic).
- The large-store byte accounting and the reserved-quarter totals were not independently re-derived. I relied on `TESTS.md` and code reading.
- I did not exercise dashboard readers of delegations. Event `ChildSessionID` is unchanged by design (D4), and retained links appear only in `Delegations`, which existing readers already consume.
