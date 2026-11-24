# FIX-21 data-integrity review (data-opus)

- Run: data-opus, first pass, focus reviewer-data (data integrity and failure handling)
- Requested model/effort: Claude Opus 5.5 / medium. Runtime model ID reported to the agent: claude-opus-5-5. Usage and cost are not available to the reviewer.
- Reviewed revision: base 133100bb .. candidate 0f73f759 (`git diff 133100bb 0f73f759`), plan 030bf0ec
- Verdict: CHANGES_REQUESTED (one reproduced acceptance gap, F1)

## Hypotheses and disposition

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | `RetainAcceptedChildLinks` invents a link, duplicates a child across delegations, or keeps a contradicted child | Rejected. Randomized property probe (200k cases, isolated copy of `domain/child_links.go`): no output child absent from fresh/accepted delegations of that owner, no duplication beyond the fresh input, no contradicted child kept |
| H2 | The rule is not idempotent, or drifts across repeated passes | Rejected. f(f(x,A),A) == f(x,A) in 200k random cases. Re-running with the published result as the accepted ledger and the same fresh read reproduces the same parent (200k cases) |
| H3 | Retention resurrects a legitimately contradicted link | Rejected for supported paths. A contradicted fresh/accepted child is cleared/refused before any carry. Only theoretical: a blocked publication plus an unread child whose ledger record still names this parent. ParentID is immutable session_meta, so not practical |
| H4 | Accepted links are still lost through a merge path (registry snapshot, archive regroup, ledger merge, pending snapshots, restore) | Partly confirmed. See F1 (registry snapshot after an interrupted pass). Regroup, ledger merge (`mergeSessions` after CLI retention), pending snapshots (written from the post-retention result) and restore (retention also applied to `RestoredSessions`) are covered |
| H5 | Retained links counted twice, or not at all, in `PendingChildren` | Mostly rejected. Adapter and CLI deduplicate through `countedChildLinks`, and regroup runs before the count. Minor over-count only: see F3 |
| H6 | Budget split/deferral can loop, starve, or overrun `max_source_bytes` | Rejected. Deferral happens only while the split is unreleased. The final leftover refresh runs once, after `releaseSplit`. Charges clamp to their shares. Identification exhaustion sets `discoveryPending` (visible) |
| H7 | Header cache serves stale or partial data | Rejected. Key is (physical path, size, mtime). Budget refusals and context errors are not cached. Admission requires min(size, 64 KiB) to fit, so a read is never abandoned partway. A `.gz` header can still be abandoned, because `.gz` uses the compressed size; sweeps skip `.gz` |
| H8 | UUIDv7 file/directory pre-filters can exclude a real child | Rejected. A child is created after its parent, with a 5 s tolerance. The directory end plus 14 h covers offsets from UTC-12 to UTC+14. `archived_sessions`, malformed names and non-v7 names are always read. Identity still comes from `session_meta`, and the nested `parent_thread_id` fallback is preserved in `decodeSessionMetadata` |
| H9 | Unassigned relocation adopts a file without proof, or isolation hides a real failure | Rejected. Requires `s.project`, the `ValidateProject` working directory, store containment, a unique same-UUID or unconventional match, and `keepsAcceptedCaptureParent`. Non-not-found and I/O errors still fail. Budget errors keep their retry. The isolated row keeps its snapshot (Pending plus detail), and the diagnostic attempt is partial |
| H10 | Cancellation between per-owner commit and family annotation strips accepted links | Confirmed. See F1 |

## Findings

### F1: An interrupted pass permanently strips accepted links from the registry snapshot, and the next pass reports them as complete (severity 6)

- **Location:** `backend/internal/adapters/capture/reconcile.go` `refreshTicketOwners`.
  - The per-owner `saveCaptureRead` (around L800-830) commits `current.Snapshot` from a single-transcript extraction, where task-name links are unresolved and so empty.
  - The family `finishCaptureAnnotations` retention runs later, in a separate transaction (around L900-930).
- **Classification:** the window is pre-existing. FIX-21 adds retention only in the annotation step, so the window defeats the new guarantee.
- **Expectation:** the spec (capture-child-link-retention, "Parent refreshed while some children are not read") says every accepted `child_session_id` remains, both in the published ledger and in the capture registry snapshot. AGENTS.md rule 6 says accepted facts are never deleted. Partial coverage stays visible.
- **Trigger:** the pass context is cancelled, or the annotation fails, after the parent's per-owner snapshot commits but before the annotation transaction commits.
  - `reconcileCaptureGenerationAtWithPolicy` runs under `PassTimeoutMS`, so a worker pass timeout on a large store is a realistic trigger. That is the incident shape.
  - Other errors returned between the two steps also trigger it, for example a failed `queueCodexSourceStarts` or cursor write, or a combine/summary error.
- **Consequence:**
  - The parent's registry snapshot loses all task-name links.
  - The next successful partial pass has no accepted link in `previousSnapshots` to retain. Unread children's links are gone from the registry and from `RefreshedSessions`/`Sessions`, and `PendingChildren` is 0. The loss is invisible.
  - Under `--capture-auto` with a committed ledger that already holds the links, `retainCaptureChildLinks` restores them in the published ledger. The registry stays stripped.
  - The published ledger also loses them, with no signal, in two cases:
    - the links were never published (for example, a first publication that is still pending or blocked);
    - the consumer is registry-based (`capture reconcile`, `Sessions`).
  - The CLI helper does not consult pending snapshots (`pendingSessions`), and they fill missing owners only.
- **Evidence:** isolated overlay probe. No product files changed. It reuses `newChildLinkRegistry` and trips a custom context after N `Err`/`Done` checks.
  ```
  go test -overlay <scratch>/overlay.json ./internal/adapters/capture/ -run TestProbeAnnotationWindowStripsRegistryLinks -count=1 -v
  limit=475 err=context canceled registry parent links now ["" "" "" ""]
    after next partial pass: registry=[c1 c2 "" ""] refreshed=[c1 c2 "" ""] pending=0
  ```
- **Suggested direction (parent/coder decides):** apply `RetainAcceptedChildLinks` against `previousSnapshots` when the per-owner snapshot is built, so the intermediate commit never holds fewer links. Alternatively, commit per-owner snapshots and annotations atomically. Then add a cancellation regression test of this shape.

### F2: Accepted links that the rule refuses to carry are dropped without any incompleteness signal (severity 3, suggestion)

- **Location:**
  - `domain/child_links.go`: outcomes `refused_ambiguous`, `refused_held_by_sibling`, and the unknown invocation time path, which makes no decision.
  - `cmd/agentic-lens/capture.go` `retainCaptureChildLinks`, whose stderr line counts only kept and cleared links.
- **Expectation:** refusing ambiguous moves is approved (D1-A). Deleting an accepted fact still ought to be visible.
- **Trigger:** two delegations of one parent share agent type, `task_name` and timestamp string, or an accepted delegation lacks `invoked_at`, and the fresh read leaves the link empty.
- **Consequence:** the published ledger loses the accepted link. `PendingChildren` and the diagnostics do not reflect it, so the attempt can be recorded as completed.
- **Likelihood:** rare. The Codex description is `task_name`, and explicit `session_id` links survive without the child.
- **Suggestion:** count refused accepted links as unresolved child evidence, or mention them in the diagnostic and stderr line.

### F3: Possible double count in `PendingChildren` for a queued child path whose child is not yet registered (severity 2)

- **Location:** `capture/children.go` `pendingChildLinks`. A `pending_children` path maps to a child only through a registered row's `TranscriptPath`.
- **Trigger:** a queued path for an unregistered child, plus a retained link to that same child.
- **Consequence:** the same child is counted twice. This over-reports and stays partial, so no fact is lost. It deviates slightly from "counted once".

## Other observations (no finding)

- The CLI helper records a second reconciliation diagnostic (partial) after reconcile and before the lifecycle check. A `RecordDiagnostic` contention error exits 1, not with the busy exit code. Facts are unaffected. This is consistent with the pending-snapshot diagnostic write.
- `deferred` also triggers on discovery session-claim exhaustion (`ErrDiscoveryBudget` from `claim`) while the split is unreleased. That costs one retry in the leftover refresh and is otherwise harmless.
- Isolated unassigned rows persist as Pending. In a later worker pass that does not attempt them, they count as failures under the existing rule. This is the same behaviour as any unattempted pending row, so it is not new.

## Verification performed

- Read the diff and the unchanged consumers listed in the brief: `resolveSourceChildren`, `mergeSessions`/`fillMissingCaptureSessions` and the `main.go` order, `capture_worker`, the archive regroup, `decodeSessionMetadata`, `readSessionContext` (old preflight), and `pendingChildWork`.
- `go test ./internal/domain/ -run 'ChildLink|RetainAccepted' -count=1`: ok.
- Domain property probes (isolated module under the scratchpad, copy of `child_links.go` with stub types): idempotence, repeat-pass stability, no invention, no duplication and no contradicted child, 200k cases each, 0 failures.
- Overlay cancellation probe for F1. It is reproducible and uses only `t.TempDir` fixtures.
- Package suites were not rerun. Reused `evidence/TESTS.md` per the brief.

## Gaps

- I did not probe the F1 window end-to-end through `generate --capture-auto` (CLI ledger backstop). That conclusion comes from tracing the code.
- I did not test cross-machine pending-snapshot recovery combined with F1.
- I did not exercise `.gz` header admission or OpenCode identification paging.
