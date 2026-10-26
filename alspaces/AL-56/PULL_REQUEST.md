# feat(AL-56): add Codex session discovery and metrics generation

Adds opt-in Codex rollout discovery and metrics extraction through `agentic-lens generate --source codex`, while keeping OpenCode as the default. Session selection requires a literal nonce or explicit thread ID; ambiguous nonces fail, ticket/workspace heuristics are excluded, and linked child rollouts are included.

Finalizes the unreleased schema v4 as a session-owned ledger. Each session contains its invocations, per-tool call/error summaries, and retained events; session identity is stored once. Cross-session tool totals and dashboard rows are derived from session facts. OpenCode and Codex generation, migration, repository loading, backfill, and UI use this shape.

Persist anomaly summaries, invocation facts, and tool totals without event rows, evidence references, or generated anomaly titles/descriptions. Anomaly records keep rule ID, severity, and structured context; CLI/API/UI format readable labels from those values. Legacy presentation strings are discarded during normalization. The ledger records its harness source once. Local event views rebuild Codex details from `CODEX_HOME` rollouts and OpenCode details from SQLite, then compute event/anomaly associations in the response. Missing, unknown, or incomplete sources return an empty unhydrated timeline while preserving stored anomaly summaries. Retry detection is session-scoped, and rework detection respects invocation parent groups.

Migration converts v1-v3 and interim-v4 ledgers, rejects mixed or malformed data without modifying the source, and enforces event privacy. Legacy tool aggregates without session attribution are removed with a count and recovery guidance in migration notes and CLI output. This branch does not modify any other ticket ledger; other ledgers remain for backfill on the machine that has their source data. No workspace-wide backfill was run.

Verification: `make` passed, including Go race/vet, 728 frontend tests, lint/type checks, frontend build, and four platform builds. Independent reviews approved the summary-only ledger, local hydration, and derived anomaly text. Final AL-56 generation removes event, evidence, title, and description fields. The final size comparison is recorded in iteration 08 verification. No other ticket ledger changed and no workspace-wide backfill ran.

Code attribution uses `--code-ref HEAD` by default; ticket/branch-based discovery remains removed. No HTTP routes or import UI were added.

## Limits

Codex rollout persistence is private and version-sensitive. Discovery does not tail live sessions. Cost is unavailable in observed rollout data; nullable cost remains a separate follow-up.
