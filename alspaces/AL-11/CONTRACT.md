# Contract: AL-11 Native Go Telemetry Extractor and Active Session Metric Generation

## 1. Problem Statement
Active session telemetry extraction in AgenticLens currently depends entirely on external execution via scripts/extract_agentic_metrics.py. This reliance on an external Python interpreter creates an operational friction point for the single-binary Go distribution goal, requires dual runtime environments, and complicates automated telemetry capture during SDLC iterations.

Ticket AL-11 delivers the vertical slice for telemetry independence: a native Go extraction engine residing within internal/adapters that queries the OpenCode SQLite database (~/.local/share/opencode/opencode.db) and executes Git repository analysis to compute multi-language code metrics. The Go binary (cmd/agentic-lens) exposes a dedicated generate subcommand capable of resolving active sessions, constructing atomic fact ledgers (matching the domain.AgenticFactLedger contract), projecting aggregations (domain.AgenticMetricsLedger), writing formatted JSON files, and idempotently updating Markdown summary tables.

## 2. Functional Core, Imperative Shell (FCIS) Invariants
All implementation must strictly satisfy the architectural boundaries defined in TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
Dependencies flow strictly inward across layers:
1. CLI Subcommand Layer (cmd/agentic-lens/main.go) handles CLI argument parsing, top-level flags, process exit codes, and wires adapters to domain orchestrations.
2. Imperative Adapters Layer (internal/adapters/opencode/, internal/adapters/git/, internal/adapters/):
   - internal/adapters/opencode/ handles SQLite database connection, SQL queries, JSON payload extraction from database records, and session resolution.
   - internal/adapters/git/ handles Git command execution, branch and revision resolution, numstat line parsing, file classification, and workspace iteration discovery.
   - internal/adapters/repository.go and internal/adapters/schemas.go handle disk persistence, filesystem scanning, and JSON serialization.
   - internal/adapters/formatters.go handles Markdown table rendering and idempotent summary updates.
3. Functional Core Domain Layer (internal/domain/):
   - Pure domain models (internal/domain/models.go), value objects, and typed string enums (internal/domain/enums.go).
   - Pure mathematical rollups and projections (internal/domain/aggregations.go, internal/domain/constants.go).
   - Absolute isolation: zero imports of database/sql, os, net/http, or outer adapter packages. Zero system clocks (time.Now) or disk I/O in domain logic.

### Pure Domain Core Isolation (internal/domain/)
1. The domain models InvocationTokens, InvocationFact, ToolFact, CodeFacts, AgenticFactLedger, and AgenticMetricsLedger remain immutable value structures.
2. The domain function DeriveMetricsLedger in internal/domain/aggregations.go remains the sole authority for deriving workflow efficiency, rollups, thread execution trees, role metrics, model metrics, and totals from atomic fact ledgers.
3. Domain derivations defend against division by zero (such as zero tokens, zero duration, or zero iterations).

### Imperative Boundary Shell (internal/adapters/ and cmd/)
1. Pure Go SQLite Driver: The SQLite adapter must use modernc.org/sqlite to ensure static compilation and pure Go execution with CGO_ENABLED=0. CGO-dependent drivers (such as mattn/go-sqlite3) are strictly prohibited.
2. Read-Only Database Access: OpenCode SQLite connections must open in read-only mode (file:...?...&mode=ro) to avoid write locking or database schema collisions with concurrent OpenCode agent processes.
3. Safe Process Execution: Git command executions must avoid fragile shell pipelines, invoke git directly via os/exec, and gracefully handle empty outputs, missing remotes, or uncommitted worktrees.
4. Resilient Fallbacks: When database access fails or a ticket session cannot be located, existing ledgers on disk must be loaded to avoid data loss.
5. Strict Zero Emojis Mandate: No emojis in log messages, CLI output, generated JSON files, summary tables, or code comments.
6. Strong Typing: Zero any or empty interface across all Go packages. Concrete structs and typed errors only.

## 3. Scope Boundaries: Go Backend and Delimiters

### In Scope for AL-11
1. Pure Go SQLite Extraction Adapter (internal/adapters/opencode/):
   - Package establishing connection to ~/.local/share/opencode/opencode.db (or configured/flagged database path).
   - Root session resolution implementing prioritized matching:
      - Priority 1: Exact Nonce Matching against part table data.
      - Priority 2: Explicit session ID passed via command flag.
      - Priority 3: Ticket ID matching against session table title.
      - Priority 4: Most recent session message fallback when no ticket or nonce is supplied.
   - Extracting session hierarchies (root and child subagent sessions).
   - Extracting message assistant tokens (input, output, cache read, reasoning, peak context), durations, and financial costs.
   - Extracting tool executions and failure statuses from part records.
   - Counting context compactions from part records.
2. Multi-Language Git Adapter (internal/adapters/git/):
   - Git user identification from git config user.email or user.name.
   - Robust revision resolution: active branch diff against merge-base with origin/main, local commits in git log matching ticket ID, remote feature branch inspection, and local commit fallback.
   - Hybrid file classification: loading optional code_metrics pattern configurations from .agenticlens/config.json, with universal multi-language defaults (Go, TypeScript, Python, web assets, Makefiles) distinguishing production code from test code.
   - Exclusion of documentation (*.md), manifests (*.json, *.lock, go.sum, package-lock.json), and workspace state.
   - Ticket title inference from workspace SUMMARY.md, CONTRACT.md, or git log commit subjects.
   - Ticket type classification (feature, bug, chore, refactor, documentation) from branch prefix or commit convention.
   - SDLC iteration counting by inspecting workspaces/<ticket_id>/iterations/ directory.
3. CLI Command Engine (cmd/agentic-lens/main.go):
   - Subcommand generate supporting flags: --ticket-id, --nonce, --session-id, --ticket-title, --ticket-type, --started-at, --completed-at, --json-output, --append, and --db-path.
   - Orchestrating existing ledger loading, database extraction, git analysis, domain projection, JSON serialization, and Markdown table updating.
   - Preserving existing CLI behaviors (server, scan, list).
4. Automated Table-Driven Test Suite:
   - In-memory SQLite fixture tests for OpenCode schema querying and session resolution.
   - Git numstat parsing and file classification unit tests.

### Out of Scope for AL-11
1. Batch backfill command across all historical workspaces (governed by AL-12).
2. Decommissioning and removing scripts/extract_agentic_metrics.py from repository (governed by AL-5).
3. Chronological event stream extraction and schema version 2 (governed by AL-6).
4. TypeScript frontend changes or view extensions (governed by AL-4).

## 4. Architectural Contracts and Data Structures

### SQLite Schema Mapping (internal/adapters/opencode/)
1. Table session:
   - Columns: id (text primary key), parent_id (nullable text), agent (nullable text), model (nullable text), title (nullable text), time_created (integer timestamp in milliseconds), time_updated (integer timestamp in milliseconds).
2. Table message:
   - Columns: id (text primary key), session_id (text foreign key), time_created (integer timestamp in milliseconds), data (json text containing modelID, role, time with created and completed, tokens with input, output, cache.read, reasoning, and cost).
3. Table part:
   - Columns: id (text primary key), session_id (text foreign key), message_id (text foreign key), time_created (integer timestamp in milliseconds), data (json text containing type e.g. compaction or tool, tool name, and state with status and output).

### Session Model Accumulator and Role Deduction
1. Role Deduction:
   - If parent_id is null, or if agent is build or orchestrator, role is assigned as orchestrator.
   - Otherwise, role is assigned as the session agent string (e.g. planner, coder, tester, reviewer, general).
2. Token Accumulation:
   - Invocations are grouped by session and model. For each assistant message, input, output, cache read, and reasoning tokens are summed. Peak context is tracked as the maximum of input plus cache read tokens.
   - Active duration is accumulated as the sum of message completed minus created timestamps.
   - Compaction counts for the session are attached to its primary invocation fact.

### Git Adapter and Hybrid Classification (internal/adapters/git/)
1. Numstat Parsing: Converts raw git diff --numstat output into structured tuples of added lines, deleted lines, and relative file path.
2. File Classification Rules:
   - Test files: matches *_test.go, *.test.ts, *.test.tsx, *.spec.ts, *.spec.tsx, test_*.py, *_test.py, or resides within test/, tests/, or __tests__/.
   - Production files: extensions .go, .ts, .tsx, .py, .js, .jsx, .html, .css, or exact name Makefile (excluding matches identified as test files).
   - Ignored files: extensions .md, .json, .lock, go.sum, package-lock.json, or paths residing under workspaces/.
   - Custom overrides: if code_metrics is present in .agenticlens/config.json, glob patterns specified in include, exclude, test_patterns, and prod_patterns take precedence.

### Output Serialization Contract
The generated JSON written to --json-output must match the existing AgenticFactLedger schema:
- Top-level properties: ticket_id, ticket_title, ticket_type, started_at, completed_at, updated_at, code, invocations, tools.
- When passing --append, the table generated by internal/adapters/formatters.go FormatMarkdownMetricsTable must be idempotently written into the target Markdown summary.

## 5. Acceptance Criteria
1. Running agentic-lens generate --ticket-id AL-11 --json-output workspaces/AL-11/agentic_metrics.json successfully queries OpenCode SQLite and Git, producing a valid ledger with zero Python dependencies.
2. Extracted metrics (token counts, financial costs, invocations, tools, and code lines added/deleted) exhibit parity with values extracted by the legacy Python script on identical session fixtures.
3. Passing --append workspaces/AL-11/SUMMARY.md updates or appends the Workflow Efficiency markdown table idempotently.
4. The Go binary compiles with CGO_ENABLED=0 and passes go test -race ./... and go vet ./... cleanly.
5. All new packages contain zero any types, zero empty interfaces, and zero emojis.
