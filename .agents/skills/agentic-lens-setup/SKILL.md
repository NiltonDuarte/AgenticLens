---
name: agentic-lens-setup
description: Autonomous setup skill to bootstrap AgenticLens in Codex and OpenCode repositories, downloading precompiled release binaries, configuring project parameters, explaining telemetry features, and reviewing agent guidelines with mandatory user confirmation.
---

# AgenticLens Cross-Project Autonomous Setup Skill

Guides autonomous bootstrapping and adoption of AgenticLens in Codex and OpenCode software repositories using precompiled release artifacts from GitHub Releases (https://github.com/NiltonDuarte/AgenticLens/releases).

AgenticLens is an observability and developer telemetry platform for OpenCode and Codex workflows. OpenCode traces come from its local SQLite database; Codex traces come from rollout files under the Codex home directory.

This skill follows the autonomous agentic distribution model: it is completely self-contained and executable by any autonomous coding agent reading this manifest directly via shared repository URL or local path.

## Operational Mandates
- Strict Release-Only Mandate: Never compile binaries from Go/TypeScript source or request Go/Node.js toolchain installation in target repositories. Retrieve official precompiled release artifacts from GitHub Releases: https://github.com/NiltonDuarte/AgenticLens/releases.
- Operator Skill and Binary Placement: Install the AgenticLens operator skill under `.agents/skills/agentic-lens-operator/` and place the multi-platform wrapper and precompiled platform binaries in its `bin/` subdirectory. Keep a protective `.gitignore` in that binary directory containing `*` and `!/.gitignore`.
- Mandatory User Confirmation Gate: Inspect existing agent instructions and hook/plugin configuration. Present the exact additive setup changes before modifying target repository files; affirmative approval already covering those changes satisfies the gate. Never modify third-party skills or personas to obtain telemetry.
- Lifecycle Capture: Use the native release binary to install AgenticLens-owned callbacks, request explicit session-specific ticket binding from the agent and reconcile authoritative source records. Never infer ownership, parents, model, effort or usage from checkout proximity or an agent assertion. Preserve unrelated hooks, trust settings and conflicting local edits. Installation alone does not prove runtime capture.
- Local State: Keep capture registry, pending snapshots, source paths and request tokens local and ignored. Only final generation publishes the ticket ledger, summary and digest-bound coverage report.
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

#### 1.2 Operator Skill and Binary Destination
The operator skill is the day-to-day interface for using the installed platform. Before writing files, present the exact operator skill files to be installed (the entrypoint and each reference), their source URLs, and the destination `.agents/skills/agentic-lens-operator/`; include the binary destination under its `bin/` directory. Obtain affirmative authorization for these repository changes. Do not write the skill files or binaries before approval. If an operator skill already exists, inspect it and present the exact diff; preserve local edits and ask before replacing or updating conflicting files.

Install the operator skill files from the public AgenticLens repository:

```bash
OPERATOR_SKILL_DIR=".agents/skills/agentic-lens-operator"
BIN_DIR="$OPERATOR_SKILL_DIR/bin"
RAW_SKILL_URL="https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/.agents/skills/agentic-lens-operator"
mkdir -p "$OPERATOR_SKILL_DIR/references" "$BIN_DIR"

curl -fsSL -o "$OPERATOR_SKILL_DIR/SKILL.md" "$RAW_SKILL_URL/SKILL.md"
curl -fsSL -o "$OPERATOR_SKILL_DIR/references/cli.md" "$RAW_SKILL_URL/references/cli.md"
curl -fsSL -o "$OPERATOR_SKILL_DIR/references/dashboard.md" "$RAW_SKILL_URL/references/dashboard.md"
curl -fsSL -o "$OPERATOR_SKILL_DIR/references/configuration.md" "$RAW_SKILL_URL/references/configuration.md"

if [ ! -f "$BIN_DIR/.gitignore" ]; then
  cat <<'EOF' > "$BIN_DIR/.gitignore"
# Prevent accidental git tracking of precompiled binaries
*
!.gitignore
EOF
fi
```

For Windows (PowerShell), fetch the same files before installing the binary:

```powershell
$OPERATOR_SKILL_DIR = ".agents/skills/agentic-lens-operator"
$BIN_DIR = "$OPERATOR_SKILL_DIR/bin"
$RAW_SKILL_URL = "https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/.agents/skills/agentic-lens-operator"
New-Item -ItemType Directory -Force -Path "$OPERATOR_SKILL_DIR/references", $BIN_DIR | Out-Null
Invoke-WebRequest -Uri "$RAW_SKILL_URL/SKILL.md" -OutFile "$OPERATOR_SKILL_DIR/SKILL.md"
Invoke-WebRequest -Uri "$RAW_SKILL_URL/references/cli.md" -OutFile "$OPERATOR_SKILL_DIR/references/cli.md"
Invoke-WebRequest -Uri "$RAW_SKILL_URL/references/dashboard.md" -OutFile "$OPERATOR_SKILL_DIR/references/dashboard.md"
Invoke-WebRequest -Uri "$RAW_SKILL_URL/references/configuration.md" -OutFile "$OPERATOR_SKILL_DIR/references/configuration.md"
if (-not (Test-Path "$BIN_DIR/.gitignore")) {
  Set-Content -Path "$BIN_DIR/.gitignore" -Value "*`n!.gitignore`n" -NoNewline
}
```

#### 1.3 Multi-Platform Wrapper and Binary Installation
Autonomous agents frequently execute inside containerized Linux sandboxes (such as devcontainers or Docker) while developers interact from macOS or Windows host workstations. Install the portable wrapper at `.agents/skills/agentic-lens-operator/bin/agentic-lens`; it dispatches to native platform executables (`agentic-lens-${OS}-${ARCH}`) in the same directory and can download a missing host binary from GitHub Releases on demand.

If a local downloader script exists at `./.agents/skills/agentic-lens-setup/scripts/download.sh`, execute it:
```bash
# Basic setup for current execution environment (e.g. Linux container)
./.agents/skills/agentic-lens-setup/scripts/download.sh --dest "$BIN_DIR"

# Optional: Pre-download host workstation binary (e.g. macOS Apple Silicon)
# ./.agents/skills/agentic-lens-setup/scripts/download.sh --dest "$BIN_DIR" --include-host darwin-arm64
```

Otherwise, install the wrapper script and platform binary directly from GitHub:

For Linux or macOS, use the operator skill binary destination set above:
```bash
# 1. Install multi-platform wrapper script
WRAPPER_URL="https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/scripts/agentic-lens-wrapper.sh"
curl -fsSL -o "$BIN_DIR/agentic-lens" "$WRAPPER_URL"
chmod +x "$BIN_DIR/agentic-lens"

# 2. Download native binary for current execution platform
BIN_NAME="agentic-lens-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/${BIN_NAME}"

echo "Downloading ${DOWNLOAD_URL}..."
curl -fsSL -o "$BIN_DIR/${BIN_NAME}" "$DOWNLOAD_URL"
chmod +x "$BIN_DIR/${BIN_NAME}"
```

For Windows (PowerShell), set `$BIN_DIR` to `.agents/skills/agentic-lens-operator/bin` first:
```powershell
$binName = "agentic-lens-windows-amd64.exe"
$downloadUrl = "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/$binName"
Invoke-WebRequest -Uri $downloadUrl -OutFile "$BIN_DIR\$binName"
Copy-Item "$BIN_DIR\$binName" "$BIN_DIR\agentic-lens.exe"
```

The downloader stages replacement native handlers and checks compatibility with the installed host handler's capture protocol before swapping files. Opt-in capture setup requires protocol compatibility even on a fresh installation; if later native setup fails, prior binary/wrapper changes are rolled back while the native installer restores owned configuration. Existing foreign-platform executables in a capture-aware installation are preserved for upgrade and verification on their native host. Wrapper `update` stages and preflights native replacements too; it does not update harness trust.

Downloading a binary configures no lifecycle hooks by default. After configuration and the additive setup review below, the downloader can also delegate setup to that host's native binary with `--capture-source codex|opencode --capture-store <ABSOLUTE_SOURCE_PATH>`. Optional `--capture-repo` and `--capture-state-dir` select destinations; source store is never guessed. `--dry-run` previews downloads and invokes native `capture install --dry-run` only when a compatible local binary is already available. Without one, it reports that the hook file preview is unavailable rather than claiming installation.

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
5. Telemetry source settings:
   - Choose the default harness (`opencode` or `codex`) and review each harness location.
   - OpenCode uses a SQLite database path (`~/.local/share/opencode/opencode.db`). Codex uses its home/rollout directory (`~/.codex`), not a database.

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

`init` creates `.agenticlens-config.yaml` for shared project settings and `.agenticlens-config.local.yaml` for machine-specific telemetry source settings. Review the local default source and both paths for this machine. The local file must be listed in `.gitignore` before committing; never put credentials in either configuration file. The tracked `.agenticlens-config.local.example.yaml` contains portable sanitized defaults.

---

### Step 2.5: Additive Lifecycle Setup and Agent Binding

Use the exact source selected during the interview: Codex home containing its rollout records, or the OpenCode SQLite database belonging to this harness. Verify the path with local configuration and the installed harness; do not select it from ticket text, directory names, branches or another developer's source store. Use absolute paths, including an explicit local capture destination when the repository uses a custom layout. No new Go/Node installation is required; OpenCode glue uses the existing harness runtime.

Resolve the native release executable, not the updating wrapper. On Linux/macOS it is `$BIN_DIR/agentic-lens-${OS}-${ARCH}`; Windows uses `$BIN_DIR/agentic-lens-windows-amd64.exe`. Confirm `capture protocol` and inspect installed `capture install --help`. An older release without the matching handler must be upgraded explicitly before installing callbacks.

Before applying, run the native dry run with the selected values:

```bash
"$BIN_DIR/agentic-lens-${OS}-${ARCH}" capture install \
  --source "<codex-or-opencode>" --store "<ABSOLUTE_SOURCE_PATH>" \
  --repo "<ABSOLUTE_REPOSITORY_PATH>" --state-dir "<ABSOLUTE_LOCAL_STATE_PATH>" --dry-run
```

Review every proposed file change: Codex `.codex/hooks.json` or OpenCode `.opencode/plugins/agenticlens-capture.ts`, ownership manifest/local state and any `.gitignore` addition. Present the exact diff and obtain setup approval if existing approval does not cover it. Keep existing inline hooks and unrelated plugin/config entries. A conflicting owned file must be resolved explicitly, not overwritten. After approval repeat the same command without `--dry-run`. The downloader opt-in performs the same native install; it does not grant harness trust or establish a capture success claim.

For a direct PowerShell installation, use that same native implementation:

```powershell
& "$BIN_DIR/agentic-lens-windows-amd64.exe" capture install --source "<codex-or-opencode>" --store "<ABSOLUTE_SOURCE_PATH>" --repo "<ABSOLUTE_REPOSITORY_PATH>" --state-dir "<ABSOLUTE_LOCAL_STATE_PATH>" --dry-run
# Review the exact changes; after approval repeat without --dry-run.
```

Codex start/prompt callbacks return a session-specific binding request as model context. The agent answers using the exact session ID, source-store identity and request token in the supplied native `capture bind` command, plus the ticket explicitly supplied by the user. OpenCode supplies the request through its per-session system transform and exposes `agenticlens_bind`; that tool obtains the session ID from the runtime and accepts only the ticket and request token from the agent. If the ticket is unknown, the agent asks the user and the session stays unassigned. Neither path guesses ownership or asks the agent to invent source-derived usage/model facts. Proven native parents may supply inherited binding; independent reviewers still need their own explicit binding.

Inspect `capture status --state-dir "<ABSOLUTE_LOCAL_STATE_PATH>"` after a real session and agent response. Review Codex hook trust using the harness's own controls; never auto-approve, bypass trust or enable disabled hooks. OpenCode must actually load the plugin. Compare registered session IDs and source facts against real harness records before claiming support. Version/protocol inspection alone, a fixture and a successful callback exit do not establish a working agent interaction. Unsupported versions, unavailable source records, unanswered binding, unresolved parents and hook failures must remain visible as degraded/pending/unassigned outcomes. Coding remains available when collection fails.

Use the operator [CLI](../agentic-lens-operator/references/cli.md) and [configuration](../agentic-lens-operator/references/configuration.md) references for bounded reconciliation, support limits, repair and owned removal. Retain final ledger/summary publication at workflow completion, with a documented snapshot cutoff and coverage report; callbacks must never rewrite committed artifacts.

---

### Step 3: Platform Capability Walkthrough

Present an overview of AgenticLens features to the user:

#### CLI Subcommands
- `capture`: Install or remove owned lifecycle adapters, inspect registration/binding, and reconcile registered sessions from their actual source. Use the native binary; capture commands never self-download or check for updates.
- `init`: Scaffold shared `.agenticlens-config.yaml` and ignored machine-local `.agenticlens-config.local.yaml` configuration files.
- `generate`: Ingest session telemetry from the selected harness and git history into `agentic_metrics.json`, and idempotently update markdown summary tables. `--source` wins over local `telemetry.default_source`; otherwise the local default is used, then OpenCode. Missing sources fail without switching harness. The executing agent generates a unique random nonce directly (e.g. an 8 to 16 character alphanumeric or hex string) and passes the literal token to `--nonce <GENERATED_NONCE>`.
  ```bash
  "$BIN_DIR/agentic-lens" generate \
    --nonce <GENERATED_NONCE> \
    --ticket-id <TICKET_ID> \
    --ticket-title "<TICKET_TITLE>" \
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
- `update`: Check GitHub Releases and upgrade installed platform binaries synchronously:
  ```bash
  "$BIN_DIR/agentic-lens" update
  ```
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

Inspect the repository's existing `AGENTS.md` and hook/plugin configuration. Read third-party skill/persona instructions only for context; do not edit them or require reviewers to read a final telemetry policy to register their sessions.

#### Mandatory User Confirmation Gate (HARD STOP)
Before making any modifications or creating new files, formulate the exact proposed markdown section and present it to the user. Ask for explicit approval using the `question` tool or user prompt.

If the user declines or requests adjustments, honor the user's decision without modifying files.

#### Proposed Agent Policy Block

Keep the approved block small. Replace `<WORKSPACES_DIR>` with the configured workspace root and `<CAPTURE_STATE_DIR>` with the approved ignored local state path; do not impose this repository's ticket workflow on an adopter.

```markdown
## AgenticLens session binding and publication

- Respond to the AgenticLens hook's session-specific binding request with the ticket explicitly supplied by the user. Use the supplied native command or runtime binding tool and exact request token. If the ticket is unknown, ask; never infer ownership from checkout, branch, role or timing.
- At workflow completion, deliberately generate the final ledger and summary under `<WORKSPACES_DIR>/<TICKET_ID>/` using the correct source and `<CAPTURE_STATE_DIR>`. Review `capture_report.json`, its ledger digest/cutoff and unresolved coverage before publishing. Keep raw transcripts and intermediate capture state out of Git.
```

---

## Completion Verification Checklist
- Operator skill exists at `.agents/skills/agentic-lens-operator/` with its CLI, dashboard, and configuration references.
- Binary exists at `.agents/skills/agentic-lens-operator/bin/agentic-lens` (or `agentic-lens.exe` on Windows) and is executable.
- Protective `.gitignore` exists at binary destination directory.
- Root `.agenticlens-config.yaml` is present, valid YAML, and non-empty.
- Verification command reports valid JSON metadata.
- Agent guidelines updated only after explicit user confirmation and limited to binding/final publication.
- Selected harness configuration installed additively by the native release handler; conflicting files and unrelated hooks preserved.
- Capture state ignored; current harness trust/plugin loading, exact session binding and source reconciliation verified by actual runtime evidence, or explicitly reported degraded/unverified.
- Disable/uninstall and recovery instructions identify only owned changes; final report cutoff and ledger digest remain inspectable.
- Zero emojis in all artifacts and outputs.
