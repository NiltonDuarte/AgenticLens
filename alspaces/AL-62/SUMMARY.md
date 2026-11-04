# AL-62: Session-owned skill and delegation facts

## Delivered behavior

- Domain: schema 6 stores skill/task occurrences only in their owning sessions, retaining native item/message/call/turn provenance, repeated occurrences and explicit outcomes. Metrics invocation rows are not occurrence owners.
- Adapters/CLI: Codex and OpenCode v1/v2 producers feed generate, capture, backfill, persistence and HTTP. Complete exact owners replace independently, including equal/older/empty snapshots; unavailable or malformed owners survive. Safe source-linked occurrence events remain available offline with explicit partial-event coverage.
- Presentation: nested current API, owner/session filters, fleet/comparison analytics, native timeline badges and child navigation use owned facts. Unknown outcomes never become measured success/failure. Raw agent references remain separate from proven child sessions.
- User-approved breaking behavior: current reads require schema 6. Explicit migration drops ownerless legacy skill/task records and audits counts. No root compatibility arrays/unresolved-history store remain. No bulk migration or change to live AL-55 data was performed.

## Scope, source evidence and approval

Initial Git status was clean on AL-62/fix-codex-skill-delegation-facts at baseline 19d70574d0e22b00f185c32d048baaf62792743e. User explicitly approved the initial contract/plan, then directed on 2026-11-04: "Delete ownerless records on migration. Drop backward compatibility." [Current contract](CONTRACT.md) and [iteration 02 delta](iterations/02/PLAN.md) supersede the earlier compatibility proposal; preserve known unavailable-owned history and all remaining guarantees.

Read-only real Codex 0.166.0 evidence: AL-55 root 01a19267-636b-7d61-85c6-db8d3d3eba77 and its four children exist; four explicit spawn calls join their outputs by native call ID. Outputs carry task_name references, resolved only by exact parent/agent-path metadata in read projections. Catalogue/path mentions are not skill activation. AL-62 root 01a1a3dc-3627-79c0-8b42-8c9cea2041f3 records direct selection (msg_01a1a3dc-3cc3-71f2-95e8-43f79fb7905f) plus attachment (msg_01a1a3dc-3cd7-7953-8be3-a7c4578f04f1), one activation request with unknown outcome. Detailed privacy-safe references remain in CONTRACT.md; raw transcripts stay outside Git.

## Consolidated review: APPROVED

Reviewed source fingerprint 7bc5d7ef512cdc2a031f087c9e6ab1826a2ae6e7c4f25dc1eaccf1a3b7104168 (442 files), committed as 8ec7bca8ff935b4632f489d1da7900183d37aa89. Current origin/main 19d70574 was merged with no changes. Parent consolidated approval follows the passing integrated checks below; no unresolved finding or material evidence gap remains.

Eight initial discovery cells covered behavior, data/failure, architecture and test-effectiveness on BOTH gpt-6.1-sol/high and gpt-6-luna/medium: seven CHANGES_REQUESTED and Luna test-effectiveness APPROVED/no findings. All final required correction cells APPROVED. The data Sol recheck retained DATA-01 as open for malformed outer fields, then approved its corrected source 3 regression. Sol behavior/architecture source 2 approvals carry because source 3 changes only malformed outer/required-field warnings and tests; successful parsing, schema, ownership and UI paths are unchanged. Remaining final cells reviewed source 3 directly. Model/effort reflects accepted explicit/inherited dispatch configuration, not independently inspected source execution metadata.

Permanent compact evidence: [REVIEW_SUMMARY.json](iterations/02/REVIEW_SUMMARY.json). Sixteen canonical findings retain numeric severity and every discovery tuple; verification actors are separate. Initial 12 review findings and 4 integration corrections are fixed:

| Finding and issue | Discovering actor / model / effort | Fix and implementing actor / model / effort |
| --- | --- | --- |
| DATA-01: Malformed recognized Codex skill arguments authorize deletion of stored facts. | Data reviewer / gpt-6.1-sol / high | malformed native arguments/outer fields mark the owner incomplete; accepted ledger bytes survive. Backend coder / gpt-6.1-sol / high. |
| DATA-02: Private registration and pinned-file history gates block valid portable exact-owner refresh. | Data reviewer / gpt-6.1-sol / high; Architecture reviewer / gpt-6.1-sol / high | valid history and complete exact refresh need no private receipt or pinned-file equality. Backend coder / gpt-6.1-sol / high. |
| DATA-03: Missing stored child prevents its complete parent snapshot from refreshing. | Data reviewer / gpt-6.1-sol / high | refresh complete owners independently and retain unavailable or malformed children. Backend coder / gpt-6.1-sol / high. |
| BEH-01: Distinct native skill attachments collapse into one occurrence. | Data reviewer / gpt-6.1-sol / high; Behavior reviewer / gpt-6.1-sol / high; Test reviewer / gpt-6.1-sol / high | merge only one adjacent selection/attachment envelope; retain distinct requests. Backend coder / gpt-6.1-sol / high. |
| BEH-02: Freshness guards reject successfully complete equal or older registered source snapshots. | Behavior reviewer / gpt-6.1-sol / high | complete equal/older capture snapshots replace their owner; true partial reads preserve history. Backend coder / gpt-6.1-sol / high. |
| BEH-03: Persisted namespace-derived child UUID is later treated as explicit source identity. | Behavior reviewer / gpt-6.1-sol / high; Architecture reviewer / gpt-6.1-sol / high | persist raw child references only; resolve namespace-derived UUIDs in read projections. Backend coder / gpt-6.1-sol / high. |
| BEH-04: Known child namespace contradiction is missed when the reported namespace has no matches. | Behavior reviewer / gpt-6.1-sol / high | report explicit child namespace contradictions even when no namespace candidate matches. Backend coder / gpt-6.1-sol / high. |
| BEH-05: OpenCode capture bypasses owned skill and task fact extraction. | Architecture reviewer / gpt-6.1-sol / high | exact OpenCode v1/v2 capture uses the shared owned occurrence producer. Backend coder / gpt-6.1-sol / high. |
| BEH-06: Native selected-skill requests have no produced timeline event despite the activation badge test. | Test reviewer / gpt-6.1-sol / high | source-backed activation and occurrence metadata survive save/reload and drive native timeline badges. Backend and frontend coders / gpt-6.1-sol / high. |
| DATA-04: Opaque Codex tool outputs are promoted to known successful skill or delegation outcomes. | Data reviewer / gpt-6-luna / medium | opaque outputs and child references stay outcome-unknown without positive outcome evidence. Backend coder / gpt-6.1-sol / high. |
| BEH-07: Opaque Codex agent_id is preferred over explicit child session_id and used as a navigable session. | Behavior reviewer / gpt-6-luna / medium | retain opaque agent IDs separately and navigate using explicit or proven session identity. Backend and frontend coders / gpt-6.1-sol / high. |
| ARCH-01: Current v6 task links may omit explicit unknown/succeeded/failed outcome. | Architecture reviewer / gpt-6-luna / medium | current task outcomes are required; legacy defaulting is migration-only. Backend and frontend coders / gpt-6.1-sol / high. |
| BEH-08: Frontend skill badges omit supported load_skill events and native call_id matching after persistence. | Orchestrator / gpt-6.1-sol / high | load_skill badges use native call_id independently of event ID representation. Frontend coder / gpt-6.1-sol / high. |
| DATA-05: Sparse retained occurrence events can erase cached anomalies or fabricate consecutive-failure streaks. | Backend coder / gpt-6.1-sol / high | mark sparse persisted events partial and preserve full-source anomaly summaries. Backend coder / gpt-6.1-sol / high. |
| DATA-06: Ticket anomaly aggregation excludes retained unavailable owners before publication. | Tester / gpt-6.1-sol / high | ticket anomaly aggregation includes every retained owner without rewriting historical DTOs. Backend coder / gpt-6.1-sol / high. |
| DATA-07: Capture with no fresh owners bypasses an explicit local source-to-other-ticket association. | Tester / gpt-6.1-sol / high | empty refresh still rejects positive same-store other-ticket bindings; absent receipts stay allowed. Backend coder / gpt-6.1-sol / high. |

Verification actors: tester and implementation coders (gpt-6.1-sol/high) ran development regressions; each focused reviewer on its configured profile verified corrected scope, including raw CLI and producer-to-HTTP/DOM controls where reported. Parent orchestrator (gpt-6.1-sol/high) integrated, ran make and standalone verification. A supplied-issue verifier is not credited as a discoverer. Detailed active review records are compacted only after this approval.

## Integrated verification

- make: PASS on implementation commit above after main integration; Go vet/race across 13 packages, frontend typecheck/lint and 818 tests in 36 suites; Python distribution/installation/capture checks (14+10+15); canonical frontend build and five native platform binaries.
- Standalone Linux binary version identifies commit 8ec7bca. Its actual HTTP index/JS/CSS exactly match web/dist and the embedded tree, with the source DB absent. JS index-mKcHXOp5.js SHA256 c95a2c84ef6b791e99ec90b79367ffc7479367278cd1922c19f8b711aae71bc9; CSS index-Bn43kjvY.css SHA256 586f9f4eb5f2b1cb7fe64ad5fd0f32a1eaccb14cb4cd36feaaee1105e7352541. Test server stopped.
- Actual CLI regressions cover unregistered/copy-path history, remote/mixed-source preservation, growth/repeat/complete-empty refresh, complete parent with missing/malformed child, malformed native payload byte preservation, migration drop counts, current-only rejection, safe offline occurrence metadata/anomaly preservation and scoped source/ticket/store/project guards. [TESTING.md](iterations/02/TESTING.md) records commands/oracle transfers and bounded evidence.
- Initial standalone attempt incorrectly used installed-launcher --al-* flags against raw source binary; help confirmed unsupported flags and the corrected raw server command passed. Final telemetry uses the built source binary from project CWD, with explicit source codex/type fix; no wrapper installation is required.

## Remaining limits and delivery

Real Codex evidence is grounded in the formats above; other/unrecorded formats are not claimed equivalent. OpenCode v1/v2 coverage uses self-contained SQLite/source fixtures; no live provider session or browser automation was claimed. Remote hosting/CI is not yet verified. Migration intentionally discards ownerless legacy auxiliary data; older ledgers need explicit migration. Existing original source logs and other ticket ledgers remain untouched.

Final telemetry extraction/identity verification and PR publication follow this approval as the final delivery steps.

## Telemetry coverage boundary

Independent raw session-meta inventory before final extraction found the AL-62 native Codex root and 13 direct children (14 sessions). Final verification will compare this exact-ID inventory with the generated ledger. Eight T3-delegated Luna review threads have service-level parent bindings, but the available thread metadata does not expose their native Codex session IDs; native parent links for those service threads are not proven. Do not invent sessions, parent links, delegation occurrences or usage to fill this gap. Report native-source-graph coverage separately from the broader service workflow; model execution settings are dispatch-acknowledged, not source-session cost evidence.

Lifecycle preparation moved the completed AL-62 specification to plan/done and promoted AL-63 to Ready because its sole AL-62 dependency is satisfied. No AL-63 implementation was performed.

Final ingestion first failed before extraction because the configured shared local capture registry has an unsupported coordinate layout (missing worktree mapping table). Existing operational state was preserved. Retry uses the documented per-command --capture-dir override with a fresh private temporary registry, normal Codex source home and explicit source/type. No configuration or application change, reset, source-log rewrite or empty ledger workaround was used.

The per-command override retry also failed because installed-state preflight validates the configured default before the override. A temporary ignored .agenticlens-config.local.yaml now selects the same fresh state path; it was created exclusively with 0600 permissions while preserving the shared configuration SHA and old registry. No hooks exist for this checkout and none were changed. This operational override will be removed after final identity verification; source code/reviewed fingerprint remain unchanged.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 5
- **Active Agent Duration:** 3h 27m (excludes user wait time)
- **Context Window Health:** CRITICAL — Root agent peak 272.8k tokens (136% of 200k healthy ceiling, 1 compactions) | Subagents peak 449.8k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.0 avg turns/run (peak: 6 turns) | Planner 2.0, Coder 5.5, Tester 4.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.7:1 (+2,617 test lines / +1,537 prod lines)
- **Code Delivered:** +4,333 / -2,595 lines across 104 code files (excluding docs)
- **Reasoning Effort:** 204,210 reasoning tokens (28.7% of generated output)
- **Tool Health & Self-Correction:** 1,250 tool calls (77 self-corrections, 93.8% success rate)

### Anomalies & Health Flags
- **Status:** 16 execution anomalies detected (9 CRITICAL, 7 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.1k tokens) — Context tokens reached 200.1k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (201.0k tokens) — Context tokens reached 201.0k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.9k tokens) — Context tokens reached 200.9k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.2k tokens) — Context tokens reached 200.2k, breaching the 200k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (25 cycles) — Subagent roles 'coder' and 'tester' alternated for 25 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (53 cycles) — Subagent roles 'coder' and 'tester' alternated for 53 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (5 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (66 cycles) — Subagent roles 'coder' and 'tester' alternated for 66 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (10 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 10 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.9k tokens) — Context tokens reached 160.9k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.0k tokens) — Context tokens reached 160.0k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (4 cycles) — Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 82,284 | 11,263 | 1,216,896 | 93.7% | 90.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 561,141 | 195,414 | 90,487,680 | 99.4% | 407.2k | CRITICAL (204%) | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 360,570 | 120,614 | 77,260,160 | 99.5% | 449.8k | CRITICAL (225%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 495,353 | 108,338 | 49,773,568 | 99.0% | 272.8k | BLOATED (136%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 2 | 208,687 | 16,771 | 4,379,648 | 95.5% | 134.4k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 218,900 | 17,747 | 3,742,592 | 94.5% | 131.3k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 3 | 221,581 | 22,073 | 3,466,624 | 94.0% | 117.4k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-tests** | 2 | 205,025 | 14,798 | 2,609,664 | 92.7% | 108.6k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **14** | **2,353,541** | **507,018** | **232,936,832** | **99.0%** | **449.8k** | **CRITICAL** | **Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 14 | 2,353,541 | 507,018 | 232,936,832 | 99.0% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **Total** | **14** | **2,353,541** | **507,018** | **232,936,832** | **99.0%** | **Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations)** |

Capture observation finalized: 2026-11-04T02:14:32.236614681Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.

### Final telemetry verification

Successful built-source CLI generation used literal nonce 7ea42988833fd9d0, --source codex, --ticket-type fix and the temporary private capture selection. It wrote schema 6 raw facts and rendered the metrics above. Independent metadata inventory and saved ledger match exactly: 14 native sessions (root+13 direct children), every native source/parent identity preserved, one root skill activation request with native primary message/attachment evidence and unknown outcome, and 13 distinct native delegation call IDs/child references. No root skills/task_links/unresolved store or namespace-inferred persisted child UUID exists. AL-62 had no prior ledger/other-harness sessions to preserve; actual cross-machine mixed-harness preservation is proven by the separate command regressions.

A standalone isolated API run reloaded this ACTUAL generated ledger with source DB absent: all 13 child-session links resolve from stored native metadata; fleet skill_adoption contains exactly one invocation, 0 known outcomes, 1 unknown outcome. Current API contains only session-owned arrays, and GET left ledger bytes unchanged. Test server stopped. This is real source->generation->save/reload->API/fleet verification, distinct from synthetic regression fixtures and without changing live AL-55 data. Frontend source-shaped fixture/DOM tests cover the rendered consumers.

The temporary ignored local overlay was removed and shared configuration SHA 6271a7a0e1b8b6482df6ea0dabfef4aa7624f16ea396c6dbf1e780c5551a7339 remained unchanged. Old shared operational state was not reset or upgraded. Source-native metrics cover only the 14 proven sessions, not the 8 unbound T3 service review sessions described above. Source costs are unavailable; no estimates or measured zeros were invented. Rendered native metrics: 3h 27m active agent duration, 99.0% cache hits, 449.8k peak context/CRITICAL. Capture collection counters describe the separate registered collection, not the 14 exact native rollout reads.
