# AL-56 Research & OpenCode Compatibility Contract

## Goal

Provide first-class Codex session discovery and telemetry extraction in AgenticLens while strictly preserving backward compatibility, extraction fidelity, and non-breakage across existing OpenCode workflows, historical schemas, CLI conventions, and the TypeScript dashboard.

## Research & Parity Deliverables

- Review AgenticLens extraction, live collection, domain facts, ledger derivation, and Git analysis.
- Inspect local CODEX_HOME rollout files and SQLite databases read-only; identify usable metadata, projections, diagnostic logs, and sensitive/non-metric stores.
- Compare Codex rollout history, opt-in rollout traces, codex exec --json, and app-server events for metric coverage and privacy.
- Implement a fixture-backed Codex rollout adapter and CLI integration that generates a schema-v3 agentic_metrics.json ledger.
- Keep the checked-in parser fixture synthetic, but generate the session ledger from real local Codex data; exclude transcript/tool payloads and distinguish observed from unavailable metrics.
- Commit the real-session agentic_metrics.json and generate its metrics summary in SUMMARY.md as lifecycle deliverables. Keep raw rollouts local and test fixtures synthetic.
- Maintain ticket materials in alspaces/AL-56, advancing phased plans under iterations/ as implementation progresses.

## OpenCode Non-Breakage & Backward Compatibility Invariants

1. Default Source Invariant:
   - The CLI flag --source must default to "opencode".
   - Existing OpenCode pipelines, automated generation scripts, and test suites omitting --source must continue executing against OpenCode SQLite without change.
2. OpenCode SQLite Extraction Fidelity:
   - Root session resolution in backend/internal/adapters/opencode/resolver.go using exact nonce matching, explicit session ID, and ticket ID fallback must remain fully supported.
   - Fact extraction in backend/internal/adapters/opencode/extractor.go for invocations, tools, events, timestamps, and steering turns must experience zero logic modification or regressions.
3. Code Facts & Git Comparison Safety:
   - The flag --code-ref defaults to "HEAD" for committed change comparison.
   - When running in detached states, shallow clones, or repositories where an upstream base branch cannot be resolved, failure boundaries must be explicit and informative without silent corruption.
   - Existing code facts in historical ledgers must be protected against unintentional zeroing during regeneration or recomputation.
4. File Classification Fallback Discipline:
   - Projects initialized with empty test_patterns or prod_patterns in .agenticlens-config.yaml must not suffer silent metric collapse (such as test-to-production ratio decaying to 0:0).
   - Safe language-standard conventions (for Go, TypeScript, JavaScript, CSS) must apply when pattern lists are unconfigured or empty.
5. Backfill Integrity:
   - Batch execution of backfill across multi-ticket workspaces must never apply a single --code-ref diff globally across all historical tickets.
   - Recomputing Git code facts via --recompute-git must be strictly scoped to single-ticket targets (--ticket-id) or accompanied by per-ticket branch resolution to prevent historical ledger corruption.
6. Frontend Transparency:
   - All emitted ledgers adhere to schema-v3 domain models in backend/internal/domain/models.go.
   - The TypeScript frontend in web/src/ continues to consume ledgers identically regardless of whether the originating telemetry harness was OpenCode or Codex.

## Scope Limits

- The CLI supports explicit Codex-source discovery across CODEX_HOME active and archived rollout directories, including gzip archives and child threads; it does not yet tail active sessions.
- The session ledger uses a real local rollout; a nonce lookup in the projected history database correlates its thread ID with this interaction (see DATA_PROVENANCE.md).
- The session ledger has no Codex-billed USD cost. The shared schema currently emits numeric zero, which is ambiguous with free usage; a cross-cutting nullable cost representation is tracked under AL-57.
- Git-derived code facts are not inferred from Codex transcript/file-change content and are joined through the existing Git analyzer.
- No prompts, assistant text, reasoning content, command arguments, or tool input/output are returned or saved.

## Verification

- Parser tests cover token/cache semantics, tool lifecycle/error projection, unknown/malformed records, and privacy filtering.
- CLI tests cover discovery, repeated generation, and model changes.
- OpenCode regression suites verify that OpenCode generation, database queries, and backfill workflows remain completely unhindered.
- Parity test fixtures across testdata/parity_workspaces confirm identical domain rollups and formatters.

## Repository Independence

Following FIX-4, source code must receive ticket identity and workspace layout through user input or configuration. Research ticket IDs are tracking metadata only; renaming a ticket must not require application or test changes. Tests use stable synthetic identifiers and fixtures under testdata or t.TempDir(). CLI tests must run from their own temporary project with an explicit fixture configuration, never inherit the development checkout configuration or live ticket directories.

## Code and Iteration Metrics

--code-ref defaults to HEAD and is required for every Git code comparison, independently of ticket or branch naming. Compare its merge base with the configured base branch, filter by configured source folders and classification patterns, and reject missing or invalid refs without falling back to ticket-based discovery or cached code facts. An empty explicit diff is zero. Codex SDLC iterations count root user messages, preferring the semantic user-message event stream over duplicate response-item representations; child delegation and lifecycle/model events do not increment this count.

## Configured File Classification

Test directory/name and production extension heuristics must not silently override explicit configuration. Classification uses code_metrics.test_patterns, prod_patterns, and include from project configuration, with workspace/exclude precedence. When explicit configuration arrays are empty or unconfigured, standard fallback conventions must maintain classification integrity for monitored source extensions. Slash-containing globs are repository-relative; ** matches complete directory components.

## Child Rollout Attribution

Codex child rollouts must retain their own first session identity and explicit or nested parent linkage. Object-valued subagent source metadata is supported. Inherited parent history must not overwrite child identity or contribute parent token, event, duration, or turn facts to the child. When role metadata is absent, derive only the final persona component from agent_path as a raw source role, matching OpenCode's extraction of session.agent. Do not resolve project role_mappings in Codex extraction. Explicit role metadata takes precedence; absent persona data remains UnassignedSubAgent, and the full path stays private.
