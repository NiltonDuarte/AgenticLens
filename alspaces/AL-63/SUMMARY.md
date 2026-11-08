# AL-63 capability audit

Codex historical hydration is supported and succeeded against the copied AL-55 ledger: 232 events across all five recorded owners. OpenCode v1 and v2 recorded-source hydration returned 27 and four events respectively. Live observation differs: the current reader uses v1 tables, fails for v2-only owners and has no Codex live implementation.

The user corrected the initial planning-only interpretation and requested execution of the original ticket. Its specification was restored byte-for-byte from eb7d169c; [iteration 02](iterations/02/PLAN.md) records the steering. [CAPABILITIES.md](CAPABILITIES.md) owns the full matrix, version/source identities, commands, evidence classes, findings and unexecuted cases. No application, tests, schemas, runtime scripts or defaults changed.

Disposable HTTP probes preserved ledger bytes and exposed complete/partial/unavailable coverage. With sources unavailable, AL-55 still returned five owners/four delegations and AL-62 returned 26 owners/three skills/25 delegations. Additional v2 CLI probes verified growing skill/task occurrences, unchanged repeats, malformed/missing-source preservation, other-owner identity and exactly one failed read among two calls. These close the test review's audit evidence gaps, without adding durable tests or claiming native parity. Native Codex backfill on the AL-55 copy also passed twice: all five identities/four delegations survived and repeated session arrays were identical.

Application revision f4152225 has unchanged application/test/build inputs from the successful integrated make at 0a8cfcee: backend race/integration, 190 frontend unit tests, 678 integration tests, static/release/install/capture/orchestration checks and multi-platform builds. Focused agent commands also passed. Current origin/main remains eb7d169c; no upstream application delta was introduced. Reuse applies only to those unchanged inputs. npm advisories, bundle-size and React act warnings remain outside this audit. Document checks, original-ledger hashes and final telemetry coverage are recorded at delivery.

Confirmed existing limits need follow-up: OpenCode metadata child links, v2 live reporting (FIX-18), configured Codex-home propagation (AL-70), Codex function-output error counters, live incompleteness diagnostics (AL-69), and mixed-family recovery policy (REV-7). No fixes were implemented or new tickets allocated. Architecture/Sol found the metadata issue; behavior/Sol found live/configuration/error-counter/diagnostic issues; data/Sol reproduced the known mixed-family limitation. Actor profiles and all review/evidence outcomes are retained in iteration 02; no later source-only approval refutes the probes.

Fresh active/native walkthroughs, native v2 positive skills/tasks/navigation, browser click-through and enabled v2 heuristic fallback remain unexecuted. AL-62 available-source hydration timed out once at a 25-second client limit; no performance defect is asserted. Original AL-55/AL-62 bytes remain intact. Audit completion must not be read as parity or defect-free source support.

All eight discovery reports are retained in the [consolidated review](iterations/02/REVIEW.md), with [candidate severity/discovery attribution](iterations/02/REVIEW_SUMMARY.json). Luna/data's native-verification INCOMPLETE and Sol/tests' initial evidence INCOMPLETE are preserved; parent probes supply the bounded evidence described above, while unexecuted native cases remain explicit matrix states. No product correctness approval is claimed.

Delivery verification: full make at audit commit 1e95b333 exited 0 after no-op origin/main integration. All required lanes executed, including backend race/integration, 190 frontend unit and 678 integration tests, static/boundary checks and multi-platform builds. Subsequent changes are audit documentation/registry only. Original acceptance was archived byte-for-byte in plan/done, AL-63 removed from the active index and AL-65 promoted to Ready. This closes discovery, not follow-up fixes or approval of defaults. Original-ledger hashes, report links, candidate attribution and source-only diff checks passed.

Final telemetry baseline: existing AL-63 root01a1bd1f-69c1-7043-9bf4-8d4fe9a53c9b must survive. Native source metadata independently identifies four direct Sol child sessions (architecture, behavior, data, tests), all gpt-6.1-sol/high. Four Luna service delegates have acknowledged model/effort and retained reports, but their native session/parent linkage is unavailable; do not manufacture owner rows or usage. Final ledger coverage is bounded to the native graph, with current-root duration provisional and native costs unavailable where unrecorded.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 8
- **Active Agent Duration:** 48m 55s (recorded duration including 657s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 188.8k tokens (42% of 450k healthy ceiling, 0 compactions) | Subagents peak 129.5k (Reviewer-behavior)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 24,948 reasoning tokens (21.1% of generated output)
- **Tool Health & Self-Correction:** 250 tool calls (23 self-corrections, 90.8% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 156,469 | 51,297 | 12,643,328 | 98.8% | 188.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 1 | 96,524 | 11,312 | 1,920,640 | 95.2% | 103.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 108,653 | 13,625 | 2,883,968 | 96.4% | 129.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 92,417 | 10,347 | 2,242,432 | 96.0% | 110.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 85,512 | 6,449 | 1,026,048 | 92.3% | 102.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **5** | **539,575** | **93,030** | **20,716,416** | **97.5%** | **188.8k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 5 | 539,575 | 93,030 | 20,716,416 | 97.5% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **5** | **539,575** | **93,030** | **20,716,416** | **97.5%** | **Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations)** |

Capture observation finalized: 2026-11-08T22:25:44.895732985Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 1; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 1; reserved bytes: 2646003. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1bd1f-69c1-7043-9bf4-8d4fe9a53c9b: current; last successful source read 2026-11-08T22:25:44.880012534Z; accepted source checkpoint 2026-11-08T22:25:44.784Z.

Final ingestion succeeded with literal nonce e3a902c7f6b1485d and explicit Codex/docs selection: schema6 contains the existing root plus exactly four source-proven native children, each retaining the exact parent/source pair. No existing owner was lost; four unlinked Luna service sessions remain excluded from native usage. Cost/caller identity and open-turn timing retain their source-reported uncertainty. Final whitespace, unchanged application scope, original acceptance bytes and original AL-55/AL-62 ledger hashes passed.
