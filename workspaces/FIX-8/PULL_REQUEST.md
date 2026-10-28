# fix(backfill): preserve Codex rollout session trees (FIX-8)

## Summary

Refresh each stored Codex root rollout tree once and replace the complete tree session set together. Failed discovery or parsing leaves stored sessions intact, and unrelated Codex and other-source sessions are preserved.

## Context

- Contract: `workspaces/FIX-8/CONTRACT.md`
- Work summary: `workspaces/FIX-8/SUMMARY.md`
- Iteration plan and review: `workspaces/FIX-8/iterations/01/`

Closes FIX-8.
