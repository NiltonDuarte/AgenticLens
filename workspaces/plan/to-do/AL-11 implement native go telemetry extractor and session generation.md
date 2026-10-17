h3. Title: AL-11: Implement Native Go Telemetry Extractor and Active Session Metric Generation

*TARGET*: AgenticLens (adapters | domain | cli | tests)

*CONTEXT*

Currently, active session telemetry extraction depends entirely on `scripts/extract_agentic_metrics.py`. As part of the single-binary Go rewrite and operational decoupling, the Go application (`cmd/agentic-lens`) must natively extract active session telemetry directly from the OpenCode SQLite database and git repository, writing the atomic facts ledger and updating markdown summary tables.

*PORTING REQUIREMENTS & PYTHON EXTRACTOR ISSUES TO FIX (FROM AL-10)*

Investigation of `scripts/extract_agentic_metrics.py` revealed several critical flaws that must be resolved in the native Go implementation (`internal/adapters/git`):
1. *Eliminate Hardcoded Path Filtering*: The Python script restricted `git numstat` parsing to path prefixes starting with `('src/', 'tests/', 'infra/', 'cli/')`, which filtered out all Go files in `internal/`, `cmd/`, and TypeScript files in `web/`, resulting in 0 lines added/deleted. The Go adapter must process all repository files without hardcoded directory exclusions.
2. *Multi-Language File Classification (Hybrid Configuration)*: The Python script only identified test files if they started with `tests/` and prod files if they started with `src/`. This misclassified Go `*_test.go` and TS `*.test.ts`/`*.test.tsx` as non-test or discarded them. The Go adapter must:
   - Support optional `code_metrics` configuration from `.agenticlens/config.json` (`include`, `exclude`, `test_patterns`, `prod_patterns`).
   - When omitted, apply universal multi-language defaults:
     - Test files: `*_test.go`, `*.test.ts`, `*.test.tsx`, `*.spec.ts`, `*.spec.tsx`, `test_*.py`, `*_test.py`, `test/**`, `tests/**`, `__tests__/**`.
     - Production files: `.go`, `.ts`, `.tsx`, `.py`, `.js`, `.jsx`, `.html`, `.css`, `Makefile` (excluding test files).
     - Excluded: `*.md`, `*.json`, `*.lock`, `go.sum`, `package-lock.json`, `workspaces/**`, `.agenticlens/**`, `dist/**`, `node_modules/**`, `.git/**`, `.opencode/**`.
3. *Robust Git Revision Range & Branch Resolution*: The Python script relied heavily on `origin/main...HEAD` queries, which fail when `origin/main` is unpushed, absent, or when running in detached HEAD or worktree environments. The Go adapter must implement a multi-stage fallback resolution strategy:
   - Determine base branch merge-base dynamically (`git merge-base HEAD main` or `origin/main`).
   - If active branch matches the ticket ID, diff against merge-base.
   - If not on ticket branch, scan commit history with `git log --all --grep=<ticket_id>` across local and remote commits.
   - If a ticket branch exists locally or remotely, diff its head against `main` merge-base.
   - Fall back to commit body search if commit subjects do not match.
4. *Division by Zero & Ratio Safety*: Ensure test-to-production ratio derivations handle zero production lines gracefully without panic or invalid floating-point values.
5. *Clean Idempotent Summary Updates*: Ensure updating Markdown summaries is idempotent, preserves existing sections, and maintains zero emojis across all output.

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
* Code facts accurately reflect multi-language code changes (Go, TypeScript, Python) with proper test vs prod breakdown and zero hardcoded directory exclusions.
* Hybrid pattern overrides in `.agenticlens/config.json` are respected, falling back to universal multi-language defaults when omitted.
* Robust git revision resolution correctly identifies changes across active branches, merge-bases, and commit searches even without remote `origin/main`.
* Passing `--append <path>` updates the Markdown summary table idempotently.
* Pure Go build succeeds with CGO_ENABLED=0 (`go test -race ./...`, `go vet ./...`).
* Zero `any` or empty `interface{}` across new Go packages.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* End-to-end extraction verified against active session.
