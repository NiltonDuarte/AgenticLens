# Implementation Plan: AL-46 Collapsible Failed Tool Calls in the Live Inspector

## 1. Goal and Iteration Scope

Implement the full presentation slice defined by `../CONTRACT.md`: show failed tool calls in a default-collapsed, high-visibility summary, provide an accessible native toggle, preserve current error card details when expanded, and keep the chronological message stream visible beneath the summary.

## 2. Grounded References

- `web/src/components/LiveTraceInspector.tsx`: `LiveTraceInspector` currently renders `.inspector-failed-tools-section` before `.inspector-messages-section`; the section heading displays `trace.failed_tool_calls`, and `trace.recent_errors` maps to `FailedToolAlertCard` with anomaly streak and parent-message context.
- `web/src/components/FailedToolAlertCard.tsx`: existing card presentation to retain without modification.
- `web/src/components/LiveTraceInspector.test.tsx`: existing component test suite and rendering helpers to extend.
- `web/src/index.css`: existing `.inspector-failed-tools-section`, `.failed-tools-heading`, `.failed-tools-cards-list`, and message section styles.

## 3. Implementation Tasks

### Task 1: Add default-collapsed local state and accessible summary toggle

- Add `isFailedToolsExpanded` state to `LiveTraceInspector`, initialized to `false`.
- Keep the current condition that omits the section when there are no failed tools or no recent error details.
- Replace the plain heading presentation with a compact summary header containing a prominent badge whose text reports the exact `trace.failed_tool_calls` count.
- Add a native `button` with `type="button"` and `data-testid="failed-tools-toggle-btn"`; use its `aria-expanded` value to communicate state and a relationship to the collapsible card list.
- Conditionally render the existing cards list only while expanded. Preserve all existing card props and metadata derivation.
- Leave the execution message section in its current position immediately after the failure section.

### Task 2: Style the summary row and failure badge

- In `web/src/index.css`, style the summary header as a compact, aligned row that remains legible at narrow inspector widths.
- Give the failure count badge a prominent existing error color treatment and clear contrast.
- Style the toggle as a visible, focusable control with hover and focus-visible affordances consistent with existing inspector controls.
- Ensure expanded cards retain their current vertical spacing and layout.

### Task 3: Extend inspector tests

- Verify the failed tool section is collapsed on initial render and the card details are absent.
- Verify the badge shows the exact failure count and the toggle exposes the collapsed state.
- Verify the execution message stream is visible in the same initial render below the compact summary.
- Verify clicking the toggle reveals the existing cards and clicking again hides them.
- Verify a zero-failure trace omits the section.
- Verify native keyboard activation with Enter and Space expands/collapses the section and that state remains accurately reflected by `aria-expanded`.

## 4. Verification

- Run `npm run typecheck`, `npm run lint`, and `npm test` from `web/`.
- Confirm tests exercise both directions of toggle interaction, the exact count, default state, empty state, and message stream visibility.
- Review the diff for strict TypeScript, accessible semantics, unchanged card metadata, and zero unrelated backend or API changes.
