# Contract: FIX-3 Resolution of False Tool Errors, Anomaly Misattributions, Waterfall Collapsibility, and Debug Identifiers

## 1. Problem Statement

Four interrelated operational issues impair operator observability and anomaly investigation during agent execution runs:

1. False Tool Call Execution Errors:
   Content inspection and file-reading tools (specifically read, grep, glob, and webfetch) that complete execution successfully with exit status completed are being scanned for compilation faults, syntax errors, and error prefixes inside their output content within the Go domain classifier in backend/internal/domain/live.go (ClassifyToolFailure), the live service adapter in backend/internal/adapters/live/service.go, and the OpenCode extraction adapter in backend/internal/adapters/opencode/extractor.go. When an agent reads technical documentation containing error examples (such as TECHNICAL_GUIDELINES.md), inspection tests, or log files, the classifier misinterprets the retrieved document text as an execution failure. This mutates status from completed to error, triggers false consecutive failure escalations (for example, escalating to 7 consecutive failures and HIGH severity alerts), and corrupts tool reliability metrics. Successful executions of content-reading and inspection tools must never treat document contents as tool failures.

2. Tool Retry Streak Anomalies Flagging Successes and Misattributing Sessions:
   In the frontend execution waterfall (web/src/components/TraceWaterfall.tsx), the anomaly-matching helper findMatchingAnomaly associates TOOL_RETRY_STREAK (and related tool failure anomalies) with events that completed successfully (status === completed). This occurs because matching logic fails to verify event execution status and matches solely on turn index while ignoring session identifier discrepancies, causing cross-session misattribution. Furthermore, in backend/internal/domain/anomalies.go, detectToolRetryStreaks tracks consecutive errors across session boundaries without resetting on session transitions. In the frontend trace waterfall, tool failure anomalies must strictly attach only to failed tool events (status === error) within the matching session, accurately highlighting failed retries so operators can inspect actual failure details.

3. Execution Trace Waterfall Excessive Height and Lack of Collapsibility:
   The Execution Trace Waterfall section in web/src/components/TraceWaterfall.tsx renders all chronological event rows, filter buttons, and multiselect dropdowns by default. In ticket deep dives containing hundreds of turns, this element dominates the viewport. The section must be collapsible and collapsed by default. While collapsed, it must present a concise summary bar containing the title, total event count, data source badge (Local SQLite versus Committed Ledger), error and anomaly totals summary, and an explicit expand action. When expanded, the full filter toolbar and chronological timeline become visible.

4. Absence of Concrete Debug Identifiers with Copy to Clipboard Actions:
   When an operator identifies an anomalous event, failure alert, or unexpected tool outcome, debugging requires locating the underlying agent message turn in raw logs or the SQLite database. Currently, the UI displays fragmented labels without a unified, copyable debug identifier. The system must expose concrete identifiers (session ID, message ID, turn identifier, or composite debug ID) across event inspection views (web/src/components/AnomalyDrawer.tsx, web/src/components/TraceWaterfall.tsx, and web/src/components/FailedToolAlertCard.tsx) equipped with a one-click Copy to Clipboard button providing temporary feedback.

5. Primitive Obsession Across Declared Domain Enums (Iteration 05 Scope Extension):
   A systematic audit identified that existing declared enums in the domain layer (ContextHealthStatus, AgentRole, FailureSeverity, ConventionalTicketType / StandardTicketType) are bypassed in favor of raw strings across domain structs in both Go and TypeScript. This degrades compile-time safety and introduces risks of transposition errors in function signatures. Existing declared types must be systematically wired into the domain core and presentation boundaries.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must strictly preserve the architectural boundaries specified in TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Functional Core (backend/internal/domain/):
   - In backend/internal/domain/live.go, ClassifyToolFailure remains a pure function operating exclusively on string values and integer counts. When toolName corresponds to a content inspection tool (read, grep, glob, webfetch) and status equals completed, the function immediately returns false, SeverityNone, and a nil error message, bypassing all content error scans.
   - In backend/internal/domain/anomalies.go, detectToolRetryStreaks remains a pure calculation over the immutable slice of EventFact structs. Error streaks must be scoped per session: when event.SessionID changes or event.Status is completed, any active streak is finalized and the streak counter resets to zero.
   - In backend/internal/domain/models.go, the pure EventFact struct is augmented with an optional MessageID field (*string with json tag message_id,omitempty) to preserve concrete message provenance without external dependencies.
   - Zero I/O, zero network calls, zero clock calls, and zero external packages within the domain package.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - In backend/internal/adapters/live/service.go, processRawPart and convertToEventFact propagate MessageID from SQLite PartRow records into domain EventFact and LivePartFact instances. When ClassifyToolFailure returns false for completed content inspection tools, consecutive error counters reset to zero and status remains completed.
   - In backend/internal/adapters/opencode/extractor.go, ExtractFacts maps scanned message_id column values into EventFact.MessageID. When evaluating part error state, content inspection tools (read, grep, glob, webfetch) with status completed must not scan output text for error substrings, keeping status as completed and avoiding false increments to tool error accumulators.
3. TypeScript Domain and Contracts (web/src/types/, web/src/domain/):
   - In web/src/types/telemetry.ts, the EventFact interface is updated with an optional readonly message_id property matching backend serialization.
   - Domain invariants for anomaly rule classification ensure tool failure rules (such as TOOL_RETRY_STREAK and TOOL_HIGH_FAILURE_RATE) are treated as failure-specific rules requiring error status on target events.
4. TypeScript Boundary Presentation Shell (web/src/components/, web/src/views/):
   - In web/src/components/TraceWaterfall.tsx, findMatchingAnomaly enforces two strict predicates: session ID matching whenever the anomaly context defines session_id, and an error status check (event.status === error) for tool failure rules.
   - In web/src/components/TraceWaterfall.tsx, collapsibility state is managed locally with a collapsed-by-default initial state. Collapsed rendering outputs a compact summary header with expand toggle, while expanded rendering exposes full filters and timeline.
   - In web/src/components/AnomalyDrawer.tsx, web/src/components/TraceWaterfall.tsx, and web/src/components/FailedToolAlertCard.tsx, concrete identifiers are displayed alongside copy-to-clipboard buttons utilizing modern navigator.clipboard with fallback to document.execCommand, resetting visual feedback after a 2-second timeout.

### Strong Typing and Zero Any Invariant
1. Zero usage of any, unknown type casts, or empty interface{} across Go backend and TypeScript frontend codebases.
2. Go struct fields and DTOs use concrete scalar types and explicit pointer-to-string fields for optional identifiers.
3. TypeScript definitions use strict readonly properties and precise string literal unions for event statuses and tool classifications.

### Zero Emojis Mandate
Zero emojis permitted across all Go source files, TypeScript components, test suites, comments, docstrings, commit messages, PR drafts, UI elements, badges, tooltips, or workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Go Domain Layer (backend/internal/domain/):
   - Update ClassifyToolFailure in backend/internal/domain/live.go to prevent false error classification on completed executions of read, grep, glob, and webfetch tools.
   - Update detectToolRetryStreaks in backend/internal/domain/anomalies.go to enforce per-session scoping of consecutive tool failure streaks.
   - Add MessageID (*string) to EventFact in backend/internal/domain/models.go.
   - Update and expand unit tests in backend/internal/domain/live_test.go, backend/internal/domain/anomalies_test.go, and backend/internal/domain/events_test.go.
2. Go Adapters Layer (backend/internal/adapters/):
   - Update processRawPart and convertToEventFact in backend/internal/adapters/live/service.go to map message ID and respect pure classifier invariants.
   - Update ExtractFacts in backend/internal/adapters/opencode/extractor.go to map message ID and bypass content error scanning for completed inspection tools.
   - Update unit tests in backend/internal/adapters/live/sqlite_test.go and backend/internal/adapters/opencode/extractor_test.go.
3. TypeScript Telemetry Types (web/src/types/):
   - Add readonly message_id?: string | null | undefined to EventFact in web/src/types/telemetry.ts.
4. TypeScript Presentation Components (web/src/components/, web/src/views/):
   - Update findMatchingAnomaly in web/src/components/TraceWaterfall.tsx to enforce session ID matching and restrict tool failure anomalies strictly to failed events.
   - Implement collapsible section pattern in web/src/components/TraceWaterfall.tsx, collapsed by default, displaying concise summary metrics when collapsed and full interactive controls when expanded.
   - Add concrete debug identifiers and Copy to Clipboard controls in web/src/components/TraceWaterfall.tsx, web/src/components/AnomalyDrawer.tsx, and web/src/components/FailedToolAlertCard.tsx.
   - Update and expand test suites in web/src/components/TraceWaterfall.test.tsx, web/src/components/AnomalyDrawer.test.tsx, web/src/components/FailedToolAlertCard.test.tsx, and web/src/views/TicketDeepDive.test.tsx.
5. Domain Typing & Primitive Obsession Remediation (Iteration 05 Extension):
   - In Go domain (backend/internal/domain/): wire ContextHealthStatus, AgentRole, FailureSeverity, and ConventionalTicketType into domain structs. Update OrderRoles to accept generic string types. Return FailureSeverity directly from ClassifyToolFailure.
   - In Go adapters (backend/internal/adapters/): align schemas DTO converters, git inference helpers, live service, and opencode extractor.
   - In TypeScript domain (web/src/domain/, web/src/types/): define and export CanonicalAgentRole and autocomplete union AgentRole, export FailureSeverity, align StandardTicketType and TicketType, and update domain interfaces to use ContextHealthStatus, AgentRole, FailureSeverity, and TicketType.
   - In TypeScript presentation components (web/src/components/): update TicketSelector and FailedToolAlertCard props and alignment.

### Out of Scope
1. Modifying SQLite schema migrations or altering database column definitions in OpenCode.
2. Introducing branded or nominal entity identifier types (TicketID, SessionID, MessageID, PartID); deferred to dedicated architectural ticket REF-1.
3. Altering Git commit diff extraction or file churn calculations in backend/internal/adapters/git/.
4. Changing HTTP server endpoint routes or server configuration APIs.
5. Modifying live WebSocket or Server-Sent Events protocols.

---

## 4. Architectural Contracts and Data Specifications

### Tool Failure Classification Contract (backend/internal/domain/live.go)
1. Function Signature:
   ClassifyToolFailure(toolName string, status string, output string, consecutiveErrors int) (bool, string, *string)
2. Content Inspection Tool Definition:
   Tools categorized as content inspection or reading tools comprise read, grep, glob, and webfetch (matched case-insensitively).
3. Invariant Rule:
   If cleanStatus is "completed" and cleanTool is one of the content inspection tools:
   - Return false for isError.
   - Return SeverityNone for severity string.
   - Return nil for errorMessage pointer.
   - Do not inspect output for keywords such as "syntax error", "panic:", "build failed", "permission denied", or "Error:".
4. Non-Inspection Tools or Error Status:
   If status is "error" or "failed", or if the tool is not a content inspection tool (for example, bash, python, edit), existing classification rules for tracebacks, compilation errors, system faults, and command failures continue to apply as defined.

### Per-Session Tool Retry Streak Contract (backend/internal/domain/anomalies.go)
1. Function Signature:
   detectToolRetryStreaks(events []EventFact) []Anomaly
2. Scoping Invariant:
   Consecutive errors are tracked within an individual session ID.
   When encountering a tool call event:
   - If event.Status is "error":
     - If event.ToolName matches activeTool and event.SessionID matches activeSessionID:
       - Increment consecutive error counter.
       - Update lastTurn to event.Turn.
     - Otherwise (tool name differs or session ID differs):
       - If prior streak meets threshold (at least 3 consecutive errors), emit streak anomaly for prior session and tool.
       - Initialize activeTool to event.ToolName, activeSessionID to event.SessionID, and consecutive error counter to 1.
   - If event.Status is "completed" or non-error:
     - If prior streak meets threshold, emit streak anomaly.
     - Reset activeTool, activeSessionID, and streak counter to initial zero values.
   - Upon loop completion, emit any remaining active streak that meets the threshold.

### Event Fact Model Contract (backend/internal/domain/models.go and web/src/types/telemetry.ts)
1. Go EventFact Struct Additions:
   - MessageID *string with json tag "message_id,omitempty"
2. TypeScript EventFact Interface Additions:
   - readonly message_id?: string | null | undefined
3. Provenance Mapping:
   - In opencode.ExtractFacts: assign mid.String to EventFact.MessageID when mid.Valid and non-empty.
   - In live.Service convertToEventFact: assign pr.MessageID to EventFact.MessageID when non-empty.

### Anomaly Matching Contract (web/src/components/TraceWaterfall.tsx)
1. Function Signature:
   findMatchingAnomaly(event: EventFact, anomalies?: readonly Anomaly[]): Anomaly | undefined
2. Predicate Rules:
   An anomaly a matches event if and only if all applicable criteria are satisfied:
   - Session Identity Criterion:
     If a.context.session_id is defined and non-empty, and event.session_id is defined and non-empty:
     The anomaly and event must share identical session IDs (a.context.session_id === event.session_id). If they differ, the candidate does not match.
   - Tool Failure Anomaly Status Criterion:
     If a.rule_id is TOOL_RETRY_STREAK, TOOL_HIGH_FAILURE_RATE, or any tool-targeted failure rule:
     The candidate matches only if event.status === "error". Successful events (status === "completed" or "success") must never receive a tool failure anomaly badge.
   - Tool Name and Turn Match Criterion:
     If a.context.tool_name matches event.tool_name (case-insensitively):
     - If a.context.turn is defined, event.turn must match a.context.turn.
     - If a.context.turn is omitted, session match suffices.

### Collapsible Trace Waterfall Contract (web/src/components/TraceWaterfall.tsx)
1. Default State:
   Collapsed by default upon initial component mount.
2. Collapsed View Rendering:
   Renders a compact summary container displaying:
   - Section Title: "Execution Trace Waterfall"
   - Source Badge: "Source: Local SQLite" or "Source: Committed Ledger"
   - Total Event Count: "X Events Recorded"
   - Issues Summary: "X Errors, Y Anomalies" (or "Clean Execution" when zero issues exist)
   - Expand Button: "Expand Trace Waterfall" with aria-expanded="false" and test id "waterfall-expand-btn"
3. Expanded View Rendering:
   Renders full waterfall header including:
   - Collapse Button: "Collapse Trace Waterfall" with aria-expanded="true" and test id "waterfall-collapse-btn"
   - Filter Button Group: "All Events", "Tool Calls Only", "Errors and Anomalies Only"
   - Multiselect Filter Dropdowns: Role Filter, Tool Filter, Anomaly Filter
   - Chronological Timeline Item Track with expandable payload details

### Concrete Debug Identifiers and Clipboard Action Contract
1. Identifier Formatting:
   - In AnomalyDrawer: Diagnostic context displays Session ID, Message ID (when present), and Turn Debug ID (formatted as "turn-[turnNumber]" or composite "[sessionId]:turn:[turnNumber]").
   - In TraceWaterfall: Timeline item header or details row displays concrete identifier tokens (session ID and message ID).
   - In FailedToolAlertCard: Displays part identifier or session reference.
2. Copy Action Behavior:
   - Button labeled "Copy ID" or "Copy" with aria-label describing the specific identifier copied.
   - Clicking invokes clipboard copy via navigator.clipboard.writeText with fallback to document.execCommand('copy').
   - Button text transitions to "Copied" for 2000 milliseconds before reverting to "Copy".
   - Safe exception handling ensuring clipboard permission denial does not crash the UI.

### Domain Enums and Typing Contract (Iteration 05 Extension)
1. Context Window Health Typings:
   - Go: GroupedMetric.ContextHealth, ContextWindowHealth.Status, ModelMetric.ContextHealth, RoleMetric.ContextHealth, and TicketComparisonRow.ContextHealthStatus must use ContextHealthStatus.
   - TypeScript: GroupedMetric.context_health, ContextWindowHealth.status, ModelMetric.context_health, RoleMetric.context_health, and TicketComparisonRow.context_health_status must use ContextHealthStatus.
2. Agent Role Typings:
   - Go: InvocationFact.Role, EventFact.Role, LiveMessageFact.Role, RoleMetric.Role, ThreadRoleMetric.Role, and ContextWindowHealth.SubagentsPeakRole must use AgentRole. OrderRoles and OrderSubagentRoles must be generic over ~string.
   - TypeScript: CanonicalAgentRole must be 'orchestrator' | 'planner' | 'coder' | 'tester' | 'reviewer'. AgentRole must be autocomplete union CanonicalAgentRole | (string & {}). EventFact.role, InvocationFact.role, LiveMessageFact.role, LiveSessionNode.role, RoleMetric.role, ThreadRoleMetric.role, and ContextWindowHealth.subagents_peak_role must use AgentRole.
3. Failure Severity Typings:
   - Go: LivePartFact.Severity must use FailureSeverity. ClassifyToolFailure signature must return (bool, FailureSeverity, *string). LiveAnomalyAlert.Severity must use LiveAnomalySeverity.
   - TypeScript: FailureSeverity must be 'none' | 'low' | 'medium' | 'high'. LivePartFact.severity must use FailureSeverity.
4. Conventional Ticket Type Typings:
   - Go: AgenticFactLedger.TicketType, AgenticMetricsLedger.TicketType, and TicketComparisonRow.TicketType must use *ConventionalTicketType.
   - TypeScript: StandardTicketType must include canonical types, TicketType must be StandardTicketType | (string & {}), and ticket_type across ledgers and rows must use TicketType | null.
