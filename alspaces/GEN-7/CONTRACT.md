# Technical Contract: Ticket GEN-7

## Problem Statement
1. Multi-Platform Container vs Host Discrepancy: When autonomous coding agents run inside a containerized sandbox (e.g. Linux x86_64 or ARM64 in Docker/devcontainer) while the developer works on a host workstation (e.g. macOS Apple Silicon or Intel), installing a single platform-specific binary to `bin/agentic-lens` causes execution failures on the opposite platform. If setup downloads a Linux binary, the host user cannot run `agentic-lens server` to view telemetry. Asking the user for an either/or choice during setup breaks the agent inside the container.
2. Update Verification Without Execution Latency: Developers need visibility when newer AgenticLens releases become available without embedding update logic into the Go binary (which would overwrite the multi-platform wrapper script) and without adding network latency to CLI commands (`generate`, `server`, `scan`).

## Core Architectural Invariants

### 1. Multi-Platform Wrapper Architecture
- `bin/agentic-lens` is a lightweight POSIX shell script installed from `scripts/agentic-lens-wrapper.sh`.
- Compiled native platform binaries reside side-by-side in the same directory using the naming convention `agentic-lens-${OS}-${ARCH}${BIN_EXT}` (e.g. `agentic-lens-linux-amd64`, `agentic-lens-darwin-arm64`).
- At runtime, `bin/agentic-lens` detects `uname -s` and `uname -m` dynamically and dispatches execution to the appropriate platform binary.

### 2. Self-Healing Binary Download
- If the requested platform binary is missing from the directory when executed (for example, first run of `./bin/agentic-lens server` on a macOS host after setup was performed inside a Linux container), the wrapper script automatically retrieves the missing platform binary directly from GitHub Releases via `curl` (with fallback to `wget`), sets executable permissions, and executes it.
- A protective `.gitignore` in the binary directory excludes all platform binaries (`*` and `!.gitignore`).

### 3. Non-Blocking Background Update Check
- The wrapper script performs update checks in a detached background subshell (`( check_latest & ) >/dev/null 2>&1`) to introduce zero latency to foreground commands.
- Update checks are throttled to run at most once every 24 hours using a timestamp sentinel file (`.last_update_check`).
- Discovered new versions are stored in `.latest_version`. On subsequent runs, if a newer version is known, the wrapper prints a single concise notice to stderr:
  `Notice: AgenticLens update available: <latest> (current: <current>). Run './bin/agentic-lens update' to upgrade.`
- Can be disabled via `AGENTIC_LENS_NO_UPDATE_CHECK=1`. Zero CI flags needed.

### 4. Explicit Update Subcommand
- Running `./bin/agentic-lens update` triggers an explicit, synchronous upgrade:
  - Resolves latest release tag from GitHub Releases.
  - Updates all existing platform binaries present in the binary directory.
  - Updates `.latest_version` and `.last_update_check`.

### 5. Downloader and Setup Integration
- `scripts/download-agentic-lens.sh` and `.opencode/skills/agentic-lens-setup/scripts/download.sh` install the wrapper script as `bin/agentic-lens` and download the current execution environment's platform binary (`agentic-lens-${OS}-${ARCH}`).
- Both downloaders support optional flags `--all-platforms` and `--include-host <os-arch>` to pre-cache host binaries during initial setup.
- `.opencode/skills/agentic-lens-setup/SKILL.md` and `docs/SETUP_MANUAL.md` document the wrapper architecture, multi-platform container/host workflow, and background update behavior.

### 6. Quality & Zero Emojis
- Full automated test suite verifying platform detection, self-healing downloads, background update checks, and update subcommands.
- Zero emojis across all scripts, docstrings, documentation, and workspace artifacts.
- `make check` and `make test` pass cleanly.

## Scope Boundaries

### In-Scope Files
- `scripts/agentic-lens-wrapper.sh`
- `scripts/download-agentic-lens.sh`
- `.opencode/skills/agentic-lens-setup/scripts/download.sh`
- `.opencode/skills/agentic-lens-setup/SKILL.md`
- `docs/SETUP_MANUAL.md`
- `scripts/tests/wrapper_test.sh` (or test scripts for wrapper and downloader)
- `alspaces/GEN-7/` workspace artifacts (`CONTRACT.md`, `SUMMARY.md`, `PULL_REQUEST.md`)

### Out-of-Scope Files
- Backend domain calculation engines (`backend/internal/domain/*`)
- Frontend dashboard components (`web/src/*`)
