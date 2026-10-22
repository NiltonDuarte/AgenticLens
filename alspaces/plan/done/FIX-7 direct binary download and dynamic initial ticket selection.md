# FIX-7: Direct Binary Download and Dynamic Initial Ticket Selection

## Metadata
- **Ticket ID**: FIX-7
- **Subsystem**: Scripts / Web / Setup / Presentation / Tests
- **Status**: Done
- **Type**: fix

## Problem Statement
1. Artifact downloader scripts (`scripts/download-agentic-lens.sh` and `.opencode/skills/agentic-lens-setup/scripts/download.sh`) attempted to download `.tar.gz`/`.zip` compressed archives and unpack them, and attempted to use `gh` CLI. Standalone precompiled binaries are directly distributed on GitHub Releases with embedded frontend assets and runtime dependencies, rendering archive handling and `gh` invocations unnecessary.
2. In `web/src/App.tsx`, the initial state for `selectedTicketId` was hardcoded to `'AL-1'`, causing the Ticket Deep Dive view to display a "Workspace AL-1 not found" error when running in projects with custom ticket prefixes or without an `AL-1` ticket.

## Solution
1. Refactored both downloader scripts to download the standalone binary directly (`agentic-lens-${OS}-${ARCH}${BIN_EXT}`) via `curl` with `wget` fallback, removing all archive tools and `gh` invocations.
2. Updated `web/src/App.tsx` to initialize `selectedTicketId` to `null`, allowing Ticket Deep Dive to dynamically default to the first available ticket in the project workspace catalog.
3. Removed silent fallback redirect in `TicketDeepDive.tsx` so genuine errors are accurately surfaced when an explicit invalid ticket ID is requested.
