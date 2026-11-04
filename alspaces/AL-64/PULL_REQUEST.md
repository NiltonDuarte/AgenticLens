# fix(duration): account for source-proven active turns (AL-64)

Telemetry generated during supported unfinished Codex or OpenCode v1/v2 activity includes elapsed time through a recorded cutoff and identifies the contribution as provisional. Completed measurements retain their source-specific boundaries; unsupported evidence and legacy coverage remain explicit.

Compact duration evidence survives persistence and projections. Source-backed refresh replaces provisional contributions after completion without accumulating snapshots or discarding unavailable sessions. CLI and dashboard duration consumers expose coverage and cutoffs. Native OpenCode v1 auxiliary-table detection accepts the observed empty bootstrap layout while retaining strict ambiguous-layout rejection.

Reader-first v6 expansion: upgrade consuming readers before writer rollout; pre-change strict readers reject new evidence fields. No release rollout is authorized here.

Draft pending source-interval liveness policy clarification and final delivery gates. See [technical contract](CONTRACT.md) and [work summary](SUMMARY.md).
