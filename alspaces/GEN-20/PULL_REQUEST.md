# chore(capture): use unique worktree directory identifiers (GEN-20)

Replace sequential names for AgenticLens state worktree directories with stable qualified Git worktree identifiers where available and a persisted short UUID fallback. Keep allocation isolated across concurrent processes and resolve committed paths consistently for previews, installation and relocation.

The database schema is unreleased; no compatibility migration or old-binary protocol fence is included, as requested.

The mapping stores its value as directory_name; corresponding Go helpers, variables and diagnostics call it a directory name.

Baseline contract: [CONTRACT.md](CONTRACT.md). Current scope delta: [iterations/02/PLAN.md](iterations/02/PLAN.md). Delivery context: [SUMMARY.md](SUMMARY.md).
