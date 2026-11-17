# docs(audit): verify cross-harness source capabilities (AL-63)

## Outcome

Codex historical hydration returned 232 events for all five recorded AL-55 owners. The v1 and v2 OpenCode records returned 27 and four events respectively from single-family copies. Current historical reads reject an intact mixed-family store even when the owner ID is known. Live observation works for v1, fails for v2 and has no Codex reader. [Full matrix, versions and source IDs](CAPABILITIES.md)

## Verification and review

The November 17 review used application revision `387624b3`, which integrates main `0e9c8b36`. Exact-owner backfill completed twice for every source on disposable ledgers. Every second run preserved the entire session array. The OpenCode checks use real rows in family-isolated copies and constructed selector ledgers. The Codex check uses a copied real ticket and real rollout files. The original ticket is byte-identical to its initial revision. AL-55 and AL-62 match main, including the upstream tag addition.

Focused checks used `go test -count=1` from `backend/`. The CLI, backfill, OpenCode and live packages passed the selected cases for child read errors, manual generation, duplicate location selection, insufficient discovery coverage, child proof, structured execution and mixed-source hydration. The three additional backfill cases for skill/task snapshots, v2 grandchildren and growing duration passed. The November 17 service probes and the CLI command are recorded in the matrix.

Eight independent reviews evaluated the original audit revision. Their findings, identities and initial incomplete outcomes remain attributed. The audit owner rechecked the affected paths and discovered the mixed-store historical qualification during this refresh. [Review record](evidence/REVIEW.md) · [Findings and dispositions](evidence/REVIEW_SUMMARY.json)

Full `make` exited 0 on the integrated revision, including backend race and integration tests, 231 frontend unit tests, 804 frontend integration tests and four native build targets. Strict OpenSpec validation passed, as did all three repository workflow tests. The layout update passed strict OpenSpec validation and a link check across the change. The application and build inputs match the tested revision.

## Telemetry

The change carries the existing AL-63 ledger. The integrated engine publishes the refresh because the shared installed pin predates the OpenSpec workspace layout. This run uses a temporary worktree-local copy of the selected machine policy with a private publication state directory. Capture stays disabled. The refreshed schema 6 docs ledger preserves all five previous session IDs, sources and parent links, including the root and four direct reviewers. The source reported an open turn at extraction. Service-level delegates remain unattributed in that graph.

## Limits

OpenCode v1 metadata-backed child links still disappear. Codex failed/error function outputs still leave tool error counts at zero. OpenCode v2 lacks a native positive skill/delegation sample in the selected records. Malformed live-part diagnostics have only source-inspection evidence. A newly active live session, browser navigation and heuristic backfill were outside this replay. Four service-delegated reviewer sessions lack proven native linkage.

## Pull request

This audit updates the eight-capability matrix for OpenCode v1, OpenCode v2 and Codex and collects its review records directly under the OpenSpec change’s evidence directory. A replay against integrated main corrected historical OpenCode coverage. Both single-family stores can hydrate their recorded owners, while an intact mixed store returns unavailable with a diagnostic. Codex hydrated all five AL-55 owners and exposed partial coverage when four rollouts were missing.

Exact-owner backfill passed twice per source on disposable copies. The report retains the broken v1 child-link and Codex error-count cases, the v2 live limit and the bounds of the native evidence. Original roadmap acceptance is preserved.

Full `make` passed (231 frontend unit and 804 integration tests), alongside strict OpenSpec validation and three workflow tests.

[Capability matrix](https://github.com/NiltonDuarte/AgenticLens/blob/AL-63/verify-source-capabilities-across-harnesses/openspec/changes/al-63-verify-source-capabilities-across-harnesses/CAPABILITIES.md) · [AL-63 pull request](https://github.com/NiltonDuarte/AgenticLens/pull/133)

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 13
- **Active Agent Duration:** 1h 14m (recorded duration including 99s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 259.9k tokens (58% of 450k healthy ceiling, 1 compactions) | Subagents peak 129.5k (Reviewer-behavior)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 36,450 reasoning tokens (21.1% of generated output)
- **Tool Health & Self-Correction:** 356 tool calls (33 self-corrections, 90.7% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (2 WARNING).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (252.5k tokens) — Context tokens reached 252.5k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 2 | 807,756 | 94,362 | 26,877,056 | 97.1% | 259.9k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-architecture** | 1 | 96,524 | 11,312 | 1,920,640 | 95.2% | 103.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 108,653 | 13,625 | 2,883,968 | 96.4% | 129.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 92,417 | 10,347 | 2,242,432 | 96.0% | 110.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 85,512 | 6,449 | 1,026,048 | 92.3% | 102.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **6** | **1,190,862** | **136,095** | **34,950,144** | **96.7%** | **259.9k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 5 | 1,050,148 | 118,834 | 30,536,192 | 96.7% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **gpt-6.2-astra** | 1 | 140,714 | 17,261 | 4,413,952 | 96.9% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **6** | **1,190,862** | **136,095** | **34,950,144** | **96.7%** | **Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations)** |
<!-- agentic-lens:metrics:end -->
