# FIX-8 Technical Contract

## Problem

Codex backfill recomputation in `backend/internal/adapters/backfill/engine.go` currently discovers rollout paths once for each stored Codex session. A root session can resolve to a rollout tree containing the root and descendant sessions. `SessionsFromFactsWithSource` groups that tree into multiple sessions, but the recomputation keeps only `freshCodex[0]` and replaces only the current slot. This can discard child-session telemetry and leave stale or duplicate root entries in the resulting ledger.

## Required behavior

When Codex database recomputation is requested and stored Codex sessions exist:

1. Resolve/discover each distinct Codex root rollout tree once, rather than rediscovering the same tree for each stored session in it.
2. Parse and combine the discovered transcripts, then group all refreshed facts into the complete set of Codex sessions for that tree.
3. Replace the stored Codex sessions represented by that root tree as one logical operation. The resulting Codex collection contains every refreshed session in the tree once, with no stale versions or duplicate root IDs.
4. Preserve stored Codex sessions unrelated to the refreshed tree and preserve all non-Codex sources unchanged.
5. Treat discovery, parsing, combining, or grouping failure as a failed refresh for that tree: retain its original stored sessions intact. Successful refreshes for other independent trees may proceed.

The existing output ordering by session ID remains in effect. Existing DB recomputation status fields must indicate an update only when at least one tree refresh succeeds, and indicate a found Codex session only when refresh yields session data.

## Architecture boundaries

- Keep filesystem discovery and transcript parsing in the adapter shell (`backend/internal/adapters/backfill`).
- Keep session grouping through the existing pure domain API (`domain.SessionsFromFactsWithSource`); do not introduce I/O or discovery concerns into `backend/internal/domain`.
- Use concrete Go types and preserve the repository's zero-`any`/zero-`interface{}` policy.
- Add no emoji to code, comments, logs, or artifacts.

## Acceptance criteria

- A root-plus-child rollout refresh replaces the complete tree and retains telemetry for both root and child sessions exactly once.
- A one-session rollout refresh replaces its existing session without adding or dropping sessions.
- A discovery failure leaves the affected stored Codex sessions unchanged and does not damage unrelated Codex sessions or other sources.
- Existing unrelated OpenCode and other-source sessions remain unchanged by Codex tree replacement.
