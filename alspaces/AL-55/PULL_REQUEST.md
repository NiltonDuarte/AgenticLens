# feat(backfill): preserve incomplete historical snapshots (AL-55)

Explicit source recompute now refreshes complete stored session trees without requiring Force. Missing owners, children or malformed source records retain their previous trees; incomplete OpenCode recovery also retains whole skill/task snapshots and reports mixed-age data per ticket and in the command summary.

Existing global source-backfill obligations require all known owners across required sources. Failed optional refresh keeps valid canonical ledgers usable. File reads provide empty internal skill/task arrays when optional v4 fields are absent or null, while preserving malformed-value validation and populated rows.

Keeps schema v4, optional wire fields, existing migration chain and backup interfaces. No cost or UI changes. Recovery is bounded by known owners and discovery; retained collections can be older than refreshed sessions.

See [contract](CONTRACT.md), [summary](SUMMARY.md) and [iteration plan](iterations/01/PLAN.md). Shared specification lifecycle is deferred pending the manager's PO handoff. Manager owns merge sequencing; no auto-merge or release action is requested.
