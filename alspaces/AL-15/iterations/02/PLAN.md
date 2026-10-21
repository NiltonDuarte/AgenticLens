# Implementation Plan: AL-15 Universal Multi-Version Schema Migration Engine (Iteration 02)

## 1. Iteration Goal and Scope

Iteration 01 established the universal schema migration domain core, the imperative migration runner, the `migrate` CLI subcommand, and foundational test suites. Iteration 02 addresses user feedback, documentation gaps, CLI discoverability shortcomings, and upstream git conflicts:

1. Unblock test execution by cleanly resolving the merge conflict in `backend/internal/adapters/schemas_test.go`, preserving both upstream anomaly DTO mapping tests and AL-15 schema migration boundary tests.
2. Fix root CLI help output in `backend/cmd/agentic-lens/main.go` so that running `agentic-lens --help`, `agentic-lens -h`, or `agentic-lens help` explicitly enumerates all four primary subcommands (`server`, `generate`, `backfill`, `migrate`) alongside root inspection flags.
3. Update `README.md` to thoroughly document the telemetry backfill engine (`agentic-lens backfill`) and the universal schema migration engine (`agentic-lens migrate`), detailing their flags, behavioral semantics, and concrete CLI usage patterns.
4. Add automated CLI integration tests in `backend/cmd/agentic-lens/main_test.go` asserting that root help commands output all subcommands and flags.
5. Re-run complete project quality gates to verify clean compilation, zero race conditions, and zero typing errors.

---

## 2. Grounded Reference Architecture and Concrete Paths

### Adapters Test Layer (`backend/internal/adapters/`)
- `backend/internal/adapters/schemas_test.go`:
  Target of active git merge conflict. Contains upstream functions `TestBrittleToolSummaryDTO_BidirectionalMapping` and `TestFleetAnomalySummaryDTO_BidirectionalMapping` alongside Iteration 01 function `TestSchemaMigrationDTOBoundaries`. Both test suites must be reconciled and preserved in sequential order with all conflict markers removed.

### CLI Application Layer (`backend/cmd/agentic-lens/`)
- `backend/cmd/agentic-lens/main.go`:
  Application entrypoint. Currently contains dispatch routing for `server`, `generate`, `backfill`, and `migrate`, but falls through to standard Go `flag.Parse()` without a custom `flag.Usage` handler. Must be updated to:
  - Intercept `help`, `-h`, and `--help` as explicit first-argument subcommands.
  - Define a structured `printRootUsage()` function assigned to `flag.Usage` before parsing root flags.
  - Explicitly document all four subcommands (`server`, `generate`, `backfill`, `migrate`) with purpose statements and flag pointers (`agentic-lens <subcommand> --help`).
  - Document all root workspace inspection flags (`--workspaces`, `--ticket-id`, `--list`, `--anomalies`, `--append`, `--json-output`).
- `backend/cmd/agentic-lens/main_test.go`:
  CLI integration test suite. Must be extended with tests asserting help output contents across invocation styles (`--help`, `-h`, and `help`).

### Documentation (`README.md`)
- `README.md`:
  Root developer documentation. Currently documents architecture, local server setup, and execution anomaly detection, but completely omits the telemetry backfill engine and schema migration engine. Must be expanded with:
  - Comprehensive `Telemetry Backfill Engine` section covering `agentic-lens backfill`.
  - Comprehensive `Universal Schema Migration Engine` section covering `agentic-lens migrate`.
  - Updated `Architecture` and `CLI & Server` sections referencing `backfill` and `migrate`.

---

## 3. Phased Implementation Tasks

### Phase 1: Git Conflict Reconciliation in Adapters (`backend/internal/adapters/schemas_test.go`)
- Task 1.1: Audit the conflict region in `backend/internal/adapters/schemas_test.go` between lines 239 and 453.
- Task 1.2: Preserve `TestBrittleToolSummaryDTO_BidirectionalMapping` completely:
  - Subtest asserting fully populated brittle tool mapping from domain to DTO and round-trip to domain.
  - Subtest asserting JSON serialization round-trip for `BrittleToolSummaryDTO`.
- Task 1.3: Preserve `TestFleetAnomalySummaryDTO_BidirectionalMapping` completely:
  - Subtest asserting fully populated fleet anomaly summary mapping from domain to DTO and back.
  - Subtest asserting JSON serialization round-trip for `FleetAnomalySummaryDTO`.
- Task 1.4: Preserve `TestSchemaMigrationDTOBoundaries` completely:
  - Subtest asserting schema version 2 initializes a non-nil empty events slice in serialized JSON.
  - Subtest asserting schema version 1 does not inject synthetic event arrays when events are nil.
  - Subtest asserting malformed types (string version, float version, non-array events, invalid code object) trigger deserialization errors.
  - Subtest asserting empty JSON objects deserialize safely to domain without panics.
- Task 1.5: Remove all conflict markers (`<<<<<<< Updated upstream`, `=======`, `>>>>>>> Stashed changes`).
- Task 1.6: Verify standard library imports in `backend/internal/adapters/schemas_test.go` including `strings`, `testing`, `encoding/json`, `agenticlens/internal/adapters`, and `agenticlens/internal/domain`.
- Task 1.7: Run `go test ./backend/internal/adapters/...` to confirm tests compile and pass.

### Phase 2: CLI Subcommand Help Enhancement (`backend/cmd/agentic-lens/main.go`)
- Task 2.1: Implement custom root usage printer `printRootUsage()` in `backend/cmd/agentic-lens/main.go`:
  - Output clear header: AgenticLens Developer Telemetry and Observability CLI.
  - Output usage syntax showing both subcommand execution pattern and root inspection flag pattern.
  - Output structured subcommands section:
    - `server`: Start the HTTP dashboard server and API service.
    - `generate`: Ingest and compute facts from OpenCode SQLite DB and Git history.
    - `backfill`: Batch telemetry backfill, Git code facts recomputation, and schema upgrades.
    - `migrate`: Universal stepwise schema migration engine for telemetry ledgers.
  - Output structured root flags section:
    - `--workspaces`: Path to root workspaces directory containing agentic_metrics.json ledgers.
    - `--ticket-id`: Target ticket identifier to filter and report on.
    - `--list`: List all discovered ticket workspaces and summary metrics.
    - `--anomalies`: Display anomalies and health flags for target ticket.
    - `--append`: Path to markdown summary file to idempotently update with metrics table.
    - `--json-output`: Path to write computed metrics ledger as JSON.
  - Output informational footnote instructing users to run `agentic-lens <subcommand> --help` for subcommand-specific options.
- Task 2.2: Assign `flag.Usage = printRootUsage` in `main()` before `flag.Parse()`.
- Task 2.3: Add explicit argument interceptor at the start of `main()`:
  - If `len(os.Args) > 1` and `os.Args[1]` is `help`, `-h`, or `--help`: invoke `printRootUsage()` and exit with code 0.
- Task 2.4: Ensure zero emojis in usage text and formatting.

### Phase 3: Comprehensive CLI Help Test Suite (`backend/cmd/agentic-lens/main_test.go`)
- Task 3.1: Add `TestCLIHelpCommand` in `backend/cmd/agentic-lens/main_test.go`:
  - Subtest testing `--help`: Assert process exits with code 0 (or exit status acceptable to flag parser), stdout or stderr contains `server`, `generate`, `backfill`, and `migrate`, and lists root inspection flags.
  - Subtest testing `-h`: Assert process outputs all four subcommands.
  - Subtest testing `help`: Assert process exits with code 0 and outputs all four subcommands.
- Task 3.2: Add subtest verifying `backfill --help` lists `--migrate-schema`, `--recompute-git`, `--recompute-db`, `--dry-run`, `--update-summaries`, and `--force`.
- Task 3.3: Add subtest verifying `migrate --help` lists `--workspaces`, `--ticket-id`, `--target-version`, `--dry-run`, `--backup`, and `--update-summaries`.

### Phase 4: Documentation Updates (`README.md`)
- Task 4.1: Update `Architecture` and `CLI & Server` overview in `README.md`:
  - List all four primary subcommands (`server`, `generate`, `backfill`, `migrate`).
- Task 4.2: Add dedicated section `Telemetry Backfill Engine (`agentic-lens backfill`)`:
  - Explain purpose: batch workspace reconciliation, recomputing Git code facts, extracting OpenCode database telemetry, and upgrading schema versions across historical workspaces.
  - Document all supported flags: `--workspaces`, `--ticket-id`, `--recompute-git`, `--recompute-db`, `--migrate-schema`, `--dry-run`, `--update-summaries`, `--db-path`, and `--force`.
  - Provide copy-pasteable CLI command examples:
    - Dry-run preview of Git facts recomputation.
    - Full telemetry backfill with schema migration and markdown summary updates.
    - Single ticket filtered backfill.
- Task 4.3: Add dedicated section `Universal Schema Migration Engine (`agentic-lens migrate`)`:
  - Explain purpose: universal stepwise multi-version schema migration for `agentic_metrics.json` ledgers, upgrading unversioned v1 ledgers to schema version 2.
  - Highlight core invariants: non-destructive field preservation, seven-point invariant audit, per-workspace error isolation, dry-run simulation, and atomic file replacement with optional `.bak` backups.
  - Document all supported flags: `--workspaces`, `--ticket-id`, `--target-version`, `--dry-run`, `--backup`, and `--update-summaries`.
  - Provide copy-pasteable CLI command examples:
    - Dry-run preview of pending schema migrations across all workspaces.
    - Batch migration upgrading workspaces with automatic backup creation.
    - Single ticket targeted migration with summary table regeneration.
- Task 4.4: Ensure zero emojis across all newly added markdown documentation.

### Phase 5: Verification & Quality Gates
- Task 5.1: Run `go vet ./...` in `backend/` to verify zero static analysis warnings.
- Task 5.2: Run `go test -race ./...` in `backend/` to confirm all unit, invariant, adapter, and CLI tests pass cleanly without race conditions.
- Task 5.3: Run `npm run typecheck` in `web/` to confirm zero frontend type regressions.
- Task 5.4: Build CLI executable via `go build -o bin/agentic-lens ./cmd/agentic-lens` and run `./bin/agentic-lens --help` manually to confirm complete help formatting.
- Task 5.5: Audit git diff for strict compliance with the zero-emoji and zero-any mandates.

---

## 4. Invariant and Risk Analysis

1. Test Conflict Risk:
   - Risk: Accidentally dropping either the anomaly tests or the migration boundary tests during conflict resolution.
   - Mitigation: Explicitly retain both test suites as separate named functions and verify both in `go test ./backend/internal/adapters/...`.
2. Flag Parsing Precedence:
   - Risk: Standard Go `flag.Parse` calling `os.Exit(2)` on unknown flags or default `-h`.
   - Mitigation: By overriding `flag.Usage` and directly intercepting `help`, `-h`, and `--help` before calling `flag.Parse()`, the CLI maintains predictable exit codes (0 for explicit help requests) while formatting all subcommands.
3. Documentation Consistency:
   - Risk: Documenting flags in `README.md` that diverge from actual CLI flags registered in `main.go`.
   - Mitigation: Direct verification against `runBackfill` and `runMigrate` flag sets in `backend/cmd/agentic-lens/main.go`.

---

## 5. Acceptance Criteria for Iteration 02

1. `backend/internal/adapters/schemas_test.go` has zero git conflict markers and passes all tests.
2. Running `./bin/agentic-lens --help`, `./bin/agentic-lens -h`, and `./bin/agentic-lens help` clearly displays all four subcommands (`server`, `generate`, `backfill`, `migrate`) along with their descriptions and root inspection flags.
3. `README.md` documents both `agentic-lens backfill` and `agentic-lens migrate` with complete flag listings and concrete usage examples.
4. CLI help integration tests pass in `backend/cmd/agentic-lens/main_test.go`.
5. All backend tests pass with `go test -race ./...`.
6. Frontend type check passes with `npm run typecheck`.
7. Zero emojis and zero `any` declarations exist across modified files.
