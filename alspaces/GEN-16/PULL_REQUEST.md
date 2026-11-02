# chore(GEN-16): retain compact structured review findings

Completed reviews currently retain findings only in prose. Require a permanent per-iteration REVIEW_SUMMARY.json with severity, a brief issue, every reporting persona/model/effort tuple and disposition before detailed evidence is deleted.

Preserve shared discovery credit, refuted candidates, identity uncertainty and the existing approval gate. Link the compact file from SUMMARY.md and align contributor/development/attribution guidance. Documentation only; historical audit records are unchanged.

Validation: both edited skills pass metadata validation; JSON example parses; changed-document local links and git diff --check pass. Integrated make passed, including Go race tests, 781 frontend tests, lint/typecheck, release/installation checks and four platform binary builds. Parent documentation review found no issues; no delegated or behavioral forward tests. Telemetry facts and snapshot limits are recorded in SUMMARY.md.
