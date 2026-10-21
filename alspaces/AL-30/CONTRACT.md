# Contract: AL-30 Fail Telemetry Generation on Unresolved Root Session

## 1. Problem Statement

When `agentic-lens generate` executes, if the telemetry engine fails to resolve a root session from the OpenCode SQLite database (via nonce, explicit session ID, ticket title matching, or workspace path) and no cached facts exist in an existing ledger on disk, it currently completes with exit code 0. It records zero invocations, zero agent duration, zero token metrics, and zero tools.

Silently generating empty records corrupts the telemetry ledger with misleading data and conceals session resolution failures. Having zero telemetry defeats the purpose of AgenticLens. Telemetry generation must not silently succeed when no root session can be identified.

Ticket AL-30 enforces strict root session resolution in `agentic-lens generate`. If neither the live database nor existing cached ledger facts provide identified root thread IDs or invocations, the CLI aborts execution with exit code 1, prints an actionable error message to standard error, and strictly prevents creating, modifying, or overwriting output files.

To maximize automatic session identification before failing, the resolver integrates workspace path fallback matching (`workspaces/<ticket>/`) when session title lookup yields no match. An explicit escape hatch flag (`--allow-empty`, default `false`) is added for synthetic unit tests where a live database is intentionally bypassed.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes across the Go backend must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. CLI Application Layer (`backend/cmd/agentic-lens/main.go`):
   - Entrypoint and command parsing for `runGenerate`.
   - Parses the new `--allow-empty` flag.
   - Enforces the root session resolution precondition before invoking file writer adapters.
   - Writes the actionable error message to `os.Stderr` and calls `os.Exit(1)`.
   - Contains zero database queries and zero mathematical rollup derivations.
2. Imperative Adapters Layer (`backend/internal/adapters/opencode/`):
   - Database boundary adapter.
   - `backend/internal/adapters/opencode/resolver.go` encapsulates SQL queries against OpenCode SQLite tables (`session`, `message`, `part`).
   - Enhances `ResolveActiveRootSession` with automatic fallback to `ResolveSessionByWorkspacePath(db, cleanTicket, 0, 0)` under Priority 3 when ticket title lookup returns `sql.ErrNoRows`.
   - Keeps data retrieval isolated from CLI execution policies and file writing side effects.
3. Functional Core Domain Layer (`backend/internal/domain/`):
   - Pure domain models (`AgenticFactLedger`, `AgenticMetricsLedger`, `InvocationFact`).
   - Pure mathematical aggregations, context window health derivations, and anomaly rules.
   - Pure functional core isolation: zero disk I/O, zero network, zero database access, and zero system clocks.
   - Domain logic is untouched by this ticket; boundary validation occurs entirely in the imperative shell.
4. Presentation Layer (`web/`):
   - Out of scope for this backend CLI and boundary adapter ticket.
   - Zero frontend modifications required.

### Pure Domain Core Isolation
- Domain models and aggregation functions remain pure value objects.
- No database handles, flags, or CLI error formatting constructs are imported or referenced in `backend/internal/domain/`.

### Imperative Shell Boundary Rules
- CLI Error Format:
  When root session resolution fails, output to `os.Stderr` must match:
  `Error: unable to identify root session for ticket '<TICKET_ID>'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.`
- Non-Zero Exit Code:
  Resolution failure must call `os.Exit(1)`.
- Side Effect Prevention:
  On resolution failure, neither `adapters.SaveFactLedger` (`--json-output`) nor `adapters.UpdateSummaryFile` (`--append`) may be called. No files may be created or overwritten.
- Zero Any Mandate:
  Zero `any` or `interface{}` across Go code.
- Zero Emoji Mandate:
  Zero emojis across code, docstrings, error messages, terminal logs, tests, and documentation.

---

## 3. Scope Boundaries: Go Backend and Tests

### In Scope
1. CLI Flag and Validation Enforcement (`backend/cmd/agentic-lens/main.go`):
   - Register boolean flag `--allow-empty` (default `false`) on the `generate` flag set in `runGenerate`.
   - After database extraction, check if `len(knownThreadIDs) == 0 && len(invocations) == 0 && !*allowEmptyFlag`.
   - Print the standardized actionable error message to `os.Stderr` and exit with status 1.
   - Guard `SaveFactLedger` and `UpdateSummaryFile` behind this validation check so that failed runs never touch disk.
2. Boundary Adapter Fallback (`backend/internal/adapters/opencode/resolver.go`):
   - In `ResolveActiveRootSession`, update Priority 3 (ticket ID matching): when querying `session` by title returns `sql.ErrNoRows`, invoke `ResolveSessionByWorkspacePath(db, cleanTicket, 0, 0)` as an automatic fallback.
   - Update docstring for `ResolveActiveRootSession` documenting the fallback.
3. Resolver Unit Tests (`backend/internal/adapters/opencode/extractor_test.go`):
   - Add test case verifying `ResolveActiveRootSession` resolves the root session via workspace path when session title does not match ticket ID.
   - Add test case verifying `ResolveActiveRootSession` returns `nil, nil` when neither title nor workspace path matches.
4. CLI Integration and Failure Tests (`backend/cmd/agentic-lens/main_test.go`):
   - Add test verifying `generate --ticket-id NON_EXISTENT` exits with status 1, writes the expected error message to stderr, and writes no files to disk.
   - Add test verifying that an existing summary file is not modified when resolution fails.
   - Update existing synthetic unit tests in `main_test.go` that intentionally omit database sessions to supply `--allow-empty`.
   - Add test verifying `--allow-empty` permits generation without active sessions.
   - Verify that valid tickets with identifiable sessions continue to succeed.

### Out of Scope
1. Changes to TypeScript web frontend (`web/`).
2. Schema migrations or backfill engine logic (`internal/adapters/backfill/`).
3. Changes to OpenCode SQLite schema or table structures.
4. Modifications to HTTP server routes or streaming endpoints.

---

## 4. Architectural Specifications

### Resolution Precedence in `ResolveActiveRootSession`
1. Priority 1: Exact Nonce matching against `part` table data.
2. Priority 2: Explicit session ID passed directly as parameter.
3. Priority 3: Ticket ID matching against root session titles. If no row matches (`sql.ErrNoRows`), automatically fallback to searching root sessions whose part data contains `workspaces/<ticket>/` via `ResolveSessionByWorkspacePath(db, cleanTicket, 0, 0)`.
4. Priority 4: Fallback to the most recent message session parent or ID only when neither nonce nor ticket ID is supplied.

### CLI Execution and Guard Logic in `runGenerate`
1. Parse flags: include `--allow-empty` (bool, default `false`).
2. Load cached facts from `--json-output` if file exists. Populate `knownThreadIDs` and `invocations`.
3. Resolve root session via `opencode.ResolveActiveRootSession` if database is accessible.
4. Extract facts via `opencode.ExtractFacts` if any root sessions are known.
5. Session resolution check:
   - If `len(knownThreadIDs) == 0 && len(invocations) == 0 && !*allowEmptyFlag`:
     - Determine ticket identifier to display: `activeTicketID` if non-empty, otherwise `UNKNOWN`.
     - Write error to standard error: `Error: unable to identify root session for ticket '<TICKET_ID>'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.`
     - Terminate immediately with `os.Exit(1)`.
6. Proceed with code metrics, ledger derivation, and disk output only if check passes or `--allow-empty` is set.

---

## 5. Acceptance Criteria

1. `agentic-lens generate --ticket-id NON_EXISTENT` exits with status 1 and logs error to stderr matching `Error: unable to identify root session for ticket 'NON_EXISTENT'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.`.
2. No JSON ledger or markdown summary file is created, modified, or overwritten when root session identification fails.
3. Automatic fallback in `ResolveActiveRootSession` checks workspace path (`workspaces/<ticket>/`) before failing when ticket title does not match.
4. Supplying `--allow-empty` allows synthetic execution to proceed with exit code 0 when sessions are absent.
5. All synthetic unit tests in `main_test.go` pass without regression.
6. `make test-backend`, `go vet ./...`, and `make check` pass with zero errors.
7. Zero emojis across all code, tests, logs, and artifacts.
8. Zero `any` or empty `interface{}` across Go code.
