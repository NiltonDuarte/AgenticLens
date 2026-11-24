# FIX-21: Technical design

## Scope and boundaries

This change has three task groups. Each has its own acceptance spec:

1. Retain accepted Codex delegation child links through partial automatic capture refresh ([spec](specs/capture-child-link-retention/spec.md)).
2. Fit a ticket family's discovery and extraction within the capture budget ([spec](specs/capture-discovery-budget/spec.md)).
3. Recover, or isolate, an unassigned rollout that moved to `archived_sessions/` ([spec](specs/capture-archived-unassigned-recovery/spec.md)).

Intent is in the [proposal](proposal.md).

The following are excluded:
- manual `generate` and explicit backfill behavior (open PR #162, AL-94)
- OpenCode and Claude Code link resolution
- source parsing
- relocation for bound rows, which already exists
- any ledger or registry schema change

## Decisions and ownership

### Link retention

**Root cause.** `resolveSourceChildren` resolves a task-name result only against transcripts loaded in the same `CombineTranscriptsWithPolicy` call. Three places then overwrite or replace the parent's delegations:
- `finishCaptureAnnotations` (`reconcile.go`) overwrites the refreshed owner's registry snapshot delegations with those results.
- `regroupArchiveFamily` (`archive_projection.go`) does the same.
- `mergeSessions` replaces the ledger owner wholesale.

The stored `DelegationFact` has no call ID.

**Pure domain rule.** `domain.RetainAcceptedChildLinks` in `backend/internal/domain/child_links.go` takes the accepted sessions, the fresh sessions and the target harness (Codex). It returns the reconciled fresh sessions and a typed report.

Child parent evidence has three typed states:
- **confirmed:** the fresh child's `ParentID` is this parent. Without a fresh child, the accepted record's `ParentID` is this parent.
- **contradicted:** a fresh child names another parent or none, or an accepted record names another parent.
- **unobserved:** there is no record, or the accepted record has no parent.

The rule applies these decisions:
- A fresh link to a contradicted child is cleared.
- A different fresh child replaces the accepted link only when that child is confirmed.
- An accepted link is carried onto an empty fresh delegation only when all of these hold:
  - the agent type, description and invoked-at key is unique on both sides;
  - the invoked-at time is known;
  - the child is not contradicted;
  - no sibling delegation already holds that child.

Owners with nil delegations, owners from other harnesses and owners that were not refreshed are untouched, and delegation order is preserved. Applying the rule again changes nothing. Events never receive retained links.

**Invocation.** The rule runs in two places only. Neither adds a fallback to `mergeSessions`.

1. **Capture adapter.** `finishCaptureAnnotations` and `regroupArchiveFamily` apply it against the pass's `previousSnapshots`.
2. **Generation.** `--capture-auto` applies it once, through a helper in `cmd/agentic-lens/capture.go` that `main.go` calls once, right after reconcile succeeds.
   - It runs against `baseline.Current`, on `RefreshedSessions` and `RestoredSessions`.
   - It runs before pending sync, merge, restore acknowledgement and publication.

**Coverage.** Each distinct retained (parent, child) pair adds 1 to `ReconcileResult.PendingChildren`. No new counter or field is added.
- **Deduplication.** A pair is not added again when one of these already counts it:
  - a `pending_starts` row with the same store, the same parent and `agent` equal to the child ID;
  - a `pending_children` row for the same parent whose path equals the child's registered transcript path;
  - an earlier retained pair in the same pass.
- **When the count is added.**
  - The adapter adds its pairs after `pendingChildWork` sets the figure.
  - The CLI helper adds its pairs through the deduplicating `ReconcileResult.AddRetainedChildLinks`, and records the updated partial reconciliation (from `ReconcileResult.ReconciliationDiagnostic`) through the existing diagnostic writer. Otherwise a machine whose registry never registered the children would show the pass as complete.
- **What does not change.** `UnresolvedParentLinks` is never changed. A non-zero `PendingChildren` already marks the attempt as partial.

### Discovery budget

**Measured cause.** `extractExactAt` refuses a rollout that is larger than the remaining budget. Header identification spends that same pool:
- `readMetadata` (`capture/children.go`) charges a flat `min(size, 64 KiB)` per file.
- `inspectRollout` and `listRolloutPaths` (`codex/discovery.go`) claim and read every file in the store for each exact lookup.
- Headers are never reused within a pass.

A sweep of 1,909 headers is estimated at about 120 MiB, and several sweeps happen per pass.

**Changes.** Capture paths only:
- **Window pre-filter (child sweeps).**
  - **Clock evidence.** Rollout directory and file-name times are machine-local; on the user's store, CET and CEST appear. Codex file-name UUIDs are UUIDv7, equal to `session_meta.id`, and their first 48 bits are the exact UTC creation milliseconds.
  - **File rule.** A candidate whose UUIDv7 instant is more than 5 seconds before the parent's instant is skipped without being claimed or read. The 5 seconds cover NTP clock steps only; the smallest observed parent-to-child gap was 6.7 s across 949 relations.
  - **Parent instant.** It comes from the parent's own UUIDv7, or else from its `session_meta` timestamp. Without either, that sweep does not prune.
  - **Directory rule.** It only reduces listing cost. A local year, month or day directory is skipped only when its local period ends more than 14 hours (the maximum UTC offset) plus the 5-second tolerance before the parent's instant.
  - **Always read.** Outside a pruned date directory, a child sweep reads malformed names and non-v7 UUIDs as it does today. Names and directories only exclude candidates: inclusion, identity and parent always come from the candidate's own `session_meta`.
- **Exact-lookup pre-filter.** `codex.DiscoverExactCaptureRolloutPath` skips conventionally named files whose UUID differs from the request, v7 or not. Same-UUID copies are still read, and duplicates are reported with `codex.ErrOwnerAmbiguous`.
- **Header cache.** One header reader with a pass-local `codex.HeaderCache`, keyed by physical path, size and mtime, serves child sweeps, owner preflight and exact lookup.
- **Actual-bytes charging.** Header reads are charged the bytes actually consumed, measured as the current discovery reader does, within the 64 KiB bound. A header read starts only when `min(size, 64 KiB)` fits in the remaining allowance, so no read is abandoned partway.
- **Reserved identification ceiling.**
  - **Rotating rounds.** During the work-class loop in `ReconcileAt`, header identification may use at most a quarter of `max_source_bytes`. That quarter is reserved, so owner extraction in those rounds may use at most three quarters.
  - **Final leftover refresh.** Extraction may use everything that remains of the total, including identification's unused share.
  - **What counts as identification.** Only child discovery: Codex child sweeps, queued-child header reads, the parent header read for the window, and OpenCode child pages. A known owner's own preflight and exact lookup are charged to extraction, and they cost nothing once the header cache holds them.
  - **No borrowing by identification.** It stops at its ceiling with discovery pending, not failed, and never draws on extraction. Header reads no longer reduce the extraction session count.
  - **Deferred owners.** An owner refused only because of the reservation is marked not attempted. Its session slot is refunded, and the final leftover refresh retries it.
  - **Totals.** The two never exceed `max_source_bytes` combined, and no configuration key is added. Reported `SourceBytes` is their sum. `SourceAttempts` (the coverage summary's "Source attempts") counts owner reads only, not header reads, so attempts never exceed `max_sessions`.

Manual `generate` and backfill use `DiscoverRolloutPaths` and stay unchanged. That path runs without a pass budget, and a nonce scan reads whole files before the root is known. Pruning there would add risk of losing a member without any correctness benefit. Live hydration keeps the current exact lookup, and the pre-filters are opt-in options that only capture passes.

### Archived unassigned recovery

**Root cause.** The incident log's lstat failure comes from the unassigned pass (`Reconcile` with an empty ticket, `capture_worker.go`). `inspectRegisteredCodexPath` (`source_database.go`) returns the `EvalSymlinks` error when `TicketID` is empty, so the existing relocation never runs. The worker then counts the failed unassigned row as a pass failure (`capture_worker.go`, around line 411) and retries with backoff until its lifetime runs out.

**What the guard protects.** All of these remain:
- no ticket or binding is inferred for an unbound source;
- no file outside the registered store is adopted;
- the session's identity comes from its own unique `session_meta`;
- project proof is required;
- the accepted parent is kept;
- errors other than not-found still fail.

**Changes.**
- **(A) Relocation on exact proof.** Relocation is allowed for unassigned rows under the same checks binding already applies (`verifyBindingSource`): `resolveRegisteredCodexSnapshot`, `validateCodexOwnerMetadata` and `keepsAcceptedCaptureParent`. The row stays unassigned.
- **(B) Isolation.** An unproven, ambiguous or conflicting outcome records the row's state and diagnostic. It adds the row to the request-local `ReconcileResult.IsolatedSources` as an `IsolatedSource{Identity, Reason, Detail}`, with a typed reason (unavailable, ambiguous or conflicting). The worker excludes that list from pass success and from `lastFailure`. Budget outcomes keep their existing retry behavior.

### Known limits

- **Contradiction warnings.** Capture never surfaced the "delegation child metadata contradicts native parent" warning, and this change keeps it that way.
  - A task-name-only result raises no warning, because `resolveSourceChildren` never matches a child whose parent differs.
  - For an explicit `session_id` result, the warning stays in the owner's source warnings and does not reach capture output.
  - The guarantee FIX-21 makes is that a contradicted link is cleared.
- **Renamed rollouts.** An exact lookup would miss a renamed rollout whose file name carries a UUID other than its `session_meta.id`. On the user's real store, the file-name UUID equalled `session_meta.id` in 1909 of 1909 files.

### Owners

The coder owns these product files:
- `domain/child_links.go`
- `capture/reconcile.go`, `capture/archive_projection.go`, `capture/children.go`, `capture/source_database.go` and `capture/codex_context.go`
- `codex/discovery.go` and `codex/discovery_budget.go`
- `cmd/agentic-lens/capture.go`, `capture_worker.go` and the single call in `main.go`

The tester owns the new and updated test files named in [tasks](tasks.md). The orchestrator owns integration, evidence, telemetry and the PR.

**Overlap with PR #162.** Two overlaps are expected:
- In `main.go`, only one call.
- Conceptually with backfill's `source_budget.go`. This change does not edit backfill.

## Implementation approval

Approved by the user on 2026-11-24, in the main development conversation. The approval covers all three groups: link retention, the discovery budget and archived relocation. Its basis is decisions D1 to D7 and the final D6 rules: the UUIDv7 window with a 5-second tolerance, and a reserved quarter for identification whose unused share is released to the final leftover refresh. The user asked whether the unused identification share could go to extraction. Shown this design in answer, they replied "sim". Any material change to these boundaries needs renewed approval.

## Implementation approach

**Regression boundaries** from `docs/REGRESSION_PREVENTION.md`:
- **Telemetry extraction and refresh:** root with children, partial read, unchanged repeat, unrelated harness, failed reads preserving facts.
- **Persisted facts:** accepted facts stay visible while children are pending.
- **Portable ledgers across machines:** B refreshes A's committed ledger with empty capture state, and a complete local refresh still wins.
- **Polling and asynchronous updates:** worker pass cost grows with the family window, not the store, and an unrecoverable source does not cause retry loops.

**Existing coverage owners** (`codex/`, `capture/`, `cmd/agentic-lens/`):
- `codex/rollout_test.go`: task-name resolution.
- `capture/reconcile_nested_codex_test.go`: contradiction.
- `capture/reconcile_test.go`: budget accounting, unassigned isolation and the coverage summary.
- `capture/reconcile_archived_rollout_test.go`: bound relocation and binding of archived unbound rows.
- `capture/fix19_budget_test.go` and `capture/fix19_candidate_budget_internal_test.go`.
- `codex/discovery_test.go`, `codex/discovery_budget_test.go` and `codex/fix19_discovery_test.go`.
- `owned_refresh_regression_test.go` and `capture_history_test.go`: retention across machines.
- `ledger_timestamps_test.go`: FIX-20.
- `main_codex_test.go`: manual linking.
- `capture_corrections_test.go`: worker lifetime.

Budget-accounting assertions that change are updated with a recorded rationale, not deleted. Fixtures are self-contained and never touch a live ledger or `~/.codex`. Exact byte accounting against known fixture sizes is the byte oracle. File permissions are not used, because a root user would bypass them.

**Review** covers:
- data integrity: no invented links, idempotence, attribution;
- behavior through the worker and `--capture-auto`;
- architecture: pure domain rule, minimal CLI diff;
- test strength, including targeted mutations that remove the carry-forward, the window pre-filter or the isolation.

## Steering

**Decisions D1 to D7.** The user settled them in the main conversation on 2026-11-24, relayed by the coordinator. They revise the first draft as follows.

- **Retained as drafted:**
  - D1-A: match on agent type, description and invoked-at time, refuse ambiguous or unknown cases, and make no call-ID schema change.
  - D2: a different child replaces a link only when confirmed.
  - D3-A: apply the rule in the registry and once in `--capture-auto`.
  - D4: events receive no retained links.
  - Codex only.
- **D5, changed.** The draft's new retained-link count and field are dropped. Retained links now feed `PendingChildren` under the deduplication rule above, and never `UnresolvedParentLinks`. The coverage requirement and its tests were rewritten. The CLI case also records the partial diagnostic.
- **D6, now in scope** (option B plus refinements 2 to 4). This supersedes the draft's split and adds the discovery-budget capability, its task group, a measurement task and large-store tests. Manual discovery stays unchanged, by decision.
- **D7, now in scope** (A plus B). This supersedes the draft's split and adds the archived-recovery capability and tests for the unassigned pass and worker isolation.

**D6 refinements.** The user settled the two remaining D6 questions on 2026-11-24, relayed by the coordinator, after the orchestrator's store analysis of 1,909 rollouts.
- **Time window.** UUIDv7 creation instants with a 5-second tolerance replace the draft's local-time file-name window and its 24-hour margin. Day directories are skipped only beyond the 14-hour maximum UTC offset. The spec now has separate requirements for the per-file rule, the directory rule and the exact lookup. Tests add a child within tolerance, malformed and non-v7 names, and a CET/CEST fixture whose local date differs from its UTC date.
- **Allowance.** A reserved quarter, released to the final leftover refresh, replaces the draft's fixed split without refunds. Tests add extraction scheduled ahead of discovery, the release of the leftover share, identification stopping at its ceiling as pending, and exact byte totals.

No open decisions remain.

**Implementation refinements.** The coder's implementation report proposed five refinements, and the orchestrator accepted them on 2026-11-24 as within the approved scope. None needs renewed approval.
1. **Identification covers child discovery only.** A known owner's preflight and exact lookup are charged to extraction. This is what lets known owners still be extracted when identification hits its ceiling.
2. **`SourceAttempts` excludes header reads.** This keeps attempts at or below `max_sessions`. Header bytes still count in `SourceBytes`. The coverage summary's "Source attempts" figure changes as a result.
3. **Owners refused by the reservation are deferred.** Each is marked not attempted, its session slot is refunded, and the final leftover refresh retries it.
4. **Stricter directory rule.** It uses 14 hours plus 5 seconds and also applies to year and month directories. The rule skips less.
5. **Header read admission.** A header read starts only when `min(size, 64 KiB)` fits in the remaining allowance.

The discovery-budget spec and tasks were aligned with these.

**Spec wording alignment.** On 2026-11-24 the tester reported two places where the spec's wording contradicted the accepted design, and the orchestrator settled both. Neither changes behavior or scope.
- **Contradiction scenarios.** They now promise that the link is cleared and that existing source-level warning behavior is unchanged. They no longer promise a capture-visible warning, and the limit is recorded under Known limits.
- **File names.** The requirement "File names never establish identity" became "File names only exclude candidates", plus a separate requirement that unconventional names are read. These permit the exact-lookup and date-directory exclusions, while inclusion still requires the candidate's own `session_meta`. The residual risk from renamed rollouts is recorded under Known limits.

Verification asserts that the link is cleared, not that a warning appears, and covers both permitted exclusions. Verification adds assertions for deferral and retry, the attempts figure, year and month directories at the boundary, and refusal to admit a read.
