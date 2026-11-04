# AL-77 Work Summary

## Delivered behavior and authority

Fleet Overview inspects the unfiltered loaded project ledger dataset for exact session IDs recorded in distinct tickets. Expand the warning for complete session IDs, affected tickets, source context and existing ticket navigation. Same-ticket repetitions do not create conflicts. Diagnostic/omitted scope is partial; failed inspection is unavailable. Refresh replaces results, clears resolved conflicts, and ignores obsolete completions. Metrics filters and inspection state remain independent; reads do not repair, reassign, delete or deduplicate ledger data.

The user approved [CONTRACT.md](CONTRACT.md) and [iteration 01](iterations/01/PLAN.md), then explicitly instructed: "Do not fill missing id from directory name. Ticket id is a required field on the data". [Iteration 02](iterations/02/PLAN.md) records that authoritative delta. Persisted `ticket_id` is required and nonblank; invalid ledgers report diagnostics while valid siblings remain available. Both directory/request fallbacks were removed. The proposed recorded_ticket_id metadata solution was never implemented. No schema bump, new endpoint, source-authenticity prerequisite or path-equality rule was introduced.

## Baseline and layers

Initial revision `a84adbe5be46c2ef52ede0987a417562715d38e1`; clean working tree on `AL-77/warn-duplicate-sessions-across-tickets`. Feature branch `feat/AL-77-warn-duplicate-sessions`; implementation commit `fe6d71b8`. Configuration selects workspace root `alspaces`. Upstream fetch/merge reported already up to date; no integration source changes invalidate the review.

- Domain/presentation: pure readonly TypeScript overlap index; accessible native disclosure and named ticket controls; Fleet Overview owns independent loading and refresh. Existing Go aggregation policy and valid ledger totals are unchanged.
- Adapters: canonical DTO validation requires recorded ticket identity; scan/direct-load inference is deleted. Source-unavailable persisted facts remain usable. Rejected files remain untouched and valid siblings retain totals.
- Documentation: canonical schema/runtime guidance in docs/OWNED_OCCURRENCES.md describes required identity. Specification moved to plan/done and its active index entry removed. No downstream active ticket depended on AL-77.

## Review outcome and attribution

Final consolidated outcome: APPROVED for candidate `AL77-candidate-2-3871cfb9066dccfb`, content SHA-256 `3871cfb9066dccfba4b962b5193f41376f5dba61defaf269325878c22b4da1a8`, delivered in `fe6d71b8`. All ten persona/model cells completed: behavior, data/failure, architecture, test effectiveness and generic reviewer with frontend async/accessibility specialty, each gpt-6.1-sol/high and gpt-6-luna/medium. Source rollout turn_context records independently confirmed all requested model/effort settings. Nine corrected-candidate child verdicts were APPROVED; one architecture CHANGES_REQUESTED candidate was refuted by parent evidence. Parent approval follows successful integrated checks, not child verdict counts.

- [Iteration 01 compact findings](iterations/01/REVIEW_SUMMARY.json): BEH-01, severity 6, false membership from inferred ticket IDs, independently discovered by behavior reviewers gpt-6.1-sol/high and gpt-6-luna/medium plus data/failure reviewer gpt-6.1-sol/high. Fixed by coder gpt-6.1-sol/high through required-ID validation and deletion of both fallbacks under user authority. Tester gpt-6.1-sol/high verified real boundaries and a guard-removal mutation; both behavior profiles and both data profiles verified the corrected paths. Original three CHANGES_REQUESTED runs retain shared discovery credit; seven queued original cells were superseded before dispatch and received corrected-candidate coverage.
- [Iteration 02 compact findings](iterations/02/REVIEW_SUMMARY.json): ARCH-01, claimed severity 5 cleanup bypass, reported by architecture reviewer gpt-6-luna/medium. Parent gpt-6.1-sol/high refuted it: runtime validation reaches canonical DTO.Validate before cleanup publication. A disposable overlay probe rejected absent/null/empty/whitespace IDs and accepted a valid completed-migration control; no product fix was needed.
- NAV-01, severity 4, preexisting detail-navigation limitation when recorded ticket ID differs from directory: independently reported by behavior and data/failure reviewers gpt-6.1-sol/high. Repository maintainers own separate identity-to-location follow-up; existing navigation is reused and no new equality rule is imposed.
- TEST-01, severity 2, unchanged nested-session rejection fixture masked by an earlier schema error: reported by test-effectiveness reviewer gpt-6.1-sol/high. Repository maintainers own separate fixture/oracle maintenance. It does not undermine the focused AL-77 validation oracles.

No unresolved introduced finding or material review gap remains. Preexisting accepted debt stays explicit in compact records. Private active review reports/run records were compacted only after final approval; raw source transcripts remain outside Git.

## Verification and limits

Coder and tester gpt-6.1-sol/high ran focused frontend/domain checks (23 tests), full frontend checks, typecheck, lint, production build and relevant schema/repository/HTTP race checks. Full backend `go test -race ./...` passed. Existing positive schema fixtures received explicit IDs without weakening their original error oracles. Disposable removal of only the required-ID guard made schema, direct-load and HTTP regression tests fail. Test fixtures are self-contained; no live ticket history or developer source database is used.

Parent gpt-6.1-sol/high ran integrated `make` successfully on `fe6d71b8` after upstream integration: dependency install, vet/lint/typecheck, workflow/installation/capture scripts, backend race suite, frontend 833/833 tests across 37 files, canonical frontend build and all configured platform binary builds. Existing MetricProgression React act warnings and the Vite chunk-size warning remain nonfatal. Frontend generated trees are ignored and were not committed.

Standalone host binary reports commit `fe6d71b8`. Parent temporary HTTP probe verified all three generated files (index, CSS, JavaScript) match web/dist and embedded dist byte-for-byte, bundle contains the shared-session disclosure, and SPA fallback serves the same current index. Process was stopped after verification. An initial probe incorrectly passed wrapper-only --al-offline/--al-project flags to the native binary; actual native help established the correct invocation and the corrected probe passed. No application change was needed.

Real browser, assistive technology and narrow-viewport geometry were not exercised. Native details/summary semantics and wrapping were inspected; jsdom tests verify focus, click disclosure, and Enter/Space ticket activation. StrictMode/unmount safety is source-based request-generation evidence, not a claimed browser experiment.

## Handoff correction

User requested relative file links instead of GitHub blob URLs. Restored relative contract, plan and summary links in PULL_REQUEST.md and the published PR description. This is a documentation-only correction within iteration 02; implementation and prior verification remain unchanged. No new plan or application checks are needed. Whitespace validation passed; telemetry was refreshed while preserving all 18 source sessions.

## Telemetry

Final ingestion uses source-grounded explicit selection. Source inventory identifies root `01a1a7d9-4a6e-71b1-8d22-746dc07621b3`, four native children and thirteen independent T3 review roots (mapped to exact run prompts and observed model/effort). Independent review roots have no native parent linkage; all thirteen were ingested explicitly without fabricating one. Root-tree extraction ran last and verification confirmed exact equality with all 18 inventoried source session IDs: the root, four native children with their original parent links, and thirteen independent review roots without invented parents. Every recorded source is codex; every supplemental review session survived final root extraction. No preexisting AL-77 ledger or other-harness session existed to preserve. The built native CLI is used directly from the project directory; --al-offline/--al-project are wrapper flags, and no installed wrapper was available on PATH. Supported native --source codex, --session-id, --nonce and --ticket-type feat were confirmed with generate --help. Initial extraction stopped before publication because the configured local capture coordinate layout lacks its worktree mapping table. Existing state was preserved; explicit --capture-dir selects a new private temporary registry for this manual source-backed extraction. The CLI reports that configured-state limitation on each run; no ledger or successful-empty source result was manufactured. Supplemental review roots use supported explicit --session-id because a nonce in this parent transcript would select the parent rather than those independent roots. The final parent-tree extraction used a literal nonce. All 18 invocation costs remain explicitly unavailable; no pricing or role identity was invented. The rendered capture-observation footer covers the thirteen registry-bound review roots, while the canonical ledger additionally includes the five manually extracted parent-tree sessions. It is not a common as-of snapshot or proof that all eighteen owners were registry-bound. Native role metadata identifies two planners, one coder and one tester; the parent and thirteen T3 roots have no native role metadata and remain Unknown in rendered metrics. The table predates final commit/push/PR bookkeeping by design.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 43
- **Active Agent Duration:** 1h 27m (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 219.8k tokens (49% of 450k healthy ceiling, 0 compactions) | Subagents peak 144.8k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.8 avg turns/run (peak: 2 turns) | Planner 1.5, Coder 2.0, Tester 2.0
- **Test-to-Production Code Ratio:** 1.6:1 (+386 test lines / +241 prod lines)
- **Code Delivered:** +627 / -68 lines across 11 code files (excluding docs)
- **Reasoning Effort:** 33,315 reasoning tokens (19.1% of generated output)
- **Tool Health & Self-Correction:** 397 tool calls (19 self-corrections, 95.2% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+106.1k tokens) — Context tokens surged by 106.1k to 106.1k tokens.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 142,257 | 11,721 | 2,133,376 | 93.7% | 127.7k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **coder** | 1 | 108,083 | 18,655 | 5,515,392 | 98.1% | 143.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 202,967 | 9,686 | 3,130,112 | 93.9% | 144.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 14 | 1,032,905 | 101,023 | 26,869,248 | 96.3% | 219.8k | HEALTHY | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **Total** | **18** | **1,486,212** | **141,085** | **37,648,128** | **96.2%** | **219.8k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 6 | 308,178 | 22,795 | 2,267,392 | 88.0% | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **gpt-6.1-sol** | 12 | 1,178,034 | 118,290 | 35,380,736 | 96.8% | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **Total** | **18** | **1,486,212** | **141,085** | **37,648,128** | **96.2%** | **Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations)** |

Capture observation finalized: 2026-11-04T20:29:27.32653253Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 18; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 18; reserved bytes: 17040741. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1a813-81ce-77c3-bd6e-0312d173487c: current; last successful source read 2026-11-04T20:29:27.085650511Z; accepted source checkpoint 2026-11-04T19:42:16.108Z.
Owner codex/01a1a813-87f1-7293-b34f-0dcfd86b7ad4: current; last successful source read 2026-11-04T20:29:27.110809113Z; accepted source checkpoint 2026-11-04T19:42:29.751Z.
Owner codex/01a1a7ec-e048-7742-8033-1b440046bbf9: current; last successful source read 2026-11-04T20:29:27.124391633Z; accepted source checkpoint 2026-11-04T18:59:30.801Z.
Owner codex/01a1a7e5-1b59-7f22-9b21-91f2fd14bef6: current; last successful source read 2026-11-04T20:29:27.152330016Z; accepted source checkpoint 2026-11-04T19:37:28.151Z.
Owner codex/01a1a81a-a3bb-7c93-aa4e-2440c372b161: current; last successful source read 2026-11-04T20:29:27.171957631Z; accepted source checkpoint 2026-11-04T19:49:48.921Z.
Owner codex/01a1a818-8fe0-7b12-ac4d-7c4829c57bc0: current; last successful source read 2026-11-04T20:29:27.18614525Z; accepted source checkpoint 2026-11-04T19:47:07.215Z.
Owner codex/01a1a816-ea2f-73a1-b63a-3baf390543b6: current; last successful source read 2026-11-04T20:29:27.207010649Z; accepted source checkpoint 2026-11-04T19:45:59.106Z.
Owner codex/01a1a7d9-b373-7271-8835-553129325ce8: current; last successful source read 2026-11-04T20:29:27.222648385Z; accepted source checkpoint 2026-11-04T18:39:40.258Z.
Owner codex/01a1a7ec-dd54-7532-8b02-2503011e8341: current; last successful source read 2026-11-04T20:29:27.239471485Z; accepted source checkpoint 2026-11-04T19:00:20.587Z.
Owner codex/01a1a815-a44c-7572-a847-6c2f8bb4d530: current; last successful source read 2026-11-04T20:29:26.849382536Z; accepted source checkpoint 2026-11-04T19:43:29.783Z.
Owner codex/01a1a816-58e4-7010-840d-a610eed6b2cc: current; last successful source read 2026-11-04T20:29:26.870862548Z; accepted source checkpoint 2026-11-04T19:45:17.865Z.
Owner codex/01a1a81a-a722-7ea0-aa17-4e4cef3bbe17: current; last successful source read 2026-11-04T20:29:26.885163764Z; accepted source checkpoint 2026-11-04T19:49:29.978Z.
Owner codex/01a1a813-8493-7fc0-954a-5bd21da98ca3: current; last successful source read 2026-11-04T20:29:26.905766616Z; accepted source checkpoint 2026-11-04T19:41:56.461Z.
Owner codex/01a1a7d9-4a6e-71b1-8d22-746dc07621b3: current; last successful source read 2026-11-04T20:29:26.977745669Z; accepted source checkpoint 2026-11-04T20:29:26.783Z.
Owner codex/01a1a7e8-f267-78b1-b3a7-789b1f929f70: current; last successful source read 2026-11-04T20:29:27.006423148Z; accepted source checkpoint 2026-11-04T19:39:15.581Z.
Owner codex/01a1a816-5b56-7092-890b-40e2d72e4905: current; last successful source read 2026-11-04T20:29:27.02599005Z; accepted source checkpoint 2026-11-04T19:45:19.215Z.
Additional owner freshness details remain in local capture diagnostics.
