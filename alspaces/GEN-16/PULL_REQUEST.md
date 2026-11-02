# chore(GEN-16): retain compact structured review findings

Completed reviews currently retain findings only in prose. Require a permanent per-iteration REVIEW_SUMMARY.json with a numeric severity score from 0 (low) to 10 (high), a brief issue, every reporting persona/model/effort tuple and disposition before detailed evidence is deleted.

Omit reviewed_revision from the compact JSON; retain revision evidence in SUMMARY.md. Preserve shared discovery credit, refuted candidates, identity uncertainty and the existing approval gate. Link the compact file from SUMMARY.md and align contributor/development/attribution guidance. Documentation only; historical audit records are unchanged.

Validation: both edited skills pass metadata validation; JSON example parses; changed-document local links and git diff --check pass. Integrated make passed, including Go race tests, 781 frontend tests, lint/typecheck, release/installation checks and four platform binary builds. Parent documentation review found no issues; no delegated or behavioral forward tests. Telemetry facts and snapshot limits are recorded in SUMMARY.md.

Follow-up format validation: severity is a JSON number in [0, 10], reviewed_revision is absent, and shared reporters are preserved. Reused the passing make result for unchanged application/test/runtime code; origin/main remains integrated.
