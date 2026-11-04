# AL-76 work summary

Shared lossless ledger formatting and the selective `format` command are delivered. [Contract](CONTRACT.md), [execution and steering](iterations/01/PLAN.md), [compact review findings](iterations/01/REVIEW_SUMMARY.json).

## Implementation

Serialization adapters preserve scalar lexemes, member/array order, unknown fields, nulls and absent fields. Recognized tool/token/evidence/anomaly records compact onto one line; expanded structure uses two spaces and one trailing newline. Fact, computed, backfill and migration writers share the policy.

Filesystem/CLI supports all, single and explicit ticket-set selection, configured/explicit roots, dry-run/check and deterministic results. Preflight parses all targets under canonical ownership before atomic replacement. Unsafe physical paths, selected hard-link aliases and pending journals are rejected; supported symbolic aliases, permissions and retired sidecars are preserved. Backfill/migration own each target before baseline read through publication; computed exports own their output without nested scanner locks. Domain/frontend behavior and the release pin were not changed by AL-76. Only disposable fixtures were formatted.

## Review and fixes

All four focuses ran gpt-6.1-sol/high and gpt-6-luna/medium. Sol behavior, data and architecture independently discovered PUB-01 (severity 8): formatting could overwrite newer migration/backfill publications. Behavior and architecture independently reproduced schema rollback; data traced the same interleaving. Coder gpt-6.1-sol/high fixed ownership; tester gpt-6.1-sol/high added real contention regressions. All three Sol correction re-reviews approved product `c32c8937d8897179a6772c6aa75300eafe94351f`; Luna's independent product reviews approved it without further findings.

Sol test-effectiveness discovered TEST-01 (severity 4, missing late-I/O partial-result oracle) and TEST-02 (severity 3, missing computed anomaly layout oracle). Tester gpt-6.1-sol/high added both; Sol re-review and independent Luna test review approved `0dedb33110bafe2fc290ec99b86d3635b07c81b5`. Parent orchestrator gpt-6.1-sol/high caught PATH-01 (severity 3, slash-key structural collision) before formal review; coder escaped path segments and tester added coverage. Discovery credit is separate from verification. Exact models/efforts were verified in local Codex source turn metadata at final delivery; Luna dispatch also acknowledged its settings.

## Integrated verification and approval

Initial tree clean at `122e983f8c365a2aa248b1b6d492252555e922ae`. One clean merge incorporated main `8b1de5f761ebb92f78b2b462c1ff583724ecbdb5` at tested revision `3bccb9af8f4af6ee0a7c570ec94cbf1d05344f70`, preserving AL-60 capture repairs/fixture barriers, AL-61 historical timestamps, predecessor index removals and current shared frontend assets. AL-76 source/tests and its physical-path/lock/journal dependencies were unchanged by integration. All eight approvals carry forward with that dependency rationale; no additional product corrections were needed.

Full `make` PASS on integrated revision: static checks, workflow/install/distribution tests, all Go race packages, 36 frontend files / 790 tests, frontend build and five-platform binaries. CLI race package: 106.174s; capture: 67.065s. No skipped stages. Source and embed asset trees match; built host binary contains current JS/CSS assets. Native command help and disposable dry-run/check/write/clean-check plus large-integer/exponent preservation passed. Earlier focused evidence is retained in [TESTING.md](iterations/01/TESTING.md).

Earlier capture installer/worker failures also occurred on manager's unchanged baseline; AL-60 repaired them. The integrated gate now passes. No unrelated new bug found or routed. Browser/runtime release hosting and full Windows execution were not exercised; five-platform compilation passed. Final publication changes are ticket documentation/lifecycle/telemetry only; reviewed/tested application source remains identical.

Parent consolidated local verdict: APPROVED after integrated make and revision-aware carry-forward. All reported findings fixed; no remaining review gap. Detailed completed reports compacted only after approval. MGT-2 owns remote merging; local approval does not authorize this owner to merge.

## Telemetry and delivery

Final source extraction uses explicit Codex roots: native root `01a1a283-edc6-7820-9bd5-969709b32a26` plus seven source-linked children, and four independently recorded delegated Luna roots. Source does not encode those four as native children; retain their actual root facts without inventing parent links. No prior ticket ledger existed; no other-harness sessions were present to preserve. Final source/count/limits are verified after extraction below. Native CLI uses supported `generate` flags; wrapper-only `--al-offline/--al-project` are unavailable in this checkout. Raw source transcripts remain outside Git.

PR: https://github.com/NiltonDuarte/AgenticLens/pull/102 . MGT-2 receives exact final head and verification state directly. Final hosted checks remain visible to the merge owner.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 18
- **Active Agent Duration:** 1h 12m (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 163.6k tokens (36% of 450k healthy ceiling, 0 compactions) | Subagents peak 121.4k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.9 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 2.0, Tester 2.0, Reviewer-architecture 2.0, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 0.9:1 (+739 test lines / +816 prod lines)
- **Code Delivered:** +1,555 / -123 lines across 14 code files (excluding docs)
- **Reasoning Effort:** 23,890 reasoning tokens (17.6% of generated output)
- **Tool Health & Self-Correction:** 348 tool calls (26 self-corrections, 92.5% success rate)

### Anomalies & Health Flags
- **Status:** 5 execution anomalies detected (5 WARNING).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+68.3k tokens) — Context tokens surged by 68.3k to 68.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.7k tokens) — Context tokens reached 160.7k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 71,294 | 3,925 | 541,184 | 88.4% | 73.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 84,435 | 19,570 | 4,028,800 | 97.9% | 121.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 85,781 | 10,290 | 2,591,232 | 96.8% | 116.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 5 | 416,853 | 46,047 | 18,464,256 | 97.8% | 163.6k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-architecture** | 1 | 94,855 | 8,264 | 1,467,648 | 93.9% | 111.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 100,366 | 10,478 | 1,739,136 | 94.5% | 105.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 68,985 | 6,246 | 1,102,464 | 94.1% | 84.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 71,716 | 6,740 | 1,174,144 | 94.2% | 87.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **12** | **994,285** | **111,560** | **31,108,864** | **96.9%** | **163.6k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 257,874 | 14,548 | 1,554,816 | 85.8% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 8 | 736,411 | 97,012 | 29,554,048 | 97.6% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **12** | **994,285** | **111,560** | **31,108,864** | **96.9%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |
