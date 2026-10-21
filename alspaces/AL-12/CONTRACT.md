# Contract: AL-12 Go Telemetry Backfill Engine and Distributed Multi-Machine Schema Migration

## 1. Problem Statement
Telemetry schemas in AgenticLens evolve over time. Historical workspaces contain fact ledgers (stored in `agentic_metrics.json`) that lack recent schema attributes, possess zeroed Git code metrics due to legacy extraction passes, or require enrichment against local OpenCode session traces.

A primary operational constraint in this environment is distributed multi-machine development:
The OpenCode SQLite database (`~/.local/share/opencode/opencode.db`) is not centralized. It resides on the host machines of individual developers. A developer executing on Machine A has access to SQLite session traces created on Machine A, but does not possess SQLite records for sessions initiated on Machine B or Machine C. Conversely, the Git repository is a shared source of truth across all machines once commits and branches are fetched.

Ticket AL-12 delivers a native Go telemetry backfill engine and schema migration command (`agentic-lens backfill`) that solves this distributed operational challenge:
1. It scans workspace directories and detects historical ledgers requiring schema migration or metric backfill.
2. It recomputes Git code facts across all reachable ticket branches and commits using the repository history.
3. It performs safe, non-destructive OpenCode SQLite session re-querying when local database records exist.
4. When a session trace is absent from the local machine's database, it preserves existing invocation and tool telemetry without data loss, marking the file with a structured `partial` migration state and recording pending migration steps.
5. It enables partially migrated ledgers to be committed to Git, merged into `main`, and subsequently resumed on another developer's machine where the missing OpenCode database records reside.
6. It provides dry-run previews, resilient per-ticket error isolation, atomic file writes, and automatic Markdown summary table regeneration.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation must strictly comply with the architectural boundaries in `TECHNICAL_GUIDELINES.md`:

### Inward Dependency Flow
Dependencies flow strictly inward across layers:
1. CLI Subcommand Layer (`cmd/agentic-lens/main.go`) parses CLI flags (`--workspaces`, `--ticket-id`, `--recompute-git`, `--recompute-db`, `--dry-run`, `--update-summaries`, `--db-path`, `--force`), formats human-readable progress and summary reports, and wires adapters to domain logic.
2. Imperative Adapters Layer (`internal/adapters/backfill/`, `internal/adapters/opencode/`, `internal/adapters/git/`, `internal/adapters/`):
   - `internal/adapters/backfill/` coordinates workspace discovery, loads existing ledgers, invokes Git and OpenCode adapters, isolates per-ticket failures, performs atomic file writes, and updates summaries.
   - `internal/adapters/opencode/` manages read-only SQLite queries to locate root sessions and extract invocation facts.
   - `internal/adapters/git/` analyzes commit numstats, classifies test versus production code, and counts SDLC iterations.
   - `internal/adapters/schemas.go` handles JSON boundary serialization, validation, and DTO conversion.
   - `internal/adapters/formatters.go` renders Markdown metrics tables and idempotently updates `SUMMARY.md`.
3. Functional Core Domain Layer (`internal/domain/`):
   - Houses pure domain models (`MigrationStatus`, `MigrationInfo`, `AgenticFactLedger`, `AgenticMetricsLedger`), typed constants, and pure migration state evaluation functions.
   - Pure domain functions receive immutable data structures and return modified copies or explicit errors.
   - Zero disk I/O, zero network calls, zero SQL queries, and zero system clocks inside `internal/domain/`. Clocks (such as timestamps) and file paths must be injected from the outer imperative shell.

### Pure Domain Core Isolation
1. `MigrationStatus` is defined as a typed string enum with explicit states: `pending`, `partial`, and `completed`.
2. `MigrationInfo` is an immutable value object capturing schema version, migration status, completed steps, pending steps, migration timestamp, author, and contextual notes.
3. Pure functions evaluate migration state transitions deterministically based on attempted operations, available facts, and previous migration history.
4. Mathematical rollups in `internal/domain/aggregations.go` remain purely functional and guard against division by zero.

### Imperative Boundary Shell Rules
1. Non-Destructive Preservation Mandate: If `--recompute-db` is requested but the target session cannot be found in the local SQLite database, the engine must never overwrite, blank out, or delete existing invocations or tool telemetry already present in `agentic_metrics.json`.
2. Per-Ticket Error Isolation: An unparseable or corrupt ledger file in one workspace must be logged as a non-fatal warning and must never abort the batch processing of other workspaces.
3. Atomic File Writes: Modifying `agentic_metrics.json` must use atomic write patterns (such as writing to a sibling temporary file followed by an atomic rename) to avoid corrupting files during interruptions.
4. Read-Only SQLite Access: OpenCode SQLite connections must open exclusively with read-only URI parameters (`mode=ro`) to prevent locks against concurrent agent processes.
5. Strict Zero-Emoji Mandate: Zero emojis across log statements, CLI outputs, generated JSON files, summary tables, and code comments.
6. Static Compilation: All SQLite interactions rely on the pure Go driver `modernc.org/sqlite` with `CGO_ENABLED=0`.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Domain Migration Engine (`internal/domain/`):
   - Add `MigrationStatus` enum (`pending`, `partial`, `completed`) to `internal/domain/enums.go`.
   - Add `MigrationInfo` struct to `internal/domain/models.go`.
   - Extend `AgenticFactLedger` in `internal/domain/models.go` with optional `SchemaVersion` and `Migration` metadata fields.
   - Add schema constants (`CurrentSchemaVersion = 2`, step identifiers `git_code_facts` and `opencode_db`) to `internal/domain/constants.go`.
   - Implement pure domain function `EvaluateMigrationStatus` in `internal/domain/migration.go` to compute state transitions deterministically.
2. Boundary DTOs and Serialization (`internal/adapters/schemas.go`):
   - Update `AgenticFactLedgerDTO` and create `MigrationInfoDTO` to support bidirectional JSON mapping of migration metadata while preserving full backward compatibility with unversioned historical files.
3. Backfill and Migration Adapter (`internal/adapters/backfill/`):
   - Implement workspace directory scanner discovering all `agentic_metrics.json` ledgers.
   - Implement Git recomputation engine invoking `internal/adapters/git/` to refresh code lines, test/production split, and iterations.
   - Implement OpenCode database re-querying invoking `internal/adapters/opencode/` to enrich invocations, tokens, durations, and tools when sessions exist locally.
   - Implement migration coordinator evaluating partial versus completed status and preserving existing facts when sessions are absent.
   - Implement atomic JSON file writer and Markdown summary updater.
   - Support dry-run execution reporting planned changes without disk mutation.
4. CLI Subcommand (`cmd/agentic-lens/main.go`):
   - Add `backfill` subcommand supporting flags: `--workspaces`, `--ticket-id`, `--recompute-git`, `--recompute-db`, `--dry-run`, `--update-summaries`, `--db-path`, and `--force`.
   - Output structured, informative console summaries detailing scanned ledgers, updated files, partial migrations, completed migrations, skipped files, and warnings.
5. TypeScript Frontend Contract Synchronization (`web/src/domain/types.ts`):
   - Update TypeScript interfaces with `MigrationStatus`, `MigrationInfo`, and optional `schema_version` and `migration` on `AgenticFactLedger`.
6. Automated Test Suites:
   - Pure domain unit tests for migration state transitions and validations.
   - Adapter tests for multi-workspace scanning, partial migration preservation, resumption across mock databases, corrupted file resilience, and dry-run guarantees.
   - CLI flag parsing and execution tests.

### Out of Scope
1. Network synchronization or direct machine-to-machine database replication (state sharing occurs exclusively via Git repository commits and pull requests).
2. Deleting historical workspaces or modifying historical Git commit objects.
3. Chronological event extraction or v2 timeline stream generation (governed by AL-6).
4. New dashboard chart views or UI metric visualizations (governed by frontend tickets).

---

## 4. Multi-Machine Distributed Migration Lifecycle

The distributed migration operates as a cooperative multi-step workflow across developer hosts:

### Step 1: Initial Pass on Author Machine (Machine A)
1. Developer A runs `agentic-lens backfill --workspaces workspaces/ --recompute-git --recompute-db --update-summaries`.
2. The backfill engine scans all workspaces:
   - For tickets whose sessions were run on Machine A:
     - Git metrics are recomputed from the Git repository.
     - OpenCode SQLite finds the root session and child sessions in Machine A's database.
     - Invocation facts and tool facts are refreshed.
     - `schema_version` is set to 2.
     - Migration status is marked as `completed`.
     - `steps_completed` records `git_code_facts` and `opencode_db`.
     - `steps_pending` is empty.
   - For tickets whose sessions were run on Machine B or Machine C:
     - Git metrics are recomputed from the shared Git repository.
     - OpenCode SQLite query returns no matching session in Machine A's local database.
     - The engine preserves all existing `invocations`, `tools`, and duration facts from disk.
     - `schema_version` is set to 2.
     - Migration status is marked as `partial`.
     - `steps_completed` records `git_code_facts`.
     - `steps_pending` records `opencode_db`.
     - `notes` records that local database session resolution was not found on Machine A and telemetry was preserved for subsequent host migration.
3. Updated ledgers and Markdown summaries are written to disk.
4. Developer A reviews, commits the partially migrated files to a Git branch, opens a pull request, and merges into `main`.

### Step 2: Resumption on Second Developer Machine (Machine B)
1. Developer B pulls the latest `main` branch.
2. Developer B runs `agentic-lens backfill --workspaces workspaces/ --recompute-db --update-summaries`.
3. The backfill engine scans workspaces:
   - Fully completed tickets are skipped unless `--force` is specified.
   - Tickets marked with status `partial` and having `opencode_db` in `steps_pending` are evaluated.
   - The engine queries Developer B's local SQLite database using the ticket ID or stored root thread IDs.
   - For tickets authored on Machine B, the session is found. Fresh invocations, token breakdowns, and tool statistics are extracted.
   - Migration status advances from `partial` to `completed`.
   - `opencode_db` is moved from `steps_pending` to `steps_completed`.
   - `last_migrated_by` and `last_migrated_at` are updated to Developer B's identity and current timestamp.
   - If a ticket was authored on Machine C (neither Machine A nor Machine B), it remains `partial` with existing telemetry intact until run on Machine C.
4. Developer B commits the completed ledgers to Git and merges into `main`.

---

## 5. Acceptance Criteria

1. Running `agentic-lens backfill --workspaces workspaces/ --recompute-git` successfully refreshes Git code facts (lines added, lines deleted, test versus production breakdown, file counts, and SDLC iterations) across all workspaces without modifying invocations or tool telemetry.
2. Running `agentic-lens backfill --workspaces workspaces/ --recompute-db` on a machine lacking session records for specific tickets preserves all pre-existing invocations, marks those ledgers with status `partial`, sets `steps_pending` to include `opencode_db`, and advances `schema_version` to 2.
3. Re-running `agentic-lens backfill --workspaces workspaces/ --recompute-db` on a machine possessing the missing session records successfully locates the session, enriches invocation and tool facts, and transitions migration status from `partial` to `completed`.
4. Running `agentic-lens backfill --ticket-id <ID> --dry-run` outputs detailed planned updates (schema transition, Git metric deltas, database session resolution status) without modifying any files on disk.
5. Corrupted, malformed, or unreadable JSON files in the workspaces directory produce explicit error warnings in the log output without halting or aborting the backfill of remaining workspaces.
6. When `--update-summaries` is supplied, corresponding `SUMMARY.md` files are idempotently updated with the newly derived metrics tables.
7. All migrated and partially migrated `agentic_metrics.json` ledgers pass JSON validation, conform to the updated domain schema, and can be read by `agentic-lens server` and the web dashboard without errors.
8. Quality checks pass cleanly: `go vet ./...`, `go test -race ./...`, `npm run typecheck`, and single-binary build `go build ./cmd/agentic-lens`.
9. The codebase, CLI outputs, generated ledgers, and workspace markdown files contain zero emojis and zero `any` or empty interface declarations.
