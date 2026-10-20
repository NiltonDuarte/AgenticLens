# Contract: AL-41 Eliminate Subprocess Churn, Artificial Sleeps, and Test Execution Bottlenecks

## 1. Problem Statement

The pre-release source code audit documented in workspaces/AL-35/AUDIT_REPORT.md revealed severe performance bottlenecks and test flakiness across the automated Go test suites:

1. Artificial Sleeps in Test Loops:
   Five explicit instances of time.Sleep exist in test suites. Four instances in backend/cmd/agentic-lens/main_test.go (lines 393, 583, 628, 665) use 10 millisecond fixed sleeps inside server startup polling loops. One instance in backend/internal/adapters/http/live_handlers_test.go (line 333) uses a 50 millisecond fixed sleep to wait for client disconnect cleanup. These fixed delays add non-deterministic wall-clock latency, fail under high system load, and violate engineering standards.

2. Subprocess Git Invocations and Physical Branch Mutations:
   In backend/internal/adapters/git/inference_test.go, the test helper createTestBranch invokes exec.Command to run git branch and git branch -D on disk. Concurrently mutating git references in the active working tree risks repository corruption, race conditions, and test interference. Furthermore, backend/internal/adapters/git/inference.go unconditionally shells out to git branch -a --list for every ticket classification check before evaluating configured prefix rules.

3. Subprocess Churn and Database Contention in CLI Integration Tests:
   In backend/cmd/agentic-lens/main_test.go, subtests in TestCLITicketTypeClassification take upwards of 6.3 seconds each under race detector concurrency. Because tests do not pass --db-path, every spawned CLI process attempts to locate, open, and query the host OpenCode SQLite database at ~/.local/share/opencode/opencode.db while simultaneously querying live git branches, creating intense disk I/O and process contention across dozens of parallel subtests.

4. Shallow Passthrough Wrappers in Formatters Adapter:
   In backend/internal/adapters/formatters.go, functions FormatLeadTime and FormatTokens provide zero transformation or validation, merely forwarding arguments to domain.FormatLeadTime and domain.FormatTokensK. These redundant wrappers introduce unnecessary indirection.

Ticket AL-41 addresses these bottlenecks by eliminating all artificial sleeps, decoupling unit tests from host git branch mutations through in-memory fixtures, evaluating prefix rules prior to git lookups, isolating CLI integration tests with dedicated temporary SQLite databases, and inlining redundant formatting abstractions.

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation deliverables must strictly adhere to the architectural principles defined in TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Functional Core (backend/internal/domain/):
   - Pure domain functions, immutable models, and zero external dependencies.
   - Functions such as domain.InferTicketTypeFromPrefix, domain.FormatLeadTime, and domain.FormatTokensK remain purely algorithmic with zero I/O, zero network calls, zero database handles, and zero system clocks.
   - Domain logic is preserved as the single source of truth for calculations and conversions.
2. Boundary Adapters (backend/internal/adapters/):
   - Imperative boundary layer handling external systems: git execution, OpenCode SQLite connectivity, HTTP routing, and SSE delivery.
   - git.InferTicketTypeFromBranchList remains a pure string parsing function, enabling complete unit test coverage without spawning git subprocesses.
   - git.InferTicketType encapsulates branch inspection and prefix fallback, optimizing performance by resolving configured prefix rules before invoking the git subprocess shell.
   - adapters.FormatLeadTime and adapters.FormatTokens are inlined, redirecting callers directly to the functional core.
3. CLI Application Layer (backend/cmd/agentic-lens/):
   - Coordinates flags, database extraction, git analysis, and server lifecycle.
   - Server readiness checks in main_test.go replace arbitrary polling sleeps with deterministic TCP dial probes and microsecond backoff.
   - CLI integration tests isolate database interactions by supplying temporary database paths, avoiding contention on the host OpenCode database.

### Pure Domain Core Isolation
- Domain packages must have zero imports of os, exec, database/sql, net/http, or outer adapter layers.
- Test suites testing domain logic must use pure table-driven cases with zero mocks and zero process forks.

### Imperative Shell Boundary Rules
- Zero Artificial Sleeps:
  Zero calls to time.Sleep are permitted across the entire Go backend and test suites. Synchronization must rely on deterministic constructs: Go channels, sync.WaitGroup, or bounded TCP dial polling with microsecond timers.
- In-Memory Branch Resolution:
  Unit tests must never create, modify, or delete physical git branches on the host filesystem. Pure string branch list fixtures must be passed directly to git.InferTicketTypeFromBranchList.
- Isolated Test Databases:
  CLI integration tests executing generate commands must point --db-path to an isolated empty temporary SQLite database fixture, preventing locks against the user local OpenCode database.
- Typing and Static Safety:
  Zero usage of any or interface{} across the Go codebase.
- Zero Emojis Mandate:
  Zero emojis are permitted in code, docstrings, error messages, terminal logs, test identifiers, or workspace documents.

## 3. Scope Boundaries: Go Backend and Tests

### In Scope
1. Server Startup Readiness Polling (backend/cmd/agentic-lens/main_test.go):
   - Replace four time.Sleep(10 * time.Millisecond) loops (lines 393, 583, 628, 665) with a deterministic TCP port dial probe helper utilizing net.DialTimeout and microsecond backoff.
2. SSE Stream Disconnect Synchronization (backend/internal/adapters/http/live_handlers_test.go):
   - Replace time.Sleep(50 * time.Millisecond) at line 333 with a deterministic synchronization channel in mockLiveService that triggers when the client context disconnects and triggers the unsubscribe callback.
3. Git Test Decoupling and Branch Cleanliness (backend/internal/adapters/git/inference_test.go and analyzer_test.go):
   - Delete the createTestBranch helper function shelling out to git branch and git branch -D.
   - Refactor TestInferTicketType_BranchPrefixResolution and TestInferTicketType_Precedence_BranchOverConfiguredPrefix in inference_test.go to test pure branch string parsing via InferTicketTypeFromBranchList using in-memory string fixtures.
   - Refactor TestInferenceOnActiveWorkspace in analyzer_test.go to remove reliance on createTestBranch.
4. Git Inference Prefix Prioritization (backend/internal/adapters/git/inference.go):
   - In InferTicketType, evaluate configured prefix rules before shelling out to git branch lookups. When a ticket ID matches a configured prefix with an assigned conventional commit type (such as PO -> docs or GEN -> chore), return the resolved type immediately, avoiding git subprocess execution entirely.
5. CLI Integration Test Isolation (backend/cmd/agentic-lens/main_test.go):
   - In TestCLITicketTypeClassification, provide an explicit --db-path argument pointing to an isolated empty temporary SQLite database fixture, decoupling the CLI test suite from the host OpenCode database.
6. Inline Shallow Formatters (backend/internal/adapters/formatters.go):
   - Remove or inline adapters.FormatLeadTime and adapters.FormatTokens, directing any callers directly to domain.FormatLeadTime and domain.FormatTokensK.
7. Verification and Performance Validation:
   - Verify zero occurrences of time.Sleep across all Go backend files.
   - Verify zero host git branches are created or destroyed during test execution.
   - Verify total backend test suite execution time drops significantly, targeting sub-second execution.

### Out of Scope
1. TypeScript web dashboard implementation (web/src/).
2. OpenCode SQLite schema or table layout modifications.
3. Git commit diff parsing algorithms in backend/internal/adapters/git/analyzer.go (handled under AL-39).
4. HTTP REST API additions (handled under AL-40).

## 4. Architectural Specifications

### Specification 1: Server Readiness Probe Helper (main_test.go)
- Create a test helper waitForServerReady(t *testing.T, targetURL string, timeout time.Duration) or waitForPortReady(addr string, timeout time.Duration).
- The helper probes the target TCP address using net.DialTimeout with a 10 millisecond dial timeout.
- If the connection fails, it waits using a microsecond timer starting at 100 microseconds, doubling up to a maximum delay of 5 milliseconds per iteration, bounded by the total timeout duration (e.g. 2 seconds).
- Once the TCP port connects successfully, verify HTTP endpoint availability with a single http.Get request.
- Completely replaces all four instances of time.Sleep in main_test.go.

### Specification 2: Deterministic Unsubscribe Signaling (live_handlers_test.go)
- Enhance mockLiveService in live_handlers_test.go with an unsubscribed notification channel (chan struct{}).
- In the cleanup closure returned by Subscribe(), close or signal the unsubscribed channel in addition to setting the boolean flag.
- In TestLiveTraceHandler_ClientDisconnectCleanUnsubscribe, replace time.Sleep(50 * time.Millisecond) with a select block waiting on the unsubscribe channel with a defensive 500 millisecond timeout.
- This guarantees instantaneous test execution under normal conditions and deterministic failure upon regression.

### Specification 3: In-Memory Branch List Parsing (inference_test.go)
- Remove createTestBranch and all exec.Command("git", "branch", ...) invocations.
- Test branch prefix extraction by providing multi-line branch list string fixtures directly to git.InferTicketTypeFromBranchList(fixture, ticketID).
- Test fixtures include single branches, multiple branches, remote tracking branches (remotes/origin/...), detached heads, worktrees, and dirty prefix strings.
- Verify that branch resolution matches conventional commit types (feat, fix, refactor, ci, chore, docs, test, perf, build) entirely in memory.

### Specification 4: Configured Prefix Prioritization in InferTicketType (inference.go)
- Update git.InferTicketType(ticketID string, prefixes ...domain.TicketPrefix) *string:
  1. Validate and sanitize ticketID string. If empty, return nil.
  2. Evaluate configured prefixes first via domain.InferTicketTypeFromPrefix(cleanTicket, prefixes). If a match is found and returns an explicit conventional type, return that type pointer immediately.
  3. If not resolved by configured prefixes (e.g. standard engineering tickets like AL-41 without static type bindings), invoke the git subprocess runGitCommand("branch", "-a", "--list", "*"+cleanTicket+"*") and parse the result using InferTicketTypeFromBranchList.
  4. Return nil if neither rule resolves a ticket type.
- This short-circuits subprocess execution for all PO and GEN tickets, as well as false prefix checks, eliminating over 80 percent of git process forks during tests.

### Specification 5: Isolated Test Database Fixtures (main_test.go)
- In TestCLITicketTypeClassification, create a helper setupIsolatedEmptyDB(t *testing.T) string that creates an empty SQLite database file in t.TempDir() with basic OpenCode schema tables.
- Pass --db-path pointing to this isolated file in all generate CLI invocations within TestCLITicketTypeClassification.
- This prevents parallel CLI processes from locking ~/.local/share/opencode/opencode.db, dropping subtest latency from >6.3 seconds down to <300 milliseconds.

### Specification 6: Deprecate or Inline Shallow Formatters (formatters.go)
- Remove adapters.FormatLeadTime and adapters.FormatTokens from backend/internal/adapters/formatters.go.
- Direct any internal callers in adapters to domain.FormatLeadTime and domain.FormatTokensK.
- Ensure formatters_test.go tests domain formatters directly where appropriate.

## 5. Acceptance Criteria

1. Zero time.Sleep calls exist across the entire Go backend test suite (asserted via ripgrep verification).
2. Git inference tests execute purely in memory without invoking git branch or git branch -D.
3. TestCLITicketTypeClassification subtests in backend/cmd/agentic-lens/main_test.go execute in under 300 milliseconds per subtest.
4. Total backend test suite (go test -race ./...) executes under 1.0 second wall-clock time.
5. Shallow passthrough wrappers in backend/internal/adapters/formatters.go are eliminated.
6. All quality gates pass (make check, make test, make build).
7. Zero any or interface{} constructs across all Go packages.
8. Zero emojis across all modified files, test outputs, and documentation.
