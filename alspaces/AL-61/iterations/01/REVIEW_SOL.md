# AL-61 independent narrow review

Verdict: CHANGES_REQUESTED. Scope: historical display behavior, data failure, architecture and test effectiveness. This verdict applies to the frozen first-pass candidate only, not later corrections or the whole repository.

Review identity: general reviewer, requested profile `gpt-6.1-sol` / `high`; actual runtime model/effort are not independently exposed to this reviewer. No separate run ID was supplied; local candidate ID is SOL-1. Base: `122e983f8c365a2aa248b1b6d492252555e922ae`. Raw diff: `/tmp/al61-review.diff`, verified SHA256 `fec2f0c8a850ec41504df7e4f8b96d01298d33736e69555643cc3f3f0825f24f`.

## Finding

SOL-1 — severity 5/10 — Impossible calendar dates are normalized into fabricated historical dates.

Locations in the frozen candidate: `web/src/domain/formatters.ts:59` (new historical formatter), `web/src/components/AnomalyDrawer.tsx:289` (new historical value and validity check), and equivalent `Date.parse` checks in the other label owners.

Trigger: an event timestamp such as `2026-02-30T10:15:30Z`. `new Date` and `Date.parse` accept this impossible date and normalize it into March 2. The new formatter therefore returns a plausible date rather than preserving the source text or displaying an unavailable state. The drawer previously showed this source timestamp verbatim, so the changed drawer introduces observable evidence loss. It also marks the impossible source as a valid machine-readable datetime. This violates the accepted invalid-value rule.

Reachability: unchanged `web/src/api/validation.ts:21` validates event timestamps only as strings; `isEventFact` retains them. Persisted `SessionEventFact.Timestamp` and exported `EventFact.Timestamp` are strings (`backend/internal/domain/models.go:45` and `:96`), and frontend `sessionEvents` preserves the source field. There is no intervening calendar validity guard on the rendering path. The old clock formatter has the same permissive parsing behavior, but the drawer's prior raw value does not; this report distinguishes that pre-existing behavior from the introduced display regression.

Independent reproduction: bundled the actual `src/domain/formatters.ts` and dependencies with esbuild into `/tmp/al61-sol-formatters.mjs`, imported it with Node v24.22.1, and set `TZ=America/New_York`. For that source string the production formatter returned `03/02/2026, 05:15:30`; its tooltip returned `03/02/2026, 05:15:30.000 [UTC-05:00]`; the production `Date.parse` validity expression returned true. In contrast, `not-a-date` remained raw. The current invalid-value tests cover only unparseable strings, so they do not detect this normalization path.

Required correction: reject impossible calendar timestamps before historical formatting and machine-readable datetime assignment, retaining raw invalid evidence or an explicit unavailable state. Preserve accepted valid source instants, offsets, local conversion, existing live precision and source strings. Add a coherent regression covering this normalization case through the formatter and drawer.

## Other hypotheses assessed

- Local date rollover and daylight-saving conversion: no additional defect found. The independent production-formatter probe converted `2026-01-01T02:03:04.987Z` to `12/31/2025, 21:03:04`, and the two autumn transition instants `2026-11-01T05:30:00.123Z` / `2026-11-01T06:30:00.456Z` both displayed `11/01/2026, 01:30:00` with distinct tooltip offsets `UTC-04:00` / `UTC-05:00`. Visible values omitted fractions; tooltips and unchanged live values retained them.
- Missing values: the formatter produces N/A; Started retains its existing fallback; missing Completed remains In Progress; Updated and recorded skill/event labels use an explicit N/A when rendered. The drawer retains its existing omission of an absent timestamp. No new fabricated completion or fallback instant was introduced.
- Source identity and unaffected consumers: inspected API validators/client decoding, domain event reconciliation/sorting, historical callers, `ThreadViewer`, and `LiveTraceInspector`. The diff changes presentation only; it does not rewrite stored/API strings, sorting, durations or raw thread diagnostics. The live inspector still calls unchanged `formatLocalTimeWithMs`.
- Layout and accessibility: inspected `.header-meta-row`, `.timestamp-group`, `.timestamp-item`, `.timeline-top-row`, `.timeline-timestamp`, drawer metadata grid and table containment. Existing flex wrapping and new overflow wrapping support the longer values. Skills remain inside a keyboard-focusable horizontal scroll region; existing buttons and navigation handlers are preserved. No substantiated new layout defect was found by source inspection. Browser geometry was not exercised.
- Architecture: the new helper has meaningful formatting/fallback responsibility and stays within the existing typed presentation/domain boundary. No new I/O, dependencies, state ownership or transport changes were introduced. No additional blocking architectural defect found.
- Test effectiveness: coherent existing suites now assert date visibility, seconds, absence of fractions, source datetime, tooltips, malformed raw values, and workspace lifecycle defaults. Controlled New York formatter expectations independently cover rollover and spring DST offsets. Existing interaction tests continue to cover navigation/collapse. The impossible-calendar case above is a concrete missing oracle; suite success does not refute it.

## Verification and limits

Executed: frozen-diff SHA256 verification; targeted source/dependency inspection; production-formatter bundle/probe described above; `git diff --check` (passed before corrections began). No product or test files were modified by this reviewer.

Reused coordinator/tester evidence for the unchanged candidate: all 244 focused cases passed (177 formatter/drawer/trace and 67 TicketDeepDive), typecheck and lint passed; canonical frontend build and embedded asset trees matched; disposable-workspace standalone HTTP probe matched bundled JS/CSS to `web/dist`. These were reported results, not independently rerun here. The parent owns the integrated `make` gate, final asset checks, correction verification and consolidated approval.

Limits: no real-browser geometry or assistive-technology validation; no full-repository audit; no independent verification of the final integrated revision. The coordinator acknowledged SOL-1 and began routing a correction after the original frozen-candidate investigation; later source changes require a separate affected re-review.
