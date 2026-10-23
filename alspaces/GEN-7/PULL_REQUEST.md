# GEN-7: Multi-Platform Binary Selection and Host-Container Wrapper Support in Setup

## Summary
Resolves cross-platform execution failures when autonomous coding agents run inside Linux containers (e.g. Docker, devcontainers) while developers work on host workstations (macOS Apple Silicon / Intel, Windows). Transforms `bin/agentic-lens` into a self-healing POSIX wrapper script that dynamically delegates to native platform binaries (`agentic-lens-${OS}-${ARCH}`), auto-downloads missing binaries from GitHub Releases on demand, and performs zero-latency background update checks.

## Changes
- **`scripts/agentic-lens-wrapper.sh`**:
  - Dynamically detects runtime OS and CPU architecture.
  - Automatically fetches missing platform binaries on first run across host/container boundaries.
  - Implements non-blocking detached background update checks throttled to 24 hours (`.last_update_check`).
  - Emits non-intrusive stderr update notification when a newer release is detected in `.latest_version`.
  - Added dedicated synchronous `update` subcommand (`./bin/agentic-lens update`) with `--version` support.
  - Suppresses update checks for development builds (`-dev`, `-dirty`) and respects `AGENTIC_LENS_NO_UPDATE_CHECK=1`.
- **`scripts/download-agentic-lens.sh` & `.opencode/skills/agentic-lens-setup/scripts/download.sh`**:
  - Installs the wrapper script as `bin/agentic-lens`.
  - Downloads the current execution platform's native binary into `bin/agentic-lens-${OS}-${ARCH}`.
  - Added `--include-host <target>` flag to pre-cache host workstation binaries during container setup.
  - Added `--all-platforms` flag to pre-cache all five official release targets.
- **`.opencode/skills/agentic-lens-setup/`**:
  - Co-located `agentic-lens-wrapper.sh` in `scripts/` for standalone skill distribution.
  - Updated `SKILL.md` to reflect multi-platform wrapper installation and `update` subcommand.
- **`docs/SETUP_MANUAL.md`**:
  - Documented container sandbox vs host machine execution architecture (Section 6.6).
  - Updated setup pathways to direct binary and wrapper distribution.
  - Added `agentic-lens update` to CLI reference (Section 5.1).
- **Tests & Verification**:
  - Added `backend/cmd/agentic-lens/wrapper_and_downloader_test.go` verifying dispatching, update notifications, local installs, and zero-emoji compliance.
  - Added `scripts/tests/test_wrapper.sh` for POSIX shell verification.
  - Verified `make check` and `make test-backend` pass cleanly.
