# AL-82: Filter and compare metrics by tags

Completed at implementation revision `57b1a45c`; [delivery evidence](../../AL-82/SUMMARY.md) records checks and review. The user's approval, “Approved. Be sure the tags are in the ticket scope/level”, supersedes the original PO-21 session-level ownership and Mixed contributor proposal. The [approved contract](../../AL-82/CONTRACT.md) and [iteration 02 delta](../../AL-82/iterations/02/PLAN.md) own current acceptance.

## Goal

Deliver ticket-level tags from shared/local configuration and CLI through durable metric records and API responses to dashboard filters and comparison identity. Preserve user identity, portable history and numerical metrics when no tag filter is selected.

## Acceptance criteria

* Shared and private local YAML accept a top-level string mapping. Repeatable `generate --tag KEY=VALUE` overrides local, which overrides shared, per key; unrelated keys survive. Tags-only local files work. CLI repeats use the last value, splitting at the first equals sign. Trim outer Unicode whitespace and preserve internal text/case. Reject blank/non-string entries and duplicate normalized YAML keys before publication. Document examples and that effective local tags become shared ledger data.
* Tags are one map on the ticket ledger, independent of SessionFact.User and harness. Successful manual/automatic generation merges effective keys over saved ticket tags; absent keys never erase metadata. Failed handled publication restores the prior accepted set. Interrupted or obstructed recovery remains explicitly incomplete rather than claiming success.
* Save/reload, source-unavailable reading, backfill, supported explicit migration and cross-machine Git sharing preserve tags and session facts. Reader configuration does not relabel history. Untagged schema-6 ledgers remain valid without forced migration; older schemas retain the existing explicit migration gate.
* Dashboard filters start unset; OR values within a key and AND keys on one ticket. Compose with existing filters and include Missing distinctly from literal values. Select whole tickets and complete totals, never owner allocations. Keep choices before tag filtering, reconcile fresh server values, and retain usable reset/empty-result controls.
* Comparison Color by tag defaults off, retaining completion coloring. Selecting a key uses stable identities across unchanged-data filtering/sorting, with legend and textual/accessibility identification. Preserve one point per ticket, coordinates, completion status, unavailable-metric and manual exclusions. Missing remains neutral and visible; literal Missing and Mixed are ordinary values, with no contributor Mixed state.
* Tests prove precedence/config round trips/validation; root ticket publication through both harnesses and the actual worker; unavailable and cross-developer history; absent-input and repeated refresh/backfill preservation; supported migration; no-filter compatibility; complete multi-session ticket totals; Missing/literal distinction, refreshed-catalog recovery and accessible stable comparison identities.

Shared product history: [PO-21](../../PO-21/SPEC.md). Related AL-70 configuration preservation and AL-72 UI correctness were coordination references, not prerequisites. [Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. No downstream ticket depends on AL-82.
