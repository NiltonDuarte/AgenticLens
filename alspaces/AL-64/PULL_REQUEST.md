# fix(duration): include provisional active-turn duration (AL-64)

Generation during unfinished Codex or OpenCode v1/v2 activity includes elapsed time through the snapshot cutoff. The existing active_duration_seconds total includes an optional positive provisional_seconds portion. Completion refresh replaces the prior snapshot without double counting.

Backend readers accept the earlier draft duration evidence and normalize it to numeric fields on canonical writes. CLI summaries, APIs and dashboard displays show the recorded total and positive provisional portion. No reasoning field, configuration toggle or process-liveness check is added. Stored open intervals are provisional estimates and may outlive their process.

The branch preserves upstream optional metric omission and shared-session warnings. Native OpenCode v1 bootstrap-table detection accepts the observed empty auxiliary layout while retaining ambiguous-layout rejection.

Upgrade consuming readers before enabling new scalar writers; older strict readers may reject expanded v6 fields. No release rollout is authorized. Existing unrelated delegation-retention and legacy-null compatibility debt is recorded in the work summary.

See the [technical baseline](alspaces/AL-64/CONTRACT.md), [approved numeric-only plan](alspaces/AL-64/iterations/03/PLAN.md) and [work summary](alspaces/AL-64/SUMMARY.md).
