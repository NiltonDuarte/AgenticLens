# Contract: AL-1 Go Domain Aggregations, Workspace Scanner, and CLI Engine

## 1. Problem Statement
The existing telemetry aggregation, filesystem scanning, and CLI summary generation in AgenticLens are written in Python. They depend on an external Python 3.14 interpreter, Pydantic schemas, and third-party libraries. This setup incurs startup latency, increases runtime overhead, and prevents single-binary deployment.

The objective of AL-1 is to deliver the foundational Go implementation providing full functional parity with the Python domain engine and CLI. This includes the pure Go domain aggregations, boundary repository scanner for workspace telemetry facts, JSON boundary serialization, idempotent Markdown summary updating, and the standalone CLI executable.

## 2. Functional Core, Imperative Shell (FCIS) Invariants
All implementation must strictly comply with TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
Dependencies flow strictly inward: CLI entrypoint in `cmd/agentic-lens` depends on adapters and domain; adapters in `internal/adapters` depend on domain; domain in `internal/domain` has zero external dependencies and imports no outer packages.

### Functional Core Isolation
Package `internal/domain` must contain pure mathematical rollups, context window health evaluations, token formatting utilities, and immutable data structures.
The domain layer must not import disk or network packages such as `os`, `net/http`, or `database/sql`. It must not invoke system clocks (such as `time.Now`) or random number generators. All inputs are provided as parameters, and all outputs are returned as value structures or explicit errors.
Every mathematical rollup must defensively handle zero values, including empty ledgers, zero durations, zero tokens, zero turns, zero tool calls, and zero production code lines.

### Imperative Boundary Shell
Package `internal/adapters` must isolate all filesystem I/O, path resolution, directory traversal, JSON parsing, schema validation, and summary file modification.
External JSON parsing must use boundary schema transfer objects before transforming them into domain entities.
Errors must be caught and wrapped using `fmt.Errorf` with descriptive context and the `%w` verb. Non-fatal scan warnings for corrupted individual files must be accumulated in scan reports without terminating the overall scan.
Path traversal attacks must be prevented when accessing workspace artifacts.

### Presentation Shell
Package `cmd/agentic-lens` must implement the command line interface using standard library flag parsing.
CLI commands must support listing workspaces, filtering by ticket identifier, generating schema-compliant JSON, and idempotently updating summary Markdown tables.
All log messages, console output, and Markdown tables must have zero emojis.

## 3. Scope Boundaries: Go Backend and TypeScript Frontend
AL-1 focuses strictly on the Go backend foundation. It does not introduce TypeScript frontend code or HTTP server routes.

### In Scope for AL-1
- Go module initialization with Go 1.23 or newer.
- Pure Go domain models, constants, enums, and rollup functions in `internal/domain`.
- Boundary JSON schema serialization and deserialization in `internal/adapters`.
- Filesystem repository scanner for workspace telemetry facts in `internal/adapters`.
- Markdown summary table formatter and idempotent file updater in `internal/adapters`.
- CLI application in `cmd/agentic-lens` supporting flags for workspaces directory, ticket identifier, list output, json output, and append output.
- Comprehensive table-driven unit tests in `internal/domain` with zero mocks, and integration tests in `internal/adapters` using temporary test directories.

### Out of Scope for AL-1
- HTTP REST API endpoints and web server routes (deferred to AL-2).
- Single-binary static asset embedding of `web/dist` (deferred to AL-2 and AL-5).
- TypeScript SPA dashboard, React components, and Vite tooling (deferred to AL-3 and AL-4).
- Deletion or deprecation of legacy Python files (deferred to AL-5).

## 4. Architectural Contracts and Data Structures

### Domain Entities (`internal/domain`)
- `InvocationTokens`: Immutable token counter storing input, output, cache read, reasoning, and peak context token amounts as 64-bit integers.
- `InvocationFact`: Atomic execution fact storing invocation ID, optional parent ID, role, model, user, started timestamp, completed timestamp, active duration in seconds, turns, tokens, cost in USD as float64, and compaction count.
- `ToolFact`: Tool usage record storing tool name, total calls, and error count.
- `CodeFacts`: Code modification metrics storing lines added, lines deleted, test lines added, test lines deleted, prod lines added, prod lines deleted, files changed, and SDLC iterations count.
- `AgenticFactLedger`: Root raw session facts ledger containing ticket ID, ticket title, ticket type, started timestamp, completed timestamp, updated timestamp, code facts, invocation facts slice, and tool facts slice.
- `GroupedMetric`: Aggregate metric record containing invocation count, input tokens, output tokens, cache read tokens, cache hit rate percentage, cost in USD, peak context tokens, and context health string.
- `SubagentTurnVelocity`: Velocity metrics storing average turns per invocation, peak turns, and average turns grouped by role.
- `TestToCodeRatio`: Test to production code ratio string representation and raw line counts.
- `CodeDeliveredStats`: Aggregate delivered code counts for added lines, deleted lines, and changed files.
- `ToolExecutionHealth`: Total tool calls, error and correction count, and success rate percentage.
- `ContextWindowHealth`: Healthy ceiling tokens constant, overall status string, orchestrator peak tokens, orchestrator saturation percentage, subagent peak tokens, subagent peak role, compaction count, and peak tokens by role map.
- `WorkflowEfficiency`: Aggregate efficiency metrics combining SDLC iterations, active duration seconds, formatted active duration string, role effort distribution percentage map, subagent turn velocity, test to code ratio, reasoning tokens, reasoning ratio percentage, code delivered stats, tool execution health, and context window health.
- `ModelMetric` and `RoleMetric`: Breakdown records capturing invocation counts, tokens, cache hit rates, cost in USD, peak context tokens, and health labels grouped by model and role respectively.
- `ThreadRoleMetric` and `ThreadExecutionRecord`: Hierarchical session structures grouping invocations by root session identifier.
- `AgenticMetricsLedger`: Full projected ledger combining fact ledger, workflow efficiency, thread execution records, totals, roles breakdown, and models breakdown.
- `ScanReport`: Scanned ledgers slice and non-fatal scan error messages slice.
- `TicketArtifacts`: Container for ticket ID, CONTRACT.md content, and SUMMARY.md content.

### Rollup Functions (`internal/domain`)
- `ComputeTotals`: Aggregates total invocations by counting distinct session IDs per role, sums input, output, and cache read tokens, calculates cache hit rate percentage, sums total cost rounded to 4 decimals, derives peak context tokens, and computes base context health status.
- `ComputeWorkflowEfficiency`: Calculates cumulative active duration, formatted duration (hours and minutes or minutes and seconds), subagent role effort distribution percentages based on cost, subagent turn velocity, test-to-production ratio string, reasoning token ratio percentage, code delivered statistics, tool success rate percentage, and context window health metrics against the 200,000 token ceiling.
- `ComputeRoleMetrics`: Groups invocations by role in canonical order (orchestrator, planner, coder, tester, reviewer) followed by custom roles, calculating tokens, cache hit rate, cost, peak context, and granular health labels.
- `ComputeModelMetrics`: Groups invocations by model, aggregating tokens, cache hit rate, cost, peak context, and health label, sorted by cost in descending order.
- `ComputeThreadSummaries`: Groups invocations by root session ID via recursive parent traversal, grouping nested roles and models.
- `DeriveMetricsLedger`: Orchestrates all pure rollup functions to produce an AgenticMetricsLedger from an AgenticFactLedger.
- `CalculateFleetSummary`: Computes fleet-wide rollups across multiple metrics ledgers.
- `BuildComparisonRows`: Normalizes multiple ledgers into comparison records sorted by completion date.

### Boundary Repositories and Presenters (`internal/adapters`)
- `ResolveWorkspacesDir`: Resolves workspace directory path inspecting parameter override, container volume (`/app/workspaces`), config file (`.agenticlens/config.json`), current working directory, and parent traversal.
- `ScanWorkspaces`: Recursively globs or walks directories for `agentic_metrics.json`, unmarshals JSON via adapter transfer objects, derives ticket ID from parent directory if omitted, applies domain rollup derivations, and returns a `ScanReport`.
- `ReadTicketArtifacts`: Safely loads CONTRACT.md and SUMMARY.md ensuring paths remain strictly inside the workspaces directory boundary to avoid directory traversal exploits.
- `FormatMarkdownMetricsTable`: Reconstructs the exact Markdown metrics table matching the Python formatter, detailing workflow efficiency and role/model tables.
- `UpdateSummaryFile`: Idempotently replaces or appends the metrics table inside an existing or new SUMMARY.md file by targeting the Workflow Efficiency or Agentic Execution Metrics section header.

## 5. Quality Standards and Verification Gates
- Clean compilation using Go 1.23 or newer.
- Zero any or empty interface across domain, adapters, and CLI. Concrete structs and typed errors only.
- Strict error handling with zero discarded errors.
- Zero mocks in domain tests; pure value assertions only.
- Passing `go vet ./...` and `go test -race ./...` with zero failures and zero race conditions.
- Strict zero emojis invariant across all generated files, strings, CLI output, and error messages.
