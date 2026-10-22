# Verification Plan: AL-33 Iteration 02 Filesystem Integrity Audit and Verification

## 1. Goal and Iteration Scope

Iteration 02 addresses user feedback regarding disk space pressure and potential loss during the node_modules cleanup incident. The scope is strictly verification and audit. Tester and reviewer subagents must validate that no production code, domain logic, boundary adapters, UI components, or test suites were lost, degraded, or silently skipped.

The target is an exhaustive audit across four dimensions:
1. Complete Go backend test execution with race detection enabled across all packages.
2. Complete TypeScript test execution across all 21 test files (423 tests) with zero skips after restoring web/node_modules.
3. Multi-platform binary compilation and embedded asset integrity check.
4. Formal verification of CONTRACT.md invariants: in-memory transient hydration without disk writes, path traversal prevention, zero any, and zero emojis.

---

## 2. Grounded Inventory and Integrity Audit Findings

### Layer Integrity Assessment
- Go Functional Core (backend/internal/domain):
  - models.go: WorkspaceEventsResponse struct and EventFact diagnostic fields (PromptContent, ReasoningContent, ToolInput, ToolOutput) intact.
  - enums.go: DataSourceLocalSQLite and DataSourceCommittedLedger constants intact.
  - events.go: PruneEventsForPersistence explicitly sets all diagnostic payload pointers to nil.
  - events_test.go: Serialization and sanitization tests intact.
- Go Imperative Shell (backend/internal/adapters):
  - adapters/live/sqlite.go: Non-locking reader pragmas (query_only = ON, busy_timeout = 5000, read_uncommitted = ON) and methods FetchSessionTraceMulti and ResolveChildSessionIDs intact.
  - adapters/live/service.go: HydrateWorkspaceEvents trace assembly with fallback semantics intact.
  - adapters/live/sqlite_test.go: Multi-session queries, concurrent write isolation, and degradation tests intact.
  - adapters/http/dto.go: Generic ResponsePayload constraint interface and WriteEventsResponse intact.
  - adapters/http/handlers.go: WorkspaceEventsHandler with WorkspaceEventHydrator and path traversal checks intact.
  - adapters/http/server.go: Router wiring intact.
  - adapters/http/server_test.go: HTTP integration tests for dual-plane events, 404, and traversal intact.
- Web Presentation (web/src):
  - types/telemetry.ts and domain/types.ts: WorkspaceEventsResponse and DataSource interfaces intact.
  - api/client.ts and api/client.test.ts: Dual-plane response parsing and legacy fallback normalization intact.
  - components/TraceWaterfall.tsx and TraceWaterfall.test.tsx: Offline indicator banner and rich interactive payload drawers intact.
  - components/AnomalyDrawer.tsx and AnomalyDrawer.test.tsx: Diagnostic payload previews intact.
  - views/TicketDeepDive.tsx and TicketDeepDive.test.tsx: Hydration state propagation intact.
- Pre-compiled Embedded Assets:
  - backend/internal/adapters/web/dist: index.html, index-3eg4qbze.css, and index-CMhIx-jd.js intact.
  - web/dist: index.html, index-3eg4qbze.css, and index-CMhIx-jd.js intact.
- Filesystem State:
  - web/node_modules was cleaned up during disk reclamation. It must be reinstalled via npm install before executing frontend test commands.

---

## 3. Verification Tasks for Tester Subagent

### Phase 1: Environment Readiness
- Task 1.1: Restore web dependencies by executing npm install in the web directory.
- Task 1.2: Verify git status reports a clean working tree with zero untracked scratch files.

### Phase 2: Go Backend Test Suite and Race Condition Audit
- Task 2.1: Run all backend tests with race detection from the backend directory using go test -race -v ./...
- Task 2.2: Confirm all 11 packages pass with zero failures.
- Task 2.3: Verify domain/events_test.go exercises:
  - WorkspaceEventsResponse JSON round-trip with is_hydrated set to true and false.
  - Adversarial Unicode characters, null payloads, and empty event slices.
  - PruneEventsForPersistence clearing PromptContent, ReasoningContent, ToolInput, ToolOutput, and ErrorMessage on retained events.
- Task 2.4: Verify adapters/live/sqlite_test.go exercises:
  - Multi-session trace ordering across parent and child sessions.
  - Active concurrent SQLite write transactions during hydration queries under WAL mode.
  - Graceful fallback to DataSourceCommittedLedger when the database is absent, corrupted, or locked.
- Task 2.5: Verify adapters/http/server_test.go exercises:
  - Endpoint GET /api/workspaces/{ticketId}/events returning dual-plane responses.
  - Status 404 Not Found for non-existent ticket workspaces.
  - Status 400 Bad Request for path traversal attacks including dot-dot slashes, backslashes, and encoded variations.
- Task 2.6: Audit test sources for zero occurrences of t.Skip.

### Phase 3: Frontend TypeScript Typecheck and Vitest Suite Execution
- Task 3.1: Execute npm run typecheck in web to confirm zero type errors and zero any violations under strict mode.
- Task 3.2: Execute npm test in web to run the complete Vitest test suite.
- Task 3.3: Confirm all 21 test files and all 423 tests pass cleanly.
- Task 3.4: Verify specific AL-33 test assertions:
  - TraceWaterfall.test.tsx validates offline banner rendering when isHydrated is false.
  - TraceWaterfall.test.tsx validates payload preview drawer rendering when isHydrated is true.
  - api/client.test.ts validates legacy array parsing and normalization into WorkspaceEventsResponse.
  - TicketDeepDive.test.tsx validates hydration prop forwarding.
- Task 3.5: Audit frontend test sources for zero occurrences of it.skip, test.skip, or describe.skip.

### Phase 4: Single Binary Compilation and Embedded Asset Verification
- Task 4.1: Execute make build from project root to compile multi-platform binaries into bin/.
- Task 4.2: Verify host binary execution with ./bin/agentic-lens version --json.
- Task 4.3: Verify backend/internal/adapters/web/embed_test.go passes, ensuring embedded files are accessible from Go fs.FS.

### Phase 5: Contract Compliance and Invariant Audit
- Task 5.1: Zero Disk Persistence: Verify no files are written or modified in workspaces/ during SQLite trace hydration.
- Task 5.2: Path Traversal Protection: Verify the HTTP server rejects directory traversal attempts before filesystem access.
- Task 5.3: Type Discipline: Confirm zero any or interface{} keywords across all Go packages and TypeScript modules.
- Task 5.4: Design System: Confirm zero emojis across all code, tests, UI strings, and documentation files.

---

## 4. Audit Checkpoints for Reviewer Subagent

1. Verification Report Integrity:
   - Confirm backend test logs document 11 passing packages with race detection enabled and zero skips.
   - Confirm frontend test logs document 21 passing test files with 423 passed tests and zero skips.
2. Architecture Compliance:
   - Confirm backend/internal/domain imports zero I/O packages (no database/sql, os, net/http).
   - Confirm backend/internal/adapters isolates all SQLite operations, non-blocking pragmas, and HTTP serialization.
   - Confirm web/src contains zero any and zero emojis.
3. Incident Closure:
   - Confirm no work, tests, or documentation artifacts from AL-33 were lost or skipped due to the filesystem incident.
   - Produce Iteration 02 REVIEW.md detailing verification evidence and final approval.
