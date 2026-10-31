# AgenticLens Platform Setup Manual and Telemetry Integration Guide

AgenticLens is a developer telemetry and observability platform for OpenCode and Codex workflows. It provides real-time monitoring for OpenCode and offline telemetry ingestion from either harness.

OpenCode traces are extracted read-only from its local SQLite database (`~/.local/share/opencode/opencode.db`). The supported evidence window is OpenCode v1.19.5 (`session`/`message`/`part`) and npm `@opencode/cli` v2.2.0 (`session_v2`/`session_message`), whose `latest` tag was resolved on 2026-10-29. AgenticLens validates the complete required table and column set and automatically selects exactly one schema family. Partial, mixed, malformed, and unknown schemas fail with diagnostics; schema detection does not identify the exact runtime release. `--db-path` remains the explicit path override and the default OpenCode database path remains unchanged.

Compatibility workflows are `make opencode-compat-v1`, `make opencode-compat-v2`, and `make opencode-compat`. They require the corresponding runtime (`OPENCODE_V1_BIN` / `OPENCODE_V2_BIN` can select explicit binaries) and a local deterministic provider already running. The workflows create isolated temporary HOME, XDG data/config/state roots, project roots, and absolute `OPENCODE_DB` paths, then remove them at exit. Ordinary adapter tests use deterministic SQLite fixtures and need no OpenCode installation. Evidence fixtures are sanitized runtime-produced records, not synthesized runtime integration runs. Nonzero native-provider pricing, imported-history extraction after migration, partial migration states, duplicate handling, and releases outside v1.19.5/v2.2.0 are not verified.
Codex traces are read from rollout files beneath a Codex home directory (`~/.codex`). OpenCode and Codex are independent sources; selecting one never falls back to the other.

Use this guide to install, configure, and operate AgenticLens in an adopting repository.

---

## 1. Quick Start and Setup Pathways

AgenticLens offers two primary setup pathways: an autonomous agentic setup via a shared skill manifest, and a direct manual developer installation using official precompiled release archives. Target repositories never need to compile code from source or install Go, Node.js, npm, or C compiler toolchains.

### 1.1 Pathway A: Agentic Distribution and Setup via SKILL.md

#### Step 1: Share Skill Reference with Your Agent
Provide your autonomous coding agent in OpenCode with the reference to the AgenticLens setup skill. You can either:
- Point the agent directly to the canonical remote URL:
  `https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/.agents/skills/agentic-lens-setup/SKILL.md`
- Or copy `.agents/skills/agentic-lens-setup/SKILL.md` into your repository.

Prompt your agent:
```
Read the instructions at https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/.agents/skills/agentic-lens-setup/SKILL.md and set up AgenticLens in this repository.
```

#### Step 2: Install the Wrapper and Native Binary

Follow [manual installation](#12-pathway-b-manual-developer-installation-via-github-releases), using `.agents/skills/agentic-lens-setup/bin` as the destination. The POSIX wrapper detects Linux, macOS, or Windows shell environments and amd64/arm64, selects `agentic-lens-${OS}-${ARCH}` (`.exe` on Windows), and downloads missing platform binaries. See [container/host operation](#66-container-sandbox-vs-host-machine-execution-multi-platform-wrapper).

If the setup downloader is available:

```bash
./.agents/skills/agentic-lens-setup/scripts/download.sh
# Optional: pre-download the host binary from a Linux container.
./.agents/skills/agentic-lens-setup/scripts/download.sh --include-host darwin-arm64
```

#### Step 3: Verify Binary Execution
The agent validates that the executable operates cleanly on the host system:
```bash
./.agents/skills/agentic-lens-setup/bin/agentic-lens version --json
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

The agent runs `init` to write shared `.agenticlens-config.yaml` and machine-local `.agenticlens-config.local.yaml`:
```bash
./.agents/skills/agentic-lens-setup/bin/agentic-lens init \
  --project-name "BillingService" \
  --ticket-prefixes "BILL,PAY,GEN" \
  --src-folders "services,pkg,cmd" \
  --workspaces "workspaces"
```

#### Step 5: Mandatory User Confirmation Gate
Before updating or creating any OpenCode agent policy files (`AGENTS.md` or `.agents/agent/*.md`), the agent must present the exact policy text to the user and receive explicit approval.

---

### 1.2 Pathway B: Manual Developer Installation via GitHub Releases

Developers who prefer direct manual installation without an agent setup skill can install AgenticLens in three quick steps.

#### Step 1: Install Wrapper and Platform Binary
Navigate to official GitHub Releases:
`https://github.com/NiltonDuarte/AgenticLens/releases`

Standalone precompiled binaries are distributed directly with embedded frontend assets and runtime dependencies. No archive extraction tools (`tar`, `unzip`) or compilation toolchains are required.

Download via `curl`:
```bash
mkdir -p ./bin

# Write protective .gitignore
cat <<'EOF' > ./bin/.gitignore
*
!.gitignore
EOF

# Install multi-platform wrapper script
curl -fsSL -o ./bin/agentic-lens "https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/scripts/agentic-lens-wrapper.sh"
chmod +x ./bin/agentic-lens

# Example for Linux amd64
curl -fsSL -o ./bin/agentic-lens-linux-amd64 "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/agentic-lens-linux-amd64"
chmod +x ./bin/agentic-lens-linux-amd64

# Example for macOS Apple Silicon
curl -fsSL -o ./bin/agentic-lens-darwin-arm64 "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/agentic-lens-darwin-arm64"
chmod +x ./bin/agentic-lens-darwin-arm64
```

Alternatively, use the automated downloader script:
```bash
curl -fsSL https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/scripts/download-agentic-lens.sh | bash -s -- --dest ./bin
```

On Windows (PowerShell):
```powershell
New-Item -ItemType Directory -Force -Path .\bin
Invoke-WebRequest -Uri "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/agentic-lens-windows-amd64.exe" -OutFile ".\bin\agentic-lens-windows-amd64.exe"
Copy-Item ".\bin\agentic-lens-windows-amd64.exe" ".\bin\agentic-lens.exe"
```

#### Step 2: Scaffold Configuration and Launch Dashboard
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

Live SSE uses OpenCode; offline generation supports OpenCode and Codex. Sources are independent and never silently substituted after a failure.

### 2.1 Complete Methodology Neutrality within OpenCode

Adopters choose their own single-agent, human-assisted, hierarchical, or CI workflows, review gates, and personas. They do not inherit this repository's maintainer skills (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`).

### 2.2 Custom Conventions and Taxonomy

Configure ticket prefixes as equal peers, any workspace root, and custom persona mappings. Trunk-based, GitFlow, GitHub flow, and short-lived branches are supported; unmapped personas can remain custom roles.

## 3. Configuration Reference

Shared project settings live in `.agenticlens-config.yaml`. Machine-specific source settings live in optional `.agenticlens-config.local.yaml`, which is ignored by Git. Review its default harness and paths, verify it is ignored before committing, and never put credentials in either file. The sanitized `.agenticlens-config.local.example.yaml` can be copied as a starting point.

```yaml
telemetry:
  default_source: opencode
  sources:
    opencode:
      database_path: ~/.local/share/opencode/opencode.db
    codex:
      home: ~/.codex
```

### Telemetry Source and Session Selection

`agentic-lens init` creates shared `.agenticlens-config.yaml` and machine-local `.agenticlens-config.local.yaml`. The local file is ignored by Git; check `.gitignore` before committing and keep credentials out of both files. Configure `telemetry.default_source`, `telemetry.sources.opencode.database_path` (OpenCode SQLite), and `telemetry.sources.codex.home` (Codex rollout directory). The tracked `.agenticlens-config.local.example.yaml` provides sanitized portable defaults. `generate --source` overrides the local default; otherwise `telemetry.default_source` is used, then OpenCode. Source locations resolve from CLI path override, local harness path, legacy shared OpenCode `database.path`, then adapter defaults. A missing selected source is an error and never triggers a switch to the other harness.

Use `--nonce <literal>` to select the exact Codex rollout containing a unique run nonce, or `--session-id <thread-id>` to select a known thread. `--codex-home <path>` overrides local Codex home, `$CODEX_HOME`, and `~/.codex`.

Code metrics always compare an explicit Git ref with the configured `code_metrics.base_branch`; `generate` defaults to `HEAD`, and `backfill --recompute-git` requires `--code-ref`. Empty or invalid refs fail instead of trying to infer a branch or commits from the ticket ID. Comparisons include committed changes and exclude the uncommitted working tree. Codex session discovery requires `--nonce` or `--session-id`; ticket text and working-directory heuristics are not reliable session identifiers. Codex SDLC iterations count root-session user messages, not folders, commits, model requests, or child-agent prompts. Schema-v4 sessions own invocation records, per-session tool summaries, and retained events; cross-session totals are derived for display. Migration reports and removes legacy aggregate tool rows without reliable session attribution. Re-extract from source logs to recover those counts when available.

### Code Classification and Interpretation

Code classification is configured under `code_metrics` in `.agenticlens-config.yaml`:
`test_patterns` identifies tests and fixtures; `prod_patterns` and `include` identify production files. There are no implicit test directory, test filename, or production extension rules. Empty patterns leave files unclassified (`Other`), which can contribute to total lines without contributing to test/production subtotals.

Patterns containing `/` are relative to the repository root; a `**` path component matches zero or more directories. Patterns without `/` match basenames. Workspace exclusions and `exclude` win, followed by test patterns, then production/include patterns. The existing lockfile/dotfile exclusions apply to otherwise unmatched files. Both Git discovery modes monitor the required `code_metrics.source_roots` plus configured test, production, and include patterns, allowing fixtures outside source roots.

Interpret ratios using the configured classification and counting basis. Test helpers can be counted as production if no test pattern matches them; generated files, fixtures, comments, and blank lines also affect comparisons. Whole-repository source counts and session diff-added counts answer different questions. There is no universal target ratio; use the [test metrics guidance](../TECHNICAL_GUIDELINES.md#interpreting-test-metrics) and [contributor test review steps](../CONTRIBUTING.md#51-before-adding-or-removing-tests) to assess coverage and maintenance cost.

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
  source_roots:
    - src
    - pkg
    - cmd
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
- **`source_roots`** (List of Strings, Required): Source code folders monitored for Git changes and code metrics (e.g. `src`, `pkg`, `cmd`).
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

Complete implementation, testing, and review before generating telemetry. Generate immediately before the final commit, then stage the task-owned deliverables and metrics together. Earlier extraction omits later turns, tool calls, compactions, and subagent handoffs.

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
Generate a unique 8–16 character alphanumeric or hex nonce and pass it literally. Never use `$NONCE` or `$(...)`: OpenCode stores tool inputs before shell expansion, so searching for the expanded value cannot find that input. See [CLI flags](#agentic-lens-generate) for output and source options.

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

Configure `role_mappings` as shown in the [configuration example](#31-complete-configuration-schema-example). Tokens, costs, and durations aggregate under mapped roles while raw ledgers preserve persona identifiers.

### 4.5 Standard Repository Agent Guidelines Policy Block

Below is the standard policy block that can be added to your repository's OpenCode agent instructions file (`AGENTS.md` or `.agents/agent/*.md`):

```markdown
## AgenticLens Telemetry Guidelines

1. Ticket Tracking Mandate:
   - Every task requires a ticket ID and a workspace under `workspaces/<TICKET_ID>/`.
   - Use an existing ticket when one is provided. For ad hoc work without an ID, assign a `GEN-#` identifier; do not create a planning specification or index entry under `plan/` solely for it.
   - Agents work on one ticket at a time. All turns belong to the active ticket.

2. Plan Lifecycle:
   - When a matching specification exists under `plan/to-do/`, move it to `plan/done/` when its ticket is completed.
   - Remove the ticket entry from `plan/TICKETS.md` if present. Do not create a specification or index entry just to complete this step; ad hoc tickets still have an ID and workspace.

3. Dedicated Workspace Allocation:
   - Store task documentation and metrics in `workspaces/<TICKET_ID>/`.
   - Maintain `CONTRACT.md` (scope boundaries) and `SUMMARY.md` (execution log) as required by the project's artifact list.

4. Delayed Telemetry Ingestion:
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
Scaffolds `.agenticlens-config.yaml` and `.agenticlens-config.local.yaml` at the repository root. The local file contains only source settings and is Git-ignored.
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
  - `--session-id <ID>` (Optional): Explicit source session ID override.
  - `--db-path <PATH>` (Optional): OpenCode SQLite path override; see [source selection](#telemetry-source-and-session-selection) for Codex and harness options.
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

#### `agentic-lens update`
Upgrades installed platform binaries to the latest (or specified) official release from GitHub Releases.
- Flags:
  - `--version <tag>` (Optional, Default: `latest`): Target release tag to install (e.g. `v1.0.1`).
- Notes:
  - Executed directly by the multi-platform wrapper script (`bin/agentic-lens`).
  - Synchronously refreshes all installed platform binaries (`agentic-lens-*`) in the binary directory and updates the version sentinel.

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
- Store the binary in a dedicated directory (such as `./bin/` or `.agents/skills/agentic-lens-setup/bin/`) and ensure a protective `.gitignore` contains `*` and `!/.gitignore` to prevent tracking compiled binaries in git.

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

### 6.6 Container Sandbox vs Host Machine Execution (Multi-Platform Wrapper)

A very common configuration in OpenCode development is running the autonomous agent inside a Linux sandbox (such as Docker, Dev Containers, or a remote container) while the engineer interacts from a macOS host workstation:
- **The Problem**: If setup installs a single platform executable (e.g. Linux ELF binary) to `./bin/agentic-lens`, running `./bin/agentic-lens server` on macOS fails with binary execution errors (`cannot execute binary file`). Conversely, installing only a macOS binary prevents the containerized agent from executing `./bin/agentic-lens generate` or `./bin/agentic-lens version`.
- **The Solution**: AgenticLens resolves this by installing a lightweight dispatcher wrapper script at `bin/agentic-lens` alongside native platform executables (`agentic-lens-${OS}-${ARCH}`).
- **Self-Healing Download**: If an engineer runs `./bin/agentic-lens server` on their macOS workstation after the setup skill ran inside the Linux container, the wrapper detects that `agentic-lens-darwin-arm64` is missing locally and downloads it automatically from GitHub Releases before executing.
- **Background Update Notification**: The wrapper checks for newer versions in a background subshell (throttled to at most once per 24 hours) without blocking CLI commands. If an update is detected, a notice is displayed on stderr. The check can be disabled by setting `AGENTIC_LENS_NO_UPDATE_CHECK=1`.
- **Upgrading**: Run `./bin/agentic-lens update` to upgrade all installed platform binaries to the latest release tag.

---

## Appendix: Platform Architecture and Core Concepts

### A.1 Functional Core, Imperative Shell (FCIS) Architecture

See [architecture and subsystem responsibilities](../ARCHITECTURE.md) and the mandatory [technical boundaries](../TECHNICAL_GUIDELINES.md#2-functional-core-imperative-shell-fcis).

### A.2 Single Binary Zero-Dependency Model

Go 1.23+ builds with `CGO_ENABLED=0`; `modernc.org/sqlite` needs no C compiler or shared C libraries. Vite compiles the TypeScript dashboard into `web/dist`, embedded via `embed.FS`. The resulting executable provides both CLI and web UI without external runtimes.

### A.3 Local Immutable Telemetry Fact Ledgers (`agentic_metrics.json`)

Rather than transmitting proprietary code or execution telemetry to external cloud services, AgenticLens operates entirely locally:
- Telemetry facts are captured per ticket or task in isolated workspace directories: `workspaces/<TICKET_ID>/agentic_metrics.json`.
- Schema-v4 sessions own invocation records, per-session tool summaries, session-specific anomaly summaries, and sparse direct-evidence events. Cross-session anomaly summaries (including high tool failure rates) live at ticket level without session IDs or event rows. Metrics are derived from summaries and tool totals, not by rerunning sequence detectors over sparse evidence. When OpenCode hydration is unavailable, the timeline shows committed anomaly evidence only; full Codex event hydration is not provided. Migration reports and removes legacy aggregate tool rows without reliable session attribution; re-extract from OpenCode/Codex source logs to recover those counts when available.
- Each ledger is an immutable JSON document recording session identifiers, wall-clock duration, token volume (input, output, cache-read, and cache-write tokens), financial costs calculated from precise model pricing matrices, turn counts, tool invocation distributions, error rates, and git commit diff volumes (insertions, deletions, and affected files).
- These ledgers are committed directly to version control alongside code deliverables, establishing an auditable, repository-native history of agent engineering metrics with zero cloud egress.

#### Recovering unscoped tool totals

Preserve the old ledger first. Run `generate` for each known root session ID with `--json-output` pointing to a new, initially nonexistent file. Reuse that same new output path for later root sessions; generation merges the session facts by session and tool name. For Codex, selecting a root thread also discovers its child rollouts. For OpenCode, generate once for each root session. Do not point `--json-output` at the old ledger: its unscoped tool rows fail validation before generation can merge new facts.

```bash
cp workspaces/<TICKET_ID>/agentic_metrics.json workspaces/<TICKET_ID>/agentic_metrics.unscoped-backup.json
./bin/agentic-lens generate --source codex --session-id <ROOT_THREAD_ID_1> --ticket-id <TICKET_ID> --ticket-title "<TICKET_TITLE>" --ticket-type chore --json-output workspaces/<TICKET_ID>/agentic_metrics.recovered.json
./bin/agentic-lens generate --source codex --session-id <ROOT_THREAD_ID_2> --ticket-id <TICKET_ID> --ticket-title "<TICKET_TITLE>" --ticket-type chore --json-output workspaces/<TICKET_ID>/agentic_metrics.recovered.json

# Compare the IDs and per-session tool rows with the sessions you intended to recover.
jq -r '.sessions[].id' workspaces/<TICKET_ID>/agentic_metrics.recovered.json | sort -u
jq '[.sessions[].tools[]] | length' workspaces/<TICKET_ID>/agentic_metrics.recovered.json

# Replace the active ledger only after verifying all expected sessions and counts.
mv workspaces/<TICKET_ID>/agentic_metrics.recovered.json workspaces/<TICKET_ID>/agentic_metrics.json
```

Use `--source opencode` and repeat with each OpenCode root session ID for OpenCode recovery. Verify all expected session IDs (including Codex child rollouts) and tool counts before replacing the original. An unscoped aggregate cannot be accurately divided among sessions; if the source logs are unavailable, those old aggregate counts cannot be recovered. Keep the backup and do not invent IDs or redistribute the totals.

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
