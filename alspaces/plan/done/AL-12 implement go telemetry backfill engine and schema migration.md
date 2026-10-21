h3. Title: AL-12: Implement Go Telemetry Backfill Engine and Schema Migration Command

*TARGET*: AgenticLens (adapters | domain | cli | tests)

*CONTEXT*

When telemetry schemas evolve (e.g. adding chronological events in v2, updating anomaly indicators, or correcting code metric formulas), existing `agentic_metrics.json` files on disk must be upgraded and recomputed without manual scripts. A dedicated, robust backfill command in the Go application is required to iterate over workspace directories, recalculate facts, migrate schemas, and persist updated ledgers idempotently.

*GOAL*

Deliver a native Go backfill engine in `internal/adapters` and expose the `agentic-lens backfill` CLI subcommand to scan, recompute git metrics, re-query DB traces for linked sessions, migrate schemas, and update `agentic_metrics.json` and `SUMMARY.md` files.

*VERTICAL SLICE SCOPE*

* Adapter Core: Create `internal/adapters/backfill/` or `internal/adapters/backfill.go` providing:
  - Workspace directory scanner discovering all `agentic_metrics.json` ledgers.
  - Recomputation engine for Git code metrics.
  - Session trace re-querying and enrichment against OpenCode SQLite DB.
  - Atomic JSON rewrite with schema preservation and validation.
* CLI Engine: Add `backfill` subcommand to `cmd/agentic-lens/main.go` accepting:
  - `--workspaces <dir>`: Target workspaces directory (defaults to configured workspaces).
  - `--ticket-id <id>`: Optional single-ticket filter.
  - `--recompute-git`: Explicitly recompute git code facts.
  - `--recompute-db`: Explicitly re-extract invocations and tools from OpenCode DB.
  - `--dry-run`: Preview changes without modifying files.
  - `--update-summaries`: Automatically regenerate markdown summary tables.
* Quality: Unit and integration tests covering multi-ticket backfill, corrupted file handling, and dry-run guarantees.

*ACCEPTANCE CRITERIA*

* Running `agentic-lens backfill --workspaces workspaces/ --recompute-git` successfully refreshes code metrics across all workspaces.
* Running `agentic-lens backfill --ticket-id AL-1 --dry-run` displays planned field updates without writing changes.
* Corrupted or unparseable JSON files produce clear warning logs without aborting the batch process.
* All updated `agentic_metrics.json` files pass schema validation and can be read by `agentic-lens server`.
* Go checks pass cleanly (`go vet ./...`, `go test -race ./...`).

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Verified end-to-end backfill across test workspaces.
