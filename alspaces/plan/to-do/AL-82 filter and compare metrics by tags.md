h3. Title: AL-82: Filter and compare metrics by tags

*GOAL*

Deliver tags from shared/local configuration and CLI through durable metric records and API responses to dashboard filters and comparison point identity. Preserve existing user identity, portable history and numerical metrics when no tag filter is selected.

*CONTEXT*

Users need arbitrary key-value cohort metadata in their metrics. A single user attribution field cannot express experiments, teams and environments. Existing comparison points distinguish completion status. Historical untagged metrics must stay valid and visible.

*ACCEPTANCE CRITERIA*

* Shared .agenticlens-config.yaml and private .agenticlens-config.local.yaml accept a top-level tags string mapping. CLI repeatable --tag KEY=VALUE overrides local, which overrides shared, per key; unrelated lower-layer keys survive. A tags-only local file works. CLI repeats use the last value, splitting at the first equals sign. Invalid/blank keys or values, non-string YAML values and duplicate normalized YAML keys fail before any ledger publication. Trim outer whitespace, preserve internal text and case. Document examples and that effective local tags become shared ledger data.
* Tags are additional metadata, independent of SessionFact.User. Fresh root/child owners for Codex and OpenCode receive effective tags, including automatic publication. Existing owner keys remain unless an effective key replaces them; no tag deletion syntax is included. Missing tag inputs do not erase recorded metadata.
* Preserve tags and facts for every owner outside a successful complete local refresh. Save/reload, source-unavailable reading, backfill, supported migration and cross-machine Git sharing preserve tags. Do not use current reader configuration to relabel historical metrics. Untagged/legacy ledgers remain valid without a forced migration unless canonical schema policy requires an explicit compatible transition.
* Dashboard tag filters start unset; values within a key are OR, keys are AND, and existing filters compose. Offer Missing distinctly from literal values. Reset restores tagged and untagged records. Build choices before tag filtering; show an actionable empty result. A conjunction must match one owner, not unrelated owners in the same ticket. Recompute session-derived totals from matching owners; ticket-only metrics remain clearly ticket-scoped without invented allocations.
* Comparison Color by tag starts off, preserving completion coloring. Selecting a key uses stable value identities with legend and textual/accessibility identification, without filtering or changing point count, identity or coordinates. Keep completion status available. Missing-key points remain visible with a neutral identity. A ticket aggregating multiple values or tagged/untagged contributors is Mixed with a tooltip listing its membership; preserve one point per ticket. Missing/Mixed categories cannot collide with literal tag values. Existing unavailable-metric exclusions and explicit user exclusions remain supported.
* Tests prove full three-layer precedence and retained keys; config round trips and tags-only local files; root/child and capture publication; persisted tags with unavailable sources; two developers and harnesses sharing history; absent-input refresh/backfill preservation and idempotence; no-filter compatibility; owner-aware filter totals; Missing/Mixed and accessible stable comparison identities.

*Status*: Ready
*Blocked by*: None
*Specification*: [Metric tags specification](../../PO-21/SPEC.md)
*Work type*: feat
*Priority*: P2

*IMPLEMENTATION NOTES*

*Related*: AL-70 configuration preservation; AL-72 UI correctness. Coordinate shared paths; neither is a prerequisite.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
