# AL-47 Work Summary

## Session baseline and approval

- Baseline: `1d20117711aab96a230679d3cdb9406853fe3f04`; initial Git status clean.
- Branch: `feat/AL-47-live-trace-dynamic-refresh-active-session-polling`.
- User approved CONTRACT.md and iteration 01 PLAN.md on 2026-10-31. Implementation stayed within those boundaries.
- Parent orchestrated planner, coder, tester, and reviewer; parent owns integration and delivery.

## Domain

Complete meaningful trace comparison detects in-place content, reasoning, tool, token, child-link, and anomaly changes. Frontend pure helpers identify active statuses and unchanged snapshots.

## Adapters

A stable read-only SQLite observer supplies connection-scoped source generations. Shared bounded trace snapshots eliminate repeated full-history reconstruction on unchanged tree, REST, and SSE reads. Source changes or uncertain observations invalidate reuse; failed reads do not become empty successful snapshots. Existing SSE protocol, nonblocking delivery, configured backend cadence, and source-independent tree status calculation remain intact.

## Presentation

Selected running/active traces recover through a coalesced 2000 ms REST loop alongside SSE. Selection/lifecycle ownership ignores obsolete successes and errors; newer SSE snapshots supersede earlier REST requests. Completion, stall, missing selection, pause, close, and unmount stop automatic refresh. Manual refresh and resume reconcile explicitly. Trace failures retain visible data and survive unrelated successful tree traffic. Stable message identity preserves reasoning choices, and reading anchors preserve position for same-message growth and new turns in both existing orderings.

## Verification

Tester evidence: [iterations/01/TESTING.md](iterations/01/TESTING.md). Full Go race suite passed. Full frontend suite passed 720 tests before two final stopped-polling cases; final focused view/inspector rerun passed 49 tests. Typecheck, lint, focused API tests, and diff checks passed. No tests removed or silently skipped. Existing React act warnings remain in unrelated MetricProgression coverage.

Independent review approved the diff. Implementation commit `f531208` was merged with `origin/main` without conflicts at tested head `6968687122d6146010e4255ae46f3ea0f1b6936b`. Upstream changed planning documents only; its AL-47 clarification matches the accepted scope. Full integrated `make` passed: Go vet/race tests, frontend typecheck/lint, all 34 frontend files and 722 tests, frontend build, and linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 builds. Embedded `TestGetFS` passed within that gate. `diff -qr web/dist backend/internal/adapters/web/dist` passed. The standalone linux/amd64 binary served byte-identical index and JS/CSS over HTTP from an isolated workspace with Node absent from PATH, then shut down cleanly. Built generate help and version JSON were checked. No generated assets were force-added.

The specification moved to plan/done and AL-47 was removed from the ticket index; no downstream ticket declares AL-47 as a blocker. Final Codex telemetry was extracted and verified; PR delivery follows. DOM checks verify controlled reading-anchor arithmetic and state, not browser geometry; deterministic mutation exactly between cache generation observations is untested. Raw source transcripts stay outside Git.


## Telemetry provenance

New ticket ledger: no prior AL-47 ledger or other-harness sessions to preserve. Actual Codex session metadata under the normal ~/.codex/sessions store establishes root `01a191f5-413e-7250-a3bf-0b3f4e5be953` and four direct children: planner `01a191f5-8f30-72b1-8e01-07a9f3fc3825`, coder `01a191fa-d821-7490-9edc-e6a10fdeeb77`, tester `01a19200-2d16-7db3-b0ec-8f788079261f`, reviewer `01a19203-e0c8-7d41-a003-809db2aa2c8e`. Final extraction must retain all five with Codex provenance. Source transcripts are not committed.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 7
- **Active Agent Duration:** 42m 26s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 146.9k tokens (74% of 200k healthy ceiling, 0 compactions) | Subagents peak 153.8k (Reviewer)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0, Coder 1.0, Tester 1.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 17.2:1 (+155 test lines / +9 prod lines)
- **Code Delivered:** +164 / -6 lines across 2 code files (excluding docs)
- **Reasoning Effort:** 11,553 reasoning tokens (14.9% of generated output)
- **Tool Health & Self-Correction:** 241 tool calls (18 self-corrections, 92.5% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+128.6k tokens) — Context tokens surged by 128.6k to 128.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+66.2k tokens) — Context tokens surged by 66.2k to 66.2k tokens.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 171,431 | 8,696 | 1,351,680 | 88.7% | 136.5k | ELEVATED | $0.00 |
| **coder** | 2 | 111,388 | 16,570 | 3,832,320 | 97.2% | 149.6k | ELEVATED | $0.00 |
| **tester** | 2 | 83,543 | 14,115 | 3,571,328 | 97.7% | 151.6k | ELEVATED | $0.00 |
| **reviewer** | 2 | 87,407 | 4,924 | 1,888,640 | 95.6% | 153.8k | ELEVATED | $0.00 |
| **rootagent** | 1 | 144,078 | 21,594 | 11,514,880 | 98.8% | 146.9k | ELEVATED | $0.00 |
| **Total** | **9** | **597,847** | **65,899** | **22,158,848** | **97.4%** | **153.8k** | **ELEVATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 9 | 597,847 | 65,899 | 22,158,848 | 97.4% | $0.00 |
| **Total** | **9** | **597,847** | **65,899** | **22,158,848** | **97.4%** | **$0.00** |

## Final telemetry validation

The built CLI reported five Codex rollouts. The generated ledger was checked against all five expected session IDs, Codex source labels, and four child-to-root links; all matched. There was no pre-existing AL-47 ledger, so no other-harness facts were replaced. Generation used a literal nonce, explicit root ID, source codex, type feat, and ticket title. Generated metrics record 18m 39s active duration, 96.9% cache hit rate, and elevated context health with zero root compactions. The renderer prints $0.00, but the extracted sessions carry no source cost fields; this is unavailable billing evidence, not proof that the run was free. No pricing estimate is asserted. Extraction succeeded without a source error. Raw transcripts remain outside Git.


## Approved startup correction preparation (2026-10-31)

User/MGT-1 routed a confirmed AL-47 SSE startup defect from the AL-53 integration gate. This remains iteration 01 and preserves the original feature contract and evidence. New clean fix branch `fix/AL-47-sse-startup-subscription` starts at fetched main `6bc1064943ae22e1b6586a6d145543c6a9cd23f2`. Original feature PR #76 is merged (5da845a); correction delivery requires a new PR.

The original failed AL-53 gate occurred before SQLite mutation: HTTP initial readTrace timed out. `/tmp/al53-sse-ordering-proof.log` reports “Observed concrete service publish first trace while HTTP initial-tree call is paused before Subscribe”; “HTTP stream subscribed afterward: session_tree=1 trace_update=0 stream_error=context deadline exceeded”; “Confirmed subscription ordering gap: unchanged trace is not replayed to the late subscriber”. The planner read this real-service/race proof and verified the matching source ordering; it has not independently rerun the proof. Passing repeated package tests does not erase the reproduced schedule.

Appended plan directs a minimal HTTP subscriber registration before connection flush/initial tree computation, retaining initial-tree-first frames and deferred cleanup. Deterministic regression will hold the initial lookup while the concrete SQLite-backed service broadcasts, then prove the queued trace is delivered. No application or test edits, verification, or telemetry extraction performed during this planning step. Existing ledger identities must be preserved; manager owns merges and is not a telemetry child/root. Full make and independent review are required before the new fix PR is handed back.


## Parent startup correction integration gate (2026-10-31)

The correction was independently APPROVED and committed as `c72eb25317e8b6317448a08100364940402cd319`, based on current main `6bc1064943ae22e1b6586a6d145543c6a9cd23f2`. Fresh upstream integration was already current; no product changes followed review. Full `make` passed at this actual integrated head: Go vet and all race suites, frontend typecheck/lint, 35 files / 736 frontend tests, canonical frontend assets and four platform binaries. Log: `/tmp/al47-startup-make.log` outside Git. Existing MetricProgression act warnings remain unrelated. No tests were skipped. `diff -qr web/dist backend/internal/adapters/web/dist` passed; embedded `TestGetFS` ran in the full gate. Host version JSON and built generate help passed. A bounded standalone HTTP smoke test served byte-identical index/JS/CSS from isolated configuration with Node absent from PATH, then exited cleanly. Frontend/domain/live-service sources are unchanged by this follow-up; application/test diff is limited to the two HTTP paths. Test diff fingerprint remains `bd7cdce1e044a17425e92be6eda58a38eff3baa69f0c50b8fdbb1fd396c2637a`.

The original AL-53 gate failure at merged `62a7963bb421b3111025fc537bcd91d3d6dc5fe9` remains evidence: `/tmp/al53-integration-make.log` failed the initial readTrace before UPDATE. Retained proof source `/tmp/al53-sse-ordering-proof.go` and output `/tmp/al53-sse-ordering-proof.log` demonstrate the valid schedule; they do not establish the uninstrumented failed run's exact scheduling. The corrected controlled regression independently verifies the ordering and actual concrete service/SQLite delivery. No timeout increase, skipped test, test-only startup bypass, or generic replay guarantee was introduced.

Original feature PR #76 merged as `5da845a7c244a6b2582102c0dff0d221844d9bfc`; only the new follow-up PR will deliver this correction. The completed specification remains in plan/done; no completed ticket index entry was resurrected. Manager owns merges; this owner neither merges nor enables auto-merge. Original feature contract, iteration evidence and five ledger source identities are preserved. Final own-root Codex extraction uses conventional fix classification; the ticket's original feature evidence remains historical rather than reclassified as proof of this correction.


Final correction telemetry validation: generation discovered nine real own-root Codex rollouts. Compared with ledger at base 6bc1064, all five original IDs/source labels/parent links remain, with exactly four new correction children (planner 01a1924c-1c67-78b0-9303-0da57f121e33, coder 01a1924d-ab13-7841-9313-12d490c6876f, tester 01a1924f-6268-7580-82ec-2612063cacb9, reviewer 01a19251-80ea-7090-a487-bcf0905e0f42) parented only to AL-47 root 01a191f5-413e-7250-a3bf-0b3f4e5be953. No manager or other independent root was imported; no other-harness sessions existed in the prior ledger. Ticket classification is fix for this follow-up. Extraction reported no source errors. Native ledger totals cover the historical feature and correction together; T3 MCP activity may be omitted by the decoder, not proven zero activity. Billing remains unavailable despite zero renderer output. Raw rollouts/proof/gate logs stay outside Git. Final delivery adds only documentation/telemetry after the tested implementation head.
