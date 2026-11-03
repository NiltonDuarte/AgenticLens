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
- **Active Agent Duration:** 1h 4m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 155.9k tokens (78% of 200k healthy ceiling, 0 compactions) | Subagents peak 121.4k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.9 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 2.0, Tester 2.0, Reviewer-architecture 2.0, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 0.9:1 (+739 test lines / +816 prod lines)
- **Code Delivered:** +1,555 / -123 lines across 14 code files (excluding docs)
- **Reasoning Effort:** 22,398 reasoning tokens (17.2% of generated output)
- **Tool Health & Self-Correction:** 341 tool calls (25 self-corrections, 92.7% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (3 WARNING).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+68.3k tokens) — Context tokens surged by 68.3k to 68.3k tokens.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 71,294 | 3,925 | 541,184 | 88.4% | 73.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 84,435 | 19,570 | 4,028,800 | 97.9% | 121.4k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 85,781 | 10,290 | 2,591,232 | 96.8% | 116.0k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 5 | 411,299 | 42,187 | 16,865,152 | 97.6% | 155.9k | ELEVATED | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-architecture** | 1 | 94,855 | 8,264 | 1,467,648 | 93.9% | 111.9k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 100,366 | 10,478 | 1,739,136 | 94.5% | 105.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 68,985 | 6,246 | 1,102,464 | 94.1% | 84.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 71,716 | 6,740 | 1,174,144 | 94.2% | 87.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **12** | **988,731** | **107,700** | **29,509,760** | **96.8%** | **155.9k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 257,874 | 14,548 | 1,554,816 | 85.8% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 8 | 730,857 | 93,152 | 27,954,944 | 97.5% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **12** | **988,731** | **107,700** | **29,509,760** | **96.8%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

Capture observation finalized: 2026-11-03T21:08:17.541861905Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 4; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 4; reserved bytes: 784733. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1a296-5708-7ae2-8133-b71616d928dc: current; last successful source read 2026-11-03T21:08:17.529384192Z; accepted source checkpoint 2026-11-03T18:06:21.913Z.
Owner codex/01a1a294-5379-7c82-90f1-1805be0b5494: current; last successful source read 2026-11-03T21:08:17.535329518Z; accepted source checkpoint 2026-11-03T18:04:15.325Z.
Owner codex/01a1a297-ee6f-7201-bc87-9e4da76ac22c: current; last successful source read 2026-11-03T21:08:17.517375748Z; accepted source checkpoint 2026-11-03T18:07:44.477Z.
Owner codex/01a1a295-80bb-7e40-9d43-cd37595bb8cd: current; last successful source read 2026-11-03T21:08:17.52390398Z; accepted source checkpoint 2026-11-03T18:05:34.272Z.

Telemetry receipt: verified exact source-ID set of 12 Codex sessions (five actual roots and seven native children), eight gpt-6.1-sol and four gpt-6-luna invocations; each explicit delegated-root ingestion survived final native-root refresh. Native child parent_id values match source metadata. Costs are unavailable for all 12 invocations and were not converted to zero. Root/delegated role metadata is unavailable for five invocations and remains Unknown. Rendered aggregate cache-hit ratio is 96.8%; peak context 155.9k/ELEVATED. Registry advisory coverage concerns its four bound delegated owners; explicit native-root extraction covers the other eight source records. Extraction is a source-read checkpoint and excludes later verification/push/handoff activity; no complete future-session claim. No source failures or other-harness baseline to preserve.
