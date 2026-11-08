# AL-79: Native user-level Codex capture across worktrees

[PR130](https://github.com/NiltonDuarte/AgenticLens/pull/130). Consolidated review: APPROVED for the current user-agreed scope; no merge is authorized. Initial checkout was clean at4df45967c801824500ac2ea39a8f52fa9cbdcf66. Branch: feat/AL-79-native-user-codex-capture; configured workspace root: alspaces.

User approved [CONTRACT.md](CONTRACT.md) and [iteration01 plan](iterations/01/PLAN.md). The later instruction, "Remove windows test CI, I will remove support from windows later", approves [iteration02's delta](iterations/02/PLAN.md): remove the Windows job/build edge, retain runtime scripts/binaries/release targets, and stop further Windows verification. Other acceptance boundaries remain. AL-79 is moved to done and AL-80 is Ready.

## Delivered behavior

Six additive native Codex hooks install once in the effective user home. Callbacks validate exact session/source context and select one whole local configuration: worktree-local first, otherwise Git-base-local. Shared policy supplies executable selection; capture enablement, source and state remain local. Disabled capture stops before capture-only I/O. Unassigned facts stay local until explicit binding.

Adapters preserve root/child identities, ancestry and independently bound child tickets; retain immutable worker builds across rebuilds; and coalesce publication with bounded retry, durable recovery and deletion retention. Portable v6 history accepts an external parent without importing another ticket's metrics. Partial group context preserves owner/cached analysis without a false clean conclusion. Pure domain logic remains separate from I/O. Manual pins/cwd, normal trust and standalone embedded assets remain supported. Setup uses the common native resolver.

Repository adoption: the user subsequently requested binary_dir in shared configuration. The one-line agenticlens.binary_dir: bin change retains the exact fallback pin. Native resolution and the current repository wrapper selected/executed bin/agentic-lens-linux-amd64 from backend cwd in an isolated home/cache with no pinned binary. The existing global launcher rejects the new key and needs a separate upgrade; its files/hooks were preserved. Runtime source/review evidence is unchanged.

## Review and verification

Reviewed source160db423ce493942a7d0ca0cca6e56fe98185d75, source diff SHA256139d0e87e1e014ef0035b4aff4a09edfef843056352cd0545d2479e1309cf672; final six-file correction fingerprint69893e0f69877f377c9dcdf59989893b9768e6e0f6e69754112b2e0d58759adc. Behavior, data/failure, architecture, tests, security and build/distribution reviews used gpt-6.1-sol/high and gpt-6-luna/medium. Initial CHANGES_REQUESTED findings were corrected, later test-only scopes passed both profiles, and final bounded CI/distribution review passed both. Earlier Windows INCOMPLETE evidence was superseded by the explicit scope withdrawal, never relabeled a platform pass.

[Canonical findings](iterations/01/REVIEW_SUMMARY.json) retain28 candidates, numeric severity, every discovering persona/model/effort, fixes, the refuted routing claim and accepted Windows fixture debt. [Final correction review](iterations/02/REVIEW_SUMMARY.json) found no new candidate. Architecture/build reviewers and the independent parent, all gpt-6.1-sol/high, share process-lifetime discovery. Data reviewer gpt-6-luna/medium discovered executable retention; its architecture peer's routing objection was refuted. Go/distribution testers gpt-6.1-sol/high discovered recovery/scaffold/fixture issues; the parent gpt-6.1-sol/high found integrated assertion and native transport/diagnostic failures. Implementation engineer gpt-6.1-sol/high fixed runtime; Go/distribution testers gpt-6.1-sol/high fixed oracles. Independent reviewers and parent verified corrections; discovery is kept separate.

Full make passed on8b15b2ee: all Go races/recovery locks,190 frontend units,678 integration tests, installation/distribution/workflow checks and five-platform compilation. All314 Go/module/embed files remain identical: f1b674667f1247eeccfe01640f5353d27a27e9794831aa2b99ce6d917bd6195b. Current corrections also pass POSIX installation10/distribution21 and dependency-graph checks. All six retained [CI checks](https://github.com/NiltonDuarte/AgenticLens/actions/runs/42514876409) passed on reviewed source160db423; subsequent delivery changes are documentation/telemetry only. [Test evidence](iterations/01/TEST_EVIDENCE.md) and [distribution evidence](iterations/01/DISTRIBUTION_EVIDENCE.md) distinguish passes, failures and stops.

Genuine normal-trust Codex0.167.0/Linux root/child capture and automatic publication passed in two worktrees with exact source facts and preserved history. The user signed in and trusted six isolated hooks normally. Unchanged callback/extraction/binding paths retain that revision-bounded [native proof](iterations/01/NATIVE_EVIDENCE.md).

Windows remains uncertified: actual win32 plugin6 and distribution12/14 plus isolated completion passed; DrvFS/setup and loaded-upgrade assurance remained limited. The earlier unavailable-local-runtime assumption was corrected. These retained limits belong to the user's planned support removal.

Telemetry is a final Codex checkpoint:41 source sessions, comprising the root,25 native children and15 separately rooted tool-delegated reviews. Exact source membership/parentage and preservation of prior imports passed validation; no prior other-harness ledger existed. Native parentage for the tool-delegated roots is unavailable and was not invented. Isolated acceptance-probe usage stays in its separate-project evidence. The main root remains provisional; unavailable costs are not fabricated.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 61
- **Active Agent Duration:** 10h 31m (recorded duration including 677s provisional)
- **Context Window Health:** BLOATED — Root agent peak 450.1k tokens (100% of 450k healthy ceiling, 3 compactions) | Subagents peak 451.4k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.0 avg turns/run (peak: 12 turns) | Planner 1.0, Coder 12.0, Tester 4.8, Reviewer 1.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.0:1 (+2,377 test lines / +2,453 prod lines)
- **Code Delivered:** +4,830 / -595 lines across 74 code files (excluding docs)
- **Reasoning Effort:** 650,168 reasoning tokens (34.6% of generated output)
- **Tool Health & Self-Correction:** 2,236 tool calls (132 self-corrections, 94.1% success rate)

### Anomalies & Health Flags
- **Status:** 17 execution anomalies detected (11 CRITICAL, 6 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.1k tokens) — Context tokens reached 450.1k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (451.4k tokens) — Context tokens reached 451.4k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+120.5k tokens) — Context tokens surged by 120.5k to 120.5k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (71 cycles) — Subagent roles 'coder' and 'tester' alternated for 71 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> tester (9 cycles) — Subagent roles 'reviewer-architecture' and 'tester' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (16 cycles) — Subagent roles 'coder' and 'tester' alternated for 16 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer (22 cycles) — Subagent roles 'reviewer-tests' and 'reviewer' alternated for 22 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (30 cycles) — Subagent roles 'coder' and 'tester' alternated for 30 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (14 cycles) — Subagent roles 'coder' and 'tester' alternated for 14 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer (13 cycles) — Subagent roles 'reviewer-tests' and 'reviewer' alternated for 13 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (7 cycles) — Subagent roles 'coder' and 'tester' alternated for 7 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+98.5k tokens) — Context tokens surged by 98.5k to 98.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.3k tokens) — Context tokens reached 250.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.9k tokens) — Context tokens reached 253.9k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (257.5k tokens) — Context tokens reached 257.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.2k tokens) — Context tokens reached 251.2k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 91,461 | 8,921 | 771,840 | 89.4% | 98.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 1,384,039 | 166,026 | 72,547,968 | 98.1% | 446.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 4 | 2,714,396 | 343,148 | 142,070,272 | 98.1% | 451.4k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer** | 5 | 390,921 | 41,431 | 7,422,464 | 95.0% | 129.3k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Unknown** | 16 | 2,021,216 | 532,781 | 134,427,776 | 98.5% | 450.1k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 16 unavailable / 16 invocations) |
| **reviewer-architecture** | 4 | 446,218 | 39,441 | 7,805,824 | 94.6% | 177.0k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 3 | 347,832 | 31,929 | 8,083,456 | 95.9% | 152.6k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-data** | 3 | 331,904 | 38,039 | 8,578,816 | 96.3% | 151.5k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-tests** | 4 | 309,400 | 26,986 | 5,904,512 | 95.0% | 168.0k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **41** | **8,037,387** | **1,228,702** | **387,612,928** | **98.0%** | **451.4k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 41 unavailable / 41 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 15 | 1,069,019 | 62,167 | 7,651,712 | 87.7% | Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations) |
| **gpt-6.1-sol** | 26 | 6,968,368 | 1,166,535 | 379,961,216 | 98.2% | Unavailable (0 observed, 0 partial, 26 unavailable / 26 invocations) |
| **Total** | **41** | **8,037,387** | **1,228,702** | **387,612,928** | **98.0%** | **Unavailable (0 observed, 0 partial, 41 unavailable / 41 invocations)** |

Capture observation finalized: 2026-11-08T20:08:34.980294304Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 39; retained older facts: 1; pending: 2; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 41; reserved bytes: 66615179. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1bc5e-2601-7840-9e69-810cde5028e3: current; last successful source read 2026-11-08T20:08:34.510221213Z; accepted source checkpoint 2026-11-08T18:15:47.97Z.
Owner codex/01a1bc5c-1650-7532-9e70-b53d09e3c1ca: current; last successful source read 2026-11-08T20:08:34.535260924Z; accepted source checkpoint 2026-11-08T18:13:49.667Z.
Owner codex/01a1bc55-a30a-7e52-8b72-ee63618fb744: current; last successful source read 2026-11-08T20:08:34.557557488Z; accepted source checkpoint 2026-11-08T18:07:25.057Z.
Owner codex/01a1bc15-64b6-7f42-9544-6b75d7cc0825: current; last successful source read 2026-11-08T20:08:34.575445112Z; accepted source checkpoint 2026-11-08T16:55:39.706Z.
Owner codex/01a1bc02-d12f-7f12-944e-e9ae65bfca5b: current; last successful source read 2026-11-08T20:08:34.620017418Z; accepted source checkpoint 2026-11-08T16:41:25.475Z.
Owner codex/01a1bc6c-298f-76f2-8011-2a307ef663d6: current; last successful source read 2026-11-08T20:08:34.638275142Z; accepted source checkpoint 2026-11-08T18:31:03.925Z.
Owner codex/01a1bc0a-17b8-74a3-ba6f-83e308083e66: current; last successful source read 2026-11-08T20:08:34.6695387Z; accepted source checkpoint 2026-11-08T16:48:26.304Z.
Owner codex/01a1bbd3-f01a-7772-ae01-e52423c52f48: pending; last successful source read unavailable; accepted source checkpoint unknown.
Owner codex/01a1bc19-7489-7c63-9dff-88dfe73a1645: current; last successful source read 2026-11-08T20:08:34.706115735Z; accepted source checkpoint 2026-11-08T17:01:23.855Z.
Owner codex/01a1bc5e-c2a5-70b0-a604-b33e8fb00e1e: pending; last successful source read 2026-11-08T19:46:27.385382763Z; accepted source checkpoint 2026-11-08T18:16:29.027Z.
Owner codex/01a1bc3a-a845-7d73-8665-ec31767ae319: current; last successful source read 2026-11-08T20:08:33.401823165Z; accepted source checkpoint 2026-11-08T17:43:22.688Z.
Owner codex/01a1bc9a-3851-7a43-8faf-bb765748798e: current; last successful source read 2026-11-08T20:08:33.42349735Z; accepted source checkpoint 2026-11-08T19:21:40.09Z.
Owner codex/01a1bc03-0564-75f3-a03d-f90a9ab91b63: current; last successful source read 2026-11-08T20:08:33.460418533Z; accepted source checkpoint 2026-11-08T16:41:08.672Z.
Owner codex/01a1bbf4-3a95-7283-951c-166f9c3037da: current; last successful source read 2026-11-08T20:08:33.489354903Z; accepted source checkpoint 2026-11-08T16:22:53.365Z.
Owner codex/01a1bc7e-c0db-7af2-b725-9c698832119a: current; last successful source read 2026-11-08T20:08:33.507921723Z; accepted source checkpoint 2026-11-08T18:51:09.626Z.
Owner codex/01a1bc9f-2567-7fc1-9200-9a5612c5ce48: current; last successful source read 2026-11-08T20:08:33.523134429Z; accepted source checkpoint 2026-11-08T19:26:05.722Z.
Additional owner freshness details remain in local capture diagnostics.

Retained recorded history: 1 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
