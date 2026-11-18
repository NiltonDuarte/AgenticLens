# AL-95: Preview source-backed defaults and decide on rollout

## Goal

Operators can try exact-source backfill by opt-in. The product owner can use that evidence to approve, defer or decline a later default switch.

## Context

Explicit recomputation already works without Force. The discovery batch reused five rollouts across 12 ledgers, leaving representative batch cost and external script compatibility open. [AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145)

## Acceptance criteria

- Offer an explicit opt-in for exact recorded owners, including completed ledgers. Provide a write-free dry run reporting predicted coverage and ledger/summary writes. On identical inputs compare the preview with the opt-in applied run, then repeat to verify stable bytes and mtimes.
- Demonstrate separate Codex and supported single-family OpenCode cases, a root with missing children, absent stores, unsupported mixed SQLite layout and mixed-harness ledgers. Retain prior trees and skill/task snapshots wherever evidence is incomplete. Use the selected local source. OpenCode heuristic fallback requires explicit opt-in; Codex has no title fallback.
- Document and exercise a maintenance-only invocation for scripts. Keep ordinary behavior, Force, explicit source recompute, Git-only ticket/ref selection, combined flags and failure signals compatible during preview.
- Measure representative multi-ticket elapsed time and source reads against configured budgets. Exercise exhaustion and targeted retry; distinguish measured costs from estimates. Present evidence and compatibility options for a later rollout decision. Product review selects the opt-in and maintenance-only interface before implementation.

- Review those results against agreed batch size, runtime and read budgets. Include summary-only repair and retained-owner diagnostics. Identify unverified workloads and external-script assumptions.
- Record an explicit go, defer or retain-current-default decision with rationale. Deferral or retention is a valid outcome. Preview completion, roadmap readiness and ticket creation confer no authority to switch.
- Before approving a switch, agree release timing, announcement, maintenance-only invocation, Force support or deprecation window, failure/exit compatibility and a way to restore prior behavior. Show migration examples for legacy batch, Git-only, explicit source and combined calls.
- Any rollout requires explicit product approval and separately authorized implementation. Preserve explicit harness and Git ticket/ref selection, existing recovery safeguards and opt-in-only heuristics throughout.
