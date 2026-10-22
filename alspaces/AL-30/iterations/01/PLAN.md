# Implementation Plan: AL-30 Fail Telemetry Generation on Unresolved Root Session

## 1. Goal and Iteration Scope

Enforce strict root session resolution in the AgenticLens CLI generate subcommand. The command must terminate with non-zero exit code 1, print an actionable error message to standard error, and prevent creating or modifying output files whenever no root session can be identified and no invocations are loaded.

In addition, integrate automatic workspace path fallback matching in the OpenCode boundary resolver to maximize session identification prior to failure, and add the --allow-empty flag to preserve isolated synthetic testing capabilities.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation maps to the following concrete files and packages:

### Boundary Adapters Layer (backend/internal/adapters/opencode/)
- backend/internal/adapters/opencode/resolver.go: Update ResolveActiveRootSession to call ResolveSessionByWorkspacePath as an automatic fallback under Priority 3 when ticket title lookup finds no matching rows.
- backend/internal/adapters/opencode/extractor_test.go: Add unit tests for ResolveActiveRootSession verifying workspace path fallback and unresolved session behavior.

### CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go: Add --allow-empty flag in runGenerate. Add guard after session extraction to print actionable error to os.Stderr and exit with status 1 when thread IDs and invocations are zero and --allow-empty is false.
- backend/cmd/agentic-lens/main_test.go: Update existing synthetic tests needing --allow-empty. Add new CLI integration tests asserting non-zero exit code, exact stderr message, and prevention of output file writes on failure.

---

## 3. Phased Implementation Tasks

### Phase 1: Boundary Adapter Workspace Path Fallback Integration
- Task 1.1: Update ResolveActiveRootSession in backend/internal/adapters/opencode/resolver.go:
  - Locate Priority 3 ticket ID title query block.
  - When query scan returns sql.ErrNoRows, invoke ResolveSessionByWorkspacePath passing database handle, cleanTicket, 0 for minTimeMs, and 0 for maxTimeMs.
  - If ResolveSessionByWorkspacePath returns a non-nil, non-empty session ID pointer and nil error, return that session ID pointer and nil error.
  - If ResolveSessionByWorkspacePath returns an error, wrap with context "resolving session by workspace path fallback: %w" and return nil and the wrapped error.
  - If neither title nor workspace path matches, allow execution to proceed to Priority 4 or return nil, nil.
  - Update docstring for ResolveActiveRootSession to document the workspace path fallback behavior under Priority 3.
- Task 1.2: Add Resolver Unit Tests in backend/internal/adapters/opencode/extractor_test.go:
  - Add test case in TestResolveActiveRootSession where session title does not contain ticket identifier, but part data references workspaces/<ticket>/. Assert root session ID is resolved via fallback.
  - Add test case where neither title nor part data references ticket identifier. Assert resolver returns nil session pointer and nil error.
  - Run package tests with race detector to confirm passing.

### Phase 2: CLI Resolution Enforcement and Flag Handling
- Task 2.1: Register --allow-empty Flag in backend/cmd/agentic-lens/main.go:
  - In runGenerate, declare allowEmptyFlag using genFlags.Bool("allow-empty", false, "Allow empty metrics generation even if no root session or invocations are resolved (test escape hatch)").
- Task 2.2: Add Enforcement Check in runGenerate:
  - Directly after the database extraction block (following opencode.ExtractFacts), evaluate len(knownThreadIDs) == 0 && len(invocations) == 0 && !*allowEmptyFlag.
  - When condition evaluates to true:
    - Determine display ticket string: activeTicketID if non-empty, otherwise UNKNOWN.
    - Format error message: Error: unable to identify root session for ticket '<TICKET_ID>'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.
    - Write error message to os.Stderr.
    - Terminate execution immediately via os.Exit(1).
  - Ensure this check occurs strictly before git.ComputeCodeFacts, domain.DeriveMetricsLedger, adapters.SaveFactLedger, and adapters.UpdateSummaryFile.

### Phase 3: Update Existing Synthetic CLI Tests
- Task 3.1: Update TestCLIAnomaliesCommand in backend/cmd/agentic-lens/main_test.go:
  - In subtest "generate subcommand with --anomalies flag on ticket with anomalies", add --allow-empty flag to arguments so synthetic seed ledger without invocations can execute without database.
  - In subtest "generate subcommand with --anomalies flag on clean ticket", add --allow-empty flag to arguments.
- Task 3.2: Update TestCLIGenerate_AdversarialEdgeCases in backend/cmd/agentic-lens/main_test.go:
  - In subtest "generate with corrupted existing JSON output preserves execution and safely recovers", add --allow-empty flag to arguments to test corrupted JSON recovery without requiring live database session.
  - In subtest "generate with non-existent database falls back gracefully without crash":
    - Refactor to test the escape hatch: assert that passing --allow-empty allows empty generation without error.

### Phase 4: CLI Integration Tests for AL-30 Failure Invariants
- Task 4.1: Add Test for CLI Failure on Unresolved Session in backend/cmd/agentic-lens/main_test.go:
  - Define test "generate fails with status 1 and logs error when root session cannot be identified".
  - Execute agentic-lens generate --ticket-id NON_EXISTENT --db-path <temp_empty_db> --json-output <temp_json> --append <temp_md>.
  - Assert process exits with non-zero exit code (exit status 1).
  - Assert standard error contains: Error: unable to identify root session for ticket 'NON_EXISTENT'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.
  - Assert target JSON output file does not exist on disk.
  - Assert target markdown append file does not exist on disk.
- Task 4.2: Add Test for Summary File Preservation on Failure:
  - Create pre-existing markdown file with known initial content.
  - Run generate targeting non-existent ticket without --allow-empty.
  - Assert command exits with status 1.
  - Read markdown file and assert file content remains strictly identical to initial content.
- Task 4.3: Add Test for Successful Session Resolution:
  - Seed SQLite database with root session title matching ticket identifier and message/parts.
  - Run generate targeting that ticket identifier without --allow-empty.
  - Assert command exits with status 0, creates valid JSON output, and updates markdown file.

### Phase 5: Verification and Static Compliance
- Task 5.1: Run go vet ./... from backend directory.
- Task 5.2: Run go test -race ./... from backend directory.
- Task 5.3: Run make check from repository root.
- Task 5.4: Verify zero emojis across all modified files and new test assertions.
- Task 5.5: Verify zero any or empty interface{} across Go backend.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Boundary Adapter Workspace Path Fallback) establishes automatic fallback session detection and unit test coverage.
2. Phase 2 (CLI Enforcement and Flag Handling) introduces --allow-empty and the hard exit gate in runGenerate.
3. Phase 3 (Update Existing Synthetic CLI Tests) adapts existing test cases that bypass live databases to use --allow-empty.
4. Phase 4 (CLI Integration Tests for Failure Invariants) proves acceptance criteria: non-zero exit code, exact stderr message, and output file prevention.
5. Phase 5 (Verification and Compliance) runs automated quality gates and asserts zero any and zero emojis.

---

## 5. Verification Plan

### Test Commands
- Adapter tests: cd backend && go test -v -race -run TestResolve ./internal/adapters/opencode/...
- CLI tests: cd backend && go test -v -race -run TestCLI ./cmd/agentic-lens/...
- Static analysis: cd backend && go vet ./...
- Complete test suite: make test-backend
- Quality gate: make check

### Manual Verification Scenarios
- Run without session: ./bin/agentic-lens generate --ticket-id NON_EXISTENT --json-output test.json -> must fail with status 1, log error to stderr, and create no test.json.
- Run with escape hatch: ./bin/agentic-lens generate --ticket-id NON_EXISTENT --allow-empty --json-output test.json -> must succeed with status 0 and write test.json.
- Run with valid workspace path in database: verify fallback resolves root session without requiring title match.
