h3. Title: AL-41: Eliminate Subprocess Churn, Artificial Sleeps, and Test Execution Bottlenecks

*TARGET*: AgenticLens (tests | adapters | git | cli | http)

*CONTEXT*

The pre-release source code audit (AL-35) revealed severe execution bottlenecks across the automated test suites. In `backend/cmd/agentic-lens/main_test.go`, subtests in `TestCLITicketTypeClassification` take over 6.3 seconds each under `-race` concurrency because nearly 100 subprocess executions compete for live git branch lookups and host OpenCode SQLite database access. In `backend/internal/adapters/git/inference_test.go`, test helpers shell out to `git branch` and `git branch -D` on disk, creating physical branch churn in the host repository. Furthermore, 5 instances of `time.Sleep` exist in test loops (`main_test.go:393, 583, 628, 665` and `live_handlers_test.go:333`), adding non-deterministic latency. Finally, shallow passthrough functions in `backend/internal/adapters/formatters.go` (`FormatLeadTime`, `FormatTokens`) introduce unnecessary layers of indirection.

*GOAL*

Eliminate test bottlenecks, remove all artificial sleeps, decouple unit tests from host git branch mutations by using in-memory branch resolution, isolate CLI integration test environments with temporary SQLite databases, and inline shallow forwarding wrappers.

*VERTICAL SLICE SCOPE*

* Banish Artificial Sleeps (`backend/cmd/agentic-lens/main_test.go`, `internal/adapters/http/live_handlers_test.go`):
  - Replace 10ms server readiness polling loops in `main_test.go` with deterministic TCP port dial checks and microsecond backoff.
  - Replace 50ms sleep in `live_handlers_test.go` with explicit channel notification upon event broadcast.
* In-Memory Branch Resolution for Git Tests (`backend/internal/adapters/git/inference_test.go`):
  - Remove `createTestBranch` helper shelling out to `git branch`.
  - Refactor test cases to test pure string branch parser `InferTicketTypeFromBranchList` using deterministic in-memory fixtures.
  - Evaluate configured prefix rules before shelling out to git in `InferTicketType`.
* CLI Integration Test Isolation (`backend/cmd/agentic-lens/main_test.go`):
  - Pass `--db-path` pointing to an isolated empty temporary SQLite database or disable live DB resolution during pure ticket classification CLI tests.
  - Mitigate subprocess contention in table-driven classification tests.
* Inline Shallow Passthrough Wrappers (`backend/internal/adapters/formatters.go`):
  - Remove or inline `adapters.FormatLeadTime` and `adapters.FormatTokens`, directing callers straight to pure domain functions `domain.FormatLeadTime` and `domain.FormatTokensK`.
* Quality & Verification:
  - Assert backend test suite executes under 1.0 second wall-clock.
  - Assert zero `time.Sleep` calls remain across `*_test.go`.
  - Verify zero git branch creation or deletion occurs during test execution.

*ACCEPTANCE CRITERIA*

* Zero `time.Sleep` calls remain in the Go backend test suite.
* Git unit tests execute in memory without creating host git branches.
* CLI ticket type classification test suite runtime drops from >6s per subtest to <300ms.
* Shallow passthrough functions are eliminated.
* Total test execution time is significantly reduced.
* Quality gates pass (`make check`, `make test`, `make build`) with zero emojis.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Zero any and zero emojis across all files.
