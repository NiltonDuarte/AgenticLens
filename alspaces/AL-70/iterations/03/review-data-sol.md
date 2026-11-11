# AL-70 iteration 03 data integrity review

APPROVED for the assigned data integrity scope. No substantiated correctness finding; candidate IDs: none.

Discovery run: a12_data_sol, reviewer-data. Requested profile: gpt-6.1-sol/high; dispatch acknowledgment belongs to the parent record. Runtime source metadata, usage and cost were not independently observed.

Reviewed baseline `54d58be40c74864d645c5168ba23820a21389c01` against frozen candidate fingerprint `974d1449fc440115b563067461d43b9b6381baf60e178ff45e6854f884b10e6a`. All 29 manifest entries matched their SHA-256 values at entry and exit. Product and tests remained read-only; unrelated shared-checkout changes were preserved.

The investigation traced OpenCode v1/v2 extraction, Codex rollout lifecycle/model buckets/finalization, live model resolution, domain model/thread/fleet aggregation, saved-ledger API reads, and model labels in Markdown and frontend consumers against the approved A12 contract.

Hypotheses and dispositions:

- **Empty key mistaken for absence:** refuted. Codex `assignTurnModel` and `turnNumber` now test map presence; every remaining `turnModels` read was traced. Unknown turns retain their original tool-event turn references, and finalization creates an empty-key aggregate for metadata-only rollouts.
- **Late context corrupts attribution:** refuted for the representation change. Only a present empty-model turn can be reassigned; observed `default` and `unknown` are retained. Existing semantics intentionally leave pre-context usage/completed duration in the original bucket while moving turn membership. The empty bucket is deleted only when turns, usage and active duration are all empty. Multiple-turn lookup and late-context counters pass.
- **Unknown rows disappear or collide with observed names:** refuted. OpenCode keeps metadata-backed zero-token rows and their iteration/turn counters; mixed missing/default/unknown identities remain distinct. V2's assistant `model.id` guard remains unchanged. Existing single-model versus multiple-model turn accounting is unchanged; splitting the former placeholder collision follows the approved grouping boundary.
- **History is relabeled or requires private source state:** refuted on the affected path. A standalone schema-6 fixture without a harness store survives repeated production workspace-detail reads with original `default`/`unknown` rows, tokens, cost, duration and counters; ledger bytes remain identical. Save/reload also preserves newly extracted empty model strings. No schema, migration, persistence authorization or refresh algorithm changed.
- **Human labels rewrite canonical grouping:** refuted. Domain maps, sort comparisons, React keys, color lookup and cost-summary lookup use raw keys; Unknown is applied during rendering only.

Executed five bounded Go checks, all passing: Codex's five model/late-context/map-presence tests; OpenCode's four metadata/model/validation tests; live's model-resolution and repeated-history-read tests; `TestModelSummaryLabelsUnknownWithoutChangingGroups`; and `TestComputeModelMetrics_AdversarialEdgeCases` (all with `-count=1`). SQLite, rollout and API fixtures are self-contained. Reused `/tmp/al70-a12-tester-evidence.json` for the broader affected backend and 112-unit/149-integration frontend checks, with source/test oracles inspected independently.

Limits: this is a scoped diff review, not a new complete cross-machine refresh/migration audit. Those unchanged mechanisms retain earlier evidence. Parent owns main integration, canonical make, standalone binary verification, consolidation and publication. No material unresolved gap within the assigned data integrity scope.
