# fix(duration): account for source-proven active turns (AL-64)

Generation during supported unfinished Codex or OpenCode v1/v2 activity records elapsed time through a snapshot cutoff and marks it provisional. Completion refresh replaces that contribution without adding both snapshots. Compact duration evidence survives persistence, CLI/API projections and dashboard displays; legacy and unsupported coverage remain explicit.

The branch includes main through 40b16def, preserving optional metric omission and shared-session warnings. Native OpenCode v1 bootstrap-table detection accepts the observed empty auxiliary layout while retaining strict ambiguous-layout rejection.

Draft boundary: iteration 02 proposes exact-caller verification and supported read-only checks for other sessions; that scope is unapproved and unimplemented. Stored open markers can outlive a process. Native Codex fresh generation/refresh remains unverified after an isolated project-ownership rejection. A pre-existing registry delegation-retention concern is recorded as INT-DATA-01.

Reader-first v6 expansion requires upgrading consuming readers before writer rollout; pre-change strict readers reject the evidence fields. No release rollout is authorized.

Telemetry contains the root and 18 native children; 12 delegated reviews lack native parent links and are not included. Pricing is unavailable.

See [technical contract](alspaces/AL-64/CONTRACT.md), [work summary](alspaces/AL-64/SUMMARY.md) and [iteration 02 proposal](alspaces/AL-64/iterations/02/PLAN.md).
