h3. Title: AL-49: Reverse Event Chronology, Scroll-Lock, and Quick-Scroll Navigation in Live Trace Inspector

*TARGET*: AgenticLens (presentation / web / components / views / tests)

*CONTEXT*
 
In the Live Trace Inspector (`div#root > div.app-container > main.main-content > div.live-observability-view > div.live-split-layout > div.live-split-pane.right-inspector-pane > div.live-trace-inspector-container > section.inspector-messages-section > div.inspector-messages-list`), message turns and events are rendered in chronological order (turn 1 at top, latest turn at bottom).
In long agent sessions with 20 to 50+ turns, operators monitoring real-time activity must repeatedly scroll to the bottom. Furthermore, if reverse chronology is enabled (latest at top), incoming streaming events prepending to the list can violently displace the user's viewport downward if the user is reading an earlier turn, causing disorientation unless scroll locking is enforced.

*GOAL*
 
Order message turns in reverse chronological order (latest events at top) by default, provide an interactive order toggle allowing operators to switch back to chronological order (Latest First vs Oldest First), implement a pause-on-scroll lock when scrolled down so streaming events do not displace reading position, and add dedicated "Scroll to Top" and "Scroll to Bottom" navigation buttons.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Pure array reversal / sorting helper without mutating the underlying `LiveTraceView.messages` readonly array.
* Boundary Adapter: Zero backend changes; handled purely within inspector presentation.
* Presentation:
  - In `LiveTraceInspector.tsx`:
    - Display message turns in reverse chronological order (newest at top) by default.
    - Provide an order toggle button in the messages section header (`data-testid="turn-order-toggle-btn"`) allowing operators to toggle back to chronological order (Oldest First).
    - Implement pause-on-scroll lock: attach scroll listener to the container. If `scrollTop > 50px` (user has scrolled down to inspect older turns), lock the viewport position so new incoming turns prepend without shifting the visible view, or display a "New turns available [Jump to Top]" indicator.
    - Add quick-navigation buttons in the inspector controls: "Scroll to Top" (`data-testid="scroll-to-top-btn"`) and "Scroll to Bottom" (`data-testid="scroll-to-bottom-btn"`).
    - Maintain consistent logical turn badges (`Turn 1` through `Turn N`) regardless of display order.
* Quality: Unit tests in `LiveTraceInspector.test.tsx` verifying:
  - Message turns render with the newest turn first by default.
  - Clicking the order toggle switches back to chronological order (oldest first).
  - Pause-on-scroll lock prevents viewport displacement when scrolled down.
  - Turn badges maintain correct logical numbering.
  - "Scroll to Top" and "Scroll to Bottom" buttons invoke DOM scroll handlers.

*ACCEPTANCE CRITERIA*
 
* The message turns list in `LiveTraceInspector` renders the latest turn at the top by default.
* An accessible toggle button allows operators to switch between "Latest First" and "Oldest First" ordering.
* When scrolled down into earlier turns, incoming streaming turns do not violently shift or displace the user's viewport (pause-on-scroll lock).
* Turn badges continue to display correct logical turn numbering (`Turn 1` through `Turn N`).
* The header/toolbar provides accessible buttons to scroll directly to the top and bottom of the messages list.
* `npm run typecheck`, `npm run lint`, and `npm test` pass with zero errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
