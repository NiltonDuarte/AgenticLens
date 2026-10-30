# AL-42 Technical Contract

## Problem

Tool execution failures are currently inferred partly from output text in `backend/internal/domain/live.go` through `ClassifyToolFailure`. That mixes process execution status with semantic content and can label successful tools as failed. Offline extraction in `backend/internal/adapters/opencode/extractor.go` independently derives errors, while `backend/internal/adapters/live/service.go` calls the domain classifier. The two pipelines can therefore report different tool error counts and event statuses for the same execution.

## Goal

Define one pure domain execution classifier and use it in both offline OpenCode extraction and live processing. Count only genuine execution failures: explicit error/failed statuses, non-zero process exit results, process crashes or signal termination, timeouts, and permission denial. Successful grep/glob calls with no matches remain successful. Preserve consistent severity escalation and capped error messages across both boundaries.

## Contract and acceptance boundaries

- `ClassifyToolExecution(toolName, status, output, consecutiveErrors)` returns a structured `ToolExecutionResult` containing `IsError`, `Severity`, and optional sanitized `ErrorMessage`.
- Output text alone must not turn an otherwise completed execution into a failure merely because it contains traceback, syntax, missing-resource, or other semantic error wording. Explicit execution-failure signals remain classifiable.
- Grep and glob no-match outcomes, including their conventional exit code 1, are clean completions.
- Retry-streak and high-failure-rate anomalies count only authentic execution failures.
- `backend/internal/adapters/opencode/extractor.go` and `backend/internal/adapters/live/service.go` use the same domain classifier for tool error and event status derivation.
- Both paths preserve the classifier's severity escalation and sanitize/cap the error message at 2048 characters.
- Identical underlying execution records produce matching offline tool error totals and live/offline event statuses.

## Architecture boundaries

- `backend/internal/domain/live.go` owns deterministic execution classification and its value result. It performs no I/O, database access, clock reads, or adapter imports.
- `backend/internal/domain/anomalies.go` evaluates tool error counts/statuses from the normalized execution facts; it does not infer failures from arbitrary output content.
- `backend/internal/adapters/opencode/` translates persisted OpenCode records into classifier inputs and facts.
- `backend/internal/adapters/live/service.go` translates live records into the same classifier inputs and facts.
- Tests belong beside the domain and adapter behavior they verify. No frontend or schema changes are in scope.

## Out of scope

- Semantic interpretation of successful tool output. That is a separate concern and is explicitly excluded from execution failure counts.
- UI changes, new telemetry fields, persistence schema changes, and unrelated anomaly policy changes.

## Verification

Table-driven domain tests cover statuses, exit codes, signal/crash conditions, timeouts, permission failures, benign grep/glob no-match cases, escalation, and message truncation. Adapter parity tests feed equivalent persisted records into offline extraction and live service and compare tool error totals and event statuses. Run `go test -race ./...` and the repository quality gates; report any pre-existing or environment-related failures explicitly.
