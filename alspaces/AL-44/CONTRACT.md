# AL-44 Technical Contract

## Objective
Render trace event and message times as compact client-local `HH:mm:ss.SSS` values, preserve complete localized date and UTC offset in tooltips, and show the browser timezone in the live inspector header.

## Boundaries
- Implement pure browser date formatting and timezone resolution in `web/src/domain/formatters.ts`.
- Apply formatting to Live Trace Inspector, Trace Waterfall, and Ticket Deep Dive.
- Keep API and network contracts unchanged.
- Invalid values fall back to their original string or an empty string.
- Maintain strict TypeScript and zero `any` use.

## Acceptance
Formatter and component tests cover localized rendering, millisecond padding, invalid values, tooltip context, and timezone badge. Web typecheck, lint, and tests pass.
