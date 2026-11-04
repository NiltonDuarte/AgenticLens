# AL-67 technical contract

## Baseline and approval boundary

Ticket: AL-67. Work type: docs. Baseline revision: a84adbe5be46c2ef52ede0987a417562715d38e1. Workspace: alspaces/AL-67. Source requirement: [ticket specification](../plan/to-do/AL-67%20review%20telemetry%20repository%20interfaces%20and%20dependency%20injection.md).

This contract proposes a documentation-only architecture assessment of telemetry repository contracts and dependency wiring across OpenCode schema families v1 and v2 and Codex. Explicit user approval of this contract and iteration 01 plan is required before review execution. Invoking the development skill does not approve a repository refactor.

## Problem and delivered result

Determine which useful source operations already share contracts, how each consumer selects and obtains its source, and where source/storage details cross responsibility boundaries. Deliver a source-grounded account of current behavior and a bounded target design that maximizes meaningful sharing without inventing parity or combining ledger persistence, source extraction and live observation into one oversized repository.

The durable deliverable is ARCHITECTURE_ASSESSMENT.md: current and proposed dependency diagrams, an operation/implementation matrix, three source-family request walkthroughs, consumer-specific selection and lifecycle evidence, prioritized findings and staged recommendations. Distinguish implemented architecture from proposals in every section. SUMMARY.md records verification and review outcomes; PULL_REQUEST.md describes the documentation delivery. Temporary review reports follow canonical retention after consolidation; the architectural assessment remains the product deliverable.

## Read scope and architectural invariants

Read Go domain types and pure aggregation, adapter implementations, CLI composition, configuration, backfill/recovery, capture integration when reached, HTTP and live services, and the TypeScript contracts/consumers receiving their output. No application, schema, runtime-script, manifest, generated asset or test edits are authorized. Existing tests may run after approval; missing coverage becomes a recommendation rather than new tests.

Preserve Functional Core, Imperative Shell: domain calculations stay pure and storage/schema selection belongs at imperative boundaries. Favor small consumer-owned interfaces, with separate optional capabilities where evidence differs. A v1/v2 dispatcher behind an OpenCode implementation is acceptable; inspect its actual guarantees separately from harness selection.

The portable ledger contract and fact fidelity rules in TECHNICAL_GUIDELINES.md govern the assessment. Exact stored session/source identity and source-proven parent links remain authoritative routing evidence. Distinct occurrences, unavailable owners, other-harness records and persisted baseline visibility must survive partial or absent enrichment. Unsupported, unavailable, partial and complete-empty outcomes are distinct; a missing source is neither successful empty data nor measured zero. Review recommendations cannot impose a new authenticity or refresh policy.

## Required evidence and acceptance boundaries

1. Inventory exact interfaces, owners, implementations, constructors and callers; distinguish workspace-ledger access, native source operations and higher-level services. Interface absence is a valid observation when established by source evidence.
2. Compare discovery, exact-ID lookup, extraction, child traversal, trace hydration, refresh/recovery and live observation across all three source families. Record ownership, identity, snapshot completeness, ordering, errors and capabilities, with evidence and explicit unknowns.
3. Trace one concrete end-to-end request per family from effective configuration/flags to selected implementation and consumer output. Independently inspect generate, backfill, server/hydration and live paths, including local settings, explicit overrides and mixed-source owner routing. Never infer live capability from offline extraction support.
4. Assess resource lifetime/read-only access, cancellation/timeouts, concurrency/subscribers, caches/staleness, work bounds and cleanup. Distinguish justified dispatch from coupling with observable consequences.
5. Tie findings to reachable triggers and supported requirements; record numeric severity, confidence, source references and whether pre-existing debt or a recommendation. Reconcile counterevidence. Each proposed remediation names affected paths, compatibility guarantees, verification needed and a bounded sequence. Existing defects may remain accepted documentation findings with an explicit proposed follow-up; do not imply they were fixed.
6. Record tested revision, existing commands/results, reused evidence, gaps and environment limitations. Synthetic tests demonstrate fixture behavior only. Real-source availability and runtime versions require separate supported evidence.
7. Publish a documentation PR after required review and integration checks; do not merge it. Telemetry is generated with the explicit Codex source at final delivery, checked against available session evidence and honestly marked incomplete if sources cannot be extracted.

## Adjacent ticket boundaries

Reuse relevant AL-55 and AL-62 evidence without modifying those workspaces or treating historical findings as current proof. AL-63 owns source capability parity audit, AL-62 skill/delegation ownership fixes, AL-64 duration fixes and AL-66 generation toggles. This review traces their relevant contracts but does not duplicate implementations, create blocking dependencies or silently approve their remaining scope.

## Guidance

Apply TECHNICAL_GUIDELINES.md, docs/REGRESSION_PREVENTION.md and the development/review skills. All ticket artifacts use plain prose without emojis. The iteration plan supplies the execution roster and test selection; material scope changes require recorded steering and renewed approval when outside this contract.
