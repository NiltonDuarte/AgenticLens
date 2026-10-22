# Implementation Delta Plan: AL-20 Iteration 02 - Eliminate Runtime Mock Mode from Live Observability

## 1. Goal and Iteration Scope

Iteration 01 established the reactive live observability frontend: SSE client streaming, fallback polling, hierarchical session tree inspection, under-the-hood message trace inspection, failed tool alert cards, and real-time anomaly banners. However, Iteration 01 included a runtime mock mode switch and silent synthetic mock data fallbacks in the live API adapter. If the Go backend server is unreachable, the client silently substitutes mock data instead of alerting the user.

The goal of Iteration 02 is to eliminate runtime mock mode and silent synthetic mock data fallbacks from the live observability subsystem. The dashboard must interact directly with real backend endpoints (/api/live/stream, /api/live/sessions, /api/live/sessions/{id}/trace). When connectivity fails, the application must emit typed errors, transition to an explicit error state, and display a prominent warning banner with the exact message "Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running." along with an interactive "Retry Connection" button. Mock data in mockData.ts will be strictly reserved as static test fixtures for Vitest test suites.

## 2. Grounded Reference Architecture and File Paths

The delta implementation touches the following files:

- Client Network and Boundary Layer:
  - web/src/api/liveApi.ts: Remove mockData imports. Remove isMock parameter and mock dispatch branches from subscribeToLiveStream, fetchLiveSessionTree, and fetchLiveTrace. Convert silent catch-block fallbacks to thrown typed Error instances and status error transitions.
  - web/src/api/mockData.ts: Unlink from production runtime bundle. Retain exclusively as Vitest test fixtures.
  - web/src/api/liveApi.test.ts: Remove tests for isMock mode. Add adversarial tests asserting that fetchLiveSessionTree and fetchLiveTrace reject with descriptive error messages on 404, 500, and network dropouts. Assert that subscribeToLiveStream propagates real error objects to onError and marks connection status as error when fallback polling fails.
- UI Presentation and Views:
  - web/src/views/LiveObservability.tsx: Remove isMock state, onMockChange callback handling, mock toggle button, and mock data seeding effect. Add connection failure banner rendering "Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running." with an active "Retry Connection" button. Maintain the existing empty state for when the live backend reports zero active sessions.
  - web/src/views/LiveObservability.test.tsx: Remove mock mode toggle tests. Add integration tests for explicit server failure banners, retry connection clicks, and zero-session empty states.
  - web/src/App.tsx: Remove onMockChange prop passing from the LiveObservability view branch.
  - web/src/index.css: Add dark-theme styles for the live connection failure banner and retry button.
- Verification and Quality Gates:
  - Makefile: make test-frontend, npm run typecheck, npm run lint, and make build.

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Boundary Layer Mock Mode Purge (web/src/api/liveApi.ts)
- Task 1.1: Remove mock data imports from web/src/api/liveApi.ts. Delete the import of mockLiveAnomalyAlerts, mockLiveSessionTree, mockLiveTraceView, and mockLiveTraceViews from ./mockData.
- Task 1.2: Refactor fetchLiveSessionTree in web/src/api/liveApi.ts:
  - Remove the isMock parameter signature entirely so it accepts no arguments and returns Promise of LiveSessionTree.
  - Make a GET request to /api/live/sessions using fetch.
  - If the response is not ok (res.ok is false), throw a typed Error describing the HTTP failure with the status code and status text.
  - Parse the JSON body and pass it through parseLiveSessionTree.
  - In the catch block, do not catch and swallow errors. Re-throw a typed Error so callers can distinguish between network failure and successful zero-session responses.
- Task 1.3: Refactor fetchLiveTrace in web/src/api/liveApi.ts:
  - Remove the isMock parameter signature so it accepts only sessionId: string and returns Promise of LiveTraceView.
  - Clean and encode the session identifier.
  - Make a GET request to /api/live/sessions/{encodedId}/trace using fetch.
  - If the response is not ok, throw a typed Error with a descriptive message including the session ID and HTTP status code.
  - Parse the JSON body and pass it through parseLiveTraceView.
  - Re-throw any caught network errors instead of returning fallback mock data.
- Task 1.4: Refactor subscribeToLiveStream in web/src/api/liveApi.ts:
  - Remove the isMock parameter so the function signature is subscribeToLiveStream(callbacks: LiveStreamCallbacks): () => void.
  - Remove the entire if (isMock) code block, including mock intervals, mock anomaly loops, and synthetic status events.
  - In startFallbackPolling, execute the poll loop against fetchLiveSessionTree. If polling fails, pass the resulting Error to callbacks.onError. If multiple consecutive polling failures occur, invoke callbacks.onStatusChange with 'error'.
  - In connectSSE, when EventSource is unavailable or throws, report the error via callbacks.onError and transition to fallback polling. If fallback polling also fails, transition to 'error'.
  - Ensure the returned teardown function cancels all timers and closes any active EventSource instance.

### Phase 2: View Error Banner and Interaction Updates (web/src/views/LiveObservability.tsx and web/src/App.tsx)
- Task 2.1: Clean up state and interface in web/src/views/LiveObservability.tsx:
  - Remove mockLiveSessionTree import from ./api/mockData.
  - In LiveObservabilityProps, remove onMockChange from the interface.
  - Remove isMock state (useState) and handleToggleMock function.
  - Remove the useEffect that seeds mockLiveSessionTree when isMock is active.
  - Remove the mock mode toggle button from the header status bar.
- Task 2.2: Implement Live Server Connection Error Banner:
  - Introduce retry logic in LiveObservability: create a handleRetryConnection callback that resets error messages, sets connectionStatus to 'connecting', and re-subscribes to live streaming or refetches the tree.
  - When connectionStatus is 'error' or an unrecoverable connection failure occurs, display a prominent error banner above the split-pane layout.
  - The banner container must feature a distinct role of alert and data-testid of live-connection-error-banner.
  - The banner text must state: "Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running."
  - Include an actionable button within the banner labeled "Retry Connection" that calls handleRetryConnection when clicked.
- Task 2.3: Preserve Empty State for Zero Active Sessions:
  - Verify that when the live backend is running and returns an empty session tree (total_active_sessions is 0 and root_sessions is empty), LiveSessionTree renders its empty state card without triggering the connection error banner.
- Task 2.4: Update web/src/App.tsx:
  - Update the render block for activeTab === 'live' to invoke LiveObservability with only onSelectTicket={handleSelectTicketFromList}, removing onMockChange.

### Phase 3: Styling Integration (web/src/index.css)
- Task 3.1: Add CSS rules for the live connection error banner in web/src/index.css:
  - Class .live-connection-error-banner: red-tinted background, solid border in status-red, flex layout aligning warning icon, text description, and retry action.
  - Class .live-retry-btn: prominent button with hover highlight and accessible focus outline.

### Phase 4: Test Suite Modernization (web/src/api/liveApi.test.ts and web/src/views/LiveObservability.test.tsx)
- Task 4.1: Update web/src/api/liveApi.test.ts:
  - Remove the describe block for subscribeToLiveStream in Mock Mode.
  - Remove tests asserting isMock behavior in fetchLiveSessionTree and fetchLiveTrace.
  - Add tests verifying that fetchLiveSessionTree rejects when fetch fails with a network error.
  - Add tests verifying that fetchLiveSessionTree rejects when fetch returns HTTP 500 or 404.
  - Add tests verifying that fetchLiveTrace rejects when fetch fails or returns HTTP 404/500.
  - Update fallback polling tests to verify that polling errors propagate to callbacks.onError and transition status to 'error' on persistent network dropout.
- Task 4.2: Update web/src/views/LiveObservability.test.tsx:
  - Remove the test: toggles mock mode on and off invoking onMockChange callback.
  - Add test: displays connection error banner with retry button when stream and fallback polling fail to connect.
  - Add test: clicking retry connection button attempts reconnection and clears prior error banner.
  - Add test: renders empty session tree card when backend connects successfully but has zero active sessions, without showing the connection error banner.

## 4. Adversarial Testing Strategy for Tester

### Unit Tests for liveApi Boundary Layer (web/src/api/liveApi.test.ts)
- Test Case 1: fetchLiveSessionTree Network Rejection
  - Configure fetch mock to reject with TypeError("Failed to fetch").
  - Call fetchLiveSessionTree().
  - Assert the promise rejects with an Error instance and does not return mockData.
- Test Case 2: fetchLiveSessionTree HTTP 500 Error
  - Configure fetch mock to return status 500 and ok false.
  - Call fetchLiveSessionTree().
  - Assert the promise rejects with an Error mentioning HTTP 500.
- Test Case 3: fetchLiveTrace HTTP 404 Error
  - Configure fetch mock to return status 404 and ok false for session ID ses-missing.
  - Call fetchLiveTrace('ses-missing').
  - Assert the promise rejects with an Error mentioning session ID and status 404.
- Test Case 4: Fallback Polling Persistent Network Failure
  - Configure EventSource to error out immediately.
  - Configure fetch for /api/live/sessions to reject persistently.
  - Assert that onStatusChange is called with 'fallback-polling' followed by 'error'.
  - Assert that onError receives the network failure errors.
- Test Case 5: SSE Lifecycle without Mock Mode
  - Connect to /api/live/stream via subscribeToLiveStream.
  - Dispatch valid session_tree, anomaly_alert, and trace_update events.
  - Assert respective callbacks receive parsed domain data.

### Component and Integration Tests for LiveObservability View (web/src/views/LiveObservability.test.tsx)
- Test Case 6: Live Connection Failure Banner
  - Simulate EventSource throwing and fetch failing for /api/live/sessions.
  - Render LiveObservability.
  - Assert that live-connection-error-banner is displayed in the document.
  - Assert the banner contains exact text "Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running."
  - Assert the "Retry Connection" button is present.
- Test Case 7: Retry Connection Interaction
  - Start with failed connection displaying the error banner.
  - Update fetch mock to succeed with valid session tree data.
  - Click "Retry Connection" button.
  - Assert that reconnect logic runs and error banner is dismissed upon connection.
- Test Case 8: Zero Active Sessions Parity
  - Mock fetch to return ok true with root_sessions empty array.
  - Render LiveObservability.
  - Assert connection status is Connected or Fallback Polling.
  - Assert LiveSessionTree renders "No Active Sessions" empty state card.
  - Assert live-connection-error-banner is not present.
- Test Case 9: Absence of Mock Controls
  - Render LiveObservability.
  - Assert that no button with text matching /mock mode/i exists in the DOM.

### Quality and Technical Invariant Gates
- Strict Typing: Zero any or untyped unknown across modified files.
- Zero Emojis: Scan all files to verify zero emojis in UI strings, comments, and logs.
- Full Verification Suite: Execute make test-frontend, npm run typecheck, npm run lint, and make build.

## 5. Review Criteria for Reviewer

### Axis 1: Spec and User Feedback Alignment
- Is runtime mock mode completely removed from LiveObservability view and liveApi?
- Are silent fallbacks to mockData.ts eliminated on fetch failure or stream disconnection?
- Is mockData.ts isolated strictly to test fixtures for Vitest?
- Does the view display the exact requested error banner ("Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running.") with a "Retry Connection" button when backend connectivity fails?
- Does the dashboard correctly display the empty state when zero active sessions exist on a healthy backend?

### Axis 2: Engineering Standards and Technical Invariants
- FCIS Discipline: Is the domain layer pure with zero side-effects? Does liveApi handle network I/O cleanly with typed errors?
- Strict Typing: Is there zero usage of any across all modified code?
- Zero Emojis: Are code, UI copy, and tests completely free of emojis?
- Suite Integrity: Do all frontend unit, component, and integration tests pass without skips or warnings?
