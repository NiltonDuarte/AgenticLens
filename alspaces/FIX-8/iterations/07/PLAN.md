# Iteration 07: Backfill AL-56 Codex Session Data

## Goal

Refresh Codex session facts stored in the AL-56 ledger using the AgenticLens backfill command, as an operational validation of the FIX-8 Codex tree refresh.

## Steps

1. Run the backfill scoped to `AL-56`, using `--source codex`, `--recompute-db`, and `--force` so completed Codex data is recomputed from local rollout files.
2. Update AL-56's rendered summary metrics and verify Codex sessions were refreshed while non-Codex session data remains present.
3. Record the command outcome in FIX-8's workspace and commit the resulting ticket artifacts on the FIX-8 branch.

## Acceptance

- The CLI reports one target ledger (`AL-56`) and a successful Codex backfill.
- The AL-56 ledger retains Codex root and child sessions and any non-Codex session records.
- The FIX-8 summary and review capture the outcome.
