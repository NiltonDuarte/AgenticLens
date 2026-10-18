# Pull Request: AL-15 Universal Multi-Version Schema Migration Engine

## Summary
Implements a universal multi-version schema migration engine for AgenticLens telemetry ledgers in Go. Supports deterministic upgrade path planning across arbitrary schema versions (e.g. unversioned/v1 to v2), invariant validation ensuring zero data corruption or loss, atomic disk writes with backup creation, a dedicated CLI `migrate` subcommand, integration with the existing `backfill` engine via `--migrate-schema`, structured root CLI `--help` discovery across all subcommands, and comprehensive documentation in `README.md`.

For complete architectural details, see `workspaces/AL-15/CONTRACT.md` and `workspaces/AL-15/SUMMARY.md`.

## Architecture & Changes

### 1. Pure Domain Migration Core (`backend/internal/domain/migration`)
- `types.go`: Defines `MigrationStep` interface and sentinel error types (`ErrUnsupportedVersion`, `ErrDowngradeNotAllowed`, `ErrMissingStep`, `ErrDuplicateStep`, `ErrInvariantViolation`, `ErrMalformedJSON`).
- `registry.go`: Thread-safe migration step registry with topological upgrade planning (`PlanChain(from, to)`).
- `validator.go`: Pure `InvariantValidator` verifying that migrations preserve ticket ID, timestamps, code facts, tool invocations, and historical event facts.
- `step_v1_v2.go`: Baseline `StepV1ToV2` migrating unversioned/v1 ledgers to schema version 2 by initializing an empty `events: []` slice and schema version tag.
- `engine.go`: Coordinated `Engine` with version sniffing, step execution, and invariant enforcement.

### 2. Imperative Shell & Adapters (`backend/internal/adapters/backfill`)
- `options.go` & `report.go`: Added `MigrationOptions`, `MigrateSchema` flag on `BackfillOptions`, and structured migration reporting.
- `migration_runner.go`: Implemented `MigrationRunner` scanning workspace directories, executing migrations, atomically swapping files via temporary files (`.tmp`), creating backups (`.bak`), and optionally re-rendering `SUMMARY.md` markdown tables.
- `engine.go`: Integrated schema migration step into `backfill.Engine` when `--migrate-schema` is enabled.
- `schemas.go`: Ensured schema v2 empty events slices serialize cleanly as `"events": []`.

### 3. CLI Subcommand & Documentation (`backend/cmd/agentic-lens`, `README.md`)
- Added `agentic-lens migrate` subcommand supporting flags: `--workspaces`, `--ticket-id`, `--target-version`, `--dry-run`, `--backup`, and `--update-summaries`.
- Added `--migrate-schema` flag to `agentic-lens backfill`.
- Implemented structured root usage `printRootUsage()` listing all 4 subcommands (`server`, `generate`, `backfill`, `migrate`) and root inspection flags.
- Intercepted `help`, `-h`, and `--help` for clear discoverability and exit code 0.
- Expanded `README.md` with detailed sections and examples for both `backfill` and `migrate`.
