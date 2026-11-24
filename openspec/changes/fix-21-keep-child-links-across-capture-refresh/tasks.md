# FIX-21: Execution tasks

[design.md](design.md) sets the scope and records approval. The [link retention](specs/capture-child-link-retention/spec.md), [discovery budget](specs/capture-discovery-budget/spec.md) and [archived recovery](specs/capture-archived-unassigned-recovery/spec.md) specs define observable acceptance. No task starts until approval is recorded in the design.

## Link retention

- [x] **Coder:** add the pure domain rule `domain.RetainAcceptedChildLinks` in `domain/child_links.go`, with typed parent-evidence and outcome states, a typed report and Codex scoping.
- [x] **Tester:** add table cases in `domain/child_links_test.go`:
  - a child that is absent and is retained
  - a confirmed child whose fresh link is empty
  - a fresh child that contradicts, including one with no parent
  - an accepted child record that contradicts
  - a different child, both confirmed and unconfirmed
  - duplicate provenance and unknown provenance
  - a child already held by a sibling delegation
  - nil delegations, a non-Codex owner and an owner that was not refreshed
  - idempotence
- [x] **Coder:** apply the rule in `finishCaptureAnnotations` and `regroupArchiveFamily`. Add the deduplicated retained pairs to `PendingChildren` after `pendingChildWork` runs, through `ReconcileResult.AddRetainedChildLinks`.
- [x] **Tester:** add `capture/reconcile_child_links_test.go`. The fixture is one parent with four task-name-only `spawn_agent` delegations and four child rollouts. Cover these cases:
  - A full pass links all four.
  - The parent grows and a budget reads it with only two children. All four links are kept and unresolved child evidence rises by 2. One child already queued in `pending_starts` is counted once. Parent links are unchanged and the attempt is partial.
  - A repeated pass leaves the snapshot identical.
  - A contradicting child is read together with the parent. Its link is cleared, and source-level warning behavior is unchanged.
  - The spawn events of retained delegations in `SourceEvents` carry no child session ID.
- [x] **Coder:** add the generation helper in `cmd/agentic-lens/capture.go` and its single call in `main.go`. It rewrites refreshed and restored sessions, adds their pairs with `AddRetainedChildLinks`, records the partial diagnostic from `ReconciliationDiagnostic` and prints a line to stderr.
- [x] **Tester:** add `cmd/agentic-lens/capture_child_links_test.go`, run through `generate --capture-auto`. Seed a committed ledger with all links and empty capture state. Read the parent and only some children. Assert all of the following:
  - every link is kept
  - other owners and other-harness facts are unchanged
  - unresolved child evidence includes the unregistered children
  - the latest reconciliation is partial
  - a repeated run keeps the ledger bytes and timestamps
- [x] **Tester:** confirm that a full manual `generate --source codex` links task-name-only children. Add a case to `main_codex_test.go` only if no existing case covers this.
- [x] **Tester (review DATA-02):** cover refused accepted links. In `domain/child_links_test.go`, each typed reason (ambiguous, unknown provenance, held by a sibling) is reported once, and no refused link is moved. Through `capture.RetainAcceptedChildLinks` and `generate --capture-auto`, refused links count once as unresolved child evidence beside queued work, the attempt is partial, and stderr reports "could not place N accepted link(s)".
- [x] **Tester (ordinal group retention):** in `domain/child_links_test.go`, cover the ordinal carry for duplicate keys, a larger fresh group with appended delegations and unknown invoked_at. Cover refusal and reporting when a fresh link in the group is set or the fresh group is smaller. Cover a contradicted child and a child held by another final delegation within a group, idempotence, and repeated passes with the published output as the accepted ledger. In `capture/reconcile_child_links_test.go`, an ambiguous group's links survive two consecutive partial passes and stay counted as unresolved.

## Discovery budget

- [x] **Tester:** build a large-store fixture generator in test helpers. It creates thousands of earlier live and archived rollouts, names outside the convention, one root with children, and same-UUID copies. Measure the bytes and claims of each current consumer: child sweep, owner preflight, exact lookup and extraction. Record the measurements in `evidence/TESTS.md` before any code change.
- [x] **Coder:** in `codex/discovery.go` and `discovery_budget.go`, add the opt-in UUIDv7 window, the year, month and day directory pre-filters, and `DiscoverExactCaptureRolloutPath` with its exact-UUID pre-filter and `ErrOwnerAmbiguous`. Add the shared header reader with a pass-local `HeaderCache`, plus actual-bytes charging and read admission.
- [x] **Coder:** in `capture/children.go`, `reconcile.go`, `source_database.go` and `codex_context.go`, wire in the pre-filters, the cache and the reserved child-discovery ceiling. Its leftover is released to the final refresh. Defer owners refused by the reservation, and keep header reads out of `SourceAttempts`. Leave manual and backfill discovery and live exact lookup unchanged.
- [x] **Tester:** add `codex/discovery_window_test.go`. Cover these cases:
  - a child created after its parent, or earlier within the 5-second tolerance, is found
  - a candidate before the window costs zero bytes and zero session slots
  - malformed names and non-v7 UUIDs are read
  - a name that matches but whose metadata disagrees is decided by its metadata
  - a parent with no instant disables pruning
  - a CET/CEST-style fixture whose local directory date differs from the UTC date still finds the child
  - year, month and day directories at the 14-hour-plus-5-second bound are descended or skipped correctly
  - an exact lookup skips other UUIDs and still reports a duplicate same-UUID copy as ambiguous
  - an archived same-name copy is found
- [x] **Tester:** add `capture/reconcile_large_store_test.go`. Cover these cases:
  - the family is extracted within the budget, with exact byte totals against the fixture's known sizes
  - each header is read once, and a changed file is reread
  - extraction scheduled before discovery in a rotation phase cannot consume the reserved identification quarter
  - unused identification bytes are available to the final leftover refresh
  - identification stops at its ceiling with discovery pending, not failed, while known owners are still extracted, including their preflight charged to extraction
  - an owner refused only by the reservation is deferred with its slot refunded and is extracted by the final leftover refresh
  - a header read whose `min(size, 64 KiB)` does not fit is not started or charged
  - `SourceAttempts` excludes header reads and stays at or below `max_sessions`
- [x] **Tester:** update the changed accounting assertions in `capture/reconcile_test.go`, `fix19_budget_test.go`, `fix19_candidate_budget_internal_test.go` and `codex/discovery_budget_test.go`, with a recorded rationale for each. Confirm that the manual discovery tests in `codex/discovery_test.go` still pass unchanged.

## Archived unassigned recovery

- [x] **Coder:** allow unassigned relocation in `source_database.go` under the binding-time checks. Add `ReconcileResult.IsolatedSources` (`IsolatedSource{Identity, Reason, Detail}`). Exclude isolated rows from pass success and `lastFailure` in `capture_worker.go`.
- [x] **Tester:** extend `capture/reconcile_archived_rollout_test.go`. Cover these cases:
  - an unassigned child rollout moved to `archived_sessions/` is still read, and stays unassigned with the same parent
  - missing, ambiguous, foreign-project and conflicting-parent cases appear in `IsolatedSources` with their reason, and no other file's facts are adopted
- [x] **Tester:** add a worker case next to `capture_corrections_test.go` where an unrecoverable unassigned rollout no longer causes a lifetime-exhausted failure, and other tickets still publish.

## Integration and delivery

- [x] Run the affected package tests, raising the `./cmd/agentic-lens` timeout under host load. Record the revision, commands and results in `evidence/TESTS.md`.
- [ ] Get independent review of data integrity, behavior, architecture and tests. Resolve the findings and reverify the affected paths.
  - Candidate: `133100bb..0f73f759`. Each focus runs on both configured profiles, Opus 5.5 at medium effort and Sonnet 5.5 at high effort, giving eight discovery runs.
  - Working reports go to `evidence/review/<focus>-<profile>.md`. The orchestrator owns `evidence/REVIEW.md` and compacts it into `evidence/REVIEW_SUMMARY.json`.
- [ ] Commit the candidate and merge main. Resolve any `main.go` overlap with PR #162 so both sides' behavior is kept. Run the integrated `make` gate, which covers the frontend build and embedded assets.
- [ ] Write `delivery.md`, generate the final telemetry, then push and open the PR.

## Workflow follow-up

Archive only after delivery is verified. FIX-21 has no roadmap entry to reconcile.
