# FIX-21 behavior and integration review (beh-sonnet)

- Run: FIX-21 discovery run `beh-sonnet`, focus behavior and integration. Requested identity: Claude Sonnet 5.5, effort high. Runtime metadata was not inspected beyond that.
- Reviewed revision: frozen candidate `0f73f759` against base `133100bb` (`git diff 133100bb 0f73f759`). The worktree differs from the candidate only in `tasks.md`.
- Scope: end-to-end automatic capture path (worker, `Reconcile`/`ReconcileAt`, `generate --capture-auto`, ledger and delivery publication) against the three delta specs and design D1-D7.
- Verdict: **CHANGES_REQUESTED** (one medium defect, F1). Everything else in scope is approved.

## Findings

### F1 (medium, severity 4) - A resumed child sweep that cannot admit its first candidate is closed instead of left pending

- Location: `backend/internal/adapters/capture/children.go`, starts-class sweep in `recoverChildren`: the `previous := last` / `last = previous` handling on `opencode.ErrCaptureBudget`, followed by the `discovery_cursor` write.
- Classification: introduced by this change (the new resume branch), with an analogous pre-existing weakness at base (see below).
- Expectation: design D6 and the spec "Identification never borrows from extraction" say that when identification reaches its ceiling, discovery stays pending and the next pass continues from that candidate. The code comment says "Resume at this candidate once identification has bytes again."
- Defect: `last` starts as `""` on every sweep, and entries at or before the saved cursor return before `last` is updated. When a sweep resumes from a file-level cursor and the first candidate after the cursor cannot be admitted (remaining identification < `min(size, 64 KiB)`), `previous` is `""`, so `last = previous = ""`. The code then writes `discovery_cursor.path = ''`, which the rest of the code reads as "sweep completed, closed until new evidence" (`recoverChildren` `continue`, `pendingChildWork` `NOT EXISTS ... path=''`, `CodexChildDiscoveryCursors` excludes it).
- Reachable trigger: automatic generation uses `ReconcileAt`, which does not reopen closed sweeps (only `Reconcile` with a ticket does, via `reopenTicketCodexChildSweeps`). Two or more Codex parents with pending starts in one ticket (a root plus a child that also spawns, both with SubagentStart hints), a modest `max_source_bytes` (the setup manual's example is 8 MiB, so identification gets 2 MiB), and a store with many in-window rollouts. Parent A's sweep drains the quarter to below one header. Parent B resumes from its file-level cursor and its first post-cursor header does not fit.
- Consequence: B's unread children are never discovered by the worker or `--capture-auto` until new native evidence reopens the sweep (a new SubagentStart callback or the parent's source changing). Meanwhile `DiscoveryPending` was true for one pass only, and the worker's `CodexChildDiscoveryCursors` sees no pending sweep, so it can exit. Discovery looks complete but is not.
- Evidence: probe in a scratch copy of the candidate (`.../scratchpad/probe/backend/internal/adapters/capture/zz_probe_test.go`, test `TestProbeResumedSweepClosed`). Two task families under one ticket, both parents bound, four SubagentStart hints each. Small passes via `ReconcileAt` at budgets 12 KB to 200 KB, followed by one pass with 64 MiB.
  - With a stderr trace added to the scratch copy, every sampled pass logged `cursor write "" exhausted true visited 1`.
  - Candidate as is: after the huge pass, `rows=6` of 10, `DiscoveryPending=true`, `PendingChildren=4`, cursors closed (every budget from 12 KB to 200 KB).
  - Same probe with `s.Reconcile` (which reopens sweeps): all 10 owners found. The loss is specific to the automatic path.
  - One-line change in the scratch copy (`if previous == "" { previous = cursor }` before `last = previous`): all 10 owners found at every budget, and cursors stay open while pending.
  - The base commit shows the same closure at tiny budgets in the same probe (copy in `.../scratchpad/base/backend`), because the old code also wrote `last = ""` when bytes remained but too few for any candidate. So the closure is not new in kind. The new branch was written to prevent it and does not on resume.
- Test gap: the large-store and window tests start each ceiling case from a fresh sweep, where the root directory entry makes `previous` non-empty. No test resumes a file-level cursor at the ceiling. Suggested test: two parents, first sweep exhausted mid-store, second pass drains the quarter with the other parent, assert the cursor is not `''` and a later pass discovers the child.
- Suggested fix (parent owns): keep the incoming `cursor` as the resume position when `previous` is empty, or do not overwrite the cursor when exhaustion happens before any candidate was admitted.

### F2 (suggestion, severity 2) - Accepted links dropped by the rule's refusals are not visible in capture output

- Location: `domain/child_links.go` outcomes `refused_ambiguous`, `refused_held_by_sibling`, `refused_contradicted`, and `cmd/agentic-lens/capture.go` `retainCaptureChildLinks`.
- Classification: suggestion; consistent with the approved spec (those links are intentionally not moved).
- Observation: the CLI reports retained and cleared counts only. An accepted link refused for ambiguity (identical agent type, description and invoked-at on both sides), for an unknown invocation time (for example delegations migrated from older ledgers), or held by a sibling is dropped when the fresh owner replaces the ledger owner, with no stderr line, no coverage count and no diagnostic. If the report only has refusals, nothing is printed at all.
- Consequence: a partial refresh can still shorten a parent's link list without the operator being told, which sits uneasily with "never delete accepted facts" even though the spec allows it.
- Suggestion: print and persist the refused count with the existing partial diagnostic, or state the limit in the delivery summary.

### F3 (suggestion, severity 1) - `retainCaptureChildLinks` store open differs from neighboring helpers

- `capture.go` opens the registry with `capture.OpenProject` and a 2 s diagnostic context, while the adjacent generation helpers use `OpenProjectContext`. A contended registry here exits with status 1 ("Cannot record retained child links") instead of the busy exit code the worker treats as a short retry. Neighboring pending-snapshot helpers also exit 1, so this is consistent in practice, hence only a suggestion.

## Hypotheses investigated

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | Worker -> `generate --capture-auto` loses accepted links when the registry has no earlier snapshot (portable ledger, machine B) | Refuted. `retainCaptureChildLinks` runs once in `main.go` after reconcile and before pending sync, merge, restore acknowledgement and publication, against `baseline.Current`. The CLI test `TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState` passes on this candidate (focused run, 5.5 s). |
| H2 | Retained links in pending snapshots, restore completion or `mergeSessions` disagree with the ledger | Refuted by reading. Pending sync, `completeMaterializedCaptureRestores` and `verifyPendingPromotionCoverage` consume the already mutated `RefreshedSessions` and `RestoredSessions`, so every consumer sees the same links. |
| H3 | Retention fabricates or moves links | Not found. Keys require unique agent type, description and known invoked-at on both sides, contradicted children are cleared, a different child replaces an accepted link only when confirmed, and sibling-held children are refused. Idempotence is argued from the fixpoint and covered by tests. |
| H4 | Partial coverage is not reported when the registry never registered the children | Refuted. `AddRetainedChildLinks` dedupes against `pending_starts` and `pending_children`, and the CLI helper re-records the diagnostic as partial through `ReconciliationDiagnostic`. `CoverageSummary` is rendered after the helper. A later link is added to `PendingChildren` only, never to `UnresolvedParentLinks`. Consequence noted: a link to a fully registered, current child still counts as "unresolved child evidence" whenever the parent is refreshed without that child, as the spec states. |
| H5 | Unassigned archived rollout still loops the worker | Refuted. `inspectRegisteredCodexPath` no longer fails on not-found for unassigned rows, relocation requires exact proof (`resolveRegisteredCodexSnapshot`, `validateCodexOwnerMetadata`, `keepsAcceptedCaptureParent`, project proof), and unproven, ambiguous or conflicting outcomes are isolated (row `pending`, facts and path kept) and excluded from worker success and `lastFailure`. Other errors (symlink loop, I/O) still fail the pass. Isolated rows are retried on later passes, bounded by the existing refresh cycle rather than retried until lifetime expiry. |
| H6 | Manual `generate`, backfill, live hydration unchanged | Confirmed by reading. `main.go` adds one call under `--capture-auto`. Backfill and the manual path still call `codex.DiscoverRolloutPaths`. Live hydration still calls `DiscoverExactRolloutPathContext`. The window, exact-UUID skip and header cache are only reachable through capture-only entry points (`DiscoverExactCaptureRolloutPath`, `childCreationWindow`). The legacy ambiguity error keeps its message and now also satisfies `ErrOwnerAmbiguous`. |
| H7 | OpenCode or Claude Code behavior changed | Claude Code: no code path touched. OpenCode: link retention and window rules are Codex only, but the budget split applies to all capture passes. OpenCode extraction is capped at three quarters during the rotating rounds and OpenCode child pages draw on the identification quarter. An owner refused only by the reservation is deferred and retried in the final leftover refresh with every remaining byte, so reads that fit the total budget still happen; an owner larger than the total is attempted twice and ends pending as before. This is a bounded accounting change, noted as a limit. |
| H8 | The UUIDv7 window or directory rule drops real children | Not found on reading. Names and directories only exclude; inclusion, identity and parent come from `session_meta` (nested `thread_spawn` parent included through `decodeSessionMetadata`). Residual risks are the ones the design records (5 s clock tolerance, renamed rollouts). |
| H9 | Header cache or admission rule charges wrongly or abandons reads | Not found. Admission requires `min(size, 64 KiB)` to fit, errors that are not budget errors are cached per pass, budget refusals are not cached. Minor edge: a `.jsonl.gz` header can pass admission on compressed size and then stop at the allowance mid-read (charged what was read, not cached). Not reachable in a way that changes accepted facts. |
| H10 | Identification stops at its ceiling as "pending, not failed" on every path | Refuted for the resume case, see F1. |
| H11 | Registry snapshot and ledger diverge after the adapter retains links and the CLI retains again | Refuted. A link already held by the fresh snapshot is not a decision, so it is counted once, by the adapter. |

## Verification performed

- Read the full diff for `domain/child_links.go`, `capture/reconcile.go`, `children.go`, `source_database.go`, `archive_projection.go`, `codex_context.go`, `codex/discovery.go`, `discovery_budget.go`, `cmd/agentic-lens/capture.go`, `capture_worker.go` and the `main.go` call site, and the unchanged consumers listed in the brief (`resolveSourceChildren`, `mergeSessions` call path, pending-snapshot sync, restore completion, worker success logic, `CodexChildDiscoveryCursors`, `reopenTicketCodexChildSweeps`).
- Focused run in the worktree: `go test ./cmd/agentic-lens -run TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState -count=1` passed.
- Isolated probes in scratch copies only (`git archive` of `0f73f759` and `133100bb` under the scratchpad). The probe test and an instrumented, patched `children.go` exist only in that copy. No product or test file in the worktree was modified; the only file written there is this report.
- Reused, not rerun: whole-package results, race runs, large-store test and mutations M1-M6 recorded in `evidence/TESTS.md`.

## Gaps and limits

- I did not run the full `capture`, `codex` or `cmd/agentic-lens` suites (host load, per brief).
- F1 was demonstrated at probe budgets far below production defaults. Its production reachability rests on the multi-parent plus modest-budget reasoning above, not on a production-scale reproduction.
- I did not exercise the worker process end to end. The worker-isolation conclusion rests on code reading plus the existing tests (`capture_corrections_isolation_test.go`, M4).
- Cross-focus: no data-integrity or architecture defect beyond F1 and F2 was encountered.
