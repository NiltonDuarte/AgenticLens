# Review PRs merged in the past 40 hours

REV-8 is an ad hoc audit of AgenticLens. Review every pull request merged from 2026-11-17 07:47:05 UTC through 2026-11-18 23:47:05 UTC. Identify reproducible regressions or unmet requirements and distinguish them from existing defects that remain in the merged revision.

The scope comes from the audit request. GitHub merge timestamps determine inclusion and Git commit parents define each diff.

The follow-up repairs the two regressions reported in the audit handoff, the related batch maintenance race, and the unavailable shared engine pin. The [remediation requirements](specs/review-remediation/spec.md) define the outcomes.

The planned third iteration lets capture refresh already-bound Codex sessions whose saved rollout paths disappear after archival. The six REV-8 rollouts found under `archived_sessions/` establish the motivating case. The [requirements](specs/review-remediation/spec.md#requirement-recover-archived-codex-rollouts-for-bound-sessions) define when capture may repair a path and when it preserves the prior snapshot.
