---
name: agentic-lens-setup
description: Guided interactive setup skill to bootstrap AgenticLens in any repository, downloading precompiled release binaries, configuring project parameters, explaining telemetry features, and reviewing agent guidelines with mandatory user confirmation.
---

# AgenticLens Cross-Project Interactive Setup Skill

Guides autonomous bootstrapping and adoption of AgenticLens in any software repository.

## Operational Mandates
- Strict Release-Only Mandate: Never compile binaries from source or request Go/Node.js toolchain installation in target repositories. Retrieve official precompiled release artifacts from GitHub Releases.
- Skill Co-Location Mandate: Download and store the precompiled binary inside this skill directory (`.opencode/skills/agentic-lens-setup/bin/agentic-lens`) alongside a protective `.gitignore` preventing accidental git tracking.
- Mandatory User Confirmation Gate: Inspect existing agent instructions (`AGENTS.md`, `.opencode/agent/*.md`, `CLAUDE.md`) and propose operational policies. Never modify target repository files without presenting the exact text and receiving affirmative user confirmation.
- Zero Emojis Mandate: Zero emojis permitted across all prompts, CLI output, generated files, configurations, and documentation.

---

## Step-by-Step Setup Workflow

### Step 1: Precompiled Release Retrieval

Execute the co-located artifact downloader script to detect the host operating system (Linux, macOS, Windows) and architecture (amd64, arm64), fetch the precompiled release binary, unpack it, and enforce executable permissions:

```bash
./.opencode/skills/agentic-lens-setup/scripts/download.sh
```

If the downloader script is invoked from the repository root, `scripts/download-agentic-lens.sh` may also be used:

```bash
./scripts/download-agentic-lens.sh
```

#### Verification
Confirm the binary is operational and inspect the build metadata:

```bash
./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json
```

Verify that valid JSON is returned containing:
- `version`: semantic release version string
- `commit`: git SHA hash of the build
- `build_date`: build timestamp in RFC3339 format

---

### Step 2: Interactive Interview and Configuration Scaffolding

Prompt the developer or inspect the repository to collect configuration parameters:

1. Project Name:
   - Descriptive identifier for the repository.
   - Default: current root directory name.
2. Primary Ticket Prefix and Classification:
   - Primary prefix for feature tickets (e.g. AL, PROJ, APP).
   - Standard secondary prefixes: `GEN` (chore) and `PO` (docs).
3. Workspaces Directory:
   - Directory where ticket artifacts and session telemetry ledgers are stored.
   - Default: `workspaces`.
4. Monitored Source Folders:
   - Comma-separated list of source code directories to observe for git metrics and diff volume (e.g. `src`, `backend`, `web`, `lib`, `pkg`).
5. OpenCode SQLite Database Path:
   - Path to local OpenCode database.
   - Default: `~/.local/share/opencode/opencode.db`.
6. Agent Role Mappings:
   - Personas mapped to canonical roles (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`).

#### Generating Configuration
Run the CLI `init` command using the collected parameters:

```bash
./.opencode/skills/agentic-lens-setup/bin/agentic-lens init \
  --project-name "<PROJECT_NAME>" \
  --ticket-prefix "<TICKET_PREFIX>" \
  --workspaces "<WORKSPACES_DIR>" \
  --src-folders "<SRC_FOLDERS>" \
  --db-path "<DB_PATH>"
```

Alternatively, review generated YAML with `--dry-run` before writing:

```bash
./.opencode/skills/agentic-lens-setup/bin/agentic-lens init --dry-run
```

This writes a fully commented, valid `.agenticlens-config.yaml` file at the repository root.

---

### Step 3: Platform Capability Walkthrough

Present an overview of AgenticLens features to the user:

#### CLI Subcommands
- `init`: Scaffold and generate `.agenticlens-config.yaml` configuration.
- `generate`: Ingest session telemetry from OpenCode SQLite DB and git history into `agentic_metrics.json`, and idempotently update markdown summary tables.
  ```bash
  NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
  ./.opencode/skills/agentic-lens-setup/bin/agentic-lens generate \
    --nonce $NONCE \
    --ticket-id <TICKET_ID> \
    --ticket-type <CONVENTIONAL_TYPE> \
    --json-output workspaces/<TICKET_ID>/agentic_metrics.json \
    --append workspaces/<TICKET_ID>/SUMMARY.md
  ```
- `scan`: Scan workspace ledgers, reporting active sessions, iteration counts, context window health, and token costs:
  ```bash
  ./.opencode/skills/agentic-lens-setup/bin/agentic-lens --list
  ```
- `server`: Start embedded web dashboard and real-time streaming API service:
  ```bash
  ./.opencode/skills/agentic-lens-setup/bin/agentic-lens server --port 8080
  ```
- `backfill`: Batch telemetry backfill, git diff fact recomputation, and schema upgrades.
- `migrate`: Stepwise schema migration engine across ledger versions.
- `version`: Print binary version, git commit, build date, and schema information.

#### Real-Time Observability and Live Monitoring
- Automatic SQLite polling discovers active agent sessions in real time.
- Automated anomaly detection identifies:
  - Stalls (sessions exceeding turn inactivity threshold).
  - High Error Rate (ratio of tool invocation failures exceeding threshold).
  - Context Exhaustion (token utilization approaching model context window capacity).

#### Embedded Web Dashboard
- Single binary distribution with embedded TypeScript web application.
- Visualizations for token consumption, model breakdown, turn velocity, SDLC iteration loops, and cost tracking.

---

### Step 4: Agent Guidelines Review and Mandatory Confirmation Gate

Inspect the repository for existing agent instructions:
- `AGENTS.md`
- `.opencode/agent/*.md`
- `CLAUDE.md`

#### Mandatory User Confirmation Gate (HARD STOP)
Before making any modifications or creating new files, formulate the exact proposed markdown section and present it to the user. Ask for explicit approval using the `question` tool or user prompt.

If the user declines or requests adjustments, honor the user's decision without modifying files.

#### Proposed Agent Policy Block
When approved by the user, append the following standardized operational policies to the repository's agent instructions:

```markdown
## AgenticLens Operational Policies

### 1. Ticket Tracking Mandate (Zero Untracked Work)
- Every task, feature, bugfix, or refactoring must be tracked by an explicit ticket identifier matching configured prefixes (e.g. `<PREFIX>-<ID>`, `GEN-<ID>`, `PO-<ID>`).
- Single-ticket exclusivity: Agents must never work on more than one ticket simultaneously.
- Every subsequent prompt must be treated as steering feedback for the active ticket.

### 2. Dedicated Workspace Allocation
- Every ticket must have an isolated workspace under `workspaces/<TICKET_ID>/`.
- Standard artifacts:
  - `workspaces/<TICKET_ID>/CONTRACT.md`: Frozen technical contract and scope invariants.
  - `workspaces/<TICKET_ID>/SUMMARY.md`: Summary of completed work and verification metrics table.
  - `workspaces/<TICKET_ID>/PULL_REQUEST.md`: Pull request description.
  - `workspaces/<TICKET_ID>/agentic_metrics.json`: Telemetry fact ledger.

### 3. Delayed Telemetry Ingestion Mandate
- Telemetry generation must be executed as the very final turn before git commit:
  ```bash
  NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
  ./.opencode/skills/agentic-lens-setup/bin/agentic-lens generate \
    --nonce $NONCE \
    --ticket-id <TICKET_ID> \
    --ticket-type <CONVENTIONAL_TYPE> \
    --json-output workspaces/<TICKET_ID>/agentic_metrics.json \
    --append workspaces/<TICKET_ID>/SUMMARY.md
  ```
- Premature ingestion omits subsequent tool calls, turn counts, compactions, and subagent sessions.
```

---

## Completion Verification Checklist
- Binary exists at `.opencode/skills/agentic-lens-setup/bin/agentic-lens` and is executable.
- Protective `.gitignore` exists at `.opencode/skills/agentic-lens-setup/bin/.gitignore`.
- Root `.agenticlens-config.yaml` is present, valid YAML, and non-empty.
- Verification command `./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json` succeeds.
- Agent guidelines updated only after explicit user confirmation.
- Zero emojis in all artifacts and outputs.
