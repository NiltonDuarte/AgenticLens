# AL-61 Independent Review — gpt-6-luna / medium

## Outcome

**APPROVED** for the assigned narrow scope on the frozen candidate diff (`/tmp/al61-review.diff`, SHA256 `fec2f0c8a850ec41504df7e4f8b96d01298d33736e69555643cc3f3f0825f24f`), based on baseline `122e983f8c365a2aa248b1b6d492252555e922ae`. This is a scoped approval for historical display behavior, data failure, architecture, and test effectiveness; it is not whole-repository approval.

No blocking findings (severity 0).

## Investigation and evidence

- Compared the frozen diff to the AL-61 contract and iteration plan. The four timestamp consumers are limited to TicketDeepDive Started/Completed/Updated, TraceWaterfall event labels, SkillsSubagentsPanel Loaded At, and AnomalyDrawer selected event Timestamp.
- Traced timestamp values through `web/src/api/validation.ts`: workspace `started_at`/`completed_at`/`updated_at`, event timestamps, and skill timestamps are checked as strings, with no parsing or rewrite. The presentation retains the source string in `dateTime` for parseable values; malformed values omit the machine-readable attribute. No backend, persistence, ordering, duration, or API behavior changed.
- `formatLocalDateTime` converts the supplied instant with `Intl.DateTimeFormat` in the viewer timezone, includes year/month/day and hour/minute/second, omits milliseconds, emits `N/A` for absent inputs, and returns invalid input verbatim. Existing `formatLocalTimeWithMs` is unchanged. The Started missing state still uses the existing `formatIsoDateTimeShort` fallback; Completed still uses `In Progress` and Updated uses `N/A` through the shared formatter.
- Tooltip calls remain in place at each changed consumer. Event labels expose parseable source instants through `dateTime`; the anomaly drawer previously showed raw timestamp text and now retains that source instant in `dateTime` while adding the localized tooltip.
- CSS changes are scoped to timestamp rows, skills table time elements, historical drawer values, and timeline timestamps. They enable wrapping and min-width shrinkage without altering table overflow or controls. The inspected skills table still has its existing row focus-visible treatment; timestamps themselves do not become interactive controls.

## Hypotheses assessed

1. **Local date rollover or DST could produce misleading dates/offsets.** The formatter uses the browser's local timezone; new formatter tests explicitly set `America/New_York` and assert a UTC-to-local date rollover, the spring DST jump, and tooltip offsets on both sides. This is appropriate direct evidence for the formatter contract.
2. **Malformed or missing source timestamps could fabricate a date or lose diagnostic evidence.** Source-level path and assertions show malformed text is returned as text and title fallback, while missing lifecycle values retain their explicit states. No evidence of date fabrication.
3. **A display change could mutate stored values or alter existing live timestamp precision.** API validators preserve strings; the diff changes only the historical consumers and leaves the shared milliseconds formatter intact. Existing precision test remains as a guard.
4. **The tests may pass without detecting omitted date/seconds or source-instant loss.** The formatter test asserts concrete local date/time values and millisecond omission; component tests cover label owners and source `dateTime`/tooltip behavior for several consumers. AnomalyDrawer and TraceWaterfall additionally check malformed fallback and omission of invalid `dateTime`. The tests are self-contained and do not depend on the live workspace.
5. **Long values may escape their layout or become unreachable.** Scoped wrapping/min-width rules support shrinking timestamp flex items and wrapping table/drawer/timeline values. No browser geometry evidence was supplied; DOM/CSS review cannot prove viewport rendering at narrow widths.

## Verification reused / limits

The coordinator supplied tester results: all 244 focused cases passed (177 formatter/drawer/trace and 67 TicketDeepDive), and typecheck plus lint passed. A standalone binary HTTP probe verified bundled JS/CSS against `web/dist` using a disposable empty workspace; canonical asset trees match. I did not rerun these checks. The full `make` integration gate is parent-owned and outstanding. No browser geometry evidence was supplied; DOM/CSS review cannot prove viewport rendering at narrow widths. Record the `make` result and geometry limitation in final consolidation.

## Architecture and test effectiveness

No FCIS boundary change, speculative abstraction, or domain impurity was introduced. The formatter remains a small pure presentation helper and is reused at each historical label. Tests add relevant assertions to existing formatter/component recipients and retain the distinct live precision guard. I found no assertion deletion or unrelated test expansion. The test evidence is stronger for semantic formatting than for rendered narrow-screen geometry.
