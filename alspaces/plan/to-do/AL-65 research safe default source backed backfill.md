h3. Title: AL-65: Research default source-backed recomputation and the future of Force

*TARGET*: AgenticLens (Backfill CLI semantics / source recovery)

*Status*: Blocked
*Work type*: docs
*Blocked by*: AL-63
*Blocks*: None
*Origin*: PO-18; user review dated 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679.

*CONTEXT*

The user expects backfill to follow sessions already recorded in a ledger and recompute when original source metrics are available, while retaining stored facts when source data is unavailable. Current --force only bypasses completed-ledger skipping; --recompute-db and --recompute-git already bypass that skip and select different work. Force alone neither rereads sources nor enables heuristic matching. Research is requested; no default/flag behavior change is approved.

*GOAL*

As an operator, I want a recommendation for understandable, safe backfill defaults so that available source facts refresh without requiring Force while unavailable history remains intact.

*VERTICAL SLICE SCOPE*

* Domain Core: Evaluate per-session availability, snapshot completeness, preservation, idempotency, and incomplete/mixed-age reporting.
* Boundary Adapter: Evaluate exact-identity source-backed recovery across OpenCode v1/v2 and Codex using AL-63 evidence; distinguish source refresh, Git recomputation, migration, and ledger processing.
* Presentation: Propose default behavior, flag/help semantics and available/unavailable recovery diagnostics with concrete command scenarios; this ticket changes no command behavior or help.
* Quality: Research report with reproducible non-destructive experiments, script compatibility assessment, and an explicit recommendation/approval boundary. No application or test code changes.

*ACCEPTANCE CRITERIA*

* Interpret availability as original session/metric source records needed for recomputation, never merely previously saved ledger metrics.
* Evaluate recomputing exact stored session identities by default without skipping complete ledgers; unavailable sessions retain existing facts rather than receiving zero/empty replacements.
* Preserve incomplete-tree safeguards, previous trees and skill/task snapshots, unrelated harness facts, and explicit opt-in heuristic matching under every proposal; do not publish destructive partial replacements or claim unavailable recovery complete.
* Compare ordinary backfill, --force, --recompute-db, --recompute-git and combinations. Explain useful work Force performs without recompute flags and whether clearer help alone would meet the expectation.
* Assess runtime, needless writes/timestamp churn, unchanged-repeat stability, source availability per harness, and compatibility for existing scripts/workflows using evidence with stated limits.
* Recommend retaining, changing or replacing --force only after research; distinguish the requested default source-backed behavior from simply enabling current Force.
* Provide proposed available/unavailable/retained outcome messages and migration/compatibility options. Any behavior change requires subsequent explicit approval and a separate implementation session.

*DELIVERY CONSTRAINTS*

* Use AL identifiers, not AL2.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. FCIS; Go 1.23+ and TypeScript 5.x strict; zero any/interface{} and zero emojis.
* Tests and experiments use self-contained fixtures/disposable copies, never live ticket workspaces. Preserve alspaces/AL-55/agentic_metrics.json.
* No new capability is inferred from missing data. Separate source evidence, recorded observations, and hypotheses.

*DEFINITION OF DONE*

* Deliver the evidence report, limitations and follow-up recommendation in the active ticket workspace; no application fixes are part of this ticket.
* Record exact checks executed, failures, unavailable evidence and unexecuted checks. Existing test results may be reused for unchanged code.
* Maintain ticket summary/review/PR artifacts and source-grounded telemetry; publish a reviewable PR. Completion requires evidence, not just a plan.
