# AgenticLens Platform Setup Manual and Telemetry Integration Guide

AgenticLens is a developer telemetry and observability platform built specifically for OpenCode workflows. It provides real-time monitoring, longitudinal cost and token tracking, automated anomaly detection, and context window health diagnostics across autonomous coding agents running in OpenCode.

AgenticLens operates by extracting execution traces directly from OpenCode's local SQLite database (`~/.local/share/opencode/opencode.db`), specifically querying the `session`, `message`, and `part` tables. Consequently, AgenticLens strictly requires an active OpenCode installation and its local SQLite database; it cannot function without OpenCode and does not support alternative agent harnesses.

This manual serves as the comprehensive, authoritative guide for engineering teams, platform architects, and autonomous coding agents adopting AgenticLens in OpenCode projects.

---

## 1. Quick Start and Setup Pathways

AgenticLens offers two primary setup pathways: an autonomous agentic setup via a shared skill manifest, and a direct manual developer installation using official precompiled release archives. Target repositories never need to compile code from source or install Go, Node.js, npm, or C compiler toolchains.

### 1.1 Pathway A: Agentic Distribution and Setup via SKILL.md

In modern agentic workflows, autonomous agents consume skills natively through Markdown instruction manifests (`SKILL.md`) accessed directly via shared repository links or raw URLs. No archive packaging, zip files, or custom unpacking scripts are required.

#### Step 1: Share Skill Reference with Your Agent
Provide your autonomous coding agent in OpenCode with the reference to the AgenticLens setup skill. You can either:
- Point the agent directly to the canonical remote URL:
  `https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/.opencode/skills/agentic-lens-setup/SKILL.md`
- Or copy `.opencode/skills/agentic-lens-setup/SKILL.md` into your repository.

Prompt your agent:
```
Read the instructions at https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/.opencode/skills/agentic-lens-setup/SKILL.md and set up AgenticLens in this repository.
```

#### Step 2: Autonomous Host Detection and Binary Retrieval
The agent inspects the host platform and architecture using standard system utilities:
```bash
OS_RAW=$(uname -s)
ARCH_RAW=$(uname -m)

case "$OS_RAW" in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*|Windows_NT) OS="windows" ;;
  *) echo "Unsupported OS: $OS_RAW" >&2; exit 1 ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported Architecture: $ARCH_RAW" >&2; exit 1 ;;
esac
```

The agent resolves the release artifact from GitHub Releases:
- Linux x86_64: `agentic-lens-linux-amd64.tar.gz`
- Linux ARM64: `agentic-lens-linux-arm64.tar.gz`
- macOS Apple Silicon: `agentic-lens-darwin-arm64.tar.gz`
- macOS Intel: `agentic-lens-darwin-amd64.tar.gz`
- Windows x86_64: `agentic-lens-windows-amd64.zip`

The agent creates the binary destination directory, downloads the archive, extracts the binary, and sets executable permissions:
```bash
DEST_DIR=".opencode/skills/agentic-lens-setup/bin"
mkdir -p "$DEST_DIR"

# Write protective .gitignore to prevent committing binaries to version control
cat <<'EOF' > "$DEST_DIR/.gitignore"
*
!.gitignore
EOF

# Download and unpack precompiled binary
RELEASE_URL="https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/agentic-lens-${OS}-${ARCH}.tar.gz"
curl -fsSL "$RELEASE_URL" | tar -xz -C "$DEST_DIR"
chmod +x "$DEST_DIR/agentic-lens"
```

#### Step 3: Verify Binary Execution
The agent validates that the executable operates cleanly on the host system:
```bash
./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json
```
Expected output:
```json
{
  "version": "v1.0.0",
  "commit": "a1b2c3d",
  "build_date": "2026-10-22T12:00:00Z"
}
```

#### Step 4: Project Scaffolding
The agent queries the developer or inspects the codebase to determine:
- Project Name (e.g. `BillingService`)
- Ticket prefixes (e.g. `BILL,PAY,GEN`)
- Monitored source code folders (e.g. `services,pkg,cmd`)
- Workspaces directory (default: `workspaces`)

The agent runs `init` to write `.agenticlens-config.yaml`:
```bash
./.opencode/skills/agentic-lens-setup/bin/agentic-lens init \
  --project-name "BillingService" \
  --ticket-prefixes "BILL,PAY,GEN" \
  --src-folders "services,pkg,cmd" \
  --workspaces "workspaces"
```

#### Step 5: Mandatory User Confirmation Gate
Before updating or creating any OpenCode agent policy files (`AGENTS.md` or `.opencode/agent/*.md`), the agent must present the exact policy text to the user and receive explicit approval.

---

### 1.2 Pathway B: Manual Developer Installation via GitHub Releases

Developers who prefer direct manual installation without an agent setup skill can install AgenticLens in three quick steps.

#### Step 1: Download Release Archive
Navigate to official GitHub Releases:
`https://github.com/NiltonDuarte/AgenticLens/releases`

Download the appropriate archive for your operating system and architecture.

Alternatively, download via `curl`:
```bash
# Example for Linux amd64
curl -fsSL -o agentic-lens.tar.gz "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/agentic-lens-linux-amd64.tar.gz"

# Example for macOS Apple Silicon
curl -fsSL -o agentic-lens.tar.gz "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/agentic-lens-darwin-arm64.tar.gz"
```

#### Step 2: Extract and Set Executable Permissions
```bash
mkdir -p ./bin
tar -xzf agentic-lens.tar.gz -C ./bin
chmod +x ./bin/agentic-lens
rm agentic-lens.tar.gz
```

On Windows (PowerShell):
```powershell
New-Item -ItemType Directory -Force -Path .\bin
Invoke-WebRequest -Uri "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/agentic-lens-windows-amd64.zip" -OutFile ".\agentic-lens.zip"
Expand-Archive -Path ".\agentic-lens.zip" -DestinationPath ".\bin"
Remove-Item ".\agentic-lens.zip"
```

#### Step 3: Scaffold Configuration and Launch Dashboard
Verify the binary:
```bash
./bin/agentic-lens version --short
```

Initialize your repository configuration:
```bash
./bin/agentic-lens init \
  --project-name "MyProject" \
  --ticket-prefixes "PROJ,GEN" \
  --src-folders "src,pkg" \
  --workspaces "workspaces"
```

Start the embedded dashboard:
```bash
./bin/agentic-lens server --port 8085
```
Open `http://localhost:8085` in your browser to view the Fleet Overview and Live Observability feeds.

---

## 2. Agnostic Adoption for OpenCode Projects

AgenticLens is engineered specifically for OpenCode environments. It is hardwired to OpenCode internals: its telemetry ingestion engine, session resolver, and live SSE streaming service directly query OpenCode's local SQLite database (`session`, `message`, and `part` tables). It cannot function without an active OpenCode installation and does not support alternative agent harnesses.

While tightly coupled to OpenCode at the platform layer, AgenticLens imposes zero engineering methodology constraints on adopting teams. Within OpenCode, it is completely agnostic to how teams structure their development workflows.

### 2.1 Complete Methodology Neutrality within OpenCode
Every engineering team operating in OpenCode has distinct development lifecycles, branching strategies, and agent configurations. AgenticLens accommodates any way of working:
- **Workflow Topology Agnostic**: Operates equally well with single-agent prompt loops, human-in-the-loop interactions, hierarchical multi-agent delegations, or automated CI/CD maintenance pipelines running inside OpenCode.
- **Decoupled from Maintainer Personas**: Adopters do not inherit this repository's internal maintainer personas (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`). Teams configure their own agents or subagents freely.
- **Review and QA Independence**: Adopters define their own testing standards, linting gates, and code review criteria. AgenticLens records what occurred during execution without prescribing how development must be managed.

### 2.2 Custom Conventions and Taxonomy
Adopting repositories configure AgenticLens to align with their existing organizational standards:
- **Custom Ticket Prefixes**: Support for any ticketing taxonomy (e.g. `JIRA-`, `ISSUE-`, `FEAT-`, `BUG-`, `CORE-`, `PAY-`, `DEVOPS-`). Multiple prefixes are treated as equal peers.
- **Custom Workspace Hierarchies**: Telemetry can be housed in `workspaces/`, `alspaces/`, `.tasks/`, `specs/`, or any custom repository path.
- **Custom Git Branching**: Operates seamlessly with trunk-based development, GitFlow, GitHub flow, or short-lived feature branches.
- **Custom Persona Role Mapping**: Arbitrary agent persona strings configured in OpenCode can be mapped into canonical reporting roles via `role_mappings` in `.agenticlens-config.yaml` or tracked as standalone custom roles.

---

## 3. Configuration Reference (.agenticlens-config.yaml)

AgenticLens is configured through a YAML manifest located at the repository root (`.agenticlens-config.yaml`).

### 3.1 Complete Configuration Schema Example

```yaml
# AgenticLens Global Configuration Reference
project_name: PaymentGateway

ticket:
  prefixes:
    - prefix: PAY
      description: Payment core transactions and ledger logic
      ticket_type: feat
    - prefix: SEC
      description: Security scanning, vulnerability remediation, and audits
      ticket_type: fix
    - prefix: OPS
      description: Infrastructure, deployment pipelines, and chores
      ticket_type: chore
    - prefix: DOCS
      description: Architecture decision records and API references
      ticket_type: docs
  case_insensitive: true
  custom_patterns: []

workspaces:
  dir: workspaces
  src_folders:
    - src
    - pkg
    - cmd
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
  security-scanner: security
  perf-analyzer: tester
  integration-tester: tester
  frontend-specialist: developer
  backend-specialist: developer

role_order:
  - security
  - developer
  - tester
```

### 3.2 Schema Field Definitions

#### `project_name`
- **Type**: String
- **Default**: None (Mandatory during `init`)
- **Description**: Human-readable name of the project. Displayed in web dashboard headers, page titles, CLI summary tables, and JSON export metadata.

#### `ticket`
Controls ticket identifier parsing, prefix matching, and normalization.
- **`prefixes`** (List of Objects or Strings): Recognized ticket prefixes.
  - `prefix` (String): Prefix string (e.g. `PAY`, `SEC`, `OPS`).
  - `description` (String): Operational description of the ticket domain.
  - `ticket_type` (String, Optional): Default conventional commit type inferred for this prefix (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).
- **`case_insensitive`** (Boolean, Default: `true`): When true, matches ticket identifiers regardless of casing (e.g. `pay-101` normalizes to `PAY-101`).
- **`custom_patterns`** (List of Strings, Default: `[]`): Additional regular expressions to match custom ticket naming formats when standard `<PREFIX>-<NUMBER>` is insufficient.

#### `workspaces`
Defines filesystem layout for ticket telemetry and change tracking.
- **`dir`** (String, Default: `workspaces`): Directory where ticket workspace folders live (e.g. `workspaces/<TICKET_ID>/`).
- **`src_folders`** (List of Strings, Mandatory during `init`): Monitored source code folders. Git diffs, line additions, and line deletions occurring outside these folders are filtered out when calculating project code metrics.
- **`artifacts`** (List of Objects, Default: `[]`): Workspace documentation files tracked and displayed in the web dashboard artifact viewer.
  - `path` (String): File path relative to the ticket workspace (e.g. `CONTRACT.md`).
  - `title` (String): Display title in web dashboard tabs.
  - `description` (String, Optional): Explanatory subtitle.

#### `database`
Specifies local OpenCode SQLite database connectivity parameters.
- **`path`** (String, Default: `~/.local/share/opencode/opencode.db`): Filesystem path to the OpenCode SQLite database. Supports home directory tilde expansion (`~`).
- **`query_timeout_ms`** (Integer, Default: `5000`): Maximum execution duration in milliseconds for database queries before timing out.
- **`busy_timeout_ms`** (Integer, Default: `5000`): SQLite busy handler wait duration in milliseconds before returning a busy lock error.
- **`read_only`** (Boolean, Default: `true`): Opens database connection in immutable read-only mode to prevent accidental writes or lock contention with OpenCode.

#### `live`
Configures real-time session polling and automated anomaly detection.
- **`active_threshold_seconds`** (Integer, Default: `300`): Sliding window in seconds. Sessions with turn activity within this window are classified as active.
- **`poll_interval_ms`** (Integer, Default: `2000`): Frequency in milliseconds at which the embedded server polls SQLite for new session turns.
- **`max_trace_buffer_messages`** (Integer, Default: `500`): Maximum number of recent trace events buffered in server memory per active session for live web streaming.
- **`anomaly_thresholds`**:
  - `stall_threshold_seconds` (Integer, Default: `120`): Elapsed seconds without turn activity in an active session before raising a stall anomaly.
  - `error_rate_threshold` (Float, Default: `0.25`): Ratio of tool errors to total invocations triggering a high error rate alert (e.g. 0.25 = 25%).
  - `context_exhaustion_pct` (Float, Default: `0.85`): Token usage percentage of model context window capacity triggering a saturation warning (e.g. 0.85 = 85%).

#### `code_metrics`
Configures git commit diff extraction and file classification.
- **`base_branch`** (String, Optional, Default: automatic probing): Comparison branch for computing diff metrics. If omitted or empty, AgenticLens probes `origin/HEAD`, `origin/main`, `main`, `origin/master`, and `master`.
- **`include`** (List of Strings, Default: `[]`): Glob patterns explicitly included as production code.
- **`exclude`** (List of Strings, Default: `[]`): Glob patterns excluded from diff and code volume calculations (e.g. vendor folders, minified assets).
- **`test_patterns`** (List of Strings, Default: `[]`): Glob patterns identifying test code and fixtures.
- **`prod_patterns`** (List of Strings, Default: `[]`): Glob patterns identifying production runtime code.

#### `role_mappings`
Maps raw agent persona strings recorded in OpenCode traces to canonical role categories:
- Unmapped personas are categorized under their raw name or as unassigned subagents.
- Custom personas can be mapped directly to canonical roles or preserved as standalone custom roles.

#### `role_order`
- **Type**: List of Strings
- **Default**: `[orchestrator, planner, coder, tester, reviewer]`
- **Description**: Explicit sorting sequence for agent roles in CLI tables, summary markdowns, and web dashboard charts. Listed roles appear first in the specified order; any unlisted roles follow in alphabetical order.

---

## 4. OpenCode Agent Workflow Integration Guide

Integrating AgenticLens into your repository requires zero changes to agent reasoning loops, model selection, or coding styles. Integration is achieved through a single command invocation at the conclusion of each task.

### 4.1 The Delayed Telemetry Ingestion Mandate

The most critical operational principle for accurate telemetry is the **Delayed Telemetry Ingestion Mandate**:
- Telemetry generation must always be executed as the very final turn of an agent session, immediately before committing work to git.
- Running telemetry generation early (for example, mid-session or during code generation) causes all subsequent turns, tool calls, compactions, and subagent handoffs to be permanently omitted from the session ledger.
- The command sequence must follow this exact order:
  1. Complete implementation, testing, and reviews.
  2. Execute `./bin/agentic-lens generate ...` to record final telemetry facts into `agentic_metrics.json` and append the markdown table to `SUMMARY.md`.
  3. Execute `git add -A && git commit -m "..."` to commit code and metrics atomically.

### 4.2 Nonce Generation and Ingestion Command (`generate`)

To guarantee deterministic session resolution across concurrent agents or subagent delegations, pass a unique execution nonce directly to the `generate` command:

```bash
./bin/agentic-lens generate \
  --nonce <GENERATED_NONCE> \
  --ticket-id "<TICKET_ID>" \
  --ticket-type "<CONVENTIONAL_TYPE>" \
  --json-output "workspaces/<TICKET_ID>/agentic_metrics.json" \
  --append "workspaces/<TICKET_ID>/SUMMARY.md"
```

#### Literal Nonce Mandate (FIX-5)
The executing agent must generate a unique random nonce directly (e.g. an 8 to 16 character alphanumeric or hex string) and supply the concrete literal token in the command.

**Important**: Shell variable expansions (`$NONCE`) or command substitutions (`$(...)`) must never be used. OpenCode records tool input arguments verbatim into its SQLite database before shell variable expansion or subshell execution occurs. If an agent executes a command containing `$NONCE` or `$(...)`, OpenCode records the literal string `"$NONCE"` in the database part record. Subsequent database lookups matching the runtime expanded token against the database will fail to find a match. Always supply the literal nonce value directly.

#### Command Flags Breakdown:
- `--ticket-id <ID>` (Required): The active ticket or issue identifier (e.g. `PAY-42`, `SEC-101`).
- `--nonce <STRING>` (Recommended): A unique execution nonce used to correlate concurrent subagent sessions and resolve the precise execution window. Must be passed as a literal string without shell variable expansion.
- `--ticket-type <TYPE>` (Required): Conventional commit classification (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).
- `--json-output <PATH>`: Destination path for the immutable JSON telemetry ledger.
- `--append <PATH>`: Destination path for appending an idempotent Markdown verification table.
- `--db-path <PATH>` (Optional): Explicit path to OpenCode SQLite database (overrides configuration).
- `--anomalies` (Optional): Evaluate and record anomaly alerts directly into the output ledger.

### 4.3 Conventional Commit Type Classification

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

### 4.4 Custom Persona Role Mapping

Adopting repositories often utilize specialized agent personas. Map them in `.agenticlens-config.yaml` to unify cross-team reporting:

```yaml
role_mappings:
  security-scanner: security
  perf-analyzer: tester
  integration-tester: tester
  frontend-specialist: developer
  backend-specialist: developer
```

When AgenticLens processes execution traces, it automatically aggregates tokens, costs, and durations under the designated canonical role while preserving the granular persona identifier in the raw telemetry ledger.

### 4.5 Standard Repository Agent Guidelines Policy Block

Below is the standard policy block that can be added to your repository's OpenCode agent instructions file (`AGENTS.md` or `.opencode/agent/*.md`):

```markdown
## AgenticLens Telemetry Guidelines

1. Ticket Tracking Mandate:
   - Every task must be tracked by an explicit ticket identifier matching configured prefixes (e.g. PAY-101).
   - Agents work on one ticket at a time. All turns belong to the active ticket.

2. Dedicated Workspace Allocation:
   - Store task documentation and metrics in `workspaces/<TICKET_ID>/`.
   - Maintain `CONTRACT.md` (scope boundaries) and `SUMMARY.md` (execution log).

3. Delayed Telemetry Ingestion:
   - Ingest telemetry as the very final command before committing code.
   - The agent must generate a unique random nonce directly (e.g. an 8 to 16 character alphanumeric or hex string) and pass the literal token. Do not use shell variables (`$NONCE`) because OpenCode records tool inputs into SQLite before subshell expansion occurs:
     ./bin/agentic-lens generate \
       --nonce <GENERATED_NONCE> \
       --ticket-id "<TICKET_ID>" \
       --ticket-type "<CONVENTIONAL_TYPE>" \
       --json-output "workspaces/<TICKET_ID>/agentic_metrics.json" \
       --append "workspaces/<TICKET_ID>/SUMMARY.md"
   - Stage deliverables and telemetry together: `git add -A && git commit -m "<message>"`.
```

---

## 5. CLI Commands and Embedded Web Dashboard Reference

AgenticLens provides a unified command-line interface and an embedded web dashboard for observability.

### 5.1 CLI Subcommands Reference

#### `agentic-lens init`
Scaffolds and writes `.agenticlens-config.yaml` at the repository root.
- Flags:
  - `--project-name <NAME>` (Required): Human-readable name of the project.
  - `--ticket-prefixes <LIST>` (Required): Comma-separated list of ticket prefixes (e.g. `PAY,SEC,OPS`).
  - `--src-folders <LIST>` (Required): Comma-separated list of monitored source code directories.
  - `--workspaces <DIR>` (Optional, Default: `workspaces`): Root directory for ticket workspaces.
  - `--db-path <PATH>` (Optional): Custom path to OpenCode SQLite database.
  - `--output <FILE>` (Optional, Default: `.agenticlens-config.yaml`): Destination output file.
  - `--force` (Optional): Overwrite existing configuration file without prompting.
  - `--dry-run` (Optional): Print generated configuration YAML to stdout without writing to disk.

#### `agentic-lens generate`
Ingests session telemetry from OpenCode's SQLite database and git commit history into a workspace ledger.
- Flags:
  - `--ticket-id <ID>` (Required): Active ticket identifier.
  - `--ticket-type <TYPE>` (Required): Conventional commit classification.
  - `--nonce <STRING>` (Optional): Unique execution nonce for OpenCode session resolution. Must be passed as a directly generated literal token; shell variables (`$NONCE`) must not be used.
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
  - `--recompute-git` (Optional): Recompute Git code facts from an explicit ref; requires `--code-ref`.
  - `--code-ref <REF>` (Required with `--recompute-git`): Branch or commit to compare with the configured base branch.
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

---

### 5.2 Embedded Web Dashboard Views

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

## 6. Operational Troubleshooting and Best Practices

### 6.1 SQLite Concurrency, Busy Timeouts, and WAL Mode

AgenticLens accesses the local OpenCode SQLite database concurrently while OpenCode writes session traces:
- **Busy Timeouts**: If OpenCode holds an exclusive lock during a database transaction, SQLite queries can experience transient lock contention. AgenticLens mitigates this using a configurable busy timeout handler (`database.busy_timeout_ms`, default 5000 ms).
- **WAL Mode**: OpenCode operates in SQLite Write-Ahead Logging (WAL) mode, allowing concurrent readers without blocking active writers.
- **Read-Only Connections**: Keep `database.read_only: true` enabled in `.agenticlens-config.yaml`. This guarantees that AgenticLens opens connections with `SQLITE_OPEN_READONLY`, ensuring zero lock escalation and eliminating the risk of corrupting OpenCode state.

### 6.2 Cross-Platform Database Paths with Tilde Expansion

The OpenCode SQLite database location varies across operating systems. Configure `database.path` accordingly:
- **Linux**: `~/.local/share/opencode/opencode.db`
- **macOS**: `~/Library/Application Support/opencode/opencode.db`
- **Windows**: `%APPDATA%\opencode\opencode.db` (or `C:\Users\<USER>\AppData\Roaming\opencode\opencode.db`)

AgenticLens automatically performs tilde expansion (`~`) on Linux and macOS, resolving paths to the user home directory. If running in a containerized or CI environment, set the database path via command-line flag: `--db-path /path/to/opencode.db`.

### 6.3 Executable Permissions and macOS Quarantine Flags

When installing precompiled binaries manually on Unix systems (Linux and macOS):
- Ensure the executable bit is set: `chmod +x ./bin/agentic-lens`.
- If Gatekeeper blocks execution on macOS, clear quarantine flags:
  ```bash
  xattr -d com.apple.quarantine ./bin/agentic-lens
  ```
- Store the binary in a dedicated directory (such as `./bin/` or `.opencode/skills/agentic-lens-setup/bin/`) and ensure a protective `.gitignore` contains `*` and `!/.gitignore` to prevent tracking compiled binaries in git.

### 6.4 Git Upstream Base Branch Inference and Fallback Probing

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

### 6.5 Workspace Directory Path Resolution

AgenticLens resolves workspace directories relative to the current working directory or git repository root:
- Paths with spaces (e.g. `my workspaces/ticket 1`) are fully supported by wrapping arguments in quotes:
  ```bash
  ./bin/agentic-lens --workspaces "my workspaces"
  ```
- Always verify that the configured `workspaces.dir` in `.agenticlens-config.yaml` matches the directory used in CLI commands and agent prompts.

---

## Appendix: Platform Architecture and Core Concepts

AgenticLens is constructed around the Functional Core, Imperative Shell (FCIS) pattern and packaged as a single, self-contained binary executable with zero external runtime dependencies.

### A.1 Functional Core, Imperative Shell (FCIS) Architecture

The architecture strictly separates pure domain computations from side-effecting operations:
- **Functional Core (`backend/internal/domain/`)**: Contains pure value objects, cost models, statistical anomaly calculation formulas, and context window health algorithms. The functional core contains zero file I/O, zero network dependencies, and zero system clocks. All calculations are completely deterministic and easily verified via unit tests.
- **Imperative Shell (`backend/internal/adapters/` and `backend/cmd/agentic-lens/`)**: Manages external interactions including OpenCode SQLite database querying, git repository diff extraction, configuration file parsing, HTTP server lifecycle management, and CLI argument binding.

### A.2 Single Binary Zero-Dependency Model

Traditional developer tools often require complex runtime dependencies such as Node.js, Python runtimes, or external database servers. AgenticLens eliminates this operational overhead:
- The backend is written in Go 1.23+ and compiles to a native static executable with CGO disabled (`CGO_ENABLED=0`).
- SQLite database connectivity is powered by a pure Go driver (`modernc.org/sqlite`), eliminating the requirement for a host C compiler or shared C libraries.
- The web dashboard is built with TypeScript and Vite. During the build process, compiled static frontend assets (`web/dist/`) are embedded directly into the Go executable via `embed.FS`.
- Adopting repositories run a single executable file that serves both the command-line interface and the full-featured interactive web dashboard.

### A.3 Local Immutable Telemetry Fact Ledgers (`agentic_metrics.json`)

Rather than transmitting proprietary code or execution telemetry to external cloud services, AgenticLens operates entirely locally:
- Telemetry facts are captured per ticket or task in isolated workspace directories: `workspaces/<TICKET_ID>/agentic_metrics.json`.
- Each ledger is an immutable JSON document recording session identifiers, wall-clock duration, token volume (input, output, cache-read, and cache-write tokens), financial costs calculated from precise model pricing matrices, turn counts, tool invocation distributions, error rates, and git commit diff volumes (insertions, deletions, and affected files).
- These ledgers are committed directly to version control alongside code deliverables, establishing an auditable, repository-native history of agent engineering metrics with zero cloud egress.

### A.4 Real-Time Live Observability and SSE Event Streaming

In addition to post-turn historical ledgers, AgenticLens provides live observability of active agent sessions:
- The embedded server continuously polls the local OpenCode SQLite database at configurable intervals (`poll_interval_ms`, default 2000 ms).
- Active sessions are identified based on recent turn activity within a sliding time window (`active_threshold_seconds`, default 300 seconds).
- Telemetry events, tool calls, and status updates are broadcast to connected browsers via Server-Sent Events (SSE) through the `/api/live/stream` endpoint.
- An in-memory trace buffer retains recent execution events, providing immediate visibility into subagent handoffs, prompt dispatches, and tool outputs as they occur.

### A.5 Automated Anomaly Detection

AgenticLens continuously analyzes active sessions and historical ledgers against configurable statistical thresholds to detect engineering inefficiencies:
- **Session Stalls**: Sessions that remain open with no new turn activity beyond a configurable duration (`stall_threshold_seconds`, default 120 seconds).
- **High Tool Error Rates**: Sessions where the ratio of failed tool executions to total invocations exceeds a critical limit (`error_rate_threshold`, default 0.25 or 25%).
- **Context Window Exhaustion**: Sessions where cumulative prompt and generation tokens reach or exceed a critical percentage of model capacity (`context_exhaustion_pct`, default 0.85 or 85%).

### A.6 Context Window Health States

Context saturation severely impacts model reasoning quality, instruction adherence, and cost. AgenticLens classifies context health into four discrete states based on utilization percentage:
- **HEALTHY** (0% to 50% utilization): Optimal operating envelope. The model retains complete instruction adherence and context fidelity with minimal compaction pressure.
- **ELEVATED** (50% to 75% utilization): Normal working range for complex implementations. The agent operates effectively, but monitoring is recommended.
- **BLOATED** (75% to 85% utilization): High context saturation. Model reasoning degradation and prompt compaction risks increase significantly.
- **CRITICAL** (85% to 100% utilization): Severe exhaustion. Immediate session compaction or subagent delegation is advised to avoid catastrophic forgetting or dropped instructions.
