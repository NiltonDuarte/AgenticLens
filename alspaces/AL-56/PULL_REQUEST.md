# feat(AL-56): add Codex session discovery and metrics generation

Adds opt-in Codex rollout discovery and metrics extraction through `agentic-lens generate --source codex`, while keeping OpenCode as the default. Codex selection requires a literal nonce or explicit thread ID; ambiguous nonces fail, ticket text and working-directory heuristics are excluded, and root selection includes linked child rollouts. Real-session smoke extraction produced the committed ledger from local Codex rollout data.

The ledger remains schema v4 and now stores tool facts per `(session_id, tool_name)`. Repeated generation replaces matching session facts and retains facts from other sessions; dashboards aggregate by tool name. SDLC iteration totals preserve exact zero and derive from root-session user messages. Every v4 tool row requires a nonempty `session_id`; old unscoped aggregate rows are intentionally rejected, which is a breaking migration/load behavior.

To recover an old unscoped ledger, preserve it and re-extract each root session into one new JSON output path, reusing that path to merge sessions. Codex root discovery includes child rollouts. Verify expected session IDs and tool counts before replacing the old ledger. Generation against the old path fails validation. If source logs are unavailable, unscoped aggregate counts cannot be split or recovered.

Code attribution requires an explicit ref (generation defaults to `HEAD`); ticket/branch-based discovery is removed. Backfill Git recomputation requires `--code-ref`. File classification uses configured test and production paths, including configured paths outside source roots. Frontend types and tool aggregation were updated; there are no HTTP routes or import UI changes.

## Limits

Rollout persistence is a private, version-sensitive Codex format. Discovery does not tail live sessions or select a harness automatically. Missing source timestamps remain unavailable. Cost is not exposed by rollout data; current numeric zero is ambiguous and nullable cost is tracked separately. No new dependencies were added.
