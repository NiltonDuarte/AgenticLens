# Contract: AL-15 Universal Multi-Version Schema Migration Engine for Telemetry Ledgers

## 1. Problem Statement and Context
Telemetry ledgers in AgenticLens (`agentic_metrics.json`) evolve across project iterations. Earlier iterations utilized an unversioned v1 schema containing coarse token counts and tool counts without schema version markers, event streams, or migration audit records. Ticket AL-12 introduced schema version 2 and distributed migration metadata, and Ticket AL-6 added chronological event fact tracking. Future tickets will continue to introduce schema additions (such as anomaly streams, tool failure telemetry, and live session trees).

Workspaces created during older iterations or restored from previous checkouts become incompatible or incomplete when read by newer engines. The system currently lacks a universal, stepwise schema migration engine. Upgrades between arbitrary historical versions must be deterministic, stepwise, non-destructive, and idempotent.

Ticket AL-15 establishes a universal multi-version migration engine in Go:
1. It introduces a pure functional domain migration framework that chains discrete version migration steps sequentially from source version to target version.
2. It implements the baseline migration step from unversioned v1 to version 2, guaranteeing non-destructive field preservation, schema version initialization, and empty event slice setup.
3. It enforces strict domain invariant validation at every migration transition, ensuring no custom metrics, Git code facts, invocation records, tool records, or timestamps are lost or corrupted.
4. It implements an imperative boundary migration runner in the adapters layer capable of scanning workspaces, detecting schema versions, executing stepwise upgrades, generating backup files, and atomically persisting changes.
5. It exposes a dedicated CLI subcommand (`agentic-lens migrate`) with dry-run support, ticket filtering, target version selection, and backup capabilities.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

The implementation strictly satisfies the architectural boundaries defined in `TECHNICAL_GUIDELINES.md`:

### Inward Dependency Flow
Dependencies flow inward across layers:
1. Application Entrypoint & CLI (`cmd/agentic-lens/main.go`):
   Parses command-line arguments for the new `migrate` subcommand and the `--migrate-schema` flag on `backfill`. Formats terminal reports, progress metrics, and warning tables. Zero business logic in CLI entrypoints.
2. Imperative Adapters Layer (`internal/adapters/`):
   - `internal/adapters/backfill/`: Discovers workspace directories, loads file contents, generates `.bak` backups when requested, executes atomic file writes using temporary files and filesystem renames, and triggers Markdown summary regenerations.
   - `internal/adapters/schemas.go`: Provides bidirectional DTO mapping, JSON encoding/decoding, and boundary validation between serialized JSON files and domain models.
   - `internal/adapters/repository.go`: Provides filesystem traversal and workspaces directory path resolution.
3. Pure Functional Core Domain Layer (`internal/domain/` and `internal/domain/migration/`):
   - `internal/domain/models.go` and `internal/domain/constants.go`: Immutable domain entities, schema version constants, and value objects.
   - `internal/domain/migration/`: Stepwise migration interface, migration step registry, migration chain planner, invariant validator, and pure step implementations.
   - Zero External Dependencies: Pure functions operating exclusively on immutable structs and values. Zero imports of disk I/O, network libraries (`net/http`, `os`, `database/sql`), or outer layers (`cmd/`, `internal/adapters/`). System clocks and file paths are injected from the imperative shell.

### Pure Domain Core Isolation
1. Migration Step Abstraction:
   Each migration step defines its source schema version, target schema version, human-readable description, and a pure transformation function that maps an input ledger to an upgraded ledger.
2. Step Registry & Sequential Chain:
   The registry maintains registered steps, enforces continuous version transitions (for example, version 1 to 2, version 2 to 3), rejects ambiguous or duplicate steps, and plans sequential execution paths.
3. Deterministic Invariant Validation:
   Every migration transition executes an invariant audit. The validator inspects the ledger before and after the step, verifying that core identity fields, Git code facts, invocations, tools, and timestamps remain intact.
4. Idempotency:
   Executing migration against a ledger that is already at the target version performs zero modifications and returns the ledger unchanged.

### Imperative Boundary Shell Rules
1. Non-Destructive Preservation Mandate:
   Migration must never drop or overwrite existing telemetry facts. If unexpected fields exist, they must be preserved across migration cycles.
2. Per-Ticket Error Isolation:
   A corrupted or unparseable JSON file in one workspace directory must generate a warning in the execution report and must not abort batch migration for remaining workspaces.
3. Atomic File Persistence:
   All disk mutations must write to a sibling temporary file (`agentic_metrics.json.tmp`) followed by an atomic filesystem rename to ensure that system interruptions or crashes cannot leave partially written ledgers.
4. Optional Ledger Backups:
   When backup mode is enabled, the adapter must create a `.bak` copy of the original ledger before writing the upgraded file.
5. Strict Zero-Emoji Mandate:
   Zero emojis across CLI output, logs, comments, documentation, and serialized files.
6. Static Typing and Zero Any:
   Zero `any` or empty interface declarations in Go and TypeScript.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Go Domain Migration Core (`internal/domain/migration/`):
   - Define `MigrationStep` interface: `SourceVersion() int`, `TargetVersion() int`, `Description() string`, and `Apply(domain.AgenticFactLedger) (domain.AgenticFactLedger, error)`.
   - Implement `Registry`: thread-safe step registration, version validation, and migration chain discovery (`PlanChain(sourceVersion, targetVersion)`).
   - Implement `InvariantValidator`: pure function comparing pre-migration and post-migration ledgers for metric integrity.
   - Implement `StepV1ToV2`: pure migration step upgrading unversioned v1 ledgers to schema version 2 by assigning schema version 2, initializing non-nil empty event slices, and establishing baseline migration metadata.
   - Implement `Engine`: high-level coordinator detecting source schema version from raw JSON payload, resolving migration steps, executing transformations, validating invariants, and returning the upgraded ledger.
2. Boundary Adapters Layer (`internal/adapters/backfill/`):
   - Extend `BackfillOptions` with migration flags.
   - Create `MigrationRunner` in `internal/adapters/backfill/` that coordinates directory scanning, version detection, domain engine invocation, dry-run simulation, optional `.bak` file generation, and atomic persistence.
   - Expose migration reporting structures capturing scanned count, upgraded count, skipped count, and warning details.
3. CLI Layer (`cmd/agentic-lens/main.go`):
   - Add dedicated `migrate` subcommand parsing flags: `--workspaces`, `--ticket-id`, `--target-version`, `--dry-run`, `--backup`, and `--update-summaries`.
   - Add `--migrate-schema` flag to existing `backfill` subcommand to trigger universal migration during backfills.
   - Render clean, zero-emoji progress and summary tables.
4. TypeScript Contract Synchronization (`web/src/domain/types.ts`):
   - Verify that domain types remain strictly aligned with schema version 2 and ready for multi-version ledgers.
5. Automated Test Suites:
   - Pure domain table-driven tests for registry, step planning, invariant validation, and step execution.
   - Invariant tests asserting that running migrations on up-to-date ledgers produces zero mutations.
   - Adapter tests for multi-workspace scanning, dry-run guarantees, backup creation, and corrupted JSON resilience.
   - Race condition verification using Go race detector.

### Out of Scope
1. Database schema migrations for external relational stores.
2. Direct network transmission or remote RPC synchronization of ledgers.
3. Modifying historical Git commit objects or altering Git history.
4. UI component rewrites (dashboard presentation consumes already-migrated v2 ledgers).

---

## 4. Migration Architecture & Stepwise Execution Lifecycle

### Version Detection & Chain Discovery
1. The migration runner reads `agentic_metrics.json` as raw bytes.
2. `Engine.DetectVersion(input)` inspects the `schema_version` field:
   - If `schema_version` is missing, null, or 0, the source version is classified as 1 (unversioned historical ledger).
   - If `schema_version` is present and positive, the source version is that integer value.
3. The runner checks whether `sourceVersion == targetVersion`:
   - If equal, the ledger is up-to-date; execution is skipped or marked idempotent no-op.
   - If `sourceVersion > targetVersion`, an error is returned because downgrades are forbidden.
4. `Registry.PlanChain(sourceVersion, targetVersion)` resolves the required path:
   - For an unversioned v1 ledger targeting version 2, the plan contains `StepV1ToV2`.
   - For future versions (e.g. v1 to v3), the plan contains `StepV1ToV2` followed by `StepV2ToV3`.
   - If a version transition step is missing, planning fails immediately with a descriptive error.

### Step Execution and Invariant Validation
1. The raw payload is deserialized into the initial `domain.AgenticFactLedger`.
2. For each step in the resolved migration chain:
   - The engine captures a copy of the pre-step ledger.
   - The engine invokes `step.Apply(currentLedger)`.
   - The engine passes both the pre-step and post-step ledgers to `InvariantValidator.Validate(before, after)`.
   - If invariant validation fails, migration halts immediately without writing changes.
   - The transformed ledger becomes the input for the next step.
3. The engine returns the fully migrated ledger.

### Persistence Lifecycle
1. Dry-run Mode:
   - The runner records the planned transition (for example, version 1 to 2) and diff summary in the migration report.
   - Zero files on disk are touched.
2. Mutation Mode:
   - If `--backup` is specified, the runner copies `agentic_metrics.json` to `agentic_metrics.json.bak`.
   - The runner serializes the migrated ledger to JSON via `AgenticFactLedgerDTOFromDomain` with 2-space indentation and trailing newline.
   - The runner writes to `agentic_metrics.json.tmp` and renames to `agentic_metrics.json`.
   - If `--update-summaries` is active, it regenerates the summary table in `SUMMARY.md`.

---

## 5. Invariant Validation & Non-Destructive Preservation Rules

The `InvariantValidator` enforces the following invariant assertions between pre-migration and post-migration states:
1. Ticket Identification Invariant:
   `TicketID`, `TicketTitle`, and `TicketType` must match between before and after.
2. Timestamps Invariant:
   `StartedAt` and `CompletedAt` must remain identical. `UpdatedAt` must not be cleared.
3. Git Code Facts Invariant:
   `LinesAdded`, `LinesDeleted`, `TestLinesAdded`, `TestLinesDeleted`, `ProdLinesAdded`, `ProdLinesDeleted`, `FilesChanged`, and `SDLCIterations` must match identically.
4. Invocations Invariant:
   The count of invocations must not change. For every invocation, `ID`, `Role`, `Model`, `User`, `ActiveDurationSeconds`, `Turns`, `Tokens`, `CostUSD`, and `Compactions` must match.
5. Tools Invariant:
   The count of tool records must not change. For every tool fact, `Name`, `Calls`, and `Errors` must match.
6. Event Facts Invariant:
   All pre-existing `EventFact` items must remain identical in order and content. A migration step may initialize an empty non-nil event slice if nil, but must never drop pre-existing events.
7. Schema Version Invariant:
   The post-migration schema version must match the step's `TargetVersion()`.

---

## 6. Acceptance Criteria

1. Executing `agentic-lens migrate --workspaces workspaces/` upgrades all historical `agentic_metrics.json` files to `CurrentSchemaVersion`.
2. Historical unversioned files and explicit v1 files migrate cleanly to v2 without data loss, field corruption, or loss of custom metrics.
3. Re-running migration on an already upgraded ledger produces zero modifications (idempotency).
4. Running `agentic-lens migrate --dry-run` reports planned version upgrades and field additions without writing to disk.
5. Running `agentic-lens migrate --backup` generates `.bak` copies before overwriting ledgers.
6. Corrupted or unparseable JSON files log descriptive error warnings and do not prevent other workspaces from migrating.
7. Running `agentic-lens backfill --migrate-schema` executes schema migration alongside telemetry backfill.
8. Quality checks pass cleanly: `go vet ./...`, `go test -race ./...`, `npm run typecheck`, and `go build ./cmd/agentic-lens`.
9. The implementation contains zero emojis and zero `any` or empty interface declarations.
