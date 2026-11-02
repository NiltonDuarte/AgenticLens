---
name: agentic-lens-setup
description: Autonomous setup skill to bootstrap AgenticLens in OpenCode and Codex repositories, downloading precompiled release binaries, configuring project parameters, explaining telemetry features, and reviewing agent guidelines with mandatory user confirmation.
---

# AgenticLens Cross-Project Autonomous Setup Skill

Guides autonomous bootstrapping and adoption of AgenticLens in Codex and OpenCode software repositories using precompiled release artifacts from GitHub Releases (https://github.com/NiltonDuarte/AgenticLens/releases).

AgenticLens is an observability and developer telemetry platform for OpenCode and Codex workflows. OpenCode traces come from its local SQLite database; Codex traces come from rollout files under the Codex home directory.

This skill follows the autonomous agentic distribution model: it is completely self-contained and executable by any autonomous coding agent reading this manifest directly via shared repository URL or local path.

## Operational Mandates
- Strict Release-Only Mandate: Never compile binaries from Go/TypeScript source or request Go/Node.js toolchain installation in target repositories. Retrieve official precompiled release artifacts from GitHub Releases: https://github.com/NiltonDuarte/AgenticLens/releases.
- Complete Local Installation: Install this setup skill and all its scripts alongside `.agents/skills/agentic-lens-operator/`. Use the operator `scripts/agentic-lens.sh` launcher and keep platform executables under its `bin/` directory. Git tracking and ignore choices for scripts and executables belong to the adopting repository owner; do not stage files, commit them, or create/modify ignore rules during binary/skill installation or repair.
- Lifecycle Capture: Install reviewed additive callbacks through the native release handler. Exact source/project/session proof and an explicit agent binding response establish collection scope; never infer ownership from branch, role or timing. A bounded native worker publishes automatically after binding. Preserve unrelated hooks, edited files and user-controlled trust.
- Mandatory User Confirmation Gate: Inspect existing OpenCode agent instructions (`AGENTS.md` or `.agents/agent/*.md`) and propose operational policies. Never modify target repository files without presenting the exact text and receiving affirmative user confirmation.
- Zero Emojis Mandate: Zero emojis permitted across all prompts, CLI output, generated files, configurations, and documentation.

---

## Step-by-Step Setup Workflow

### Step 1: Install Complete Skills and Precompiled Binaries

Before modifying the adopting repository, present the exact proposed files, source release URLs, destinations and any existing-file diffs. Obtain affirmative approval, or use approval already granted for that scope. Preserve locally edited files. Binary/skill installation and repair must not run `init`, change existing configuration, stage files, or impose Git ignore rules. Reviewed native lifecycle setup in Step 2.5 is a separate opt-in operation.

#### 1.1 Complete Installation Manifest

Install both skills, not only the operator. All paths below are relative to `.agents/skills/`:

| File | Purpose |
| --- | --- |
| `agentic-lens-setup/SKILL.md` | Setup and repair instructions |
| `agentic-lens-setup/scripts/install.sh` | Complete skill bundle installer |
| `agentic-lens-setup/scripts/download.sh` | Explicit platform binary installation/repair |
| `agentic-lens-setup/scripts/agentic-lens-wrapper.sh` | Legacy binary command compatibility |
| `agentic-lens-operator/SKILL.md` | Day-to-day operation |
| `agentic-lens-operator/references/cli.md` | CLI and worktree recovery |
| `agentic-lens-operator/references/configuration.md` | Configuration reference |
| `agentic-lens-operator/references/dashboard.md` | Dashboard reference |
| `agentic-lens-operator/scripts/agentic-lens.sh` | Launcher outside binary storage |

The installer also writes `agentic-lens-setup/RELEASE_VERSION`, containing the selected release tag. That same tag is the default for later binary repair. It preflights the full bundle and refuses differing existing files, including the version record, before copying any skill files. Review and reconcile conflicts explicitly; do not delete edited files to make installation pass.

#### 1.2 Bootstrap When Setup Scripts Are Missing

Choose an official release tag from https://github.com/NiltonDuarte/AgenticLens/releases that contains this complete skill bundle. Replace `<RELEASE_TAG>` below with that tag. Read and review the selected release's manifest and installer before running it; download the installer into a temporary location, not into an existing installed skill. Do not pipe an unreviewed script into a shell.

```bash
# Run from the adopting repository root, after reviewing/approving the files.
AGENTIC_LENS_RELEASE="<RELEASE_TAG>"
AGENTIC_LENS_BOOTSTRAP="$(mktemp -d)"
curl -fsSL -o "$AGENTIC_LENS_BOOTSTRAP/install.sh" \
  "https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/$AGENTIC_LENS_RELEASE/.agents/skills/agentic-lens-setup/scripts/install.sh"
# Inspect the downloaded installer before execution.
bash "$AGENTIC_LENS_BOOTSTRAP/install.sh" --version "$AGENTIC_LENS_RELEASE" --dest .
rm -rf "$AGENTIC_LENS_BOOTSTRAP"
```

Alternatively, an already available installer can use `--version latest`: it resolves the official latest release once and pins every skill file to that tag. Never mix skill scripts from moving `main` with a selected binary release. A release lacking these files fails explicitly; do not silently substitute another revision.

On Windows, these shell commands require Git Bash. For native PowerShell without Bash, retrieve the same manifest files from one selected release tag and save `RELEASE_VERSION` with that tag, preserving local edits; this is a manual installation, not a PowerShell automation provided by this skill. Fetch the Windows amd64 binary from that same release and invoke it directly as described below. Native Windows arm64 release artifacts are not provided.

#### 1.3 Explicit Binary Installation or Worktree Repair

From the adopting repository root, after authorization for the binary download:

```bash
bash .agents/skills/agentic-lens-setup/scripts/download.sh
# Optional: also install a macOS Apple Silicon host binary from a Linux container.
bash .agents/skills/agentic-lens-setup/scripts/download.sh --include-host darwin-arm64
```

The downloader defaults to the sibling operator `bin/` directory and uses `RELEASE_VERSION` unless `--version` or `AGENTIC_LENS_VERSION` explicitly selects another release. Existing binary contents are retained unless replacement was explicitly authorized with `--force`; explicit repair restores a missing execute permission before reporting success. If the owner chose not to carry `RELEASE_VERSION` into a worktree (or the file is empty), choose the intended release explicitly with `--version <RELEASE_TAG>` or `AGENTIC_LENS_VERSION`. The installed downloader fails before downloading when release identity is unavailable; it never assumes `latest`. Offline `--local-bin` repair does not require a release tag because it selects no remote artifact. `--dest` or `AGENTIC_LENS_BIN_DIR` selects another directory; use that same environment variable with the launcher. `--dry-run` prints destination/artifact URLs without writing or downloading. `--local-bin` installs an already available native prebuilt executable for offline repair. No Go/Node compilation is involved.

The ordinary launcher does not download or check for updates. It detects the current OS/architecture and executes the native file in `operator/bin`, preserving the current directory, arguments and exit status. It reports an explicit repair command when that file is missing. The legacy `bin/agentic-lens` wrapper is retained for existing callers and keeps its established automatic-download/update-check behavior; use the new launcher for explicit-download operation.

For native Windows PowerShell, with `$release` set to the same selected release tag, download to a temporary file and replace only after success. Confirm replacement authorization if a destination already exists:

```powershell
$binDir = ".agents/skills/agentic-lens-operator/bin"
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
Invoke-WebRequest -Uri "https://github.com/NiltonDuarte/AgenticLens/releases/download/$release/agentic-lens-windows-amd64.exe" -OutFile "$binDir/agentic-lens-windows-amd64.exe.tmp"
Move-Item -Path "$binDir/agentic-lens-windows-amd64.exe.tmp" -Destination "$binDir/agentic-lens-windows-amd64.exe"
& "$binDir/agentic-lens-windows-amd64.exe" version --json
```

Binary repair and capture setup have distinct effects. Ordinary installation/repair preserves configuration and Git policy. Explicit opt-in capture setup uses --capture-source codex|opencode and the verified absolute --capture-store, with optional --capture-repo and --capture-state-dir. Preview the exact native setup changes, including its owned configuration, manifests and local-state ignore entry; apply only within that approval. A binary download or repair alone does not install or trust callbacks.

The downloader stages native replacement and checks installed capture-protocol compatibility before swapping an active handler. Failed opt-in native setup restores prior binary/wrapper files while the native installer rolls back owned configuration changes. Preserve foreign-platform artifacts whose compatibility cannot be verified locally. A missing compatible handler makes setup incomplete; do not claim a dry-run configuration preview without one.

Native PowerShell capture-aware upgrades must stage the candidate at a separate path, compare capture protocol with the working handler, and preserve a backup until reviewed native setup succeeds. Restore it on failure. Do not overwrite an active hook handler using the initial-install example without this compatibility/rollback procedure. No binary repair changes harness trust.

#### 1.4 Worktree Diagnosis and Verification

A worktree receives only files tracked in its checked-out revision. The repository owner decides whether to track or ignore installed skills, launchers, binaries, and local configuration. Explain these choices; do not change Git policy for them.

- If the launcher and setup scripts exist but the native binary does not, use the explicit local binary repair command above. If the binary exists without execute permission, that repair restores permission without replacing its contents or accessing the network. If the release marker is absent, explicitly select the intended release or use `--local-bin`; do not infer a release from the skill text.
- If setup/launcher files are also absent, use the complete bootstrap in section 1.2. Do not invoke a nonexistent local repair script or search another worktree for executables.
- If `.agenticlens-config.local.yaml` is absent, diagnose it separately. Select the actual harness explicitly with `--source codex` or `--source opencode`, and use this machine's `--codex-home` or `--db-path` where needed. Do not copy another worktree's machine settings or rerun `init` over the existing shared configuration.
- If repair is offline and no local prebuilt executable is available, report the missing artifact and retry command; do not claim setup succeeded.

Verify installed binary metadata from the current worktree:

```bash
AGENTIC_LENS=".agents/skills/agentic-lens-operator/scripts/agentic-lens.sh"
bash "$AGENTIC_LENS" version --json
```

Check the returned `version`, `commit`, and `build_date`. Download success and executable metadata do not establish cryptographic integrity; verify against the selected release's `SHA256SUMS` when available before executing newly downloaded artifacts.

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
bash "$AGENTIC_LENS" init \
  --project-name "<PROJECT_NAME>" \
  --ticket-prefixes "<TICKET_PREFIXES>" \
  --workspaces "<WORKSPACES_DIR>" \
  --src-folders "<SRC_FOLDERS>" \
  --db-path "<DB_PATH>"
```

Alternatively, review generated YAML with `--dry-run` before writing:

```bash
bash "$AGENTIC_LENS" init \
  --project-name "<PROJECT_NAME>" \
  --ticket-prefixes "<TICKET_PREFIXES>" \
  --src-folders "<SRC_FOLDERS>" \
  --dry-run
```

`init` creates `.agenticlens-config.yaml` for shared project settings and `.agenticlens-config.local.yaml` for machine-specific telemetry source settings. Review the local default source and both paths for this machine. The local file must be listed in `.gitignore` before committing; never put credentials in either configuration file. The tracked `.agenticlens-config.local.example.yaml` contains portable sanitized defaults.

---

### Step 2.5: Additive Lifecycle Setup and Agent Binding

Use the exact source selected during the interview: Codex home containing its rollout records, or the OpenCode SQLite database belonging to this harness. Verify the path with local configuration and the installed harness; do not select it from ticket text, directory names, branches or another developer's source store. Use absolute paths, including an explicit local capture destination when the repository uses a custom layout. No new Go/Node installation is required; OpenCode glue uses the existing harness runtime.

Resolve the native release executable, not the updating wrapper. Set BIN_DIR to the verified operator bin directory (default `.agents/skills/agentic-lens-operator/bin`, or the explicit AGENTIC_LENS_BIN_DIR installation override), and identify OS/ARCH from the executing environment. On Linux/macOS the native path is `$BIN_DIR/agentic-lens-${OS}-${ARCH}`; Windows uses `$BIN_DIR/agentic-lens-windows-amd64.exe`. Confirm `capture protocol` and inspect installed `capture install --help`. An older release without the matching handler must be upgraded explicitly before installing callbacks.

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

Codex start/prompt callbacks return a session-specific binding request as model context. The agent answers using the exact session ID, source-store identity and request token in the supplied native `capture bind` command, plus the ticket explicitly supplied by the user. OpenCode supplies the request through its per-session system transform and exposes `agenticlens_bind`; that tool obtains the session ID from the runtime and accepts explicit ticket/title and request token from the agent. If the ticket is unknown, the agent asks the user and the session stays unassigned. Neither path guesses ownership or asks the agent to invent source-derived usage/model facts. Proven native parents may supply inherited binding; independent reviewers still need their own explicit binding.

Inspect `capture status --state-dir "<ABSOLUTE_LOCAL_STATE_PATH>"` after a real session and agent response. Review Codex hook trust using the harness's own controls; never auto-approve, bypass trust or enable disabled hooks. OpenCode must actually load the plugin. Compare registered session IDs and source facts against real harness records before claiming support. Version/protocol inspection alone, a fixture and a successful callback exit do not establish a working agent interaction. Unsupported versions, unavailable source records, unanswered binding, unresolved parents and hook failures must remain visible as degraded/pending/unassigned outcomes. Coding remains available when collection fails.

Use the operator [CLI](../agentic-lens-operator/references/cli.md) and [configuration](../agentic-lens-operator/references/configuration.md) references for bounded reconciliation, support limits, repair and owned removal. Native callbacks schedule automatic ledger/summary publication after explicit binding; verify its cutoff and coverage at handoff. Manual generate is a recovery fallback. Detailed reports are optional; when disabled use compact summary coverage and retain old reports as stale. Publication never commits files.

---


### Step 3: Platform Capability Walkthrough

Present an overview of AgenticLens features to the user:

#### CLI Subcommands
- `capture`: Install/remove owned adapters, inspect registration/binding and reconcile exact source sessions. Native capture handlers stay offline and schedule bounded automatic publication.
- `init`: Scaffold shared `.agenticlens-config.yaml` and ignored machine-local `.agenticlens-config.local.yaml` configuration files.
- `generate`: Ingest session telemetry from the selected harness and git history into `agentic_metrics.json`, and idempotently update markdown summary tables. `--source` wins over local `telemetry.default_source`; otherwise the local default is used, then OpenCode. Missing sources fail without switching harness. The executing agent generates a unique random nonce directly (e.g. an 8 to 16 character alphanumeric or hex string) and passes the literal token to `--nonce <GENERATED_NONCE>`.
  ```bash
  bash "$AGENTIC_LENS" generate \
    --nonce <GENERATED_NONCE> \
    --ticket-id <TICKET_ID> \
    --ticket-type <CONVENTIONAL_TYPE> \
    --json-output workspaces/<TICKET_ID>/agentic_metrics.json \
    --append workspaces/<TICKET_ID>/SUMMARY.md
  ```
- `scan`: Scan workspace ledgers, reporting active sessions, iteration counts, context window health, and token costs:
  ```bash
  bash "$AGENTIC_LENS" --list
  ```
- `server`: Start embedded web dashboard and real-time streaming API service:
  ```bash
  bash "$AGENTIC_LENS" server --port 8080
  ```
- `backfill`: Batch telemetry backfill, git diff fact recomputation, and schema upgrades.
- `migrate`: Stepwise schema migration engine across ledger versions.
- Binary upgrades: explicitly select the intended release and authorize replacement, then run the setup downloader with `--version <RELEASE_TAG> --force`. Skill bundle updates require reviewing/reconciling differing installed files separately; do not silently upgrade scripts during binary repair. The legacy wrapper also supports its existing `update` command.
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

Inspect the repository for existing OpenCode agent instructions:
- `AGENTS.md`
- `.agents/agent/*.md`

#### Mandatory User Confirmation Gate (HARD STOP)
Before making any modifications or creating new files, formulate the exact proposed markdown section and present it to the user. Ask for explicit approval using the `question` tool or user prompt.

If the user declines or requests adjustments, honor the user's decision without modifying files.

#### Proposed Agent Policy Block

Keep the approved block small. Replace `<WORKSPACES_DIR>` with the configured workspace root and `<CAPTURE_STATE_DIR>` with the approved ignored local state path; do not impose this repository's ticket workflow on an adopter.

```markdown
## AgenticLens session binding and publication

- Respond to the AgenticLens hook's session-specific binding request with the ticket explicitly supplied by the user. Use the supplied native command or runtime binding tool and exact request token. If the ticket is unknown, ask; never infer ownership from checkout, branch, role or timing.
- Verify automatic native publication under `<WORKSPACES_DIR>/<TICKET_ID>/` after explicit binding. Use manual generate only for recovery. When detailed reports are enabled, review `capture_report.json`, its ledger digest/cutoff and unresolved coverage before committing. When disabled, use compact summary coverage and leave existing reports retained as stale. Keep raw transcripts and intermediate capture state out of Git.
```

---

## Completion Verification Checklist
- Operator skill exists at `.agents/skills/agentic-lens-operator/` with its CLI, dashboard, and configuration references.
- Setup skill exists with every script in section 1.1 and the selected `RELEASE_VERSION`.
- Operator launcher exists outside binary storage; the native binary for this host exists and executes.
- Git index is unchanged. Binary/skill installation and repair preserve ignore policy; separately approved native capture setup may add its exact local-state ignore entry. The owner decides which installed files to track.
- Root `.agenticlens-config.yaml` is present, valid YAML, and non-empty.
- Verification command reports valid JSON metadata.
- Agent guidelines updated only after explicit user confirmation.
- Zero emojis in all artifacts and outputs.

- Reviewed lifecycle setup preserves unrelated hooks, local edits and user trust. Native capture installation may add the approved local-state ignore entry; binary/skill repair alone does not change Git policy.
- Real callback delivery, explicit agent response and automatic source publication are verified, or support remains explicitly degraded/unverified. Version output alone is insufficient.
- Capture defaults and report policy are documented in the [operator configuration](../agentic-lens-operator/references/configuration.md#automatic-publication-policy). Detailed report disabled by configuration leaves any older report stale; summary coverage remains available.
