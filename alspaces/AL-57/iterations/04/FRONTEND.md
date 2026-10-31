# AL-57 iteration 04 frontend implementation

Baseline: 52234b3c099fe1a6c1b2e1dd3d62c1db08ad7242. Parent documentation edits preserved. Scope: frontend presentation and filtered comparison benchmark only; no backend/schema changes or commits.

## Changes

Shared CostKPICard wraps the existing KPICard and InfoTooltip. Primary text is currency or a dash. Secondary text carries the explicit known-subtotal/partial, unavailable or empty state with existing view context. Keyboard-accessible tooltip contains the full invocation coverage and explains complete versus partial cost. Deep dive, fleet overview and progression pass their actual invocation counts for legacy numeric fallbacks. No CSS shrinking, clipping or global typography changes were made; existing secondary text wraps naturally. Existing tables, CLI, charts and JSON contract are unchanged.

Average Spend / Ticket now sums and counts complete finite costs in filteredRows only, including zero and excluding partial/unavailable cost or known subtotals. The visible eligible count explains its narrower denominator; all other benchmarks, rows and scatter populations still use the full current filtered dataset. Current search/type/prefix/skill/refetched data flows determine the result; no new time-window control was introduced.

## Verification

- npm run typecheck: passed.
- npm run lint: passed after removing unnecessary test regex escapes.
- Focused Vitest: Components, TicketDeepDive, FleetOverview, MetricProgression and TicketComparisons suites; 5 files, 146 tests passed, no skips.
- npm run build: passed on final production copy, including corrected tooltip wording about current fetched data.
- diff -qr web/dist backend/internal/adapters/web/dist: passed; generated embed assets match and remain ignored.
- git diff --check: passed.

Regression tests extend existing suites: short primary versus detailed secondary/tooltip with keyboard coverage access, complete zero and unknown/empty states; actual partial cost-card wiring in all three views; eligible mixed-cost average with search/reset and unknown/empty recovery. Existing literal-skill/type/prefix interaction case was left unchanged for the tester to extend; tester now owns all frontend test paths and will add missing composition/refetch guards.

Full integrated make, standalone binary asset validation and final review remain parent-owned. No preview/browser verification ran; DOM/accessibility tests establish content and interactions, not measured browser geometry.
