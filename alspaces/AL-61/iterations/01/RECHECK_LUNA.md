# AL-61 Affected Re-review — Historical Timestamp Calendar Validation

## Outcome

**APPROVED** for the assigned narrow correction on the supplied candidate diff, SHA256 `f0c01c50091bd020205c68891e240e314df32b884f91ac7c3324737921839c61`, against baseline `122e983f8c365a2aa248b1b6d492252555e922ae`. No blocking findings (severity 0).

Review identity: reviewer requested as `gpt-6-luna` / `medium`; actual runtime model and effort are not independently exposed in this context. No run ID was supplied. This verdict covers the calendar-validation correction and AL-61 historical timestamp display behavior only; it does not certify the whole repository or the final integrated revision.

## Finding disposition

The prior SOL-1 finding (severity 5/10), that impossible dates such as `2026-02-30T10:15:30Z` normalize into a fabricated display date, is corrected in the candidate. `getHistoricalTimestampDateTime` validates month/day combinations with Gregorian leap-year rules before returning the original instant. `formatLocalDateTime` and `formatHistoricalDateTimeTooltip` share that guard, and all four consumer paths use the helper for `dateTime`. Therefore an impossible source date remains raw in visible text and tooltip and is omitted from the machine-readable attribute.

## Investigation and evidence

- Read the accepted behavior and boundaries in `CONTRACT.md` and `iterations/01/PLAN.md`. Inspected the corrected formatter and all four display consumers in TicketDeepDive, TraceWaterfall, SkillsSubagentsPanel, and AnomalyDrawer, plus their changed tests.
- Checked the original normalization counterexample against the corrected test cases: `2026-02-30`, non-leap `2025-02-29`, and `2026-04-31` are preserved as invalid text by formatter tests; AnomalyDrawer tests also assert an impossible date remains raw and has neither a valid `datetime` nor normalized tooltip. This exercises the prior failure condition and affected component boundary.
- Valid date behavior is retained: tests cover `2024-02-29`, equivalent instants with `Z` and `-05:00`, non-UTC date rollover, spring DST conversion, and offset-specific tooltips. The implementation accepts RFC3339-style timezone offsets with `Date.parse` rejecting out-of-range values, while it rejects out-of-range clock fields. Go adapter source inspected in the relevant timestamp paths uses `time.RFC3339Nano` formatting/parsing; the frontend pattern accepts those fractional and offset forms. The frontend API validators preserve timestamp strings without rewriting them.
- The guard is presentation-only. It does not modify source strings, transport, persistence, ordering, or live timestamp formatting. No additional introduced regression was found in the inspected callers or unchanged API path.

## Verification and limits

Executed the four affected frontend suites: formatter tests (73 passed), AnomalyDrawer tests (20 passed), TraceWaterfall tests (84 passed), and TicketDeepDive tests (67 passed), total 244 passed. `git diff --check` passed. Parent reports corrected typecheck and lint passed; reused rather than rerun. Parent owns the integrated `make` gate after integration.

No browser viewport geometry or assistive-technology validation was performed, so wrapping behavior at narrow widths remains source/CSS-inspected rather than browser-proven. No full-repository review was performed. These limits do not prevent approval of the specific calendar-validation correction.
