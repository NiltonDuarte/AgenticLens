# fix(web,scripts): direct binary download and dynamic initial ticket selection (FIX-7)

## Summary of Changes
- **Direct Binary Downloader (`scripts/download-agentic-lens.sh`, `.opencode/skills/agentic-lens-setup/scripts/download.sh`)**:
  - Streamlined downloader scripts to fetch the precompiled standalone binary directly (`agentic-lens-${OS}-${ARCH}${BIN_EXT}`) from GitHub Releases via `curl` / `wget`.
  - Removed archive extraction overhead (`.tar.gz`/`.zip`, `tar`, `unzip`), temporary directories, and `gh` CLI invocations.
- **Dynamic Initial Ticket Selection (`web/src/App.tsx`, `web/src/views/TicketDeepDive.tsx`)**:
  - Initialized `selectedTicketId` to `null` instead of hardcoding `'AL-1'`.
  - Allowed Ticket Deep Dive to naturally default to the first available workspace in the scanned catalog when launching the server in projects with custom ticket prefixes.
  - Removed silent fallback redirection on not-found errors so explicitly requested missing IDs report errors accurately.
- **Embedded Web Assets (`backend/internal/adapters/web/dist/`)**:
  - Recompiled production frontend assets and updated embedded distribution files.
