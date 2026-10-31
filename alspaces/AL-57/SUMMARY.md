# AL-57: Evidence-aware telemetry cost

## Status and authorization

Implementation authorized. User selected AL-57, reviewed the proposed scope, and answered the approval question: "remove legacy_unverified type. treat legacy costs as observed. keep it simple". The first approved contract includes this amendment; iteration 02 records the acceptance delta. Valid legacy numeric values, including zero and historical Codex zero, are observed. No historical special state, count or display is implemented. AL-55 is delivered and AL2 development remains queued.

## Baseline and ownership

Initial Git status was clean at 45cb679f21475cd14b7dd4cbc25996a39d734403. The existing worktree was retained and the ticket branch is fix/AL-57-cost-semantics. Workspace root is alspaces, read from .agenticlens-config.yaml. Backend, frontend, relevant compatibility documentation, AL-57 workspace documents and its index/spec lifecycle paths are task-owned. Parent owns integration and delivery; backend and frontend coders have disjoint writable application paths.

## Work

- Read the attached PO-17 thread through its final page and the delivered local recommendations.
- Assigned the planner sole ownership of CONTRACT.md and iterations/01/PLAN.md, with zero application, schema, or test writes.
- Created the proposed CONTRACT.md and iterations/01/PLAN.md. Parent reviewed them for consistency with the planning-only authorization and PO-17 hold.
- Proposal includes partial coalesced invocation evidence, known subtotal and coverage, legacy numeric amounts treated as observed, v5 persistence/API changes, and explicit read-only canonical-v4 compatibility.
- Recorded explicit amended acceptance in CONTRACT.md and iterations/02/PLAN.md; dispatched separate backend and frontend coders with disjoint ownership.
- The original iteration 01 proposal is historical, superseded where it conflicts with the accepted legacy numeric policy.
- Initial planning whitespace/scope checks passed, followed by approved implementation, adversarial testing and independent review. Reviewed code is committed; final telemetry and PR delivery follow the successful integration gate.

## Implementation

- Domain: three cost evidence states and composable invocation coverage, nullable complete-cost projections, precise complete-first cost ordering, unavailable cost-derived ratios and cumulative totals.
- Adapters: presence-aware OpenCode/Codex extraction, partial coalesced source amounts, strict v5 persistence/API validation, read-only v4 compatibility, explicit v4-v5 migration, legacy v1-v3 presence preservation, and exact-owner refresh safeguards. Valid historical numeric amounts remain observed.
- Presentation: CLI/Markdown and dashboard show known subtotals with coverage, preserve true zero, sort incomplete values last, and leave chart gaps for unavailable costs. Partial task references show known subtotals without invocation coverage and remain excluded from invocation totals.
- Documentation: cost/API compatibility guide linked from README and setup guidance; original unapproved history policy explicitly superseded.

## Verification and limits

Coder Go package tests/vet and frontend typecheck/lint/build passed. Independent tester's make test passed 12 Go packages under race and 754 frontend tests; subsequent focused domain/HTTP race tests and 64 task-panel tests passed after review corrections. See iterations/02/TESTING.md for tested source fingerprints and commands; the final integrated make gate passed on implementation commit 71b3dcf92c8f1ef2428ef374b40cce18f8e63dfb.

Review corrections: fleet null-map values require matching coverage; cost_usd HTTP query selects the actual comparator; micro-cost sorting uses complete precision with stable identity ties; partial task amounts retain their subtotal and stay excluded from invocation totals; built-in migration registration errors are handled explicitly.

Current source fingerprint (217 backend/web paths; sorted path + NUL + bytes + NUL, ignored generated files excluded): 82d6d7261258a2b33dd8350b64a5215185c2e24929a5f5301e0f25d9f6ac5c81. No browser geometry claim. Existing frontend React act warnings are recorded by tester. No live historical ledgers were bulk migrated. Old binaries reject v5; numeric-only API clients must update.

Telemetry extraction is intentionally deferred to final delivery; no ledger has been fabricated.

## Telemetry source evidence (not an extracted ledger)

The actual Codex root is 01a192c4-c9a6-7022-97da-19f433d6c5c3 in the normal ~/.codex/sessions store. session_meta source is vscode. Child session_meta records explicitly link parent_thread_id to that root:

- Planner: 01a192c6-4542-7501-99f4-9e010bd1977f.
- Backend coder: 01a192d0-c616-7413-b55b-cfc48d4f0a22.
- Frontend coder: 01a192d0-f684-7b60-b166-7429e80de682.

- Tester: 01a192d8-86bf-7201-a128-524172605e87.
- Reviewer: 01a192dd-93d1-7a32-a417-c3de8554a73b.

All five child session_meta records identify the actual Codex root as parent. Final ledger must retain all six sessions. Raw rollout files remain outside Git.


## Integration checkpoint

Reviewed implementation committed as 71b3dcf92c8f1ef2428ef374b40cce18f8e63dfb. git fetch origin main and git merge origin/main --no-edit completed with no new changes; origin/main is 45cb679f21475cd14b7dd4cbc25996a39d734403. Required full make passed on that implementation commit: Go vet/race checks, frontend typecheck/lint, all 755 tests in 35 files with zero skips, current frontend build, and linux/amd64, linux/arm64, darwin/arm64, darwin/amd64 binaries. Existing branch PR search returned none; delivery will use a new PR.


## Final delivery verification

- Independent review APPROVED with no open findings; final reviewed application fingerprint is unchanged by delivery documentation.
- Full make passed on 71b3dcf92c8f1ef2428ef374b40cce18f8e63dfb against origin/main 45cb679f21475cd14b7dd4cbc25996a39d734403.
- Built generate/migrate/backfill/server help inspected; schema migration defaults to 5 and source codex/explicit session selection are supported.
- Standalone linux/amd64 binary served exact current HTML, JS and CSS plus SPA fallback with an empty temporary workspace and no executable lookup PATH (Node unavailable). Process stopped after smoke verification.
- web/dist and backend/internal/adapters/web/dist match and remain ignored under current tracking policy. No generated assets forced into Git.
- Completed AL-57 specification moved to plan/done and its index entry removed. No indexed downstream ticket lists AL-57 as a blocker; AL2 development remains queued.
- Existing npm audit advisories and React act warnings were emitted; no dependency versions changed and all gates exited successfully.
- Telemetry is generated at the final step below using the actual Codex root and five children; no unavailable-source facts or numeric cost estimates are fabricated.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 18
- **Active Agent Duration:** 3h 34m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 363.6k tokens (182% of 200k healthy ceiling, 0 compactions) | Subagents peak 372.5k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 4.8 avg turns/run (peak: 8 turns) | Planner 6.0, Coder 5.3, Tester 7.0, Reviewer 6.0, Explorer 1.5
- **Test-to-Production Code Ratio:** 0.5:1 (+1,538 test lines / +2,949 prod lines)
- **Code Delivered:** +4,487 / -1,646 lines across 76 code files (excluding docs)
- **Reasoning Effort:** 171,365 reasoning tokens (28.3% of generated output)
- **Tool Health & Self-Correction:** 871 tool calls (55 self-corrections, 93.7% success rate)

### Anomalies & Health Flags
- **Status:** 23 execution anomalies detected (12 CRITICAL, 11 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.5k tokens) — Context tokens reached 200.5k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+113.9k tokens) — Context tokens surged by 113.9k to 113.9k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (201.3k tokens) — Context tokens reached 201.3k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.4k tokens) — Context tokens reached 200.4k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (201.6k tokens) — Context tokens reached 201.6k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (205.2k tokens) — Context tokens reached 205.2k, breaching the 200k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (15 cycles) — Subagent roles 'coder' and 'tester' alternated for 15 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (16 cycles) — Subagent roles 'coder' and 'tester' alternated for 16 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> tester (7 cycles) — Subagent roles 'reviewer' and 'tester' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (9 cycles) — Subagent roles 'coder' and 'tester' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: explorer <-> planner (10 cycles) — Subagent roles 'explorer' and 'planner' alternated for 10 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (164.6k tokens) — Context tokens reached 164.6k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.2k tokens) — Context tokens reached 163.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.2k tokens) — Context tokens reached 162.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165.8k tokens) — Context tokens reached 165.8k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.9k tokens) — Context tokens reached 160.9k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.7k tokens) — Context tokens reached 163.7k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> coder (3 cycles) — Subagent roles 'reviewer' and 'coder' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> tester (4 cycles) — Subagent roles 'reviewer' and 'tester' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> reviewer (3 cycles) — Subagent roles 'tester' and 'reviewer' alternated for 3 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 195,253 | 29,620 | 7,984,256 | 97.6% | 199.7k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 3 | 1,159,583 | 140,210 | 55,306,112 | 97.9% | 324.6k | BLOATED (162%) | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **tester** | 1 | 317,470 | 70,626 | 40,916,736 | 99.2% | 372.5k | BLOATED (186%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 1 | 781,695 | 36,480 | 21,988,736 | 96.6% | 359.7k | BLOATED (180%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **explorer** | 2 | 156,374 | 11,791 | 2,473,344 | 94.1% | 105.9k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **rootagent** | 1 | 296,339 | 146,268 | 78,735,232 | 99.6% | 363.6k | BLOATED (182%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **9** | **2,906,714** | **434,995** | **207,404,416** | **98.6%** | **372.5k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 9 | 2,906,714 | 434,995 | 207,404,416 | 98.6% | Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations) |
| **Total** | **9** | **2,906,714** | **434,995** | **207,404,416** | **98.6%** | **Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations)** |

## User review correction: JSON key order

User identified that the generated ledger's keys were reordered. Cause: cost projection decoded JSON objects into maps and marshaled them alphabetically. This session is correcting that boundary to retain established root/nested DTO order and keep new evidence adjacent to its cost, without changing values or cost semantics. Existing AL-57 PR #89 remains open. Focused tests and independent review approved the correction; integrated verification passed and final ledger refresh follows.


## JSON key-order correction verification

Correction commit 522ed96 was merged with current origin/main 36bea1f18f2256392428329136b3241e98d09854 as 8257196afaf6727c48ae3f4afebe280c1537352a. Upstream integration added planning documents only; reviewed application fingerprint remains 9f3f40c6ccfc5b5b40e3227277ccc1ff62baa504f1e096e5907692520020d9ae across 218 paths.

Full make passed on that integrated commit: Go vet/race checks, frontend typecheck/lint, 755 tests in 35 files, production assets and four platform binaries. Standalone binary smoke passed current HTML/JS/CSS and an actual ordered HTTP version response with Node unavailable on PATH; asset trees match. Existing React act warnings/advisories remain recorded, with no dependency changes or skipped gates.

The corrected serializer preserves established root, session, invocation, token, code and evidence field order; new evidence follows cost_usd. Cost values, availability semantics, session identity and array order are preserved. The final telemetry refresh updates this session ledger using the corrected built binary, retaining the existing six source sessions.


Corrected ledger verification passed: established field order at root/code/session/invocation/token/evidence levels, all six prior Codex source sessions retained with matching parent/source identities and invocation multiplicity, and current unavailable cost semantics intact. Usage is refreshed from the real source sessions; this is not a manual reformat of an outdated ledger.


## Iteration 03: typed cost boundaries

Source: user rejected the nesting/structural byte dispatch in cost_json.go, accepted the proposed typed DTO replacement with "$agentic-lens-dev fix it", and explicitly included domain/migration/cost_decode.go. This pre-approves the structural refactor while preserving the frozen cost and compatibility contract. No repeat permission request is needed for this scope.

Initial status is clean at ee7558749c7101ddc4134cbdd22961bcba0a5a09 on fix/AL-57-cost-semantics. PR #89 remains open. Planner owns iterations/03/PLAN.md; application implementation remains delegated through coder/tester/reviewer roles. Remove generic rewriting and custom ordered-object machinery, rather than moving their nesting into smaller helpers. Typed fields and conversions own nullable values and field order; version-specific decoders preserve cost presence.


Iteration 03 plan is approved by the user's explicit request to implement the typed alternative. Adapter coder owns adapter/HTTP/export DTOs and shared pure cost validation; migration coder owns only domain/migration. Existing files and source facts are preserved across the shared checkout. Tests and review will verify actual wire payloads and version decoding, including transfer of meaningful assertions from the removed arbitrary-JSON transformer tests.


Iteration 03 implementation replaces generic JSON rewriting with concrete nullable fact/task, aggregate, export and HTTP DTOs and typed version-selected migration readers. The adapter walker and custom ordered-object serializer are deleted. Shared amount/evidence validation remains pure in domain/cost.go. Actual output order now follows declared fields directly, and unrelated response families bypass cost projection.

Telemetry source addition: migration coder 01a19405-8d62-7f70-b9f9-31f06c174e3d explicitly identifies root 01a192c4-c9a6-7022-97da-19f433d6c5c3 as parent in session_meta. Final extraction must preserve the existing six sessions and include this seventh real session. Raw rollouts remain outside Git.


## Iteration 03 review and integration

Independent testing and review approved the typed refactor with no open findings. Reviewed source fingerprint 3edc1375f8dd66983375d8b73d3b9cb3f73b2cf7d67abe08878d932f3cd591f1 covers 219 existing backend/web source files, with deleted walker files excluded. Every meaningful removed generic-transform assertion has a concrete schema/persistence/HTTP/migration/CLI recipient. Focused owner tests passed with and without race detection.

Implementation committed as 3c8f09331c925cf0c2ecc8c4d78ada64414ef099; fetched/merged origin/main 36bea1f18f2256392428329136b3241e98d09854 with no new changes. Full make passed on that exact implementation commit: Go vet/race checks across all 12 packages, frontend typecheck/lint, 755 frontend tests in 35 files, current embedded assets and four platform binaries. Source semantics/schema/API shapes are unchanged by this structural correction; no new migration rollout or frontend changes are required.


## Iteration 03 final verification

The reviewed application source is committed as 3c8f09331c925cf0c2ecc8c4d78ada64414ef099 and integrated with origin/main 36bea1f18f2256392428329136b3241e98d09854. Full make passed without skips. Standalone linux/amd64 binary served byte-identical current HTML/JS/CSS, preserved actual version response field order, and returned empty fleet cost null with an empty cost summary; no Node executable was available on its PATH. The two generated asset trees match and remain ignored.

The semantic contract and schema/API shapes are unchanged. Both generic adapter serializer files are removed; migration cost decoding now reads concrete known-schema records, preserving legacy presence, invalid-v5 rejection, source ownership and multiplicity. Existing arbitrary-transform-only tests were removed with their unsupported API; meaningful assertions remain in named real boundary owners documented in TESTING.md and REVIEW.md. No additional live historical ledger migration was performed.

Final extraction below must retain the six existing real Codex sessions and add the explicit migration coder session, for seven sessions total. Old cost states and source ownership remain intact; current usage is refreshed from the actual source files rather than fabricated.


Iteration 03 final telemetry verification passed: all six prior real Codex sessions remain, the explicit migration child is included (seven total), parent/source identities and invocation multiplicity are preserved, and usage/tools are nonempty. The regenerated file retains established root/code/session/invocation/token/evidence key order. All fresh Codex costs remain explicitly unavailable. Raw source transcripts stay outside Git.


## Iteration 04: usable cost cards and filtered spend average

Source: user rejected the cost-card presentation in ticket deep dive, metric progression and fleet overview, and explicitly required comparison average spend to exclude tickets with unavailable costs and respect filters. Initial status clean at 52234b3c099fe1a6c1b2e1dd3d62c1db08ad7242; PR #89 remains open. This authorizes the bounded UI/benchmark correction.

The large card value currently contains formatCostUSD's full coverage sentence. Replace it with a short financial headline and appropriate compact state/coverage support. Comparison average spend must use only complete numeric costs from filteredRows, including observed zero; unavailable/partial tickets are excluded from this calculation and its denominator. All-unavailable/empty remains unavailable. This supersedes the former blanket-unavailable average rule only for this benchmark; other metrics retain their own filtered population. Stored evidence, aggregate completeness, nullable API contracts and schema behavior remain unchanged.


## Iteration 04 implementation and review

The three affected views now share CostKPICard: primary values contain only currency or a dash; partial/unknown/empty state and useful view context remain visible in secondary text, with full keyboard-accessible coverage in the existing tooltip. The verbose table formatter and backend/schema/chart semantics are unchanged.

Comparison Average Spend independently accumulates complete finite nonnegative ticket costs from filteredRows, includes zero, excludes unavailable/partial values and shows the eligible count. Other comparison metrics and rows retain all filtered tickets. The actual comparison view has no time-window control; its existing search/type/prefix/skill and refetch paths are verified without adding new controls.

Coder focused suites, independent comparison filter/refetch tests, typecheck/lint and canonical frontend build passed. Independent review approved with no findings at 220-source fingerprint 2aba254e521adf59dd0ebfc5d79b3fba6c9af117a2aefee3cc1902d94d01335f. Implementation committed as 3fb374cfd93aca6bbe26097e35214183ae78e1b8; fetched/merged current origin/main with no additional changes. Full make passed on this exact implementation commit: Go vet/race checks, frontend typecheck/lint, all 761 tests in 35 files, canonical frontend assets and four platform builds.


## Iteration 04 final verification

Full make passed on 3fb374cfd93aca6bbe26097e35214183ae78e1b8 after current-main integration, with no skips. Standalone linux/amd64 smoke served byte-identical current HTML/JS/CSS, verified the new card/eligible-count bundle and implementation version, and ran without Node on PATH. Frontend and embed asset trees match and remain ignored. DOM/accessibility/content/filter tests establish the intended behavior; no browser geometry inspection is claimed.

Filtered-average checks include mixed complete/zero/partial/unavailable records, actual skill/type/prefix/search composition and recovery, all-unavailable/empty selections, and refetched dataset replacement. Other benchmark denominator/table/scatter populations remain all filtered rows. Legacy numeric-observed policy, backend schema/wire contracts and chart behavior are unchanged. No live historical ledgers were migrated for this UI correction.

Final telemetry below preserves the existing seven real Codex source sessions and their identities; it refreshes session facts using the built binary rather than manufacturing a successful ledger.


Iteration 04 telemetry verification passed: all seven existing real Codex sessions retained with the same parent/source identities and invocation multiplicity, nonempty usage/tools, schema 5 and established key order. Fresh Codex costs remain unavailable without estimates or invented zero. Raw rollouts remain outside Git.


## Iteration 04 review correction: remove bucket subtitle

User asked to remove "time buckets" from the progression cost card. The count describes aggregated intervals, not necessarily visible bars. Remove only that redundant subtitle and update its existing UI expectation; cost states/coverage and time grouping/charts remain unchanged. Baseline c7c87587e9366c44b79f9ec5295a22c70bbfce6c was clean, PR #89 open. No new tests planned for the reversible copy correction; existing owner checks and required delivery gates are reused/run as appropriate.


Bucket-subtitle correction: focused testing/review approved with no findings at 220-source fingerprint 460adc0af5b0db02ef39eb57cfeb48367f061ed276d2af9feb7a46ca7a4e7440. The only production edit removes the redundant secondary prop; the existing partial-state expectation retains currency and accessible coverage assertions. Typecheck/lint and all 16 progression tests passed. Implementation committed as 9736d8b057b5bb08ced88f07bb04deb70c00311f; current main integrated without changes. Required full make passed: Go vet/race, frontend typecheck/lint, 761 frontend tests, canonical assets and four platform builds.


Final bucket-copy gate passed on 9736d8b: full make, matching frontend/embed trees and standalone binary serving exact current assets with the redundant label absent from its JS bundle. No Node runtime was available. No bucket/chart math changed, no new tests were added, and no browser geometry claim is made. Final telemetry refresh retains the existing seven actual Codex sessions.


Copy-correction telemetry verification passed: all seven existing actual Codex sessions, parent/source identity, invocation multiplicity and unavailable cost evidence retained; usage is nonempty. Raw source rollouts remain outside Git.


## Iteration 05: validation and serialization ownership plan

Source: user requested a concrete plan before implementation for six demonstrated/structural concerns: frontend shape-inferred cost validation, DTO responsibility/placement, fact codec cohesion, redundant scalar/summary calculations, test-only computed reverse conversions, and duplicated strict decoder mechanics. This is planning authorization only; earlier implementation approval does not bypass the explicit new planning gate.

Baseline 412bc08e3199af04c03009ed60ca1bd3b592dd85 was clean, same AL-57 branch/worktree, PR #89 open. Preserve the frozen financial semantics and accepted iteration 04 filtered-average behavior, historical numeric zero, read-only v4, strict required v5 evidence, source ownership/multiplicity, JSON key order and persisted baseline on failed optional hydration.

Planner owns iterations/05/PLAN.md. Separate read-only explorers trace DTO/reverse-conversion consumers and scalar/summary compatibility/precision; no production/schema/test changes or early telemetry extraction are authorized. The plan must distinguish proposed fixes, necessary retained duplication and broader deferred redesign, with actual assertion recipients and verification boundaries.


Iteration 05 planning evidence: both frontend validation defects reproduced through actual API clients. Independent audits confirm computed-metrics ToDomain and its exclusive reverse subtree have only test callers; fact-ledger ToDomain has live repository/backfill consumers. SerializedLedgerDTO has only three test declarations and no production consumer, and validateAnomalyEvidence has none. The plan proposes removing those proven unused paths, preserving independent emitted-JSON and live ingestion assertions.

Scalar audit distinguishes safe local derivation from rounded cross-ledger compatibility. Two source amounts 0.00004 each demonstrate why fleet scalar 0 and raw summary 0.00008 cannot be unified by rounding the merged summary without changing behavior. Broader representation/rounding redesign is explicitly deferred. Current plan 05 and parent checkpoint review await user approval; production files remain unchanged.


Iteration 05 approval: user explicitly selected "Approve the full plan" after the concrete plan was presented. All six bounded corrections, justified retentions and deferred broader redesign are accepted. Implementation is now authorized; do not ask for the same permission again. The only preexisting staged changes at this checkpoint are this ticket's planning artifacts. Baseline remains 412bc08e3199af04c03009ed60ca1bd3b592dd85.


## Iteration 05 implemented dispositions

1. Fixed frontend shape inference with endpoint/schema-selected readers; dynamic role/model keys are opaque and v5 invocation/task cost evidence is required. Both demonstrated failures are regression-tested through real API client entry points.
2. Fixed DTO ownership: shared metric projections, compact exports and persisted facts remain in responsibility-specific adapter files; HTTP-only response DTOs/converters now live in the HTTP boundary with no reverse dependency or facade aliases. Distinct privacy/wire contracts are retained with justification.
3. Fixed fact cohesion by colocating per-kind definitions, codecs and conversions. Production AgenticFactLedgerDTO.ToDomain and its live fact/anomaly chain are retained.
4. Fixed redundant local accumulation/caches, preserving exact legacy precision. Rounded fleet/role accumulations and independent in-memory compatibility fallbacks are retained with demonstrated caller/rounding reasons. Global representation/rounding redesign is deferred outside approved scope.
5. Fixed unused code: removed computed metrics reverse graph, test-only alias, dead fleet/brittle/root event/tool DTO paths and unused anomaly helpers. Meaningful assertions now target real emitted JSON and live ingestion/output paths; unsupported computed-input/stdlib-only promises are intentionally retired.
6. Fixed decoder duplication with narrow configured-decoder/EOF/presence mechanics; explicit v4/v5 policies and historical migration tolerance remain separately owned.

Independent review APPROVED, no open findings. Final reviewed 230-file source fingerprint f49d9f8a69b019dbeba07fa2e5891c1da6509a40683197bc6bfffed4c02998ad; earlier a14 snapshot is kept distinct in TESTING.md. All 25 retained DTO field/tag sequences match baseline. Code committed as 3fcf1575161390a73202babf921c2e9e894f8873 and integrated current origin/main unchanged. Full make passed: Go vet/race across 12 packages, frontend typecheck/lint, all 775 tests in 35 files, canonical assets and four platform binaries.

New real source sessions for the two read-only audits: DTO audit 01a19431-b6ea-70c1-ae3e-f0336e71274e and cost audit 01a19431-db4b-79a1-85dd-bbdbea811a51. Each session_meta explicitly links to root 01a192c4-c9a6-7022-97da-19f433d6c5c3; final telemetry must retain prior seven sessions and include these two, nine total. Raw source files remain outside Git.


## Iteration 05 final delivery gate

Full make passed on 3fcf1575161390a73202babf921c2e9e894f8873 after current-main integration. Standalone linux/amd64 smoke served byte-identical current HTML/JS/CSS, preserved version field order, and returned the moved HTTP fleet/workspace contracts correctly for an empty fixture. Node was unavailable on PATH; generated frontend and embed trees match and remain ignored. No browser geometry claim. Existing npm advisories/React act warnings remain recorded; no dependency changes or skipped gates.

All six implementation findings are fixed, with compatibility retentions and broader deferred representation/rounding/loader policy documented in REVIEW.md. No new schema bump, source refresh policy or live historical migration was introduced. Fact ToDomain remains production-required; computed reverse/test-only aliases and the approved dead helper families are gone.

Final telemetry below must preserve prior seven real Codex sessions plus the two explicitly linked audit sessions, nine total, using the actual source and literal nonce.


Iteration 05 final telemetry verification passed: prior seven actual Codex sessions retained and both explicit audit sessions included (nine total), with correct parent/source identity, prior invocation multiplicity, established key order and nonempty usage/tool records. Fresh Codex costs remain unavailable without estimates or invented zero. No raw transcripts are committed.
