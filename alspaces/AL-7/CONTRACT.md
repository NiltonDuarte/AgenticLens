# Contract: AL-7 Domain Anomaly Detection and Rule Engine

## 1. Problem Statement

Multi-agent development workflows encounter execution degradation patterns that are invisible to high-level metric aggregates:
1. Tool Retry Streaks: Subagents getting caught in repetitive tool execution failure loops (for example, repeated failing bash invocations or invalid file edit arguments).
2. High Tool Failure Rates: Tools exhibiting persistent error frequencies above acceptable quality thresholds across an entire development session.
3. Context Saturation Spikes: Sudden single-turn token jumps that risk overflowing the context window or breaching the critical saturation ceiling.
4. Subagent Rework Loops: Ping-pong handoff thrashing where worker roles (such as coder and tester) alternate repeatedly without making progress.

Ticket AL-6 established chronological turn-level event extraction and schema evolution in Go, storing ordered EventFact records in the AgenticFactLedger. Ticket AL-7 leverages this telemetry to build a pure functional anomaly detection engine in the Go domain. The engine evaluates deterministic rules against both v2 event streams and legacy v1 aggregate ledgers, surfaces health flags in CLI summaries, and maintains full contract parity with the TypeScript frontend.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation across Go backend and TypeScript frontend must adhere strictly to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. CLI Subcommand Layer (backend/cmd/agentic-lens/main.go):
   - Handles CLI flag parsing (--anomalies, --list, --append), console output rendering, and adapter invocation.
   - Contains zero anomaly detection algorithms or calculation heuristics.
2. Imperative Adapters Layer (backend/internal/adapters/):
   - Maps domain anomaly structs to serialization DTOs (backend/internal/adapters/schemas.go).
   - Formats anomaly diagnostics into human-readable Markdown review tables (backend/internal/adapters/formatters.go).
   - Handles file updates for summary markdown (UpdateSummaryFile).
   - Contains zero heuristic threshold definitions or rule evaluation loops.
3. Functional Core Domain Layer (backend/internal/domain/):
   - Houses pure anomaly models (Anomaly, AnomalySeverity, AnomalyRuleID, AnomalyContext) and rule evaluation logic (DetectAnomalies).
   - Remains 100% pure functional: zero disk I/O, zero network calls, zero external processes, and zero system clocks.
   - Operates purely on immutable structs passed as value parameters and returns deterministic slices of anomalies.

### Pure Domain Core Isolation
1. AnomalySeverity is defined as an immutable typed string enum in backend/internal/domain/enums.go with values WARNING and CRITICAL.
2. AnomalyRuleID is defined as an immutable typed string enum in backend/internal/domain/enums.go with explicit constants:
   - RuleToolRetryStreak = "TOOL_RETRY_STREAK"
   - RuleToolHighFailureRate = "TOOL_HIGH_FAILURE_RATE"
   - RuleContextSaturationSpike = "CONTEXT_SATURATION_SPIKE"
   - RuleSubagentReworkLoop = "SUBAGENT_REWORK_LOOP"
3. DetectAnomalies(ledger AgenticFactLedger) []Anomaly is a pure deterministic function. Given identical ledger facts, it must return an identical, deterministically sorted slice of Anomaly value objects.
4. Graceful v1 Fallback Mandate: When ledger.Events is nil or empty (legacy v1 ledgers), DetectAnomalies must gracefully evaluate rules that can operate on aggregate metrics (such as tool error rates and peak context tokens) while skipping sequence-dependent rules without panic, error, or data loss.
5. DeriveMetricsLedger in backend/internal/domain/aggregations.go attaches evaluated anomalies to AgenticMetricsLedger immutably. Existing mathematical rollup calculations remain unmodified.

### Imperative Shell Boundary Rules
1. Backward-Compatible Serialization: AnomalyDTO in backend/internal/adapters/schemas.go must use pointer semantics and standard JSON tags (omitempty) so that older consumers ignore the field without error.
2. Idempotent Summary Formatting: FormatMarkdownMetricsTable and UpdateSummaryFile must format an "Anomalies & Health Flags" section. If zero anomalies are detected, a clean session banner is rendered. If anomalies exist, each anomaly is formatted with severity badges, rule identifier, and context descriptions.
3. Zero-Emoji Mandate: Zero emojis across all log lines, CLI table columns, summary markdown outputs, JSON schemas, code comments, and documentation.
4. Static Single-Binary and Zero Any: Single Go binary distribution without CGO or Node.js runtime dependencies. Zero any or empty interface{} across all Go and TypeScript source files.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Domain Core Models and Rules (backend/internal/domain/):
   - Define AnomalySeverity and AnomalyRuleID enums in backend/internal/domain/enums.go.
   - Define AnomalyContext and Anomaly structs in backend/internal/domain/anomalies.go.
   - Extend AgenticMetricsLedger in backend/internal/domain/models.go with Anomalies []Anomaly.
   - Implement pure DetectAnomalies(ledger AgenticFactLedger) []Anomaly in backend/internal/domain/anomalies.go.
   - Integrate anomaly detection into DeriveMetricsLedger in backend/internal/domain/aggregations.go.
2. Boundary Adapters and Serialization (backend/internal/adapters/):
   - Implement AnomalyDTO and AnomalyContextDTO in backend/internal/adapters/schemas.go with bidirectional domain conversion.
   - Extend AgenticMetricsLedgerDTO with Anomalies []AnomalyDTO (json tag anomalies,omitempty).
   - Update FormatMarkdownMetricsTable in backend/internal/adapters/formatters.go to render the Anomalies and Health Flags section.
3. CLI Interface Integration (backend/cmd/agentic-lens/main.go):
   - Add --anomalies flag to CLI root and generate commands to output detailed anomaly diagnostics to stdout.
   - Update --list terminal table output to include an Anomalies count column.
   - Ensure generate and default commands update SUMMARY.md with anomaly findings when --append is specified.
4. TypeScript Presentation Layer Parity (web/src/domain/types.ts):
   - Add AnomalySeverity, AnomalyRuleID, AnomalyContext, and Anomaly interfaces to TypeScript domain types.
   - Extend AgenticMetricsLedger in TypeScript with optional readonly anomalies?: readonly Anomaly[].
5. Comprehensive Test Suites:
   - Unit tests in backend/internal/domain/anomalies_test.go covering clean runs, individual rule triggers, v1 fallbacks, multi-anomaly ordering, and empty inputs. Zero mocks.
   - Formatter tests in backend/internal/adapters/formatters_test.go verifying markdown output formatting.
   - CLI integration tests in backend/cmd/agentic-lens/main_test.go verifying --anomalies and --list table columns.
6. Documentation and Architectural Overview (README.md):
   - Update repository documentation with existing capabilities:
     - Multi-agent telemetry extraction from OpenCode SQLite databases and git repository metadata.
     - Dual-schema fact and metric ledgers, spanning v1 aggregate models and v2 chronological event streams.
     - Automated backfill engine for historical workspace telemetry reconciliation.
     - Single-binary distribution with embedded TypeScript web UI.
     - CLI command suite: generate, backfill, server, scan, and list.
   - Update repository documentation with new AL-7 anomaly detection features:
     - Pure functional rule engine detecting tool retry streaks, tool failure rates, context saturation spikes, and subagent rework loops.
     - CLI inspection with the --anomalies flag and the list view anomalies column.
     - Automated Markdown summary health section formatting in workspace deliverables.

### Out of Scope
1. HTTP REST endpoints for anomalies (GET /api/workspaces/:ticketId/anomalies) and fleet rollups (ComputeFleetAnomalies), which belong to AL-8.
2. Web UI waterfall visualization, timeline event playback, and interactive anomaly inspector components, which belong to AL-9.
3. Database mutations or backfill schema changes (all event data is already populated by AL-6).

---

## 4. Anomaly Detection Engine Architecture and Rule Definitions

The anomaly detection engine evaluates four deterministic rules against the fact ledger:

### Rule 1: TOOL_RETRY_STREAK
- Description: Detects consecutive execution failures of the same tool in the chronological event stream.
- Rule ID: TOOL_RETRY_STREAK
- Evaluation Precondition: Ledger must contain populated Events slice (v2 ledger).
- Evaluation Logic:
  1. Filter events where EventType is tool_call, preserving chronological order.
  2. Traverse events tracking the active tool name and consecutive error count.
  3. When an event with Status equal to error is encountered:
     - If the tool matches the previous failed tool, increment the consecutive failure counter.
     - If the tool differs, reset the streak to 1 for the new tool.
  4. When an event with Status equal to completed is encountered for that tool, reset the consecutive failure counter to 0.
  5. If the consecutive failure counter reaches 3 or more:
     - Generate an Anomaly.
     - Severity is CRITICAL if streak reaches 5 or more failures; WARNING if streak is 3 or 4 failures.
     - Context captures ToolName, Count (streak length), SessionID, and Turn of the latest error.
- v1 Fallback: Skipped gracefully when Events is empty (cannot reconstruct consecutive streaks from aggregate totals).

### Rule 2: TOOL_HIGH_FAILURE_RATE
- Description: Detects tools exhibiting an excessively high failure frequency across the entire ticket lifecycle.
- Rule ID: TOOL_HIGH_FAILURE_RATE
- Evaluation Precondition: Minimum of 5 total calls for a given tool.
- Evaluation Logic (v2 Event Streams):
  1. Aggregate total calls and error calls per tool from tool_call events in Events.
  2. For each tool with total calls >= 5, calculate FailureRate = float64(errors) / float64(total_calls).
  3. If FailureRate > 0.20 (20% error rate):
     - Generate an Anomaly.
     - Severity is CRITICAL if FailureRate >= 0.50 (50% or higher); WARNING if FailureRate is between 0.20 and 0.50.
     - Context captures ToolName, Count (error count), TotalCalls, and FailureRatePct.
- v1 Fallback Logic:
  When Events is empty, iterate over ledger.Tools []ToolFact. For each ToolFact with Calls >= 5 and float64(Errors)/float64(Calls) > 0.20, generate the Anomaly using the same threshold and severity criteria.

### Rule 3: CONTEXT_SATURATION_SPIKE
- Description: Detects sudden single-turn context window expansions or breaches of the critical context ceiling.
- Rule ID: CONTEXT_SATURATION_SPIKE
- Thresholds:
  - Single-Turn Jump Threshold: 40,000 tokens (WARNING), 80,000 tokens (CRITICAL).
  - Critical Saturation Ceiling: 160,000 tokens (CRITICAL).
- Evaluation Logic (v2 Event Streams):
  1. Group events by SessionID and sort chronologically.
  2. For each event, compare ContextTokens with the preceding event in that session.
  3. If DeltaTokens = CurrentContext - PreviousContext > 40,000 tokens:
     - Generate an Anomaly.
     - Severity is CRITICAL if DeltaTokens > 80,000 tokens; WARNING otherwise.
     - Context captures SessionID, Turn, TokenDelta, and CurrentTokens.
  4. If CurrentContext >= 160,000 tokens:
     - Generate a separate Anomaly with CRITICAL severity for ceiling breach.
     - Context captures SessionID, Turn, CurrentTokens, and Ceiling (160000).
- v1 Fallback Logic:
  When Events is empty, inspect ledger.Invocations. If any InvocationFact has Tokens.PeakContext >= 160,000 tokens, emit a CRITICAL Anomaly for ceiling breach. Single-turn jumps are skipped gracefully.

### Rule 4: SUBAGENT_REWORK_LOOP
- Description: Detects cyclic ping-pong role transitions between worker subagents indicating thrashing or deadlock.
- Rule ID: SUBAGENT_REWORK_LOOP
- Thresholds: 3 or more repetitive alternating cycles between two identical roles (WARNING); 5 or more cycles (CRITICAL).
- Evaluation Logic (v2 Event Streams and v1 Invocations):
  1. Extract ordered sequence of subagent roles (excluding orchestrator), ordered by timestamp or invocation start time.
  2. Scan for repeating alternating patterns of length 2 (for example, role A -> role B -> role A -> role B).
  3. If consecutive ping-pong cycles reach 3 or more:
     - Generate an Anomaly.
     - Severity is CRITICAL if cycles >= 5; WARNING if cycles is 3 or 4.
     - Context captures PrimaryRole, SecondaryRole, CycleCount, and Threshold.

### Deterministic Anomaly Sorting
To guarantee repeatable outputs and assertions, anomalies returned by DetectAnomalies are sorted deterministically:
1. Primary sort: Severity descending (CRITICAL before WARNING).
2. Secondary sort: RuleID ascending.
3. Tertiary sort: Turn ascending (or ToolName ascending when turn is absent).

---

## 5. Acceptance Criteria

1. DetectAnomalies correctly flags 3 or more consecutive tool errors in v2 event streams with TOOL_RETRY_STREAK.
2. DetectAnomalies correctly flags tools with error rate exceeding 20% (minimum 5 calls) with TOOL_HIGH_FAILURE_RATE for both v2 events and v1 aggregate ledgers.
3. DetectAnomalies correctly flags single-turn context surges exceeding 40k tokens and context ceiling breaches exceeding 160k tokens with CONTEXT_SATURATION_SPIKE.
4. DetectAnomalies correctly flags alternating worker subagent rework loops with SUBAGENT_REWORK_LOOP.
5. Legacy v1 ledgers lacking events are evaluated gracefully without errors or nil pointer panics.
6. Clean sessions without rule violations return an empty slice of anomalies.
7. FormatMarkdownMetricsTable renders an "Anomalies & Health Flags" section in Markdown output: clean status when empty, formatted list with severity badges when anomalies exist.
8. CLI command agentic-lens --ticket-id <ID> --append <SUMMARY.md> idempotently updates SUMMARY.md with the anomaly section.
9. CLI command agentic-lens --list displays an Anomalies column reflecting detected anomaly counts per ticket.
10. CLI flag --anomalies prints detailed anomaly context diagnostics to stdout.
11. TypeScript interfaces in web/src/domain/types.ts mirror Go domain anomaly types with strict typing and zero any.
12. Quality gates pass cleanly:
    - go vet ./... passes with zero warnings.
    - go test -race ./... passes across all packages with zero race conditions.
    - npm run typecheck in web/ passes with zero errors.
13. Strict zero-emoji policy enforced across all code, tests, documentation, and summary markdown outputs.
14. README.md is comprehensively updated documenting existing telemetry extraction from OpenCode SQLite/git, schema v1/v2 with chronological events, backfill engine, single-binary embedded web UI, CLI command suite (generate, backfill, server, scan, list), and new AL-7 anomaly detection features (pure rule engine, CLI --anomalies flag, anomalies column, markdown summary health section), with strictly zero emojis.
