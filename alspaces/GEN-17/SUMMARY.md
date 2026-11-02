# GEN-17: Preserve source session personas

## Delivered behavior

- Domain: retain the unknown role bucket in ordering, role/thread totals, effort and peaks; structural root/child calculations continue to use parent_id. Shared pure persona evidence preserves explicit metadata and rejects conflicting assistant candidates.
- Adapters: Codex supports root agent_path fallback; OpenCode v1/v2 recover only unambiguous assistant personas. New facts synthesize neither RootAgent nor UnassignedSubAgent. Successful source hydration projects recovered roles onto matched events through an explicit source-only reconciler. Failed metadata reads preserve committed OpenCode facts and independent Codex enrichment. Live trace and persona evidence share one immutable cached snapshot.
- Presentation: missing roles display Unknown in dashboard, trace filters, metrics and CLI summaries. Existing known ledger personas remain when optional live evidence adds no persona; explicit source evidence overrides them. Historical labels remain readable. No new persisted root field, schema bump or bulk ledger rewrite.

## Scope, ownership and ticket correction

User approval: "apply this recommendation" approved Unknown for missing roles in roots and children, expanding the proposed baseline's child fallback. [Iteration 02](iterations/02/PLAN.md) controls that delta; CONTRACT.md retains the proposed baseline. Planner authored the plans, coder owned production implementation, tester owned regression changes, independent reviewers owned scoped investigation, and the parent owned integration and delivery.

Initial Git status was clean at 99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e on generated branch GEN-13/root-session-role. Initial local workspace/index/reference checks allocated GEN-16. User later explicitly corrected this: "GEN-16 was taken. change the workspace to GEN-17 and merge main into the branch". The same work moved to alspaces/GEN-17 and chore/GEN-17-preserve-root-session-role; no application/test changes were required for relocation. Historical finding IDs retain GEN16 as provenance. Checkpoint b8c087a committed owned work; merge 124419d integrated origin/main 8f0fd4f cleanly. Incoming independent GEN-16 artifacts were preserved. No matching GEN-17 planning specification or index entry required lifecycle updates.

## Consolidated review: APPROVED

Reviewed final content: integrated 124419da3aadd170b9fe6e15e0d7bdcc1ebef200 plus backend/web diff fingerprint 3b083d30617cefbe537a8838c4fc441881988aff7f2588fcbaa5d60916260f60. The fingerprint is SHA-256 of git diff HEAD --binary restricted to backend/web followed by sorted untracked source path, NUL and bytes. Application/test content was frozen through the final reviews. Integration changed guidance only. Sol behavior/architecture correction approvals carry forward across the final failclosed metadata branch; final data/test investigations and targeted controls cover that branch.

[Compact findings and every reporting persona/model/effort tuple](iterations/02/REVIEW_SUMMARY.json) permanently retain all six candidates, including shared discovery, refutation and accepted debt. Model/effort attribution below is dispatch-acknowledged configuration, not independently source-verified identity.

| Finding | Discovering responsibility and model/effort | Disposition and verification |
| --- | --- | --- |
| GEN16-01 | Behavior, data/failure and test-effectiveness reviewers, each gpt-6.1-sol/high | Matched events discarded recovered persona. Coder gpt-6.1-sol/high added explicit source reconciliation; tester gpt-6.1-sol/high added stable-ID, unique legacy and appended controls. Independent pre-fix probes failed; parent and reviewers verified corrected behavior. |
| GEN16-02 | Architecture reviewer, gpt-6.1-sol/high | Split cache reads could lose a stable persona. Coder gpt-6.1-sol/high paired trace/evidence snapshots; tester gpt-6.1-sol/high added deterministic real-SQLite invalidation guard. Parent bounded concurrent race probe and focused reviewers verified correction. |
| GEN16-03 | Data/failure reviewer, gpt-6-luna/medium | Confirmed pre-existing global invalidation cost. Parent verified identical baseline behavior; accepted debt owned by live-service maintenance for future per-session invalidation work. No new history query introduced. |
| GEN16-04 | Test-effectiveness reviewer, gpt-6-luna/medium | Test gap only. Tester gpt-6.1-sol/high extended the existing orphan builder test to assert original parent_id and input ancestry. Focused reviewers verified the oracle; no production defect claimed. |
| GEN17-05 | Data/failure re-reviewer, gpt-6.1-sol/high | Failed metadata lookup erased an event-only recorded role. Coder gpt-6.1-sol/high now fails closed for OpenCode; tester gpt-6.1-sol/high added partial-failure and mixed-source controls. Pre-fix overlay failed three cases; corrected races and final data/test reviews passed. This distinct defect was discovered during re-review. |
| GEN17-06 | Data/failure re-reviewer, gpt-6-luna/medium | Refuted interpretation: absent optional persona evidence does not prove a known recorded role stale. Existing persisted-planner control and baseline-preservation contract support retention; final test review agreed. No product change. |

Initial eight discovery runs completed: all four Sol scopes requested changes; Luna behavior/architecture approved, data raised existing performance debt, tests identified the ancestry assertion gap. Re-review verified corrections under both profiles. Final behavior, architecture and test scopes approve under both profiles; final Sol data approves. Luna data's completed dissent is retained and independently refuted as above. Parent consolidated APPROVED only after integrated make and standalone checks passed; this does not authorize merging the PR. Verification actors are separate from discovering reporters; no supplied-finding verification is counted as independent discovery.

## Verification

- Integrated make passed on the final reviewed content: all 12 Go packages under the race detector; frontend 35 files / 783 tests; typecheck and ESLint; Python release and installation checks; canonical frontend build and Linux/macOS amd64/arm64 builds. Log: /tmp/gen17-make.log (local evidence, not a committed transcript).
- Parent verified frontend and Go embed output trees byte-for-byte, then ran the Linux/amd64 standalone binary with Node absent from PATH in a disposable workspace. Current index, JavaScript and CSS served byte-identically; server terminated. Generated assets remain ignored per policy.
- Targeted regression and before/after evidence: [TEST_REPORT.md](iterations/02/TEST_REPORT.md). Final metadata-failure and mixed-source controls preserve ledger bytes, accepted facts and other-source multiplicity. Legacy labels remain readable with enrichment unavailable. git diff --check passed.
- The original 2000-poll race/count3 stress attempt timed out after 40 seconds and is not a pass. A bounded 20-poll race/count1 copy passed in 1.889 seconds; the durable snapshot test uses deterministic interleaving. Browser geometry and actual external-session persona recovery were not exercised.

## Remaining limits and telemetry

Database-wide invalidation can reread histories of unchanged active sessions after unrelated writes; idle polls reuse snapshots. This is existing accepted performance debt, not changed-session-only polling. Pricing, cost, source identity and session completeness must come from final extraction, not reviewer assignments. Final telemetry is generated after all review/integration work; extraction facts and limits are recorded below when available.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 18
- **Active Agent Duration:** 1h 38m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 212.3k tokens (106% of 200k healthy ceiling, 0 compactions) | Subagents peak 207.6k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.5 avg turns/run (peak: 4 turns) | Planner 2.0, Coder 4.0, Tester 4.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 3.1:1 (+622 test lines / +202 prod lines)
- **Code Delivered:** +824 / -176 lines across 34 code files (excluding docs)
- **Reasoning Effort:** 41,527 reasoning tokens (19.6% of generated output)
- **Tool Health & Self-Correction:** 524 tool calls (35 self-corrections, 93.3% success rate)

### Anomalies & Health Flags
- **Status:** 12 execution anomalies detected (6 CRITICAL, 6 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.5k tokens) — Context tokens reached 200.5k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.3k tokens) — Context tokens reached 200.3k, breaching the 200k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (8 cycles) — Subagent roles 'coder' and 'tester' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (5 cycles) — Subagent roles 'coder' and 'tester' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (6 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 6 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.9k tokens) — Context tokens reached 160.9k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 4 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 70,102 | 6,506 | 1,125,120 | 94.1% | 73.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 170,459 | 16,619 | 7,678,336 | 97.8% | 177.8k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 142,308 | 32,229 | 11,555,072 | 98.8% | 207.6k | BLOATED (104%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 185,333 | 61,990 | 28,846,848 | 99.4% | 212.3k | BLOATED (106%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 2 | 127,728 | 10,404 | 1,957,120 | 93.9% | 95.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 207,494 | 10,717 | 1,886,592 | 90.1% | 102.9k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 3 | 227,499 | 18,100 | 3,101,696 | 93.2% | 91.3k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-tests** | 3 | 187,194 | 13,542 | 2,163,200 | 92.0% | 88.1k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Total** | **14** | **1,318,117** | **170,107** | **58,313,984** | **97.8%** | **212.3k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 14 | 1,318,117 | 170,107 | 58,313,984 | 97.8% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **Total** | **14** | **1,318,117** | **170,107** | **58,313,984** | **97.8%** | **Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations)** |

## Final telemetry validation and delivery revision

Implementation commit 6215fad contains exactly the reviewed application/test content (124419d plus frozen final diff); a subsequent fetch/merge confirmed origin/main already up to date. Final extraction used the built CLI with literal nonce 5b6dcad992b26d86, explicit --source codex and --ticket-type chore. No previous GEN-17 ledger existed, so no pre-existing other-harness sessions were replaced.

Verified raw source identity: root 01a19e15-72d0-7c60-b25f-3289182eeb8c, one matching default Codex rollout, source metadata origin vscode, no agent_role/agent_path/parent. Ledger source is codex. Its blank persona therefore correctly displays Unknown. Ledger contains the root plus all 13 expected native child sessions, each with the correct parent_id: planner, coder, tester, four initial Sol reviewers, four Sol correction reviewers, and two final Sol reviewers. Source roles remain planner/coder/tester and the exact focused reviewer personas; no synthetic root/child role was assigned.

Telemetry is explicitly incomplete for the eight T3-delegated Luna runs: their T3 executions and model/effort were acknowledged, but those runs are not in the extracted Codex parent tree. Their costs/tokens are not represented by this ledger and no ancestry or session facts were fabricated. All 14 captured invocations report cost evidence unavailable, not measured zero. Rendered durations, cache rates and context health describe captured sessions only. The root peak was ELEVATED and the tester peak BLOATED in this snapshot. Final extraction occurred after all review, corrections and integrated checks; the short validation/commit/push/PR handoff tail naturally follows it. Raw transcripts remain outside Git.

## Naming follow-up

User steering: "I prefer role_resolution" after the filename recommendation. Renamed backend/internal/domain/persona.go to role_resolution.go only; file bytes, symbols, resolution rules and accepted runtime boundaries are unchanged. This is a mechanical follow-up within iteration 02, explicitly approved by the user; no new plan or tests are needed. Coder gpt-6.1-sol/high performed the rename, and the parent verified byte identity against HEAD (SHA-256 ea8b18276da992e90478e46f16c7a4a8d7bd9d1aab882313d9d126e814f02609). Prior substantive review approvals carry forward because no implementation bytes or call paths changed. No new reviewer focus/profile runs are selected for this filename-only exception; parent rename verification and the required integrated make gate cover this follow-up. Final telemetry will refresh the existing ledger, preserving prior sessions and explicit delegated-run/cost limits.

Rename verification completed at369e9e3: latest origin/main already up to date; full make passed (12 Go race packages, frontend35files783tests, static/Python checks, frontend and four-platform builds). Log /tmp/gen17-rename-make.log. The prior standalone asset proof remains valid for unchanged implementation/frontend bytes. PR96 remained OPEN before follow-up delivery. This follow-up changes the filename only; no new review finding or acceptance boundary was introduced. Telemetry refreshed with literal nonce6d27469e29b514a2 after these checks; the prior ledger is retained outside Git for session-preservation comparison.

Rename telemetry validation: all14existingCodex sessions retained, same verified root and13nativechildren. The eight delegated Luna runs remain excluded and costs remain unavailable; prior telemetry limitations still apply.

## Final session refresh

User accepted the current role-resolution behavior ("ok, its fine") and requested metrics generation, commit and push. No application changes or new acceptance boundaries were introduced by the explanatory discussion. Refreshing the existing Codex ledger with literal nonce17d240a0483436fa and conventional type chore, preserving prior captured sessions. Prior integrated make and standalone verification remain applicable to unchanged application content. Existing delegated-run coverage and unavailable-cost limits remain explicit.

Final refresh validation: same verified Codex root and13nativechildren; all14previous sessions retained. The eight T3-delegated Luna runs remain outside this source parent tree, and captured cost evidence remains unavailable. No application/test changes or additional verification claims.
