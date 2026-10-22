# Iteration 02 Plan: Codebase, Test Suite, and Skills Refactoring (AL-16)

## Iteration Objective
Execute the second refactoring delta for ticket AL-16. This iteration resolves six specific engineering debts:
1. Eliminate full filesystem directory traversal during single-ticket event queries in the HTTP adapter by delegating directly to targeted workspace loading.
2. Relocate event ordering logic from the HTTP handler into a pure functional core domain function with unit tests.
3. Eliminate the triplicated invocation accumulation loop across ComputeTotals, ComputeRoleMetrics, and ComputeModelMetrics in the functional core using a pure helper function.
4. Encapsulate JSON marshaling, formatting, and filesystem persistence for fact ledgers and metrics ledgers inside repository adapters, removing redundant directory creation and write operations from the CLI entrypoint.
5. Add parallel execution markers to independent CLI subprocess test cases running against isolated temporary directories.
6. Prevent redundant fleet ledger network requests in the frontend deep dive view when switching tickets if the fleet catalog is already resident in memory.

---

## Phase 1: Pure Domain Event Sorting

### Target Files
- backend/internal/domain/events.go
- backend/internal/domain/events_test.go

### Architectural Specification
- Package: domain.
- Define pure function SortEvents(events []EventFact) []EventFact.
- Purity and Value Semantics:
  - Zero imports outside standard library sort and slices. Zero I/O, zero network, zero wall-clock dependencies.
  - Value semantics and immutability: The input slice must not be modified in place. SortEvents must allocate a separate target slice, copy the elements, and sort the copy.
  - Nil or empty input handling: If events is nil or empty, return an allocated empty non-nil slice.
- Sort Ordering:
  - Primary sort key: Timestamp string in ascending lexicographical order.
  - Secondary sort key: Turn integer in ascending numerical order when timestamps match.
  - Ordering stability: Preserve original relative order for events with identical timestamp and turn values using sort.SliceStable.

### Unit Testing Specification
- Package: domain_test.
- File: backend/internal/domain/events_test.go.
- All tests declare t.Parallel().
- Scenarios:
  - Nil input slice returns empty slice with zero length.
  - Empty input slice returns empty slice.
  - Single event returns equivalent single event.
  - Pre-sorted slice maintains ordering.
  - Reversed slice correctly sorts to ascending timestamps.
  - Same timestamp resolves by ascending turn number.
  - Same timestamp and same turn number preserve original arrival sequence.
  - Mutation check: Assert the underlying array of the input slice remains unmodified after SortEvents returns.

---

## Phase 2: Targeted Event Query in HTTP Adapter

### Target Files
- backend/internal/adapters/http/handlers.go
- backend/internal/adapters/http/server_test.go

### Architectural Specification
- Function: WorkspaceEventsHandler(workspacesDir string) http.HandlerFunc.
- Replace full directory traversal (adapters.ScanWorkspaces) with single-ticket loading (adapters.LoadTicketWorkspace).
- Steps:
  1. Validate ticketId parameter for emptiness and directory traversal characters (containing .., /, or \), returning 400 Bad Request via WriteError on failure.
  2. Call adapters.LoadTicketWorkspace(ticketID, workspacesDir).
  3. If LoadTicketWorkspace returns an error or nil, return 404 Not Found via WriteError with code not_found and message Ticket workspace not found.
  4. Extract events from targetLedger.Facts.Events.
  5. Pass the events slice to domain.SortEvents.
  6. Return the sorted slice via WriteEventsResponse(w, sortedEvents).

### Testing Specification
- Verify existing HTTP test TestWorkspaceEventsAndAnomaliesHTTP in backend/internal/adapters/http/server_test.go.
- Add an explicit boundary test verifying that querying events for a missing ticket ID returns HTTP 404 without scanning the entire workspace directory tree.
- Verify hostile ticket IDs return HTTP 400.

---

## Phase 3: Fowler Duplicated Code Smell Elimination in Aggregations

### Target Files
- backend/internal/domain/aggregations.go
- backend/internal/domain/aggregations_test.go

### Architectural Specification
- Problem: The accumulation loop tracking distinct invocation session IDs, summing input tokens, output tokens, cache read tokens, summing cost, tracking peak context tokens, and computing cache hit rate percentage is triplicated in ComputeTotals, ComputeRoleMetrics, and ComputeModelMetrics.
- Refactoring steps:
  1. Define an unexported struct named invocationAccumulator in backend/internal/domain/aggregations.go with fields:
     - invocations of type int64
     - inputTokens of type int64
     - outputTokens of type int64
     - cacheReadTokens of type int64
     - cacheHitRatePct of type float64
     - costUSD of type float64
     - peakContextTokens of type int64
  2. Define an unexported pure function accumulateInvocations(invs []InvocationFact) invocationAccumulator.
     - Responsibilities:
       - Maintain a set of seen invocation IDs to count distinct sessions.
       - Sum input, output, and cache read tokens.
       - Sum cost and round to four decimal places.
       - Track maximum peak context tokens.
       - Calculate cache hit percentage rounded to one decimal place, guarding against division by zero.
  3. Refactor ComputeRoleMetrics:
     - Group invocations by role into mapped slices.
     - Iterate through canonical and custom roles.
     - Call accumulateInvocations for each role slice.
     - Assign accumulated fields to RoleMetric and compute ContextHealth via GetHealthLabel(acc.peakContextTokens, HealthyCeilingTokens).
  4. Refactor ComputeModelMetrics:
     - Group invocations by model into mapped slices.
     - Iterate through models.
     - Call accumulateInvocations for each model slice.
     - Assign accumulated fields to ModelMetric, compute ContextHealth via GetHealthLabel(acc.peakContextTokens, HealthyCeilingTokens), and sort descending by CostUSD.
  5. Refactor ComputeTotals:
     - Use accumulateInvocations or roll up metrics from role slices to compute total invocations, token sums, total cost, cache hit rate, and peak context tokens.
     - Set ContextHealth via string(GetBaseHealthStatus(acc.peakContextTokens, HealthyCeilingTokens)).
     - Return GroupedMetric.

### Unit Testing Specification
- Ensure all existing unit tests in backend/internal/domain/aggregations_test.go pass with zero modifications to public test contracts.
- Confirm exact arithmetic parity across edge cases: zero invocations, shared session IDs across roles, and zero cache tokens.

---

## Phase 4: Repository Encapsulation and CLI Decoupling

### Target Files
- backend/internal/adapters/repository.go
- backend/internal/adapters/repository_test.go
- backend/cmd/agentic-lens/main.go

### Architectural Specification
- Implement persistence helpers in backend/internal/adapters/repository.go:
  1. SaveFactLedger(path string, ledger domain.AgenticFactLedger) error:
     - Determine parent directory using filepath.Dir.
     - Create directory hierarchy with os.MkdirAll(parentDir, 0755).
     - Marshal ledger with two-space indentation using json.MarshalIndent.
     - Append a trailing newline character.
     - Write file using os.WriteFile(path, data, 0644).
     - Wrap any filesystem or marshaling errors with contextual descriptors using %w.
  2. SaveMetricsLedger(path string, ledger domain.AgenticMetricsLedger) error:
     - Determine parent directory using filepath.Dir.
     - Create directory hierarchy with os.MkdirAll(parentDir, 0755).
     - Transform ledger into external format via SerializeLedger(ledger).
     - Marshal formatted DTO with two-space indentation using json.MarshalIndent.
     - Append a trailing newline character.
     - Write file using os.WriteFile(path, data, 0644).
     - Wrap any filesystem or marshaling errors with contextual descriptors using %w.
- Refactor backend/cmd/agentic-lens/main.go:
  1. In the generate command handler, replace lines creating directories and writing jsonOutputFlag with a single call to adapters.SaveFactLedger(*jsonOutputFlag, factLedger).
  2. In the filter command handler, replace lines creating directories and writing jsonOutputFlag with a single call to adapters.SaveMetricsLedger(*jsonOutputFlag, *targetLedger).

### Testing Specification
- Add unit tests in backend/internal/adapters/repository_test.go:
  - TestSaveFactLedger: Asserts directory auto-creation, valid file permissions, JSON formatting, trailing newline, and error return on permission failure.
  - TestSaveMetricsLedger: Asserts correct DTO serialization and persistence to disk.

---

## Phase 5: Test Concurrency and Parallelism

### Target Files
- backend/cmd/agentic-lens/main_test.go

### Architectural Specification
- Add t.Parallel() to independent CLI subprocess tests running against isolated temporary directories:
  - TestCLIListCommand
  - TestCLIAnomaliesCommand
  - TestCLITicketFilterAndMarkdownOutput
  - TestCLIJsonOutput
  - TestCLIAppendIdempotency
  - TestCLINoValidLedgersExit
  - TestCLIServerSubcommandEdgeCases and its internal subtests
  - TestCLIServerInvalidPortRange
  - TestCLIGenerateCommand and its internal subtests
  - TestCLIGenerate_AdversarialEdgeCases
  - TestCLIBackfillCommand
  - TestCLIVersionCommand
- Keep tests that bind fixed network ports (e.g., TestCLIServerAddressAlreadyInUse) serial to avoid socket collisions.

---

## Phase 6: Frontend Ticket Switching Optimization

### Target Files
- web/src/views/TicketDeepDive.tsx
- web/src/views/TicketDeepDive.test.tsx

### Architectural Specification
- In web/src/views/TicketDeepDive.tsx:
  - Decouple fleet catalog loading from ticket detail loading.
  - Load the full ledger fleet via fetchWorkspaceLedgers('all') once when the component mounts or if the ledgers array is empty.
  - When the user selects a new ticket from the dropdown, verify if ledgers are already loaded. If present, bypass fetchWorkspaceLedgers and call only fetchWorkspaceById(selectedId).
  - Update local state (detail, activeTicketId, loading states) cleanly without resetting the ticket selector options.

### Frontend Testing Specification
- In web/src/views/TicketDeepDive.test.tsx:
  - Add test: changing ticket selection in the dropdown fetches detail for the selected ticket but does not re-fetch the fleet ledger catalog if already loaded.
  - Assert fetchWorkspaceLedgers is called only once.
  - Assert fetchWorkspaceById is called with the newly selected ticket ID.
  - Verify all existing tests in TicketDeepDive.test.tsx pass.

---

## Phase 7: Verification and Quality Gates

### Verification Steps
1. Static Analysis:
   - Execute go vet ./... across all backend packages. Zero warnings permitted.
2. Race Detector and Tests:
   - Execute go test -race ./... across backend packages. All tests must pass cleanly.
3. Frontend Validation:
   - Execute npm typecheck in web/. Zero type errors permitted.
   - Execute npm lint in web/. Zero lint warnings permitted.
   - Execute npm test in web/. All test suites must pass.
4. Invariants Check:
   - Confirm zero code blocks or triple backticks in planning documents.
   - Confirm zero emojis across all code, tests, and documentation.
   - Confirm zero any or interface{} in Go and TypeScript code.
