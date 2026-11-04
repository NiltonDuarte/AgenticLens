# AL-66: Optional generation metric output

Iteration 02 active: user approved lean field-presence output (omitted / recorded-empty / nonempty), removing generation_policy and anomaly_availability from new JSON. Iteration01 evidence below remains historical; current plan supersedes its persistence decisions. Initial iteration02 Git status clean; baseline3372830f0e891dbe8fd13a2568d9853bc1300d0a; PR115 remains open.

All iteration02 discovery reviews complete: both test profiles approved (Sol additionally mutation-tested recorded-empty oracle); five scoped approvals do not refute the three independently reproduced issues. Coder and tester are assigned bounded production/test correction ownership. No new JSON metadata or schema is introduced; current user-approved field-presence scope remains unchanged.


Iteration02 DATA-01 (severity6), discovered by data reviewer gpt-6.1-sol/high: actual backfill/migration reject supported former metadata before normalization; normalized cleanup bytes must also avoid re-emission. Parent independently reran both recovery workflows: marker-free control passes, root-policy and owner-marker cases fail (exit1). ARCH-01 (severity4), discovered by architecture reviewer gpt-6.1-sol/high: permissive v1 incomplete owner still returns recorded-empty anomalies; parent reran actual SQLite selected-malformed-part probe (exit1). Strict publication rejects that fixture, so no strict persisted loss is claimed. Data reviewer gpt-6-luna/medium approved inspected paths without probes; confirmed findings remain open.


Iteration02 review BEH-01 (severity4), discovered by behavior reviewer gpt-6.1-sol/high: first source-backed backfill saves recovered root analysis but derives Markdown from the earlier ledger, so summary recovery lags one run. Parent gpt-6.1-sol/high independently reproduced with actual Execute/save/read/summary overlay; exit1 on first-run mismatch. This is a new root-presence transition, distinct from the iteration01 refuted owner-count hypothesis. Product remains frozen for independent discovery; correction pending.


Iteration02 uses the existing Go1.26 omitzero capability for lean absent/empty/value serialization, without changing toolchain requirements. Tester baseline overlay reproduced fabricated root[] and lost empty skills; production build, typecheck/lint and257 focused frontend tests pass. Product frozen; final test evidence, review and integration pending. Historical iteration01 metrics below have not yet been refreshed. Iteration02 test handoff is complete: affected full CLI/domain/HTTP race passed, all initially green unchanged packages remain valid, corrected frontend cases passed87affected tests, and typecheck/lint/diff checks passed. Candidate262de9f2b527d548d6fb5a2251c44357b8e1d2cbe6579c0cdcfdc2f87efc4fa7 (58paths) frozen for eight independent reviews; no final approval yet.


Implemented independent enabled-by-default settings: `generation.metrics.skills`, `generation.metrics.delegations` and `generation.metrics.anomalies`. Disabled categories are omitted from generated JSON and managed Markdown, including retained history. Core execution metrics, child relationships and enabled history survive. Invalid settings fail before output replacement.

Schema 6 carries optional generation policy and per-owner unavailable-analysis metadata. Reads distinguish disabled, unavailable and recorded-empty evidence; complete source-backed refresh can recover omitted facts. Backfill preserves uncertainty without inheriting current generation settings. Missing original sources cannot recreate intentionally omitted facts. This is a generation output policy, not global source redaction.

## Scope and revision

User approved the contract/plan with the delegations naming correction; no task_links alias. See [contract](CONTRACT.md) and [iteration plan](iterations/01/PLAN.md). Initial Git status was clean. Baseline: a84adbe5be46c2ef52ede0987a417562715d38e1. Implementation commit: 047a98e46df67b8ef69f4afdbdc904500ba43c77. Final reviewed product/test/docs fingerprint: 094daa566b3bce8d327813bcb86a155e7395a96c5cc5a9a88e9e890f238d3683 (55 paths). origin/main integration introduced no changes.

Domain owns typed defaults, immutable projection and availability-aware derivation. Adapters own strict configuration, source policy routing, post-validation retained-DTO omission and canonical serialization. CLI/Markdown/API/UI communicate effective policy and preserve independent tool evidence. Coder and tester: gpt-6.1-sol/high; identities confirmed in source sessions.

## Review and corrections

[Compact findings and dispositions](iterations/01/REVIEW_SUMMARY.json) retain every candidate and all reporting persona/model/effort tuples, including the refuted claim.

| Finding | Severity | Discovering/reporting reviewer | Disposition and fix actor |
| --- | --- | --- | --- |
| BEH-01 | 5 | Behavior, gpt-6-luna/medium | Fixed by coder gpt-6.1-sol/high: fleet tool calls/errors and their table survive disabled anomaly analysis. |
| DATA-01 | 8 | Data integrity, gpt-6.1-sol/high | Fixed by coder gpt-6.1-sol/high: off/on unavailable-owner output is canonical and explicitly unavailable, never null or invented empty. |
| TEST-01 | 3 | Test effectiveness, gpt-6.1-sol/high | Nonblocking suggestion implemented by tester gpt-6.1-sol/high: exact enabled-history values and multiplicity. |
| TEST-02 | 3 | Test effectiveness, gpt-6.1-sol/high | Nonblocking suggestion implemented by tester gpt-6.1-sol/high: event-only recovery limits, canonical reload, summary recovery and actual source restoration. |
| DATA-02 | 7 | Data integrity and architecture, both gpt-6.1-sol/high | Shared independent discovery; fixed by coder gpt-6.1-sol/high: backfill preserves omitted-owner uncertainty before dropping policy. |
| DATA-03 | 4 | Data integrity and architecture, both gpt-6-luna/medium | Refuted by parent gpt-6.1-sol/high exact-count output probe and complete-source producer trace; no product patch. |

Initial eight discovery cells requested changes for BEH-01/DATA-01. Eight corrected-revision cells verified those fixes but found DATA-02. Final six data/architecture/test cells verified the narrow engine/test correction; behavior approvals under both models carry from the unchanged generate/domain/serialization/API/UI/fleet revision. Review profiles: gpt-6.1-sol/high and gpt-6-luna/medium. All 22 review session model/effort settings matched actual Codex source records.

Verification credit is separate from discovery: parent gpt-6.1-sol/high independently reproduced DATA-01/DATA-02 and refuted DATA-03. Tester gpt-6.1-sol/high produced durable red-before/green-after regressions for all three confirmed defects. Final architecture gpt-6.1-sol/high and test reviewers under both profiles corroborated DATA-03 refutation. Architecture/test gpt-6-luna/medium confirmations of supplied DATA-02 are verification, not new discovery. The two final Luna change-request opinions are retained; their sole remaining DATA-03 hypothesis is refuted by concrete output/source evidence.

Final consolidated verdict: APPROVED by parent gpt-6.1-sol/high after the required integrated gate on implementation commit 047a98e46df67b8ef69f4afdbdc904500ba43c77. No unresolved confirmed finding or material review hypothesis.

## Verification and limits

`make` passed (exit 0) after clean upstream integration: dependencies, Go vet, TypeScript typecheck/lint, release/installation/capture script checks, full backend race tests, 837 frontend tests in 36 files, canonical frontend build and all configured platform binaries. Parent executed the built host binary and verified each frontend file matches the embedded tree and occurs byte-for-byte in the executable. Git status remained clean after the gate.

Focused cases cover all eight toggle combinations; first creation and retained same/other-harness owners; v1/v2/Codex/registered generation; strict invalid config/metadata before output changes; actual JSON absence and managed-summary replacement; canonical off/on reload/API/repeat; exact retained facts and genuine missing-to-readable recovery; mixed/all-disabled fleet metrics; missing/partial backfill owners beside complete-empty and recorded nonempty findings.

Prior full Go race and frontend suite (837 tests), typecheck/lint, canonical build and frontend=embed=standalone asset checks passed. Review probes and affected backfill race also passed. No manual browser geometry, remote CI or release execution is claimed. Raw source transcripts and detailed working review evidence stay outside Git; no historical ticket audit was modified.

## Telemetry

Final ingestion is intentionally last. Source audit identified the actual Codex root 01a1a7d7-b906-7f32-a4a9-56e8f9f1f176, three native child sessions and 22 T3 review sessions. T3 source metadata records those reviews as independent native roots with no parent ID; preserve that evidence rather than invent parent links. Explicit source IDs are required to include them, followed by the final literal-nonce root refresh. Extraction completeness and actual metrics are verified at finalization below. The built CLI rejects legacy wrapper flags --al-offline/--al-project; supported native generate flags are used from this project directory. Session extraction uses explicit source codex and conventional type feat; original source metadata/hierarchy are preserved. Configured capture coordinates are unsupported by this source build; explicit manual generation uses a fresh task-owned state directory while preserving the existing configured state. No source evidence is synthesized or treated as empty to bypass this operational incompatibility.

Telemetry verification passed: schema 6, ticket AL-66, conventional type feat, exactly all 26 audited Codex session IDs, three native parent links, and all 22 T3 review roots retained. Every explicit import preserved all prior accepted sessions; the final nonce 291d579a6c7a selected the actual root plus its three native children. Canonical CLI read/export also passed. There was no ledger at baseline and therefore no pre-existing other-harness history to preserve. Source role metadata is unknown for the root and 22 T3 roots (23 Unknown roles); it is not rewritten to reviewer/orchestrator guesses. Review-role attribution is separately retained above with source-observed model/effort. All 26 invocation costs remain explicitly unavailable, never billed zero. Aggregate active-agent duration is 2h 40m; cache hit97.3%; context ELEVATED.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 70
- **Active Agent Duration:** 2h 40m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 257.0k tokens (57% of 450k healthy ceiling, 0 compactions) | Subagents peak 234.5k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.3 avg turns/run (peak: 3 turns) | Planner 1.0, Coder 3.0, Tester 3.0
- **Test-to-Production Code Ratio:** 1.2:1 (+919 test lines / +756 prod lines)
- **Code Delivered:** +1,675 / -195 lines across 51 code files (excluding docs)
- **Reasoning Effort:** 70,708 reasoning tokens (20.4% of generated output)
- **Tool Health & Self-Correction:** 767 tool calls (60 self-corrections, 92.2% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (1 CRITICAL, 2 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (31 cycles) — Subagent roles 'coder' and 'tester' alternated for 31 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.4k tokens) — Context tokens reached 251.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (257.0k tokens) — Context tokens reached 257.0k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 88,993 | 7,454 | 1,006,080 | 91.9% | 94.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 178,877 | 48,044 | 17,875,712 | 99.0% | 234.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 157,421 | 44,308 | 16,560,896 | 99.1% | 231.2k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 23 | 1,910,357 | 176,193 | 47,361,536 | 96.1% | 257.0k | ELEVATED | Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations) |
| **Total** | **26** | **2,335,648** | **275,999** | **82,804,224** | **97.3%** | **257.0k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 26 unavailable / 26 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 11 | 791,883 | 49,017 | 5,371,520 | 87.2% | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **gpt-6.1-sol** | 15 | 1,543,765 | 226,982 | 77,432,704 | 98.0% | Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations) |
| **Total** | **26** | **2,335,648** | **275,999** | **82,804,224** | **97.3%** | **Unavailable (0 observed, 0 partial, 26 unavailable / 26 invocations)** |

Capture observation finalized: 2026-11-04T19:59:52.429066354Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 22; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 22; reserved bytes: 17511981. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1a7f9-7700-7622-b7be-5ee0deddf8cc: current; last successful source read 2026-11-04T19:59:51.801989368Z; accepted source checkpoint 2026-11-04T19:14:34.559Z.
Owner codex/01a1a81b-8929-7850-a731-884d293f646c: current; last successful source read 2026-11-04T19:59:51.820531537Z; accepted source checkpoint 2026-11-04T19:50:04.032Z.
Owner codex/01a1a80b-e82b-7381-97e3-90e20302b230: current; last successful source read 2026-11-04T19:59:51.849525919Z; accepted source checkpoint 2026-11-04T19:34:26.93Z.
Owner codex/01a1a817-a03c-78f3-823e-68d4dc781bea: current; last successful source read 2026-11-04T19:59:51.869662342Z; accepted source checkpoint 2026-11-04T19:46:03.477Z.
Owner codex/01a1a7fa-f39e-7d33-beea-3c50bd9d0f32: current; last successful source read 2026-11-04T19:59:51.895190798Z; accepted source checkpoint 2026-11-04T19:16:09.818Z.
Owner codex/01a1a81a-dad9-7293-a4e8-fdda363e5a08: current; last successful source read 2026-11-04T19:59:51.915519843Z; accepted source checkpoint 2026-11-04T19:49:51.855Z.
Owner codex/01a1a7fa-f009-75b3-a19e-d2b866405f88: current; last successful source read 2026-11-04T19:59:51.938947292Z; accepted source checkpoint 2026-11-04T19:15:30.974Z.
Owner codex/01a1a817-9d74-7913-b128-995d1fde451f: current; last successful source read 2026-11-04T19:59:51.960389088Z; accepted source checkpoint 2026-11-04T19:46:03.098Z.
Owner codex/01a1a7fc-8ee0-7a90-9cd9-1f15597d3852: current; last successful source read 2026-11-04T19:59:51.981929092Z; accepted source checkpoint 2026-11-04T19:17:09.842Z.
Owner codex/01a1a80f-4c04-7ce1-8832-10d464c4f3df: current; last successful source read 2026-11-04T19:59:52.004231328Z; accepted source checkpoint 2026-11-04T19:37:51.383Z.
Owner codex/01a1a80e-2a8b-7800-8bce-8056e15184da: current; last successful source read 2026-11-04T19:59:52.023506504Z; accepted source checkpoint 2026-11-04T19:36:23.924Z.
Owner codex/01a1a817-a2ee-7140-bb51-08b783a346d3: current; last successful source read 2026-11-04T19:59:52.042451211Z; accepted source checkpoint 2026-11-04T19:46:52.6Z.
Owner codex/01a1a819-5494-70f2-b182-f13dbcec3e8a: current; last successful source read 2026-11-04T19:59:52.063071766Z; accepted source checkpoint 2026-11-04T19:48:17.679Z.
Owner codex/01a1a809-5dd8-7b03-a585-62023a2dfc47: current; last successful source read 2026-11-04T19:59:52.082595574Z; accepted source checkpoint 2026-11-04T19:30:56.636Z.
Owner codex/01a1a809-5889-7500-881b-ab0cd989dd6e: current; last successful source read 2026-11-04T19:59:52.107767962Z; accepted source checkpoint 2026-11-04T19:31:50.995Z.
Owner codex/01a1a809-5ae7-7de2-b05c-298d915ddf2f: current; last successful source read 2026-11-04T19:59:52.130987707Z; accepted source checkpoint 2026-11-04T19:30:58.515Z.
Additional owner freshness details remain in local capture diagnostics.
