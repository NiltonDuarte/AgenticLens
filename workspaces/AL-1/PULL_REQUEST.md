# feat(domain): implement Go domain aggregations, workspace scanner, and CLI engine

## Context
Migrates the core telemetry aggregation engine, filesystem scanner, and CLI from Python to native Go as part of ticket **AL-1** in the AgenticLens rewrite roadmap.

## Key Changes
- **Functional Core (`internal/domain`)**:
  - Implemented immutable value structures for token counters, invocations, tools, and code facts.
  - Implemented typed string enums (`ContextHealthStatus`, `AgentRole`, `TimeGranularity`, `TimeWindowPreset`) and canonical role orders.
  - Implemented pure token and health helpers (`FormatTokensK`, `GetBaseHealthStatus`, `GetHealthLabel`) referencing the 200k healthy ceiling.
  - Implemented pure rollups (`ComputeTotals`, `ComputeWorkflowEfficiency`, `ComputeRoleMetrics`, `ComputeModelMetrics`, `ComputeThreadSummaries`, `DeriveMetricsLedger`, `CalculateFleetSummary`, `BuildComparisonRows`, `AggregateMetricsOverTime`).
  - Strict FCIS adherence: zero I/O imports, zero system clocks, defensive zero-division guarding.
- **Boundary Adapters (`internal/adapters`)**:
  - Boundary JSON DTOs (`schemas.go`) mapping input ledgers to domain entities.
  - Workspace scanner (`repository.go`) with multi-strategy directory resolution, non-fatal parsing error collection, and path traversal protection.
  - Markdown summary table formatter and idempotent section updater (`formatters.go`).
- **CLI Executable (`cmd/agentic-lens`)**:
  - Standalone CLI supporting `--list`, `--ticket-id`, `--workspaces`, `--append`, and `--json-output`.
- **Build & Developer Tooling (`Makefile`, `.opencode/`)**:
  - Root `Makefile` with targets: `all`, `check`, `test`, `test-backend`, `test-frontend`, `vet`, `lint`, `build`, `clean`.
  - Updated skills (`agentic-lens-dev`), guidelines, and agent definitions (`orchestrator`, `coder`, `tester`, `reviewer`) to standardize on `make`.
- **Quality & Invariants**:
  - Zero `any` or empty `interface{}` across domain, adapters, and CLI.
  - Zero emojis across code, logs, and presentation.

## References
- Contract: [workspaces/AL-1/CONTRACT.md](workspaces/AL-1/CONTRACT.md)
- Summary: [workspaces/AL-1/SUMMARY.md](workspaces/AL-1/SUMMARY.md)
