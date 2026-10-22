# Implementation Plan: AL-19 Iteration 02 - Post-Filesystem Recovery Verification, Contract Reconciliation, and Type Parity Hardening

## 1. Goal and Iteration Scope

Iteration 02 performs a rigorous post-filesystem recovery verification of all deliverables implemented for ticket AL-19 (Live Anomaly Detection Engine and Real-Time Session Health Rules). Following a filesystem and environment disruption, this plan audits the Go backend functional core, imperative adapters, HTTP presentation layer, and TypeScript frontend models against workspaces/AL-19/CONTRACT.md, prior iteration artifacts (workspaces/AL-19/iterations/01/PLAN.md and REVIEW.md), and repository technical guidelines.

Four primary objectives define this iteration:
1. Reconcile and verify the integrity of all four pure live anomaly rules, the composite evaluator, deterministic multi-key sorting, and table-driven unit tests in the Go domain.
2. Confirm the complete operational status of the live SQLite ingestion pipeline, token metrics calculation, thread-safe Server-Sent Events anomaly broadcasting, and HTTP DTO serialization.
3. Close concrete model gaps identified in frontend TypeScript interfaces by adding missing fields (parent_id on LiveSessionNode, output_snippet on LivePartFact, and reasoning_content on LiveMessageFact) and authoring a dedicated unit test suite for web domain live types.
4. Eliminate residual interface{} occurrences in the HTTP test suite to satisfy the strict typing mandate, verify npm and node_modules filesystem stability, and execute the full validation matrix across make check, make test, and make build.

## 2. Grounded Reference Architecture and Symbols

The verification and hardening steps target exact files, symbols, and architectural boundaries across the codebase:

- Pure Functional Domain Core:
  - File: backend/internal/domain/live.go
    - Symbols: LivePartFact, LiveMessageFact, LiveSessionNode, LiveSessionTree, LiveTraceView, ClassifyToolFailure, BuildSessionTree.
  - File: backend/internal/domain/live_anomaly.go
    - Symbols: LiveAnomalyAlert struct, RuleConsecutiveToolErrors constant, RuleContextCeilingApproach constant, RuleTokenVelocityRunaway constant, RuleToolThrashing constant, LiveSeverityWarning, LiveSeverityHigh, LiveSeverityCritical, ConsecutiveToolErrorsStreakWarning, ConsecutiveToolErrorsStreakCritical, ContextCeilingApproachRatioWarning, ContextCeilingApproachRatioCritical, TokenVelocityWarningTokens, TokenVelocityCriticalTokens, TokenVelocityWarningDeltaTokens, TokenVelocityCriticalDeltaTokens, ToolThrashingMinFailures, ToolThrashingCriticalFailures, EvaluateConsecutiveToolErrors function, EvaluateContextCeilingApproach function, EvaluateTokenVelocityRunaway function, EvaluateToolThrashing function, SortLiveAnomalies function, EvaluateLiveAnomalies composite function.
  - File: backend/internal/domain/live_anomaly_test.go
    - Symbols: TestConsecutiveToolErrorsRule, TestContextCeilingApproachRule, TestTokenVelocityRunawayRule, TestToolThrashingRule, TestEvaluateLiveAnomalies_Composite, TestSortLiveAnomalies_DeterministicOrdering.

- Imperative Shell Adapters:
  - File: backend/internal/adapters/live/service.go
    - Symbols: LiveService struct, LiveEventType enum, LiveEventAnomalyAlert constant, LiveEventSessionTree constant, LiveEventTraceUpdate constant, LiveEvent struct, NewLiveService constructor, LiveService.Start, LiveService.pollAndBroadcast, LiveService.checkAndBroadcastAnomalies, LiveService.GetSessionTree, LiveService.GetSessionTrace, parsePartRecord function, validateSessionID function, extractMessageContent function, resolveModelName function.
  - File: backend/internal/adapters/live/sqlite.go
    - Symbols: SQLiteHook struct, OpenReadOnlyHook function, FetchActiveSessions method, FetchSessionTrace method, SessionExists method.
  - File: backend/internal/adapters/live/cache.go
    - Symbols: TraceBuffer struct, NewTraceBuffer constructor, TraceBuffer.SetMessages, TraceBuffer.GetMessages.
  - File: backend/internal/adapters/live/sqlite_test.go
    - Symbols: TestReadOnlyPragmaEnforcement, TestFetchSessionTrace_TokenExtraction, TestFetchSessionTrace_LiveAnomaliesPopulated, TestGetSessionTree_WithAnomalies, TestLiveService_AnomalyAlertBroadcast.

- HTTP Handlers and DTO Serialization:
  - File: backend/internal/adapters/http/live_handlers.go
    - Symbols: LiveServiceInterface interface, LiveSessionsHandler function, LiveTraceHandler function, LiveStreamHandler function.
  - File: backend/internal/adapters/http/dto.go
    - Symbols: ResponsePayload type constraint, WriteJSON generic function, WriteLiveSessionTreeResponse function, WriteLiveTraceResponse function, WriteLiveAnomalyAlertResponse function, WriteLiveAnomalyAlertsResponse function.
  - File: backend/internal/adapters/http/live_handlers_test.go
    - Symbols: mockLiveService struct, TestLiveSessionsHandler_Success, TestLiveSessionsHandler_AnomaliesField, TestLiveTraceHandler_AnomaliesField, TestLiveStreamHandler_AnomalyAlertBroadcast, TestLiveHandlers_NilSlicesSerializeAsEmptyArrays.

- Web TypeScript Frontend Models and Tests:
  - File: web/src/domain/live.ts
    - Symbols: LiveAnomalySeverity type, LiveAnomalyRule type, LiveAnomalyAlert interface, LivePartFact interface, LiveMessageFact interface, LiveTraceView interface, LiveSessionNode interface, LiveSessionTree interface.
  - File: web/src/domain/types.ts
    - Symbols: Re-exported live types from web/src/domain/live.
  - File: web/src/domain/live.test.ts (Target New Verification Test Suite)
    - Symbols: Test suites covering LiveAnomalyAlert validation, LiveSessionTree hierarchy serialization, LiveTraceView token aggregation, and optional field resilience.

- Build and Environment Configuration:
  - Files: Makefile, web/package.json, web/package-lock.json, .agenticlens/config.yaml.

## 3. Grounded Codebase Verification Findings

Grounding against the repository confirmed that core backend files survived the filesystem disruption intact, but highlighted four specific items requiring action:

1. Domain Rules and Test Suite Status (backend/internal/domain/):
   - All four rules in live_anomaly.go are fully implemented according to CONTRACT.md specifications.
   - Rule 1 (ConsecutiveToolErrors): Correctly tracks tool part errors, resets streak to zero on success, ignores non-tool parts, triggers HIGH at streak 3-4, and triggers CRITICAL at streak 5 or greater.
   - Rule 2 (ContextCeilingApproach): Correctly compares peak context against the healthy ceiling (defaulting to 200,000 tokens), triggering WARNING at 80% (160k tokens) and CRITICAL at 100% (200k tokens).
   - Rule 3 (TokenVelocityRunaway): Correctly monitors assistant turn token consumption, triggering WARNING at 30,000 total tokens or 40,000 delta, and CRITICAL at 60,000 total tokens or 80,000 delta, safely handling compaction drops.
   - Rule 4 (ToolThrashing): Correctly monitors repeated failing calls to identical tools, evaluating both trace-ending streaks and recent five-call windows, triggering WARNING at 3 failures and CRITICAL at 5 failures.
   - Composite Evaluator and Sorting: EvaluateLiveAnomalies executes all rules, handles nil trees and traces defensively, deduplicates identical alert tuples, and applies deterministic multi-key sorting (Severity descending, RuleName ascending, Timestamp descending, SessionID ascending, TriggerEvent ascending).
   - Domain unit test file live_anomaly_test.go contains 1,269 lines of comprehensive table-driven tests with zero mocks and zero test skips.

2. Adapters and Live Engine Status (backend/internal/adapters/live/):
   - service.go correctly unmarshals raw SQLite message JSON to extract input, output, cache read, and reasoning tokens, computing ContextTokens (input + cache read) and TotalTokens (input + output + cache read).
   - GetSessionTrace attaches evaluated anomalies to LiveTraceView.
   - GetSessionTree populates node anomalies and aggregates active anomalies into LiveSessionTree.ActiveAnomalies.
   - checkAndBroadcastAnomalies streams anomaly alerts over SSE with thread-safe deduplication in seenAnomalies map.
   - sqlite_test.go contains integration tests verifying token extraction, anomaly attachment, and SSE broadcasts with zero test skips.

3. HTTP Handlers and DTO Serialization Status (backend/internal/adapters/http/):
   - live_handlers.go validates path traversal characters, handles nil service instances with empty fallback arrays, and streams SSE events with proper flusher management.
   - dto.go incorporates live models into ResponsePayload and ensures nil slices serialize as empty JSON arrays instead of null.
   - Defect Identified in Test Suite: lines 496, 501, 509, 510, 511, 682, and 701 of live_handlers_test.go utilize map[string]interface{} and []interface{} type assertions. These must be replaced with concrete structs or map[string]json.RawMessage to fulfill the repository zero-interface invariant.

4. Web TypeScript Types Status (web/src/domain/):
   - live.ts correctly defines LiveAnomalySeverity, LiveAnomalyRule, and LiveAnomalyAlert.
   - Model Gaps Identified:
     - LiveSessionNode in web/src/domain/live.ts omits parent_id?: string | null (present in Go domain as ParentID *string).
     - LivePartFact in web/src/domain/live.ts omits output_snippet?: string (present in Go domain as OutputSnippet string).
     - LiveMessageFact in web/src/domain/live.ts omits reasoning_content?: string | null (present in Go domain as ReasoningContent *string).
   - Missing Frontend Test Suite: Unlike events, migrations, formatters, and progression, web/src/domain/ currently lacks a live.test.ts verifying live payload parsing and optional property boundaries.

5. Filesystem Permissions and Build Pipeline Status:
   - File permissions across web/ and web/node_modules/ must be verified for uncorrupted write and execution access.
   - Go module verification (go mod verify) is required to ensure backend vendor checksum consistency.
   - Full build pipeline targets (make check, make test, and make build) must run cleanly end-to-end.

## 4. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Environment and Filesystem Stability Verification
- Task 1.1: Verify filesystem permissions in the web directory. Ensure user write permissions on web/node_modules, web/dist, and backend/internal/adapters/web/dist. Re-run dependency installation if any broken symlinks or missing packages exist.
- Task 1.2: Verify Go dependency integrity in backend by executing go mod verify and go mod tidy to ensure go.mod and go.sum are synchronized and uncorrupted.
- Task 1.3: Run git diff --check to verify the repository working tree is free of whitespace errors, merge conflict leftovers, or broken index references.

### Phase 2: Web TypeScript Domain Interface Reconciliation
- Task 2.1: Update LiveSessionNode in web/src/domain/live.ts:
  - Add optional readonly property parent_id?: string | null.
  - Maintain readonly modifiers across all properties and strict typing with zero any.
- Task 2.2: Update LivePartFact in web/src/domain/live.ts:
  - Add optional readonly property output_snippet?: string.
  - Ensure input_snippet and output_snippet match the payload structure serialized by the Go backend.
- Task 2.3: Update LiveMessageFact in web/src/domain/live.ts:
  - Add optional readonly property reasoning_content?: string | null.
  - Ensure all token fields (context_tokens, input_tokens, output_tokens, cache_read_tokens, reasoning_tokens, total_tokens) remain strictly typed as number.
- Task 2.4: Verify re-exports in web/src/domain/types.ts:
  - Ensure all updated interfaces are cleanly re-exported without breaking downstream views and components.

### Phase 3: HTTP Handler Test Suite Zero-Interface Hardening
- Task 3.1: Refactor TestLiveSessionsHandler_AnomaliesField in backend/internal/adapters/http/live_handlers_test.go:
  - Replace untyped map[string]interface{} deserialization with a dedicated local struct or map[string]json.RawMessage.
  - Assert that active_anomalies and root_sessions arrays deserialize as valid non-null JSON arrays containing the expected alert and node fields.
- Task 3.2: Refactor TestLiveHandlers_NilSlicesSerializeAsEmptyArrays in backend/internal/adapters/http/live_handlers_test.go:
  - Replace rawTree and rawTrace map[string]interface{} assertions with map[string]json.RawMessage.
  - Inspect raw JSON byte values to assert that uninitialized slices serialize as string representations of empty arrays rather than null.
- Task 3.3: Verify zero occurrences of interface{} remain across the entire backend/internal/adapters/http/ package.

### Phase 4: Frontend Live Domain Unit Test Suite
- Task 4.1: Author web/src/domain/live.test.ts using Vitest:
  - Include Node test environment directive at the top of the file.
  - Test LiveAnomalyAlert deserialization across all severity levels (WARNING, HIGH, CRITICAL) and canonical rule names.
  - Test LiveSessionNode deserialization with populated parent_id and null or undefined parent_id.
  - Test LiveSessionTree hierarchy deserialization with multiple root sessions and nested sub-sessions.
  - Test LiveMessageFact deserialization with token breakdown fields, presence of reasoning_content, and absence of reasoning_content.
  - Test LivePartFact deserialization with tool name, input snippet, output snippet, error status, and error message.
  - Test deserialization of empty anomaly slices and verify that missing optional fields default safely without runtime exceptions.

### Phase 5: Comprehensive Test Suite and Build Verification
- Task 5.1: Execute backend static analysis:
  - Run cd backend && go vet ./... and verify zero warnings.
- Task 5.2: Execute backend unit and integration test suites:
  - Run cd backend && go test -v -race ./... and verify all packages pass cleanly with zero data races and zero skipped tests.
- Task 5.3: Execute frontend static checks and unit tests:
  - Run cd web && npm run typecheck.
  - Run cd web && npm run lint.
  - Run cd web && npm test.
- Task 5.4: Execute repository root quality check:
  - Run make check to verify vet, lint, and git diff check.
  - Run make test to verify full test suite execution across backend and frontend.
  - Run make build to compile multi-platform distribution binaries (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64) and verify binary wrapper placement in bin/agentic-lens.

## 5. Adversarial Testing Strategy for Tester

### Pure Domain Rule Boundary Tests (backend/internal/domain/live_anomaly_test.go)
- Test Case 1: Consecutive Tool Errors Recovery Invariant:
  - Construct a trace with two failing tool parts, followed by a non-tool reasoning part, followed by a third failing tool part.
  - Assert that non-tool parts do not reset the streak and an alert is triggered.
  - Append a successful tool part followed by two failing tool parts; assert that the streak reset prevents any alert from firing.
- Test Case 2: Exact Ceiling Threshold Boundaries:
  - Test ceiling 200,000 tokens with peak context 159,999 tokens (assert zero alerts).
  - Test peak context 160,000 tokens (assert WARNING alert with 80% utilization message).
  - Test peak context 199,999 tokens (assert WARNING alert).
  - Test peak context 200,000 tokens (assert CRITICAL alert with 100% utilization message).
  - Test peak context 250,000 tokens (assert CRITICAL alert with 125% utilization message).
- Test Case 3: Token Velocity Surges and Compaction Clamping:
  - Test assistant message with total tokens exceeding 30,000 (WARNING) and 60,000 (CRITICAL).
  - Test turn-to-turn context token delta exceeding 40,000 (WARNING) and 80,000 (CRITICAL).
  - Test context reduction turn (e.g. 150k context dropping to 40k after context compaction); assert delta clamps to zero and triggers no runaway alert.
- Test Case 4: Tool Thrashing Under Sliding Window and Sequence:
  - Test four consecutive failures of tool bash (triggers WARNING).
  - Test five consecutive failures of tool bash (triggers CRITICAL).
  - Test a window of five calls containing three bash failures and one intervening bash success; assert failure accumulation reset prevents thrashing alert.
- Test Case 5: Composite Evaluator Stability and Sorting:
  - Pass nil tree and nil trace to EvaluateLiveAnomalies; assert return is empty non-nil slice with zero panics.
  - Trigger all four anomaly rules simultaneously; assert returned slice is strictly ordered: CRITICAL alerts first, HIGH alerts second, WARNING alerts third; alerts of equal severity ordered by RuleName ascending, then Timestamp descending.

### HTTP and Adapter Boundary Tests (backend/internal/adapters/http/ and live/)
- Test Case 6: Typed Nil-Slice Serialization:
  - Verify that endpoints /api/live/sessions and /api/live/sessions/:id/trace return raw JSON containing square brackets [] for root_sessions, active_anomalies, messages, recent_errors, and anomalies when backend fields are uninitialized nil slices.
- Test Case 7: Real-Time SSE Stream Anomaly Delivery:
  - Connect a test client to /api/live/stream.
  - Publish an anomaly alert event on the service channel; verify client receives event type anomaly_alert with valid JSON payload.
  - Publish an identical anomaly alert on the subsequent polling cycle; verify the client receives zero duplicate events due to seenAnomalies deduplication.

### Web Frontend Unit Tests (web/src/domain/live.test.ts)
- Test Case 8: Live Domain Interface Round-Trip Deserialization:
  - Deserialize complete JSON payloads representing LiveSessionTree, LiveSessionNode, LiveTraceView, LiveMessageFact, and LiveAnomalyAlert.
  - Verify properties including parent_id, output_snippet, and reasoning_content are correctly mapped and accessible.
  - Deserialize payloads omitting optional properties; verify all fields degrade cleanly to undefined or null without throwing exceptions.

## 6. Review Criteria for Reviewer

### Axis 1: Spec and Contract Alignment
- Completeness of Rules: Are all four rules (ConsecutiveToolErrors, ContextCeilingApproach, TokenVelocityRunaway, ToolThrashing) fully grounded in domain code and verified by unit tests?
- Metric Thresholds: Do threshold values in domain/live_anomaly.go match CONTRACT.md (streak 3/5, ceiling 80%/100% of 200k, velocity 30k/60k tokens and 40k/80k delta, thrashing 3/5 failures)?
- Deterministic Ordering: Does SortLiveAnomalies strictly enforce multi-key ordering (Severity descending, RuleName ascending, Timestamp descending, SessionID ascending, TriggerEvent ascending)?
- SSE Alert Contract: Does the SSE stream emit event: anomaly_alert with deduplication against active anomalies?
- Frontend Type Parity: Does web/src/domain/live.ts match the Go domain models for LiveSessionNode (including parent_id), LivePartFact (including output_snippet), and LiveMessageFact (including reasoning_content)?

### Axis 2: Engineering Standards and Quality Invariants
- FCIS Discipline: Does backend/internal/domain/ remain 100% pure with zero I/O imports, zero database calls, and zero time.Now() invocations?
- Strict Typing: Are Go production and test files completely free of any and interface{}? Is TypeScript completely free of any?
- Zero Emojis: Are all files, comments, docstrings, tests, and error strings completely free of emojis?
- Zero Test Skips: Are there zero occurrences of t.Skip in the Go test suite and zero skipped tests in Vitest?
- Quality Gates: Do make check, make test, and make build pass cleanly without errors or warnings?
