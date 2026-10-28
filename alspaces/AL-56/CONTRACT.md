# AL-56 Research & OpenCode Compatibility Contract

## Goal

Provide first-class Codex session discovery and telemetry extraction in AgenticLens while preserving OpenCode workflow behavior and an explicit upgrade path for historical ledgers. Schema v4 has a structurally incompatible ledger shape: ordinary runtime ingestion must not interpret schema v1-v3 as v4.

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
6. Explicit Legacy Upgrade Boundary (supersedes the prior non-destructive loader invariant):
   - `LoadExistingFactLedger` and every ordinary server/API/generate ingestion path accept only the current canonical schema-v4 shape. They reject schema v1-v3 with an actionable error directing the operator to the explicit migration and source-backed backfill workflow; they must not parse, merge, normalize, or silently upgrade legacy shapes.
   - Preserve a supported explicit migration path for v1-v3. The operator first runs schema migration, then runs source-backed backfill for the ledger's declared/selected harness to recover facts that a shape-only transform cannot reconstruct. Migration alone does not make a legacy ledger runtime-ingestion-ready.
   - Backfill/migration may read legacy shapes only inside their explicit workflow. Preserve recoverable facts and report unrecoverable/missing source data; do not claim successful backfill when source evidence is unavailable. Keep existing backup, dry-run, ticket scoping, and Git-fact preservation guardrails.
   - After successful source-backed backfill, persist and validate canonical v4, including required v4 atomic/session-owned fields. Ordinary ingestion then accepts the ledger. Malformed, ambiguous, unsupported, or still-incomplete ledgers fail closed without modifying the source file.
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

## AL-56 v4 Compatibility Change

The earlier invariant in §6 promising that `LoadExistingFactLedger` parses v1-v3 directly conflicts with the requested v4 compatibility boundary and is superseded by the explicit legacy-upgrade invariant above. This is an intentional breaking change to ordinary ingestion. Supported remediation is explicit schema migration followed by source-backed backfill; do not remove or strand that operator path. Server/API reads and `generate` must surface the same actionable legacy-schema rejection, and must not partially merge old facts before failing.

### Acceptance Criteria

- Ordinary server, API, and `generate` existing-ledger loads reject schema v1, v2, and v3 before parsing or merging legacy fact collections; errors identify the detected version and direct operators to the supported migration plus source-backed backfill sequence.
- Current canonical v4 loads continue to work, including existing OpenCode workflows.
- Explicit migration remains able to transform supported v1-v3 layouts, with existing dry-run/backup/scope safeguards.
- A subsequent source-backed backfill can reconstruct source-available facts, report unavailable source data as incomplete/failure, and produce a validated canonical v4 ledger; no ordinary runtime ingestion proceeds on a migrated-only or incomplete ledger.
- API/server behavior exposes the rejection clearly and consistently without rewriting the ledger. The frontend receives canonical v4 responses only and needs no legacy-shape parser.
- Regression coverage exercises each legacy version through ordinary loader, generate, and server/API boundaries, plus successful explicit migrate-then-backfill and failure/incomplete-source cases.

## AL-56 Migration Metadata Lifecycle (Iteration 11)

This section supersedes iteration 10's requirement for a durable universal `source_backfill` readiness marker and any contract wording that requires completed v4 ledgers to retain migration metadata.

- Ordinary runtime accepts a structurally canonical schema-v4 ledger when migration metadata is absent. Canonical v4 structure is the runtime eligibility criterion for completed ledgers; there is no additional persisted `source_backfill` marker requirement.
- Migration metadata is temporary workflow state. It may be persisted while a multi-step migration or source-backed backfill is in progress so the operator can resume it on another system or in a later step.
- Any persisted migration-in-progress or incomplete state makes the ledger ineligible for ordinary runtime loading. Runtime rejects it before returning or merging facts and directs the operator to resume the explicit migration/backfill workflow.
- The successful final migration/backfill step validates and persists canonical v4, then removes all migration metadata before reporting completion. A ledger must not become runtime-ready while an in-progress or incomplete marker remains.
- Schema v1-v3 remain rejected by ordinary runtime before legacy fact parsing. Their supported route remains explicit schema migration followed by source-backed backfill.

### Iteration 11 Acceptance Criteria

- Ordinary runtime accepts a structurally canonical schema-v4 ledger with no migration metadata, including existing completed v4 ledgers.
- Ordinary runtime rejects schema v1, v2, and v3 before parsing legacy facts, with detected version and explicit migration plus source-backed backfill guidance.
- Ordinary runtime rejects any persisted migration-in-progress or incomplete state; the explicit migration workflow can resume that state across systems/steps.
- Successful completion of the final migration/backfill step removes migration metadata before the ledger is reported as runtime-ready.
- No custom `source_backfill` marker remains as a universal eligibility requirement for completed v4 ledgers.
- Tests cover completed v4 without metadata, rejection and resumption of intermediate migration state, deletion of metadata on successful final migration, and v1-v3 pre-parse rejection.
