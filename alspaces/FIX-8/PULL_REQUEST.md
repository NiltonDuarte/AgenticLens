# fix(backfill): preserve Codex rollout session trees (FIX-8)

## Summary

Refresh each stored Codex root rollout tree once and replace the complete tree session set together. Failed discovery or parsing leaves stored sessions intact, and unrelated Codex and other-source sessions are preserved.

Update `AGENTS.md` to define `<WORKSPACES_DIR>` from `.agenticlens-config.yaml`'s `workspaces.dir`, with `workspaces` as the fallback when the file or key is absent. The current configuration resolves to `alspaces`; ticket, plan, iteration, telemetry, and PR paths consistently use the configurable root.

## Context

- Contract: `alspaces/FIX-8/CONTRACT.md`
- Work summary: `alspaces/FIX-8/SUMMARY.md`
- Iteration plans and reviews: `alspaces/FIX-8/iterations/`

Closes FIX-8.
