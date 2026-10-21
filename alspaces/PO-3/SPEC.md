# Feature Specification: Native Go Extraction, Backfill Engine, and Multi-Language Git Metrics

## Problem Statement
Currently, session metrics extraction relies entirely on a monolithic Python script (`scripts/extract_agentic_metrics.py`). This creates two critical operational issues:
1. **Coupled Generation and Backfill**: Active session metric extraction and historical ledger backfilling are conflated into a single execution path. When the telemetry schema changes or expands (such as adding chronological events or new aggregation fields), there is no dedicated, resilient Go CLI command to backfill raw JSON data across existing workspace ledgers without triggering unwanted session resolution logic.
2. **Defective Git Metrics Calculation**: The Python script currently reports zero lines added and deleted across all dev ticket workspaces (e.g. AL-1, AL-2). Root-cause analysis revealed multiple hardcoded paths and flawed git query assumptions:
   - Line 363 & Line 398 in `scripts/extract_agentic_metrics.py`: diffs are filtered with `r[2].startswith(("src/", "tests/", "infra/", "cli/"))`, discarding all files in Go/TypeScript directory layouts (`cmd/`, `internal/`, `web/`, `scripts/`).
   - Line 594-597: test vs production lines are classified strictly via `startswith("tests/")` and `startswith("src/")`, causing Go test files (`*_test.go`) and TS test files (`*.test.ts`, `*.spec.ts`) as well as Go production code (`internal/`, `cmd/`) to evaluate to 0.
   - Git revision resolution queries `origin/main` directly via `git log origin/main`, failing when commits exist on local branches or on local `main` prior to pushing upstream.

We need to fix the git calculation across Python and Go, provide a dedicated Go active session generator command (`agentic-lens generate`), and provide a dedicated Go historical backfill command (`agentic-lens backfill`) capable of recalculating metrics and applying future schema expansions idempotently.

## Target Personas & Use Cases
- **AI Systems Engineers & Agent Developers**: Need accurate code velocity metrics (lines added/deleted, test-to-production ratios) recorded in active ticket summaries to evaluate agent efficiency.
- **Platform Maintainers & Schema Authors**: Need an idempotent, batch backfill CLI command in Go to reprocess existing `agentic_metrics.json` ledgers whenever schema fields or calculation algorithms evolve.

## User Stories
1. As an agent developer, I want `scripts/extract_agentic_metrics.py` and the Go engine to accurately calculate git code metrics for Go and TypeScript repositories, so that my ticket telemetry reflects real changes delivered.
2. As a platform engineer, I want a dedicated `agentic-lens generate` CLI command in Go to extract the active session's telemetry from OpenCode SQLite DB and git without depending on Python.
3. As a platform maintainer, I want a dedicated `agentic-lens backfill` CLI command in Go to update existing `agentic_metrics.json` ledgers across workspaces when schemas evolve or calculations are corrected.

## Functional Requirements

### 1. Git Metrics Calculation Engine (Hybrid Model: Defaults + Config Overrides)
- **Convention-Over-Configuration Philosophy**:
  - Out of the box with zero configuration, the engine accurately classifies code files across Go, TypeScript, Python, and other common languages.
  - If `.agenticlens/config.json` defines a `code_metrics` object, explicit include, exclude, test, and production patterns override or supplement the defaults.
- **Default Multi-Language Path Classification**:
  - Test files: Match `*_test.go`, `*.test.ts`, `*.test.tsx`, `*.spec.ts`, `*.spec.tsx`, `test_*.py`, `*_test.py`, `test/**`, `tests/**`, `__tests__/**`.
  - Production code: Match `.go`, `.ts`, `.tsx`, `.py`, `.js`, `.jsx`, `.html`, `.css`, and `Makefile` (excluding test files).
  - Excluded paths: Ignore `workspaces/**`, `.opencode/**`, `docs/**`, `*.md`, `*.json`, `*.lock`, `go.sum`, `package-lock.json`, and dotfiles.
- **Configurable Overrides via `.agenticlens/config.json`**:
  - Support optional `code_metrics` schema in `.agenticlens/config.json`:
    ```json
    {
      "code_metrics": {
        "include": ["cmd/**", "internal/**", "web/**", "scripts/**"],
        "exclude": ["workspaces/**", "web/dist/**", "*.md", "*.json", "go.sum"],
        "test_patterns": ["**/*_test.go", "**/*.test.ts", "**/*.spec.ts", "**/tests/**"],
        "prod_patterns": ["internal/**/*.go", "cmd/**/*.go", "web/src/**/*.{ts,tsx}", "scripts/**/*.py"]
      }
    }
    ```
- **Git Revision Resolution**:
  - Active feature branch: Diff against merge base with `main` or `origin/main`, plus any uncommitted staged/unstaged modifications.
  - Merged ticket commits: Scan git log across `--all` for `[clean_ticket]` or `(<ticket_id>)` and aggregate numstat across matching commits.
  - Standalone local commits: Fallback to local commit ranges matching ticket identifier in subject or body.
- **Immediate Python Fix**:
  - Patch `scripts/extract_agentic_metrics.py` to read `.agenticlens/config.json` when present, fall back to universal defaults, and remove the hardcoded `src/`/`tests/` checks.
  - Run backfill across all existing ticket workspaces (`workspaces/AL-1`, `workspaces/AL-2`, `workspaces/GEN-1`, `workspaces/GEN-2`, `workspaces/PO-1`, `workspaces/PO-2`) to restore code metrics.

### 2. Native Go Session Generation Command (`agentic-lens generate`)
- **CLI Interface**:
  ```bash
  agentic-lens generate --ticket-id <TICKET_ID> [--nonce <NONCE>] [--session-id <ID>] [--json-output <PATH>] [--append <SUMMARY_PATH>]
  ```
- **OpenCode DB Adapter (`internal/adapters/sqlite/` or `opencode/`)**:
  - Query OpenCode SQLite DB (`~/.local/share/opencode/opencode.db`) using pure Go SQLite driver (`modernc.org/sqlite`).
  - Resolve active session using nonce, ticket ID, or latest active session.
  - Extract session durations, role classifications, token counters (input, output, cache read/write, reasoning), compactions, and tool execution facts.
- **Git Status Adapter (`internal/adapters/git/`)**:
  - Execute git commands to compute lines added, lines deleted, test lines added/deleted, prod lines added/deleted, and files changed.
- **Output & Summary**:
  - Serialize to `agentic_metrics.json` in accordance with domain schema.
  - Idempotently update markdown table in `SUMMARY.md` when `--append` is supplied.

### 3. Native Go Telemetry Backfill Command (`agentic-lens backfill`)
- **CLI Interface**:
  ```bash
  agentic-lens backfill [--workspaces <DIR>] [--ticket-id <TICKET_ID>] [--dry-run] [--recompute-git] [--recompute-db]
  ```
- **Execution Workflow**:
  - Discover existing `agentic_metrics.json` ledgers within target workspaces directory.
  - `--recompute-git`: Recalculates code facts using the Go git adapter and updates the ledger.
  - `--recompute-db`: Re-queries OpenCode SQLite DB for referenced sessions to backfill new schema fields or repair missing facts.
  - `--dry-run`: Reports planned modifications without writing to disk.
  - Idempotently write updated JSON with formatted indentation.
  - Optionally refresh `SUMMARY.md` tables in modified workspaces.

## Non-Functional & Architectural Invariants
- **Functional Core, Imperative Shell (FCIS)**:
  - Pure metrics calculations, aggregations, and classification heuristics remain in `internal/domain`.
  - SQLite querying, git execution, and filesystem operations reside exclusively in `internal/adapters`.
- **Pure Go SQLite Dependency**:
  - Use `modernc.org/sqlite` to maintain CGO-free, cross-compilable Go binary distributions.
- **Strict Typing**:
  - Zero `any` or empty `interface{}` across all new Go structs and interfaces.
- **Zero Emojis**:
  - Strictly no emojis in code, CLI logs, terminal outputs, JSON payloads, or generated Markdown tables.
- **Parity & Backward Compatibility**:
  - Go extraction output must match legacy Python schema structure (`v1` and future `v2`), producing identical aggregate numbers for tokens, cost, and code metrics.

## Decomposition & Delivery Roadmap
- [ ] Ticket AL-10: `AL-10: Fix Git Metrics Calculation in Python Extractor and Backfill Existing Workspace Ledgers` (Ready / Immediate Unblocked Bugfix)
- [ ] Ticket AL-11: `AL-11: Implement Native Go Telemetry Extractor and Active Session Metric Generation` (Blocked by AL-10, AL-2)
- [ ] Ticket AL-12: `AL-12: Implement Go Telemetry Backfill Engine and Schema Migration Command` (Blocked by AL-11)
