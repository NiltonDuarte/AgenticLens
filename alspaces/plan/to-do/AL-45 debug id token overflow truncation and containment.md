h3. Title: AL-45: Debug ID Token Middle Truncation and Responsive Layout Containment

*TARGET*: AgenticLens (presentation / web / css / components / tests)

*CONTEXT*
 
In the Live Trace Inspector (`div#root > div.app-container > main.main-content > div.live-observability-view > div.live-split-layout > div.live-split-pane.right-inspector-pane > div.live-trace-inspector-container > section.inspector-messages-section > div.inspector-messages-list > div.trace-message-turn.role-assistant > div.message-turn-header > div.message-turn-meta > div.debug-id-token-group > span.debug-id-token.mono`), long UUIDs and identifier tokens span out of their container box.
Monospace tokens expand `.message-turn-meta` beyond horizontal limits on narrowed split panes. Furthermore, standard CSS tail ellipsis (`ses_0192a83e...`) destroys ending entropy where unique subagent session IDs differ. Branch `fix/FIX-3-centralize-live-anomalies-debug-identifiers` introduced these debug ID tokens and will be merged soon.

*GOAL*
 
Implement middle truncation (`ses_0192...304a`) for visible debug ID tokens and enforce flexbox container containment (`min-width: 0`, `max-width`), ensuring distinct subagent sessions are identifiable at a glance while the full uncut identifier remains accessible via tooltip and is copied intact by the prompt copy button.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Pure string middle-truncation utility `truncateMiddle(str: string, maxLeading?: number, maxTrailing?: number): string` in `web/src/domain/formatters.ts` preserving 8 leading and 6 trailing characters (e.g. `ses_0192...304a`), handling strings shorter than the cutoff without modification.
* Boundary Adapter: Zero backend/database changes; purely frontend presentation and formatting.
* Presentation:
  - In `LiveTraceInspector.tsx`: Apply `truncateMiddle` to message ID and part ID rendered inside `span.debug-id-token.mono`.
  - In `web/src/index.css`: Constrain `.debug-id-token-group`, `.debug-id-token.mono`, and `.message-turn-meta` with `max-width`, `min-width: 0`, `overflow: hidden`, and `white-space: nowrap`.
  - In copy handler: Ensure `copyTextToClipboard` copies the raw, uncut identifier.
* Quality: Unit tests in `formatters.test.ts` verifying `truncateMiddle` boundaries, edge cases, and short strings. Component tests in `LiveTraceInspector.test.tsx` verifying middle-truncated DOM text, full ID in `title` tooltip attribute, and clipboard copy action preserving the full identifier.

*ACCEPTANCE CRITERIA*
 
* Monospace debug identifier tokens do not overflow the horizontal boundaries of `.message-turn-meta` or the message turn card at any split-pane width down to 320px.
* Identifiers are middle-truncated (showing first 8 characters, ellipsis `...`, and last 6 characters) so unique trailing entropy is preserved.
* The element's `title` attribute contains the full, uncut identifier string.
* Clicking the adjacent "Copy" / "Copy Prompt" button copies the original full identifier without truncation.
* Flexbox layout allows adjacent role pills, timestamp, and action buttons to remain cleanly aligned.
* `npm run typecheck`, `npm run lint`, and `npm test` pass cleanly.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
