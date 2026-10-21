# Architectural Contract: AL-19 Live Anomaly Detection Engine and Realtime Session Health Rules

## 1. Problem Statement
During autonomous multi-agent development sessions, LLM-based agents can enter degenerative failure modes, including:
1. Repeated and escalating tool execution failure streaks where an agent repeatedly runs into errors without self-correction.
2. Context window saturation approaching or breaching the healthy context ceiling (HealthyCeilingTokens = 200,000 tokens), causing degraded reasoning or abrupt truncation.
3. Abnormally high token consumption per turn (token velocity runaway), indicating unbounded generation, giant file dumps, or infinite loops.
4. Tool thrashing where the agent fixates on repeatedly invoking the exact same failing tool without making forward progress.

Currently, live observability endpoints exposed by AL-17 provide session trees and trace views, but lack automated real-time health evaluations. Operators and downstream observability dashboards (AL-20) require sub-millisecond, in-memory live anomaly detection to identify and flag degraded sessions as they occur, before token budgets are wasted or sessions fail silently.

## 2. Functional Core, Imperative Shell (FCIS) Invariants
The live anomaly detection system strictly adheres to the FCIS architecture:

### Functional Core (backend/internal/domain/live_anomaly.go and backend/internal/domain/live.go)
- Pure functional, deterministic rule evaluation functions operating on immutable structs.
- Zero external imports: absolutely no net/http, database/sql, os, or filesystem dependencies.
- Zero nondeterminism: no time.Now() calls inside rule evaluation logic. All timestamps, durations, and thresholds are passed as explicit parameters or read directly from message and trace models.
- Mathematical safety: defensive guards against empty message slices, zero tool invocations, zero context tokens, negative token deltas, and division by zero.
- Immutability: evaluators accept read-only pointers or value copies of LiveSessionTree and LiveTraceView, returning freshly allocated anomaly slices without mutating input structures.

### Imperative Shell (backend/internal/adapters/live/ and backend/internal/adapters/http/)
- backend/internal/adapters/live/sqlite.go queries raw session records, message payloads, and part data through read-only pragmas.
- backend/internal/adapters/live/service.go orchestrates transient in-memory buffering, extracts token metrics from raw message JSON payloads, executes the domain's pure LiveAnomaly evaluators, and broadcasts alert events.
- backend/internal/adapters/live/cache.go provides the thread-safe, bounded in-memory ring buffer (TraceBuffer) for active sessions.
- backend/internal/adapters/http/live_handlers.go and backend/internal/adapters/http/dto.go serialize enriched tree and trace models to JSON and stream live events across Server-Sent Events (SSE) connections.
- Clocks (time.Now().UnixMilli()) and system interactions are strictly confined to this outer shell layer.

## 3. Scope Boundaries and Rule Specifications

### Rule 1: ConsecutiveToolErrorsRule
- Purpose: Detect escalating execution failure streaks where tool calls fail repeatedly without intervening recovery.
- Evaluator: Evaluates the chronological sequence of tool parts in LiveTraceView.
- Trigger Condition: When the count of consecutive tool execution failures (where is_error is true) reaches or exceeds 3.
- Recovery Reset: Any successful tool execution (where is_error is false) immediately resets the consecutive error counter to 0.
- Severity Tiers:
  - High: Error streak count of 3 or 4 consecutive tool failures.
  - Critical: Error streak count of 5 or more consecutive tool failures.
- Output Alert Metadata:
  - RuleName: ConsecutiveToolErrors
  - Severity: HIGH or CRITICAL
  - Message: Tool execution error streak reached N consecutive failures without recovery (last failing tool: ToolName).
  - TriggerEvent: Part ID of the latest failing tool invocation and sanitized error snippet.

### Rule 2: ContextCeilingApproachRule
- Purpose: Warn operators when an active session's context token accumulation approaches or breaches the defined healthy context window ceiling.
- Evaluator: Evaluates the highest context token count observed across messages in LiveTraceView against HealthyCeilingTokens (200,000 tokens).
- Trigger Condition and Severity Tiers:
  - Warning: Context tokens reach or exceed 80% of ceiling (>= 160,000 tokens and < 200,000 tokens).
  - Critical: Context tokens reach or exceed 100% of ceiling (>= 200,000 tokens).
- Output Alert Metadata:
  - RuleName: ContextCeilingApproach
  - Severity: WARNING or CRITICAL
  - Message: Context window utilization reached N tokens (P% of healthy ceiling 200k tokens).
  - TriggerEvent: Turn timestamp or message ID where the ceiling was first approached or breached.

### Rule 3: TokenVelocityRunawayRule
- Purpose: Detect sudden token consumption surges in a single turn indicating runaway agent loops, huge context injections, or verbose LLM regurgitation.
- Evaluator: Evaluates token consumption per message turn across assistant messages in LiveTraceView, inspecting total tokens per turn or turn-to-turn token delta.
- Trigger Condition and Severity Tiers:
  - Warning: A single turn consumes greater than 30,000 tokens or causes a context surge delta greater than 40,000 tokens.
  - Critical: A single turn consumes greater than 60,000 tokens or causes a context surge delta greater than 80,000 tokens.
- Output Alert Metadata:
  - RuleName: TokenVelocityRunaway
  - Severity: WARNING or CRITICAL
  - Message: Abnormally high token velocity of N tokens consumed in turn T (threshold: Thresh tokens).
  - TriggerEvent: Message ID and turn number of the surging turn.

### Rule 4: ToolThrashingRule
- Purpose: Detect agent fixation loops where the agent repeatedly calls the exact same tool with failing outcomes without making progress.
- Evaluator: Evaluates the sequence of tool calls in LiveTraceView, tracking repeat invocations of the identical tool name where the executions fail without any intermediate success.
- Trigger Condition:
  - The same tool name is invoked and fails 3 or more times consecutively, or fails 3 or more times within the most recent window of 5 tool calls without intervening progress.
- Severity Tiers:
  - Warning: 3 failing calls to the same tool without progress.
  - Critical: 5 or more failing calls to the same tool without progress.
- Output Alert Metadata:
  - RuleName: ToolThrashing
  - Severity: WARNING or CRITICAL
  - Message: Tool thrashing detected on tool ToolName with N failing invocations without progress.
  - TriggerEvent: Tool name and part IDs of the failing invocations.

## 4. Domain Model Definitions and Evaluator Signatures

### LiveAnomalyAlert Struct Specification (backend/internal/domain/live_anomaly.go)
- SessionID (string, JSON field session_id): Target session identifier.
- RuleName (string, JSON field rule_name): Canonical rule name string (ConsecutiveToolErrors, ContextCeilingApproach, TokenVelocityRunaway, ToolThrashing).
- Severity (string, JSON field severity): Severity tier string (WARNING, HIGH, CRITICAL).
- Message (string, JSON field message): Diagnostic description explaining the health issue and threshold breach.
- Timestamp (string, JSON field timestamp): ISO-8601 formatted timestamp of the triggering event.
- TriggerEvent (string, JSON field trigger_event): Identifying context (part ID, message ID, or tool signature).

### LiveMessageFact Extension (backend/internal/domain/live.go)
- Extend LiveMessageFact to record token metrics extracted from raw message JSON payloads:
  - ContextTokens (int64, JSON field context_tokens): Sum of input tokens and cache read tokens.
  - InputTokens (int64, JSON field input_tokens): Raw input tokens for the message turn.
  - OutputTokens (int64, JSON field output_tokens): Raw output tokens for the message turn.
  - CacheReadTokens (int64, JSON field cache_read_tokens): Cached input tokens read.
  - ReasoningTokens (int64, JSON field reasoning_tokens): Reasoning/thought tokens generated.
  - TotalTokens (int64, JSON field total_tokens): Total tokens consumed in the turn.

### LiveTraceView Extension (backend/internal/domain/live.go)
- Extend LiveTraceView to include detected live anomalies:
  - Anomalies ([]LiveAnomalyAlert, JSON field anomalies): Slices of active anomaly alerts detected for this trace view.

### LiveSessionNode and LiveSessionTree Extension (backend/internal/domain/live.go)
- Extend LiveSessionNode to include:
  - Anomalies ([]LiveAnomalyAlert, JSON field anomalies): Anomalies identified for the specific session node.
- Extend LiveSessionTree to include:
  - ActiveAnomalies ([]LiveAnomalyAlert, JSON field active_anomalies): Rollup of all active anomaly alerts across active sessions.

### Pure Domain Evaluator Signatures
- EvaluateConsecutiveToolErrors(trace *LiveTraceView) []LiveAnomalyAlert
- EvaluateContextCeilingApproach(trace *LiveTraceView, ceilingTokens int64) []LiveAnomalyAlert
- EvaluateTokenVelocityRunaway(trace *LiveTraceView, warningThreshold int64, criticalThreshold int64) []LiveAnomalyAlert
- EvaluateToolThrashing(trace *LiveTraceView) []LiveAnomalyAlert
- EvaluateLiveAnomalies(tree *LiveSessionTree, trace *LiveTraceView) []LiveAnomalyAlert
  - Master composite evaluator. Evaluates all active health rules. Safely handles nil tree or nil trace arguments. Returns a deterministically ordered slice sorted primarily by Severity descending (CRITICAL, HIGH, WARNING), secondarily by RuleName ascending, and tertiary by Timestamp descending.

## 5. API Response Schemas and SSE Event Contracts

### 1. Active Sessions Tree Endpoint: GET /api/live/sessions
Response HTTP 200 Status: application/json
Top-level JSON Object:
- root_sessions: Array of LiveSessionNode objects. Each node contains id, parent_id, agent, role, model, title, time_created, time_updated, status, guessed_ticket_id, ticket_confidence, anomalies (Array of LiveAnomalyAlert), and sub_sessions (Array of LiveSessionNode).
- last_refreshed_at: String (ISO-8601 formatted timestamp).
- total_active_sessions: Integer count of active sessions across the hierarchy.
- active_anomalies: Array of LiveAnomalyAlert objects representing all active session anomalies across the tree.

### 2. Session Trace View Endpoint: GET /api/live/sessions/:id/trace
Response HTTP 200 Status: application/json
Top-level JSON Object:
- session_id: String session identifier.
- messages: Array of LiveMessageFact objects. Each message contains id, role, content, timestamp, reasoning_content, context_tokens, input_tokens, output_tokens, cache_read_tokens, reasoning_tokens, total_tokens, and parts.
- total_tool_calls: Integer count of tool executions.
- failed_tool_calls: Integer count of failing tool executions.
- recent_errors: Array of LivePartFact objects representing recent tool errors ordered newest first.
- anomalies: Array of LiveAnomalyAlert objects detected for this session trace.

### 3. Real-Time Event Stream Endpoint: GET /api/live/stream (SSE)
Response HTTP 200 Status: text/event-stream; charset=utf-8
Headers: Cache-Control: no-cache, Connection: keep-alive, X-Accel-Buffering: no
Event Types:
- event: session_tree
  - data: Serialized LiveSessionTree JSON object (pushed on connection and whenever tree topology, status, or session anomalies change).
- event: trace_update
  - data: Serialized LiveTraceView JSON object (pushed when active session trace messages, tool calls, or anomalies update).
- event: anomaly_alert
  - data: Serialized LiveAnomalyAlert JSON object (pushed immediately when a new live anomaly alert is triggered).

## 6. Zero Code Mandate Verification
This contract contains zero code fence blocks and strictly defines domain specifications, types, thresholds, and boundary contracts through descriptive prose and structured specifications.
