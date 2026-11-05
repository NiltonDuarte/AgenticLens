# AL-66: Lean optional metric output

Generation settings remain independent and enabled by default: `generation.metrics.skills`, `.delegations` and `.anomalies`. New ledger/API output contains no `generation_policy`, `anomaly_availability` or replacement policy/reason metadata.

| Collection shape | Meaning |
| --- | --- |
| Omitted | Not recorded; no zero/healthy claim |
| `[]` | Actually recorded-empty |
| Nonempty array | Recorded facts |

This contract applies at each canonical owner, including session skills/delegations/anomalies and root anomaly summaries. Runtime generation can explain disabled configuration; persisted omission deliberately does not retain its cause. Complete supported source observations can establish or recover empty/nonempty collections. Missing/partial/other-source owners preserve known values and presence; ordinary tools, child relationships, invocations and costs remain intact. Unsupported or unavailable sources cannot recreate intentionally omitted facts. Schema 6 stays current; supported former metadata validates on reads and is stripped on explicit writes, without bulk rewrites of other ledgers.

## Approval and revision

User explicitly approved the field-presence approach. The [contract amendment](CONTRACT.md) and [iteration 02 plan](iterations/02/PLAN.md) supersede iteration01 metadata-persistence decisions. Config keys, strict defaults/precedence, command scope, portable-history/safe-write guarantees and source-backed recovery limits remain.

Iteration02 baseline: 3372830f0e891dbe8fd13a2568d9853bc1300d0a; initial Git status clean. Implementation commit:65114b77f5bb105488c6a583f036b2947a93fc05. Final reviewed product/test/current-doc fingerprint: 7bcd5968cc9f895dc97cc06447fdbe2d843e1fe18526e139c5b6e90ef4088110, 60 paths. Current origin/main integration introduced no changes.

The existing Go1.26 `omitzero` encoding preserves nil versus allocated-empty without new custom presence serializers or a toolchain bump. Domain/value conversions and source/category authority remain separate from adapter JSON/legacy normalization. Current readers preserve absence rather than reconstructing analysis from tools. Canonical/API/cache/export writers and frontend decoding keep all three states; known rows remain visible under partial coverage.

## Review and corrections

[Iteration02 compact findings](iterations/02/REVIEW_SUMMARY.json) retain every finding and original reporter identity. All 16 discovery/correction review sessions used the configured profiles, gpt-6.1-sol/high and gpt-6-luna/medium; actual Codex source model/effort matched every request.

| Finding | Severity | Discovery | Disposition and fix actor |
| --- | --- | --- | --- |
| BEH-01 | 4 | Behavior reviewer, gpt-6.1-sol/high | Fixed by coder gpt-6.1-sol/high: backfill Markdown derives the prepared ledger published as JSON, including first/repeat source recovery. |
| DATA-01 | 6 | Data integrity reviewer, gpt-6.1-sol/high | Fixed by coder gpt-6.1-sol/high: backfill/migration normalize supported former metadata before domain decoding; raw cleanup uses normalized bytes and retains original duplicate/contradiction guards. |
| ARCH-01 | 4 | Architecture reviewer, gpt-6.1-sol/high | Fixed by coder gpt-6.1-sol/high: permissive v1 inspects selected owner parts before classification and omits newly empty incomplete evidence, retaining known facts and strict rejection. No prior strict publication loss was demonstrated. |

Initial eight discovery cells produced five scoped approvals and three change requests. Parent gpt-6.1-sol/high independently reproduced all three reports; approvals did not override those counterexamples. Tester gpt-6.1-sol/high produced durable failing-before/passing-after tests using actual backfill, migration cleanup, SQLite extraction, canonical save/reload, repeated runs and invalid-file preservation. Both profiles for behavior/data/architecture/test effectiveness then approved the bounded corrected revision, with no new distinct findings. Reviewers supply verification credit, not new discovery credit for the supplied issues. Initial test-effectiveness Sol also independently mutated serialization to prove the recorded-empty oracle fails.

Final consolidated verdict: APPROVED by parent gpt-6.1-sol/high after integrated make on 65114b77f5bb105488c6a583f036b2947a93fc05. No unresolved confirmed finding or material review hypothesis.

## Verification and limits

`make` passed (exit 0) after clean main integration: dependencies, Go vet, frontend typecheck/lint, release/installation/capture script checks, full backend race tests, all 844 frontend tests in 36 files, canonical frontend build and all configured platform binaries. Parent executed the built host binary and verified every frontend asset equals the embedded tree and occurs byte-for-byte in the executable. Working tree was clean after the gate.

Baseline overlay proves real SaveFactLedger previously fabricated root `[]` from nil and dropped recorded-empty skills. Representation tests cover all three categories, complete/partial owner authority, retained raw history, all eight config combinations, actual Codex/OpenCode v1/v2/registered generation, off/on and real source restoration, backfill/first-summary/repeat, former-metadata recovery/cleanup, invalid artifact preservation, API JSON/GET no-write and frontend unknown-versus-recorded-zero states. Source parity fixtures now supply actual complete SourceEvents on both sides while preserving exact root/session parity; ordinary unsourced reads leave absent analysis absent.

Initial broad tests had explicitly disclosed obsolete metadata/absent-to-zero oracles; affected full-package race and frontend reruns passed after bounded fixture corrections. Prior valid unchanged checks were reused; the final integrated gate establishes the final tree. Embedded assets and standalone execution are verified. Manual browser geometry, remote CI and release execution are not claimed. Raw transcripts/detailed working reports stay outside Git.

## Historical iteration01

The [iteration01 review summary](iterations/01/REVIEW_SUMMARY.json) preserves six original findings/suggestions, their numeric severities, reporting identities and dispositions, including the refuted owner-count claim. Its earlier approved implementation was 047a98e46df67b8ef69f4afdbdc904500ba43c77, finalized 3372830f0e891dbe8fd13a2568d9853bc1300d0a. Its metadata representation is superseded, not retroactively described as field-presence approval. Existing fleet tool, exact history, canonical off/on and backfill recovery protections carry through the new representation.

## Telemetry

Final ingestion follows implementation/review/integration. Audited source inventory contains the actual Codex root 01a1a7d7-b906-7f32-a4a9-56e8f9f1f176, three native child sessions and 38 T3 review roots (22 iteration01, 16 iteration02), 42 sessions total. T3 metadata has no native parent IDs; preserve original hierarchy rather than invent links. Existing 26-session AL-66 history must remain a subset after refresh. Source roles not present remain Unknown; review work-role attribution above is separate and source-observed models/efforts are verified.

The source CLI supports native generate flags, not the legacy --al-offline/--al-project wrapper flags. Explicit source codex, type feat and a literal nonce select the final root refresh. Configured capture coordinates are from an unsupported older layout; explicit manual generation uses a separate task-owned state directory and leaves configured state untouched. Missing pricing stays unavailable. Final extraction verification and metrics follow below.

Final extraction verification passed: schema 6, ticket AL-66, type feat, all 42 audited Codex session IDs and all 26 pre-existing session IDs preserved. Three native parent links and 38 independent T3 source roots remain as recorded. Literal nonce 93dff0ab2532 selected the actual root plus its three native children. Both removed metadata keys are absent recursively from every JSON object; present optional categories are arrays, never null. Canonical CLI read/export passed. Native source has no role metadata for the root and 38 T3 roots, so 39 roles remain Unknown rather than inferred; review-role attribution above is separate. All42 costs remain explicitly unavailable, not billed zero. Aggregate active-agent time 5h 19m, cache hit 98.1%, context ELEVATED. The previous ledger contained only Codex sessions, so no other-harness baseline was omitted.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 120
- **Active Agent Duration:** 5h 19m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 431.3k tokens (96% of 450k healthy ceiling, 0 compactions) | Subagents peak 406.3k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 4.0 avg turns/run (peak: 5 turns) | Planner 2.0, Coder 5.0, Tester 5.0
- **Test-to-Production Code Ratio:** 2.0:1 (+1,647 test lines / +817 prod lines)
- **Code Delivered:** +2,464 / -336 lines across 65 code files (excluding docs)
- **Reasoning Effort:** 190,454 reasoning tokens (25.2% of generated output)
- **Tool Health & Self-Correction:** 1,370 tool calls (119 self-corrections, 91.3% success rate)

### Anomalies & Health Flags
- **Status:** 5 execution anomalies detected (2 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (31 cycles) — Subagent roles 'coder' and 'tester' alternated for 31 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (41 cycles) — Subagent roles 'coder' and 'tester' alternated for 41 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (255.4k tokens) — Context tokens reached 255.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.5k tokens) — Context tokens reached 250.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.4k tokens) — Context tokens reached 251.4k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 216,414 | 14,815 | 2,423,296 | 91.8% | 133.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 263,297 | 96,257 | 43,122,688 | 99.4% | 353.3k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 286,731 | 109,850 | 50,091,904 | 99.4% | 406.3k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 39 | 3,290,338 | 344,404 | 114,307,328 | 97.2% | 431.3k | ELEVATED | Unavailable (0 observed, 0 partial, 39 unavailable / 39 invocations) |
| **Total** | **42** | **4,056,780** | **565,326** | **209,945,216** | **98.1%** | **431.3k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 42 unavailable / 42 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 19 | 1,323,729 | 83,945 | 10,132,608 | 88.4% | Unavailable (0 observed, 0 partial, 19 unavailable / 19 invocations) |
| **gpt-6.1-sol** | 23 | 2,733,051 | 481,381 | 199,812,608 | 98.7% | Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations) |
| **Total** | **42** | **4,056,780** | **565,326** | **209,945,216** | **98.1%** | **Unavailable (0 observed, 0 partial, 42 unavailable / 42 invocations)** |

Capture observation finalized: 2026-11-04T21:51:46.596310156Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 42; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 42; reserved bytes: 47672560. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1a7f9-7700-7622-b7be-5ee0deddf8cc: current; last successful source read 2026-11-04T21:51:45.086304382Z; accepted source checkpoint 2026-11-04T19:14:34.559Z.
Owner codex/01a1a87f-cdfe-7662-a6cb-837d2e0d211e: current; last successful source read 2026-11-04T21:51:45.100302221Z; accepted source checkpoint 2026-11-04T21:39:57.657Z.
Owner codex/01a1a81b-8929-7850-a731-884d293f646c: current; last successful source read 2026-11-04T21:51:45.113588815Z; accepted source checkpoint 2026-11-04T19:50:04.032Z.
Owner codex/01a1a87f-2e16-7d12-a929-29f12425cffe: current; last successful source read 2026-11-04T21:51:45.12717982Z; accepted source checkpoint 2026-11-04T21:38:52.867Z.
Owner codex/01a1a80b-e82b-7381-97e3-90e20302b230: current; last successful source read 2026-11-04T21:51:45.148806457Z; accepted source checkpoint 2026-11-04T19:34:26.93Z.
Owner codex/01a1a7e3-1891-7332-95a4-9cb20251a8eb: current; last successful source read 2026-11-04T21:51:45.232687457Z; accepted source checkpoint 2026-11-04T21:29:07.826Z.
Owner codex/01a1a817-a03c-78f3-823e-68d4dc781bea: current; last successful source read 2026-11-04T21:51:45.246395035Z; accepted source checkpoint 2026-11-04T19:46:03.477Z.
Owner codex/01a1a869-d1da-71e3-962b-412d7e7a5f82: current; last successful source read 2026-11-04T21:51:45.261505126Z; accepted source checkpoint 2026-11-04T21:15:33.05Z.
Owner codex/01a1a7fa-f39e-7d33-beea-3c50bd9d0f32: current; last successful source read 2026-11-04T21:51:45.287958271Z; accepted source checkpoint 2026-11-04T19:16:09.818Z.
Owner codex/01a1a7ea-bcd8-7281-84c6-932d00907e1c: current; last successful source read 2026-11-04T21:51:45.401935134Z; accepted source checkpoint 2026-11-04T21:34:37.994Z.
Owner codex/01a1a81a-dad9-7293-a4e8-fdda363e5a08: current; last successful source read 2026-11-04T21:51:45.425304848Z; accepted source checkpoint 2026-11-04T19:49:51.855Z.
Owner codex/01a1a86f-eca9-7ac1-be3d-af090fd925c3: current; last successful source read 2026-11-04T21:51:45.450192869Z; accepted source checkpoint 2026-11-04T21:22:26.671Z.
Owner codex/01a1a7fa-f009-75b3-a19e-d2b866405f88: current; last successful source read 2026-11-04T21:51:45.483157669Z; accepted source checkpoint 2026-11-04T19:15:30.974Z.
Owner codex/01a1a817-9d74-7913-b128-995d1fde451f: current; last successful source read 2026-11-04T21:51:45.509219561Z; accepted source checkpoint 2026-11-04T19:46:03.098Z.
Owner codex/01a1a86d-742d-7cf2-beb6-c51f35088297: current; last successful source read 2026-11-04T21:51:45.538431625Z; accepted source checkpoint 2026-11-04T21:20:23.383Z.
Owner codex/01a1a7d7-b906-7f32-a4a9-56e8f9f1f176: current; last successful source read 2026-11-04T21:51:45.70670715Z; accepted source checkpoint 2026-11-04T21:51:45.046Z.
Additional owner freshness details remain in local capture diagnostics.
