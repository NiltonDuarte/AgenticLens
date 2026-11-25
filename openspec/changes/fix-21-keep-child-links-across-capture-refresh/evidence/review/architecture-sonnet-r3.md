# FIX-21 architecture re-review, run arch-sonnet-r3

- Focus: architecture and maintainability (reviewer-architecture), phase re_review round 3.
- Requested profile: Claude Sonnet 5.5 / high. Runtime model ID claude-sonnet-5-5; effort not observable.
- Reviewed revision: head 5d4df7ee (change 133100bb..5d4df7ee; corrections delta 86eb6dae..5d4df7ee).
- Scope: the rewritten position-alignment rule in `domain/child_links.go`, typed outcomes and reasons, removed dead states, comment and contract accuracy, `capture.RetainAcceptedChildLinks` and the CLI helper, design, spec and tasks consistency with code (R2-07).
- Independence: no `*-r3.md` sibling report was read. Round-2 reports and REVIEW.md were read as the re-review baseline.

## Verdict: APPROVED (scoped)

No blocking architectural defect. The alignment rule has one path, typed outcomes and reasons, and no dead states. Six non-blocking findings (ARCH-R3-01..06), all documentation or naming except ARCH-R3-02, which is a small design inconsistency in the busy-exit mapping.

## Round-2 verification (architecture-relevant items)

| ID | Status at 5d4df7ee | Evidence |
|---|---|---|
| R2-01 / R2-02 / R2-03 | Resolved in the rule | `reconcileOwnerChildLinks` has one alignment path: group by key, `switch` on fresh versus accepted size (`unmatched`, `ambiguous`, `unknown_provenance`), then pair position i with position i and apply the unique-delegation rules. The three-way `alignment` enum is gone; `alignmentUnique` (assigned, never read) no longer exists. An untimed unique delegation against an unconfirmed different child goes through the same `movable` path and yields `kept_pending_confirmation`. |
| R2-04 | Resolved | `ChildLinkRefusedAmbiguous` and `ChildParentNotEvaluated` are removed. `grep` finds no leftovers in code, tests, spec or design (design names them only in the "R2 implementation sync" note). The `ChildLinkRefusal` reasons are emitted only on `Refused`, never on `Decisions`. A smaller group and a no-counterpart key produce no decisions. |
| R2-05 | Resolved | `cmd/agentic-lens/capture.go` appends "so the reconciliation remains partial" only when `retention.Added > 0`; `Added` now has a production consumer, which closes the round-2 optional note about the exported field. |
| R2-06 | Resolved in behavior, see ARCH-R3-02 | `errors.Is(err, context.DeadlineExceeded)` joins `IsRegistryContention` in the busy-exit branch. |
| R2-07 | Mostly resolved | design.md now names `capture.RetainAcceptedChildLinks`, the per-owner application at the first snapshot commit, `Known`, the file-name deduplication hint, three evidence states and the typed refusal reasons. Residual gaps are ARCH-R3-01 and ARCH-R3-03. |
| ARCH-R2-04 (typed key inconsistency) | Resolved | A unique key is a group of one, so timed and untimed unique delegations follow the same rules; the design states it. |

## Findings

### ARCH-R3-01 (sev 2, introduced, spec accuracy)

**The coverage requirement describes only one of the four refusal paths.** `specs/capture-child-link-retention/spec.md`, requirement "Report retained and refused links as unresolved child evidence", says a pass counts "each accepted link it refused to carry onto an empty, uncontradicted fresh delegation". Under position alignment a link is also refused when its fresh group is smaller (the fresh delegations at those positions need not be empty) or when its key has no fresh delegation at all (there is no delegation to carry it onto). The requirement's own scenario "Accepted link refused" already lists those cases, so the requirement text and its scenario disagree. Neither the spec nor the design states the exemption that `refusedChildLinks` applies: an unplaced (smaller-group or unmatched) link is reported only while no final delegation links its child ("an unmatched key whose child is linked elsewhere is not reported", covered by a test). The design's "No fresh counterpart" bullet is unconditional.

Consequence: a reader of the spec cannot predict the counts for the smaller-group and unmatched cases, and a later change could "fix" the exemption as a bug or drop it as unspecified. Fix: reword the requirement to "each accepted link it retained without re-proving, or refused to place" and add the linked-elsewhere condition to the design bullets and the unmatched requirement.

### ARCH-R3-02 (sev 2, introduced, design consistency)

**The busy exit has two definitions, and the new one is undocumented against the central policy.** `capture.IsRegistryContention` documents that "a joined permanent error or expired context retains ordinary failure policy" and every other caller relies on that. `retainCaptureChildLinks` now adds `|| errors.Is(err, context.DeadlineExceeded)` locally, with the comment "The bounded wait expires only while the registry is held elsewhere." The comment is stronger than the code can establish: the 2 s context also bounds `RecordDiagnostic` after the open, so a deadline can expire for a slow open or write that is not contention, and it then exits as busy (worker retry) instead of failing. The user decision (R2-06) approves the busy mapping, so the behavior is in scope. What is missing is one owner for the rule: the capture package keeps the classification, and the CLI now carries an exception to it. R2-06 also has no test seam (TESTS.md says "None. See Limits."), so the mapping can regress silently.

Fix (optional): move the decision into capture, for example a helper that treats deadline expiry on a context the caller bounded as contention, or soften the comment to "treated as contention", and state the exception in the `IsRegistryContention` doc. A table test of the exit code would be the tests reviewer's call.

### ARCH-R3-03 (sev 1, introduced, design and tasks currency)

**Residual stale text in the owning records.**
- `design.md` "Owners" still omits `codex/context.go`, which ARCH-02 changed (`exactSessionContext`, `conflictingNativeParent`). It was reported in round 2 (ARCH-R2-01) but is not in the R2-07 consolidated list.
- `design.md` "R2 implementation sync" says "Complete groups emit no refusal decisions". The code still emits `refused_contradicted` and `refused_held_by_sibling` decisions for complete groups. The intended statement is that no ambiguity decision is emitted for a complete group. Tighten the wording.
- `tasks.md` DATA-02 line lists the typed reasons "(ambiguous, unknown provenance, held by a sibling)" without `unmatched`; the "ordinal group retention" task still describes refusal when "a fresh link in the group is set", which position alignment superseded; the review task still names candidate `133100bb..0f73f759`. The superseded entries are history, but a task list is meant to show current work.

### ARCH-R3-04 (sev 1, introduced, evidence currency)

`evidence/TESTS.md` "Revision and working tree" still reads "commit `86eb6dae`, plus the coder's uncommitted round 2 changes and the tester's uncommitted tests". The round-2 work is now committed as 5d4df7ee. The "Ordinal group retention" row in the round-1 table also still describes the superseded all-empty rule ("a group with one fresh link set ... is refused"). Update the revision to 5d4df7ee and mark the row superseded by the round-2 table, so the record does not contradict the rule it validates.

### ARCH-R3-05 (sev 1, introduced, naming)

`ReconcileResult.retainedChildLinks` (`reconcile.go` line 39), the local `retained` in `finishCaptureAnnotations` and the archive regroup now hold `Report.UnresolvedLinks()`, which is retained plus refused links. The name predates DATA-02 and now misleads: a reader would expect only kept links, while the field feeds the unresolved count. Rename to `unresolvedChildLinks`, or document the field. No behavior change.

### ARCH-R3-06 (sev 0 to 1, optional)

- `ChildLinkDecision` field meaning differs by outcome and is undocumented: for `replaced`, `ChildSessionID` carries the fresh child that won (equal to `FreshChildID`), not the accepted child that was dropped; for `cleared` it is empty. The type comment says only "one non-trivial decision". The replaced accepted child is not recorded anywhere. No consumer reads it today (`RetainedLinks` skips `replaced`), so this is a documentation note.
- The `RetainAcceptedChildLinks` doc comment has a 112-column line (line 165); rewrap it.

## Maintainability assessment of the rewritten rule

- One alignment path. `reconcileOwnerChildLinks` has a single `switch` for the three group sizes and one loop for pairing, then the held-child fixpoint, then outcome assignment. `grouped` only chooses between the `retained` and `retained_by_position` labels, and the documented meaning of both labels matches the code (`!key.timed || len(fresh) > 1 || len(accepted) > 1`).
- Typed state. `ChildParentEvidence` has three values, `ChildLinkOutcome` seven, `ChildLinkRefusal` four, all with comments that match their emission sites. No `any` or `interface{}`; `gofmt` is clean on domain, cmd, capture and codex, and `go vet ./internal/domain` is clean.
- Layering. The rule stays pure (no I/O, inputs copied), the capture adapter orchestrates counting and the diagnostic, and `main.go` still makes one call.

## Verification performed

- Read in full: `domain/child_links.go`; `capture/reconcile.go` retention paths (`RetainAcceptedChildLinks`, `addUnresolvedChildLinks`, `reconciliationDiagnostic`, first snapshot commit, `finishCaptureAnnotations`, archive regroup); the CLI helper and its delta against 86eb6dae; `IsRegistryContention`; design.md, spec, tasks.md, proposal.md and the TESTS.md round-2 sections.
- Focused run on the real tree: `go test ./internal/domain -run 'RetainAcceptedChildLinks|ChildLink' -count=1` passed. `gofmt -l` and `go vet ./internal/domain` clean.
- Isolated randomized probe on a scratch copy of the domain package (`/tmp/claude-1000/-home-niltonduarte--t3-worktrees-AgenticLens-t3-ab98b16e-74530a6275012cd8/a27e309f-c6cf-4b8e-8a78-c7568c3fffdd/scratchpad/arch-sonnet-r3/mod`): 200,000 random accepted and fresh owners (mixed key sharing, timed and untimed, random registry evidence, links drawn from a five-child pool). Checked on every case:
  - applying the rule to its own result changes nothing;
  - taking the result as the accepted ledger and re-applying against the same fresh read is stable;
  - no link is invented and no child is duplicated beyond the fresh read;
  - delegation count and order are preserved;
  - every uncontradicted accepted child either stays linked, is reported in `Refused`, or is displaced by a confirmed replacement or a contradiction (no silent loss).
  All checks held, with no violation. (A nil fresh delegation list is skipped by design and is not a counterexample; an empty non-nil list reports every accepted link as unmatched.)
- `grep` for removed names (`ChildParentNotEvaluated`, `refused_ambiguous`, `AddRetainedChildLinks`, `ReconciliationDiagnostic` as a method) confirms no stale references in code, tests, spec or design.
- Reused, not rerun: whole-package, race, large-store and mutation results from `evidence/TESTS.md`.

## Gaps

- I did not exercise the CLI busy-exit mapping or the stderr wording end to end; R2-06 has no seam, and I checked the mapping by reading the code.
- Capture-level behavior of position alignment through a worker pass was not re-probed; the domain probe covers the rule, and the adapter wiring passes `previousSnapshots` and `group.previous` as before.
- The first-commit retention does not add its links to the unresolved count; the annotation does. I found no path where a refreshed Codex owner skips the annotation, but did not prove it exhaustively (carried over from round 2).
