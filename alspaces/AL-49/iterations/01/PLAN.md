# AL-49 Iteration 01 Plan

## Goal

Deliver latest-first chronology, reversible ordering, stable logical turn numbering, safe live-update scrolling, and direct list navigation in the Live Trace Inspector.

## Grounded Files and Symbols

- `web/src/domain/live.ts`: readonly `LiveTraceView.messages` and `LiveMessageFact` contracts.
- `web/src/components/LiveTraceInspector.tsx`: `LiveTraceInspector`, `LiveMessageTurnItem`, existing `trace.messages.map`, message section header, and current source-index-based turn context.
- `web/src/components/LiveTraceInspector.test.tsx`: ordering, turn badge, and inspector interaction coverage; update tests that presently assert chronological rendering.
- `web/src/index.css`: `.inspector-messages-section`, `.messages-heading`, and `.inspector-messages-list` styles.
- `TECHNICAL_GUIDELINES.md`: strict TypeScript, pure domain logic, shell-owned DOM interactions, and no `any` or emojis.

## Implementation Tasks

1. Add a pure message-ordering helper in `web/src/domain/` that returns the messages in the requested direction without modifying the readonly input. Keep the source index or derive the stable logical turn number before reordering, so the visible order never changes the turn identity.
2. In `LiveTraceInspector`, default the order state to latest-first and add an accessible toggle in the message section header with `data-testid="turn-order-toggle-btn"`. Clearly communicate the current ordering state and the alternate action. Keep empty-trace behavior intact.
3. Keep the message list in a stable scrollable container. Track the user's scroll position relative to the active ordering; when scrolled away from the newest edge past the 50px threshold, prevent incoming messages from displacing the visible reading position and show a keyboard-accessible “New turns available” action to return to latest. Clear the affordance when the latest edge is reached or selected. Handle first render, ordering changes, and new-message updates without unwanted jumps.
4. Add accessible “Scroll to Top” and “Scroll to Bottom” controls with `data-testid="scroll-to-top-btn"` and `data-testid="scroll-to-bottom-btn"`. They must target the message-list element and behave predictably for both order modes.
5. Extend `web/src/index.css` for the header controls, pending-new-turn affordance, and scrollable list while preserving the current visual hierarchy and responsive layout.
6. Update `web/src/components/LiveTraceInspector.test.tsx` and add focused pure-helper coverage if helper tests have an established location. Verify newest-first default, toggle to oldest-first, stable logical `Turn 1` through `Turn N` badges and prompt attribution, non-mutating ordering, paused-scroll behavior when new messages arrive, pending indicator jump behavior, empty trace controls, and both navigation handlers.

## Acceptance Checks

- `npm run typecheck`
- `npm run lint`
- `npm test`

Confirm all existing inspector behaviors continue to pass, particularly failed-tool turn attribution, reasoning controls, copy-prompt turn context, and empty states. Do not change backend files or API contracts.
