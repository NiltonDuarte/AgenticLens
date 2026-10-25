# AL-56 Research Contract

## Goal

Find and validate the strongest practical way to generate AgenticLens metrics from Codex session data, with OpenCode extraction as the parity target.

## Research deliverables

- Review AgenticLens extraction, live collection, domain facts, ledger derivation, and Git analysis.
- Inspect local CODEX_HOME rollout files and SQLite databases read-only; identify usable metadata, projections, diagnostic logs, and sensitive/non-metric stores.
- Compare Codex rollout history, opt-in rollout traces, `codex exec --json`, and app-server events for metric coverage and privacy.
- Implement a fixture-backed Codex rollout adapter and CLI integration that generates a schema-v3 `agentic_metrics.json` ledger.
- Keep the checked-in parser fixture synthetic, but generate the session ledger from real local Codex data; exclude transcript/tool payloads and distinguish observed from unavailable metrics.
- Keep real-session validation ledgers and generated summary tables local and gitignored; publish only synthetic fixtures and sanitized research findings.
- Keep all ticket materials in `alspaces/AL-56`; maintain one research plan and no iteration folders.

## Scope limits

- The CLI supports explicit Codex-source discovery across `CODEX_HOME` active and archived rollout directories, including gzip archives and child threads; it does not yet tail active sessions.
- The session ledger uses a real local rollout; a nonce lookup in the projected history database correlates its thread ID with this interaction (see `DATA_PROVENANCE.md`).
- The session ledger has no Codex-billed USD cost. Schema-v3 requires numeric cost, so zero represents unavailable here and must be disclosed.
- Git-derived code facts are not inferred from Codex transcript/file-change content and are joined through the existing Git analyzer.
- No prompts, assistant text, reasoning content, command arguments, or tool input/output are returned or saved.

## Verification

- Parser tests cover token/cache semantics, tool lifecycle/error projection, unknown/malformed records, and privacy filtering.
- CLI tests cover discovery, repeated generation, and model changes. The existing dashboard consumes the generated ledger.
