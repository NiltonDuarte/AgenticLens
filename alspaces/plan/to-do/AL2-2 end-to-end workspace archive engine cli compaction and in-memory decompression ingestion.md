h3. Title: AL2-2: End-to-End Workspace Archive Engine, CLI Compaction, and In-Memory Decompression Ingestion

*TARGET*: AgenticLens (domain | adapters | cli | http | presentation | web | tests)

*CONTEXT*
 
As multi-agent development iterations accumulate across weeks and months, dozens of completed workspace ticket folders (`workspaces/<PREFIX>-NN/`) consume repository space, waste filesystem inodes, and slow down directory walking. Delivering an end-to-end vertical slice from archive creation and cleanup to transparent server-side in-memory reading and web dashboard presentation enables teams and CI to compact older workspaces into `.tar.gz` archives without losing the ability to view, query, and filter historical telemetry.

*GOAL*
 
Deliver an end-to-end workspace archiving and transparent ingestion engine, incorporating the `agentic-lens archive` subcommand with date-based and interactive compaction, atomic `.tar.gz` packaging with directory cleanup, transparent in-memory archive reading in repository scanner and HTTP endpoints, and archive status presentation in the web dashboard.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Define `ArchiveManifest` and `ArchivedTicketEntry` models in `internal/domain`. Implement pure `FilterWorkspacesForArchive(entries []WorkspaceInfo, cutoff time.Time) []WorkspaceInfo`, `IsArchived` ledger annotation, and deduplication logic prioritizing live on-disk directories over archived copies. Pure functions, zero I/O imports.
* Boundary Adapter (Archive Creation): Implement `CreateWorkspaceArchive` in `internal/adapters/archive.go`. Use standard Go `archive/tar` and `compress/gzip` to bundle ticket folders into `.tar.gz` archives with `manifest.json`. Write atomically (`.tmp`), verify checksums and tarball entries, and safely purge source directories unless `--keep` is specified. Implement interactive terminal prompt selector for non-flag invocations in a TTY.
* Boundary Adapter (Archive Ingestion): Update `ScanWorkspaces` in `internal/adapters/repository.go` to discover `.tar.gz` archives in `workspaces/` (and `archives/`). Parse `agentic_metrics.json` directly from the tar stream without extracting files to disk. Update `LoadTicketWorkspace` and `ReadTicketArtifacts` to stream `SUMMARY.md` and `CONTRACT.md` from archives.
* CLI Presentation: Add `agentic-lens archive` subcommand in `backend/cmd/agentic-lens/main.go` supporting `--before`, `--workspaces`, `--output`, `--keep`, and `--dry-run`.
* Web Presentation: Expose `is_archived` in API payloads. In `web/`, add "Archived" badges to ticket rows in `WorkspacesTable.tsx`, an archive filter toggle ("All", "Active", "Archived") in `FilterBar.tsx`, and archive origin banner in `TicketDeepDive.tsx`.
* Quality: Domain unit tests with zero mocks. Integration tests in `internal/adapters` using `t.TempDir()` verifying archive packaging, atomic safety, directory cleanup, and in-memory streaming ingestion. Vitest unit tests in `web/` verifying badge rendering and archive filtering.

*ACCEPTANCE CRITERIA*
 
* `agentic-lens archive --before <date>` compacts workspaces completed before the specified date into a `.tar.gz` archive and safely deletes original folders from disk.
* In a terminal without `--before`, the CLI presents an interactive selection prompt to select workspaces for compaction.
* Passing `--keep` leaves original folders intact on disk; `--dry-run` reports actions without writing files or deleting folders.
* `ScanWorkspaces` and the HTTP server transparently discover and read archived tickets directly in-memory with zero disk extraction.
* `GET /api/workspaces` returns both live and archived tickets, correctly tagging archived items with `is_archived: true`.
* Web dashboard table displays distinct "Archived" badges and allows filtering between Active, Archived, or All tickets.
* Strict typing maintained across Go and TypeScript with zero `any`; `make check` and `make test` pass cleanly.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
