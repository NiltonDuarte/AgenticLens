# Implementation Plan: AL-15 Universal Multi-Version Schema Migration Engine for Telemetry Ledgers

## 1. Goal and Iteration Scope

Deliver the complete universal stepwise schema migration engine in Go:
- Create a pure functional domain package `internal/domain/migration/` providing the `MigrationStep` interface, registry, step chain planner, invariant validator, and `StepV1ToV2`.
- Create the imperative migration runner in `internal/adapters/backfill/` to discover workspace ledgers, detect schema versions, orchestrate domain migrations, generate backups, and execute atomic file writes.
- Add the `agentic-lens migrate` subcommand to `cmd/agentic-lens/main.go` and wire the `--migrate-schema` flag into `agentic-lens backfill`.
- Deliver comprehensive unit, invariant, and integration test suites covering all edge cases, corrupted files, and race conditions.

---

## 2. Grounded Reference Architecture and Concrete Paths

### Domain Core Layer (`backend/internal/domain/`)
- `backend/internal/domain/constants.go`:
  Reference `CurrentSchemaVersion = 2`. Add error definitions and step identifiers where appropriate.
- `backend/internal/domain/models.go`:
  Reference `AgenticFactLedger`, `MigrationInfo`, `EventFact`, `CodeFacts`, `InvocationFact`, `ToolFact`.
- `backend/internal/domain/enums.go`:
  Reference `MigrationStatus`.
- `backend/internal/domain/migration.go`:
  Existing multi-machine status derivation from AL-12 (`EvaluateMigrationStatus`).
- `backend/internal/domain/migration/types.go` (new):
  Define `MigrationStep` interface, step metadata, and migration errors (`ErrUnsupportedVersion`, `ErrMissingStep`, `ErrInvariantViolation`, `ErrMalformedJSON`, `ErrDowngradeNotAllowed`).
- `backend/internal/domain/migration/registry.go` (new):
  Define `Registry` struct, registration logic, transition validation, and `PlanChain(sourceVersion, targetVersion)`.
- `backend/internal/domain/migration/validator.go` (new):
  Define `InvariantValidator` verifying non-destructive transformation between pre-step and post-step ledgers.
- `backend/internal/domain/migration/step_v1_v2.go` (new):
  Implement `StepV1ToV2` upgrading unversioned v1 ledgers to version 2 with schema version marker, baseline migration metadata, and empty non-nil events slice.
- `backend/internal/domain/migration/engine.go` (new):
  Define `Engine` coordinating version detection from raw JSON bytes, step execution, and invariant enforcement.
- `backend/internal/domain/migration/engine_test.go` (new):
  Pure domain tests for version detection, chain discovery, step transitions, invariant failures, and idempotency.

### Imperative Adapters Layer (`backend/internal/adapters/`)
- `backend/internal/adapters/schemas.go`:
  Reference `AgenticFactLedgerDTO`, `AgenticFactLedgerDTOFromDomain`, and JSON mapping methods.
- `backend/internal/adapters/repository.go`:
  Reference `ResolveWorkspacesDir` and `ScanWorkspaces`.
- `backend/internal/adapters/backfill/options.go`:
  Extend `BackfillOptions` with `MigrateSchema bool`. Add `MigrationOptions` struct for standalone migration.
- `backend/internal/adapters/backfill/report.go`:
  Add `MigrationStepResult` and `MigrationReport` structs for structured reporting.
- `backend/internal/adapters/backfill/migration_runner.go` (new):
  Implement `MigrationRunner` scanning workspaces, detecting schema versions, executing domain migration, managing `.bak` backups, writing temporary files, renaming atomically, and updating summaries.
- `backend/internal/adapters/backfill/migration_runner_test.go` (new):
  Tests for batch migration, dry-run preservation, backup file generation, idempotency, and corrupted file resilience.
- `backend/internal/adapters/backfill/engine.go`:
  Integrate `MigrationRunner` into `BackfillEngine.Execute` when `MigrateSchema` is true.

### CLI Application Layer (`backend/cmd/agentic-lens/`)
- `backend/cmd/agentic-lens/main.go`:
  Add `runMigrate(args []string)` handling `agentic-lens migrate` flags (`--workspaces`, `--ticket-id`, `--target-version`, `--dry-run`, `--backup`, `--update-summaries`). Add `--migrate-schema` flag to `runBackfill`.
- `backend/cmd/agentic-lens/main_test.go`:
  Integration tests for `migrate` subcommand flags, dry-run output, and error conditions.

### TypeScript Presentation Layer (`web/src/domain/`)
- `web/src/domain/types.ts`:
  Verify TypeScript domain types align with versioned ledger schemas.

---

## 3. Phased Implementation Tasks

### Phase 1: Pure Domain Migration Framework (`internal/domain/migration/`)
- Task 1.1: Create `backend/internal/domain/migration/types.go`:
  - Define `MigrationStep` interface:
    - `SourceVersion() int`: Returns integer version step upgrades from.
    - `TargetVersion() int`: Returns integer version step upgrades to.
    - `Description() string`: Returns human-readable summary of transformation.
    - `Apply(ledger domain.AgenticFactLedger) (domain.AgenticFactLedger, error)`: Performs pure transformation.
  - Define sentinel domain errors:
    - `ErrUnsupportedVersion`: Raised when target or source version is outside supported range.
    - `ErrDowngradeNotAllowed`: Raised when target version is less than source version.
    - `ErrMissingStep`: Raised when no step transition exists between two contiguous versions.
    - `ErrDuplicateStep`: Raised when registering a step for an already registered version transition.
    - `ErrInvariantViolation`: Raised when pre-step and post-step data do not satisfy non-destructive invariants.
    - `ErrMalformedJSON`: Raised when raw bytes cannot be inspected for schema version.
- Task 1.2: Create `backend/internal/domain/migration/registry.go`:
  - Implement `Registry` struct managing a thread-safe map of version transitions keyed by source version.
  - Implement `NewRegistry()` constructor.
  - Implement `Register(step MigrationStep) error`:
    - Validates that `step.TargetVersion() == step.SourceVersion() + 1`.
    - Returns `ErrDuplicateStep` if a step for that source version already exists.
  - Implement `PlanChain(sourceVersion int, targetVersion int) ([]MigrationStep, error)`:
    - If `sourceVersion == targetVersion`, return empty slice (no-op).
    - If `sourceVersion > targetVersion`, return `ErrDowngradeNotAllowed`.
    - Iterate from `sourceVersion` to `targetVersion - 1`, resolving each step. If any step is missing, return `ErrMissingStep`.
- Task 1.3: Create `backend/internal/domain/migration/validator.go`:
  - Implement `InvariantValidator` struct.
  - Implement `Validate(before domain.AgenticFactLedger, after domain.AgenticFactLedger, expectedTargetVersion int) error`:
    - Assert `after.SchemaVersion != nil` and `*after.SchemaVersion == expectedTargetVersion`.
    - Assert `TicketID`, `TicketTitle`, and `TicketType` pointer values match.
    - Assert `StartedAt` and `CompletedAt` pointer values match.
    - Assert `CodeFacts` fields (`LinesAdded`, `LinesDeleted`, `TestLinesAdded`, `TestLinesDeleted`, `ProdLinesAdded`, `ProdLinesDeleted`, `FilesChanged`, `SDLCIterations`) match identically.
    - Assert `len(after.Invocations) == len(before.Invocations)`.
    - Assert each invocation in `after.Invocations` matches `before.Invocations` on `ID`, `Role`, `Model`, `Turns`, `ActiveDurationSeconds`, `CostUSD`, `Tokens`, and `Compactions`.
    - Assert `len(after.Tools) == len(before.Tools)`.
    - Assert each tool in `after.Tools` matches `before.Tools` on `Name`, `Calls`, and `Errors`.
    - Assert `len(after.Events) >= len(before.Events)`. If `before.Events` was non-empty, assert all before events match after events in identical order.
    - If any check fails, return descriptive `ErrInvariantViolation`.

### Phase 2: Step V1 to V2 and Engine Implementation (`internal/domain/migration/`)
- Task 2.1: Create `backend/internal/domain/migration/step_v1_v2.go`:
  - Implement `StepV1ToV2` struct.
  - `SourceVersion()` returns 1.
  - `TargetVersion()` returns 2.
  - `Description()` returns "Upgrades unversioned or v1 ledgers to schema version 2 with events array and migration metadata".
  - `Apply(ledger domain.AgenticFactLedger)`:
    - Clone ledger value.
    - Set `schemaVersion := 2; ledger.SchemaVersion = &schemaVersion`.
    - If `ledger.Events == nil`, set `ledger.Events = []domain.EventFact{}`.
    - If `ledger.Migration == nil`, initialize baseline `MigrationInfo` setting `SchemaVersion: 2`, `Status: domain.MigrationStatusCompleted` (or preserve partial status if pending steps exist), `LastMigratedBy: "universal-migration-engine"`.
    - Return upgraded ledger.
- Task 2.2: Create `backend/internal/domain/migration/engine.go`:
  - Implement `Engine` struct containing `registry *Registry` and `validator *InvariantValidator`.
  - Implement `NewEngine(registry *Registry) *Engine`:
    - Defaults to standard `InvariantValidator`.
  - Implement `DetectVersion(input []byte) (int, error)`:
    - Unmarshals minimal JSON struct with `SchemaVersion *int`.
    - If unmarshal fails, returns `ErrMalformedJSON`.
    - If `SchemaVersion == nil || *SchemaVersion <= 1`, returns 1.
    - Otherwise returns `*SchemaVersion`.
  - Implement `Migrate(input []byte, targetVersion int) (domain.AgenticFactLedger, error)`:
    - Calls `DetectVersion(input)`.
    - Unmarshals `input` into `domain.AgenticFactLedger` using standard JSON unmarshaling into a clean domain struct.
    - If `sourceVersion == targetVersion`: returns unmarshaled ledger with zero mutations.
    - Resolves chain via `registry.PlanChain(sourceVersion, targetVersion)`.
    - Loops through steps, invoking `step.Apply(current)` and validating via `validator.Validate(before, after, step.TargetVersion())`.
    - Returns final migrated ledger.
- Task 2.3: Implement `backend/internal/domain/migration/engine_test.go`:
  - Table-driven tests for version detection: unversioned JSON, explicit version 1, explicit version 2, and malformed JSON.
  - Tests for registry: valid registration, duplicate registration rejection, invalid non-contiguous step rejection.
  - Tests for plan chain: single step, multi-step chain, identical version no-op, downgrade error.
  - Tests for `StepV1ToV2`: field preservation, non-nil events slice initialization, schema version 2 assignment.
  - Tests for invariant validator: triggers error when invocations dropped, when tools modified, when code facts corrupted, or when target version mismatched.
  - Invariant idempotency test: verifying migrating an already v2 ledger returns zero mutations.

### Phase 3: Imperative Boundary Adapter (`internal/adapters/backfill/`)
- Task 3.1: Extend `backend/internal/adapters/backfill/options.go`:
  - Add `MigrateSchema bool` to `BackfillOptions`.
  - Define `MigrationOptions` struct:
    - `WorkspacesDir string`
    - `TicketID string`
    - `TargetVersion int`
    - `DryRun bool`
    - `Backup bool`
    - `UpdateSummaries bool`
- Task 3.2: Extend `backend/internal/adapters/backfill/report.go`:
  - Define `TicketSchemaMigrationResult`:
    - `TicketID string`
    - `WorkspacePath string`
    - `SourceVersion int`
    - `TargetVersion int`
    - `Status string` (e.g. "migrated", "up-to-date", "skipped", "error")
    - `BackupCreated bool`
    - `DiffSummary string`
    - `Error error`
  - Define `SchemaMigrationReport`:
    - `Scanned int`
    - `Migrated int`
    - `UpToDate int`
    - `Skipped int`
    - `Errors []string`
    - `Results []TicketSchemaMigrationResult`
- Task 3.3: Implement `backend/internal/adapters/backfill/migration_runner.go`:
  - Implement `MigrationRunner` struct holding `engine *migration.Engine`.
  - Implement `NewMigrationRunner(engine *migration.Engine) *MigrationRunner`:
    - If engine is nil, instantiates default engine registered with `StepV1ToV2`.
  - Implement `Run(opts MigrationOptions) (*SchemaMigrationReport, error)`:
    - Resolves workspace directory using `adapters.ResolveWorkspacesDir(opts.WorkspacesDir)`.
    - Discovers all `agentic_metrics.json` files via `filepath.WalkDir`.
    - Filters by `opts.TicketID` if specified.
    - Iterates files with per-workspace error isolation:
      - Reads file content via `os.ReadFile`.
      - Detects source version using `engine.DetectVersion`.
      - Checks if `sourceVersion >= targetVersion`: records status "up-to-date", increments `UpToDate`, and continues.
      - Calls `engine.Migrate(content, opts.TargetVersion)`.
      - If dry-run: records status "migrated (dry-run)", describes version change and field additions, does not mutate disk.
      - If write mode:
        - If `opts.Backup`: copies file to `agentic_metrics.json.bak` using `os.WriteFile`.
        - Converts migrated ledger to `AgenticFactLedgerDTO` via `adapters.AgenticFactLedgerDTOFromDomain`.
        - Serializes to formatted JSON with two spaces and trailing newline.
        - Writes to temporary file `agentic_metrics.json.tmp`.
        - Atomically renames temporary file to `agentic_metrics.json`.
        - If `opts.UpdateSummaries`: derives `domain.DeriveMetricsLedger(migrated)` and calls `adapters.UpdateSummaryFile`.
        - Records status "migrated".
    - Returns structured report.
- Task 3.4: Integrate with `BackfillEngine.Execute` in `backend/internal/adapters/backfill/engine.go`:
  - When `opts.MigrateSchema` is true, invoke `MigrationRunner` for each processed workspace before or during backfill step evaluation.
- Task 3.5: Implement `backend/internal/adapters/backfill/migration_runner_test.go`:
  - Test batch workspace migration: creates temporary workspaces with unversioned ledgers, runs runner, asserts all files updated to version 2 with events array.
  - Test dry-run: asserts zero filesystem changes and identical byte hashes before and after dry-run.
  - Test backup creation: asserts `.bak` file created with original unmigrated content when `--backup` is specified.
  - Test resilience against corrupted JSON: creates invalid JSON files alongside valid ones, asserts valid files migrate and corrupted file produces a warning without aborting batch.
  - Test idempotence: re-running on migrated workspaces results in all files marked "up-to-date" with zero writes.

### Phase 4: CLI Integration (`cmd/agentic-lens/main.go`)
- Task 4.1: Extend subcommand routing in `backend/cmd/agentic-lens/main.go`:
  - Route `os.Args[1] == "migrate"` to `runMigrate(os.Args[2:])`.
- Task 4.2: Implement `runMigrate(args []string)`:
  - Register flags:
    - `--workspaces`: Path to root workspaces directory.
    - `--ticket-id`: Filter to specific ticket identifier.
    - `--target-version`: Target schema version (defaults to `domain.CurrentSchemaVersion`).
    - `--dry-run`: Preview planned schema migrations without mutating disk.
    - `--backup`: Create `.bak` backup copy prior to overwriting.
    - `--update-summaries`: Regenerate `SUMMARY.md` markdown metrics table.
  - Invoke `backfill.NewMigrationRunner(nil).Run(opts)`.
  - Format output:
    - Execution header with mode (`EXECUTION` vs `DRY-RUN PREVIEW`), target directory, and target schema version.
    - Table columns: `Ticket ID`, `Source Ver`, `Target Ver`, `Status`, `Backup`, `Notes`.
    - Summary line: Scanned, Migrated, Up-to-Date, Skipped, Errors.
    - Warnings list if non-fatal errors occurred.
    - Strict zero emojis.
  - Exit code: 0 on success, 1 on fatal configuration or directory errors.
- Task 4.3: Add `--migrate-schema` flag to `runBackfill` in `backend/cmd/agentic-lens/main.go`.
- Task 4.4: Add CLI tests in `backend/cmd/agentic-lens/main_test.go` covering `migrate` subcommand flag parsing, help text, and dry-run execution.

### Phase 5: TypeScript Contract Synchronization (`web/src/domain/types.ts`)
- Task 5.1: Review `web/src/domain/types.ts` to confirm complete parity with domain schema version 2 models (`schema_version`, `migration`, `events`).
- Task 5.2: Execute `npm run typecheck` in `web/` to guarantee zero frontend compiler regressions.

### Phase 6: Automated Verification and Quality Gates
- Task 6.1: Run `go vet ./...` in `backend/` to verify static analysis.
- Task 6.2: Run `go test -race ./...` in `backend/` to ensure all unit, invariant, adapter, and CLI tests pass without race conditions.
- Task 6.3: Run `make check` and `make test` from repository root.
- Task 6.4: Compile single binary executable: `make build` (or `(cd backend && go build -o ../bin/agentic-lens ./cmd/agentic-lens)`).
- Task 6.5: Verify CLI execution against existing workspaces in dry-run mode:
  - Run `./bin/agentic-lens migrate --workspaces workspaces/ --dry-run`.
  - Confirm all existing ledgers are detected, recognized as version 2, and reported as up-to-date with zero disk mutations.
- Task 6.6: Audit git diff for strict adherence to the zero-emoji and zero-any mandates.
