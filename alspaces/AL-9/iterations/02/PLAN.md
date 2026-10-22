# Implementation Plan: AL-9 Visual Trace Waterfall and Anomaly Inspector (Iteration 02 Delta)

## 1. Iteration Goal and Steering Feedback Scope

This iteration incorporates follow-up steering feedback for ticket AL-9 following main branch synchronization. It refines trace waterfall ergonomics, expands error payload fidelity, and removes spurious anomaly diagnostic states.

Iteration 02 targets four concrete deliverables:
1. Default Filter Selection: Set the Trace Waterfall component to default to the "Errors and Anomalies Only" tab on initial render.
2. Error Payload Expansion: Expand the OpenCode SQLite event extractor limit from 200 characters to 4096 characters, update Go boundary tests, and adjust frontend drawer CSS for large multi-line payloads and scroll performance.
3. Event Selection and Drawer Decoupling: Decouple timeline node selection from diagnostic drawer opening. Clicking unflagged events (turns without errors or anomalies) must highlight the timeline item without opening the drawer. Only flagged turns (turns with execution errors or correlated anomalies) or explicit badge clicks open the drawer.
4. AnomalyDrawer Truth in Advertising: Remove the fake fallback title "Operational Anomaly Detected" and the artificial rule identifier "ANOMALY_DIAGNOSTIC". Conditionally render the Rule Diagnosis section only when a genuine domain anomaly exists. When viewing an unflagged tool error, title the drawer accurately as "Tool Execution Failure" and display the full error payload without fake diagnosis.

This plan adheres strictly to the Zero-Code Mandate (100% descriptive prose and structured specifications) and the Zero-Emoji Mandate.

---

## 2. Grounded Reference Architecture and Concrete File Paths

### Go Backend Imperative Shell
- File: backend/internal/adapters/opencode/extractor.go
  - Function: ExtractFacts
  - Modification: In the tool call extraction branch, expand the error excerpt truncation bound from 200 bytes to 4096 bytes.
- File: backend/internal/adapters/opencode/extractor_test.go
  - Function: TestExtractFacts_ToolCallsAndDurations
  - Modification: Update long error string test fixture from 250 characters to exceed 4096 characters (e.g. 5000 characters), and assert that the extracted ErrorMessage string length equals 4096.

### TypeScript Frontend Presentation Layer
- File: web/src/components/TraceWaterfall.tsx
  - Component: TraceWaterfall
  - Modification: Set initial state of filterMode to errors_anomalies. Ensure clicking an unflagged event highlights the event without invoking drawer-opening actions.
- File: web/src/components/AnomalyDrawer.tsx
  - Component: AnomalyDrawer
  - Modification: Eliminate fallback title "Operational Anomaly Detected" and fake rule ID "ANOMALY_DIAGNOSTIC". Conditionally render the Rule Diagnosis section only when anomaly is present. Title tool execution errors as "Tool Execution Failure".
- File: web/src/index.css
  - Selectors: .drawer-error-box, .drawer-error-content
  - Modification: Expand max-height to 380px, configure vertical and horizontal scrolling, and add scrollbar styling to accommodate up to 4096 characters cleanly.

### TypeScript Frontend View Integration Layer
- File: web/src/views/TicketDeepDive.tsx
  - Component: TicketDeepDive
  - Function: handleSelectEvent
  - Modification: Update handler logic so selecting an event always updates selectedEvent, but only sets isDrawerOpen to true when the event has a matching anomaly or an error status. If an unflagged event is selected, keep isDrawerOpen false (and close the drawer if currently open).

### Test Suites
- File: backend/internal/adapters/opencode/extractor_test.go: Update long error truncation assertions to 4096 characters.
- File: web/src/components/TraceWaterfall.test.tsx: Update tests to verify errors_anomalies is the default filter mode, and assert that clicking unflagged events emits selection without triggering anomaly drawer inspection.
- File: web/src/components/AnomalyDrawer.test.tsx: Replace tests that asserted fake ANOMALY_DIAGNOSTIC fallback with tests asserting accurate "Tool Execution Failure" titling and omission of the Rule Diagnosis section when anomaly is undefined.
- File: web/src/views/TicketDeepDive.test.tsx: Add tests verifying that clicking an unflagged timeline item highlights it without opening the AnomalyDrawer, while clicking a tool error or anomaly badge opens the drawer.

---

## 3. Phased Implementation Tasks for Coder Subagent

### Phase 1: Backend Adapter Error Truncation Expansion

Task 1.1: Expand Error Excerpt Bound in backend/internal/adapters/opencode/extractor.go
- Locate the tool call parsing loop within ExtractFacts where isError is evaluated.
- Locate the condition checking if length of errMsg exceeds 200.
- Update the threshold from 200 to 4096 characters.
- When length exceeds 4096, truncate errMsg to the first 4096 characters (errMsg[:4096]).
- Preserve existing logic for extracting non-empty trimmed output and falling back to tool state status strings when output is blank.

Task 1.2: Update Extractor Tests in backend/internal/adapters/opencode/extractor_test.go
- In TestExtractFacts_ToolCallsAndDurations, update the synthetic error string generation for p_long_err.
- Generate a test payload of 5000 characters (e.g. "Error: " prefix followed by repeated padding characters).
- Update test assertion to verify that len(*longErrEv.ErrorMessage) equals 4096.
- Update test comments and documentation strings referencing the previous 200 character limit to 4096 characters.

### Phase 2: Frontend Presentation Refinements

Task 2.1: Default to Errors and Anomalies in web/src/components/TraceWaterfall.tsx
- In TraceWaterfall, initialize the filterMode state to errors_anomalies instead of all.
- Verify that on initial render, the "Errors and Anomalies Only" button receives the active CSS class.
- Verify that filteredEvents initially filters sortedEvents to only those where event status is error or findMatchingAnomaly returns a valid anomaly.
- Ensure that switching to "All Events" or "Tool Calls Only" works as expected.

Task 2.2: Decouple Unflagged Event Clicks in web/src/components/TraceWaterfall.tsx
- Retain the onSelectEvent callback signature accepting the EventFact and optional matching Anomaly.
- Maintain an internal selectedEvent fallback state when the selectedEvent prop is not provided, allowing standalone highlighting.
- When a timeline item is clicked or activated via keyboard:
  - If the item has a matching anomaly or its status equals error, invoke onSelectEvent passing the event and the anomaly.
  - If the item is unflagged (status is not error and no matching anomaly exists), invoke onSelectEvent passing the event and undefined for anomaly.
- When the AnomalyBadge inside an item is explicitly clicked, invoke onSelectEvent passing the event and the matching anomaly, stopping click propagation to avoid duplicate triggers.

Task 2.3: Accurate Titling and Conditional Rule Diagnosis in web/src/components/AnomalyDrawer.tsx
- Remove the constants or default fallbacks for "Operational Anomaly Detected" and "ANOMALY_DIAGNOSTIC".
- Derive the drawer title dynamically:
  - If anomaly is present and has a title, use anomaly.title.
  - If anomaly is absent and event is present with status error, use "Tool Execution Failure" (or "Execution Failure" if tool_name is absent).
  - If both are absent, fallback to a neutral label such as "Execution Details".
- Conditional Rule Diagnosis rendering:
  - Wrap the entire "Rule Diagnosis" section (including the Rule ID label and description paragraph) in a conditional check so it renders only when anomaly is defined.
  - When anomaly is undefined, the Rule Diagnosis section must not appear in the DOM.
- Context Metadata Grid:
  - Continue extracting turn, session ID, tool name, subagent role, duration, timestamp, and context tokens from event or anomaly context.
  - Only render rule-specific metrics (Failure Count, Total Calls, Failure Rate %, Threshold, Token Delta) when defined on anomaly context.
- Severity Badge in Drawer Header:
  - Only render AnomalyBadge in the header when anomaly is present. Do not render a severity badge for plain tool errors lacking a rule classification.
- Error Payload Section:
  - Ensure the Error Payload section renders whenever event.error_message is present, styled with the expanded scrollable container.

Task 2.4: Expand Drawer Error Styling and Scrolling in web/src/index.css
- Locate .drawer-error-box and .drawer-error-content.
- In .drawer-error-content:
  - Increase max-height from 250px to 380px to provide ample vertical viewport space for up to 4096 characters of stack trace or error output.
  - Ensure white-space is pre-wrap and word-break is break-word to prevent horizontal blowout while preserving indentation.
  - Set overflow-y to auto and overflow-x to auto.
  - Add line-height of 1.5 and font-size of 0.82rem for enhanced readability.
  - Add custom scrollbar styling using CSS pseudo-elements (::-webkit-scrollbar, ::-webkit-scrollbar-thumb, ::-webkit-scrollbar-track) matching the dark theme surface tokens.

### Phase 3: View Controller Integration

Task 3.1: Update Selection Logic in web/src/views/TicketDeepDive.tsx
- In handleSelectEvent:
  - Inspect the incoming event and optional anomaly.
  - Determine if the turn is flagged: true if anomaly is truthy OR if event.status equals error.
  - Always update setSelectedEvent(event) so the clicked turn is highlighted in the waterfall timeline track.
  - If the turn is flagged:
    - Set setSelectedAnomaly(anomaly ?? null).
    - Set setIsDrawerOpen(true).
  - If the turn is unflagged:
    - Set setSelectedAnomaly(null).
    - Set setIsDrawerOpen(false).

---

## 4. Adversarial Testing Plan for Tester Subagent

The tester subagent must verify all changes with zero silent skips (no test.skip, no it.skip, no t.Skip).

### Suite 1: Go Backend Extractor Tests (backend/internal/adapters/opencode/extractor_test.go)
1. Verify 4096-Character Truncation: Insert a tool part with an error message of 5000 characters. Call ExtractFacts and assert that the extracted ErrorMessage has a length of exactly 4096 characters.
2. Verify Sub-4096 Integrity: Insert a tool part with an error message of 1500 characters. Assert that the entire 1500 characters are preserved without premature truncation.
3. Verify Clean and Empty Errors: Verify tool executions with status error but empty output fall back to status text without errors.
4. Run: go test -v -race ./backend/internal/adapters/opencode/...

### Suite 2: TraceWaterfall Component Tests (web/src/components/TraceWaterfall.test.tsx)
1. Default Filter Mode: Assert that upon mounting with a list containing both normal and error events, the "Errors and Anomalies Only" tab is active by default, and only events with errors or matching anomalies are rendered.
2. Tab Switching: Click "All Events" button and assert that all events (including successful, unflagged events) become visible.
3. Unflagged Event Click: Click a successful turn (turn 1, status success, no anomaly). Assert that onSelectEvent is called with anomaly as undefined.
4. Flagged Event Click: Click an error turn or anomalous turn. Assert that onSelectEvent is called with the event and corresponding anomaly.
5. Anomaly Badge Click: Click the AnomalyBadge within an anomalous turn. Assert that onSelectEvent is called with the matching anomaly.

### Suite 3: AnomalyDrawer Component Tests (web/src/components/AnomalyDrawer.test.tsx)
1. Genuine Anomaly Rendering: When isOpen is true and anomaly is provided, assert that the drawer title matches anomaly.title, AnomalyBadge displays the severity, and the Rule Diagnosis section displays the rule ID and description.
2. Tool Execution Failure Rendering: When isOpen is true, anomaly is undefined, and event has status error with an error_message:
   - Assert drawer title is "Tool Execution Failure".
   - Assert that "Rule Diagnosis" section is NOT present in the document.
   - Assert that "ANOMALY_DIAGNOSTIC" does NOT appear anywhere in the document.
   - Assert that "Operational Anomaly Detected" does NOT appear anywhere in the document.
   - Assert that the Error Payload section is present and renders the full error message text.
3. Large Payload Handling: Render an error message of 4096 characters. Assert the component renders without crashing and the text content matches.
4. Neutral Fallback: When both anomaly and event are undefined, assert drawer renders a neutral title without fake anomaly diagnostics.

### Suite 4: View Integration Tests (web/src/views/TicketDeepDive.test.tsx)
1. Unflagged Timeline Node Click: Render TicketDeepDive with v2 ticket AL-2. Click an unflagged timeline node (e.g. Turn 1). Assert that the node receives the selected highlight class, but screen.queryByRole('dialog') remains null (drawer does NOT open).
2. Flagged Timeline Node Click: Click a timeline node with status error (or click the warning badge). Assert that screen.getByRole('dialog') appears (drawer opens).
3. Plain Error Node Inspection: Click a tool error node that has no correlated anomaly rule. Assert drawer opens with title "Tool Execution Failure" and does NOT display "Rule Diagnosis".
4. Closing Drawer via Normal Node: Open drawer via an anomalous node, then click a normal unflagged node in the waterfall. Assert the drawer closes and the normal node is highlighted.

### Boundary and Quality Gates
- Typecheck: npm run typecheck across web/ must complete with zero errors and zero any.
- Unit and Component Tests: npm test must pass 100% of tests with zero skips.
- Asset Build: npm run build must generate web/dist cleanly.
- Backend Verification: make check and make test must pass cleanly.

---

## 5. Reviewer Audit Checklist

The reviewer subagent must audit the resulting git diff along two strict axes:

### Axis 1: Spec and Feedback Alignment
- TraceWaterfall defaults to "Errors and Anomalies Only" on load.
- Extractor truncation limit in extractor.go is exactly 4096 (not 200, not unlimited).
- Extractor tests in extractor_test.go verify 4096 character boundary.
- Clicking an unflagged event in TraceWaterfall / TicketDeepDive highlights the node and does NOT open the AnomalyDrawer.
- AnomalyDrawer renders "Tool Execution Failure" when displaying a tool error without an anomaly rule.
- Fake fallback string "Operational Anomaly Detected" is completely removed.
- Fake rule identifier "ANOMALY_DIAGNOSTIC" is completely removed.
- Rule Diagnosis section is conditionally rendered only when an anomaly rule is present.
- Drawer error payload container supports vertical and horizontal scrolling up to 380px height.
- Zero silent test skips across all test suites.

### Axis 2: Engineering Standards (TECHNICAL_GUIDELINES.md)
- FCIS boundary discipline: extractor.go remains in adapters; pure models in domain remain untouched; presentation components remain pure.
- Static typing: zero occurrences of any or unsafe type assertions.
- Immutability: readonly props and arrays preserved.
- Zero emojis: zero emojis across Go code, TypeScript components, CSS, tests, docstrings, and commit messages.
- Single binary parity: web/dist builds cleanly for embedding into Go executable.
