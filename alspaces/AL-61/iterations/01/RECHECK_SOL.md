# AL-61 affected re-review

Verdict: APPROVED for the corrected historical-display boundary. No new finding candidates. This scoped approval does not replace the parent-owned integrated build, asset and make gates.

Identity: general reviewer; requested and acknowledged profile `gpt-6.1-sol` / `high`. Actual runtime model/effort metadata are not independently exposed to this reviewer. Phase: affected re-review; local report identity `RECHECK_SOL` (no separate run ID supplied). Base: `122e983f8c365a2aa248b1b6d492252555e922ae`. Candidate: `/tmp/al61-review-fixed.diff`, SHA256 `f0c01c50091bd020205c68891e240e314df32b884f91ac7c3324737921839c61`. Both that file and the full current `git diff --binary` matched this hash after probing.

## Finding disposition

SOL-1, original severity 5/10: FIXED. `getHistoricalTimestampDateTime` rejects impossible source calendar dates before formatting; visible historical values, historical tooltips and machine-readable datetime attributes use this same guard in TicketDeepDive, TraceWaterfall, SkillsSubagentsPanel and AnomalyDrawer. The original `2026-02-30T10:15:30Z` now remains raw in visible text and title and has no datetime attribute. The unchanged live formatter retains its prior behavior and precision.

The formatter regression asserts raw text/title for impossible dates, including non-leap February and a 31st day in a 30-day month. The drawer regression reaches its actual rendering path and asserts raw text/title and absence of datetime for February 30. Those expectations fail under the previous permissive Date parsing. The valid leap-date and explicit-offset assertions also guard against an overbroad rejection fix.

## Investigation and executed verification

- Read the contract, plan, reviewer procedure, initial Sol/Luna reports and corrected diff. Traced each historical caller to the shared guard and checked relevant unchanged timestamp producers and API validation. API timestamp fields remain strings and are not rewritten. OpenCode `formatEpochMilli` emits six fractional digits and `+00:00`; Codex preserves rollout RFC 3339 strings and synthesizes fallback starts with RFC3339Nano; generated ledger updates use RFC3339Nano. The stricter historical parser accepts these formats.
- Bundled the actual corrected `web/src/domain/formatters.ts` and its dependencies with esbuild into `/tmp/al61-sol-recheck-formatters.mjs`; executed Node assertions against those production exports. Twelve invalid calendar/clock/offset inputs retained source text in visible values and historical tooltips and returned undefined for datetime. Inputs included February 30, non-leap February 29, April 31, month/day zero, hour 24, minute/second 60 and invalid offsets.
- Six valid source formats/offsets produced the independently expected UTC clock: whole seconds with Z, six fractional digits with +00:00, nine fractional digits with Z, +05:30, -03:30 and +05:45. Valid leap years 2024, 2000 and 2400 were accepted; February 29 with a negative offset correctly rolled into March 1.
- Under America/New_York, concrete expected strings passed for prior-day local rollover, the spring DST jump, both autumn repeated-hour instants with distinct UTC offsets, and preserved live milliseconds. Empty/null/undefined values retained N/A and no datetime.
- Compared the existing first-pass production bundle with the corrected bundle: the former fabricated `03/02/2026, 05:15:30` for February 30, while the latter returned the exact source text. This independently establishes the before/after failure rather than relying only on suite status.
- `git diff --check` passed. No product or test files were modified by this reviewer.

## Reused evidence and limits

Coordinator-supplied corrected-candidate evidence: formatter suite 73 passed, drawer suite 20 passed, typecheck and lint passed; remaining corrected focused reruns were in progress at dispatch. Previous first-pass evidence included 244 focused cases, matching canonical asset trees and a standalone binary HTTP asset probe; it is not evidence of the final integrated correction. The parent owns the final focused-suite accounting, canonical asset rebuild, standalone delivery verification and integrated make gate.

No introduced regression was substantiated in the affected boundary. Browser geometry and assistive-technology behavior were not exercised. This is not a whole-repository audit or merge authorization.
