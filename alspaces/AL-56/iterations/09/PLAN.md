# AL-56 Iteration 09 Plan: Enforce the v4 Legacy Ingestion Boundary

## Objective

Make schema v4's structural incompatibility explicit at runtime. Ordinary existing-ledger loading for `generate` and server/API paths must reject schema v1-v3 without parsing or merging them. Retain a supported explicit schema migration followed by source-backed backfill before runtime ingestion.

This plan supersedes the older §6 `LoadExistingFactLedger` contract promise that runtime loads parse and merge v1-v3. That promise conflicts with the requested breaking compatibility change; the explicit migration/backfill workflow is the supported compatibility path.

## Current Implementation Grounding

- `backend/internal/adapters/repository.go:LoadExistingFactLedger` currently unmarshals canonical session records and legacy root-level invocation/tool/event records into the same domain result. It does not gate on schema version before parsing.
- `backend/cmd/agentic-lens/main.go` calls this loader during `generate` and currently contains a warning for schema versions below 4, indicating that legacy data can reach runtime ingestion.
- `backend/internal/adapters/backfill/migration_runner.go` and `backend/internal/domain/migration/engine.go` provide an explicit shape migration path. The backfill engine separately supports source selection/recomputation, but migration and source extraction need clear sequencing and readiness semantics.
- Server/API ledger loading must be traced to ensure it does not bypass the same schema gate or expose partially interpreted legacy data.

## Design and Work Scope

### Loader and generate boundary

- Gate on `schema_version` before decoding legacy or canonical fact collections. Reject v1, v2, and v3 with a typed/actionable error naming the detected version and directing the user to schema migration followed by source-backed backfill.
- Fail closed for missing/invalid versions where the payload shape is ambiguous or cannot be proven canonical v4. Do not mutate the file or return partial facts.
- Accept and validate canonical v4 using the existing domain/DTO validation. Preserve current v4 accumulation semantics and OpenCode default-source behavior.
- Remove any warning-and-continue behavior that permits v1-v3 records to merge in ordinary `generate` ingestion.

### Server/API and runtime behavior

- Find all existing-ledger reads used by server/API endpoints and route them through the same v4-only validation boundary, or apply equivalent boundary validation at the adapter edge.
- Return a clear client-visible error for legacy or incomplete ledgers. Avoid presenting an empty/partially parsed v4 ledger as a successful response.
- Keep API response types canonical v4; the frontend does not parse historical shapes.

### Explicit migration and source-backed backfill

- Preserve support for explicitly invoking schema migration on v1-v3. Migration is an operator-directed transformation and must not be called implicitly by loader/server/generate.
- Require a source-backed backfill after schema migration and before the resulting file is accepted by runtime ingestion. The workflow must know the harness source (`opencode`, `codex`, or explicit supported equivalent); do not infer source from opaque IDs.
- Define a completion/readiness signal so shape-migrated but not source-backed or otherwise incomplete ledgers are rejected by runtime ingestion. Backfill must mark/report unavailable source evidence as incomplete rather than claiming readiness.
- Preserve existing dry-run, backup, ticket-scoping, atomic-write, and Git-fact preservation behavior. Migration/backfill errors must not destroy the original ledger.
- Keep the migration implementation capable of reading legacy layouts only inside the explicit migration/backfill path.

## Acceptance Criteria

- Ordinary `LoadExistingFactLedger` rejects fixtures for each of schema v1, v2, and v3 before converting root-level facts; the error contains the version and explicit remediation sequence.
- `generate` against an existing v1-v3 ledger exits unsuccessfully, writes no merged/new ledger, and does not append a summary as if ingestion succeeded.
- Server/API endpoints backed by existing ledgers reject legacy and incomplete ledgers consistently, with an actionable status/error and no silent partial response.
- Canonical, complete v4 ledgers continue to load in generate and server/API; existing OpenCode extraction and defaults remain unchanged.
- Explicit migration can still transform supported v1-v3 inputs. A following source-backed backfill produces a validated, runtime-ready canonical v4 ledger when source evidence is available.
- Migration-only, missing-source, source-read failure, and incomplete-backfill outcomes remain ineligible for ordinary ingestion and are reported clearly.
- Existing migration dry-run/backup/scope protections and preservation of Git facts are retained.
- Tests cover loader, CLI/generate, server/API, migration, and backfill boundaries, including no-write/no-summary side effects on rejection.

## Files and Components to Inspect/Change During Implementation

- `backend/internal/adapters/repository.go` and repository DTO validation.
- `backend/cmd/agentic-lens/main.go` existing-ledger load/error handling.
- Server/API ledger repository and handlers under `backend/internal/adapters` and `backend/cmd`.
- `backend/internal/adapters/backfill/migration_runner.go`, backfill orchestration, and `backend/internal/domain/migration/engine.go` for explicit upgrade and readiness semantics.
- Related adapter, CLI, server/API, migration, and backfill tests.
- `alspaces/AL-56/CONTRACT.md` is the controlling acceptance boundary; update workspace summary/PR and final scoped ledger only during implementation delivery.

## Out of Scope

- Removing the v1-v3 migration chain or forcing users to discard historical data.
- Introducing new ledger content unrelated to migration readiness or v4 validation.
- Changing Codex/OpenCode extraction fidelity, data privacy constraints, or Git classification rules beyond what is required to source-backfill and validate v4.
- Workspace-wide migration/backfill during this ticket's final telemetry generation; keep any final ticket-ledger update scoped to AL-56.
