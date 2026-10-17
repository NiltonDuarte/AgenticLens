h3. Title: AL-7: Domain Anomaly Detection and Rule Engine

*TARGET*: AgenticLens (domain | cli | tests)

*CONTEXT*

Multi-agent development sessions occasionally encounter execution degradation such as repetitive tool error streaks (e.g. repeated failing edits or commands), rapid context window saturation jumps, or cyclic ping-pong handoffs between subagents. We need a deterministic rule evaluator in the pure Go domain to identify these patterns and flag them in CLI reports.

*GOAL*

Build a pure functional anomaly detection engine in the Go domain evaluating deterministic rules against `AgenticFactLedger`, complete with CLI table formatting and summary markdown integration.

*VERTICAL SLICE SCOPE*

* Domain Core: Pure structs and rule evaluator in `internal/domain/anomalies.go`:
  - Models: `Anomaly`, `AnomalySeverity` (`WARNING`, `CRITICAL`), `AnomalyRuleID`, `AnomalyContext`.
  - Rules:
    1. `TOOL_RETRY_STREAK`: Detect 3+ consecutive failures of the same tool or command sequence.
    2. `TOOL_HIGH_FAILURE_RATE`: Detect tools with error rate > 20% (minimum 5 invocations).
    3. `CONTEXT_SATURATION_SPIKE`: Single-turn context jump exceeding 40k tokens or reaching `CRITICAL` (>160k tokens) ceiling.
    4. `SUBAGENT_REWORK_LOOP`: Repeated cyclic transitions between identical worker roles exceeding threshold.
  - Pure function `DetectAnomalies(ledger AgenticFactLedger) []Anomaly`.
* Boundary Adapter & CLI:
  - Add anomaly summary section to Markdown summary formatter (`internal/adapters/formatters.go`).
  - Add `--anomalies` flag and table column to `cmd/agentic-lens/main.go`.
* Quality: Comprehensive table-driven tests in `internal/domain/anomalies_test.go` covering edge cases: clean sessions, consecutive failures, mixed successes/failures, legacy v1 fallback behavior, and multi-anomaly runs. Zero mocks.

*ACCEPTANCE CRITERIA*

* `DetectAnomalies` correctly flags 3 consecutive tool failures in v2 event streams.
* `DetectAnomalies` evaluates v1 ledgers gracefully using aggregate tool error counts and peak context.
* Running `agentic-lens --ticket-id <ID> --append <SUMMARY.md>` appends an "Anomalies & Health Flags" section if anomalies are detected, or clean status if none.
* CLI table output reflects anomaly counts per ticket.
* `go test -race ./...` and `go vet ./...` pass with zero errors.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Go type checking clean with zero any or empty `interface{}`.
* Telemetry tested with `agentic-lens` CLI.
