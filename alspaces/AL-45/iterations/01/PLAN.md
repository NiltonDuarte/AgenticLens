# AL-45 Iteration 01 Plan

## Objective

Deliver the approved frontend-only middle truncation and responsive containment slice described in [the contract](../../CONTRACT.md).

## Grounded implementation targets

- `web/src/domain/formatters.ts` contains existing pure formatting utilities; add the typed `truncateMiddle` utility there.
- `web/src/domain/formatters.test.ts` is the unit-test target for formatter behavior.
- `web/src/components/LiveTraceInspector.tsx` renders the message ID and part ID in `.debug-id-token.mono`. Its message and part prompt handlers currently pass raw IDs into prompt builders; retain that data flow.
- `web/src/components/LiveTraceInspector.test.tsx` is the component-test target for DOM presentation and copy behavior.
- `web/src/index.css` defines `.debug-id-token-group` and `.debug-id-token.mono`; update those rules and the `.message-turn-meta` rule to allow flex children to shrink and clip within the available width.

## Work sequence

1. Implement `truncateMiddle` with defaults of 8 leading and 6 trailing characters, a `...` separator, and unchanged output when the input is no longer than the retained character count. Cover defaults, custom boundaries, exact cutoff, shorter strings, empty input, and zero/invalid boundary behavior consistent with a safe pure utility.
2. Apply the formatter to visible message and part ID labels. Set each label's `title` to its raw identifier. Leave prompt generation and clipboard input sourced from the original message/part records.
3. Update the metadata and debug-ID CSS flex constraints so labels can shrink without pushing headers or cards wider; preserve visibility and alignment of role, timestamp, and copy controls.
4. Add component coverage for truncated visible message and part IDs, full-ID titles, and copy-prompt output containing the full identifier.
5. Run frontend tests and quality checks: `npm run typecheck`, `npm run lint`, `npm test`; then run the applicable repository `make` quality gates for delivery.
6. Review the diff against the contract, verify no backend or schema changes, and record the implementation/test outcome in the iteration review and ticket summary.

## Acceptance checks

- The canonical long-ID display retains 8 leading and 6 trailing characters around `...`; IDs at or below the cutoff render unchanged.
- Both message and part token elements expose the original ID in `title`.
- Copy-prompt generation uses full raw identifiers.
- CSS constraints prevent horizontal overflow down to 320px while keeping adjacent controls aligned.
- All specified frontend checks pass, with strict TypeScript and no `any` additions.
