# FIX-21 behavior re-review, round 2 (beh-opus-r2)

- Phase: re_review round 2. Focus: automatic capture end to end (worker -> Reconcile -> `generate --capture-auto` -> publication), with manual, backfill, live, OpenCode and Claude paths unchanged.
- Reviewed revision: `86eb6dae` (corrections delta `0f73f759..86eb6dae`, full change `133100bb..86eb6dae`).
- Requested profile: Claude Opus 5.5 / medium. Runtime model ID: claude-opus-5-5. The effort level could not be confirmed.
- Verdict: **CHANGES_REQUESTED**. The trigger is R2-BEH-01, which is at the parent's discretion: it is a defect unless the parent records it as an accepted limit and reconciles the spec.

## Round-1 findings in this focus

| ID | Status at 86eb6dae | Evidence |
|---|---|---|
| BEH-01 | Resolved | `children.go`: `last := cursor`, and an exhausted sweep with `last == ""` now `continue`s without writing a row. So a resumed sweep that is exhausted on its first candidate rewrites the same cursor and is never closed, and a new sweep that admits nothing stays unstarted. A sweep that completes still writes `""`. `TestResumedSweepAtTheCeilingKeepsItsFileCursor` and `TestBrandNewSweepAdmittingNothingWritesNoCursor` pass (run in a scratch copy). |
| BEH-02 / ARCH-03 | Resolved | The CLI makes one `capture.RetainAcceptedChildLinks` call. The registry is opened with `OpenProjectContext` only when `Added > 0`. On error the CLI prints "publication not completed" and exits `capturePublicationBusyExit` (75) when `IsRegistryContention` holds, otherwise 1. This matches the other call sites (main.go:592, generation_publication.go:235). The worker treats exit 75 as busy-only (capture_worker.go:501). `ReconciliationDiagnostic` and `AddRetainedChildLinks` are now unexported. `TestRetainAcceptedChildLinksReportsRegistryContention` passes. Limit: no CLI test drives exit 75 through `runGeneration`, but the path is a three-line branch. |
| DATA-02 (behavior side) | Resolved for the cases the rule handles | `Report.Refused` and `UnresolvedLinks()` now feed `addUnresolvedChildLinks` at all three registry sites (owner annotation, archive regroup, CLI), and the stderr line reports the refused count. Probe: a timed group with fresh `[C1,""]` and accepted `[C1,C2]` reports `C2` as refused and unresolved. The second pass behaves differently; see R2-BEH-01. |
| DATA-01 (cursor/annotation interplay) | Resolved | `refreshTicketOwners` applies `RetainAcceptedChildLinks` against `previousSnapshots` before `RetainIncompleteGroupSummaries`. The annotation then applies the same rule again, which is idempotent on an all-empty group (probe r7). `TestReconcileInterruptedAnnotationKeepsRegistryLinks` passes. |
| ARCH-02 parity | Behavior preserved | `exactSessionContext` and `conflictingNativeParent` reproduce both earlier checks exactly: required id and cwd, and two ancestry spellings that must agree. `RolloutHeader.SessionContext` still rejects `!First` before the shared check. Error strings and the `ErrSessionContextMismatch` wrapping are unchanged. |
| ARCH-04 (gz) | Resolved | A `.gz` header is admitted only when the full 64 KiB bound fits, and the charge is still the bytes actually read. The side effect is that an identification share under 64 KiB never admits a `.gz` header. That falls under the accepted "very small `max_source_bytes`" pending limit. |

## New findings

### R2-BEH-01: a partly re-read ambiguous group loses its unread links after one partial pass, and later passes report it as complete (introduced by the round-2 rule; severity 4)

- **Location:** `backend/internal/domain/child_links.go`, the `!empty || len(fresh) < len(accepted)` refusal branch of `reconcileOwnerChildLinks`, and `refusedChildLinks`.
- **Trigger:**
  - A Codex parent issues two or more `spawn_agent` calls with the same agent type, description and invoked_at but different task names. Parallel calls in one turn fit this.
  - Each fresh read resolves a delegation by its own task name through `ChildReference`/`AgentPath`. A pass whose budget reads child C1 but not C2 therefore produces the fresh group `[C1, ""]`, while the accepted group is `[C1, C2]`.
- **Consequence:**
  - **Pass 1:** the group is refused. C2 is dropped from the registry snapshot and from the published ledger, counted once as unresolved, and the attempt is partial.
  - **Pass 2:** nothing has changed and C2 is still unread. Both the registry and the ledger now hold `[C1, ""]`, so nothing is refused or retained. Unresolved child evidence no longer includes C2.
  - **Outcome:** the reconciliation can be recorded as complete while the committed ledger has lost an accepted link. On a machine that has the C2 rollout, a later read restores the link. On a cloned machine without it, the link is permanently lost and nothing reports it.
  - **Contract conflicts:**
    - The design's own Problem statement for "Ordinal group retention" names this failure: the published ledger lost the links, so later passes had nothing to report.
    - The spec scenario "Parent refreshed while some children are not read" says every accepted `child_session_id` remains.
    - The more specific "Retain ambiguous groups by ordinal position" requirement mandates the refusal. The spec is therefore internally inconsistent for this input.
- **Evidence:** isolated domain probe in a scratch copy of `86eb6dae`. A test-only file applied `RetainAcceptedChildLinks` twice; the second call took the first call's output as accepted.

  ```
  pass1: P[0]=C1 P[1]="" refused [{P C2 ambiguous}] unresolved [{P C2}]
  pass2: P[0]=C1 P[1]=""  refused [] unresolved []
  ```

  `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses` covers only the all-empty fresh group.
- **Suggested direction (parent's or planner's call):** one option is to allow positional carry when every set fresh link in the group equals the accepted link at the same position, a prefix-consistent alignment. That keeps the append-only argument and invents nothing. The other option is to record this as an accepted limit and reword the first scenario so it excludes duplicate-key groups with some children re-read.

### R2-BEH-02: an untimed delegation with a unique key lets an unconfirmed fresh child replace the accepted link (introduced by the round-2 rule; severity 2)

- **Location:** `reconcileOwnerChildLinks`. The unique branch requires `key.timed`. An untimed one-to-one group therefore goes to the position branch, which refuses whenever the fresh link is set.
- **Trigger:** a delegation with no invoked_at, and a fresh read whose resolved child D differs from the accepted child C1 while D's parent is not confirmed.
- **Consequence:** D stays on the delegation and C1 is reported as `unknown_provenance`. With a timed key, the same input keeps C1 (`kept_pending_confirmation`).
- **Contract conflict:** the spec says an unknown invocation time is "its own key value". The scenario "Different child without confirming metadata" says the accepted link remains until a later read confirms the new child. The two treatments of an equivalent unique key are inconsistent.
- **Mitigating factors:** the input is rare, because a fresh link set from the parent transcript usually equals the accepted one. The divergence is also reported as unresolved and partial, not silent.
- **Evidence:** probe result for the untimed case, `P[0]=D refused [{P C1 unknown_provenance}]`, against the timed case, `P[0]=C1 decision kept_pending_confirmation`.

### Observations (not findings)

- A fully re-read timed duplicate group, fresh `[C1,C2]` against accepted `[C1,C2]`, emits `refused_ambiguous` decisions with an empty `ChildSessionID`. They are never counted or shown and do not change the delegations. They are diagnostic noise only. A reader of `Report.Decisions` could mistake them for refusals.
- An accepted key that has no fresh delegation at all, for example a fresh owner read with fewer delegations, is neither retained nor reported. Codex transcripts only append, so this needs a truncated owner read. This behavior existed before round 2.

## Regression checks with no defect found

- **Double counting:** `countedChildLinks` deduplicates retained and refused links across the registry sites and the CLI call. The CLI `kept`/`refused` maps subtract kept links from refused. The DATA-03 name hint only merges unregistered queued paths and never sets identity.
- **Idempotence:** an ordinal carry re-applied to its own result changes nothing (probe r7). A repeated all-empty partial pass carries the same links again and recounts them as unresolved, so the attempt stays partial and the ledger bytes stay unchanged.
- **Misplacement:** the position branch runs only when every fresh link in the group is empty, after contradicted fresh links are cleared, and when the fresh group has at least as many delegations. The per-child contradicted and held-by-sibling guards still apply. Carry after clearing a contradicted fresh link matches the unique-key behavior.
- **Known (DATA-04):** registry snapshots take precedence over the ledger for child evidence. A Known child without a parent is unobserved, which does not block retention.
- **Unchanged paths:** retention is invoked only from capture reconciliation and the `--capture-auto` branch. Manual and backfill generation, live, OpenCode and Claude are untouched: there is no call site, and retention is gated on `Harness: SourceCodex`. `TestManualGenerateLeavesLinkToMissingChildEmpty` passes.

## Verification performed

- I read the corrections delta for domain, capture, codex and the CLI. I traced the cursor loop, `ReconcileAt` counting, the contention helpers, the worker exit-75 handling, and Codex child-link resolution (`source_occurrences.go`, `occurrences.go`).
- The domain probe ran on a `git archive 86eb6dae` copy in the scratchpad.
- Focused tests, run in that copy with the embedded dist copied in:
  - `go test ./internal/domain -run ChildLink`
  - `go test ./internal/adapters/capture -run` for the eight FIX-21 retention, cursor and window tests
  - `go test ./cmd/agentic-lens -run 'TestCaptureAutoKeepsCommittedChildLinks|TestManualGenerateLeavesLink'`

  All passed. I did not rerun the full package suites, by instruction. I reused the race, large-store and mutation results in TESTS.md.

## Gaps

- No end-to-end worker run against a contended registry, so the exit-75 path is verified by reading only.
- No real Codex transcript was checked to confirm how often parallel `spawn_agent` calls share an invoked_at value. R2-BEH-01's frequency is inferred from the delegation key and task-name resolution logic.
