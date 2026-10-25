# AL-56 Research Contract

## Goal

Find and validate the strongest practical way to generate AgenticLens metrics from Codex session data, with OpenCode extraction as the parity target.

## Research deliverables

- Review AgenticLens extraction, live collection, domain facts, ledger derivation, and Git analysis.
- Inspect local CODEX_HOME rollout files and SQLite databases read-only; identify usable metadata, projections, diagnostic logs, and sensitive/non-metric stores.
- Compare Codex rollout history, opt-in rollout traces, `codex exec --json`, and app-server events for metric coverage and privacy.
- Implement a fixture-backed Codex rollout adapter and CLI integration that generates a schema-v3 `agentic_metrics.json` ledger.
- Keep the checked-in parser fixture synthetic, but generate the session ledger from real local Codex data; exclude transcript/tool payloads and distinguish observed from unavailable metrics.
- Commit the real-session `agentic_metrics.json` and generate its metrics summary in `SUMMARY.md` as lifecycle deliverables. Keep raw rollouts local and test fixtures synthetic.
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

## Repository independence

Following FIX-4, source code must receive ticket identity and workspace layout through user input or configuration. Research ticket IDs are tracking metadata only; renaming a ticket must not require application or test changes. Tests use stable synthetic identifiers and fixtures under `testdata` or `t.TempDir()`. CLI tests must run from their own temporary project with an explicit fixture configuration, never inherit the development checkout configuration or live ticket directories.

## Code and iteration metrics

`--code-ref` explicitly selects committed Git work independently of ticket or branch naming. Compare its merge base with the configured base branch, filter by configured source folders and classification patterns, and reject invalid refs without falling back to cached facts. An empty explicit diff is zero, not stale cached code. Codex SDLC iterations count root user messages, preferring the semantic user-message event stream over duplicate response-item representations; child delegation and lifecycle/model events do not increment this count.

## Configured file classification

Test directory/name and production extension heuristics are prohibited. Classification uses `code_metrics.test_patterns`, `prod_patterns`, and `include` from project configuration, with workspace/exclude precedence and unmatched files left as Other. Slash-containing globs are repository-relative; `**` matches complete directory components. Both code-ref and automatic ticket discovery filter totals using configured source roots and inclusion patterns, including test fixtures outside source roots.

## Child rollout attribution

Codex child rollouts must retain their own first session identity and explicit or nested parent linkage. Object-valued subagent source metadata is supported. Inherited parent history must not overwrite child identity or contribute parent token, event, duration, or turn facts to the child. When role metadata is absent, derive only the final persona component from `agent_path` as a raw source role, matching OpenCode’s extraction of `session.agent`. Do not resolve project `role_mappings` in Codex extraction. Explicit role metadata takes precedence; absent persona data remains UnassignedSubAgent, and the full path stays private. Synthetic adapter and isolated CLI regressions must verify distinct root/child identities, child-only usage, and idempotent regeneration before validating the real probe.
