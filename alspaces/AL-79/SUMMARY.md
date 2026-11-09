# AL-79: Native Codex capture and separate version management

[Pull request #130](https://github.com/NiltonDuarte/AgenticLens/pull/130). The approved [contract](CONTRACT.md) and [manager/engine plan](iterations/03/PLAN.md) are implemented and installed on this Linux machine. Initial checkout was clean at 4df45967c801824500ac2ea39a8f52fa9cbdcf66; iteration 03 started from 9c39cf9d4f44bdb968cb92b0212dccf577d10255. The user explicitly approved implementation and installation and waived compatibility with older installations. No merge is authorized. AL-79's specification is done; AL-80 remains Ready.

`~/.local/bin/agentic-lens` is now the native manager. It selects an engine offline, explicitly installs exact engine versions and updates its own executable independently. `agentic-lens-engine` owns capture, generation, workers and the embedded dashboard; it has no dispatcher. The former shell launcher and sidecars are backed up outside Git. This repository's shared `agenticlens.binary_dir: bin` selects `bin/agentic-lens-engine-linux-amd64` from root or nested backend directories. Its exact fallback release pin is unchanged. An explicit companion engine installation supports commands outside projects.

Six additive native user hooks support configured checkouts and linked worktrees, authoritative session/source routing, whole-file local capture policy, independent child ticket assignment, immutable worker images and portable schema v6 history. Interrupted legacy migration retains full undo privately and shares only typed discovery metadata; all resources are leased before recovery. Receipt adoption consumes the current installer protocol. Actual home hooks/configuration and normal human trust were preserved. Earlier real Linux root/child capture evidence carries only for unchanged paths: [native proof](iterations/01/NATIVE_EVIDENCE.md).

Reviewed/built source: `a013dcfa0eedc2881c46ebaf17ad9af827a2b444`; final 71-path content fingerprint `4bd69745495d906d01e8fe48b81e49323d2cda07cdf9397e6988911feda8833c`. Main integration is current. Full `make` passes, including Go race/recovery checks, frontend units/integration, installation/distribution/orchestration checks and both artifact families across five targets. [Verification and installation](iterations/03/FINAL_EVIDENCE.md) records the actual commands, revision and hashes. Resolved test reports are consolidated there; stale private probes are removed, with the open native session and original launcher backup preserved. Installed version: `v3.0.0-rc.133+57a3503-22-ga013dcfa`. Manager/engine/capture protocols and project/nested/standalone execution pass.

Final consolidated review: APPROVED after 48 iteration 03 scoped review cells, with current or explicit carried evidence across behavior, data/failure, architecture, tests, security and build/distribution under gpt-6.1-sol/high and gpt-6-luna/medium. Initial findings were corrected; [compact findings](iterations/03/REVIEW_SUMMARY.json) retains all 24 candidates, shared discovery, refutations and accepted debt. The behavior reviewer gpt-6.1-sol/high discovered the public missing-owner bypass; the parent gpt-6.1-sol/high found cross-home migration through integration; the implementation engineer gpt-6.1-sol/high found the receipt mismatch. The parent and test engineer, both gpt-6.1-sol/high, identified and reproduced the live-disable shutdown race. Implementation engineers gpt-6.1-sol/high fixed production; testers gpt-6.1-sol/high and the focused dual-profile reviewers verified corrections, including disposable mutation controls. Luna security's C5 paired-report exposure is disclosed as correction verification, not independent discovery. Earlier [iteration 01](iterations/01/REVIEW_SUMMARY.json) and [iteration 02](iterations/02/REVIEW_SUMMARY.json) findings remain retained separately.

Windows CI was removed by explicit user instruction; target builds pass but Windows runtime remains uncertified and existing-manager replacement requires manual maintenance. No new release was published. Final Codex extraction validates 94 source sessions: the root, 78 native child sessions and 15 separately rooted tool-delegated reviews. All 41 prior sessions and actual source/parent identities are preserved. The root remains provisional; caller ownership is unverified and unavailable costs remain unknown.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 73
- **Active Agent Duration:** 17h 26m (recorded duration including 255s provisional)
- **Context Window Health:** BLOATED — Root agent peak 451.6k tokens (100% of 450k healthy ceiling, 5 compactions) | Subagents peak 451.4k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.6 avg turns/run (peak: 12 turns) | Planner 1.5, Coder 7.3, Tester 5.0, Reviewer 1.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.0:1 (+3,528 test lines / +3,654 prod lines)
- **Code Delivered:** +7,187 / -685 lines across 92 code files (excluding docs)
- **Reasoning Effort:** 1,075,215 reasoning tokens (33.8% of generated output)
- **Tool Health & Self-Correction:** 4,013 tool calls (248 self-corrections, 93.8% success rate)

### Anomalies & Health Flags
- **Status:** 38 execution anomalies detected (25 CRITICAL, 13 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.1k tokens) — Context tokens reached 450.1k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (451.4k tokens) — Context tokens reached 451.4k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+120.5k tokens) — Context tokens surged by 120.5k to 120.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+122.6k tokens) — Context tokens surged by 122.6k to 122.6k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (71 cycles) — Subagent roles 'coder' and 'tester' alternated for 71 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> tester (9 cycles) — Subagent roles 'reviewer-architecture' and 'tester' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (16 cycles) — Subagent roles 'coder' and 'tester' alternated for 16 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer (22 cycles) — Subagent roles 'reviewer-tests' and 'reviewer' alternated for 22 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (30 cycles) — Subagent roles 'coder' and 'tester' alternated for 30 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (14 cycles) — Subagent roles 'coder' and 'tester' alternated for 14 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer (13 cycles) — Subagent roles 'reviewer-tests' and 'reviewer' alternated for 13 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (7 cycles) — Subagent roles 'coder' and 'tester' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (49 cycles) — Subagent roles 'coder' and 'tester' alternated for 49 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-behavior (12 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-behavior' alternated for 12 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (9 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> reviewer-tests (13 cycles) — Subagent roles 'reviewer' and 'reviewer-tests' alternated for 13 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer (8 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> reviewer-behavior (6 cycles) — Subagent roles 'reviewer' and 'reviewer-behavior' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> reviewer-tests (7 cycles) — Subagent roles 'reviewer' and 'reviewer-tests' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (18 cycles) — Subagent roles 'coder' and 'tester' alternated for 18 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (9 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-data (10 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-data' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (5 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (5 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 5 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+98.5k tokens) — Context tokens surged by 98.5k to 98.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.3k tokens) — Context tokens reached 250.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.9k tokens) — Context tokens reached 253.9k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (257.5k tokens) — Context tokens reached 257.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.3k tokens) — Context tokens reached 253.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.6k tokens) — Context tokens reached 250.6k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.7k tokens) — Context tokens reached 250.7k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.2k tokens) — Context tokens reached 251.2k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> reviewer-behavior (4 cycles) — Subagent roles 'reviewer' and 'reviewer-behavior' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (4 cycles) — Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (4 cycles) — Subagent roles 'tester' and 'coder' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 212,624 | 19,509 | 2,252,544 | 91.4% | 103.3k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **coder** | 3 | 2,634,409 | 305,943 | 121,834,624 | 97.9% | 446.5k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **tester** | 6 | 4,826,940 | 513,276 | 206,553,088 | 97.7% | 451.4k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **reviewer** | 17 | 1,224,137 | 101,504 | 16,370,176 | 93.0% | 129.3k | HEALTHY | Unavailable (0 observed, 0 partial, 17 unavailable / 17 invocations) |
| **Unknown** | 16 | 4,159,527 | 879,496 | 234,497,280 | 98.3% | 451.6k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 16 unavailable / 16 invocations) |
| **reviewer-architecture** | 10 | 918,014 | 64,686 | 11,921,792 | 92.9% | 177.0k | HEALTHY | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **reviewer-behavior** | 13 | 964,688 | 74,004 | 15,834,240 | 94.3% | 152.6k | HEALTHY | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **reviewer-data** | 13 | 960,712 | 77,443 | 14,488,192 | 93.8% | 151.5k | HEALTHY | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **reviewer-tests** | 14 | 832,371 | 69,853 | 14,499,712 | 94.6% | 168.0k | HEALTHY | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **Total** | **94** | **16,733,422** | **2,105,714** | **638,251,648** | **97.4%** | **451.6k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 94 unavailable / 94 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 39 | 2,261,850 | 153,021 | 21,221,504 | 90.4% | Unavailable (0 observed, 0 partial, 39 unavailable / 39 invocations) |
| **gpt-6.1-sol** | 55 | 14,471,572 | 1,952,693 | 617,030,144 | 97.7% | Unavailable (0 observed, 0 partial, 55 unavailable / 55 invocations) |
| **Total** | **94** | **16,733,422** | **2,105,714** | **638,251,648** | **97.4%** | **Unavailable (0 observed, 0 partial, 94 unavailable / 94 invocations)** |

Capture observation finalized: 2026-11-09T19:58:44.729808453Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 35; retained older facts: 20; pending: 59; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 94; reserved bytes: 67051492. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1bcfb-82d0-7430-82a6-f2f191359d4f: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bd2d-769c-7092-90e3-13ad51be1c90: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bc5e-2601-7840-9e69-810cde5028e3: pending; last successful source read 2026-11-08T20:08:34.510221213Z; accepted source checkpoint 2026-11-08T18:15:47.97Z.
Owner codex/01a1bc5c-1650-7532-9e70-b53d09e3c1ca: pending; last successful source read 2026-11-08T20:08:34.535260924Z; accepted source checkpoint 2026-11-08T18:13:49.667Z.
Owner codex/01a1bc55-a30a-7e52-8b72-ee63618fb744: pending; last successful source read 2026-11-08T20:08:34.557557488Z; accepted source checkpoint 2026-11-08T18:07:25.057Z.
Owner codex/01a1bc15-64b6-7f42-9544-6b75d7cc0825: pending; last successful source read 2026-11-08T20:08:34.575445112Z; accepted source checkpoint 2026-11-08T16:55:39.706Z.
Owner codex/01a1bd7b-ad1d-7c52-8982-320a87d23632: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bc02-d12f-7f12-944e-e9ae65bfca5b: pending; last successful source read 2026-11-08T20:08:34.620017418Z; accepted source checkpoint 2026-11-08T16:41:25.475Z.
Owner codex/01a1bd5e-904c-7063-a498-976a1eef6e9c: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bc6c-298f-76f2-8011-2a307ef663d6: pending; last successful source read 2026-11-08T20:08:34.638275142Z; accepted source checkpoint 2026-11-08T18:31:03.925Z.
Owner codex/01a1bd11-cc1e-7f33-bd0b-80bde9c776af: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bd7d-7f20-7052-9757-89d6a4775e00: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bd62-cd28-71d3-80f7-9d2e4010e77c: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bd61-4cae-78a0-a531-c98fad9dcbcb: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bd6c-829f-7703-8ea9-cc96388ff78c: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bd10-d215-7271-a1e9-4823e7b95da1: pending; last successful source read unavailable; accepted source checkpoint unknown.
Additional owner freshness details remain in local capture diagnostics.

Retained recorded history: 9 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
