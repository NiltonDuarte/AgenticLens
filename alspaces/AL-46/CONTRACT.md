# Contract: AL-46 Collapsible Failed Tool Calls in the Live Inspector

## 1. Problem Statement

The Live Trace Inspector currently renders every `FailedToolAlertCard` in a failed tool call section that precedes the chronological execution message stream. When a session has repeated tool failures, the expanded cards consume vertical space and can push the message stream out of view. Collapsing those details by default should preserve immediate visibility of the failure state and its count.

## 2. Goal and Acceptance Boundaries

When failed tool calls are present, the inspector presents a compact, prominent failure summary that is collapsed by default. A keyboard-accessible toggle reveals or hides the existing failure cards. The message stream remains immediately visible after the compact summary. When no failed tool calls are present, the failure section is omitted.

Acceptance criteria:

1. The failed tool section is present only when the existing trace data indicates failed tool calls and recent errors are available.
2. The section starts collapsed on initial render.
3. Its visible summary includes the exact failed tool count in a high-visibility error badge.
4. Its toggle button uses `data-testid="failed-tools-toggle-btn"`, exposes its current expanded state accessibly, and supports standard button keyboard activation with Enter and Space.
5. `FailedToolAlertCard` instances are not rendered while collapsed and render when expanded.
6. Toggling in either direction updates the list without disrupting the message stream below it.
7. Existing failed tool detail rendering, including matching anomaly streaks and message metadata, remains intact.
8. Frontend typecheck, lint, and tests pass.

## 3. Architectural Invariants

- This is a presentation-only change. Do not change Go domain models, Go adapters, APIs, network behavior, or trace schemas.
- Reuse the existing `LiveTraceView.failed_tool_calls` and `recent_errors` data and existing `FailedToolAlertCard` behavior.
- Keep expansion state local to `LiveTraceInspector`; initialize it to collapsed.
- Use native semantic HTML controls and React state. Do not introduce custom keyboard event handling where a native button provides the required behavior.
- Preserve strict TypeScript typing and the repository-wide zero `any` / `interface{}` and zero emoji requirements.

## 4. Scope Boundaries

### In scope

- `web/src/components/LiveTraceInspector.tsx`: local collapsed state, summary header, accessible toggle, conditional rendering of cards.
- `web/src/index.css`: layout and styling for the summary bar, prominent count badge, and toggle affordance.
- `web/src/components/LiveTraceInspector.test.tsx`: coverage for default collapsed state, exact count, message stream visibility, toggle behavior, empty state, and keyboard activation.

### Out of scope

- Backend/domain data changes or changes to how failed tools are detected or counted.
- Changes to `FailedToolAlertCard` content or its error/anomaly classification.
- Persisting expansion state across sessions, inspector remounts, or page reloads.
- Changes to other inspector sections or unrelated live observability behavior.

## 5. Grounded Existing Components

- `web/src/components/LiveTraceInspector.tsx` owns the relevant section, currently guarded by `hasFailedTools && trace && trace.recent_errors.length > 0`, renders the heading using `trace.failed_tool_calls`, and maps `trace.recent_errors` to `FailedToolAlertCard`.
- `web/src/components/FailedToolAlertCard.tsx` renders the existing failure detail card and remains unchanged.
- `web/src/components/LiveTraceInspector.test.tsx` contains the inspector component tests and fixtures for trace presentation.
- `web/src/index.css` defines `.inspector-failed-tools-section`, `.failed-tools-heading`, `.failed-tools-cards-list`, and `.inspector-messages-section` styles.

## 6. Verification

Run the web package's `npm run typecheck`, `npm run lint`, and `npm test`. Also inspect the final diff for accessibility state/labels, strict typing, preservation of current error card props, and absence of unrelated changes.
