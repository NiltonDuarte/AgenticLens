h3. Title: AL-1: Implement Go Domain Aggregations, Workspace Scanner, and CLI Engine

*TARGET*: AgenticLens (domain | adapters | cli | tests)

*CONTEXT*

The Python CLI and domain logic in `agentic_lens/domain` and `agentic_lens/adapters` compute session rollups and update summary Markdown files. We need the foundational Go implementation that executes these pure aggregations and provides CLI parity with the Python script.

*GOAL*

Deliver the pure Go functional domain model, repository filesystem scanner, and CLI executable capable of scanning workspaces, serializing JSON, and idempotently updating summary Markdown tables.

*VERTICAL SLICE SCOPE*

* Domain Core: Pure structs and functions in `internal/domain/`: models (`AgenticFactLedger`, `AgenticMetricsLedger`, `WorkflowEfficiency`, `GroupedMetric`), constants/thresholds (`HEALTHY_CEILING_TOKENS = 200000`), rollups (`ComputeTotals`, `ComputeWorkflowEfficiency`, `ComputeRoleMetrics`, `ComputeModelMetrics`, `BuildMetricsLedger`, `ComputeFleetSummary`).
* Boundary Adapter: `internal/adapters/fsrepo`: directory walker traversing `workspaces/*/agentic_metrics.json`, strict JSON unmarshaling, loading `SUMMARY.md`.
* Presentation: `cmd/agentic-lens/main.go` with CLI flags: `--ticket-id`, `--workspaces`, `--append`, `--json-output`, and `--list`. Markdown table formatter.
* Quality: Table-driven unit tests in `internal/domain` covering zero division, empty slices, role aggregations, and markdown table rendering. Zero domain mocks.

*ACCEPTANCE CRITERIA*

* `go build ./cmd/agentic-lens` compiles cleanly.
* Running `agentic-lens --list` prints tabular summary of scanned workspaces matching the schema.
* Running `agentic-lens --ticket-id <ID> --append <SUMMARY.md>` updates the metrics table section idempotently.
* Running `agentic-lens --json-output <path>` produces schema-compliant JSON matching the existing `AgenticMetricsLedger` contract.
* `go test -race ./...` and `go vet ./...` pass with zero errors.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Go type checking clean with zero any or empty `interface{}`.
* Telemetry tested with `agentic-lens` CLI.
