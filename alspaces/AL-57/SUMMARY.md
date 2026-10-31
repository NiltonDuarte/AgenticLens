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
- **Software Development Life Cycle Iterations:** 13
- **Active Agent Duration:** 2h 33m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 285.0k tokens (143% of 200k healthy ceiling, 0 compactions) | Subagents peak 329.4k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 3.8 avg turns/run (peak: 6 turns) | Planner 4.0, Coder 3.7, Tester 4.0, Reviewer 4.0
- **Test-to-Production Code Ratio:** 0.6:1 (+1,233 test lines / +2,159 prod lines)
- **Code Delivered:** +3,392 / -394 lines across 65 code files (excluding docs)
- **Reasoning Effort:** 107,962 reasoning tokens (26.6% of generated output)
- **Tool Health & Self-Correction:** 636 tool calls (47 self-corrections, 92.6% success rate)

### Anomalies & Health Flags
- **Status:** 21 execution anomalies detected (11 CRITICAL, 10 WARNING).
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
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (164.6k tokens) — Context tokens reached 164.6k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.2k tokens) — Context tokens reached 163.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.2k tokens) — Context tokens reached 162.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165.8k tokens) — Context tokens reached 165.8k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.9k tokens) — Context tokens reached 160.9k, breaching the 160k warning ceiling.
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
| **planner** | 1 | 154,972 | 19,944 | 3,794,944 | 96.1% | 163.0k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 3 | 693,451 | 99,425 | 38,892,160 | 98.2% | 271.9k | BLOATED (136%) | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **tester** | 1 | 279,731 | 59,688 | 32,765,312 | 99.2% | 329.4k | BLOATED (165%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 1 | 441,465 | 24,204 | 14,313,728 | 97.0% | 298.9k | BLOATED (149%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **rootagent** | 1 | 240,064 | 94,040 | 47,358,976 | 99.5% | 285.0k | BLOATED (143%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **7** | **1,809,683** | **297,301** | **137,125,120** | **98.7%** | **329.4k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 7 | 1,809,683 | 297,301 | 137,125,120 | 98.7% | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **Total** | **7** | **1,809,683** | **297,301** | **137,125,120** | **98.7%** | **Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations)** |

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
