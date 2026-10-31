# AL-49 Technical Contract

## Problem

The Live Trace Inspector currently renders `LiveTraceView.messages` in source order, placing the oldest turn first. Operators following long sessions must repeatedly navigate to the latest activity. Changing the order must preserve each turn's logical identity and avoid disorienting viewport shifts when the trace updates.

## Scope

This ticket changes the web presentation and its tests only. It may add a pure TypeScript ordering helper under `web/src/domain/` that accepts the readonly messages array and returns an ordered readonly view without mutating the input. No Go domain, adapter, HTTP, telemetry schema, or backend behavior changes are included.

The presentation work is centered on `web/src/components/LiveTraceInspector.tsx` and its existing stylesheet in `web/src/index.css`. The inspector will default to latest-first ordering, expose an accessible order toggle, preserve source-order logical turn numbering, provide scroll-to-top and scroll-to-bottom controls, and keep newly arriving messages from moving the reader's viewport unexpectedly. A compact “new turns available” affordance that lets the reader jump to the latest turn is an acceptable scroll-lock implementation.

## Grounded Existing Behavior

- `web/src/components/LiveTraceInspector.tsx` renders `trace.messages.map((msg, idx) => ...)` directly and passes the array index to `LiveMessageTurnItem`; that index also determines visible turn badges, debug prompt turn numbers, and part prompt turn numbers.
- `LiveMessageTurnItem` uses message IDs as keys when available and currently uses its `turnIndex` as the logical source position.
- Failed tool cards derive their turn number by finding the parent message in the original `trace.messages` array. This must remain consistent with the badges after display ordering changes.
- `web/src/domain/live.ts` defines `LiveTraceView.messages` as a readonly `LiveMessageFact[]`, so ordering must not mutate it.
- `web/src/components/LiveTraceInspector.test.tsx` contains existing assumptions that message turns render chronologically. Those expectations must be revised to assert the new default and toggle behavior while retaining the existing empty state, failure cards, prompt, and reasoning coverage.
- `web/src/index.css` contains `.inspector-messages-section`, `.messages-heading`, and `.inspector-messages-list` styles to extend for the toolbar and scroll container.

## Behavioral Contract

- Default display order is newest to oldest. “Latest First” and “Oldest First” are clear, accessible order states, with the toggle communicating the current state and its action.
- The ordering function is pure, stable, and non-mutating. The existing trace and message facts remain unchanged.
- Display position is separate from logical turn identity. For a source-ordered list of N messages, the oldest retains badge `Turn 1`, the next `Turn 2`, and newest `Turn N`, independent of current visual order. Related prompt context, parts, and failed-tool turn attribution continue to use that logical number.
- When the operator scrolls away from the latest edge beyond the agreed threshold, subsequent incoming messages must not force the visible content to jump. Provide a visible, keyboard-accessible way to return to the latest edge when new messages arrive. Returning to the latest edge clears the pending indicator.
- Navigation controls scroll the message list to its first or last displayed item according to the selected ordering, using the list's DOM scroll behavior.
- Empty traces remain safe and do not expose misleading turn-navigation or pending-message controls.
- All controls use semantic buttons, accessible names, and stable test IDs specified by the ticket.

## Architectural Boundaries

- Frontend ordering is pure presentation-domain logic under `web/src/domain/`; DOM refs, scroll listeners, viewport measurements, React state, and event handling remain in the component shell.
- Styling remains in the established web stylesheet. No inline backend or API contract changes.
- TypeScript remains strict, uses no `any`, and all new UI copy contains no emojis.

## Acceptance Boundaries

Acceptance is demonstrated by component and helper tests for newest-first default order, oldest-first toggle, stable logical turn numbering, immutable input ordering, scroll displacement protection on new messages, pending-new-turn affordance, and both navigation controls. Run the ticket's specified frontend checks: `npm run typecheck`, `npm run lint`, and `npm test`. No backend changes or backend tests are required for this presentation-only slice.
