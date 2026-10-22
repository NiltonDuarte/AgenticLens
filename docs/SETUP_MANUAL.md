# AgenticLens Platform Setup Manual and Telemetry Integration Guide

## 1. Title and Document Overview

AgenticLens is a developer telemetry and observability platform designed specifically for agentic software engineering workflows. It provides real-time monitoring, longitudinal cost and token tracking, automated anomaly detection, and context window health diagnostics across autonomous coding agents.

This manual serves as the comprehensive, authoritative guide for engineering teams, platform architects, and developers integrating AgenticLens into any software repository. It documents the platform architecture, installation pathways, configuration schema, telemetry ingestion mechanics, command-line interface, embedded web dashboard, and operational troubleshooting techniques.

---

## 2. Platform Architecture and Core Concepts

AgenticLens is constructed around the Functional Core, Imperative Shell (FCIS) pattern and packaged as a single, self-contained binary executable with zero external runtime dependencies.

### 2.1 Zero-Dependency Single-Binary Architecture

Traditional developer tools often require complex runtime dependencies such as Node.js, Python runtimes, or external database servers. AgenticLens eliminates this operational overhead:
- The backend is written in Go 1.23+ and compiles to a native static executable with CGO disabled (`CGO_ENABLED=0`).
- SQLite database connectivity is powered by a pure Go driver (`modernc.org/sqlite`), eliminating the requirement for a host C compiler or shared C libraries.
- The web dashboard is built with TypeScript and Vite. During the build process, compiled static frontend assets (`web/dist/`) are embedded directly into the Go executable via `embed.FS`.
- Adopting repositories run a single executable file that serves both the command-line interface and the full-featured interactive web dashboard.

### 2.2 Local Telemetry Fact Ledgers (`agentic_metrics.json`)

Rather than transmitting proprietary code or execution telemetry to external cloud services, AgenticLens operates entirely locally:
- Telemetry facts are captured per ticket or task in isolated workspace directories: `workspaces/<TICKET_ID>/agentic_metrics.json`.
- Each ledger is an immutable JSON document recording session identifiers, wall-clock duration, token volume (input, output, cache-read, and cache-write tokens), financial costs calculated from precise model pricing matrices, turn counts, tool invocation distributions, error rates, and git commit diff volumes (insertions, deletions, and affected files).
- These ledgers are committed directly to version control alongside code deliverables, establishing an auditable, repository-native history of agent engineering metrics.

### 2.3 Real-Time Live Observability and SSE Event Streaming

In addition to post-turn historical ledgers, AgenticLens provides live observability of active agent sessions:
- The embedded server continuously polls the local OpenCode SQLite database at configurable intervals (`poll_interval_ms`, default 2000 ms).
- Active sessions are identified based on recent turn activity within a sliding time window (`active_threshold_seconds`, default 300 seconds).
- Telemetry events, tool calls, and status updates are broadcast to connected browsers via Server-Sent Events (SSE) through the `/api/live/stream` endpoint.
- An in-memory trace buffer retains recent execution events, providing immediate visibility into subagent handoffs, prompt dispatches, and tool outputs as they occur.

### 2.4 Automated Anomaly Detection

AgenticLens continuously analyzes active sessions and historical ledgers against configurable statistical thresholds to detect engineering inefficiencies:
- **Session Stalls**: Sessions that remain open with no new turn activity beyond a configurable duration (`stall_threshold_seconds`, default 120 seconds).
- **High Tool Error Rates**: Sessions where the ratio of failed tool executions to total invocations exceeds a critical limit (`error_rate_threshold`, default 0.25 or 25%).
- **Context Window Exhaustion**: Sessions where cumulative prompt and generation tokens reach or exceed a critical percentage of model capacity (`context_exhaustion_pct`, default 0.85 or 85%).

### 2.5 Context Window Health States

Context saturation severely impacts model reasoning quality, instruction adherence, and cost. AgenticLens classifies context health into four discrete states based on utilization percentage:
- **HEALTHY** (0% to 50% utilization): Optimal operating envelope. The model retains complete instruction adherence and context fidelity with minimal compaction pressure.
- **ELEVATED** (50% to 75% utilization): Normal working range for complex implementations. The agent operates effectively, but monitoring is recommended.
- **BLOATED** (75% to 85% utilization): High context saturation. Model reasoning degradation and prompt compaction risks increase significantly.
- **CRITICAL** (85% to 100% utilization): Severe exhaustion. Immediate session compaction or subagent delegation is advised to avoid catastrophic forgetting or dropped instructions.

---

## 3. Architectural Delineation: Product Platform vs. Internal Methodology

A foundational design principle of AgenticLens is the strict architectural separation between the AgenticLens product platform and the internal development methodology used by its maintainers.

### 3.1 What Constitutes the AgenticLens Product Platform

The AgenticLens product platform consists strictly of:
1. The standalone command-line executable (`agentic-lens`).
2. The embedded TypeScript web dashboard and SSE streaming engine.
3. The root configuration schema (`.agenticlens-config.yaml`).
4. The local telemetry ledger format (`agentic_metrics.json`).
5. The standalone onboarding skill (`agentic-lens-setup`).

### 3.2 What Constitutes This Repository's Internal Development Methodology

This repository uses a specific multi-agent Software Development Life Cycle (SDLC) to build AgenticLens itself:
- Internal development skills located at `.opencode/skills/agentic-lens-dev`, `.opencode/skills/agentic-lens-po`, and `.opencode/skills/agentic-lens-review`.
- The multi-agent persona lifecycle: `orchestrator` -> `planner` -> `coder` -> `tester` -> `reviewer`.
- Specific ticket prefixes such as `AL` and `PO`.
- Rigid multi-turn iteration subfolders (`iterations/<NN>/`).

### 3.3 Agnostic Adoption for External Projects

External repositories adopting AgenticLens must never adopt or duplicate this repository's internal skills, agent personas, or development procedures:
- Adopting teams define their own ways of working, agent prompts, and delivery workflows.
- AgenticLens works seamlessly with single-agent loops, custom multi-agent frameworks, LangChain or AutoGen scripts, human developers using OpenCode or Claude Code, and CI/CD pipelines.
- External projects configure their own ticket naming conventions (e.g. `JIRA-`, `ISSUE-`, `FEAT-`, `BUG-`), workspace directory names, and git branching strategies.
- AgenticLens functions purely as an agnostic telemetry ingestion, aggregation, and visualization layer.

---

## 4. Installation and Setup Pathways

AgenticLens offers two straightforward onboarding pathways: an automated, agent-driven setup skill, and a manual developer installation using precompiled release archives. Target repositories never need to install Go, Node.js, npm, or C build tools.

### 4.1 Pathway A: Agent-Driven Onboarding via `agentic-lens-setup`

For teams using OpenCode, the platform provides a self-contained onboarding skill located at `.opencode/skills/agentic-lens-setup`.

#### Step 1: Install or Package the Setup Skill
Ensure the `agentic-lens-setup` skill is placed into your repository at `.opencode/skills/agentic-lens-setup/`. The directory structure contains:
```
.opencode/skills/agentic-lens-setup/
├── SKILL.md
├── bin/
│   └── .gitignore
└── scripts/
    └── download.sh
```

#### Step 2: Execute the Setup Skill
In your OpenCode conversation, prompt the agent:
```
Load the agentic-lens-setup skill and configure AgenticLens for this repository.
```

The skill executes the following sequence autonomously:
1. **Host Detection and Binary Download**: Runs `./.opencode/skills/agentic-lens-setup/scripts/download.sh` to determine host operating system (Linux, macOS, Windows) and architecture (amd64, arm64). It downloads the latest precompiled release binary from GitHub Releases directly into `.opencode/skills/agentic-lens-setup/bin/agentic-lens` and applies executable permissions (`chmod +x`).
2. **Protective Git Exclusion**: Verifies `.opencode/skills/agentic-lens-setup/bin/.gitignore` exists with patterns `*` and `!/.gitignore` to prevent precompiled binaries from being tracked in git.
3. **Interactive Project Interview**: Inspects the repository and queries the developer for project parameters (project name, ticket prefixes, workspaces directory, and monitored source folders).
4. **Configuration Scaffolding**: Executes `./.opencode/skills/agentic-lens-setup/bin/agentic-lens init` to scaffold `.agenticlens-config.yaml`.
5. **Mandatory User Confirmation Gate**: Inspects existing repository guidelines (`AGENTS.md`, `CLAUDE.md`, or `.opencode/agent/*.md`), prepares a proposed telemetry integration policy block, presents the exact text to the developer, and requests explicit confirmation before modifying any files.

### 4.2 Pathway B: Manual Developer Installation via GitHub Releases

Developers who prefer direct manual installation can retrieve official precompiled binaries directly from GitHub Releases.

#### Step 1: Retrieve Precompiled Release Binary
Navigate to the releases page:
`https://github.com/NiltonDuarte/AgenticLens/releases`

Select the archive corresponding to your platform:
- Linux x86_64: `agentic-lens-<VERSION>-linux-amd64.tar.gz`
- Linux ARM64: `agentic-lens-<VERSION>-linux-arm64.tar.gz`
- macOS Apple Silicon: `agentic-lens-<VERSION>-darwin-arm64.tar.gz`
- macOS Intel: `agentic-lens-<VERSION>-darwin-amd64.tar.gz`
- Windows x86_64: `agentic-lens-<VERSION>-windows-amd64.zip`

#### Step 2: Extract and Install Binary
Create a local binary directory (e.g. `./bin/`) and extract the archive:

On Linux and macOS:
```bash
mkdir -p ./bin
tar -xzf agentic-lens-*-linux-amd64.tar.gz -C ./bin
chmod +x ./bin/agentic-lens
```

On Windows:
```powershell
New-Item -ItemType Directory -Force -Path .\bin
Expand-Archive -Path .\agentic-lens-*-windows-amd64.zip -DestinationPath .\bin
```

Alternatively, use the repository download script:
```bash
./scripts/download-agentic-lens.sh --dest ./bin
```

#### Step 3: Verify Binary Execution
Confirm that the binary runs and outputs valid build metadata:
```bash
./bin/agentic-lens version --json
```

Expected JSON response:
```json
{
  "version": "v1.0.0",
  "commit": "a1b2c3d",
  "build_date": "2026-10-22T12:00:00Z"
}
```

#### Step 4: Scaffold Configuration
Run the `init` command to generate `.agenticlens-config.yaml`:
```bash
./bin/agentic-lens init \
  --project-name "MyProject" \
  --ticket-prefixes "PROJ,GEN,PO" \
  --workspaces "workspaces" \
  --src-folders "src,pkg"
```

---

## 5. Configuration Reference (.agenticlens-config.yaml)

AgenticLens is configured via `.agenticlens-config.yaml` located at the repository root. Below is the complete schema reference detailing all sections, fields, types, defaults, and operational semantics.

### 5.1 Root Configuration Overview

```yaml
# AgenticLens Global Configuration Reference
project_name: MyProject

ticket:
  prefixes:
    - prefix: PROJ
      description: Primary project development features and bugfixes
      ticket_type: feat
    - prefix: GEN
      description: General maintenance, operational tasks, and ad-hoc chores
      ticket_type: chore
    - prefix: PO
      description: Product specifications and requirements documentation
      ticket_type: docs
  case_insensitive: true
  custom_patterns: []

workspaces:
  dir: workspaces
  src_folders:
    - src
    - pkg
  artifacts:
    - path: CONTRACT.md
      title: Technical Contract
      description: Frozen technical contract and scope invariants
    - path: SUMMARY.md
      title: Work Summary
      description: Implementation summary log and verification records

database:
  path: ~/.local/share/opencode/opencode.db
  query_timeout_ms: 5000
  busy_timeout_ms: 5000
  read_only: true

live:
  active_threshold_seconds: 300
  poll_interval_ms: 2000
  max_trace_buffer_messages: 500
  anomaly_thresholds:
    stall_threshold_seconds: 120
    error_rate_threshold: 0.25
    context_exhaustion_pct: 0.85

code_metrics:
  base_branch: origin/main
  include: []
  exclude: []
  test_patterns: []
  prod_patterns: []

role_mappings:
  orchestrator: orchestrator
  build: orchestrator
  planner: planner
  coder: coder
  tester: tester
  reviewer: reviewer

role_order:
  - orchestrator
  - planner
  - coder
  - tester
  - reviewer
```

### 5.2 Section: `project_name`
- **Type**: String
- **Default**: None (Mandatory during `init`)
- **Semantics**: The human-readable name of the project. Displayed in web dashboard navigation bars, page titles, CLI summary tables, and JSON export metadata.

### 5.3 Section: `ticket`
Controls ticket identifier parsing, prefix matching, and normalization.
- **`prefixes`** (List of Objects or Strings): Recognized ticket prefixes.
  - `prefix` (String): Prefix string (e.g. `PROJ`, `AL`, `GEN`, `PO`, `FIX`).
  - `description` (String): Operational description of the ticket domain.
  - `ticket_type` (String, Optional): Default conventional commit type inferred for this prefix (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).
  - *Semantics*: All configured prefixes are treated as equal peers.
- **`case_insensitive`** (Boolean, Default: `true`): When true, matches ticket identifiers regardless of casing (e.g. `proj-123` normalizes to `PROJ-123`).
- **`custom_patterns`** (List of Strings, Default: `[]`): Additional regular expressions to match unconventional ticket naming formats when standard `<PREFIX>-<NUMBER>` is insufficient.

### 5.4 Section: `workspaces`
Defines filesystem layout for ticket telemetry and change tracking.
- **`dir`** (String, Default: `workspaces`): Directory where ticket workspace folders live (e.g. `workspaces/<TICKET_ID>/`).
- **`src_folders`** (List of Strings, Mandatory during `init`): Monitored source code folders. Git diffs, line additions, and line deletions occurring outside these folders are filtered out when calculating project code metrics.
- **`artifacts`** (List of Objects, Default: `[]`): Workspace documentation files tracked and displayed in the web dashboard artifact viewer.
  - `path` (String): File path relative to the ticket workspace (e.g. `CONTRACT.md`).
  - `title` (String): Display title in web dashboard tabs.
  - `description` (String, Optional): Explanatory subtitle.

### 5.5 Section: `database`
Specifies local OpenCode SQLite database connectivity parameters.
- **`path`** (String, Default: `~/.local/share/opencode/opencode.db`): Filesystem path to the OpenCode SQLite database. Supports home directory tilde expansion (`~`).
- **`query_timeout_ms`** (Integer, Default: `5000`): Maximum execution duration in milliseconds for database queries before timing out.
- **`busy_timeout_ms`** (Integer, Default: `5000`): SQLite busy handler wait duration in milliseconds before returning a busy lock error.
- **`read_only`** (Boolean, Default: `true`): Opens database connection in immutable read-only mode to prevent accidental writes or lock contention with OpenCode.

### 5.6 Section: `live`
Configures real-time session polling and automated anomaly detection.
- **`active_threshold_seconds`** (Integer, Default: `300`): Sliding window in seconds. Sessions with turn activity within this window are classified as active.
- **`poll_interval_ms`** (Integer, Default: `2000`): Frequency in milliseconds at which the embedded server polls SQLite for new session turns.
- **`max_trace_buffer_messages`** (Integer, Default: `500`): Maximum number of recent trace events buffered in server memory per active session for live web streaming.
- **`anomaly_thresholds`**:
  - `stall_threshold_seconds` (Integer, Default: `120`): Elapsed seconds without turn activity in an active session before raising a stall anomaly.
  - `error_rate_threshold` (Float, Default: `0.25`): Ratio of tool errors to total invocations triggering a high error rate alert (e.g. 0.25 = 25%).
  - `context_exhaustion_pct` (Float, Default: `0.85`): Token usage percentage of model context window capacity triggering a saturation warning (e.g. 0.85 = 85%).

### 5.7 Section: `code_metrics`
Configures git commit diff extraction and file classification.
- **`base_branch`** (String, Optional, Default: automatic probing): Comparison branch for computing diff metrics. If omitted or empty, AgenticLens probes `origin/HEAD`, `origin/main`, `main`, `origin/master`, and `master`.
- **`include`** (List of Strings, Default: `[]`): Glob patterns explicitly included as production code.
- **`exclude`** (List of Strings, Default: `[]`): Glob patterns excluded from diff and code volume calculations (e.g. vendor folders, minified assets).
- **`test_patterns`** (List of Strings, Default: `[]`): Glob patterns identifying test code and fixtures.
- **`prod_patterns`** (List of Strings, Default: `[]`): Glob patterns identifying production runtime code.

### 5.8 Section: `role_mappings`
Maps raw agent persona strings recorded in OpenCode traces to canonical role categories (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`).
- Unmapped personas are categorized under their raw name or as unassigned subagents.
- Custom personas (e.g. `security-auditor`, `db-architect`) can be mapped directly to canonical roles or preserved as standalone custom roles.

### 5.9 Section: `role_order`
- **Type**: List of Strings
- **Default**: `[orchestrator, planner, coder, tester, reviewer]`
- **Semantics**: Explicit sorting sequence for agent roles in CLI tables, summary markdowns, and web dashboard charts. Listed roles appear first in the specified order; any unlisted roles follow in alphabetical order.

---

## 6. Agnostic Agent Workflow Integration Guide

Integrating AgenticLens into your repository requires zero changes to agent reasoning loops, model selection, or coding styles. Integration is achieved through a single command invocation at the conclusion of each task.

### 6.1 The Delayed Telemetry Ingestion Mandate

The most critical operational principle for accurate telemetry is the **Delayed Telemetry Ingestion Mandate**:
- Telemetry generation must always be executed as the very final turn of an agent session, immediately before committing work to git.
- Running telemetry generation early (for example, mid-session or during code generation) causes all subsequent turns, tool calls, compactions, and subagent handoffs to be permanently omitted from the session ledger.
- The command sequence must follow this exact order:
  1. Complete implementation, testing, and reviews.
  2. Execute `./bin/agentic-lens generate ...` to record final telemetry facts into `agentic_metrics.json` and append the markdown table to `SUMMARY.md`.
  3. Execute `git add -A && git commit -m "..."` to commit code and metrics atomically.

### 6.2 Telemetry Ingestion Command (`generate`)

To record session metrics, invoke the `generate` subcommand:

```bash
NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
./bin/agentic-lens generate \
  --nonce "$NONCE" \
  --ticket-id "<TICKET_ID>" \
  --ticket-type "<CONVENTIONAL_TYPE>" \
  --json-output "workspaces/<TICKET_ID>/agentic_metrics.json" \
  --append "workspaces/<TICKET_ID>/SUMMARY.md"
```

#### Command Flags Breakdown:
- `--ticket-id <ID>` (Required): The active ticket or issue identifier (e.g. `PROJ-42`, `BUG-101`).
- `--nonce <STRING>` (Recommended): A unique execution nonce used to correlate concurrent subagent sessions and resolve the precise execution window.
- `--ticket-type <TYPE>` (Required): Conventional commit classification (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).
- `--json-output <PATH>`: Destination path for the immutable JSON telemetry ledger.
- `--append <PATH>`: Destination path for appending an idempotent Markdown verification table.
- `--db-path <PATH>` (Optional): Explicit path to OpenCode SQLite database (overrides configuration).
- `--anomalies` (Optional): Evaluate and record anomaly alerts directly into the output ledger.

### 6.3 Conventional Commit Type Classification

AgenticLens requires explicit commit type classification to generate accurate engineering velocity rollups:
- `feat`: New product capabilities, user-facing enhancements, or API additions.
- `fix`: Defect remediation, bug fixes, and patch corrections.
- `docs`: Documentation updates, specifications, or architectural manuals.
- `chore`: Maintenance tasks, dependency updates, configuration changes, or build scripts.
- `refactor`: Code reorganization without functional behavior changes.
- `test`: Adding, updating, or fixing automated tests.
- `ci`: Continuous integration and delivery pipeline modifications.
- `perf`: Performance optimizations and latency reductions.
- `build`: Build toolchain modifications or compiler flag updates.

### 6.4 Integrating Custom Personas

If your organization uses custom agent personas (for example, `security-scanner`, `perf-analyzer`, or `frontend-specialist`), map them in `.agenticlens-config.yaml`:

```yaml
role_mappings:
  security-scanner: reviewer
  perf-analyzer: tester
  frontend-specialist: coder
```

When AgenticLens processes execution traces, it automatically aggregates tokens, costs, and durations under the designated canonical role while preserving the granular persona identifier in the raw telemetry ledger.

### 6.5 Example Agent Integration Guidelines (`AGENTS.md`)

Below is the standard policy block that can be added to your repository's agent instructions file (`AGENTS.md`, `CLAUDE.md`, or `.opencode/agent/*.md`):

```markdown
## AgenticLens Telemetry Guidelines

1. Ticket Tracking Mandate:
   - Every task must be tracked by an explicit ticket identifier matching configured prefixes (e.g. PROJ-101).
   - Agents work on one ticket at a time. All turns belong to the active ticket.

2. Dedicated Workspace Allocation:
   - Store task documentation and metrics in `workspaces/<TICKET_ID>/`.
   - Maintain `CONTRACT.md` (scope boundaries) and `SUMMARY.md` (execution log).

3. Delayed Telemetry Ingestion:
   - Ingest telemetry as the very final command before committing code:
     NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
     ./bin/agentic-lens generate \
       --nonce "$NONCE" \
       --ticket-id "<TICKET_ID>" \
       --ticket-type "<CONVENTIONAL_TYPE>" \
       --json-output "workspaces/<TICKET_ID>/agentic_metrics.json" \
       --append "workspaces/<TICKET_ID>/SUMMARY.md"
   - Stage deliverables and telemetry together: `git add -A && git commit -m "<message>"`.
```

---

## 7. CLI and Web Dashboard Reference

AgenticLens provides a unified command-line interface and an embedded web dashboard for observability.

### 7.1 CLI Subcommands Reference

#### `agentic-lens init`
Scaffolds and writes `.agenticlens-config.yaml` at the repository root.
- Flags:
  - `--project-name <NAME>` (Required): Human-readable name of the project.
  - `--ticket-prefixes <LIST>` (Required): Comma-separated list of ticket prefixes (e.g. `PROJ,GEN,PO`).
  - `--src-folders <LIST>` (Required): Comma-separated list of monitored source code directories.
  - `--workspaces <DIR>` (Optional, Default: `workspaces`): Root directory for ticket workspaces.
  - `--db-path <PATH>` (Optional): Custom path to OpenCode SQLite database.
  - `--output <FILE>` (Optional, Default: `.agenticlens-config.yaml`): Destination output file.
  - `--force` (Optional): Overwrite existing configuration file without prompting.
  - `--dry-run` (Optional): Print generated configuration YAML to stdout without writing to disk.

#### `agentic-lens generate`
Ingests session telemetry from SQLite and git commit history into a workspace ledger.
- Flags:
  - `--ticket-id <ID>` (Required): Active ticket identifier.
  - `--ticket-type <TYPE>` (Required): Conventional commit classification.
  - `--nonce <STRING>` (Optional): Unique execution nonce for session resolution.
  - `--session-id <ID>` (Optional): Explicit OpenCode session ID override.
  - `--ticket-title <TITLE>` (Optional): Title of the ticket.
  - `--strict-type` (Optional): Fail with error if ticket type does not match conventional taxonomy.
  - `--json-output <FILE>` (Optional): Destination path for `agentic_metrics.json`.
  - `--append <FILE>` (Optional): Destination Markdown file for summary table append.
  - `--anomalies` (Optional): Run anomaly detection rules and record results in output ledger.
  - `--allow-empty` (Optional): Allow generation when zero git diffs or zero session turns exist.

#### `agentic-lens scan` (or root command)
Scans workspace directories and displays summary tables of all recorded ticket ledgers.
- Flags:
  - `--list` (Optional): Scan and print tabular overview of all discovered ticket workspaces.
  - `--ticket-id <ID>` (Optional): Filter scan output to a single ticket identifier.
  - `--workspaces <DIR>` (Optional): Override workspaces directory path.
  - `--anomalies` (Optional): Display detected anomalies across scanned workspaces.
  - `--json-output <FILE>` (Optional): Export aggregated scan results as JSON.
  - `--append <FILE>` (Optional): Append scan results table to a Markdown document.

#### `agentic-lens server`
Launches the embedded HTTP server and real-time SSE live streaming service.
- Flags:
  - `--port <PORT>` (Optional, Default: `8080`): TCP port to bind HTTP server.
  - `--workspaces <DIR>` (Optional): Override workspaces directory path.
  - `--db-path <PATH>` (Optional): Override SQLite database path.

#### `agentic-lens version`
Prints binary version, build metadata, and schema version.
- Flags:
  - `--json` (Optional): Output metadata formatted as JSON.
  - `--short` (Optional): Output version string only.

#### `agentic-lens backfill`
Batch recomputes git diff metrics, recalculates token costs against updated model pricing, or upgrades schema across historical workspace ledgers.
- Flags:
  - `--workspaces <DIR>` (Optional): Directory containing ticket workspaces.
  - `--ticket-id <ID>` (Optional): Limit backfill to a single ticket.
  - `--recompute-git` (Optional): Re-extract git diff metrics from git commit history.
  - `--recompute-db` (Optional): Re-query OpenCode SQLite database for session turns and tokens.
  - `--migrate-schema` (Optional): Upgrade legacy ledger files to current schema version.
  - `--dry-run` (Optional): Simulate backfill operations without modifying disk files.
  - `--update-summaries` (Optional): Update `SUMMARY.md` tables with recomputed values.
  - `--force` (Optional): Bypass safety confirmation checks.

#### `agentic-lens migrate`
Runs stepwise schema migrations across ticket workspace ledgers.
- Flags:
  - `--workspaces <DIR>` (Optional): Directory containing ticket workspaces.
  - `--ticket-id <ID>` (Optional): Limit migration to a specific ticket.
  - `--target-version <VER>` (Optional): Target schema version to migrate toward.
  - `--dry-run` (Optional): Preview migrations without modifying files.
  - `--backup` (Optional): Create `.bak` backup copies before modifying ledgers.
  - `--update-summaries` (Optional): Refresh Markdown summary tables after migration.

### 7.2 Embedded Web Dashboard Views

When running `agentic-lens server --port 8080`, navigate to `http://localhost:8080` to access the interactive dashboard.

#### 1. Fleet Overview (`/`)
- Cross-ticket aggregate rollups showing total tokens consumed, total financial cost, total development wall-clock time, and average iterations per ticket.
- Context window health distribution gauge displaying percentages across HEALTHY, ELEVATED, BLOATED, and CRITICAL states.
- Interactive ticket inventory with search filtering by ticket ID, title, status, and prefix tags.

#### 2. Live Observability (`/live`)
- Real-time active agent monitoring powered by Server-Sent Events (SSE).
- Live session cards displaying active subagent personas, model in use, current turn count, token accumulation, and elapsed wall-clock duration.
- Real-time trace viewer streaming tool calls, parameter payloads, and output logs.
- Active anomaly alert banners highlighting detected stalls, elevated error rates, and context exhaustion in real time.

#### 3. Ticket Detail (`/ticket/:id`)
- Deep-dive analytics for an individual ticket workspace.
- Multi-iteration timeline breaking down SDLC passes, subagent roles, and turn progression.
- Tool invocation distribution chart detailing tool call frequencies and failure rates.
- Model usage and cost breakdown (e.g. GPT-4o, Claude 3.5 Sonnet, Gemini Flash).
- Interactive Artifact Viewer rendering `CONTRACT.md` and `SUMMARY.md` as formatted Markdown with raw view toggles.

#### 4. Ticket Comparisons (`/comparisons`)
- Side-by-side comparative analysis of two or more tickets.
- Differential metrics evaluating token efficiency, cost variations, tool reliability, and code diff volume per iteration.

#### 5. Metric Progression (`/progression`)
- Longitudinal trend graphs tracking agent efficiency over time.
- Rolling averages for cost-per-ticket, turn velocity, test-to-production code ratios, and context window health evolution.

---

## 8. Troubleshooting and Operational Best Practices

### 8.1 SQLite Concurrency and Database Locks

AgenticLens accesses the local OpenCode SQLite database concurrently while OpenCode writes session traces:
- **Busy Timeouts**: If OpenCode holds an exclusive lock during a database transaction, SQLite queries can experience transient lock contention. AgenticLens mitigates this using a configurable busy timeout handler (`database.busy_timeout_ms`, default 5000 ms).
- **WAL Mode**: OpenCode operates in SQLite Write-Ahead Logging (WAL) mode, allowing concurrent readers without blocking active writers.
- **Read-Only Connections**: Keep `database.read_only: true` enabled in `.agenticlens-config.yaml`. This guarantees that AgenticLens opens connections with `SQLITE_OPEN_READONLY`, ensuring zero lock escalation and eliminating the risk of corrupting OpenCode state.

### 8.2 Cross-Platform Database Paths

The OpenCode SQLite database location varies across operating systems. Configure `database.path` accordingly:
- **Linux**: `~/.local/share/opencode/opencode.db`
- **macOS**: `~/Library/Application Support/opencode/opencode.db`
- **Windows**: `%APPDATA%\opencode\opencode.db` (or `C:\Users\<USER>\AppData\Roaming\opencode\opencode.db`)

AgenticLens automatically performs tilde expansion (`~`) on Linux and macOS, resolving paths to the user home directory. If running in a containerized or CI environment, set the database path via command-line flag: `--db-path /path/to/opencode.db`.

### 8.3 Executable Permissions and Binary Placement

When installing precompiled binaries manually on Unix systems (Linux and macOS):
- Ensure the executable bit is set: `chmod +x ./bin/agentic-lens`.
- If gatekeeper blocks execution on macOS, clear quarantine flags:
  ```bash
  xattr -d com.apple.quarantine ./bin/agentic-lens
  ```
- Store the binary in a dedicated directory (such as `./bin/` or `.opencode/skills/agentic-lens-setup/bin/`) and ensure a protective `.gitignore` contains `*` and `!/.gitignore` to prevent tracking compiled binaries in git.

### 8.4 Git Upstream Base Branch Inference

When extracting commit diff metrics, AgenticLens compares the active branch against an upstream base branch:
- By default, if `code_metrics.base_branch` is omitted or empty, AgenticLens automatically probes:
  1. `origin/HEAD`
  2. `origin/main`
  3. `main`
  4. `origin/master`
  5. `master`
- In CI or detached HEAD environments where remote tracking branches may not be fetched, explicitly specify the base branch in `.agenticlens-config.yaml`:
  ```yaml
  code_metrics:
    base_branch: origin/main
  ```
- Ensure git fetch includes the base branch in CI pipelines (`git fetch origin main`).

### 8.5 Workspace Directory Path Resolution

AgenticLens resolves workspace directories relative to the current working directory or git repository root:
- Paths with spaces (e.g. `my workspaces/ticket 1`) are fully supported by wrapping arguments in quotes:
  ```bash
  ./bin/agentic-lens --workspaces "my workspaces"
  ```
- Always verify that the configured `workspaces.dir` in `.agenticlens-config.yaml` matches the directory used in CLI commands and agent prompts.
