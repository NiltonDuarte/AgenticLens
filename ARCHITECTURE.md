# AgenticLens System Architecture and Internals

AgenticLens is engineered to provide deterministic observability and telemetry aggregation for autonomous multi-agent software engineering workflows. The platform operates on the Functional Core, Imperative Shell (FCIS) pattern, distributing as a self-contained, single Go executable that embeds a strict TypeScript single-page application (SPA).

---

## 1. Architectural Philosophy: Functional Core, Imperative Shell (FCIS)

The codebase strictly segregates pure business logic from side-effectful I/O operations:

```
+-------------------------------------------------------------------------------+
|                             Imperative Shell                                  |
|                                                                               |
|   +-----------------------+   +--------------------+   +------------------+   |
|   |  CLI & HTTP Server    |   | OpenCode SQLite DB |   | Git Repository   |   |
|   | (cmd/agentic-lens)    |   | (adapters/opencode)|   | (adapters/git)   |   |
|   +-----------+-----------+   +---------+----------+   +--------+---------+   |
|               |                         |                       |             |
|               +--------------------+    |    +------------------+             |
|                                    v    v    v                                |
|   +-----------------------------------------------------------------------+   |
|   |  Adapters Layer (backend/internal/adapters)                           |   |
|   |  Workspace Traversal, Schema Validation, Backfill Engine, Asset Server|   |
|   +-----------------------------------+-----------------------------------+   |
+---------------------------------------|---------------------------------------+
                                        | (Immutable DTOs)
                                        v
+-------------------------------------------------------------------------------+
|                       Pure Functional Domain Core                             |
|                        (backend/internal/domain)                              |
|                                                                               |
|  * Value Semantics & Immutable Structs     * Pure Metric Aggregations         |
|  * Zero Network / File I/O                 * Context Window Health Models     |
|  * Deterministic Anomaly Rule Engine       * Token Cost Calculations          |
|  * Stepwise Schema Migration Contracts     * Role Mapping & Event Sorts       |
+-------------------------------------------------------------------------------+
```

---

## 2. Subsystem Breakdown

### 2.1 Pure Functional Domain Core (`backend/internal/domain`)

The domain package contains zero I/O operations, network calls, clock queries, or stateful side effects. It operates strictly on value semantics and immutable data structures:

- **Aggregations (`domain/aggregations.go`)**: Pure metric rollups across sessions, turns, tool calls, and model invocations. Recomputes token sums, cache hit percentages, reasoning ratios, and subagent turn velocity.
- **Anomaly Detection (`domain/anomalies.go`)**: Deterministic rule evaluation over chronological tool executions, context token growth series, and subagent delegation transitions.
- **Context Health Scoring (`domain/context_health.go`)**: Computes context window utilization against defined model boundaries (healthy, elevated, critical).
- **Cost Calculations (`domain/pricing.go`)**: Computes USD expenditures based on explicit per-model input, output, reasoning, and cached token pricing tables.
- **Migration Framework (`domain/migration/`)**: Pure migration pipeline contracts defining stepwise ledger upgrades and invariant assertions.

### 2.2 Imperative Boundary Shell (`backend/internal/adapters`)

The adapter package manages all interactions with the operating system, file system, database, and network:

- **OpenCode Extractor (`adapters/opencode/`)**: Opens the configured SQLite database read-only and infers one supported schema family from its complete required table/column set. Evidence-backed v1.19.5 extraction reads `session`/`message`/`part`; v2.2.0 extraction reads `session_v2`/`session_message`. Partial, mixed, malformed, and unknown layouts return diagnostics instead of falling back. `--db-path` remains the explicit override; the default remains `~/.local/share/opencode/opencode.db`. Detection identifies a schema family, not an exact OpenCode release.

The compatibility evidence window is OpenCode v1.19.5 and npm `@opencode/cli` v2.2.0 (resolved as `latest` on 2026-10-29). Run `make opencode-compat-v1`, `make opencode-compat-v2`, or `make opencode-compat`; each runtime target creates separate temporary HOME/XDG roots, project, and absolute `OPENCODE_DB`, and requires the matching runtime plus a local deterministic provider already running. Fixture-based tests do not require installed OpenCode. Nonzero native pricing, imported history, partial migration, duplicate handling, and releases outside these two pinned versions remain unverified.
- **Git Analyzer (`adapters/git/`)**: Inspects git repository history, calculates commit diffs against upstream base branches, and categorizes code changes into production code, test code, or documentation.
- **Workspace Traversal & Validation (`adapters/workspaces.go`, `adapters/schemas.go`)**: Scans workspace directories, validates `agentic_metrics.json` ledgers against JSON schema definitions, and writes atomic updates.
- **Backfill Engine (`adapters/backfill/`)**: Drives batch reconciliation across workspaces, recomputing git code facts and SQLite telemetry.
- **Migration Runner (`adapters/migration/`)**: Implements atomic file persistence, backup creation, and execution of domain migration plans.
- **Web & HTTP Delivery (`adapters/web/`)**: Configures standard library HTTP multiplexers, REST endpoints, Server-Sent Events (SSE) for live streaming, and static asset serving from embedded files.

### 2.3 CLI & Command Routing (`backend/cmd/agentic-lens`)

The command-line interface provides unified execution pathways for local observation and automation:

- `server`: Launches the HTTP server and delivers both the REST API and the embedded dashboard.
- `generate`: Final-step CLI command for active agent sessions. Ingests session telemetry, computes anomalies, and outputs JSON metrics and summary markdown blocks.
- `backfill`: Batch reconciler for historical workspace ledgers.
- `migrate`: Deterministic multi-version schema upgrader for telemetry files.
- `version`: Outputs binary build version, git commit SHA, and compilation timestamp in plain, short, or JSON formats.

### 2.4 Embedded Web Dashboard (`web/`)

The user interface is an SPA built with TypeScript and React:

- **Embedded Distribution**: Production assets (`web/dist/`) are embedded into the Go binary at compile time via `embed.FS`. The server delivers the dashboard with zero runtime Node.js or npm dependencies.
- **Live Stream Architecture**: Uses Server-Sent Events (SSE) to poll local SQLite database state, streaming live agent turn counts, tool invocations, and token saturation metrics without browser refreshes.
- **Interactive Anomaly Drawer**: Surfaces detected anomalies, severity classifications, and reproduction context directly alongside session metrics.

---

## 3. Execution Anomaly Detection Engine

The anomaly detection engine is implemented as a deterministic, pure-functional rule evaluator inside `backend/internal/domain/anomalies.go`. It analyzes tool execution sequences, context growth patterns, and subagent invocation transitions to uncover operational risks during agentic sessions.

### 3.1 Detection Rules

| Rule Identifier | Description | Warning Threshold | Critical Threshold |
|---|---|---|---|
| `TOOL_RETRY_STREAK` | Detects repeated consecutive failures on the same tool name without interleaved successful actions. | 3 consecutive failures | 5 consecutive failures |
| `TOOL_HIGH_FAILURE_RATE` | Detects elevated tool failure rates for tools with at least 5 total invocations within the session. | Failure rate > 20% | Failure rate >= 50% |
| `CONTEXT_SATURATION_SPIKE` | Identifies sudden token context surges or context window exhaustion threatening model degradation. | Single-turn delta > 40,000 tokens | Single-turn delta > 80,000 tokens, or peak tokens >= 160,000 (80% of ceiling) |
| `SUBAGENT_REWORK_LOOP` | Identifies ping-pong delegation bouncing between subagent roles (e.g. coder and tester alternating without progress). | 3 alternating cycles | 5 alternating cycles |

### 3.2 CLI Inspection and Reporting

Inspect anomalies via the CLI across one or more workspaces:

```bash
# List all workspaces with anomaly counts
./bin/agentic-lens --workspaces alspaces --list

# View detailed anomaly diagnosis table for a specific ticket
./bin/agentic-lens --workspaces alspaces --ticket-id AL-15 --anomalies

# Append or update markdown metrics and anomalies table in SUMMARY.md
./bin/agentic-lens --workspaces alspaces --ticket-id AL-15 --append alspaces/AL-15/SUMMARY.md
```

---

## 4. Telemetry Backfill Engine (`agentic-lens backfill`)

The telemetry backfill engine provides batch reconciliation across workspaces, recomputing Git code facts, extracting OpenCode database telemetry, and upgrading telemetry schema versions across historical workspaces.

### 4.1 Supported Flags

- `--workspaces <path>`: Path to root workspaces directory containing `agentic_metrics.json` ledgers. Defaults to configured repository workspaces directory.
- `--ticket-id <id>`: Filter backfill execution to a single ticket identifier.
- `--recompute-git`: Recompute Git code facts (lines added/deleted, test/prod breakdown, files changed, and SDLC iterations) from repository history.
- `--recompute-db`: Recompute OpenCode invocations, turn counts, token metrics, and tool facts from local SQLite DB.
- `--migrate-schema`: Upgrade telemetry ledger schema versions to latest supported version (v2).
- `--dry-run`: Preview planned changes and modifications without writing changes to disk.
- `--update-summaries`: Idempotently update `SUMMARY.md` markdown metrics tables in processed workspace directories.
- `--db-path <path>`: Explicit path to OpenCode SQLite database file.
- `--force`: Force recomputation and migration even if the ledger is marked as completed.

### 4.2 CLI Usage Examples

```bash
# Dry-run preview of Git code facts recomputation across all workspaces
./bin/agentic-lens backfill --workspaces alspaces --recompute-git --dry-run

# Full batch telemetry backfill with schema migration and summary table updates
./bin/agentic-lens backfill --workspaces alspaces --recompute-git --recompute-db --migrate-schema --update-summaries

# Single ticket targeted backfill with Git facts recomputation and schema migration
./bin/agentic-lens backfill --workspaces alspaces --ticket-id AL-15 --recompute-git --migrate-schema --update-summaries
```

---

## 5. Universal Schema Migration Engine (`agentic-lens migrate`)

The universal schema migration engine provides stepwise multi-version schema upgrades for `agentic_metrics.json` ledgers, transitioning historical unversioned (v1) ledgers to schema version 2 (and subsequent schemas) cleanly and deterministically.

### 5.1 Core Architectural Invariants

- **Non-Destructive Preservation**: All existing telemetry metrics, Git facts, invocations, and tool records are strictly preserved during migration.
- **Seven-Point Invariant Audit**: Every migration step applies an invariant audit between pre- and post-migration states to verify ticket identification, timestamps, Git code facts, invocations, tools, event facts, and target version.
- **Idempotency**: Running migrations against ledgers that are already up-to-date performs zero modifications.
- **Per-Workspace Error Isolation**: Unparseable or corrupt JSON in one workspace logs a descriptive warning without aborting batch processing for other workspaces.
- **Atomic File Persistence**: Writes are executed to temporary sibling files before atomic filesystem rename, with optional `.bak` backups.

### 5.2 Supported Flags

- `--workspaces <path>`: Path to root workspaces directory containing `agentic_metrics.json` ledgers.
- `--ticket-id <id>`: Target a single ticket identifier for migration.
- `--target-version <int>`: Target schema version integer (defaults to current version: 2).
- `--dry-run`: Preview planned version transitions without modifying files on disk.
- `--backup`: Create `.bak` backup copy of original `agentic_metrics.json` file prior to overwriting.
- `--update-summaries`: Idempotently update `SUMMARY.md` markdown metrics tables in migrated workspaces.

### 5.3 CLI Usage Examples

```bash
# Dry-run preview of pending schema migrations across all workspaces
./bin/agentic-lens migrate --workspaces alspaces --dry-run

# Batch upgrade all historical workspaces to version 2 with backup file generation
./bin/agentic-lens migrate --workspaces alspaces --backup

# Migrate single ticket workspace with backup creation and summary table regeneration
./bin/agentic-lens migrate --workspaces alspaces --ticket-id AL-15 --backup --update-summaries
```

---

## 6. References

- Technical Invariants and Guidelines: [TECHNICAL_GUIDELINES.md](TECHNICAL_GUIDELINES.md)
- Developer Setup and Contribution Workflow: [CONTRIBUTING.md](CONTRIBUTING.md)
- Platform Setup Manual and Skill Integration: [docs/SETUP_MANUAL.md](docs/SETUP_MANUAL.md)
