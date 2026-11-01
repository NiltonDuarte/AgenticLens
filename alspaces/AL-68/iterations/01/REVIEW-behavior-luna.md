# AL-68 Behavior and Integration Review — behavior-luna-01

## Assessment

**CHANGES_REQUESTED** for the ticket-isolation defect below. This is scoped to the frozen candidate fingerprint `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`, against baseline `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. The fingerprint was recomputed from the sorted tracked and nonignored untracked regular repository files outside `alspaces`, hashing repository-relative path bytes, NUL, file bytes, NUL; it matched `CANDIDATE.json` exactly.

The requested reviewer identity was `codex / gpt-6-luna / medium`. Runtime model/effort metadata for this review was not independently available, so this records the requested settings and makes no claim that they were observed. Only the assigned behavior report was written; application and test paths were read-only.

## Finding

### Unassigned registrations from unrelated work can consume a ticket's refresh budget and enter its published report

**Location:** `backend/internal/adapters/capture/reconcile.go:109-113, 240-246`; report redaction in `backend/internal/adapters/capture/reconcile.go:335-349`.

**Trigger and path:** A repository state directory contains an unassigned registration for an unrelated session and a bound registration for ticket A. Generating A calls `Reconcile` for A. The selector adds every `TicketID == ""` registration alongside A's registrations. Those unassigned sessions are source-extracted under A's `sessionBudget` and shared `byteBudget`. With enough unrelated sessions or source data, the bounded pass can spend its budget before refreshing A's bound session; the resulting final ledger can omit or preserve a stale snapshot for A while reporting the pass as partial. The loop does not persist the pending state it sets for registrations beyond the session budget, making the report's local status more misleading on that pass.

The same unassigned registrations are appended to `result.Registrations` for every ticket. `generate` publishes that result in the ticket's `capture_report.json`; `redactedRegistrations` removes request tokens and source-store paths but keeps session IDs and other registration fields. Thus generating A can disclose unrelated unassigned session identities in A's published report. This crosses the stated boundary that reconciliation is bounded by the requested ticket's registered owners and weakens isolation when multiple roots/tickets share a checkout.

**Evidence:** This follows directly from the selector at lines 109-113, source refresh loop at 121-150, result construction at 240-258, and report serialization/redaction at 335-349. `backend/cmd/agentic-lens/main.go:196-203, 307-326` invokes this reconciliation during `generate` and publishes the returned report. No mutation or probe was needed to establish reachability. The target-ticket selector should exclude unrelated unassigned owners from ticket refresh/publication, while preserving an explicit local diagnostic path for unassigned registrations. If source-proven parent inheritance requires examining a parent, scope that read to the relevant registered tree rather than every unassigned registration.

## Behavior traced and hypotheses assessed

- **Install to handler:** The downloader delegates opt-in lifecycle setup directly to the platform native binary. It reports binary-only setup when capture is not requested and refuses to install hooks without a native handler. The wrapper's `capture` path bypasses update/download behavior. The Go installer protocol-checks the binary and writes additive Codex hooks or an OpenCode plugin with ownership tracking and rollback. Script/setup tests are described in `SETUP_IMPLEMENTATION.md`; I reused that evidence and did not rerun tests.
- **Agent binding:** Codex `SessionStart`/`UserPromptSubmit` registration returns a session-specific binding request only when unbound. `SubagentStart` is explicitly not treated as the child identity. OpenCode's per-session transform requests binding, and the tool uses runtime-provided `context.sessionID`; it does not accept a caller-supplied session identity. This matches the no-guessing requirement at source level.
- **Parent/child and roots:** Codex child capture queues the authoritative `agent_transcript_path` at `SubagentStop`, then parses source metadata and accepts the child only when its `ParentThread` equals the callback's parent ID. OpenCode reconciliation uses the exact registered source session and source `parent_id`. Accepted source-proven parent linkage is propagated to a child; correlation does not fabricate a parent. The real OpenCode 1.19.5 Linux smoke record reports two independently bound roots and a native child with source `parent_id` matching the second root; the child binding interaction completed. Repeated reconciliation preserved the compact snapshot digest.
- **Independent roots / explicit legacy selector:** Registered ticket roots are collected by `Reconcile` independently of native-tree discovery. Without explicit selectors, a bound registration activates the registered-only path in `generate`; with `--nonce` or `--session-id`, the original Codex `DiscoverRolloutPaths` path still runs and its selected source sessions are merged with capture snapshots. I inspected the unchanged selector implementation and the changed dispatch; I did not rerun its tests.
- **Reconciliation and preservation:** The source path is exact-ID for OpenCode and uses the registered Codex transcript path. Failed/partial reads preserve prior snapshots; accepted cumulative snapshots replace owner snapshots. OpenCode capture selects a schema family by exact ID membership and rejects ambiguous duplicate ownership. I found no additional consequential behavior defect in these paths within this focus.

## Verification and evidence reused

- Candidate fingerprint recomputation: **matched** `CANDIDATE.json`.
- `TEST_REPORT.md` and `VERIFICATION.md` report the integrated `make` gate passed on the fingerprinted candidate, including Go race tests, frontend checks/build, capture distribution tests, and Linux/Darwin cross-builds. I reused that result and ran no tests or build commands.
- `SMOKE.md` records an actual OpenCode 1.19.5 Linux interaction: binding response on independent roots, binding for a native child, child source-parent verification, exact source-derived session facts, and idempotent repeated reconciliation. The same smoke record says Codex CLI 0.166.0 hook trust review and real callbacks/binding/source comparison remain pending; it explicitly reports no trust bypass. This remains an **acceptance evidence gap**, not evidence that the Codex path is defective. No claim of verified Codex runtime support is warranted until that human-controlled trust review and actual interaction/source comparison are complete.
- The reported OpenCode smoke model was `gpt-5.6-luna`, with effort unknown. That is smoke evidence, not this review's runtime identity.

## Coverage limits

This is a focused behavior/integration review, not a whole-repository review. OpenCode real runtime evidence is Linux 1.19.5 only. Codex source-level flow and fixtures were inspected, but actual Codex agent binding and trusted callback execution remain unverified. No macOS/Windows runtime, OpenCode v2 real run, resumed-session real run, or real interrupted-hook recovery was independently exercised. No tests were rerun because the raw report documents a full passing `make` on the unchanged candidate fingerprint.
