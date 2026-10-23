h3. Title: AL-46: Collapsible Failed Tool Calls Section with Prominent Failure Badge and Default-Collapsed State

*TARGET*: AgenticLens (presentation / web / components / tests)

*CONTEXT*
 
In the Live Trace Inspector (`div#root > div.app-container > main.main-content > div.live-observability-view > div.live-split-layout > div.live-split-pane.right-inspector-pane > div.live-trace-inspector-container > section.inspector-failed-tools-section > h4.failed-tools-heading`), when an agent session encounters failed tool executions, all `FailedToolAlertCard` instances are rendered fully expanded by default.
In sessions with 3 to 10+ consecutive tool retry errors, the error section takes up hundreds of vertical pixels, pushing the actual execution message stream (`section.inspector-messages-section`) completely out of view. However, simply collapsing the section risks hiding critical failure states from operators.

*GOAL*
 
Make the Failed Tool Calls section collapsible and collapsed by default, while rendering a prominent, high-visibility error badge and failure count on the collapsed header summary bar along with an expand/collapse toggle (`data-testid="failed-tools-toggle-btn"`), ensuring critical execution issues remain immediately obvious without blocking the message stream.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Zero domain model changes; utilizes existing `LiveTraceView.failed_tool_calls` and `LiveTraceView.recent_errors`.
* Boundary Adapter: Zero network changes; state is managed within the inspector component.
* Presentation:
  - In `LiveTraceInspector.tsx`: Local state `isFailedToolsExpanded` (default `false`).
  - In collapsed header: Render heading with a prominent failure pill/badge (e.g. `Failed Tool Calls: N` in error red/amber styling) alongside an expand button (`data-testid="failed-tools-toggle-btn"`).
  - In expanded view: Render the list of `FailedToolAlertCard` items.
  - In `web/src/index.css`: Add styles for collapsed header bar, prominent failure badge, and toggle button.
* Quality: Unit tests in `LiveTraceInspector.test.tsx` verifying:
  - When failed tools exist, section is collapsed by default.
  - Prominent failure badge and exact error count are visible in the collapsed header.
  - Execution message stream is immediately visible below the collapsed header bar.
  - Clicking toggle expands and collapses the card list smoothly.
  - Keyboard accessibility (`Enter`/`Space`) on toggle button.

*ACCEPTANCE CRITERIA*
 
* When `hasFailedTools` is true, `section.inspector-failed-tools-section` is collapsed by default.
* The collapsed state renders a high-visibility badge displaying the exact failed tools count (e.g. `Failed Tool Calls: N`).
* The individual `FailedToolAlertCard` elements are hidden when collapsed and rendered when expanded.
* Clicking the toggle button expands and collapses the error cards list.
* If a session has zero failed tool calls, the section is omitted entirely.
* `npm run typecheck`, `npm run lint`, and `npm test` pass with zero errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
