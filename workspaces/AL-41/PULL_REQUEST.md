# perf(tests): eliminate subprocess churn, artificial sleeps, and test execution bottlenecks

## Description
This pull request addresses the test flakiness and performance bottlenecks identified in the audit report (AL-35):

1. **Eliminate Artificial Sleeps**:
   - Replaced four `time.Sleep(10 * time.Millisecond)` server startup polling calls in `backend/cmd/agentic-lens/main_test.go` with a deterministic TCP dial probe `waitForServerReady` with exponential microsecond backoff.
   - Replaced `time.Sleep(50 * time.Millisecond)` in `backend/internal/adapters/http/live_handlers_test.go` with deterministic channel signaling on unsubscribe.
   - Achieved zero `time.Sleep` calls across the entire Go backend.

2. **Decouple Unit Tests from Host Git Branches**:
   - Deleted disk-mutating `createTestBranch` helper in `backend/internal/adapters/git/inference_test.go`.
   - Refactored branch inference tests to pass in-memory string fixtures directly to `InferTicketTypeFromBranchList`.

3. **Prefix Prioritization in Git Inference**:
   - Evaluated configured prefix rules before shelling out to `git branch -a --list` in `InferTicketType`. Configured prefix matches short-circuit subprocess invocations immediately.

4. **Isolate CLI Integration Tests**:
   - Added `setupIsolatedEmptyDB` helper in `backend/cmd/agentic-lens/main_test.go` and passed `--db-path` across all `TestCLITicketTypeClassification` subtests.
   - Prevented disk lock contention on the host OpenCode SQLite database, dropping subtest runtimes from >6.3s to <0.15s.

5. **Inline Shallow Passthrough Formatters**:
   - Removed `FormatLeadTime` and `FormatTokens` from `backend/internal/adapters/formatters.go`. Callers now directly use `domain.FormatLeadTime` and `domain.FormatTokensK`.

## Validation
- `go test -race ./...` in `backend/`: all packages pass.
- `make check && make build`: successfully passes `go vet` and compiles binaries for all supported platforms.
- Verified zero emojis in all code, comments, logs, and docs.
