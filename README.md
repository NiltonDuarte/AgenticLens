# AgenticLens

AgenticLens is a developer telemetry and observability dashboard for multi-agent software engineering workflows. It ingests, aggregates, and visualizes lifecycle telemetry generated across workspaces (`workspaces/**/agentic_metrics.json`), tracking token saturation, context window health, financial cost, iteration velocity, and tool error distributions.

## Architecture

AgenticLens is implemented in **Go** and **TypeScript** following the Functional Core, Imperative Shell (FCIS) pattern, distributed as a **single, self-contained Go executable**:

- **Domain Layer (`internal/domain`)**: Pure Go functional core. Value semantics, immutable structs, zero I/O, pure aggregations, context health scoring, cost calculations, and stepwise schema migration framework (`internal/domain/migration`).
- **Adapter Layer (`internal/adapters`)**: Imperative boundary shell. Workspace filesystem traversal, JSON schema validation, stepwise migration runners and backfill engines (`internal/adapters/backfill`), HTTP routing, and embedded asset serving (`internal/adapters/web`).
- **CLI & Server (`cmd/agentic-lens`)**: Entrypoint for CLI execution and observation subcommands (`server`, `generate`, `backfill`, `migrate`) as well as root workspace inspection commands.
- **Web Dashboard (`web/`)**: Strict TypeScript dashboard (SPA). Compiled static assets (`web/dist/`) are embedded into the Go executable at build time via `embed.FS`. Zero Node.js runtime required in production.

## Running Locally

### Development Mode

Run the Go server and Vite dev server with hot reload:

```bash
# Terminal 1: Run Go API / telemetry engine
go run ./cmd/agentic-lens server --port 8080

# Terminal 2: Run TypeScript UI with API proxy
cd web && npm run dev
```

### Standalone Production Binary

Build and run the self-contained binary:

```bash
# Build frontend assets and Go binary
cd web && npm run build && cd ..
go build -o bin/agentic-lens ./cmd/agentic-lens

# Run dashboard (serves embedded UI and API)
./bin/agentic-lens server --port 8501
```

Visit `http://localhost:8501`.

## Execution Anomaly Detection

AgenticLens features a deterministic, pure-functional anomaly detection engine running inside the domain core (`internal/domain`). During telemetry ingestion and metric projection, the engine scans tool execution logs, context growth curves, and subagent invocation transitions to identify operational hazards:

### Detection Rules

- **`TOOL_RETRY_STREAK`**: Detects consecutive failures on identical tools.
  - `WARNING`: 3 consecutive errors.
  - `CRITICAL`: 5 consecutive errors.
- **`TOOL_HIGH_FAILURE_RATE`**: Detects elevated failure rates for tools with at least 5 invocations.
  - `WARNING`: Failure rate > 20%.
  - `CRITICAL`: Failure rate >= 50%.
- **`CONTEXT_SATURATION_SPIKE`**: Detects hazardous context window growth.
  - `WARNING`: Single-turn token delta > 40,000 tokens.
  - `CRITICAL`: Single-turn token delta > 80,000 tokens, or peak context tokens >= 160,000 tokens (80% of healthy ceiling).
- **`SUBAGENT_REWORK_LOOP`**: Detects cyclical delegation bouncing between subagent roles (for example, coder and tester alternating).
  - `WARNING`: 3 alternating cycles.
  - `CRITICAL`: 5 alternating cycles.

### CLI Diagnostics

Inspect anomalies directly from the CLI:

```bash
# List discovered workspaces with anomaly summary counts
./bin/agentic-lens --workspaces workspaces --list

# View detailed anomaly diagnosis table for a ticket
./bin/agentic-lens --workspaces workspaces --ticket-id <TICKET_ID> --anomalies

# Append or update markdown metrics and anomalies table in SUMMARY.md
./bin/agentic-lens --workspaces workspaces --ticket-id <TICKET_ID> --append workspaces/<TICKET_ID>/SUMMARY.md
```

### Markdown Report Integration

When generating or updating markdown summaries via `--append`, an `### Anomalies & Health Flags` section is rendered. If no issues are found, a clean session status is displayed; otherwise, detected anomalies are formatted with severity badges (`[CRITICAL]`, `[WARNING]`), rule identifiers, and contextual descriptions.

## Telemetry Backfill Engine (`agentic-lens backfill`)

The telemetry backfill engine provides batch reconciliation across workspaces, recomputing Git code facts, extracting OpenCode database telemetry, and upgrading telemetry schema versions across historical workspaces.

### Supported Flags

- `--workspaces <path>`: Path to root workspaces directory containing `agentic_metrics.json` ledgers. Defaults to configured or repository `workspaces/`.
- `--ticket-id <id>`: Filter backfill execution to a single ticket identifier.
- `--recompute-git`: Recompute Git code facts (lines added/deleted, test/prod breakdown, files changed, and SDLC iterations) from repository history.
- `--recompute-db`: Recompute OpenCode invocations, turn counts, token metrics, and tool facts from local SQLite DB.
- `--migrate-schema`: Upgrade telemetry ledger schema versions to latest supported version (v2).
- `--dry-run`: Preview planned changes and modifications without writing changes to disk.
- `--update-summaries`: Idempotently update `SUMMARY.md` markdown metrics tables in processed workspace directories.
- `--db-path <path>`: Explicit path to OpenCode SQLite database file.
- `--force`: Force recomputation and migration even if the ledger is marked as completed.

### CLI Usage Examples

```bash
# Dry-run preview of Git code facts recomputation across all workspaces
./bin/agentic-lens backfill --workspaces workspaces --recompute-git --dry-run

# Full batch telemetry backfill with schema migration and summary table updates
./bin/agentic-lens backfill --workspaces workspaces --recompute-git --recompute-db --migrate-schema --update-summaries

# Single ticket targeted backfill with Git facts recomputation and schema migration
./bin/agentic-lens backfill --workspaces workspaces --ticket-id AL-15 --recompute-git --migrate-schema --update-summaries
```

## Universal Schema Migration Engine (`agentic-lens migrate`)

The universal schema migration engine provides stepwise multi-version schema upgrades for `agentic_metrics.json` ledgers, transitioning historical unversioned (v1) ledgers to schema version 2 (and subsequent schemas) cleanly and deterministically.

### Core Architectural Invariants

- **Non-Destructive Preservation**: All existing telemetry metrics, Git facts, invocations, and tool records are strictly preserved during migration.
- **Seven-Point Invariant Audit**: Every migration step applies an invariant audit between pre- and post-migration states to verify ticket identification, timestamps, Git code facts, invocations, tools, event facts, and target version.
- **Idempotency**: Running migrations against ledgers that are already up-to-date performs zero modifications.
- **Per-Workspace Error Isolation**: Unparseable or corrupt JSON in one workspace logs a descriptive warning without aborting batch processing for other workspaces.
- **Atomic File Persistence**: Writes are executed to temporary sibling files before atomic filesystem rename, with optional `.bak` backups.

### Supported Flags

- `--workspaces <path>`: Path to root workspaces directory containing `agentic_metrics.json` ledgers.
- `--ticket-id <id>`: Target a single ticket identifier for migration.
- `--target-version <int>`: Target schema version integer (defaults to current version: 2).
- `--dry-run`: Preview planned version transitions without modifying files on disk.
- `--backup`: Create `.bak` backup copy of original `agentic_metrics.json` file prior to overwriting.
- `--update-summaries`: Idempotently update `SUMMARY.md` markdown metrics tables in migrated workspaces.

### CLI Usage Examples

```bash
# Dry-run preview of pending schema migrations across all workspaces
./bin/agentic-lens migrate --workspaces workspaces --dry-run

# Batch upgrade all historical workspaces to version 2 with backup file generation
./bin/agentic-lens migrate --workspaces workspaces --backup

# Migrate single ticket workspace with backup creation and summary table regeneration
./bin/agentic-lens migrate --workspaces workspaces --ticket-id AL-15 --backup --update-summaries
```

## Local Cluster Deployment

To build and deploy AgenticLens container into the local Kubernetes cluster (`sinklayer-local`):

```bash
docker build -t agentic-lens:latest .
```
