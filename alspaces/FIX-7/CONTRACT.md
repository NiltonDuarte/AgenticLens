# Technical Contract: Ticket FIX-7

## Problem Statement
1. Artifact Downloader Script Overhead: `scripts/download-agentic-lens.sh` and `.opencode/skills/agentic-lens-setup/scripts/download.sh` previously attempted to download `.tar.gz`/`.zip` compressed archives and unpack them using archive tools, along with attempts to use the `gh` CLI. Standalone precompiled binaries are already distributed directly on GitHub Releases, embedding the complete frontend and runtime dependencies. The downloader scripts should directly retrieve the standalone binary via standard HTTP tools (`curl` / `wget`) without archive unpacking, temporary extraction trees, or `gh` dependencies.
2. Web Dashboard Initial Ticket Selection: When launching the server in a project that does not contain an `AL-1` workspace, navigating to the "Ticket Deep Dive" tab failed with "Workspace AL-1 not found" because `web/src/App.tsx` hardcoded the initial state `selectedTicketId` to `'AL-1'`.

## Core Architectural Invariants

### 1. Direct Binary Distribution
- `scripts/download-agentic-lens.sh` and `.opencode/skills/agentic-lens-setup/scripts/download.sh` download the platform-specific precompiled executable directly (`agentic-lens-${OS}-${ARCH}${BIN_EXT}`).
- Zero archive extraction dependencies (`tar`, `gzip`, `unzip`).
- Zero `gh` CLI requirement; downloads rely solely on standard HTTP clients (`curl`, with fallback to `wget`).
- Protects destination directory with `.gitignore` and enforces `chmod +x`.

### 2. Clean Initial Dashboard State
- `web/src/App.tsx` initializes `selectedTicketId` to `null`.
- In `web/src/views/TicketDeepDive.tsx`, when no ticket ID is specified, it naturally selects the first available ticket in the scanned catalog (`currentLedgers[0].ticket_id`), or displays the clean empty state if the project has zero workspaces.
- No silent fallback redirects: if a specific ticket ID is requested and not found, the error is accurately surfaced rather than redirecting to an unrelated workspace.

### 3. Verification & Safety Invariants
- Frontend unit test suite passes with zero regressions.
- Embedded frontend bundle rebuilt (`npm run build`) and re-embedded into Go binaries (`make build`).
- Full repository quality gates pass (`make check`, `make test`).
- Zero emojis across all scripts, code, tests, and documentation.

## Scope Boundaries

### In-Scope Files
- `scripts/download-agentic-lens.sh`: Direct binary downloader via curl/wget.
- `.opencode/skills/agentic-lens-setup/scripts/download.sh`: Skill co-located binary downloader matching direct binary flow.
- `web/src/App.tsx`: Default `selectedTicketId` to `null`.
- `web/src/views/TicketDeepDive.tsx`: Direct ticket fetch with default catalog fallback when unselected.
- `web/src/views/TicketDeepDive.test.tsx`: Unit tests for default selection without preselected ticket.
- `backend/internal/adapters/web/dist/*`: Recompiled embedded web assets.
- `alspaces/FIX-7/`: Workspace deliverables (`CONTRACT.md`, `SUMMARY.md`, `PULL_REQUEST.md`).

### Out-of-Scope Files
- Backend domain models and core metric aggregations.

## Acceptance Criteria
1. `download-agentic-lens.sh` retrieves the standalone binary directly via curl/wget without archive extraction or `gh` invocations.
2. Running `agentic-lens server` in projects without an `AL-1` ticket opens the Ticket Deep Dive tab smoothly, defaulting to the first available workspace in the project catalog.
3. `make check` and `make test` pass cleanly.
4. Zero emojis in all touched files.
