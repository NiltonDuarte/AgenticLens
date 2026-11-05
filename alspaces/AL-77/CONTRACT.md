# AL-77: Cross-ticket session overlap warning

Status: Approved baseline. Implementation approved on 2026-11-04.

## Problem and outcome

The same recorded session ID can belong to more than one ticket ledger. Show this overlap in the Fleet Overview, identify every affected ticket in the inspected project dataset, and let the user navigate to those tickets and decide how to correct attribution. Metrics may overlap; the platform cannot determine which ticket attribution was intended.

The source requirement is [AL-77](../plan/done/AL-77%20warn%20when%20a%20session%20appears%20in%20multiple%20tickets.md). This contract and the iteration 01 plan were approved for implementation.

## Frozen acceptance boundaries

1. Compare exact, nonblank stored session IDs across distinct nonblank recorded ticket IDs. Two or more distinct ticket IDs create a conflict. Repeated sessions, invocations, threads, or ledger projections for one ticket do not. Do not normalize session identity, require matching source labels, or infer missing IDs from other fields.
2. Each conflict lists the complete session ID, all affected ticket IDs, available titles and recorded source labels. Source labels are context only. Unknown context stays unknown. Duplicate ticket projections merge their contextual labels without multiplying ticket membership.
3. Inspect all successfully loaded ledgers from the unfiltered project workspace listing. Date, type, prefix, and table-search filters continue to govern existing metrics and tables, but do not narrow overlap detection. State that the warning covers loaded project ledgers independently of visible metric filters.
4. Skipped-workspace or unreadable-diagnostics warnings make the inspected scope explicitly partial. A failed overlap-data request is unavailable, never evidence of no overlap. No result claims knowledge of unloaded, invalid, or external records. Missing usable IDs are skipped without guessing, with a scope limitation when applicable.
5. Fleet Overview supplies a discoverable warning and expandable conflict details. Use native keyboard-accessible disclosure and named ticket controls that call the existing ticket-selection navigation. Without a navigation callback, show readable ticket labels rather than inert controls. Details explain potential overlapping metrics and the user's responsibility for attribution decisions. Long IDs remain readable without breaking page containment.
6. The overlap projection uses persisted facts without original harness sessions, capture databases, receipts, or hydration. Reads never rewrite ledgers. No deletion, reassignment, cross-ticket deduplication, aggregate-policy adjustment, or extraction/schema/storage redesign is permitted.
7. A successful reload replaces the previous overlap result, including clearing resolved conflicts. Provide an explicit way to reload the inspected dataset; protect against obsolete request completions and component unmounts. Loading/error states must not present stale conflicts as current. Optional overlap loading failure must not prevent supported metric data from rendering.
8. Preserve existing filters, empty-filter recovery, metrics, ticket selection, scan diagnostics, and single-binary distribution. Single-ticket and no-overlap inputs create no conflict warning. Partial/unavailable scope can still require explanatory status.

## Architecture and implementation scope

Use TypeScript pure read projection in web/src/domain, presentation in web/src/components, and FleetOverview loading/orchestration in web/src/views/FleetOverview.tsx. Reuse fetchWorkspaceLedgers from web/src/api/client.ts without filter arguments; the existing /api/workspaces response already contains ticket metadata, facts.sessions[].id, and source. No backend production changes or new API are currently needed. Do not fetch one detail/trace per ticket or introduce polling. Detection uses an indexed traversal of loaded session memberships rather than pairwise comparisons.

Follow [TECHNICAL_GUIDELINES.md](../../TECHNICAL_GUIDELINES.md) and [REGRESSION_PREVENTION.md](../../docs/REGRESSION_PREVENTION.md), particularly portable ledgers, accepted persisted baseline, unavailable versus empty data, and test proportionality. Preserve FCIS, readonly typed inputs, zero any/interface{}, zero emojis, and source-independent rendering. Tests use self-contained synthetic fixtures and temporary directories rather than live ticket history.

Backend adapters may receive focused tests for the existing read projection and absence of writes; these do not authorize changing extraction, validation, ledger formats, or aggregation policy. A newly discovered need to extend the API or alter these boundaries returns to planning and approval.

## Delivery acceptance

Verify the pure projection, actual FleetOverview consumer, accessible disclosure/navigation, full-project scope under filtering, partial/unavailable states, request ordering, corrected-record reload, and unchanged aggregate presentation. Execute required checks and integrated make after upstream integration. Build the frontend through npm run build, verify the mirrored embedded asset tree, and verify the built standalone binary serves the current assets. Follow current generated-file tracking policy; ignored build output stays ignored.

Implementation and adversarial tests belong to coder/tester. Independent review covers behavior, data/failure, architecture, test effectiveness, and frontend async/accessibility under both configured model profiles. Parent owns consolidated approval, integration, telemetry, commits, and PR publication. Approval here does not authorize merging.
