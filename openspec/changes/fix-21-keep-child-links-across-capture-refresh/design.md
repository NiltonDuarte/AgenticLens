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

**Pure domain rule.** `domain.RetainAcceptedChildLinks` in `backend/internal/domain/child_links.go` takes a `ChildLinkInput`. Its fields are `Harness` (Codex), `Accepted`, `Fresh` (rewritten), `Observed` (read in the same refresh, never rewritten) and `Known`. `Known` holds registry evidence, the accepted records of another store, and takes precedence over `Accepted` as child-parent evidence. The function returns the reconciled fresh sessions and a typed report.

Child parent evidence has three typed states:
- **confirmed:** the fresh child's `ParentID` is this parent. Without a fresh child, the known or accepted record's `ParentID` is this parent.
- **contradicted:** a fresh child names another parent or none, or a known or accepted record names another parent.
- **unobserved:** there is no record, or the record has no parent.

The rule applies these decisions:
- **Grouping.** For one parent, delegations are grouped by key: agent type, description and invoked-at. An unknown invoked-at is its own key value. Accepted and fresh groups are each taken in transcript order, and accepted position i is paired with fresh position i. A unique key is a group of one.
- **At each position, the unique-delegation rules apply:**
  - an empty fresh link keeps the accepted link;
  - an equal link confirms it;
  - a different child whose parent is confirmed replaces it, unless that child is linked from more than one fresh delegation of the parent, which never confirms a replacement;
  - a different child whose parent is unconfirmed keeps the accepted link;
  - a contradicted child is cleared, and a contradicted accepted child is never kept.
- **Held by another delegation.** A child held by another final delegation of the parent is not kept on a second one. The decision outcome `refused_held_by_sibling` records this, but the child is never reported as a refusal, because by construction it stays linked.
- **Smaller fresh group.** When the fresh group is smaller than the accepted group, as with a truncated transcript, none of its accepted links is placed, and not even the overlapping positions are aligned. Each uncontradicted link is reported as refused, with reason `ambiguous` when the key has a known invoked-at and `unknown_provenance` when it does not. Only this refusal path emits those reasons. Complete groups never emit `ambiguous`, `unknown_provenance` or `unmatched`. They can still emit the decision outcomes `refused_contradicted` and `refused_held_by_sibling`. The refusal reasons that count as unresolved are `ambiguous`, `unknown_provenance`, `unmatched` and `displaced`.
- **No fresh counterpart.** An accepted key with no fresh delegation is reported as refused, with reason `unmatched`, and its link is not placed.
- **Displaced child.** An accepted child displaced by a replacement and left unlinked in the owner's final result is reported as refused, with reason `displaced`. The replacement decision records the displaced child in `ChildLinkDecision.ReplacedChildID`. No refusal is reported for a child still linked elsewhere in that result.

Owners with nil delegations, owners from other harnesses and owners that were not refreshed are untouched, and delegation order is preserved. Applying the rule again changes nothing. Events never receive retained links.

**Invocation.** None of these adds a fallback to `mergeSessions`.

1. **Owner snapshot.** Each owner's first snapshot commit in `refreshTicketOwners` applies the rule against `previousSnapshots`, so a pass that stops before the family annotation never leaves fewer links than it started with.
2. **Family annotation.** `finishCaptureAnnotations` and `regroupArchiveFamily` apply it again to the pass's grouped owners.
3. **Generation.** `--capture-auto` applies it once, through the single entry point `capture.RetainAcceptedChildLinks`, called from a helper in `cmd/agentic-lens/capture.go` that `main.go` calls once, right after reconcile succeeds.
   - It runs against `baseline.Current`, with the registry snapshots (`result.Sessions`) as `Known`, on `RefreshedSessions` and `RestoredSessions`.
   - **Busy classification.** Capture alone decides whether a failure counts as registry contention.
     - When the bounded registry open and diagnostic write in `capture.RetainAcceptedChildLinks` fail after its deadline expires, capture wraps the error with the sentinel `errRegistryWaitExpired`.
     - `capture.IsRegistryContention` recognises that sentinel, as well as the SQLite busy and locked codes.
     - An ordinary expired context without the sentinel is not contention.
     - The CLI's pure helper `childLinkRetentionOutcome` maps the result to the busy exit only through `IsRegistryContention`.
   - It runs before pending sync, merge, restore acknowledgement and publication.

**Coverage.** Each distinct (parent, child) pair adds 1 to `ReconcileResult.PendingChildren`. A pair counts when an accepted link was kept without being re-proven. It also counts when an uncontradicted accepted link was refused: its fresh group is smaller (`ambiguous` or `unknown_provenance`), its key is unmatched (`unmatched`), or a replacement displaced it (`displaced`). A child still linked elsewhere in the owner's final result, including one held by a sibling delegation, is not reported. A refused link is never moved. No new counter or field is added.
- **Deduplication.** A pair is not added again when one of these already counts it:
  - a `pending_starts` row with the same store, the same parent and `agent` equal to the child ID;
  - a `pending_children` row for the same parent whose path equals the child's registered transcript path. For a child not yet registered, the session ID claimed by the file name (`codex.RolloutNameSessionID`) is used only as a deduplication hint for counting, never as identity;
  - an earlier retained pair in the same pass.
- **When the count is added.**
  - The adapter adds its pairs after `pendingChildWork` sets the figure.
  - In generation, `capture.RetainAcceptedChildLinks` adds its pairs under the same deduplication. When the count grows, it records the request's reconciliation again as partial through the existing diagnostic writer. Otherwise a machine whose registry never registered the children would show the pass as complete.
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
- **Children that can never be re-read.** On a machine that can never read the children again, coverage stays partial. The "continue reconciliation" hint cannot resolve it there.
- **Coverage reported as partial too often.** When the parent and a child are read in different rounds of the same pass, coverage can be reported as partial even though both were read.
- **Very small `max_source_bytes`.** A very small value can stall identification. It shows as pending.
- **OpenCode child pages.** They still consume session slots. This predates FIX-21.
- **Refused and unmatched links are reported only once.** They are counted when refused, but they are not placed on any delegation afterwards, so a later pass that cannot re-read them has nothing left to report.
- **R2-P1, pre-existing.** When two delegations share a task name and only one child is read, `resolveSourceChildren` points both at that child. FIX-21 did not introduce this. Since R3-01, a duplicated fresh child never confirms a replacement, so FIX-21 keeps the accepted links in that case. The resolver flaw itself remains for a follow-up ticket.
- **ARCH-01, accepted debt.** The capture exact lookup duplicates the manual and live lookup rules. A follow-up should unify them.

### Owners

The coder owns these product files:
- `domain/child_links.go`
- `capture/reconcile.go`, `capture/archive_projection.go`, `capture/children.go`, `capture/source_database.go` and `capture/codex_context.go`
- `codex/discovery.go`, `codex/discovery_budget.go` and `codex/context.go`
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

The discovery-budget spec and tasks were aligned with these. Verification adds assertions for deferral and retry, the attempts figure, year and month directories at the boundary, and refusal to admit a read.

**Spec wording alignment.** On 2026-11-24 the tester reported two places where the spec's wording contradicted the accepted design, and the orchestrator settled both. Neither changes behavior or scope.
- **Contradiction scenarios.** They now promise that the link is cleared and that existing source-level warning behavior is unchanged. They no longer promise a capture-visible warning, and the limit is recorded under Known limits.
- **File names.** The requirement "File names never establish identity" became "File names only exclude candidates", plus a separate requirement that unconventional names are read. These permit the exact-lookup and date-directory exclusions, while inclusion still requires the candidate's own `session_meta`. The residual risk from renamed rollouts is recorded under Known limits.

Verification asserts that the link is cleared, not that a warning appears, and covers both permitted exclusions.

**Review DATA-02.** Reported by reviewer-data and reviewer-behavior in review round 1. The orchestrator decided it to restore the user's approved acceptance criterion 3, with no new scope. The "Ambiguous delegation provenance" scenario had lost "and the pass reports them as unresolved".
- **Change.** An accepted link that the rule refuses to carry now counts as unresolved child evidence, deduplicated like retained pairs, which keeps the pass partial. This applies to an ambiguous key, an unknown invoked_at, or a child held by a sibling, when the accepted link existed, the fresh link is empty and there is no contradiction. The link is never moved.
- **Artifacts.** The scenario's THEN clause is restored, and the coverage requirement now names refused links. The other accepted review limits are listed under Known limits.
- **Verification.** It adds a refused-link case that asserts the count, the deduplication and the partial attempt.

**Ordinal group retention.** The user approved this on 2026-11-24 in the main conversation, choosing "Manter o grupo" over "Aceitar como limite". It refines D1 for ambiguous groups.
- **Problem.** The first rule refused every delegation in a group that shared agent type, description and invoked-at (or had an unknown invoked-at) when its children were not read. The pass reported them once as unresolved, but the published ledger then lost those links, so later passes had nothing to report.
- **Decision.** Accepted links are now carried by ordinal position within the key group when every fresh link in the group is empty and the fresh group is at least as large. The per-child guards and the unresolved-evidence count apply to each carried pair. The previous refusal and reporting remain when any fresh link in the group is set or the fresh group is smaller. No link is invented, and the rule stays idempotent.
- **Superseded.** This replaces D1-A's refusal of ambiguous keys and of unknown invoked-at. DATA-02's reporting now applies only to the refusal branch.
- **Artifacts.** The spec adds the "Retain ambiguous groups by ordinal position" requirement, with scenarios for ordinal retention, refusal and repetition.
- **Verification.** It adds these cases:
  - an ordinal carry for duplicate keys and for unknown invoked-at, with the published ledger keeping the links across repeated passes;
  - refusal and reporting when a fresh link is set or the fresh group is smaller;
  - a contradicted or already-held child skipped within a group;
  - idempotence.

**Position alignment.** The user approved this on 2026-11-25 in the main conversation, choosing "Alinhar por posicao" over "Manter como esta", after review round 2 (R2-01, R2-02, R2-03; see [evidence/REVIEW.md](evidence/REVIEW.md)).
- **Problem.** A group that was only partly re-read was refused as a whole, so its unread links were lost from the registry and the ledger.
- **Superseded.** This replaces the all-empty precondition of the 2026-11-24 "Ordinal group retention" entry.
- **Decision.**
  - Accepted position i is now paired with fresh position i within each key group, applying the unique-delegation rules at each position. A unique key is a group of one.
  - A group is refused and reported only when the fresh group is smaller.
  - An accepted key with no fresh counterpart is reported as refused, with an unmatched reason, and counted as unresolved.
- **Also fixed.** R2-02: an untimed unique delegation no longer accepts an unconfirmed different child.
- **Artifacts.** The spec's ordinal requirement became "Align delegation key groups by position", with a separate requirement for unmatched keys. The design now matches the implementation (R2-07): the `capture.RetainAcceptedChildLinks` entry point, per-owner application, `Known` and the file-name deduplication hint. The one-time reporting of refused links and R2-P1 are recorded under Known limits.
- **Verification.** It adds these cases:
  - ambiguous groups with all positions unread;
  - a partly re-read group;
  - interleaved and offset groups;
  - a smaller fresh group;
  - an unmatched key;
  - an untimed unique delegation against an unconfirmed child;
  - CLI alignment end to end.

**R2 implementation sync.** The coder's report on 2026-11-25 removed two values that could no longer occur: the evidence value `ChildParentNotEvaluated` and the outcome `ChildLinkRefusedAmbiguous`. Complete groups emit no `ambiguous`, `unknown_provenance` or `unmatched` refusals; R3-05 narrowed this claim. Child parent evidence therefore has three states. A smaller fresh group places none of its accepted links, not even the overlapping prefix, and reports them as `ambiguous` (known invoked-at) or `unknown_provenance`. A key with no fresh counterpart is reported as `unmatched`. There is no change to scope or approval.

**Review round 3 corrections.** The orchestrator made these within the approved acceptance: AC1 and AC3, no silent loss of accepted facts. The source is review round 3 (R3-01, R3-02, R3-05; see [evidence/REVIEW.md](evidence/REVIEW.md)).
- **R3-01.** A replacement could silently drop an accepted child: accepted `[A B]` with fresh `[B B]`, or accepted `[A B]` with fresh `["" A]`.
  - A fresh child linked from more than one fresh delegation of the parent never confirms a replacement.
  - An accepted child displaced by a replacement and left unlinked in the owner's final result is reported as refused, with reason `displaced`, so it counts as unresolved.
  - The spec adds scenarios for the duplicated child and the displaced child, and the refusal causes now include displacement.
  - Verification adds both reproductions.
- **R3-02.** Busy classification now has a single owner in capture. `capture.RetainAcceptedChildLinks` classifies registry contention, and the CLI uses that one predicate.
- **R3-05.** Stale text is corrected:
  - the coverage requirement now lists every refusal cause, plus the exemption for a child linked elsewhere;
  - "complete groups" is narrowed;
  - the Owners list gains `codex/context.go`;
  - the R2-P1 limit is updated;
  - `tasks.md` marks superseded text instead of deleting it.

**R3 implementation sync.** The coder's report on 2026-11-25 recorded these changes. There is no change to scope or approval.
- **Held-by-sibling children.** They are never reported as refused, because another final delegation holds them by construction. The refusal reason `ChildLinkRefusalHeldBySibling` was removed, and the decision outcome `refused_held_by_sibling` remains. The spec's refusal causes and the coverage text no longer list it, and the Known limit about it counting as unresolved is withdrawn.
- **Busy classification.** R3-02 is implemented through the `errRegistryWaitExpired` sentinel, `IsRegistryContention` and the `childLinkRetentionOutcome` helper.
- **Displaced child.** `ChildLinkDecision.ReplacedChildID` records it.
