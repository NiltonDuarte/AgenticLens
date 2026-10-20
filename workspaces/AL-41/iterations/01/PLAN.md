# Implementation Plan: AL-41 Eliminate Subprocess Churn, Artificial Sleeps, and Test Execution Bottlenecks

## 1. Goal and Iteration Scope

Eliminate test execution bottlenecks identified in audit report AL-35 by replacing all artificial sleeps with deterministic synchronization, eliminating host git branch mutations through in-memory string fixtures, short-circuiting git subprocess forks via prefix evaluation, isolating CLI integration tests with dedicated temporary SQLite databases, and removing shallow passthrough wrappers.

This iteration addresses the entire scope of ticket AL-41 across backend/cmd/agentic-lens, backend/internal/adapters/git, backend/internal/adapters/http, and backend/internal/adapters/formatters.go.

## 2. Grounded Reference Architecture and Concrete Paths

The implementation targets the following concrete files and packages:

### Boundary Adapters Layer (backend/internal/adapters/)
- backend/internal/adapters/formatters.go: Remove shallow forwarding functions FormatLeadTime and FormatTokens. Direct internal callers to domain.FormatLeadTime and domain.FormatTokensK.
- backend/internal/adapters/formatters_test.go: Update tests referencing removed adapter formatters to invoke domain formatters.
- backend/internal/adapters/git/inference.go: Update InferTicketType to evaluate configured prefix rules before invoking git subprocesses.
- backend/internal/adapters/git/inference_test.go: Remove createTestBranch helper. Refactor branch resolution tests to evaluate InferTicketTypeFromBranchList using in-memory string fixtures.
- backend/internal/adapters/git/analyzer_test.go: Refactor TestInferenceOnActiveWorkspace to eliminate reliance on createTestBranch.
- backend/internal/adapters/http/live_handlers_test.go: Add unsubscribe notification channel to mockLiveService and replace 50ms time.Sleep with channel synchronization.

### CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main_test.go:
  - Add deterministic TCP dial readiness probe helper and replace four 10ms time.Sleep loops.
  - Create isolated empty SQLite database fixture and pass --db-path across TestCLITicketTypeClassification subtests.

## 3. Phased Implementation Tasks

### Phase 1: Eliminate Artificial Sleeps in Test Suites
- Task 1.1: Refactor Server Readiness Polling in backend/cmd/agentic-lens/main_test.go:
  - Implement test helper waitForServerReady(t *testing.T, targetURL string, timeout time.Duration).
  - Use net.DialTimeout against the target host:port extracted from targetURL with a 10ms dial timeout.
  - If dial fails, use a select block with time.After starting at 100 microseconds, doubling up to a maximum cap of 5 milliseconds, bounded by the overall timeout duration.
  - Once TCP connection succeeds, issue http.Get to confirm endpoint responsiveness.
  - Replace time.Sleep(10 * time.Millisecond) in TestCLIServerCommand (line 393).
  - Replace time.Sleep(10 * time.Millisecond) in TestCLIServerPortEnv (line 583).
  - Replace time.Sleep(10 * time.Millisecond) in TestCLIServerPortFlagPrecedence valid flag overrides valid PORT env (line 628).
  - Replace time.Sleep(10 * time.Millisecond) in TestCLIServerPortFlagPrecedence valid flag overrides invalid PORT env (line 665).
- Task 1.2: Refactor Client Disconnect Synchronization in backend/internal/adapters/http/live_handlers_test.go:
  - Add unsubChan chan struct{} field to mockLiveService struct.
  - Initialize unsubChan in setup or provide an accessor.
  - In mockLiveService.Subscribe(), inside the cleanup closure, perform a non-blocking send or close on unsubChan under synchronization mutex.
  - In TestLiveTraceHandler_ClientDisconnectCleanUnsubscribe, replace time.Sleep(50 * time.Millisecond) (line 333) with a select block waiting on unsubChan with a 500 millisecond timeout via time.After.
  - Assert that isUnsubscribed() returns true.

### Phase 2: In-Memory Branch Resolution and Subprocess Short-Circuiting in Git Adapter
- Task 2.1: Prioritize Configured Prefix Evaluation in backend/internal/adapters/git/inference.go:
  - In InferTicketType, evaluate domain.InferTicketTypeFromPrefix(cleanTicket, prefixes) before running runGitCommand.
  - If a configured prefix matches and yields a valid conventional commit type, return a pointer to that string immediately without shelling out to git.
  - If no configured prefix matches, execute runGitCommand("branch", "-a", "--list", "*"+cleanTicket+"*") and evaluate InferTicketTypeFromBranchList.
  - If neither matches, return nil.
- Task 2.2: Decouple Unit Tests from Disk Branch Mutations in backend/internal/adapters/git/inference_test.go:
  - Delete createTestBranch helper function.
  - Refactor TestInferTicketType_BranchPrefixResolution to test InferTicketTypeFromBranchList directly with in-memory branch list fixtures representing feat, ci, refactor, and fix branches.
  - Refactor TestInferTicketType_Precedence_BranchOverConfiguredPrefix: verify precedence using in-memory branch fixture parser and pure prefix evaluator without disk mutation.
  - Ensure all subtests in TestInferTicketType_PrefixMapping and TestInferTicketType_AdversarialEdgeCases execute purely in memory.
- Task 2.3: Remove Branch Creation in backend/internal/adapters/git/analyzer_test.go:
  - In TestInferenceOnActiveWorkspace, remove call to createTestBranch.
  - Assert active workspace properties (SDLC iterations, git user) without mutating git branches.
  - Test branch inference via InferTicketTypeFromBranchList using in-memory branch string.

### Phase 3: CLI Integration Test Isolation and Database Decoupling
- Task 3.1: Create Isolated Test SQLite Database Fixture Helper in backend/cmd/agentic-lens/main_test.go:
  - Implement setupIsolatedEmptyDB(t *testing.T) string helper that initializes an empty SQLite database with the standard OpenCode schema in t.TempDir().
- Task 3.2: Isolate TestCLITicketTypeClassification Subtests in backend/cmd/agentic-lens/main_test.go:
  - In TestCLITicketTypeClassification, initialize the isolated database fixture once or per subtest group.
  - Pass --db-path <isolatedDB> to all exec.Command invocations for generate across subtests (explicit alias feature, invalid ticket type, automatic PO, automatic GEN, unmatched ticket warning, strict-type failure, cached facts normalization, mixed casing normalization, false prefix POD-1, false prefix GENTLE-1, strict-type fallback UNKNOWN).
  - Verify all subtests execute in parallel without lock contention on ~/.local/share/opencode/opencode.db.

### Phase 4: Inline Shallow Passthrough Wrappers
- Task 4.1: Eliminate Shallow Wrappers in backend/internal/adapters/formatters.go:
  - Remove FormatLeadTime(seconds *int64) string.
  - Remove FormatTokens(tokens int64) string.
  - In formatters.go internal usages (e.g. lines 220, 221, 235, 330, 343), verify calls use domain.FormatTokensK directly.
- Task 4.2: Update formatters_test.go and Adapter Callers:
  - If backend/internal/adapters/formatters_test.go contains tests for the removed passthrough functions, redirect tests to domain.FormatLeadTime and domain.FormatTokensK or remove redundant wrapper tests.
  - Inspect backend/internal/adapters/ to ensure zero unresolved references remain.

### Phase 5: Verification, Benchmarking, and Compliance Gates
- Task 5.1: Assert Zero time.Sleep in Test Suites:
  - Verify that no time.Sleep occurrences remain across all Go files using search tooling.
- Task 5.2: Assert Zero Git Branch Mutation:
  - Verify that running go test ./backend/internal/adapters/git/... creates zero physical git branches in the local repository.
- Task 5.3: Measure Test Execution Performance:
  - Execute go test -race ./backend/... and verify wall-clock execution time is reduced to target (<1.0s).
  - Benchmark TestCLITicketTypeClassification and verify subtests complete in <300ms each.
- Task 5.4: Quality Gates and Static Checks:
  - Run go vet ./... across the backend module.
  - Run make check from repository root.
  - Run make test from repository root.
  - Run make build from repository root.
  - Verify zero any or empty interface{} declarations across Go files.
  - Verify zero emojis across all code files, comments, docstrings, error messages, test titles, and markdown deliverables.

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Banish Artificial Sleeps) removes the 5 known time.Sleep delays, establishing non-blocking TCP readiness and channel-based event delivery.
2. Phase 2 (In-Memory Branch Resolution and Prefix Prioritization) removes disk branch mutation, eliminates createTestBranch, and short-circuits git subprocess execution for prefixed tickets.
3. Phase 3 (CLI Integration Test Isolation) isolates TestCLITicketTypeClassification from the host OpenCode database, dropping execution time from >6.3s to <300ms per subtest.
4. Phase 4 (Inline Shallow Passthrough Wrappers) removes dead indirection in formatters.go.
5. Phase 5 (Verification and Gates) runs automated test suites with race detection, validates performance metrics, and ensures zero any and zero emojis compliance.

## 5. Verification Plan

### Verification Commands
- Check for zero time.Sleep:
  Inspect backend/ using grep to confirm 0 matches for time.Sleep.
- Static Analysis:
  go vet ./... in backend/
- Unit and Race Detected Test Suites:
  go test -race ./... in backend/
- Full Project Verification:
  make check && make test && make build

### Manual Invariant Verification
- Confirm git status shows no untracked or modified branches created during test runs.
- Confirm TestCLITicketTypeClassification subtests complete without subprocess timeouts or SQLite busy errors.
- Confirm zero occurrences of any or interface{} in modified Go files.
- Confirm zero emojis across all diffs.
