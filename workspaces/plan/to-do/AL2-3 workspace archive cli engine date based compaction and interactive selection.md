h3. Title: AL2-3: Workspace Archive CLI Engine, Date-Based Compaction, and Interactive Selection

*TARGET*: AgenticLens (domain | adapters | cli | tests)

*CONTEXT*
 
As multi-agent development iterations accumulate across weeks and months, dozens of completed workspace ticket folders (`workspaces/<PREFIX>-NN/`) consume repository space, waste filesystem inodes, and slow down directory walking. Teams and continuous integration pipelines need an automated `archive` command that packages older workspace directories into a verified `.tar.gz` archive prior to a specific cutoff date, or interactively allows engineers to select workspaces for compaction, safely cleaning up the archived source folders upon verified archive creation.

*GOAL*
 
Implement the `agentic-lens archive` subcommand supporting date-based batch archiving (`--before <date>`) and interactive terminal workspace selection, creating verified `.tar.gz` archive bundles with embedded manifests, and safely deleting compacted ticket directories from disk while offering `--keep` and `--dry-run` safety flags.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Define `ArchiveManifest` and `ArchivedTicketEntry` models in `internal/domain`. Implement pure `FilterWorkspacesForArchive(entries []WorkspaceInfo, cutoff time.Time) []WorkspaceInfo` and archive verification validation invariants. Zero I/O imports.
* Boundary Adapter: Implement `CreateWorkspaceArchive(options ArchiveOptions) (*ArchiveResult, error)` in `internal/adapters/archive.go`. Use standard Go `archive/tar` and `compress/gzip` to bundle each ticket workspace folder (including `agentic_metrics.json`, `SUMMARY.md`, `CONTRACT.md`, and iteration reviews) alongside a generated `manifest.json`. Write atomically to temporary file before finalizing. Verify tar integrity and SHA256 checksums before purging source folders. Implement interactive terminal prompt selector for non-flag invocations in a TTY.
* Presentation: Add `agentic-lens archive` subcommand in `backend/cmd/agentic-lens/main.go` accepting `--before`, `--workspaces`, `--output`, `--keep`, and `--dry-run` flags. Print clear progress and summary table indicating archived ticket count, compressed bytes saved, and cleanup status.
* Quality: Table-driven unit tests in `internal/domain` with zero mocks. Integration tests in `internal/adapters` using `t.TempDir()` with mock workspace folders verifying tarball packaging, manifest generation, atomic safety, `--dry-run`, `--keep`, and directory cleanup.

*ACCEPTANCE CRITERIA*
 
* `agentic-lens archive --before 2026-06-01` archives all workspaces completed before June 1, 2026 into a `.tar.gz` archive and removes the archived folders from disk.
* Passing `--keep` generates the archive while leaving the original workspace folders intact on disk.
* Passing `--dry-run` prints which folders would be archived without creating files or deleting folders.
* In an interactive terminal without `--before`, the CLI presents a navigable selection list of completed workspaces with ticket IDs and completion dates.
* Archive creation is atomic; failed or interrupted runs never leave corrupted archives or prematurely delete folders.
* `go test -race ./...` and `go vet ./...` pass with zero warnings or errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
