# Implementation Plan: AL-1 Go Domain Aggregations, Workspace Scanner, and CLI Engine

## 1. Goal and Iteration Scope
The goal of iteration 01 is to establish the Go backend foundation for AgenticLens. This iteration delivers the pure functional domain core, boundary JSON schemas, workspace filesystem repository scanner, Markdown summary formatter, and the standalone CLI executable.

Upon completion of this plan, the Go CLI will compile cleanly to a native binary, scan workspaces, serialize computed telemetry JSON, list discovered workspaces, and idempotently update summary Markdown files with full parity to the legacy Python implementation.

## 2. Grounded Reference Architecture
The implementation directly ports and refines the following existing Python components into Go idiomatic structures:
- Models: `agentic_lens/domain/models.py` to `internal/domain/models.go`
- Enums & Constants: `agentic_lens/domain/enums.py` and `agentic_lens/domain/constants.py` to `internal/domain/enums.go` and `internal/domain/constants.go`
- Pure Aggregations: `agentic_lens/domain/aggregations.py` to `internal/domain/aggregations.go`
- JSON Schemas: `agentic_lens/adapters/schemas.py` to `internal/adapters/schemas.go`
- Repository & Scanner: `agentic_lens/adapters/repository.py` to `internal/adapters/repository.go`
- Markdown Formatter: `agentic_lens/presentation/views/formatters.py` to `internal/adapters/formatters.go`
- CLI Command: `agentic_lens/presentation/cli.py` to `cmd/agentic-lens/main.go`

## 3. Phased Implementation Tasks

### Phase 1: Go Module Initialization and Project Layout
- Task 1.1: Initialize the Go module in the repository root with module identifier `agenticlens` and targeting Go 1.23. Verify `go.mod` creation.
- Task 1.2: Create package directory structure matching technical guidelines:
  - `internal/domain`
  - `internal/adapters`
  - `cmd/agentic-lens`
- Task 1.3: Verify gitignore configurations ensure Go build artifacts and binaries (such as `bin/`) are ignored.

### Phase 2: Functional Core Domain Implementation (`internal/domain`)
- Task 2.1: Implement `internal/domain/enums.go`:
  - Define `ContextHealthStatus` type as a typed string with constants for `HEALTHY`, `ELEVATED`, `BLOATED`, and `CRITICAL`.
  - Define `AgentRole` type as a typed string with constants for `orchestrator`, `planner`, `coder`, `tester`, and `reviewer`.
  - Define canonical role ordering slice and subagent role ordering slice as package-level immutable variables.
- Task 2.2: Implement `internal/domain/constants.go`:
  - Define `HealthyCeilingTokens` as constant integer 200,000.
  - Implement pure helper `FormatTokensK` converting integer token counts into human-readable strings (such as 900.8k, 200k, 1.2M, or raw string for values below 1000).
  - Implement pure helper `GetBaseHealthStatus` returning `HEALTHY`, `ELEVATED`, `BLOATED`, or `CRITICAL` based on token thresholds.
  - Implement pure helper `GetHealthLabel` returning formatted string with saturation percentage for `BLOATED` and `CRITICAL` categories.
- Task 2.3: Implement `internal/domain/models.go`:
  - Define immutable structs with value semantics: `InvocationTokens`, `InvocationFact`, `ToolFact`, `CodeFacts`, `AgenticFactLedger`, `GroupedMetric`, `TestToCodeRatio`, `SubagentTurnVelocity`, `CodeDeliveredStats`, `ToolExecutionHealth`, `ContextWindowHealth`, `WorkflowEfficiency`, `ModelMetric`, `RoleMetric`, `ThreadRoleMetric`, `ThreadExecutionRecord`, `AgenticMetricsLedger`, `TimeBucketMetric`, `FleetSummary`, `TicketComparisonRow`, `ScanReport`, and `TicketArtifacts`.
  - Use 64-bit integer types (`int64`) for token counts, byte counts, line counts, turn counts, and timestamp seconds to prevent integer overflow. Use `float64` for costs and percentages.
- Task 2.4: Implement pure aggregation algorithms in `internal/domain/aggregations.go`:
  - Implement `ComputeTotals`: handle empty slice guard; count invocations via unique session IDs per role; calculate total input, output, cache read tokens, and cost; compute cache hit rate percentage; evaluate overall context health.
  - Implement `ComputeWorkflowEfficiency`: compute cumulative duration; format duration string; calculate subagent effort distribution percentages; derive turn velocities; compute test-to-production ratio; compute reasoning token percentage; derive tool execution success rate; calculate context window saturation against ceiling; guard all denominators against zero.
  - Implement `ComputeRoleMetrics`: group by role according to canonical role order; calculate totals, cache hit rate, cost, peak context, and health label for each role.
  - Implement `ComputeModelMetrics`: group by model name; calculate totals and sort by cost descending.
  - Implement `ComputeThreadSummaries`: construct root session groups by walking parent ID relations; group nested roles and models.
  - Implement `DeriveMetricsLedger`: compose totals, efficiency, role metrics, model metrics, and thread summaries into an `AgenticMetricsLedger`.
  - Implement `CalculateFleetSummary`: aggregate across multiple ledgers.
  - Implement `BuildComparisonRows`: map ledgers into comparison rows sorted chronologically.

### Phase 3: Boundary Adapters Implementation (`internal/adapters`)
- Task 3.1: Implement `internal/adapters/schemas.go`:
  - Define JSON data transfer objects matching the exact JSON structure of `agentic_metrics.json`.
  - Implement transfer object validation and conversion methods that map DTOs into pure `internal/domain` structs.
  - Define JSON serialization DTOs for the computed metrics ledger matching the format expected by consumers and legacy CLI.
- Task 3.2: Implement `internal/adapters/repository.go`:
  - Implement `ResolveWorkspacesDir` function: evaluate override path argument, check `/app/workspaces`, read `.agenticlens/config.json` using parent traversal, fall back to `./workspaces`, and parent `/workspaces`.
  - Implement `ScanWorkspaces` function: scan target workspaces directory for `agentic_metrics.json` files; parse and validate JSON; infer ticket ID from parent directory if empty; invoke domain `DeriveMetricsLedger`; aggregate non-fatal file errors into `ScanReport.Errors` without crashing.
  - Implement `ReadTicketArtifacts` function: sanitize ticket ID and resolve path; verify resolved path is within workspaces root to prevent path traversal; load `CONTRACT.md` and `SUMMARY.md` if present.
- Task 3.3: Implement `internal/adapters/formatters.go`:
  - Implement `FormatMarkdownMetricsTable` function: reconstruct the exact Markdown sections (Workflow Efficiency, Agentic Execution Metrics, Contributors, By Subagent Role table, By Model table) with identical column headers, number formatting, token abbreviations, and currency values.
  - Implement `UpdateSummaryFile` function: safely locate existing metrics section headers (Workflow Efficiency or Agentic Execution Metrics) in the target summary file; idempotently replace section or append at the end; create parent directories if missing.

### Phase 4: CLI Application Entrypoint (`cmd/agentic-lens`)
- Task 4.1: Implement `cmd/agentic-lens/main.go`:
  - Configure standard library flag set with options: `--ticket-id` (string), `--workspaces` (string), `--append` (string), `--json-output` (string), and `--list` (boolean).
  - Implement command routing:
    - If `--list` is passed: scan workspaces, format and print tabular overview of ticket ID, duration, cost, health status, SDLC iterations, and title.
    - If `--ticket-id` is passed: select matching ledger from scan report; return error if not found.
    - If no `--ticket-id`: select the most recently updated ledger.
    - If `--append` is passed: update or append metrics table into specified markdown file.
    - If `--json-output` is passed: serialize computed metrics ledger to formatted JSON at specified path.
    - If neither --append nor --json-output is passed: print formatted Markdown metrics table to standard output.
  - Implement clean exit codes: exit 0 on success, exit 1 on invalid arguments or missing data.
  - Verify zero emojis across all output text, table headers, and error messages.

### Phase 5: Testing and Verification
- Task 5.1: Create unit tests in `internal/domain/aggregations_test.go`:
  - Table-driven tests for `ComputeTotals`: empty invocation list, single invocation, multiple invocations across roles, zero tokens, zero cost.
  - Table-driven tests for `ComputeWorkflowEfficiency`: zero prod lines (test-to-code ratio with zero denominator), zero tool calls, zero turns, duration formatting boundaries (seconds only, minutes and seconds, hours and minutes).
  - Table-driven tests for `ContextWindowHealth`: thresholds below 100k (HEALTHY), 100k-200k (ELEVATED), above 200k (BLOATED with percentage), above 400k (CRITICAL with percentage).
  - Table-driven tests for `ComputeRoleMetrics` and `ComputeModelMetrics`: sorting orders, tie-breaking, canonical ordering.
  - Table-driven tests for `CalculateFleetSummary` and `BuildComparisonRows`.
  - Zero mocks in domain tests.
- Task 5.2: Create unit tests in `internal/domain/constants_test.go`:
  - Tests for `FormatTokensK` (values under 1000, thousands with exact and decimal suffixes, millions with exact and decimal suffixes).
  - Tests for `GetBaseHealthStatus` and `GetHealthLabel`.
- Task 5.3: Create unit and integration tests in `internal/adapters/repository_test.go`:
  - Test `ResolveWorkspacesDir` with configuration file discovery and override arguments.
  - Test `ScanWorkspaces` with valid workspace fixtures in `t.TempDir()`.
  - Test `ScanWorkspaces` error handling with malformed JSON fixtures (assert non-fatal warning collection).
  - Test `ReadTicketArtifacts` path traversal prevention.
- Task 5.4: Create unit tests in `internal/adapters/formatters_test.go`:
  - Test `FormatMarkdownMetricsTable` matching expected output.
  - Test `UpdateSummaryFile` idempotent replacement and append behavior with existing headers and new files.
- Task 5.5: Run verification suite:
  - Execute `go vet ./...`
  - Execute `go test -v -race ./...`
  - Compile native binary with `go build -o bin/agentic-lens ./cmd/agentic-lens`
  - Execute `bin/agentic-lens` against existing `workspaces/PO-1/agentic_metrics.json` to verify end-to-end parity with the Python script.

## 4. Key Risks and Mitigations
- Risk: Integer division by zero during ratio or percentage calculations when metrics have zero values.
  Mitigation: Explicit guards in pure domain functions defaulting ratios and percentages to 0.0 before performing division.
- Risk: Path traversal when reading ticket artifacts via arbitrary user input.
  Mitigation: Strictly enforce `filepath.Clean` and verify resolved path starts with workspaces directory boundary using `filepath.Rel`.
- Risk: Formatting discrepancies between Python and Go Markdown tables causing noisy git diffs on SUMMARY.md.
  Mitigation: Directly mirror string formats, column headers, number delimiters (commas), currency notation, and token abbreviations in `formatters.go`, verified with character-by-character unit tests against real summary files.
- Risk: Accidental introduction of third-party dependencies or external mocks.
  Mitigation: Restrict implementation entirely to Go standard library (`encoding/json`, `flag`, `fmt`, `os`, `path/filepath`, `strings`, `time`, `testing`). Zero external Go modules required for AL-1.
