# feat(AL-56): add Codex session discovery and metrics generation

Adds opt-in Codex rollout discovery and metrics extraction through `agentic-lens generate --source codex`, while keeping OpenCode as the default. Session selection requires a literal nonce or explicit thread ID; ambiguous nonces fail, ticket/workspace heuristics are excluded, and linked child rollouts are included.

Finalizes the unreleased schema v4 as a session-owned ledger. Each session contains its invocations, per-tool call/error summaries, and retained events; session identity is stored once. Cross-session tool totals and dashboard rows are derived from session facts. OpenCode and Codex generation, migration, repository loading, backfill, and UI use this shape.

Persist only events that directly support a detected anomaly; events elsewhere in the same turn are dropped. Each retained event has a stable source ID, and anomaly summaries reference the exact `(session_id, event_id)` pairs. High tool-failure summaries retain and reference the failed calls contributing to the numerator, with per-session tool totals providing the denominator. The UI highlights only exact referenced events. Missing IDs for required evidence fail extraction or migration with a diagnostic; source-less legacy aggregates do not produce untraceable anomalies. Rework evidence is grouped by invocation parent so independent child groups cannot synthesize a cycle.

Migration converts v1-v3 and interim-v4 ledgers, rejects mixed or malformed data without modifying the source, and enforces event privacy. Legacy tool aggregates without session attribution are removed with a count and recovery guidance in migration notes and CLI output. This branch does not modify any other ticket ledger; other ledgers remain for backfill on the machine that has their source data. The real AL-56 ledger has not been regenerated for iteration 06; its conversion and real size comparison are deferred. No workspace-wide backfill was run.

Verification: `make` passed, including Go race/vet, 719 frontend tests, lint/type checks, frontend build, and four platform builds. Synthetic size coverage retained 132/619 rows (21.3%) and reduced fixture JSON by 71.9%. The independent iteration 06 review approved the revised evidence contract.

Code attribution uses `--code-ref HEAD` by default; ticket/branch-based discovery remains removed. No HTTP routes or import UI were added.

## Limits

Codex rollout persistence is private and version-sensitive. Discovery does not tail live sessions. Cost is unavailable in observed rollout data; nullable cost remains a separate follow-up.
