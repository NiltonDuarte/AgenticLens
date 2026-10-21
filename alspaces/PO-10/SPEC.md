# Feature Specification: Tool Execution vs Content Error Disentanglement and Pipeline Parity (AL-42 & AL-43)

## Problem Statement

In the current AgenticLens architecture, tool failure detection suffers from two critical architectural defects:

1. **Conflation of Execution Failures and Semantic Content Errors**:
   Currently, `domain.ClassifyToolFailure` inspects tool output text strings for error keywords (`Traceback`, `panic:`, `fatal error:`, `cannot find package`, `syntax error`, `Error:`) alongside process exit indicators (`exit status`, `cleanStatus == "error"`).
   If any string match succeeds, the tool invocation is labeled as an error (`isError = true`), and callers overwrite the invocation status to `"error"`.
   This conflates two fundamentally different failure modes:
   - **Hard Tool Execution Failures**: The tool command itself failed to execute, returned an error exit code, crashed with an OS signal, timed out, or had invalid parameters.
   - **Semantic Content Errors in Successful Tool Calls**: The tool process executed cleanly and returned exit code 0 (`status: "completed"`), but its standard output payload contains compilation errors, stack traces, test failures, or lint warnings.
   Furthermore, benign operations such as `grep` or `glob` returning 0 matches (which POSIX grep signals with exit code 1) are explicitly classified as errors (`isEmptySearch`), causing normal search operations to artificially elevate failure counts and trigger false-alarm tool retry streaks.

2. **Divergence Between Offline and Live Classification Pipelines**:
   The offline telemetry extraction engine (`internal/adapters/opencode/extractor.go`) does not use `domain.ClassifyToolFailure` at all. Instead, it relies on an ad-hoc inline conditional:
   ```go
   isError := stat == "error" || strings.Contains(out, "Traceback") || strings.HasPrefix(trimmedOut, "Error:") || strings.Contains(outPrefix, "Error:")
   ```
   Meanwhile, the live observability engine (`internal/adapters/live/service.go`) calls `domain.ClassifyToolFailure`.
   This divergence causes identical agent sessions to report different tool error counts, failure rates, and anomaly states depending on whether they are inspected via real-time live streaming or post-hoc generated workspace telemetry.

Both failure modes provide critical engineering value: hard execution failures indicate brittle tooling, bad environments, or invalid invocations, while content errors reflect agent code mistakes, compilation hurdles, and test failures. However, aggregating them into a single bucket obfuscates telemetry and degrades anomaly precision.

---

## Target Personas & Use Cases

- **AI System Engineers & Agent Developers**:
  Need to distinguish whether an agent is thrashing because a CLI tool is misconfigured/crashing (execution failure) versus the agent iterating through compilation errors and unit test failures (content error).
- **Engineering Managers & Tech Leads**:
  Need accurate fleet-wide tool reliability metrics. An agent compiling code 10 times with syntax errors should not show `bash` as a broken/failing tool in the Brittle Tools leaderboard when `bash` executed exactly as instructed.
- **Autonomous Agent Operators**:
  Monitor live sessions and waterfall traces with immediate visual distinction between process aborts (red badge) and semantic code errors (amber warning badge).

---

## User Stories

1. As an agent developer, I want tool calls that return non-zero exit codes or process crashes to be classified specifically as execution failures, so that I can diagnose infrastructure and tool configuration problems.
2. As an agent developer, I want tool calls that execute cleanly (exit 0) but contain compiler errors or stack traces to be tracked as semantic content errors without marking the tool execution itself as failed.
3. As a platform user, I want offline telemetry extraction and live stream ingestion to produce identical tool metrics for the same session events.
4. As a dashboard viewer, I want the Trace Waterfall and Live Inspector to visually distinguish hard execution crashes from completed calls with content warnings using distinct badging and color schemes.
5. As an engineering lead, I want the Brittle Tools fleet summary and `TOOL_HIGH_FAILURE_RATE` anomaly rule to reflect genuine tool execution reliability rather than compiler diagnostic churn.

---

## Architectural Breakdown: Two Vertical Slices (v1 AL Prefix)

To fix this cleanly in v1 without monolithic risk, the initiative is decomposed into two distinct, vertically sliced Jira tickets:

### Issue 1 (AL-42): Unify Tool Execution Error Classification (Exit Codes & Process Failures) Across Live and Offline Pipelines
- **Scope**:
  - Focus exclusively on hard process execution failures (non-zero exit codes, tool crashes, status `error`/`failed`, unhandled process aborts, OS faults).
  - Refactor `domain.ClassifyToolFailure` to remove output string content heuristics and benign search flags (`isEmptySearch`).
  - Wire `internal/adapters/opencode/extractor.go` to use the shared domain classifier, eliminating pipeline divergence.
  - Update `TOOL_RETRY_STREAK` and `TOOL_HIGH_FAILURE_RATE` anomaly detection rules so they evaluate strictly on true tool execution failures.
  - Provide table-driven Go tests establishing offline/live parity and verifying that benign search commands (`grep`, `glob` with 0 matches) do not register as tool failures.

### Issue 2 (AL-43): Differentiate and Surface Semantic Content Errors in Successful Tool Invocations Across Fleet Analytics and Dashboards
- **Scope**:
  - Establish first-class domain models and metrics for semantic content errors on completed tool calls (`ContentErrors` in `ToolFact`, `HasContentError` / `ContentSeverity` in `EventFact` and `LivePartFact`).
  - Implement pure domain classifier `ClassifyToolContentError` detecting compiler errors, stack traces, and unhandled exceptions in completed output payloads.
  - Update telemetry schemas (`schemas.go`, `dto.go`) with backward-compatible defaults for existing ledgers.
  - Update HTTP endpoints (`/api/workspaces/...`, `/api/live/sessions/.../trace`).
  - Update TypeScript types and Web Dashboard (`TraceWaterfall.tsx`, `LiveTraceInspector.tsx`, `TicketDeepDive.tsx`, `AnomalyDrawer.tsx`) to render dual indicators (Red for Execution Failure, Amber for Content Error) and distinct summary statistics.

---

## Layer Architecture & Component Contracts

### 1. Domain Core (`internal/domain`)

#### Pure Classifier Signatures
```go
// Execution failure classification (AL-42)
type ToolExecutionResult struct {
    IsError      bool
    Severity     FailureSeverity
    ErrorMessage *string
}

func ClassifyToolExecution(toolName string, status string, output string, consecutiveErrors int) ToolExecutionResult

// Semantic content error classification (AL-43)
type ToolContentResult struct {
    HasContentError bool
    Severity        FailureSeverity
    ErrorSnippet    *string
    Category        string // "compilation", "traceback", "runtime_exception", "assertion"
}

func ClassifyToolContent(toolName string, output string) ToolContentResult
```

#### Model Enhancements
- `domain.ToolFact`:
  ```go
  type ToolFact struct {
      Name          string `json:"name"`
      Calls         int64  `json:"calls"`
      Errors        int64  `json:"errors"`         // Hard execution failures
      ContentErrors int64  `json:"content_errors"` // Semantic content errors (AL-43)
  }
  ```
- `domain.EventFact`:
  ```go
  type EventFact struct {
      // Existing fields preserved...
      Status              string  `json:"status"` // "completed", "error"
      ErrorMessage        *string `json:"error_message,omitempty"`
      HasContentError     bool    `json:"has_content_error,omitempty"`
      ContentErrorSnippet *string `json:"content_error_snippet,omitempty"`
  }
  ```
- `domain.LivePartFact`:
  ```go
  type LivePartFact struct {
      // Existing fields preserved...
      IsError             bool    `json:"is_error"`
      ErrorMessage        *string `json:"error_message,omitempty"`
      HasContentError     bool    `json:"has_content_error,omitempty"`
      ContentErrorSnippet *string `json:"content_error_snippet,omitempty"`
  }
  ```

### 2. Adapters & Shell (`internal/adapters`, `internal/adapters/live`, `internal/adapters/opencode`)
- **Offline Extractor (`opencode/extractor.go`)**:
  - Replaces inline ad-hoc parsing with calls to `domain.ClassifyToolExecution` and `domain.ClassifyToolContent`.
  - Guarantees 100% metric parity with live telemetry.
- **Live Service (`live/service.go`)**:
  - Calls `domain.ClassifyToolExecution` for process state and `domain.ClassifyToolContent` for payload inspection.
  - Formats `LiveTraceView` with separated failure counts (`FailedToolCalls` vs `ContentErrorToolCalls`).
- **Telemetry Schema & DTO (`schemas.go`)**:
  - `ToolFactDTO` includes `ContentErrors int64` with default 0 on legacy deserialization.

### 3. Presentation (`web/src`)
- **TypeScript Types (`types/telemetry.ts`, `domain/live.ts`)**:
  - Update interfaces to support `content_errors`, `has_content_error`, and `content_error_snippet`.
- **Trace Waterfall (`TraceWaterfall.tsx`)**:
  - Execution Failure: Red node border, red status badge `EXEC ERROR`.
  - Content Warning: Amber node border, amber status badge `CONTENT ERROR` on completed turns.
- **Live Trace Inspector (`LiveTraceInspector.tsx`)**:
  - Header statistics show both `Failed Tools: X` and `Content Warnings: Y`.
- **Ticket Deep Dive (`TicketDeepDive.tsx`)**:
  - Tool execution table displays separate columns for `Calls`, `Execution Errors`, and `Content Errors`.

---

## Advisor Analysis: Risks, Blind Spots & Probing Questions

### Identified Risks & Blind Spots

1. **Grep and Glob Return Code Conflation**:
   - POSIX tools like `grep` return exit code 1 when no patterns match. In `ClassifyToolFailure`, this was previously flagged as an error under `isEmptySearch`.
   - *Risk*: If exit code 1 is blindly treated as a tool execution failure without tool-specific exit code semantics, normal exploratory searches will still be misclassified.
   - *Mitigation*: The domain execution classifier must explicitly handle standard search tools (`grep`, `glob`, `find`) where exit status 1 with empty output represents 0 matches rather than an execution crash.

2. **Schema Migration & Backward Compatibility**:
   - `workspaces/*/agentic_metrics.json` contains existing serialized `Tools` records without `content_errors`.
   - *Risk*: Strict JSON unmarshaling or invariant verification during schema migrations could fail on missing keys.
   - *Mitigation*: Ensure `content_errors` uses `omitempty` in Go DTOs and defaults to 0 when parsing older schema versions (v1-v6).

3. **Anomaly Rule Sensitivity**:
   - Rules `TOOL_RETRY_STREAK` and `TOOL_HIGH_FAILURE_RATE` currently trigger on `event.Status == "error"`.
   - *Risk*: Separating content errors from `Status` will reduce the count of events with `Status == "error"`.
   - *Benefit*: This is desirable as it eliminates spurious streak warnings caused by compiler iterations. However, we should consider whether a dedicated rule (`COMPILER_RETRY_LOOP` or `CONTENT_ERROR_STREAK`) is needed for agents stuck in compiler loops.

4. **Regex / Substring False Positives in Content Analysis**:
   - If a test passes and prints `Tested 5 cases for syntax error handling`, searching for `syntax error` in stdout could falsely flag the completed call.
   - *Mitigation*: Anchor content error signatures to prefix patterns, structured trace headers, or unhandled exception markers rather than loose word matching.

### Probing Questions for Stakeholders

1. **Tool Exit Code Semantics**: Should exit code 1 for `grep` and `glob` (when returning 0 matches) be treated strictly as a normal completion (`status: "completed"`) across all tools, or should we define a whitelist of exit-code behaviors by tool name?
2. **Anomaly Rule Expansion**: Should `TOOL_RETRY_STREAK` strictly monitor execution crashes, or should we introduce a distinct anomaly rule (e.g. `CONTENT_ERROR_STREAK` or `COMPILATION_LOOP`) specifically for repeated content errors?
3. **Ledger Schema Versioning**: Should the addition of `content_errors` in `ToolFact` trigger a bump to Telemetry Schema v7, or can it be handled as an optional non-breaking backward-compatible field in v6?
