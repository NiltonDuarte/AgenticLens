# AL-45 Contract: Debug ID Token Middle Truncation and Responsive Containment

## Problem

Long message and part identifiers in the Live Trace Inspector can expand the message header beyond its available width. Tail truncation hides the identifier suffix, which can contain the entropy needed to distinguish sessions and parts.

## Approved outcome

Display long debug identifiers with middle truncation so both their leading and trailing characters remain visible. Keep the complete identifier available through the token's `title` attribute, and preserve complete identifiers in the existing copy-prompt flow. Contain the related flex layouts so the message metadata remains within the message card at narrow split-pane widths, including 320px.

The pure utility is `truncateMiddle(str: string, maxLeading?: number, maxTrailing?: number): string` in `web/src/domain/formatters.ts`, with defaults of 8 leading and 6 trailing characters and `...` as the separator. Strings at or below the combined visible-character cutoff are returned unchanged.

## Scope and boundaries

- Presentation and frontend formatting only. No backend, database, API, or telemetry schema changes.
- `web/src/domain/formatters.ts`: pure string utility; no I/O, clock, or mutable state.
- `web/src/components/LiveTraceInspector.tsx`: use the utility for rendered message and part ID labels; expose the raw identifier as the token title. Existing prompt construction and copy behavior must continue to use the raw `message.id` and `part.id` values.
- `web/src/index.css`: contain `.message-turn-meta`, `.debug-id-token-group`, and `.debug-id-token.mono` with responsive flex sizing and overflow rules (`min-width: 0`, `max-width`, `overflow: hidden`, and `white-space: nowrap`) without disrupting adjacent role, timestamp, and action controls.
- `web/src/domain/formatters.test.ts`: utility boundary, edge, and short-input tests.
- `web/src/components/LiveTraceInspector.test.tsx`: rendered truncation, full title, and full identifier copy-prompt coverage.

## Acceptance boundaries

1. Long tokens show their first 8 and last 6 characters separated by `...`; short strings remain unchanged.
2. The tooltip/title contains the complete untruncated identifier.
3. Copying a prompt retains the complete original identifier, not the display label.
4. At split-pane widths down to 320px, debug tokens do not force `.message-turn-meta` or the message card beyond their horizontal bounds, and nearby controls remain aligned.
5. `npm run typecheck`, `npm run lint`, and `npm test` pass. The repository quality gates (`make check`, `make test`, or `make`) are used for delivery verification.

## Architectural invariants

Keep truncation as a deterministic pure function in the TypeScript domain formatter module and DOM/CSS behavior in the presentation layer. Preserve strict TypeScript typing, introduce no `any`, and add no emojis to code, UI, comments, or workspace artifacts.
