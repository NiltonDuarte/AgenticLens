# FIX-12 iteration 02: Couple anomaly visibility to waterfall expansion

Role: Design/planner only. Ticket: FIX-12. Classification: `fix`.

Status: Proposed delta for subsequent implementation approval. This assignment produces only this plan. Application code and tests are unchanged; no tests, builds, commits, telemetry generation, or PR operations are performed.

## Baseline and grounding

The frozen baseline is [CONTRACT.md](../../CONTRACT.md), especially required behaviors 6 and 7: empty traces retain aggregate anomaly inspection, and existing collapse defaults and the initial errors-and-anomalies filter remain intact. [Iteration 01 plan](../01/PLAN.md) and its [approved round 2 review](../01/REVIEW.md) establish independent optional enrichment, selection rebinding, and request isolation. This delta changes how existing anomaly access is disclosed; it does not revise the contract or reopen the approved data pipeline.

- `web/src/views/TicketDeepDive.tsx`, Row 6, mounts `TraceWaterfall` and then renders `workspace-anomalies` as a sibling when `events.length === 0 && traceAnomalies.length > 0`. That sibling cannot follow the child's expansion state. Its entries already use `formatAnomaly`, severity badges, click and Enter/Space handlers, and the accessible name `Inspect anomaly: ...`.
- `web/src/components/TraceWaterfall.tsx`, `TraceWaterfallProps` and `TraceWaterfall`, owns `isWaterfallExpanded`, initialized from `defaultExpanded`, which defaults to false. Nonempty traces expose `waterfall-expand-btn` and `waterfall-collapse-btn` with `aria-expanded`. The empty-events early return bypasses both controls and always shows the chronological empty notice.
- The same component separately owns filters, local event selection, and payload expansion. Its initial `filterMode` is `errors_anomalies`; a populated trace filtered to zero rows has the distinct message `No events match the selected filters.`
- `TicketDeepDive`, `handleSelectEvent`, `selection`, and `isDrawerOpen` own inspection. Aggregate entries select an anomaly with a null event. `loadData` and the request-generation trace/drawer keys reset the current request's UI; enrichment updates data and rebinds selections without changing those keys.
- `web/src/components/TraceWaterfall.test.tsx` covers immediate empty-state rendering, default collapse, explicit `defaultExpanded`, expansion/collapse, filters, event keyboard selection, and payload inspection. `web/src/views/TicketDeepDive.test.tsx` covers summary-only omitted/empty event collections, direct aggregate inspection, updated summary context during deferred enrichment, populated trace inspection, and obsolete requests. Empty-trace assertions currently assume no expansion step and need deliberate adjustment.

## Desired visibility and defaults

1. Use one expansion control for the execution waterfall card and all of its trace details. Collapsed means the card summary remains visible while the timeline, filters, empty-state body, and eligible anomaly list are absent from the interactive content. Expanding reveals the applicable body; collapsing hides it again. The anomaly list has no separate toggle.
2. Apply the existing default, collapsed, to empty traces as well as populated traces. The collapsed empty summary retains the execution title, truthful source label, `0 Events Recorded`, and the actual anomaly count when anomalies exist, so users can discover that inspection is available. `defaultExpanded: true` continues to initialize an expanded standalone component, including with zero events.
3. For an expanded empty trace, show `Zero chronological events recorded for this execution trace.` and the existing `Detected Workspace Anomalies:` list inside the same card when anomalies exist. With no anomalies, show the empty notice without a list or placeholder anomaly entry. Empty traces do not gain event filters or fabricated timeline rows.
4. Preserve the list's eligibility: it is the fallback for a genuinely empty raw `events` collection. A populated trace retains its existing event/anomaly badges and inspection controls. Filtering that trace to zero visible rows shows the filter-empty message and does not create the aggregate fallback list.
5. Keep the schema notice, session threads, metrics, and artifacts outside this disclosure. Hydration availability, source changes, anomaly count, and schema metadata do not force the card open or prevent expansion.

## Interaction and state ownership

### TraceWaterfall owns the disclosure

- Retain `isWaterfallExpanded` as the single source of expansion state and preserve `defaultExpanded` as an initialization option. Do not add a second expansion flag in `TicketDeepDive`, synchronize flags with effects, or introduce persisted expansion preferences.
- Add a narrowly named, optional readonly presentation slot, `emptyTraceContent`, to `TraceWaterfallProps`, using React's node type and the repository's strict optional-prop conventions. Render this slot only in the expanded empty-trace body. It carries the existing anomaly markup without moving selection or drawer state into the waterfall.
- Replace the empty-events early return with an empty-body branch under the shared disclosure shell. Compute summary counts independently of whether event rows exist. Preserve current nonempty source copy, counts, filter semantics, timeline, payload details, and event callbacks.
- Keep one native toggle button mounted across expansion changes, updating its existing Expand/Collapse label, test identifier, and `aria-expanded`. Associate it through `aria-controls` with a stable, instance-specific body identifier. Keep that body wrapper present and hidden while collapsed, with interactive children rendered only when expanded. This preserves a valid control target and keyboard focus without leaving hidden controls in the tab order.
- A toggle changes visibility only. It does not fetch data, select an event or anomaly, clear filters or payload expansion, or open/close the inspection drawer. Keep the component instance mounted when toggling so existing filter, selection, and payload state survive reopening.

### TicketDeepDive owns inspection and content

- Supply the current `workspace-anomalies` markup through `emptyTraceContent`, retaining its existing raw-events/anomaly guard, formatting, severity badges, and accessible names. Remove its sibling placement so there is exactly one copy inside the waterfall's body.
- Preserve aggregate activation: clicking or pressing Enter/Space on an entry sets the selected event to null, selects that anomaly, and opens `AnomalyDrawer`. Keep the existing Space default prevention. Activating an anomaly must not toggle the card.
- Preserve `handleSelectEvent` behavior for populated traces: correlated anomalies and plain failures open the appropriate inspector; unflagged rows retain their current selection behavior. The drawer remains owned and rendered by the view, with its existing Close and Escape behavior. Closing it leaves card expansion unchanged.
- Preserve current request-generation keys. Ticket changes and explicit reloads remount the trace with its default collapsed state and clear prior inspection as they do now. Optional enrichment within the same request preserves the user's current expansion choice, including when an empty trace becomes populated. Content changes must not manufacture a new trace key.
- Retain iteration 01 selection rebinding. An inspected event-free summary can receive updated context while its drawer stays open; if events arrive, the fallback list disappears and the expanded card shows the timeline without forcing the drawer closed or inventing an associated event.

## Focused implementation scope for subsequent roles

Coder ownership is limited to `web/src/components/TraceWaterfall.tsx` and `web/src/views/TicketDeepDive.tsx`. Reuse the existing waterfall and `legacy-anomalies-*` styling; a small adjustment to `web/src/index.css` is permitted only if nesting the existing list requires spacing changes. Keep strict typing, presentation-only props, and zero emojis. This slice requires no changes to domain logic, network behavior, backend code, ledger schemas, dependencies, or anomaly matching.

Tester ownership is limited to the two existing component/view test files below. Adjust existing expectations that deliberately change under the new empty-trace default, and add coverage of the coupled behavior. Avoid duplicating the approved event reconciliation suite.

## Planned acceptance tests

### TraceWaterfall.test.tsx

1. Empty input plus supplied anomaly content starts collapsed with zero events, the correct source and anomaly count, and an Expand button. The empty notice and anomaly controls are absent until expansion. Expand, collapse, and re-expand reveal/hide both together, with no duplicate list. Also cover empty input without anomaly content.
2. Empty input with `defaultExpanded: true` immediately shows the empty notice and supplied content and provides a working Collapse button. Update `renders empty message when events array is empty` to expand explicitly or opt into this default rather than dropping the empty-state assertion.
3. A populated trace preserves its collapsed summary, initial issue filter, event counts, and source labels. Expanding reveals its filters and applicable rows; collapsing removes them. An empty-trace slot is not rendered for populated input, including when filters leave zero matches.
4. Selecting a nondefault filter and expanding a payload, then collapsing/reopening, preserves those choices and event selection. Toggle activation does not invoke event or anomaly actions.
5. Verify native button semantics, Enter/Space activation, accurate `aria-expanded`, a stable valid `aria-controls` target, and focus retained on the toggle through both transitions. Collapsed content exposes no anomaly entry, filter, or payload control to keyboard navigation or accessibility queries. Preserve existing Enter/Space event-selection coverage.

### TicketDeepDive.test.tsx

1. For both omitted and explicit empty session event collections with anomalies, wait for the collapsed zero-event summary. Confirm no list is exposed initially; expand and assert `workspace-anomalies` is contained within `trace-waterfall-card`. Collapse and re-expand and assert the list disappears and returns exactly once. Retain the no-false-legacy assertion.
2. From the expanded empty trace, activate an aggregate anomaly by click, Enter, and Space. Assert the drawer contains the expected rule/context without a fabricated event. Close it and reopen an anomaly; preserve Escape dismissal and expanded-card state. Collapsing alone must not select an anomaly or open a drawer.
3. Update the existing direct aggregate-inspection and summary-context-rebinding tests to expand before activation. Resolve the deferred event-free enrichment and verify the same inspector receives updated severity/context while expansion remains unchanged.
4. For an empty trace without anomalies, verify default collapse and an expandable truthful empty body, with no aggregate list. Retain thread/artifact availability independently of card expansion; adapt the existing summary-only tests to expand before asserting the empty body.
5. For populated committed and hydrated traces, exercise expand/collapse and existing event-level anomaly/failure inspection. Confirm no aggregate fallback list appears. Retain the initial issue filter, All Events behavior, and distinction between a genuinely empty trace and zero filter matches.
6. Using the existing deferred-response helper, start with an empty trace and anomalies, expand it, then resolve nonempty enrichment. Assert the same expansion choice survives, the fallback list is replaced by the timeline, and a selected summary remains correctly bound. Also resolve enrichment while collapsed and confirm it stays collapsed.
7. Extend the existing ticket-change/reload assertions only as needed to prove a fresh trace starts collapsed, old anomaly content/drawer state does not leak, and obsolete responses cannot change the current card's contents or expansion.

## Verification and handoff

This artifact is a proposed implementation plan, not evidence of passing tests. Grounding consisted of reading the approved artifacts and current source/test assertions. No tests or application commands were run for this assignment.

After implementation approval, run the two focused frontend suites and the existing frontend typecheck/lint checks; subsequent repository verification and embedded-asset delivery remain with the development workflow. The reviewer should check single state ownership, actual DOM containment, empty/nonempty defaults, keyboard/focus behavior, and retention of iteration 01 inspection and request-isolation guarantees.

The frozen contract, iteration 01 artifacts, summary, PR draft, telemetry, application code, and tests remain unchanged by this planner-only delta.
