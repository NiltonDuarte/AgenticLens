# FIX-8: Preserve Complete Codex Session Trees During Backfill Recompute

## Problem

Codex backfill refresh currently discovers a root rollout tree for each stored session, groups the full tree, then retains only the first refreshed session. This can discard child-session facts and create duplicate root sessions.

## Outcome

Refresh each distinct Codex root tree once and replace its complete stored session set atomically. Preserve unrelated Codex sessions and all other telemetry sources. If discovery or parsing fails, retain the affected stored tree unchanged.

## Acceptance criteria

- Root and child rollout sessions are refreshed exactly once, without stale or duplicate session IDs.
- Single-session refresh continues to work.
- Failed refresh preserves the original affected sessions.
- Unrelated Codex and non-Codex sessions remain unchanged.
