---
name: agentic-lens-setup
description: Autonomous setup skill to bootstrap AgenticLens in any repository, downloading precompiled release binaries, configuring project parameters, explaining telemetry features, and reviewing agent guidelines with mandatory user confirmation.
---

# AgenticLens Cross-Project Autonomous Setup Skill

Guides autonomous bootstrapping and adoption of AgenticLens in any software repository using precompiled release artifacts from GitHub Releases (https://github.com/NiltonDuarte/AgenticLens/releases).

This skill follows the autonomous agentic distribution model: it is completely self-contained and executable by any autonomous coding agent reading this manifest directly via shared repository URL or local path.

## Operational Mandates
- Strict Release-Only Mandate: Never compile binaries from Go/TypeScript source or request Go/Node.js toolchain installation in target repositories. Retrieve official precompiled release artifacts from GitHub Releases: https://github.com/NiltonDuarte/AgenticLens/releases.
- Binary Placement and Git Protection: Store the precompiled binary inside `.opencode/skills/agentic-lens-setup/bin/agentic-lens` (or `./bin/agentic-lens`) alongside a protective `.gitignore` containing `*` and `!/.gitignore` to prevent committing compiled binaries to version control.
- Mandatory User Confirmation Gate: Inspect existing agent instructions (`AGENTS.md`, `.opencode/agent/*.md`, `CLAUDE.md`) and propose operational policies. Never modify target repository files without presenting the exact text and receiving affirmative user confirmation.
- Zero Emojis Mandate: Zero emojis permitted across all prompts, CLI output, generated files, configurations, and documentation.

---

## Step-by-Step Setup Workflow

### Step 1: Precompiled Release Retrieval and Verification

An autonomous agent executing this skill must perform host platform detection, download the appropriate official release archive, unpack the executable, and verify its integrity.

#### 1.1 Host Platform and Architecture Detection
Run the following shell commands to determine the target OS and CPU architecture:

```bash
OS_RAW=$(uname -s)
ARCH_RAW=$(uname -m)

case "$OS_RAW" in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*|Windows_NT) OS="windows" ;;
  *) echo "Error: Unsupported OS $OS_RAW" >&2; exit 1 ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Error: Unsupported architecture $ARCH_RAW" >&2; exit 1 ;;
esac
```

#### 1.2 Binary Destination and Protective Git Exclusion
Create the binary directory and ensure a protective `.gitignore` is in place:

```bash
BIN_DIR=".opencode/skills/agentic-lens-setup/bin"
mkdir -p "$BIN_DIR"

if [ ! -f "$BIN_DIR/.gitignore" ]; then
  cat <<'EOF' > "$BIN_DIR/.gitignore"
# Prevent accidental git tracking of precompiled binaries
*
!.gitignore
EOF
fi
```

#### 1.3 Binary Download and Extraction
If a local downloader script exists at `./.opencode/skills/agentic-lens-setup/scripts/download.sh`, execute it:
```bash
./.opencode/skills/agentic-lens-setup/scripts/download.sh
```

Otherwise, download directly from GitHub Releases using `curl` or `wget`:

For Linux or macOS:
```bash
ARCHIVE_NAME="agentic-lens-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/${ARCHIVE_NAME}"

echo "Downloading ${DOWNLOAD_URL}..."
curl -fsSL "$DOWNLOAD_URL" | tar -xz -C "$BIN_DIR"
chmod +x "$BIN_DIR/agentic-lens"
```

For Windows (PowerShell):
```powershell
$archiveName = "agentic-lens-windows-amd64.zip"
$downloadUrl = "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/$archiveName"
Invoke-WebRequest -Uri $downloadUrl -OutFile "$BIN_DIR\$archiveName"
Expand-Archive -Path "$BIN_DIR\$archiveName" -DestinationPath "$BIN_DIR" -Force
Remove-Item "$BIN_DIR\$archiveName"
```

#### 1.4 Verification
Validate that the downloaded binary executes properly and outputs valid build metadata:

```bash
"$BIN_DIR/agentic-lens" version --json
```

Verify that valid JSON is returned containing:
- `version`: semantic release version string
- `commit`: git SHA hash of the build
- `build_date`: build timestamp in RFC3339 format

---

### Step 2: Interactive Interview and Configuration Scaffolding

Inspect the repository and collect configuration parameters from the developer or workspace environment:

1. Project Name:
   - Descriptive identifier for the repository.
   - Strictly mandatory flag (`--project-name`); no default is provided.
2. Ticket Prefixes:
   - Comma-separated list of recognized ticket prefixes (e.g. `PROJ,GEN,PO` or `PAY,SEC,OPS`). All configured prefixes are treated as equal peers.
   - Strictly mandatory flag (`--ticket-prefixes`); no default or synthetic fallback is provided.
3. Workspaces Directory:
   - Directory where ticket artifacts and session telemetry ledgers are stored.
   - Default: `workspaces`.
4. Monitored Source Folders:
   - Comma-separated list of source code directories to observe for git metrics and diff volume (e.g. `src,backend,web,lib,pkg`).
   - Strictly mandatory flag (`--src-folders`); no default is provided.
5. OpenCode SQLite Database Path:
   - Path to local OpenCode database.
   - Default: `~/.local/share/opencode/opencode.db`.

#### Generating Configuration
Run the CLI `init` command using the collected parameters:

```bash
"$BIN_DIR/agentic-lens" init \
  --project-name "<PROJECT_NAME>" \
  --ticket-prefixes "<TICKET_PREFIXES>" \
  --workspaces "<WORKSPACES_DIR>" \
  --src-folders "<SRC_FOLDERS>" \
  --db-path "<DB_PATH>"
```

Alternatively, review generated YAML with `--dry-run` before writing:

```bash
"$BIN_DIR/agentic-lens" init \
  --project-name "<PROJECT_NAME>" \
  --ticket-prefixes "<TICKET_PREFIXES>" \
  --src-folders "<SRC_FOLDERS>" \
  --dry-run
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
  "$BIN_DIR/agentic-lens" generate \
    --nonce "$NONCE" \
    --ticket-id <TICKET_ID> \
    --ticket-type <CONVENTIONAL_TYPE> \
    --json-output workspaces/<TICKET_ID>/agentic_metrics.json \
    --append workspaces/<TICKET_ID>/SUMMARY.md
  ```
- `scan`: Scan workspace ledgers, reporting active sessions, iteration counts, context window health, and token costs:
  ```bash
  "$BIN_DIR/agentic-lens" --list
  ```
- `server`: Start embedded web dashboard and real-time streaming API service:
  ```bash
  "$BIN_DIR/agentic-lens" server --port 8080
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
- Every task, feature, bugfix, or refactoring must be tracked by an explicit ticket identifier matching configured prefixes (e.g. `<PREFIX>-<ID>`, `MISC-<ID>`, `APP-<ID>`).
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
  ./bin/agentic-lens generate \
    --nonce "$NONCE" \
    --ticket-id <TICKET_ID> \
    --ticket-type <CONVENTIONAL_TYPE> \
    --json-output workspaces/<TICKET_ID>/agentic_metrics.json \
    --append workspaces/<TICKET_ID>/SUMMARY.md
  ```
- Premature ingestion omits subsequent tool calls, turn counts, compactions, and subagent sessions.
```

---

## Completion Verification Checklist
- Binary exists at `.opencode/skills/agentic-lens-setup/bin/agentic-lens` (or `./bin/agentic-lens`) and is executable.
- Protective `.gitignore` exists at binary destination directory.
- Root `.agenticlens-config.yaml` is present, valid YAML, and non-empty.
- Verification command reports valid JSON metadata.
- Agent guidelines updated only after explicit user confirmation.
- Zero emojis in all artifacts and outputs.
