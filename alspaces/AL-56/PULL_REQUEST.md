# feat(AL-56): add Codex session discovery, finalize schema v4, and backfill historical workspaces

## Summary

Adds opt-in Codex rollout discovery and metrics extraction through `agentic-lens generate --source codex`, while keeping OpenCode as the default harness. Session selection requires a literal nonce or explicit thread ID; ambiguous nonces fail, ticket/workspace heuristics are excluded, and linked child rollouts are attributed.

Finalizes schema v4 as a lean session-owned ledger where each session directly encapsulates its model invocations and per-tool call/error summaries. Persisted session facts store only summary data; raw event rows, evidence references, and redundant presentation strings (titles/descriptions) are omitted from storage to eliminate repository bloat. At runtime, trace events and evidence are rebuilt on demand from local source logs (`CODEX_HOME` rollouts for Codex or SQLite for OpenCode) via `/api/v1/tickets/:id/events`, while human-readable anomaly descriptions are dynamically formatted from structured context across CLI, API, and UI.

Performs a complete migration and backfill across all 67 historical ticket workspaces under `alspaces/`:
- Upgrades historical ledgers to schema v4.
- Re-extracts true session-owned tool facts directly from the local OpenCode SQLite database via `agentic-lens backfill --source opencode --recompute-db --workspaces alspaces --force --update-summaries`.
- Restores accurate tool metrics across all 67 workspaces, resolving missing tool call and error data in autopsy summaries.

## Key Changes

- **Codex Telemetry Extraction**: Rollout discovery by nonce/session ID, child rollout traversal, token usage accounting, and configurable role mapping.
- **Session-Owned Schema v4**: `SessionFact` holds `invocations` and `tools` (non-nullable arrays). Cross-session tool totals and dashboard metrics are derived on demand.
- **Event & Evidence Decoupling**: Raw telemetry events are removed from persisted `SessionFact`. Trace inspection is served dynamically from local sources without persisting raw event payloads in git.
- **Runtime Anomaly Presentation**: Anomaly records persist only `rule_id`, `severity`, and structured `context`. User-facing titles and descriptions are derived deterministically on client and server.
- **Workspace Backfill**: All 67 ticket workspaces under `alspaces/` upgraded to schema v4 with complete session-level tool accounting and refreshed `SUMMARY.md` markdown autopsies.
- **Frontend & Dashboard**: Updated TypeScript types, `TicketDeepDive` autopsy view, `AnomalyDrawer`, and trace inspection components to query dedicated trace endpoints and consume session-owned facts.

## Verification

- `make test`: All Go packages pass with race detector enabled; all 727 Vitest frontend tests pass.
- `make`: Go vet, TypeScript type check (`tsc --noEmit`), ESLint (`--max-warnings 0`), Vite production bundle, and multi-platform compilation (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`) pass cleanly.
- Telemetry Verification: Verified across all 67 workspaces that `sessions[].tools` arrays are populated with exact call and error distributions matching detected anomalies.

## Limits

- Codex rollout persistence is private and version-sensitive. Discovery does not tail live sessions.
- Cost is unavailable in observed Codex rollout data; nullable cost remains a separate follow-up.
