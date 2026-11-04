# Candidate 03 bounded behavior correction re-review

## Scope and identity

- Persona: behavior and integration reviewer.
- Requested and session-declared reviewer identity: gpt-6-luna / medium. I have no independent model telemetry beyond this task/runtime declaration.
- Candidate: branch `fix/AL-64-active-turn-duration`, baseline `a84adbe5be46c2ef52ede0987a417562715d38e1`.
- Frozen application/test manifest: `/tmp/al64-review-03/fingerprint.json`, SHA256 `f9b8bb31e5d7a3776adf9fef2b2f3f456711e76bfbceb7f276880e7e4634eb49`; all 46 listed paths were rehashed against the checkout with zero mismatches.
- Boundary: supplied DUR01-DUR06 corrections and interactions with Codex/OpenCode duration extraction and backfill fallback. This is a focused correction pass, not a broad audit.

## Correction dispositions

- **DUR01 — APPROVED.** For an invalid Codex completion boundary, `rollout.go` retains the already completed subtotal, marks the duration limitation, and does not emit a provisional interval or invent a completion timestamp. The regression table checks reversed and malformed completions against a prior valid interval: 4 known seconds remain, evidence is partial, and unknown coverage is disclosed. This avoids falsely certifying a zero-second complete result.
- **DUR02 — APPROVED.** OpenCode v1 unfinished activity is bounded by the native running-tool start (falling back to record creation only where that start is absent) or the open streaming part's native start. `time_updated` is not used as the activity start. The query gates evidence against the snapshot cutoff and requires no later message at or before that cutoff. This preserves an earlier open interval across a post-cutoff progress update.
- **DUR03 — APPROVED.** A later v1 message only supersedes the earlier activity when its creation timestamp is at or before the operation cutoff. This preserves legitimate historical supersession while excluding records that arrived after the snapshot boundary. Invalid/missing start timestamps do not create an interval.
- **DUR04 — APPROVED.** Backfill fallback extraction receives the operation's `SnapshotCutoff` through its extraction options and records that same value in provisional evidence. The targeted engine regression uses a fixed cutoff and checks the persisted evidence value, preventing refresh from silently using a different wall-clock cutoff.
- **DUR05 — APPROVED.** The v1 activity query includes open text/reasoning parts, and v2 derives current activity from its supported open-session/running-tool conditions. Tests cover no-tool streaming and idle-null v2 states, while path-removal mutants reportedly distinguish the regression. This corrects the prior tool-only blind spot. I inspected the production predicates and test locations; I did not rerun the mutation probes.
- **DUR06 — APPROVED.** A malformed or missing Codex start paired with a terminal lifecycle cannot manufacture an interval: the invalid boundary sets a limitation, and already completed known activity remains in the subtotal. The new table cases exercise this path. The change is limited to the parse-error limitation branch and corresponding regression cases; no change to valid duplicate-start or future-cutoff behavior was found in the inspected path.

## Interactions and counterexamples assessed

- Invalid timestamp parsing: malformed/reversed Codex lifecycle boundaries preserve known totals and disclose unsupported coverage; no false provisional duration is emitted for those terminal cases.
- Cutoff ordering: Codex lifecycle records after cutoff are excluded from lifecycle calculation. OpenCode v1's `NOT EXISTS` condition only treats a later message as superseding when its timestamp is no later than cutoff. OpenCode v2 similarly tests later messages against cutoff before clearing open activity.
- Missing native starts: duration evidence does not infer a start from a generic update timestamp; the affected assistant message receives an explicit limitation. Older OpenCode v1 tool records can use part creation time as the documented fallback.
- Completed semantics: supported terminal/completion timestamps at or before cutoff remain completed measurements; a future completion is clipped to provisional evidence by the shared duration interval calculation. Invalid terminal Codex evidence is partial, not complete.
- Fallback interaction: the operation cutoff flows through backfill options to fallback extraction. The fixed-cutoff test checks that exact cutoff rather than comparing against a moving clock.
- No distinct new introduced finding was demonstrated in this bounded pass.

## Verification and evidence limits

- Executed: candidate manifest verification (46 paths; zero mismatches); read-only inspection of Codex lifecycle parser/tests, OpenCode v1/v2 activity queries and duration accumulation, and backfill cutoff propagation/tests.
- Reused from the correction brief/test report: affected Codex/OpenCode/backfill race suites passed; original-code red/current-code green probes for DUR01-DUR04; streaming path-removal mutants for DUR05; DUR06 pre-fix branch overlay was red and corrected full Codex race passed. Candidate01 full Go race, frontend 834 tests, typecheck, and lint are reused for unchanged scopes.
- Not rerun: affected suites or mutants; full `make`; frontend/UI runtime checks; native harness probes. Avoided duplicating full suites absent a concrete counterexample.
- Runtime limits remain: native OpenCode v1/v2 live generation and source refresh are documented as passing by the runtime evidence owner. Native Codex CLI generation/refresh remains unverified because project ownership validation correctly rejected the copied source in a disposable project. Deterministic Codex lifecycle tests and source-record observations do not substitute for that CLI runtime path.
- The open persisted-source interval versus process-live-proof clarification remains **PENDING**. Runtime evidence shows a v1 interrupted process can leave unfinished source markers in storage. This pass does not decide or approve the provisional source-state policy. The accepted reader-first v6 compatibility limit also remains: readers must be upgraded before writer rollout; this is not release authorization.

## Verdict

**APPROVED — scoped correction pass only.** The six specified behavior corrections are supported by the inspected implementation and available regression evidence, with no new distinct defect demonstrated in this focus. This is conditional on the pending product-policy clarification and does not close the overall delivery gate. Parent consolidation, integrated `make`, policy resolution, and final approval remain outstanding.
