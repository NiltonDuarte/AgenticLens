# REV-5: Code smell audit and remediation

## Historical audit-only baseline (GEN-14)

Authorized by the user request to review all code, specifically code smells and production code used only by tests. Review only: application, tests, schemas, runtime scripts, and configuration remain unchanged. Report prioritized evidence and safe follow-up boundaries; do not delete code or tests.

Baseline: 36bea1f18f2256392428329136b3241e98d09854. Initial Git status clean. GEN-13 already belongs to documentation consolidation; GEN-14 was absent from workspaces and ticket index and is assigned to this audit. No specification or index entry required. No subagents launched.

## REV-5 iteration 02 approved remediation boundary

Source: user steering, "track this session under a REV ticket. fix the findings". The same audit task/session is reclassified as REV-5; historical GEN-14 labels and evidence remain intact. Baseline for remediation is e8d9003ea6388d1a750273478ec572863e965a72. This explicit request approves implementation of the seven findings and supersedes the audit-only restriction for iteration 02; no additional confirmation is required for these boundaries.

Fix configuration-pair failure recovery; remove the unread live buffer and confirmed unused production functions; remove the incorrectly typed obsolete progression client; preserve meaningful tests by exercising active production APIs; add runtime validation to the remaining HTTP client payload boundaries identified in F7. Preserve runtime output and ordering, session ownership/multiplicity, unavailable-data semantics, public HTTP routes, valid persisted/wire compatibility, implicit JSON/YAML codecs, and the actual revision cache. No schema version bump, new API endpoints, new abstraction framework, or change to configured ticket prefixes is required. Configured legacy trace-buffer settings may remain accepted as compatibility data; do not repurpose them to truncate current response data.

All inventory entries must receive a recorded disposition: removed with test rationale/recipient, retained with actual runtime/codec/contract evidence, or explicitly escalated if correction would change these boundaries. Test counts are not acceptance quotas. No copies of obsolete business algorithms in test-only helpers. Keep test fixtures self-contained and independent of the renamed workspace. Root owns integration, telemetry refresh with preserved session identities, commits and PR delivery.

## Upstream integration clarification

The user's subsequent request to merge main and review incoming code authorizes integrating main47af923's already shipped cost semantics and schema v5. Preserve these upstream contracts together with REV-5 fixes; the earlier prohibition on inventing a schema bump does not exclude integrating an upstream version. This remains iteration 02 routine integration. See INTEGRATION_PLAN.md for exact parents, ownership, preserved guarantees and renewed verification boundaries.

## Iteration 03 full code audit

User clarified scope to every repository code file. This expands audit coverage beyond changed code to the tracked live source/tests and delivery/configuration code, with supporting fixtures explicitly labeled. It does not authorize rewriting every discovered concern during this audit. Preserve current behavior and record evidence-based findings and per-file coverage in iterations/03; historical ticket artifacts and ignored build/dependency files are excluded.

## Iteration 04 approved P1 remediation

User explicitly requested fixing the highest-priority issues after the listener finding was reduced to P2. This authorizes ADP-01 and D3 implementation and supersedes iteration 03 audit-only acceptance for those two findings only. Preserve stored owner trees whenever Codex parsing returns warnings, including a possibly partial trailing write; expose incomplete recovery and let healthy independent trees refresh. Build releases from the exact selected tag commit before tests/build, reject absent tag identity, and keep forced rebuilds from creating unrelated releases or moving latest/purging candidates. Preserve schema v5 evidence, source/parent identity, other-harness facts, supported normal release behavior, embedded distribution and the local-only product boundary. No listener/authentication change, broad refactor, other audit finding remediation, remote release execution or release PR merge is included. Detailed grounded ownership and verification are in iterations/04/PLAN.md. Historical contract sections remain intact.

## Iteration 05 approved prevention guidance

User approved applying the research recommendations: correctness-first review skill/persona, aligned workflow references, bounded independent old/new reviewer evaluation and explicit incomplete evidence outcome. Documentation/session-audit work may proceed directly under AGENTS.md. Preserve implementation approval, product-read-only reviewer role, local-only scope and prior P1 fixes. No claim that wording alone prevents bugs; evaluation limits must be reported.

## Iteration 06 approved parallel-review design

User requested designing specialized parallel review in the skill kit. Scope is canonical documentation and aligned dispatch guidance: separate core focuses by default for substantive changes, optional risk specialists, honest capacity-limited waves, independent investigation, verified findings, revision-aware re-review and consolidated delivery gates. No new CI service, runtime scheduler, model switch or application change. Preserve previous approvals and role boundaries.

## Iteration 07 approved model policy and attribution

User requires every selected reviewer on both Sol 6.1 high and Luna 6 medium, configurable and attributable. Add coordinator-read model policy and durable run/finding reporting contract; explicitly dispatch exact settings, preserve independent discovery credit, retain no-findings/failed/refuted records and never silently substitute models. This supersedes iteration 06 no-model-switch boundary for these explicitly requested profiles only. No application telemetry schema or dashboard changes.

## Iteration 08 approved follow-up ticket

User requested a ticket for lifecycle hooks in installation. Create AL-68 in the shared backlog and index under continued REV-5 ownership. Scope is specification only: additive Codex/OpenCode hooks, session registration/capture/reconciliation, explicit ticket/parent identity, CLI diagnostics and verification. Preserve third-party skills/personas, local-only operation and no T3 source. Implementation belongs to a separate development session; current telemetry behavior and historical coverage limitations remain unchanged.
