# fix(FIX-21): keep delegation child links across capture refresh

## Outcome

Automatic capture no longer deletes accepted `child_session_id` links when a pass re-reads a parent but not all of its children. The AL-71 incident lost 7 of 9 links this way.

**Link retention.** In capture, the rule pairs each delegation of a refreshed Codex parent with its accepted record. Delegations are grouped by agent type, description and invocation time, and paired by position within each group.

- **Kept:** an empty fresh link keeps its accepted child.
- **Cleared:** a child whose own metadata names another parent is cleared.
- **Replaced:** a different child replaces the accepted link only when its parent is confirmed and no other fresh delegation also links it.
- **Reported:**
  - accepted links that cannot be placed (a smaller fresh group, an unmatched key, a displaced child) are reported;
  - both kept and unplaced links count as "Unresolved child evidence", so the pass stays visibly partial.

**Where the rule runs.** The rule is pure domain logic. Capture applies it at three points:

- at each owner snapshot commit, so an interrupted pass keeps its links;
- when it annotates or regroups a family;
- once in `generate --capture-auto`, against the committed ledger, which protects a ledger copied to a machine without the earlier capture state.

**Discovery stays within budget.** Header identification was spending 95 to 97 % of every 512 MiB pass on unrelated files, so the parent and its children were never extracted together. Capture now:

- skips child candidates created before their parent, using the UUIDv7 instant in the file name with a 5 s tolerance;
- skips local date directories that ended more than 14 h before the parent;
- skips exact-lookup files whose name carries another session ID, while identity still comes only from `session_meta`;
- reads each header once per pass and charges only the bytes actually read;
- reserves a quarter of the budget for identification, released to extraction at the end of the pass.

On a synthetic incident-scale store (1,907 rollouts, 2.75 GiB), pass 1 extracts the whole family with 9.5 % of the budget and pass 2 links all 9 children.

**Archived unassigned rollouts.** An unassigned row whose rollout moved to `archived_sessions/` is recovered only when its own metadata proves the same session, store, project and parent. Anything else is isolated with a typed reason instead of failing every worker pass.

**Unchanged:** manual `generate`, backfill, live hydration, OpenCode and Claude Code.

Product decisions and steering are recorded in [design.md](design.md#steering):

- D1 to D7;
- the UUIDv7 window;
- the reserved quarter;
- position alignment, chosen by the user on 2026-11-25.

## Verification and review

**Tested revision:** `34068194`. The branch base is `133100bb`, which is current `origin/main`, so no merge was needed.

**Integrated gate:** `GOFLAGS=-timeout=90m make` passed on its second run, covering:
- install;
- the vet, lint, release, installation, distribution and orchestration checks;
- the Go race suites in 18 packages;
- 231 frontend unit tests and 829 frontend integration tests;
- engine and manager builds for linux and darwin, amd64 and arm64.

The first run stopped on one frontend integration test, `TicketComparisons`, which timed out at 5 s under host load. That file passed 3 of 3 times on its own, and this change does not touch `web/`.

**Focused runs on the same code** (see [TESTS.md](evidence/TESTS.md)):
- the domain, codex, capture and `cmd/agentic-lens` suites;
- capture under `-race`, run repeatedly;
- the incident-scale large-store test.

**OpenSpec:** strict validation passes.

**Telemetry:** the Claude Code ledger was generated with `--source claude` from root session `a27e309f-c6cf-4b8e-8a78-c7568c3fffdd`. It holds 37 sessions: the root and its 36 native subagents.
- The subagents are 1 planner, 1 coder, 2 tester runs and 32 reviewer runs.
- Every child links to the root, and all 36 root delegations carry their `child_session_id`.
- Served models match the review profiles: 16 runs on claude-opus-5-5 and 16 on claude-sonnet-5-5. The root, planner, coder and testers ran on claude-opus-5-5.
- Cost is unavailable for every invocation, because the Claude Code source records no price evidence.

**Review.** The change went through four review rounds. Each round covered behavior, data integrity, architecture and test effectiveness on both configured profiles, claude-opus-5-5/medium and claude-sonnet-5-5/high: eight runs per round.

| Round | Candidate | Result |
|---|---|---|
| 1 | `0f73f759` | Requested changes |
| 2 | `86eb6dae` | Requested changes |
| 3 | `5d4df7ee` | Requested changes |
| 4 | `34068194` | Approved by all eight runs |

Confirmed defects and who found them:

- **Interrupted pass:** it could strip registry links. Found by reviewer-data on both profiles (DATA-01).
- **Resumed child sweep:** a sweep stopped at the ceiling was closed instead of pending. Found by reviewer-behavior on Sonnet (BEH-01).
- **Refused links:** they were dropped silently. Found by reviewer-data on both profiles and reviewer-behavior on Sonnet (DATA-02).
- **Partly re-read groups:** links were lost after one report. Found by six runs across behavior, data and architecture (R2-01). This led to the position-alignment decision.
- **Silent replacement losses:** a replacement could drop an accepted child without a report. Found by reviewer-architecture on Opus, reviewer-data on Sonnet and reviewer-behavior on both profiles (R3-01).

The tester's own adversarial cases found one idempotence defect before review. Every fix has a regression test that fails when the fix is reverted, using mutations M1 to M41. All findings, reporters and dispositions are in [REVIEW_SUMMARY.json](evidence/REVIEW_SUMMARY.json). Test evidence is in [TESTS.md](evidence/TESTS.md).

## Limits

- **Reported once:** refused or unmatched accepted links are reported for one pass and not placed afterwards.
- **Shared task names (R2-P1):** when two delegations share a task name and only one child is read, the source resolver points both delegations at that child. The flaw predates FIX-21 and the fix is a follow-up. FIX-21 keeps the accepted links, but a duplicated fresh child that was never accepted is neither placed nor reported.
- **Permanent partial:** a machine that can never re-read the children stays partial. Coverage can also be over-reported as partial within one pass.
- **Small budgets:** a very small `max_source_bytes` can stall identification, which shows as pending.
- **Busy exit:** after the 2 s retention deadline, any registry error exits 75, which the worker retries, rather than 1. No data is lost.
- **CLI exit wiring:** the exit wiring of the retention step is verified by reading only. The pure outcome helper is tested.
- **Duplicated lookup (ARCH-01, accepted debt):** the capture exact lookup duplicates the manual and live lookup rules. Unifying them is a follow-up.
- **Contradiction warning:** capture never surfaces the contradiction warning. This is pre-existing.

## Pull request

### Problem

Automatic capture refreshed a Codex parent without all of its children and replaced the accepted parent wholesale. That deleted accepted delegation `child_session_id` links (7 of 9 in AL-71) without any report.

Two causes in the same area made the loss persistent:

- header identification consumed almost the whole byte budget on unrelated rollouts;
- an unassigned rollout moved to `archived_sessions/` failed every worker pass.

### What changed

- A pure domain rule, `domain.RetainAcceptedChildLinks`, keeps accepted links by provenance and by position within each key group.
  - It changes a link only on fresh, confirmed evidence.
  - It reports every link it cannot place.
- The rule is applied at each capture owner snapshot, in family annotation and archive regrouping, and once in `generate --capture-auto` against the committed ledger.
- Kept and unplaced links count as "Unresolved child evidence", so the pass is recorded as partial.
- Capture discovery:
  - prunes candidates by UUIDv7 creation time and by session ID in the file name;
  - caches each header once per pass and charges actual bytes;
  - reserves a quarter of the budget for identification.
- Unassigned archived rollouts are recovered only with metadata proof. Otherwise they are isolated.
- Manual `generate`, backfill and live paths are unchanged.

### Validation

- Integrated `make` passed on `34068194`: Go race suites, 231 frontend unit and 829 frontend integration tests, and four platform builds.
- Mutation-backed regression tests (M1 to M41) cover every fix. Each mutation fails its test when the fix is reverted.
- On a synthetic 2.75 GiB, 1,907-rollout store, the root and all 9 children are extracted within the budget, and all 9 links stay intact.
- Four review rounds on behavior, data, architecture and tests, each on claude-opus-5-5/medium and claude-sonnet-5-5/high. The final round was approved by all eight runs. Findings are in [REVIEW_SUMMARY.json](evidence/REVIEW_SUMMARY.json).

### Limits

See [Limits](#limits) in this file. The follow-ups are the R2-P1 shared task-name resolution, unifying the exact lookups (ARCH-01) and narrowing the busy marker (R4-01).

🤖 Generated with [Claude Code](https://claude.com/claude-code)

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 9
- **Active Agent Duration:** 15h 34m (recorded duration including 53s provisional)
- **Context Window Health:** BLOATED — Root agent peak 436.4k tokens (97% of 450k healthy ceiling, 0 compactions) | Subagents peak 794.3k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.6 avg turns/run (peak: 12 turns) | Planner 12.0, Coder 6.0, Tester 3.5, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 2.1:1 (+4,703 test lines / +2,192 prod lines)
- **Code Delivered:** +6,895 / -163 lines across 29 code files (excluding docs)
- **Reasoning Effort:** 920,052 reasoning tokens (29.8% of generated output)
- **Tool Health & Self-Correction:** 1,789 tool calls (11 self-corrections, 99.4% success rate)

### Anomalies & Health Flags
- **Status:** 19 execution anomalies detected (5 CRITICAL, 14 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.5k tokens) — Context tokens reached 450.5k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.4k tokens) — Context tokens reached 450.4k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (5 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (6 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (5 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 5 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.1k tokens) — Context tokens reached 254.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.6k tokens) — Context tokens reached 250.6k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.8k tokens) — Context tokens reached 250.8k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (260.5k tokens) — Context tokens reached 260.5k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> planner (4 cycles) — Subagent roles 'coder' and 'planner' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> planner (3 cycles) — Subagent roles 'tester' and 'planner' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (4 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> planner (3 cycles) — Subagent roles 'coder' and 'planner' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> planner (4 cycles) — Subagent roles 'tester' and 'planner' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 220 | 126,950 | 19,536,456 | 100.0% | 345.1k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 352 | 219,657 | 48,350,400 | 100.0% | 486.1k | BLOATED (108%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 2 | 600 | 497,552 | 106,142,047 | 100.0% | 794.3k | BLOATED (177%) | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Unknown** | 1 | 466 | 150,354 | 41,946,166 | 100.0% | 436.4k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 8 | 374 | 192,307 | 16,088,941 | 100.0% | 146.4k | HEALTHY | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **reviewer-behavior** | 8 | 574 | 290,513 | 32,096,120 | 100.0% | 249.7k | ELEVATED | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **reviewer-data** | 8 | 492 | 329,664 | 25,572,594 | 100.0% | 223.8k | HEALTHY | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **reviewer-tests** | 8 | 600 | 363,558 | 33,216,824 | 100.0% | 196.6k | HEALTHY | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **37** | **3,678** | **2,170,555** | **322,949,548** | **100.0%** | **794.3k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 37 unavailable / 37 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-opus-5-5** | 21 | 2,504 | 1,526,479 | 255,664,791 | 100.0% | Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations) |
| **claude-sonnet-5-5** | 16 | 1,174 | 644,076 | 67,284,757 | 100.0% | Unavailable (0 observed, 0 partial, 16 unavailable / 16 invocations) |
| **Total** | **37** | **3,678** | **2,170,555** | **322,949,548** | **100.0%** | **Unavailable (0 observed, 0 partial, 37 unavailable / 37 invocations)** |
<!-- agentic-lens:metrics:end -->
