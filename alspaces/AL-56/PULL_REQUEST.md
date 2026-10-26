# feat(AL-56): add Codex session discovery and metrics generation

Adds opt-in Codex rollout discovery and metrics extraction through `agentic-lens generate --source codex`, while keeping OpenCode as the default. Session selection requires a literal nonce or explicit thread ID; ambiguous nonces fail, ticket/workspace heuristics are excluded, and linked child rollouts are included.

Finalizes the unreleased schema v4 as a session-owned ledger. Each session contains its invocations, per-tool call/error summaries, and retained events; session identity is stored once. Cross-session tool totals and dashboard rows are derived from session facts. OpenCode and Codex generation, migration, repository loading, backfill, and UI use this shape.

Migration converts v1-v3 and interim-v4 ledgers, rejects mixed or malformed data without modifying the source, and enforces event privacy. Legacy tool aggregates without session attribution are removed with a count and recovery guidance in migration notes and CLI output. This branch migrates AL-56 only; other ticket ledgers remain unchanged for backfill on the machine that has their source data. Missing source logs mean unscoped totals cannot be recovered.

Code attribution uses `--code-ref HEAD` by default; ticket/branch-based discovery remains removed. No HTTP routes or import UI were added.

## Limits

Codex rollout persistence is private and version-sensitive. Discovery does not tail live sessions. Cost is unavailable in observed rollout data; nullable cost remains a separate follow-up.
