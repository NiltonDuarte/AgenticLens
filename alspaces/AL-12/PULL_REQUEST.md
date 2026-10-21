# Pull Request: AL-12 Go Telemetry Backfill Engine and Distributed Multi-Machine Schema Migration

## Description
This pull request implements ticket AL-12, delivering a native Go telemetry backfill engine, schema migration coordinator, and CLI subcommand (`agentic-lens backfill`). The engine enables upgrading historical workspace ledgers (`agentic_metrics.json`) to schema version 2, recalculating Git code facts across repository branches, and safely re-querying local OpenCode SQLite session databases.

Crucially, it supports distributed multi-machine development workflows where OpenCode SQLite databases reside on separate developer machines. When a target session is absent from the host machine's local database, existing invocation and tool telemetry is preserved without data loss, marking the file with status `partial` and recording pending migration steps (e.g., `opencode_db`). The partially migrated file can be committed to Git, merged into `main`, and resumed on another developer's machine where the database records exist to complete migration.

For detailed architecture boundaries and delivery summary, refer to:
- Contract: `workspaces/AL-12/CONTRACT.md`
- Summary & Metrics: `workspaces/AL-12/SUMMARY.md`

## Summary of Changes
1. **Domain Core Layer (`internal/domain/`)**:
   - `enums.go`: Added `MigrationStatus` string enum (`pending`, `partial`, `completed`).
   - `constants.go`: Defined schema constants `CurrentSchemaVersion = 2`, `MigrationStepGitCodeFacts = "git_code_facts"`, and `MigrationStepOpenCodeDB = "opencode_db"`.
   - `models.go`: Added `MigrationInfo` struct with value semantics. Extended `AgenticFactLedger` with optional `SchemaVersion *int` and `Migration *MigrationInfo`.
   - `migration.go`: Pure function `EvaluateMigrationStatus` evaluating status transitions, step deduplication, and pending step requirements deterministically without I/O or system clocks.
   - `migration_test.go`: Unit tests for migration state transitions, step accumulation, and multi-machine resumption.
2. **Adapters Layer (`internal/adapters/`)**:
   - `schemas.go` & `repository.go`: Added `MigrationInfoDTO`, updated `AgenticFactLedgerDTO` with bidirectional domain mapping and backward compatibility.
   - `backfill/`: Created `BackfillEngine` coordinating workspace scanning, Git code recomputation, read-only SQLite session querying, non-destructive telemetry preservation, atomic file writes via temporary files, and summary table updates.
   - `backfill/engine_test.go`: Integration tests verifying dry-run filesystem immutability, multi-machine distributed migration across distinct mock databases, and corrupted file resilience.
3. **CLI Application Layer (`cmd/agentic-lens/`)**:
   - `main.go`: Wired `backfill` subcommand with flags (`--workspaces`, `--ticket-id`, `--recompute-git`, `--recompute-db`, `--dry-run`, `--update-summaries`, `--db-path`, `--force`) and structured tabular reporting.
   - `main_test.go`: Added CLI execution tests for flag handling and dry-run preview.
4. **Presentation Layer (`web/src/domain/`)**:
   - `types.ts`: Synchronized `MigrationStatus`, `MigrationInfo`, and updated `AgenticFactLedger`.

## Verification
- `go vet ./...` passed cleanly.
- `go test -race ./...` passed (all packages).
- `npm run typecheck`, `npm run lint`, and `npm test` (60 Vitest tests) passed.
- `make` built cleanly with single binary `bin/agentic-lens`.
