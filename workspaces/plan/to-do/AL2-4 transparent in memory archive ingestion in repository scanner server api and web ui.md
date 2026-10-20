h3. Title: AL2-4: Transparent In-Memory Archive Ingestion in Repository Scanner, Server API, and Web UI

*TARGET*: AgenticLens (adapters | http | presentation | web | tests)

*CONTEXT*
 
Following the creation of workspace `.tar.gz` archives in AL2-3, the AgenticLens server and CLI scanner must be able to load, query, and visualize archived tickets without requiring operators to decompress them back onto the filesystem. Transparent in-memory archive reading ensures historical metrics, summaries, and contracts remain available across the HTTP API and the TypeScript dashboard, while distinguishing archived records from active on-disk workspaces.

*GOAL*
 
Enhance `ScanWorkspaces`, `LoadTicketWorkspace`, and `ReadTicketArtifacts` to transparently discover and read workspace ledgers and markdown artifacts directly from `.tar.gz` archives in-memory, expose archive status across REST endpoints, and display visual archive indicators in the web dashboard.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Extend `AgenticMetricsLedger` with `IsArchived bool` and `ArchiveSource string`. Implement deduplication logic ensuring active on-disk workspaces take precedence over archived copies if duplicate ticket IDs exist.
* Boundary Adapter: Update `ScanWorkspaces` in `internal/adapters/repository.go` to scan for `.tar.gz` files in `workspaces/` (and `workspaces/archives/`). Parse `agentic_metrics.json` directly from the tar stream without extracting files to disk. Update `LoadTicketWorkspace` and `ReadTicketArtifacts` to stream `CONTRACT.md` and `SUMMARY.md` from the appropriate archive when requested.
* Presentation: Add `is_archived` field to `WorkspaceDetailResponse` and list endpoints. In `web/`, add an "Archived" badge to ticket rows in `WorkspacesTable.tsx`, add an "Archived / Active" filter toggle in `FilterBar.tsx`, and display an archive origin header in `TicketDeepDive.tsx`.
* Quality: Comprehensive Go adapter tests verifying in-memory reading of tar archives, corrupt archive handling, deduplication precedence, and artifact extraction. TypeScript Vitest unit tests verifying badge rendering, filtering by archive state, and deep dive presentation.

*ACCEPTANCE CRITERIA*
 
* `ScanWorkspaces` transparently loads ledgers from both regular workspace directories and `.tar.gz` archives located in the workspaces path.
* `GET /api/workspaces` includes archived tickets with `is_archived: true` and `archive_source: "<filename>"`.
* `GET /api/workspaces/{ticketId}` and artifact endpoints return full ledger facts, `SUMMARY.md`, and `CONTRACT.md` directly from the archive.
* If a ticket exists both on disk and in an archive, the live on-disk version is used with zero crashes or duplicate ledger errors.
* Web dashboard table displays clear "Archived" badges and allows filtering between Active, Archived, or All tickets.
* Strict typing maintained across Go and TypeScript with zero `any`.
* All Go and frontend test suites pass cleanly (`make test`).

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
