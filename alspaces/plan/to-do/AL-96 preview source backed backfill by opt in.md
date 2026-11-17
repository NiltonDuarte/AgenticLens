# AL-96: Preview source-backed backfill by opt-in

## Goal

An operator can try the proposed ordinary backfill behavior deliberately and see what it would read, retain and write before adopting it.

## Context

[AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145) found explicit source recomputation already works without Force. Its small batch reused five rollouts across 12 ledgers, leaving real multi-ticket cost and external script expectations open.

## Acceptance criteria

- Offer an explicit opt-in for exact recorded owners, including completed ledgers. Provide a write-free dry run reporting predicted coverage and ledger/summary writes. On identical inputs compare the preview with the opt-in applied run, then repeat to verify stable bytes and mtimes.
- Demonstrate separate Codex and supported single-family OpenCode cases, a root with missing children, absent stores, unsupported mixed SQLite layout and mixed-harness ledgers. Retain prior trees and skill/task snapshots wherever evidence is incomplete. Use the selected local source. OpenCode heuristic fallback requires explicit opt-in; Codex has no title fallback.
- Document and exercise a maintenance-only invocation for scripts. Keep ordinary behavior, Force, explicit source recompute, Git-only ticket/ref selection, combined flags and failure signals compatible during preview.
- Measure representative multi-ticket elapsed time and source reads against configured budgets. Exercise exhaustion and targeted retry; distinguish measured costs from estimates. Present evidence and compatibility options for a later rollout decision. Product review selects the opt-in and maintenance-only interface before implementation.
