h3. Title: AL-42: Unify Tool Execution Error Classification and Offline-Live Pipeline Parity

*TARGET*: AgenticLens (domain | adapters | live | opencode | tests)

*CONTEXT*
 
Currently, tool error classification is broken across two dimensions:
1. Conflation: `domain.ClassifyToolFailure` combines process-level failures (non-zero exit codes, crashed processes, command failed) with arbitrary stdout/stderr content searches (traces, syntax errors, and even benign empty grep/glob results with 0 matches). When string matches occur, completed tool executions are overridden to `status: "error"`, misrepresenting benign or normal tool completions as broken tooling and triggering false-positive `TOOL_RETRY_STREAK` and `TOOL_HIGH_FAILURE_RATE` anomalies.
2. Divergence: The offline telemetry generation engine (`internal/adapters/opencode/extractor.go`) does not use `domain.ClassifyToolFailure`. Instead, it uses an ad-hoc inline conditional check (`stat == "error" || strings.Contains(out, "Traceback") || strings.HasPrefix(trimmedOut, "Error:")`). The live observability service (`internal/adapters/live/service.go`) uses `domain.ClassifyToolFailure`. Consequently, identical agent runs produce different error metrics when viewed live versus after telemetry generation.

*GOAL*
 
Refactor domain tool failure classification into a pure execution-level classifier (`ClassifyToolExecution`), eliminate erroneous classifications for benign operations (such as empty search results in grep and glob), wire both the offline extractor and live service to the single unified domain function, and ensure strict parity in tool execution error counts across offline and live pipelines.

*VERTICAL SLICE SCOPE*
 
* Domain Core (`internal/domain/live.go` & `internal/domain/anomalies.go`):
  - Refactor or introduce `ClassifyToolExecution(toolName string, status string, output string, consecutiveErrors int) ToolExecutionResult` returning structured execution status (`IsError bool`, `Severity FailureSeverity`, `ErrorMessage *string`).
  - Restrict execution failure detection to hard process execution errors: non-zero exit codes, `status == "error"` or `status == "failed"`, process crash/signal terminations (e.g., segfault, killed, OOM), timeouts, and permission denied.
  - Fix benign search misclassification: ensure `grep` and `glob` with exit code 1 or empty output indicating 0 matches are classified as clean completions (`IsError: false`, `Severity: SeverityNone`).
  - Ensure `detectToolRetryStreaks` and `detectToolHighFailureRates` in `internal/domain/anomalies.go` evaluate strictly against authentic execution errors.
* Boundary Adapters (`internal/adapters/opencode/extractor.go` & `internal/adapters/live/service.go`):
  - Replace ad-hoc inline error conditionals in `internal/adapters/opencode/extractor.go` with calls to `domain.ClassifyToolExecution`.
  - Update `internal/adapters/live/service.go` to use the refactored execution classifier for both trace assembly and live event stream transformations.
  - Ensure consecutive error escalation and error message sanitization (capped at 2048 chars) behave identically in both adapters.
* Quality:
  - Table-driven Go unit tests in `internal/domain/` testing execution classification across various tools (bash, read, edit, grep, glob), exit codes, signal crashes, and timeouts.
  - Explicit test case asserting that grep/glob 0-match results return `IsError: false` and do not increment tool error counts.
  - Parity test suite feeding identical synthetic SQLite part records to both `opencode.ExtractFacts` and `live.Service` and asserting identical `ToolFact.Errors` and `EventFact.Status`.

*ACCEPTANCE CRITERIA*
 
* `ClassifyToolExecution` returns `IsError: true` strictly for genuine execution failures, process crashes, timeouts, and non-zero exit codes.
* Grep and glob operations returning 0 matches are never classified as tool execution failures.
* `internal/adapters/opencode/extractor.go` and `internal/adapters/live/service.go` both invoke the pure domain execution classifier.
* Given the same underlying SQLite part records, offline extracted facts and live streaming facts produce identical tool call error counts and event statuses.
* `TOOL_RETRY_STREAK` and `TOOL_HIGH_FAILURE_RATE` anomalies only trigger on true execution failures.
* Go test suite (`go test -race ./...`) and lint checks pass cleanly with zero warnings.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
