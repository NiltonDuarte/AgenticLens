h3. Title: AL-44: Global Localized Millisecond Timestamp Formatting and Timezone Header

*TARGET*: AgenticLens (presentation / web / formatters / components / tests)

*CONTEXT*
 
Across AgenticLens (Live Observability inspector, Trace Waterfall timeline, and Ticket Deep Dive autopsy views), event and message timestamps currently render full raw ISO-8601 UTC strings (e.g. `2026-10-22T14:32:10.123Z`).
In `LiveTraceInspector.tsx` (`span.turn-timestamp`), `TraceWaterfall.tsx`, and `TicketDeepDive.tsx`, this takes excessive horizontal space (24+ characters), causes premature wrapping of metadata chips, and forces operators to mentally compute timezone offsets from UTC. In live monitoring, operators understand that sessions naturally roll over midnight, but need compact local time (`HH:mm:ss.SSS`) with explicit timezone indicators.

*GOAL*
 
Format event and turn timestamps globally across the application as compact, localized `HH:mm:ss.SSS` in the client's local timezone with full ISO datetime preserved in a hover tooltip, and display the active client timezone indicator in the Live Inspector header.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Pure date formatter functions in `web/src/domain/formatters.ts` (`formatLocalTimeWithMs(iso: string): string` and `getClientTimezoneBadge(): string`) handling valid dates, invalid date strings, millisecond zero-padding (`007` ms), and client timezone resolution via `Intl.DateTimeFormat().resolvedOptions().timeZone`.
* Boundary Adapter: Zero network changes; leverages standard browser localization APIs.
* Presentation:
  - Update `LiveMessageTurnItem` in `LiveTraceInspector.tsx` to render the compact localized timestamp with a hover `title` containing full localized ISO datetime (`YYYY-MM-DD HH:mm:ss.SSS [TZ]`). Display the detected timezone badge in the inspector header (`.inspector-title-group`).
  - Update `TraceWaterfall.tsx` timeline node timestamps to use `formatLocalTimeWithMs` with full date tooltip.
  - Update `TicketDeepDive.tsx` event timestamps and timeline markers to use localized formatting.
* Quality: Unit tests in `formatters.test.ts` validating timestamp formatting, edge cases, invalid date fallbacks, and timezone badge resolution. Component tests in `LiveTraceInspector.test.tsx` and `TraceWaterfall.test.tsx` asserting localized rendering.

*ACCEPTANCE CRITERIA*
 
* Message turn timestamps in `LiveTraceInspector` display strictly `HH:mm:ss.SSS` in local client time (e.g. `16:42:05.120`).
* Timestamps across Trace Waterfall and Ticket Deep Dive utilize localized `HH:mm:ss.SSS` formatting.
* Timestamp elements retain a `title` attribute containing the full localized datetime and timezone offset for complete date context.
* Inspector header contains a visible badge displaying the active client timezone (e.g. `TZ: Europe/Berlin` or `TZ: America/New_York (UTC-4)`).
* Gracefully falls back to the original string or empty indicator if an unparseable timestamp is encountered.
* `npm run typecheck`, `npm run lint`, and `npm test` pass with zero errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
