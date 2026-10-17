# Implementation Plan: AL-12 Go Telemetry Backfill Engine and Schema Migration Command

## 1. Goal and Iteration Scope
The goal of iteration 01 is to deliver the complete native Go telemetry backfill engine, schema migration coordinator, and CLI subcommand `agentic-lens backfill`. This includes defining domain migration models, extending boundary DTOs, building the workspace backfill orchestrator in `internal/adapters/backfill/`, wiring CLI flags and reporting in `cmd/agentic-lens/main.go`, updating TypeScript frontend interfaces in `web/src/domain/types.ts`, and validating the distributed multi-step migration lifecycle with automated tests.

---

## 2. Grounded Reference Architecture and Paths

The implementation maps to the following concrete files and packages:

### Domain Core Layer (`internal/domain/`)
- `internal/domain/enums.go`: Add `MigrationStatus` enum type with constants `MigrationStatusPending`, `MigrationStatusPartial`, and `MigrationStatusCompleted`.
- `internal/domain/constants.go`: Add schema constants `CurrentSchemaVersion = 2`, `MigrationStepGitCodeFacts = "git_code_facts"`, and `MigrationStepOpenCodeDB = "opencode_db"`.
- `internal/domain/models.go`: Add `MigrationInfo` struct. Update `AgenticFactLedger` to include `SchemaVersion int` and `Migration *MigrationInfo`.
- `internal/domain/migration.go`: Implement pure function `EvaluateMigrationStatus` computing state transitions, steps completed, and pending requirements based on operation outcomes.
- `internal/domain/migration_test.go`: Table-driven tests verifying deterministic migration state transitions and edge cases.

### Imperative Adapters Layer (`internal/adapters/`)
- `internal/adapters/schemas.go`: Add `MigrationInfoDTO`. Update `AgenticFactLedgerDTO` with `SchemaVersion` and `Migration`. Implement bidirectional domain mapping.
- `internal/adapters/backfill/options.go`: Define `BackfillOptions` struct capturing directory paths, ticket filters, recompute flags, dry-run mode, and database paths.
- `internal/adapters/backfill/report.go`: Define `BackfillReport`, `TicketMigrationResult`, and `MigrationAction` structs for structured reporting.
- `internal/adapters/backfill/engine.go`: Implement `BackfillEngine` coordinating workspace scanning, Git analysis, OpenCode SQLite querying, non-destructive fallback preservation, atomic JSON writes, and summary updates.
- `internal/adapters/backfill/engine_test.go`: Comprehensive tests covering multi-ticket processing, partial migration preservation, resumption on mock databases, corrupted file isolation, and dry-run guarantees.

### CLI Application Layer (`cmd/agentic-lens/`)
- `cmd/agentic-lens/main.go`: Add `runBackfill` subcommand handler parsing flags (`--workspaces`, `--ticket-id`, `--recompute-git`, `--recompute-db`, `--dry-run`, `--update-summaries`, `--db-path`, `--force`), executing the engine, and printing zero-emoji summary tables.
- `cmd/agentic-lens/main_test.go`: CLI integration tests verifying flag handling, dry-run mode, and exit codes.

### TypeScript Presentation Layer (`web/src/domain/`)
- `web/src/domain/types.ts`: Synchronize domain definitions with `MigrationStatus`, `MigrationInfo`, and optional `schema_version` and `migration` on `AgenticFactLedger`.

---

## 3. Phased Implementation Tasks

### Phase 1: Domain Core Models and Migration State Logic (`internal/domain/`)
- Task 1.1: Define `MigrationStatus` string type in `internal/domain/enums.go` with values:
  - `MigrationStatusPending = "pending"`
  - `MigrationStatusPartial = "partial"`
  - `MigrationStatusCompleted = "completed"`
- Task 1.2: Define schema constants in `internal/domain/constants.go`:
  - `CurrentSchemaVersion = 2`
  - `MigrationStepGitCodeFacts = "git_code_facts"`
  - `MigrationStepOpenCodeDB = "opencode_db"`
- Task 1.3: Define `MigrationInfo` struct in `internal/domain/models.go`:
  - Fields: `SchemaVersion int`, `Status MigrationStatus`, `StepsCompleted []string`, `StepsPending []string`, `LastMigratedAt string`, `LastMigratedBy string`, `Notes string`.
- Task 1.4: Extend `AgenticFactLedger` in `internal/domain/models.go`:
  - Add `SchemaVersion int` with JSON tag `schema_version,omitempty`.
  - Add `Migration *MigrationInfo` with JSON tag `migration,omitempty`.
- Task 1.5: Implement pure function `EvaluateMigrationStatus` in `internal/domain/migration.go`:
  - Accept existing migration info, boolean flags for Git recompute requested/succeeded, DB recompute requested/succeeded/attempted, session found in DB, author identity, and current timestamp string.
  - Return updated `MigrationInfo` maintaining accumulated completed steps without duplicates.
  - Transition status to `completed` if all requested or pending steps are satisfied.
  - Transition status to `partial` if DB recompute was requested but session was not found locally, preserving prior completed steps and recording pending steps.
- Task 1.6: Create `internal/domain/migration_test.go`:
  - Table-driven unit tests verifying state transitions from unversioned to partial, partial to completed, idempotency on completed ledgers, and step accumulation.

### Phase 2: Boundary DTOs and JSON Serialization (`internal/adapters/schemas.go`)
- Task 2.1: Implement `MigrationInfoDTO` in `internal/adapters/schemas.go` with JSON tags matching `MigrationInfo`. Implement `ToDomain()` and constructor from domain.
- Task 2.2: Update `AgenticFactLedgerDTO` in `internal/adapters/schemas.go`:
  - Add `SchemaVersion int` with JSON tag `schema_version,omitempty`.
  - Add `Migration *MigrationInfoDTO` with JSON tag `migration,omitempty`.
  - Update `ToDomain()` method to map `SchemaVersion` (defaulting to 1 if omitted) and `Migration` (nil if omitted).
- Task 2.3: Verify backward compatibility: ensure historical unversioned ledgers deserialize cleanly without errors or panics.

### Phase 3: Backfill Engine and Adapter Implementation (`internal/adapters/backfill/`)
- Task 3.1: Create `internal/adapters/backfill/options.go`:
  - Define `BackfillOptions` struct: `WorkspacesDir string`, `TicketID string`, `RecomputeGit bool`, `RecomputeDB bool`, `DryRun bool`, `UpdateSummaries bool`, `DBPath string`, `Force bool`.
- Task 3.2: Create `internal/adapters/backfill/report.go`:
  - Define `TicketMigrationResult`: `TicketID string`, `WorkspacePath string`, `PreviousStatus string`, `NewStatus string`, `GitUpdated bool`, `DBUpdated bool`, `DBFound bool`, `PendingSteps []string`, `Error error`.
  - Define `BackfillReport`: `Scanned int`, `Updated int`, `Skipped int`, `Partial int`, `Completed int`, `Warnings []string`, `Results []TicketMigrationResult`.
- Task 3.3: Implement `BackfillEngine` in `internal/adapters/backfill/engine.go`:
  - Step 3.3.1: Scan target workspaces directory using `adapters.ResolveWorkspacesDir`, discovering all workspace subdirectories containing `agentic_metrics.json`. Filter by `TicketID` if specified.
  - Step 3.3.2: Iterate through discovered ledgers with per-ticket error isolation. If a file cannot be read or parsed, append a warning to the report and continue to the next workspace.
  - Step 3.3.3: For each ledger, load existing facts using `adapters.LoadExistingFactLedger`.
  - Step 3.3.4: Evaluate skip conditions: If neither `--recompute-git` nor `--recompute-db` is set, and the ledger already has `schema_version == 2` with status `completed`, skip unless `--force` is specified.
  - Step 3.3.5: Execute Git recomputation when `--recompute-git` is enabled: invoke `git.ComputeCodeFacts` using the ticket ID, classification configuration, and fallback code facts. If new numstats are found, update code facts.
  - Step 3.3.6: Execute OpenCode database enrichment when `--recompute-db` is enabled:
    - Attempt connection to SQLite database via `opencode.OpenDB`.
    - Query root session by ticket ID and cached thread IDs via `opencode.ResolveActiveRootSession`.
    - If found: extract invocation facts and tool facts using `opencode.ExtractFacts`, updating telemetry.
    - If not found: retain existing invocations, tools, and duration facts from the on-disk ledger verbatim. Flag database session as missing on this host.
  - Step 3.3.7: Invoke pure domain function `domain.EvaluateMigrationStatus` to produce updated `MigrationInfo` and assign `SchemaVersion = 2`.
  - Step 3.3.8: Re-derive projected metrics via `domain.DeriveMetricsLedger`.
  - Step 3.3.9: If `--dry-run` is false:
    - Atomically write updated `agentic_metrics.json` using formatted JSON serialization.
    - If `--update-summaries` is true: generate updated Markdown table using `adapters.FormatMarkdownMetricsTable` and persist using `adapters.UpdateSummaryFile`.
  - Step 3.3.10: Record migration result in `BackfillReport`.
- Task 3.4: Implement `internal/adapters/backfill/engine_test.go`:
  - Test scanning multiple workspaces with synthetic directories.
  - Test Git recomputation refreshing code metrics while preserving invocations.
  - Test multi-machine scenario: initial run without SQLite session marks ledger as `partial` with `opencode_db` pending; second run with in-memory SQLite session advances status to `completed`.
  - Test non-fatal handling of corrupted JSON files.
  - Test dry-run mode guaranteeing zero filesystem modifications.
  - Test summary table updating.

### Phase 4: CLI Subcommand Integration (`cmd/agentic-lens/main.go`)
- Task 4.1: Extend subcommand routing in `cmd/agentic-lens/main.go` to support `backfill`.
- Task 4.2: Implement `runBackfill(args []string)` function:
  - Configure `flag.FlagSet("backfill", flag.ExitOnError)`.
  - Register flags: `--workspaces`, `--ticket-id`, `--recompute-git`, `--recompute-db`, `--dry-run`, `--update-summaries`, `--db-path`, `--force`.
  - Default `--workspaces` via `adapters.ResolveWorkspacesDir("")`.
  - Invoke `backfill.NewEngine().Execute(options)`.
  - Format and display execution progress and final summary report with zero emojis.
  - Return exit code 0 on success, or 1 if fatal configuration errors occur.
- Task 4.3: Add tests in `cmd/agentic-lens/main_test.go` covering flag parsing, default parameter resolution, and dry-run invocation.

### Phase 5: TypeScript Frontend Contract Synchronization (`web/src/domain/types.ts`)
- Task 5.1: Add `MigrationStatus` type definition in `web/src/domain/types.ts`:
  - Type: `'pending' | 'partial' | 'completed'`.
- Task 5.2: Add `MigrationInfo` interface in `web/src/domain/types.ts`:
  - Properties: `readonly schema_version: number; readonly status: MigrationStatus; readonly steps_completed: readonly string[]; readonly steps_pending: readonly string[]; readonly last_migrated_at: string; readonly last_migrated_by: string; readonly notes?: string;`.
- Task 5.3: Update `AgenticFactLedger` interface in `web/src/domain/types.ts`:
  - Add optional `readonly schema_version?: number;`.
  - Add optional `readonly migration?: MigrationInfo;`.
- Task 5.4: Run TypeScript typecheck (`npm run typecheck`) to ensure zero typing errors across the web application.

### Phase 6: Automated Verification and Quality Gates
- Task 6.1: Run `go vet ./...` across all backend packages.
- Task 6.2: Run `go test -race ./...` covering all domain, adapter, and CLI tests.
- Task 6.3: Run `npm run typecheck` and `npm test` in `web/`.
- Task 6.4: Execute standalone binary compilation: `CGO_ENABLED=0 go build -o bin/agentic-lens ./cmd/agentic-lens`.
- Task 6.5: Verify dry-run execution against actual repository workspaces:
  - Run `./bin/agentic-lens backfill --workspaces workspaces/ --dry-run`.
  - Confirm that output accurately lists discovered ledgers and planned actions without file alterations.
- Task 6.6: Audit code diff for strict adherence to zero-emoji policy and zero `any` declarations.
