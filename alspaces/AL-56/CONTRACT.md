# AL-56 Research & OpenCode Compatibility Contract

## Goal

Provide first-class Codex session discovery and telemetry extraction in AgenticLens while strictly preserving backward compatibility, extraction fidelity, and non-breakage across existing OpenCode workflows, historical schemas, CLI conventions, and the TypeScript dashboard.

## Research & Parity Deliverables

- Review AgenticLens extraction, live collection, domain facts, ledger derivation, and Git analysis.
- Inspect local CODEX_HOME rollout files and SQLite databases read-only; identify usable metadata, projections, diagnostic logs, and sensitive/non-metric stores.
- Compare Codex rollout history, opt-in rollout traces, codex exec --json, and app-server events for metric coverage and privacy.
- Implement a fixture-backed Codex rollout adapter and CLI integration that generates a schema-v4 agentic_metrics.json ledger.
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
   - When running in detached states, shallow clones, or repositories where an upstream base branch cannot be resolved, existing code facts in historical ledgers are preserved via cached fallback with an explicit stderr warning, while uninitialized runs abort with an explicit non-zero exit.
4. Strict Zero-Fallback File Classification Policy:
   - Zero runtime fallback heuristics in backend/internal/adapters/git/classifier.go.
   - When test_patterns or prod_patterns are unconfigured or empty in .agenticlens-config.yaml, the classifier must not guess or apply default language conventions.
   - The engine emits an explicit warning notifying the user that pattern lists are empty. Unmatched files are categorized strictly as other.
   - Sensible defaults belong exclusively to the configuration generation template in backend/internal/adapters/config/writer.go during agentic-lens init, ensuring users own and inspect their configuration explicitly on disk.
5. Backfill Safety Guardrails:
   - Recomputing Git code facts via --recompute-git strictly requires an explicit --ticket-id and an explicit --code-ref.
   - Batch execution of backfill across multi-ticket workspaces without --ticket-id must reject --recompute-git to prevent applying a single Git commit diff across historical tickets.
   - Schema migration (--migrate-schema) and database recomputation (--recompute-db) without --recompute-git must preserve existing Git code facts without alteration.
6. Non-Destructive Ledger Loader:
   - backend/internal/adapters/repository.go (LoadExistingFactLedger) parses both current schema-v4 atomic ledgers and legacy schemas (v1, v2, v3) to preserve historical code facts, root session links, and event streams during incremental runs.
7. Frontend Transparency:
   - All emitted ledgers adhere to schema-v4 domain models in backend/internal/domain/models.go.
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

--code-ref defaults to HEAD and is required for every Git code comparison, independently of ticket or branch naming. Compare its merge base with the configured base branch, filter by configured source folders and classification patterns, and reject missing or invalid refs without falling back to ticket-based discovery or cached code facts when generating a new ledger. An empty explicit diff is zero. Codex SDLC iterations count root user messages, preferring the semantic user-message event stream over duplicate response-item representations; child delegation and lifecycle/model events do not increment this count.

## Configured File Classification

Test directory/name and production extension heuristics must not silently override or substitute for configuration. Classification uses code_metrics.test_patterns, prod_patterns, and include from project configuration, with workspace/exclude precedence. When explicit configuration arrays are empty or unconfigured, no fallback patterns are applied at runtime; a warning is emitted, and unmatched source files are categorized as other. Init command scaffolding populates default patterns in the written configuration file so the user explicitly owns them. Slash-containing globs are repository-relative; double-asterisk matches complete directory components.

## Child Rollout Attribution

Codex child rollouts must retain their own first session identity and explicit or nested parent linkage. Object-valued subagent source metadata is supported. Inherited parent history must not overwrite child identity or contribute parent token, event, duration, or turn facts to the child. When role metadata is absent, derive only the final persona component from agent_path as a raw source role, matching OpenCode's extraction of session.agent. Do not resolve project role_mappings in Codex extraction. Explicit role metadata takes precedence; absent persona data remains UnassignedSubAgent, and the full path stays private.

## Cross-Device SDLC Iterations & Session Accumulation

1. Platform Adherence to Database and Configuration:
   - Telemetry extraction adheres strictly to database facts (SQLite, rollout transcripts) and project configuration.
   - The platform strictly forbids inspection or file traversal of workspace iteration directories (<workspaces_dir>/<ticket_id>/iterations/). Ticket workspaces are project-level organizational conventions, not platform-enforced data stores.
2. Invocations-Level Iteration Attribution:
   - Each invocation record (InvocationFact) carries an explicit user_iterations counter representing the number of human steering turns directed to that root session.
   - For multi-model root sessions, user_iterations is assigned exclusively to the primary invocation (index 0) to guarantee zero double-counting across model splits.
   - Child/subagent invocations (non-nil parent_id) carry a user_iterations count of 0.
3. Multi-Device Invocations Merging:
   - Telemetry generation must never blindly overwrite cached invocations with local session extractions.
   - When generating facts on a new device where local database state only contains the local session, existing invocations from cached fact ledgers are merged by invocation ID and model identity.
   - Existing invocations from previous devices or earlier sessions absent from the local database are preserved in their entirety.
4. Strict SDLC Iterations Summation (Zero Fallback Policy):
   - Total code.sdlc_iterations is strictly derived as the exact mathematical sum of user_iterations across all merged invocations.
   - Zero fallback heuristics or guessing: the platform does not fall back to loose steering turn counts, cached code.sdlc_iterations, or arbitrary defaults.
   - If an existing ledger has a schema version prior to 4 or unpopulated user_iterations, fallback is explicitly rejected and migration is required via migrate-schema or backfill.
5. Schema Version 4 Elevation:
   - Schema version is bumped to 4 (CurrentSchemaVersion = 4) to formally incorporate atomic user_iterations on InvocationFact.
   - Historical ledgers created under schema versions 1, 2, and 3 require migration to schema version 4 via migrate-schema or backfill to populate user_iterations and align with the strict schema contract.
