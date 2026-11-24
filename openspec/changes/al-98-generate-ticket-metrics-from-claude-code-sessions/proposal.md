# AL-98: Generate ticket metrics from Claude Code sessions

## Roadmap ticket

[AL-98: Generate ticket metrics from Claude Code sessions](../../../alspaces/plan/to-do/AL-98%20generate%20ticket%20metrics%20from%20claude%20code%20sessions.md)

## Why

`generate`, `backfill` and local `telemetry.default_source` accept only `codex` or `opencode`. A developer working in Claude Code can therefore only record telemetry as unavailable, as [AGENTS.md rule 7](../../../AGENTS.md#1-ticketed-and-ad-hoc-work) currently requires. This change adds a manual, source-backed path so that a Claude Code root session, its native subagents and their usage, models, skills, delegations and duration reach a portable ticket ledger. Cost coverage stays explicit, and no other harness is read as a fallback.

## What changes

- A `claude` source for manual `generate` and explicit `backfill --recompute-db`. The local `telemetry.default_source: claude` and `telemetry.sources.claude.home` settings work with a new `--claude-home` flag. `init` local setup can write these settings. Precedence and no-fallback rules match the Codex home. Backfill `--source all` now includes Claude Code, while `--source both` still means OpenCode plus Codex.
- A Claude Code transcript reader for the store layout of version 2.1.331: root transcripts under `<store>/projects/<folder>/<sessionId>.jsonl` and native subagents under `<sessionId>/subagents/agent-<agentId>.jsonl`, each with a sibling `.meta.json`. Selection uses a literal nonce or an exact session ID. A subagent becomes a child only when the source proves its parent. Usage counts once for each API response. Duration follows AL-64. Skills and delegations are owned by the session that recorded them. Invocation cost stays unavailable.
- Backend ledger validation and dashboard response validation accept the `claude` session source. Ledgers that include Claude Code owners remain readable, and their facts are refreshed and preserved under the portable ledger contract.
- Setup, operator and developer guidance, the AGENTS.md rule 7 command and the source capability matrix name the `claude` source, its tested version and its retention caveat.

Compatibility: engines released before this change reject a ledger that contains a `claude` session. This includes the repository's pinned `v3.0.0-rc.156` until it advances. The ledger schema version does not change. Backfill with `--source all` now requires a readable Claude Code store.

Exclusions: automatic capture and `capture install --source claude` (AL-99), historical trace hydration and live views for Claude Code owners (these report an unsupported source), persisting effort, requested model or cache-creation tokens, API-equivalent pricing (AL2-6), effort display (AL2-8) and this repository's own Claude Code instructions (GEN-30).

## Capabilities

- New: [claude-code-session-telemetry](specs/claude-code-session-telemetry/spec.md)

## Impact

- **Users:** Developers and agents working in Claude Code gain manual ticket telemetry. Codex and OpenCode users see no change in behavior. Configuration that never selects `claude` keeps working unchanged.
- **Components:** The source and configuration resolvers, the generate and backfill CLI, a new Claude Code source adapter, backfill recovery, ledger decoding, dashboard validation and embedded assets, and the guidance listed above.
- **Preserved guarantees:** Only owners from a complete source read are replaced. Codex, OpenCode, other-machine, unavailable and cleaned-up owners survive refresh. Repeat runs are idempotent. Selection never falls back to another store or harness. Persisted facts stay readable without the original transcripts.
