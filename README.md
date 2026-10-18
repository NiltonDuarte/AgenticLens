# AgenticLens

AgenticLens is a developer telemetry and observability dashboard for multi-agent software engineering workflows. It ingests, aggregates, and visualizes lifecycle telemetry generated across workspaces (`workspaces/**/agentic_metrics.json`), tracking token saturation, context window health, financial cost, iteration velocity, and tool error distributions.

## Architecture

AgenticLens is implemented in **Go** and **TypeScript** following the Functional Core, Imperative Shell (FCIS) pattern, distributed as a **single, self-contained Go executable**:

- **Domain Layer (`backend/internal/domain`)**: Pure Go functional core. Value semantics, immutable structs, zero I/O, pure aggregations, context health scoring, cost calculations, and stepwise schema migration framework (`backend/internal/domain/migration`).
- **Adapter Layer (`backend/internal/adapters`)**: Imperative boundary shell. Workspace filesystem traversal, JSON schema validation, stepwise migration runners and backfill engines (`backend/internal/adapters/backfill`), HTTP routing, and embedded asset serving (`backend/internal/adapters/web`).
- **CLI & Server (`backend/cmd/agentic-lens`)**: Entrypoint for CLI execution and observation subcommands (`server`, `generate`, `backfill`, `migrate`) as well as root workspace inspection commands.
- **Web Dashboard (`web/`)**: Strict TypeScript dashboard (SPA). Compiled static assets (`web/dist/`) are embedded into the Go executable at build time via `embed.FS`. Zero Node.js runtime required in production.

## Running Locally

### Development Mode

Run the Go server and Vite dev server with hot reload:

```bash
# Terminal 1: Run Go API / telemetry engine
cd backend && go run ./cmd/agentic-lens server --port 8085

# Terminal 2: Run TypeScript UI with API proxy
cd web && npm run dev
```

### Standalone Production Binary

Build and run the self-contained binary:

```bash
# Build frontend assets and Go binary
cd web && npm run build && cd ..
cd backend && go build -o bin/agentic-lens ./cmd/agentic-lens

# Run dashboard (serves embedded UI and API)
./bin/agentic-lens server --port 8501
```

Visit `http://localhost:8501`.

## Execution Anomaly Detection

AgenticLens features a deterministic, pure-functional anomaly detection engine running inside the domain core (`backend/internal/domain`). During telemetry ingestion and metric projection, the engine scans tool execution logs, context growth curves, and subagent invocation transitions to identify operational hazards:

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

## Releases and Distribution

AgenticLens artifacts are built as self-contained standalone executables embedding all web dashboard assets, requiring zero external Node.js runtime.

### Release Tiers

- **Release Candidates (Pre-release)**:
  Triggered automatically on every merge to the `main` branch. Artifacts are published as pre-releases on GitHub with version pattern `vMAJOR.MINOR.PATCH-rc.RUN_NUMBER+SHORT_SHA` (for example, `v0.1.0-rc.42+a1b2c3d`).
- **Stable Releases**:
  Triggered when a semantic version tag matching `vMAJOR.MINOR.PATCH` (for example, `v0.1.0`) is pushed to the repository. Non-conforming tags fail the release pipeline validation gate.
- **Floating Latest Tag**:
  The git tag `latest` and the GitHub Release marked as `latest` are automatically synchronized to point to the most recent stable release.

### Target Platform Matrix

| Platform | OS / Architecture | Archive Format | Artifact Name |
|---|---|---|---|
| Linux x86_64 | `linux/amd64` | `.tar.gz` | `agentic-lens-<version>-linux-amd64.tar.gz` |
| Linux ARM64 | `linux/arm64` | `.tar.gz` | `agentic-lens-<version>-linux-arm64.tar.gz` |
| macOS Intel | `darwin/amd64` | `.tar.gz` | `agentic-lens-<version>-darwin-amd64.tar.gz` |
| macOS Apple Silicon | `darwin/arm64` | `.tar.gz` | `agentic-lens-<version>-darwin-arm64.tar.gz` |
| Windows x86_64 | `windows/amd64` | `.zip` | `agentic-lens-<version>-windows-amd64.zip` |

### Verification and Installation

Download the archive corresponding to your platform and the accompanying `SHA256SUMS` file from the GitHub Releases page:

```bash
# Verify archive integrity
sha256sum -c SHA256SUMS --ignore-missing

# Extract tarball (Linux / macOS)
tar -xzf agentic-lens-v0.1.0-linux-amd64.tar.gz

# Inspect version information
./agentic-lens-linux-amd64 version
./agentic-lens-linux-amd64 version --short
./agentic-lens-linux-amd64 version --json
```

### Build and Packaging Targets

The build system provides Makefile targets for local cross-compilation, packaging, and validation:

```bash
# Cross-compile binaries for all supported platforms
make release-build

# Package platform archives and generate SHA256SUMS in dist/
make release-package

# Verify checksums and validate execution of the host platform binary
make release-verify

# Clean distribution artifacts
make release-clean
```
