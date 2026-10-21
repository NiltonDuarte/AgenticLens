# Implementation Plan: AL-19 Live Anomaly Detection Engine and Realtime Session Health Rules

## 1. Goal and Objectives
Implement a pure functional real-time anomaly detection engine in Go internal/domain to evaluate transient live session execution traces against operational health rules (consecutive tool errors, context ceiling approach, token runaway velocity, tool thrashing), and expose live anomaly indicators through the live HTTP API endpoints and SSE stream.

## 2. Phase Breakdown and Task Decomposition

### Phase 1: Pure Domain Functional Core
Target Files:
- backend/internal/domain/live.go
- backend/internal/domain/live_anomaly.go (new file)
- backend/internal/domain/constants.go

Tasks:
1. Extend Domain Models in backend/internal/domain/live.go:
   - Add token count fields to LiveMessageFact:
     - ContextTokens int64 (JSON context_tokens)
     - InputTokens int64 (JSON input_tokens)
     - OutputTokens int64 (JSON output_tokens)
     - CacheReadTokens int64 (JSON cache_read_tokens)
     - ReasoningTokens int64 (JSON reasoning_tokens)
     - TotalTokens int64 (JSON total_tokens)
   - Add Anomalies slice to LiveSessionNode:
     - Anomalies []LiveAnomalyAlert (JSON anomalies)
   - Add ActiveAnomalies slice to LiveSessionTree:
     - ActiveAnomalies []LiveAnomalyAlert (JSON active_anomalies)
   - Add Anomalies slice to LiveTraceView:
     - Anomalies []LiveAnomalyAlert (JSON anomalies)
2. Define Live Anomaly Types in backend/internal/domain/live_anomaly.go:
   - Define LiveAnomalyAlert struct with fields SessionID, RuleName, Severity, Message, Timestamp, TriggerEvent.
   - Define canonical rule name constants:
     - RuleConsecutiveToolErrors = "ConsecutiveToolErrors"
     - RuleContextCeilingApproach = "ContextCeilingApproach"
     - RuleTokenVelocityRunaway = "TokenVelocityRunaway"
     - RuleToolThrashing = "ToolThrashing"
   - Define severity tier constants or reuse existing Severity tiers (SeverityWarning, SeverityHigh, SeverityCritical).
   - Define default threshold constants:
     - ConsecutiveToolErrorsStreakWarning = 3
     - ConsecutiveToolErrorsStreakCritical = 5
     - ContextCeilingApproachRatioWarning = 0.80 (160,000 tokens for 200k ceiling)
     - ContextCeilingApproachRatioCritical = 1.00 (200,000 tokens for 200k ceiling)
     - TokenVelocityWarningTokens = 30000
     - TokenVelocityCriticalTokens = 60000
     - ToolThrashingMinFailures = 3
3. Implement Pure Rule Evaluators in backend/internal/domain/live_anomaly.go:
   - Implement EvaluateConsecutiveToolErrors(trace *LiveTraceView) []LiveAnomalyAlert:
     - Guard against nil trace or empty messages.
     - Traverse all message parts in chronological order filtering for Type == "tool".
     - Increment error streak on IsError == true, reset streak on IsError == false.
     - When streak >= 3, generate LiveAnomalyAlert (HIGH for 3-4, CRITICAL for >= 5).
   - Implement EvaluateContextCeilingApproach(trace *LiveTraceView, ceilingTokens int64) []LiveAnomalyAlert:
     - Fall back to HealthyCeilingTokens (200,000) when ceilingTokens <= 0.
     - Find maximum ContextTokens across all messages in trace.
     - If peak context >= ceilingTokens, emit CRITICAL alert.
     - Else if peak context >= 80% of ceilingTokens (160,000), emit WARNING alert.
   - Implement EvaluateTokenVelocityRunaway(trace *LiveTraceView, warningThreshold int64, criticalThreshold int64) []LiveAnomalyAlert:
     - Apply default thresholds (30,000 warning, 60,000 critical) if non-positive.
     - Traverse assistant messages in trace; check TotalTokens or ContextTokens delta between consecutive turns.
     - Emit WARNING or CRITICAL alert on threshold exceedance with turn number and message ID.
   - Implement EvaluateToolThrashing(trace *LiveTraceView) []LiveAnomalyAlert:
     - Collect sequence of tool call records (tool name, is_error, part ID, timestamp).
     - Detect if the same tool name fails >= 3 times without an intervening successful turn or within the last 5 tool invocations.
     - Emit WARNING (3 failures) or CRITICAL (>= 5 failures) alert identifying the thrashing tool.
   - Implement EvaluateLiveAnomalies(tree *LiveSessionTree, trace *LiveTraceView) []LiveAnomalyAlert:
     - Safely handle nil inputs for tree or trace.
     - If trace != nil: run all 4 trace-level evaluators, concatenate results.
     - If tree != nil: inspect tree nodes for tree-level anomaly status if applicable.
     - Sort results deterministically: primary sort Severity descending (CRITICAL > HIGH > WARNING), secondary sort RuleName ascending, tertiary sort Timestamp descending.
     - Return empty slice rather than nil when no anomalies are detected.

### Phase 2: Boundary Adapters and Live Engine Integration
Target Files:
- backend/internal/adapters/live/service.go
- backend/internal/adapters/live/cache.go
- backend/internal/adapters/live/sqlite.go

Tasks:
1. Message Payload Token Extraction in backend/internal/adapters/live/service.go:
   - In GetSessionTrace, update the unmarshaling of rm.Data from SQLite message table to parse the tokens object:
     - Input, Output, Cache.Read, Reasoning.
   - Calculate ContextTokens as Input + Cache.Read.
   - Calculate TotalTokens as Input + Output + Cache.Read.
   - Assign token values to the constructed LiveMessageFact instances.
2. In-Memory Trace Anomaly Evaluation in backend/internal/adapters/live/service.go:
   - In GetSessionTrace, after assembling LiveTraceView messages and error metrics, call domain.EvaluateLiveAnomalies(nil, &trace).
   - Assign the returned alerts to trace.Anomalies.
   - Cache recent trace messages and evaluation state in TraceBuffer.
3. Tree-Level Anomaly Rollup in backend/internal/adapters/live/service.go:
   - In GetSessionTree, for each active session record, evaluate session anomalies using trace messages from the SQLite trace or TraceBuffer.
   - Populate node.Anomalies on each LiveSessionNode.
   - Aggregate all active anomalies into tree.ActiveAnomalies on LiveSessionTree.
4. Real-Time Alert Streaming in backend/internal/adapters/live/service.go:
   - Add LiveEventAnomalyAlert = "anomaly_alert" to LiveEventType enum in service.go.
   - In pollAndBroadcast, detect when new anomalies emerge between polling cycles.
   - When a new anomaly is detected, broadcast LiveEvent with Type LiveEventAnomalyAlert containing the JSON-serialized LiveAnomalyAlert.
   - Broadcast updated LiveEventSessionTree containing the updated node and tree anomalies.

### Phase 3: HTTP API Handlers and DTO Verification
Target Files:
- backend/internal/adapters/http/live_handlers.go
- backend/internal/adapters/http/dto.go
- backend/internal/adapters/http/server.go

Tasks:
1. Update Response Constraints in backend/internal/adapters/http/dto.go:
   - Verify ResponsePayload type constraint accommodates domain.LiveSessionTree, domain.LiveTraceView, domain.LiveAnomalyAlert, and []domain.LiveAnomalyAlert.
2. Verify Handlers in backend/internal/adapters/http/live_handlers.go:
   - LiveSessionsHandler: ensure JSON serialization includes anomalies on each session node and active_anomalies on the root tree.
   - LiveTraceHandler: ensure JSON serialization includes the anomalies array on the trace view response.
   - LiveStreamHandler: ensure the SSE stream delivers anomaly_alert events alongside session_tree and trace_update events.

### Phase 4: Verification and Table-Driven Test Suite
Target Files:
- backend/internal/domain/live_anomaly_test.go (new file)
- backend/internal/domain/live_test.go
- backend/internal/adapters/live/sqlite_test.go
- backend/internal/adapters/http/live_handlers_test.go

Tasks:
1. Pure Domain Unit Tests (backend/internal/domain/live_anomaly_test.go):
   - Zero mocks policy: Pure in-memory struct testing exclusively.
   - Test ConsecutiveToolErrorsRule:
     - 0 tool errors: no alert.
     - 1 and 2 tool errors: no alert.
     - 3 consecutive tool errors: triggers HIGH alert.
     - 4 consecutive tool errors: triggers HIGH alert.
     - 5 consecutive tool errors: triggers CRITICAL alert.
     - 2 tool errors followed by 1 success followed by 2 errors: streak resets, zero alerts.
     - Tool errors across multiple different tools in a row: streak triggers correctly.
   - Test ContextCeilingApproachRule:
     - 0 tokens: zero alerts, no division by zero.
     - 100k tokens (50% of ceiling): zero alerts.
     - 160k tokens (80% of ceiling): triggers WARNING alert.
     - 199k tokens (99.5% of ceiling): triggers WARNING alert.
     - 200k tokens (100% of ceiling): triggers CRITICAL alert.
     - 250k tokens (125% of ceiling): triggers CRITICAL alert.
     - Custom ceiling argument passed: honors custom ceiling rather than default.
   - Test TokenVelocityRunawayRule:
     - Normal turns with 2,000 tokens: zero alerts.
     - Sudden surge turn with 35,000 tokens: triggers WARNING alert.
     - Extreme surge turn with 65,000 tokens: triggers CRITICAL alert.
     - Turn with 0 tokens: no panic, no alert.
   - Test ToolThrashingRule:
     - Same tool called 3 times successfully: zero alerts.
     - Different tools failing once each: zero thrashing alerts.
     - Same tool failing 3 times: triggers WARNING thrashing alert.
     - Same tool failing 5 times: triggers CRITICAL thrashing alert.
     - Tool fails 2 times, succeeds once, fails once: no thrashing alert.
   - Test Composite EvaluateLiveAnomalies:
     - Nil tree, nil trace: returns empty slice, zero panics.
     - Populated tree, nil trace: handles cleanly.
     - Nil tree, populated trace: executes all rules cleanly.
     - Multiple rules triggered simultaneously: verifies deterministic sorting (CRITICAL first, then HIGH, then WARNING).
2. Adapter and Live Engine Tests (backend/internal/adapters/live/sqlite_test.go):
   - Seed SQLite database with test messages containing raw token metrics and parts with errors.
   - Call GetSessionTrace and assert that tokens are properly parsed from raw JSON.
   - Verify that GetSessionTrace returns populated anomalies.
   - Verify that GetSessionTree returns nodes with populated anomalies and active_anomalies.
3. HTTP and SSE Handler Tests (backend/internal/adapters/http/live_handlers_test.go):
   - Test LiveSessionsHandler outputs anomalies field on nodes and active_anomalies on root tree.
   - Test LiveTraceHandler outputs anomalies field on trace response.
   - Test LiveStreamHandler broadcasts event: anomaly_alert when live anomalies are published.

## 3. Strict Quality Invariants
1. Functional Core, Imperative Shell (FCIS):
   - Zero I/O, zero network, zero database, zero time.Now() in backend/internal/domain/live_anomaly.go.
2. Typing and Zero Any:
   - Zero any and zero interface{} in all new Go code.
3. Zero Mocks in Domain:
   - All domain tests construct concrete structs and assert outputs deterministically without mock libraries.
4. Zero Emojis:
   - Zero emojis permitted in source code, comments, docstrings, commit messages, or test outputs.

## 4. Verification Commands
Run the following commands to verify implementation quality:
- Static analysis: go vet ./...
- Domain unit test suite: go test -v -race ./backend/internal/domain/...
- Live adapters test suite: go test -v -race ./backend/internal/adapters/live/...
- HTTP handlers test suite: go test -v -race ./backend/internal/adapters/http/...
- Full project test suite: go test -v -race ./...
- Production binary build: go build -o bin/agentic-lens ./backend/cmd/agentic-lens
- Complete quality check: make
