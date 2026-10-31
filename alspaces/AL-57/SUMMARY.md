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
- **Software Development Life Cycle Iterations:** 5
- **Active Agent Duration:** 51m 17s (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 152.4k tokens (76% of 200k healthy ceiling, 0 compactions) | Subagents peak 213.6k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.2 avg turns/run (peak: 4 turns) | Planner 2.0, Coder 3.5, Tester 1.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 0.4:1 (+549 test lines / +1,333 prod lines)
- **Code Delivered:** +1,882 / -273 lines across 56 code files (excluding docs)
- **Reasoning Effort:** 33,155 reasoning tokens (21.1% of generated output)
- **Tool Health & Self-Correction:** 319 tool calls (26 self-corrections, 91.8% success rate)

### Anomalies & Health Flags
- **Status:** 14 execution anomalies detected (5 CRITICAL, 9 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.5k tokens) — Context tokens reached 200.5k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+113.9k tokens) — Context tokens surged by 113.9k to 113.9k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (201.3k tokens) — Context tokens reached 201.3k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.4k tokens) — Context tokens reached 200.4k, breaching the 200k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (15 cycles) — Subagent roles 'coder' and 'tester' alternated for 15 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (164.6k tokens) — Context tokens reached 164.6k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.2k tokens) — Context tokens reached 163.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.2k tokens) — Context tokens reached 162.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165.8k tokens) — Context tokens reached 165.8k, breaching the 160k warning ceiling.
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
| **planner** | 1 | 113,111 | 13,795 | 2,048,768 | 94.8% | 122.8k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 423,698 | 51,015 | 20,882,816 | 98.0% | 213.6k | BLOATED (107%) | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 180,474 | 19,993 | 10,414,208 | 98.3% | 204.1k | BLOATED (102%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 1 | 135,124 | 9,258 | 5,226,240 | 97.5% | 210.1k | BLOATED (105%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **rootagent** | 1 | 144,823 | 30,099 | 13,728,384 | 99.0% | 152.4k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **6** | **997,230** | **124,160** | **52,300,416** | **98.1%** | **213.6k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 6 | 997,230 | 124,160 | 52,300,416 | 98.1% | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **Total** | **6** | **997,230** | **124,160** | **52,300,416** | **98.1%** | **Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations)** |


Final telemetry verification: generated schema-v5 ledger contains exactly the real Codex root and all five expected children (six sessions), with correct parent/source labels, nonempty usage/tool records and committed code facts. Fresh Codex costs are explicitly unavailable, not estimated or free. No preexisting AL-57 ledger or other-harness sessions existed to preserve. Raw source rollouts remain outside Git.
