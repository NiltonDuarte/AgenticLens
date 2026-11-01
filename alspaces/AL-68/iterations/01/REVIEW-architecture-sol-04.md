# AL-68 architecture scheduler re-review

Scoped code verdict: **APPROVED**. Overall acceptance outcome: **INCOMPLETE** because required real Codex trust/callback/agent-interaction evidence remains unavailable. Parent final integrated `make` is also outstanding. No substantiated blocking code defect remains in this bounded review.

## Revision and boundary

Run `architecture-sol-04`, final bounded corrective review. Requested profile: `codex/gpt-6.1-sol/high`. Observed interface: Codex tools on Linux. Actual model/effort metadata, usage and cost are not exposed to this reviewer; requested identity is not asserted as measured runtime identity.

Baseline `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`; frozen `corrected-04`, `CANDIDATE-04.json`, fingerprint `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec`. Independently recomputed sorted tracked/nonignored-untracked regular files outside alspaces, hashing path/NUL/content/NUL, before and after investigation: exact matches. Initial Git status contained existing staged, modified and untracked AL-68 work; preserved it. Product and tests were read-only. This assigned report is the only repository write; native binary, scripts and synthetic fixtures remain under `/tmp`.

Read the assigned architecture persona/common procedure, CONTRACT/approved PLAN including the no-guessing interaction clarification, technical/regression rules, source, CORRECTIONS/FINDINGS, prior architecture-sol-03 and parent-requested data-sol-03 evidence, IMPLEMENTATION/TEST_REPORT, VERIFICATION and SMOKE. No sibling persona was loaded. This is a correction review of CAP-12 and its affected interactions, not a new broad audit.

## Fixed findings and independent native evidence

### CAP-12 / ARCH03-1: byte starvation is fixed for the exact prior counterexample

`backend/internal/adapters/capture/reconcile.go:69` now persists a four-class priority phase independent of the configured allowance. Each class receives first position once every four passes, with the full session and byte allowance still available. Subsequent classes share the remaining resources. Ordinary owner refresh therefore cannot permanently lose all bytes to retained failed metadata.

Built the exact candidate with `go build -o /tmp/al68-arch04-cli ./cmd/agentic-lens` from backend; exit zero. Binary SHA256: `c1dab39ebf858422c5cf44e955eedfc35f0d5ea05aae40bebb92c0299959c5dd`.

Adapted the original architecture-sol-03 native CLI reproduction only to select that new binary and assert the result: `/tmp/al68-arch04-budget.py`, fixture `/tmp/al68-arch03-budget-_ipbdtk_`. After accepting root input 10, retained the same malformed 1,024-byte stop entry and appended the same authoritative later record to the valid root, whose complete source is 509 bytes. All eight separate CLI reconciliations kept `--max-sessions 2 --max-source-bytes 1024`.

- Pass 1 reserved 1,024 bytes and retained input 10 as pending.
- Pass 2 read 509 bytes, accepted cumulative input 30 and advanced the successful-read checkpoint.
- Across all eight passes, source attempts stayed at two, reserved bytes never exceeded 1,024, and the malformed queue row remained durable and reported pending. Later constrained passes retained the accepted input 30 even when refresh was pending.

The original candidate-03 eight-pass failure and unchanged-source one-attempt control are recorded in architecture-sol-03 and independently replayed by the tester. This after-fix replay keeps the original failing allowance; neither a larger allowance nor evidence deletion explains progress.

### CAP-12 / data-sol-03 cross-queue starvation is fixed

`backend/internal/adapters/capture/children.go:33` separates ordinary refresh, Codex stops, Codex starts and OpenCode discovery into durable priority classes. `recoverChildren` selects class-specific parent cursors and invokes stop queue processing separately from the start metadata walk. A permanently failed stop can consume its own first turn, but cannot remain first ahead of start discovery on every pass.

Adapted the original native interaction script to this binary, isolating its cross-queue reproduction and adding bounded-counter/source-fact assertions: `/tmp/al68-arch04-interactions.py`, fixture `/tmp/al68-data03-interaction-y3vd5qka`. Retained one malformed stop and a different valid source child, delivered only SubagentStart for that child using a distinct opaque interface ID, and ran all 20 separate CLI passes with the same two-attempt allowance.

The exact native child `child-only-start` registered on pass 2 and its source snapshot was captured on pass 3. The captured parent was `root` and input was 3. No registration used the opaque interface ID. Every pass used at most two attempts and the default byte allowance; the failed stop remained in `pending_children` throughout. The first disposable assertion assumed an undiscovered child's nil snapshot would have a JSON key and raised KeyError; corrected the probe to handle the omitted nil field, then reran from a fresh fixture. That probe error was not a product failure.

### Previous within-queue fixes remain effective

Independently replayed the original OpenCode known-prefix case through `/tmp/al68-arch04-children.py`, fixture `/tmp/al68-data-children-q7cqgr_7`. With root/child-a already known and child-b added later, twelve one-attempt CLI passes reached child-b on pass 5 while retaining existing owners. All passes respected one attempt and the default byte allowance. Source-ID pagination in `CaptureChildPage` still applies `id > after` before LIMIT; the saved child cursor and class-specific parent cursor are retained.

Queued Codex paths still rotate using the persisted queued-path cursor before malformed/mismatched metadata is read. Failed metadata remains queued; source parent equality is still checked before inheritance. Reused the passing production-adapter malformed/mismatched-prefix regression after inspecting its oracle. No queue ordering or identity proof was removed to obtain class fairness.

## Architecture and dependency assessment

The scheduling state belongs to the adapter: it controls SQLite continuation, source reads and resource allocation. Domain capture functions remain deterministic, take explicit values and introduce no I/O or clocks. `workClass` and `workBudget` express concrete scheduler responsibilities rather than a speculative framework.

`refreshTicketOwners` retains the existing owner cursor, exact extraction, source snapshot acceptance and transactional `Store.Change`. A per-pass attempted-owner set prevents the optional leftover refresh from reading an owner twice. Both command reconciliation and final generation still reach the same `Store.Reconcile`; defaults and caller timeouts are unchanged. Attempts decrement from the shared counter, metadata reservations charge before parsing, and exact failed reads retain their reservation. The tested fix concerns fair allocation across passes; it does not relax either per-pass limit.

Earlier evidence carries forward with these limits and reasons:

| Guarantee | Disposition and rationale |
| --- | --- |
| CAP-01 source-checkpoint precedence | Fixed for previously assigned old/equal/unknown rereads. `filterCapturePrecedence`, `rejectOlderExplicitOwners` and the domain source-completion checkpoint policy are unchanged; only scheduling/extraction invocation order changed. Successful refreshes still populate the explicit accepted-owner projection; failed attempts retain snapshots. Reused prior native proofs and current affected CLI race coverage. |
| CAP-02 publication ownership | Unchanged lock boundary starts before canonical baseline/source reads and covers ledger/report/summary publication, including absent-registry explicit generation. Scheduler work remains inside that boundary; no new writer bypasses it. Prior overlap/crash evidence and current CLI race result carry forward. |
| CAP-03 actual source-store identity | `CanonicalStore` resolves symlinks, `SameStore` compares actual filesystem identity, and `Store.Change` checks/migrates equivalent legacy rows within the write transaction while preserving decoded state. Scheduler passes identities through these same operations. Prior alias/legacy-snapshot preservation and distinct-store control proofs remain applicable. |
| CAP-05 failed-read bounds | Charging and exact-source failure handling remain intact; both native starvation replays additionally assert shared counters on every pass. No discarded pending work or uncharged successful read was introduced. |
| CAP-06 template paths | Installer still uses one `strings.NewReplacer` pass on the original template. Scheduler changes do not touch rendering, executable paths or owned content. Prior placeholder-path proof carries forward. |
| CAP-07 stdin timeout | Native receive still uses context cancellation and closes its input on timeout; source extraction remains outside receive. Scheduler changes neither the handoff nor callback timeout. Prior blocked-input evidence carries forward. |
| CAP-10 cutoff fidelity | Reconcile still assigns completion cutoff on return; each successful owner read sets its own `refreshed_at`. Public semantics explicitly exclude a common historical snapshot. New ordering does not substitute scheduler phase or publication clock for source chronology. |
| Ticket isolation and interaction | Eligibility remains proven ticket ownership; metadata/native parent relations are validated before inheritance. No directory/branch/role/opaque agent-ID guess is introduced. Hook feedback and binding protocols remain unchanged. |

## Verification reuse and remaining limits

TEST_REPORT records matching candidate-04 before/after fingerprints and passing `go test -race ./internal/adapters/capture -run '^TestReconcileScheduler' -count=1` (1.567s), followed by complete domain/capture/CLI race packages (1.172s/5.140s/13.835s). Inspected both new regression oracles: unchanged failing allowances, eventual source-fact progress, retained failed evidence and bounded counters; the start-only test also has a no-failed-stop control. Reused those results without repeating suites. Direct source-package results carry forward for unchanged adapters/dependencies. `git diff --check` passed.

Candidate-03 full `make` PASS is prior integrated build/test evidence, not a candidate-04 final gate. Parent owns the remaining final integrated `make`, consolidation and publication. No full-suite repeat, child delegation, source/test mutation, staging, commit, telemetry or publication was performed here.

Recorded real OpenCode 1.19.5 Linux context/tool-binding/source exchange remains applicable to the unchanged feedback protocol. These synthetic filesystem/SQLite/native CLI schedules verify algorithmic corrections; they do not supply live model evidence. Required real Codex trust/callback/agent response/native-child/source verification remains pending without a bypass or new trust question. Native non-Linux execution, other harness versions/families and resumed/interrupted live interactions remain limited as SMOKE states.

**Scoped code APPROVED; overall INCOMPLETE.** Both concrete candidate-03 scheduler failures and the affected known-prefix interaction are settled on the exact candidate-04 fingerprint. No additional concrete regression was found, and no speculative additional review pass is requested. The remaining Codex evidence and parent verification gates prevent overall delivery approval.
