# Implementation Plan: FIX-3 Resolution of False Tool Errors, Anomaly Misattributions, Waterfall Collapsibility, and Debug Identifiers

## 1. Goal and Iteration Scope

Deliver complete end-to-end resolution of ticket FIX-3 across Go backend and TypeScript frontend:
- Prevent false error classifications on successful executions of content inspection tools (read, grep, glob, webfetch) in backend/internal/domain/live.go, backend/internal/adapters/live/service.go, and backend/internal/adapters/opencode/extractor.go.
- Enforce per-session scoping for tool retry streaks in backend/internal/domain/anomalies.go.
- Add MessageID to EventFact domain models in backend/internal/domain/models.go and web/src/types/telemetry.ts, propagating it from SQLite extraction.
- Restrict TOOL_RETRY_STREAK and tool failure anomaly matching in web/src/components/TraceWaterfall.tsx strictly to failed events (status === error) with matching session IDs.
- Make the Execution Trace Waterfall collapsible and collapsed by default, displaying a concise summary while collapsed and the full interactive view when expanded.
- Provide concrete debug identifiers with Copy to Clipboard actions across web/src/components/AnomalyDrawer.tsx, web/src/components/TraceWaterfall.tsx, and web/src/components/FailedToolAlertCard.tsx.
- Update Go and TypeScript test suites to assert all new behaviors, edge cases, and zero emojis.

---

## 2. Grounded Reference Architecture and File Paths

The plan grounds against existing concrete packages, files, and symbols:

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/live.go:
  - ClassifyToolFailure function.
  - Severity enum constants (SeverityNone, SeverityLow, SeverityMedium, SeverityHigh).
- backend/internal/domain/anomalies.go:
  - detectToolRetryStreaks function.
  - RuleToolRetryStreak constant and Anomaly struct.
- backend/internal/domain/models.go:
  - EventFact struct. Add MessageID (*string) with JSON serialization tag.
- backend/internal/domain/live_test.go:
  - TestClassifyToolFailure and TestClassifyToolFailure_AdversarialSuite test suites.
- backend/internal/domain/anomalies_test.go:
  - TestDetectToolRetryStreaks test cases.

### Go Boundary Adapters (backend/internal/adapters/)
- backend/internal/adapters/live/service.go:
  - processRawPart and convertToEventFact functions.
  - PartRow struct and consecutive error escalation logic.
- backend/internal/adapters/opencode/extractor.go:
  - ExtractFacts function.
  - Scanning part rows for message_id (mid sql.NullString).
  - Tool accumulator and part error evaluation logic.
- backend/internal/adapters/live/sqlite_test.go:
  - Unit tests verifying live part extraction and error classification.
- backend/internal/adapters/opencode/extractor_test.go:
  - Unit tests verifying tool fact and event fact extraction from SQLite parts.

### TypeScript Domain and Types (web/src/types/, web/src/domain/)
- web/src/types/telemetry.ts:
  - EventFact interface. Add message_id property.
  - Anomaly interface, AnomalyContext interface, AnomalyRuleID union.
- web/src/domain/live.ts:
  - LivePartFact and LiveMessageFact interfaces.

### TypeScript Presentation Components (web/src/components/, web/src/views/)
- web/src/components/TraceWaterfall.tsx:
  - findMatchingAnomaly helper function.
  - Collapsible waterfall state (isExpanded boolean defaulted to false).
  - Summary collapsed card rendering.
  - Timeline item rendering with concrete debug identifier and Copy button.
- web/src/components/AnomalyDrawer.tsx:
  - Diagnostic Context metadata section.
  - Concrete identifiers (Session ID, Message ID, Turn Debug ID) with Copy to Clipboard action.
- web/src/components/FailedToolAlertCard.tsx:
  - Header title group and identifier display with copy action.
- web/src/views/TicketDeepDive.tsx:
  - TraceWaterfall integration container.
- web/src/components/TraceWaterfall.test.tsx:
  - Existing and new test suites covering collapsibility, anomaly matching, session scoping, and copy action.
- web/src/components/AnomalyDrawer.test.tsx:
  - Existing and new test suites covering identifier rendering and clipboard interaction.
- web/src/components/FailedToolAlertCard.test.tsx:
  - Unit tests covering alert rendering and copy actions.
- web/src/views/TicketDeepDive.test.tsx:
  - Integration tests verifying TraceWaterfall rendering and interaction within the deep dive view.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Core Fixes
- Task 1.1: In backend/internal/domain/live.go, update ClassifyToolFailure:
  - Add inspection check for content reading and inspection tools (read, grep, glob, webfetch) using case-insensitive matching.
  - If status is "completed" and toolName is a content inspection tool:
    - Return false for isError, SeverityNone for severity, and nil for errorMessage.
    - Explicitly bypass compilation error, traceback, missing resource, and command failure scans on the output text.
  - Preserve existing failure classification rules for non-inspection tools (such as bash, python, edit) and for calls whose status is "error" or "failed".
- Task 1.2: In backend/internal/domain/anomalies.go, update detectToolRetryStreaks:
  - Introduce activeSessionID tracking variable alongside activeTool and consecutiveErrors.
  - When encountering an event with status "error":
    - If event.ToolName matches activeTool and event.SessionID matches activeSessionID:
      - Increment consecutiveErrors, update lastTurn.
    - If either tool or session differs:
      - Emit any streak meeting the threshold for the prior session and tool.
      - Reset activeTool, activeSessionID, and initialize consecutiveErrors to 1.
  - When encountering an event with non-error status:
    - Emit any active streak meeting threshold.
    - Reset activeTool, activeSessionID, and consecutiveErrors to zero.
  - Finalize and emit any eligible streak after processing all events.
- Task 1.3: In backend/internal/domain/models.go:
  - Extend EventFact struct with field MessageID of type *string with struct tag json message_id,omitempty.

### Phase 2: Go Boundary Adapters Fixes
- Task 2.1: In backend/internal/adapters/live/service.go:
  - In processRawPart, verify consecutive error counter escalation: ensure that when ClassifyToolFailure returns false, consecutiveErrors resets to zero and status remains completed without being forced to error.
  - In convertToEventFact, map pr.MessageID into EventFact.MessageID when non-empty. Ensure tool calls with status completed for content inspection tools retain status completed and nil error message.
- Task 2.2: In backend/internal/adapters/opencode/extractor.go:
  - In part row processing loop, assign mid.String to EventFact.MessageID when mid.Valid and non-empty.
  - Update part error determination logic: when tool name is a content inspection tool (read, grep, glob, webfetch) and raw part status is completed (or not error), do not scan out or outPrefix for "Error:" or "Traceback". Keep isError false, eventStatus as "completed", and do not increment tool error accumulators.

### Phase 3: Go Backend Test Suites and Adversarial Verification
- Task 3.1: In backend/internal/domain/live_test.go:
  - Add test cases in TestClassifyToolFailure asserting that read tool with status "completed" and output containing "syntax error: unexpected token" returns isError=false and SeverityNone.
  - Add test cases asserting that read tool reading markdown with "Traceback", "panic:", "build failed", or "Error:" returns isError=false and SeverityNone when status is "completed".
  - Add test cases asserting that grep, glob, and webfetch with status "completed" and empty or diagnostic output return isError=false and SeverityNone.
  - Update existing test cases that expected completed glob or grep with zero matches to trigger low severity, ensuring they now expect isError=false and SeverityNone.
  - Assert that read tool with status "error" and output "no such file or directory" continues to return isError=true and appropriate severity.
- Task 3.2: In backend/internal/domain/anomalies_test.go:
  - Add unit test verifying that consecutive tool errors split across different session IDs do not combine into a single streak anomaly.
  - Assert that a streak of 3 errors in session 1 and 2 errors in session 2 emits only the streak for session 1.
- Task 3.3: In backend/internal/adapters/opencode/extractor_test.go:
  - Add test asserting that a read tool part with status "completed" and output containing "Error: file parsing failed in markdown" yields an EventFact with status "completed" and does not increment tool errors.
  - Verify that EventFact instances emitted by ExtractFacts include populated MessageID matching the inserted part message_id.
- Task 3.4: Execute Go test suite:
  - Run go test -race ./backend/... and go vet ./backend/... ensuring clean passes.

### Phase 4: TypeScript Telemetry Models and Pure Matching Logic
- Task 4.1: In web/src/types/telemetry.ts:
  - Add readonly message_id?: string | null | undefined to EventFact interface.
- Task 4.2: In web/src/components/TraceWaterfall.tsx, rewrite findMatchingAnomaly:
  - When comparing an anomaly candidate against an event:
    - Enforce Session ID Guard: if anomaly context session_id is defined and non-empty, and event session_id is defined and non-empty, verify they match strictly (ctx.session_id === event.session_id). If they do not match, return false.
    - Enforce Failure Rule Status Guard: if anomaly rule_id is TOOL_RETRY_STREAK, TOOL_HIGH_FAILURE_RATE, or any tool failure rule, require event.status === error. If event status is completed or success, return false.
    - Match tool name case-insensitively and verify turn index if ctx.turn is provided.
    - If ctx.turn is provided without tool name, require turn and session to match strictly.
    - Maintain existing role matching logic with session guard.

### Phase 5: Trace Waterfall Collapsible Section and Summary Presentation
- Task 5.1: In web/src/components/TraceWaterfall.tsx:
  - Add state isWaterfallExpanded initialized to false (collapsed by default).
  - Add optional prop defaultExpanded?: boolean (defaulting to false) to allow programmatic expansion in tests when needed.
  - Compute summary statistics: totalEvents count, totalErrors count (events where status === error), and totalAnomalies count (anomalies.length).
  - When isWaterfallExpanded is false:
    - Render a compact collapsed card container with class trace-waterfall-card and trace-waterfall-collapsed.
    - Render title "Execution Trace Waterfall".
    - Render source badge ("Source: Local SQLite" or "Source: Committed Ledger").
    - Render summary count tags: total events count and issue breakdown (for example, "3 Errors, 1 Anomaly" or "Zero Errors, Zero Anomalies").
    - Render expand button with test id "waterfall-expand-btn", label "Expand Trace Waterfall", and aria-expanded="false".
    - Do not render the filter button group, multiselect dropdowns, or timeline item list while collapsed.
  - When isWaterfallExpanded is true:
    - Render the full waterfall card with collapse button with test id "waterfall-collapse-btn", label "Collapse Trace Waterfall", and aria-expanded="true".
    - Render existing filter toolbar (All Events, Tool Calls Only, Errors and Anomalies Only, Role dropdown, Tool dropdown, Anomaly dropdown).
    - Render existing timeline track.
- Task 5.2: In web/src/index.css:
  - Add clean styling rules for .trace-waterfall-collapsed, .waterfall-summary-bar, .waterfall-summary-metrics, .waterfall-toggle-btn, and associated hover/focus states, maintaining consistent design tokens and zero emojis.

### Phase 6: Concrete Debug Identifiers and Copy to Clipboard Actions
- Task 6.1: Implement copy to clipboard utility helper or local handler:
  - Use navigator.clipboard.writeText with fallback to document.createElement('textarea') and document.execCommand('copy').
  - Maintain a temporary copied identifier state for 2 seconds before reverting.
  - Wrap in try/catch to defensively ignore clipboard rejection in restricted test environments.
- Task 6.2: In web/src/components/AnomalyDrawer.tsx:
  - In Diagnostic Context metadata grid:
    - Display Session ID with an adjacent "Copy" button.
    - If event.message_id is present, render a Message ID item with an adjacent "Copy" button.
    - Display Turn Debug ID (e.g. formatted as turn-[turnNumber] or composite [sessionId]:turn:[turnNumber]) with a "Copy" button.
    - Display visual confirmation ("Copied") upon activation.
- Task 6.3: In web/src/components/TraceWaterfall.tsx:
  - In timeline item top row or details row:
    - Display concrete debug identifier token (session ID, message ID if present, or turn debug ID).
    - Provide a "Copy ID" button with aria-label describing the identifier.
    - In expanded payload container, display message ID when available with copy button.
- Task 6.4: In web/src/components/FailedToolAlertCard.tsx:
  - In failed tool header, display part identifier or tool invocation ID with a "Copy ID" button and feedback tooltip.

### Phase 7: TypeScript Frontend Test Suites and Verification
- Task 7.1: In web/src/components/TraceWaterfall.test.tsx:
  - Update tests to account for collapsed-by-default behavior (e.g. clicking expand button or passing defaultExpanded={true} to access the timeline).
  - Add test asserting that TraceWaterfall renders collapsed summary by default with title, source badge, event count, error/anomaly totals, and expand button.
  - Add test asserting that clicking the expand button reveals filter toolbar and timeline items, and clicking collapse returns to summary view.
  - Add test verifying that findMatchingAnomaly does NOT match TOOL_RETRY_STREAK to an event with status "success" or "completed".
  - Add test verifying that findMatchingAnomaly does NOT match an anomaly when session_id differs, even if turn number matches.
  - Add test verifying that clicking the "Copy ID" button copies the debug identifier to clipboard and displays "Copied" state.
  - Verify zero emojis across all rendered waterfall output.
- Task 7.2: In web/src/components/AnomalyDrawer.test.tsx:
  - Add test verifying that Message ID renders in diagnostic context when present on event.
  - Add test verifying that clicking "Copy" on session ID or message ID copies text to clipboard and shows "Copied" confirmation.
  - Verify zero emojis across all drawer states.
- Task 7.3: In web/src/components/FailedToolAlertCard.test.tsx:
  - Add test verifying copy identifier functionality and clean presentation.
- Task 7.4: In web/src/views/TicketDeepDive.test.tsx:
  - Verify integration tests pass with collapsible waterfall.

### Phase 8: End-to-End Build and Verification Quality Gates
- Task 8.1: Run Go automated checks and test suites:
  - Run go test -race ./...
  - Run go vet ./...
- Task 8.2: Run TypeScript checks and test suites:
  - Run npm run typecheck
  - Run npm run lint
  - Run npm test
- Task 8.3: Run full single-binary compilation:
  - Run npm run build
  - Run go build -o bin/agentic-lens ./backend/cmd/agentic-lens
- Task 8.4: Verify engineering invariants:
  - Zero any across all modified Go and TypeScript files.
  - Zero emojis across all source files, tests, and documentation.
