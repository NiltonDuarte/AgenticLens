h3. Title: AL-65: Research default source-backed recomputation and the future of Force

*GOAL*

As an operator, I want a recommendation for understandable, safe backfill defaults so that available source facts refresh without requiring Force while unavailable history remains intact.

*CONTEXT*

The user expects backfill to follow sessions already recorded in a ledger and recompute when original source metrics are available, while retaining stored facts when source data is unavailable. Current --force only bypasses completed-ledger skipping; --recompute-db and --recompute-git already bypass that skip and select different work. Force alone neither rereads sources nor enables heuristic matching. Research is requested; no default/flag behavior change is approved.

*ACCEPTANCE CRITERIA*

* Interpret availability as original session/metric source records needed for recomputation, never merely previously saved ledger metrics.
* Evaluate recomputing exact stored session identities by default without skipping complete ledgers; unavailable sessions retain existing facts rather than receiving zero/empty replacements.
* Preserve incomplete-tree safeguards, previous trees and skill/task snapshots, unrelated harness facts, and explicit opt-in heuristic matching under every proposal; do not publish destructive partial replacements or claim unavailable recovery complete.
* Compare ordinary backfill, --force, --recompute-db, --recompute-git and combinations. Explain useful work Force performs without recompute flags and whether clearer help alone would meet the expectation.
* Assess runtime, needless writes/timestamp churn, unchanged-repeat stability, source availability per harness, and compatibility for existing scripts/workflows using evidence with stated limits.
* Recommend retaining, changing or replacing --force only after research; distinguish the requested default source-backed behavior from simply enabling current Force.
* Provide proposed available/unavailable/retained outcome messages and migration/compatibility options. Any behavior change requires subsequent explicit approval and a separate implementation session.

*Status*: Blocked
*Blocked by*: AL-63
*Work type*: docs

*IMPLEMENTATION NOTES*

Discovery only: no application, schema, runtime or new test changes. Use existing evidence and disposable probes; label unavailable and unexecuted cases. Preserve the original AL-55 ledger. [Prior reconciliation](../../PO-19/SPEC.md) records shipped foundations and historical limits.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
