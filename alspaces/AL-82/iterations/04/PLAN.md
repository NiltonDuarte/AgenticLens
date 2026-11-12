# AL-82 iteration 04: Active tag filters and simpler clearing

## Steering and approval

The user requested a list of active filters with an X beside each item and suggested one Clear all button instead of the current Clear key / Clear tag filters buttons. This directly authorizes the refinement. Baseline is `e2255c0396dc90b4a390665d9baad963e6984e38`, initially clean; PR 141 is open. The [contract](../../CONTRACT.md) and accepted iteration 03 matching/focus guarantees remain authoritative.

Changed acceptance: show active ticket-tag clauses inside the existing popup, preserving the compact closed toolbar control. Each row/chip identifies one key and its selected values; Missing membership must remain distinct from a literal value named Missing. Its accessible X removes that row's key clause, including when another key is being edited. Replace both old clear buttons with one quiet Clear all action, shown when clauses exist. The existing All values checkbox can still clear the edited key. Removing filters never deletes stored ticket tags or resets unrelated page filters.

This supersedes only iteration 03's clearing presentation. Preserve OR values, AND keys, key-selection behavior, stale selections, whole-ticket totals, refreshed catalogs, comparison identities, four toolbar placements, bounded long-text layout and all established focus/Escape/outside-dismissal recovery. No new persistence, API or matching behavior is authorized.

## Implementation and ownership

The coder owns `web/src/components/TagFilters.tsx`, scoped tag-editor styles in `web/src/index.css`, and initial test adaptations. Render the controlled `filters` prop rather than keeping a second active-filter store. Use the clicked row's key for removal, not the current key selector. Keep full arbitrary labels accessible and use a restrained existing-style clear action. No new hooks, types, dependencies or global styling framework are needed.

After handoff the tester owns missing test changes. Parent owns integration, telemetry, retained artifacts and PR delivery. The checkout is shared; every owner preserves others' changes.

## Verification and review

Adapt `TagFilters.test.tsx`'s existing multi-key and stale-filter cases to prove readable active membership, nonedited-key removal preserving other clauses, literal Missing versus absence, last-row removal, single Clear all and keyboard focus after removing focused controls. Adapt existing Fleet/Comparison/Progression/Deep Dive clear-action queries without losing their totals, transport, refetch or mismatch recovery assertions. Reuse existing geometry and dismissal tests; add only missing oracles. Run typecheck/lint, unit-first affected DOM checks, canonical build, integrated make and standalone embedding verification, recording actual results and environment limits.

Use paired behavior and test-effectiveness reviews on gpt-6.1-sol/high and gpt-6-luna/medium, with separate private reports: behavior-sol.md, behavior-luna.md, tests-sol.md and tests-luna.md. This narrow presentation change reuses existing clause removal/reset behavior and introduces no state, ownership, dependency or layer boundary, so a separate architecture focus is inapplicable. Data integrity is inapplicable because stored facts, configuration, codecs, predicates, transport and aggregation remain unchanged. The parent must reassess coverage if implementation exceeds these limits. Freeze source/tests before discovery; retain prior evidence only with an unchanged-source rationale.

## Delivery outcome

All four behavior/test profile cells approved candidate `8d28749f56a8007014e4f329e470fdaea8db0567c540824a5914d22ade5e5676` without findings. Source was committed/integrated as `1655ba45`; origin/main needed no merge changes. Full canonical make in the isolated network, host binary checks and standalone current-asset serving passed. [Retained review](REVIEW_SUMMARY.json) and the [ticket summary](../../SUMMARY.md) record scope, actual checks and limits. Completed private review records were compacted after approval.

The user requested deletion of iteration test-result documents. The three files were removed and their essential checks, before/after outcomes, limits and raw-log locations consolidated in the ticket summary. Application/test sources and existing approvals are unchanged.
