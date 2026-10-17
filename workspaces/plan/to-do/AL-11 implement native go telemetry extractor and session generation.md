h3. Title: AL-11: Implement Native Go Telemetry Extractor and Active Session Metric Generation

*TARGET*: AgenticLens (adapters | domain | cli | tests)

*CONTEXT*

Currently, active session telemetry extraction depends entirely on `scripts/extract_agentic_metrics.py`. As part of the single-binary Go rewrite and operational decoupling, the Go application (`cmd/agentic-lens`) must natively extract active session telemetry directly from the OpenCode SQLite database and git repository, writing the atomic facts ledger and updating markdown summary tables.

*GOAL*

Implement a native Go extraction engine in `internal/adapters` and expose the `agentic-lens generate` CLI subcommand to resolve sessions, query OpenCode SQLite, compute multi-language git code metrics, and output compliant `agentic_metrics.json` ledgers.

*VERTICAL SLICE SCOPE*

* Domain Core: Ensure domain models in `internal/domain/models.go` support extraction input facts and code metrics.
* SQLite Adapter: Create `internal/adapters/opencode/` using pure Go SQLite driver (`modernc.org/sqlite`) to query `session`, `part`, and `message` tables from `~/.local/share/opencode/opencode.db`. Extract invocations, roles, tokens, costs, durations, and tool executions.
* Git Adapter: Create `internal/adapters/git/` to compute lines added/deleted, test/prod breakdown, and files changed with hybrid classification: load optional `code_metrics` patterns from `.agenticlens/config.json`, falling back to multi-language file extension and directory defaults.
* CLI Engine: Add `generate` subcommand to `cmd/agentic-lens/main.go` accepting `--ticket-id`, `--nonce`, `--session-id`, `--json-output`, and `--append`.
* Quality: Table-driven tests for OpenCode DB extraction using a test SQLite fixture and git parser tests. Zero CGO dependencies.

*ACCEPTANCE CRITERIA*

* `agentic-lens generate --ticket-id <ID> --json-output workspaces/<ID>/agentic_metrics.json` extracts and writes a valid ledger without Python.
* Generated JSON matches existing schema specifications with identical token, cost, and code metrics.
* Passing `--append <path>` updates the Markdown summary table idempotently.
* Pure Go build succeeds with CGO_ENABLED=0 (`go test -race ./...`, `go vet ./...`).
* Zero `any` or empty `interface{}` across new Go packages.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* End-to-end extraction verified against active session.
