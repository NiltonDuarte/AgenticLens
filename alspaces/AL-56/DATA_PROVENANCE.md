# AL-56 Data Provenance

The implementation was validated against a real local Codex rollout. The real-session ledger is committed as `agentic_metrics.json`, with its generated metrics summary in `SUMMARY.md`. All committed parser and discovery fixtures are synthetic.

Session correlation uses a fresh literal nonce passed to `generate --source codex`. The nonce is searched in persisted rollout records; a read-only lookup in the thread-history projection and state catalog can independently correlate the thread ID and rollout path. Actual identifiers, correlation nonces, working-directory paths, timestamps, and usage snapshots are omitted from published notes.

The extractor streams rollouts through `codex.ParseRollout`, maps observed facts into the AgenticLens domain, and prunes persisted events using the shared policy. Code facts come from the Git analyzer. Ticket metadata is supplied by the caller. Prompts, assistant/reasoning text, command arguments, and tool inputs/outputs are omitted from generated facts; working-directory paths are omitted from exported summaries. Generated ledgers still contain real session identifiers and timestamps, so they should be reviewed before sharing.
