# AL-42 Iteration 01 Plan

## Scope

Implement the unified execution-level classifier and establish offline/live parity for tool failure facts, following `../CONTRACT.md` and the AL-42 specification. Work is limited to Go domain logic, OpenCode/live adapters, and their tests.

## Grounded code locations

- `backend/internal/domain/live.go`: `FailureSeverity`, `LivePartFact`, and current `ClassifyToolFailure` provide the classifier/result neighborhood. Introduce the structured execution result and classifier here; retain or adapt compatibility only where existing callers require it.
- `backend/internal/domain/live_test.go`: existing classifier tests cover status/content classification, adversarial cases, and message truncation. Update/add table-driven tests to assert the new execution contract, including benign grep/glob no-match results.
- `backend/internal/domain/anomalies.go`: `detectToolRetryStreaksWithIndexes` and `detectToolHighFailureRates` consume tool/event facts. Ensure their failure decisions reflect normalized execution errors only.
- `backend/internal/adapters/opencode/extractor.go`: offline extraction currently computes tool failure state independently; route this derivation through `domain.ClassifyToolExecution`.
- `backend/internal/adapters/live/service.go`: classifier calls occur in trace assembly and live event transformations. Route both through the structured unified classifier and preserve matching status, severity, and sanitized error message.
- `backend/internal/adapters/opencode/` and `backend/internal/adapters/live/` test suites: add parity coverage using equivalent persisted part inputs, comparing tool error totals and event statuses.

## Implementation sequence

1. Define `ToolExecutionResult` and pure `ClassifyToolExecution` in the domain. Recognize explicit failed/error statuses, process exit failures, crash/signal termination, timeouts, and permission denial. Do not classify arbitrary semantic output text as a failed execution. Treat grep/glob no-match and conventional exit status 1 as successful completion.
2. Make severity escalation depend on authentic failure classification and preserve optional error-message sanitization with a 2048-character cap.
3. Update domain anomaly detection inputs/logic as needed so retry streaks and high failure rates only count true execution failures.
4. Replace offline extractor's inline classifier and update both live service classifier call sites. Ensure both adapters map the same execution result to `ToolFact.Errors` and `EventFact.Status`.
5. Add/adjust tests: domain classification matrix, anomaly regression for benign search results, and adapter parity over equivalent synthetic SQLite part records.

## Acceptance checks

- Completed calls containing semantic error words stay completed unless the execution record explicitly reports a hard failure.
- Genuine process failures classify consistently regardless of tool name.
- Grep/glob 0-match cases remain non-errors and do not contribute to error counts or retry anomalies.
- Offline and live paths use the same domain classifier and agree on tool error totals/event statuses for equivalent records.
- Error messages are capped at 2048 characters and consecutive failure severity escalation matches in both paths.
- `go test -race ./...` and `make check` pass, with any unrelated baseline failure documented.

## Risks and implementation notes

- Confirm the precise persisted representation of exit codes/signals and status normalization in each adapter before mapping classifier inputs; preserve distinctions needed for grep/glob exit code 1.
- Existing callers or tests may rely on `ClassifyToolFailure`; migrate them coherently or retain a narrowly scoped compatibility wrapper without preserving content-based failure inference.
- The parity test should exercise the actual offline extraction and live service transformation seams available in the current packages, avoiding mocks in domain tests.
