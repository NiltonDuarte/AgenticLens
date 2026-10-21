h3. Title: AL-19: Implement Live Anomaly Detection Engine and Real-Time Session Health Rules

*TARGET*: AgenticLens (domain | adapters | http | tests)

*CONTEXT*

During live development iterations, agents can enter destructive failure states, such as repeated tool error loops, rapid token burn, or context window saturation exceeding the healthy ceiling (200k tokens). Operators need real-time anomaly detection evaluated on live transient session data so issues can be spotted before agents waste budget or fail silently.

*GOAL*

Implement a pure functional real-time anomaly detection engine in Go `internal/domain` to evaluate active sessions against operational health rules (consecutive tool errors, token runaway, context ceiling threshold, tool thrashing), and expose live anomaly indicators through the live HTTP API and SSE stream.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Create live anomaly rules in `internal/domain/live_anomaly.go`:
    - Rule 1: `ConsecutiveToolErrorsRule`: flags sessions where N consecutive tool calls fail (streak >= 3).
    - Rule 2: `ContextCeilingApproachRule`: triggers alerts when context window reaches >= 80% and >= 100% of `HealthyCeilingTokens` (200k).
    - Rule 3: `TokenVelocityRunawayRule`: flags abnormally high token consumption per turn.
    - Rule 4: `ToolThrashingRule`: flags repeated calls to the same failing tool without progress.
  - Define `LiveAnomalyAlert` (`SessionID`, `RuleName`, `Severity`, `Message`, `Timestamp`, `TriggerEvent`).
  - Pure evaluator function: `EvaluateLiveAnomalies(tree *LiveSessionTree, trace *LiveTraceView) []LiveAnomalyAlert`.
* Boundary Adapter & HTTP:
  - Wire anomaly evaluation into `internal/adapters/live/` engine.
  - Include detected live anomalies in `/api/live/sessions` and `/api/live/sessions/:id/trace` responses.
  - Broadcast anomaly alerts across the SSE stream `/api/live/stream`.
* Quality:
  - Table-driven unit tests for each anomaly rule under normal, boundary, and pathological conditions.
  - Zero mocks in domain logic; pure deterministic evaluation.
  - Go static checks with zero `any` and race detector passing.

*ACCEPTANCE CRITERIA*

* Consecutive tool errors immediately trigger a high-severity `ConsecutiveToolErrors` alert.
* Sessions approaching or exceeding the 200k token healthy ceiling generate warning and critical context saturation alerts.
* Anomaly alerts are included in the live session JSON response and pushed through the SSE stream.
* Anomaly evaluation remains purely in-memory and executes with sub-millisecond latency.
* Go checks pass cleanly (`go vet ./...`, `go test -race ./...`).

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Verified anomaly detection with table-driven tests.
* Telemetry tested with `agentic-lens` CLI.
