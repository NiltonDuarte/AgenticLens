# FIX-21 data-integrity re-review (data-opus-r2)

- Run: data-opus-r2, re_review round 2, focus reviewer-data (data integrity and failure handling)
- Requested profile: Claude Opus 5.5 / medium. The runtime reported model ID claude-opus-5-5. The reviewer cannot see effort, usage or cost.
- Reviewed revision: frozen head 86eb6dae. Corrections delta `git diff 0f73f759 86eb6dae`; full change 133100bb..86eb6dae.
- Verdict: CHANGES_REQUESTED. One new finding, DATA-R2-01 (severity 4), reproduced end to end. Every round-1 finding in this focus is resolved.

## Round-1 findings in focus

| ID | Status at 86eb6dae | Evidence |
|---|---|---|
| DATA-01 | Resolved | `refreshTicketOwners` now applies `RetainAcceptedChildLinks` against `previousSnapshots` before the per-owner commit (reconcile.go L893-897), so the intermediate commit never holds fewer links than before. `TestReconcileInterruptedAnnotationKeepsRegistryLinks` uses a SQL trigger to abort the family annotation, then asserts the registry links and the follow-up pass's count (2). In a discarded or stale read, the carried links stay in the registry uncounted, but that row is Pending, so the pass is still partial. |
| DATA-02 | Resolved, but see DATA-R2-01 | `Report.Refused`, then `UnresolvedLinks()`, feeds the annotation, the regroup and the CLI count. The CLI stderr line reports refused links. Reporting lasts only one pass (DATA-R2-01). |
| DATA-03 | Resolved | `pendingChildLinks` falls back to `codex.RolloutNameSessionID` for a queued path whose child is not yet registered. A misleading name can only merge two counts, never zero them, because the queued path itself still counts, so completion cannot be misreported. |
| DATA-04 | Resolved | `ChildLinkInput.Known = result.Sessions`. Known takes precedence over Accepted for child evidence. A Known record without a parent counts as unobserved, which still allows the carry, so registry evidence cannot block a carry that the ledger allows. |
| BEH-02 / ARCH-03 (busy) | Resolved | There is a single `capture.RetainAcceptedChildLinks`. The registry is opened through `OpenProjectContext` only when `Added > 0`. `IsRegistryContention` walks `errors.Join`, so a contended `RecordDiagnostic` ends with `capturePublicationBusyExit`. |
| BEH-01 (cursor) | Resolved | `last := cursor`. An exhausted sweep that admitted nothing does not write a row (it stays unstarted), and a resumed sweep keeps its cursor. A completed sweep still closes with `''`. |
| ARCH-04 (gz) | Resolved | `.gz` admission requires the full 64 KiB structural bound. |
| ARCH-02 (parity) | Resolved | `readSessionContext` and `RolloutHeader.SessionContext` both call `exactSessionContext` with the same check order: id and cwd required first, then the conflict check. The error text and the `ErrSessionContextMismatch` wrapping are unchanged. |

## New finding

### DATA-R2-01: a key group whose children are only partly re-read loses its unread accepted links permanently, and the incompleteness is reported for one pass only (severity 4, introduced by the approved ordinal rule's refusal branch)

- **Location:** `domain/child_links.go` `reconcileOwnerChildLinks`, the condition `if !empty || len(fresh) < len(accepted)`, and `refusedChildLinks`.
- **Trigger:** a parent has two or more delegations with the same agent type, description and invoked_at, for example parallel `spawn_agent` calls with `agent_id` results and the same timestamp. A bounded pass reads the parent and only some of that group's children. The fresh group is then `[_, C2]`, which is not empty, so the whole group is refused. C1 is reported in `Refused`, and the published owner and the registry snapshot both become `[_, C2]`.
- **Consequence:** on the next pass the accepted baseline (the registry snapshot, or the committed ledger under `--capture-auto`) no longer contains C1. Nothing is retained or refused, and `PendingChildren` returns to 0. The accepted link is deleted and the deletion is invisible from then on. This is the "reported once, then lost" problem that the design's Ordinal group retention steering was approved to fix. The user chose "Manter o grupo" over "Aceitar como limite". It also contradicts the first requirement ("every accepted `child_session_id` remains ... in the published ledger and in the capture registry snapshot"). Partial child reads are the incident shape this change targets, so this branch is not an edge of the feature.
- **Second trigger, two-layer retention:** under `--capture-auto`, the CLI applies ledger retention to `RefreshedSessions` after the registry retention has already carried the links. If the registry snapshot held fewer links than the committed ledger, for example `[C1, C2, _]` against a ledger `[C1, C2, C3]` from another machine with an unknown or shared invoked_at, the fresh group is no longer empty. C3 is then refused, reported once, and lost. The domain probe below shows this.
- **Evidence:**
  - Adapter overlay probe. It copies the fixture from `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses`, but child 2 is small and is read each pass, while child 1 is larger than the partial budget.
    ```
    go test -overlay <scratch>/datar2/overlay2.json ./internal/adapters/capture/ -run TestProbeGroupPartialChildRead -count=1 -v
    full: [group-child-1 group-child-2]
    pass 1 read=map[group-child-2:true group-parent:true] published=["" "group-child-2"] registry=["" "group-child-2"] pending=1
    pass 2 read=map[group-child-2:true group-parent:true] published=["" "group-child-2"] registry=["" "group-child-2"] pending=0
    pass 3 ... pending=0
    ```
  - Domain overlay probe (two-layer case): fresh `[C1 C2 _]` against accepted `[C1 C2 C3]`, all with unknown invoked_at, gives `[C1 C2 _]` with Refused C3. On the next pass, with the published output as the accepted baseline, C3 no longer appears.
- **Suggested direction (planner/coder decides):** treat a group as aligned by position when every fresh link is either empty or equal to the accepted link at the same position. The per-child guards stay as they are. This still invents nothing: each carried child is the accepted child at that position, and a set fresh link equals it. It keeps idempotence, and it covers both triggers. The alternative is to record the residual as a Known limit with explicit user acceptance. That would reverse the stated intent of the steering decision, so it needs the user.
- **Test gap:** no test covers a group in which only some of the children are read.

## Hypotheses rejected

| # | Hypothesis | Result |
|---|---|---|
| H1 | Ordinal retention invents a link or duplicates a child beyond the fresh duplicates | Rejected. Randomized property probe over 300k cases: random keys and timed/untimed mixes, append-only fresh extensions, random fresh links, and confirmed or contradicted observed children. Zero invented links, zero extra duplicates. |
| H2 | Not idempotent, f(f(x,A),A) != f(x,A) | Rejected, 0 failures. The second application sends the group to the refusal branch and emits `refused_ambiguous` decisions, but the sessions are unchanged and `Refused` is empty, so nothing is counted. |
| H3 | Repeated pass with the published output as the accepted baseline drifts, f(x,f(x,A)) != f(x,A) | Rejected for the facts, 0 failures. The reporting does drift (DATA-R2-01). |
| H4 | An accepted, uncontradicted link disappears within one pass without being retained, refused or replaced | Rejected, 0 cases after excluding `replaced`. One residual is not covered by the probe: an accepted key group with no fresh counterpart at all is never iterated, because `order` comes from the fresh owner, so neither a carry nor a report happens. Append-only transcripts make this unreachable unless the fresh read differs structurally, for example a changed invoked_at parse. |
| H5 | Double counting across the annotation, the regroup and the CLI | Rejected. One `countedChildLinks` set is shared. The CLI removes kept links from refused links before printing, and `Added` is deduplicated against the registry count. |
| H6 | The busy exit is missed for a joined error | Rejected. `errors.Join(nil-free)` unwraps to the SQLite cause. A 2 s context expiry exits 1, which is ordinary failure policy. |

## Observation (pre-existing, not a finding of this change)

`resolveSourceChildren` falls back to `ChildReference` (task_name) when the explicit `agent_id` child is unread. With two delegations that share a task name and only one child read, both resolve to that child. Under distinct invoked_at values, the retention rule then sees a confirmed different child and records `replaced`, which moves the accepted C1 to C2. This behaviour predates FIX-21 (manual and earlier capture reads produce the same result). I reached it by reading the code and did not probe it end to end.

## Verification performed

- Read the corrections delta and the head versions of `domain/child_links.go`, `capture/reconcile.go` (refresh, annotation, retention API), `capture/children.go` (sweep cursor, `pendingChildLinks`), `cmd/agentic-lens/capture.go`/`main.go`, the codex context and discovery budget, and `resolveSourceChildren`/`completeOccurrence`.
- Read the interrupted-annotation and ordinal adapter tests.
- Overlay probes, with no tracked files changed. Files are in the scratchpad `datar2/`.
  - domain: `go test -overlay .../overlay.json ./internal/domain/ -run TestProbe -count=1`, which runs the 300k-case property probe and the two-layer and partial-group cases.
  - capture: `go test -overlay .../overlay2.json ./internal/adapters/capture/ -run TestProbeGroupPartialChildRead -count=1`.
- Package suites were not rerun; `evidence/TESTS.md` was reused, as the brief requires.

## Gaps

- I did not run the second trigger (two-layer) end to end through `generate --capture-auto`. That conclusion comes from the domain probe and from reading the CLI code.
- I did not probe cross-machine pending-snapshot recovery, or OpenCode paging.
