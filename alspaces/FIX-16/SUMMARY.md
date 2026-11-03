# FIX-16: Restore release candidate publishing

## Delivered correction

The existing source-build installation test now controls VERSION, GIT_COMMIT and BUILD_DATE in its subprocess environment. It exercises both Git-derived defaults and explicit candidate metadata through real make verify-build and native execution. Candidate version, commit and build date are checked together. Both modes retain installed-cache inode/bytes, selected launcher, printed native verification and wrapper-absence assertions. Only scripts/tests/test_installation.py changed outside this workspace; workflows, Makefile and shared fixtures are unchanged.

## Grounding, approval and integration

Initial working tree was clean on FIX-16/restore-release-publishing at dc44619ae15586fe1194f872743033f523635c92. Workspace root is configured as alspaces. No existing FIX-16 workspace, matching specification or ticket-index entry existed; no lifecycle move or downstream promotion was needed.

[GitHub run 41733288460](https://github.com/NiltonDuarte/AgenticLens/actions/runs/41733288460) computed v3.0.0-rc.122+dc44619 and then failed the installation assertion comparing it to Git's default version. Job records confirm publication was skipped after Static Analysis failed. The exact-revision test and workflow were checked through GitHub. Local reproduction initially lacked embedded assets; after installing dependencies and building canonical assets, the unchanged test reproduced the recorded version mismatch.

Planning feedback requested original wording and direct evidence. Iteration 01 removed the copied request and hypothesis framing, linking confirmed GitHub records instead. Scope and acceptance boundaries stayed unchanged; explicit approval preceded coder dispatch. No extra iteration was created for this clarification.

Implementation commit and reviewed/tested revision: 93341525d8faea2060a471f5fd277b1246e59b7a. Frozen test blob: d85643b9aa7234df0a7d324cea8b0f52b72f9035. origin/main remained dc44619; integration reported already up to date. Subsequent changes are ticket documentation and telemetry only.

## Verification and review

- Coder gpt-6.1-sol/high reproduced the original failure and passed the corrected focused case with recorded RC metadata exported and with metadata unset.
- Tester gpt-6.1-sol/high passed installation 10/10, release-workflow 14/14 and capture-distribution 15/15 under the recorded candidate environment. A disposable mutation dropping explicit metadata failed the candidate assertion.
- Parent gpt-6.1-sol/high ran full make with VERSION=v3.0.0-rc.122+dc44619, GIT_COMMIT=dc44619 and BUILD_DATE=2026-11-03T22:44:16Z: exit 0. Go vet/race tests, frontend type/lint checks, all three Python suites, 36 frontend files / 790 tests, canonical assets and five-platform native compilation passed. Generated frontend and embedded asset trees match exactly.
- Host Linux amd64 make release-build release-package release-verify passed: tar archive and native checksums verified, native version commands returned the explicit candidate metadata. Host-only packaging printed the existing unmatched ZIP-glob warning because no Windows archive was requested; ZIP packaging is not claimed verified.
- Independent test-effectiveness and behavior/integration reviews each ran under gpt-6.1-sol/high and gpt-6-luna/medium, all APPROVED with no findings. Exact models/efforts were observed in Codex source-session turn contexts; T3 also acknowledged both Luna settings. Sol test-effectiveness review independently confirmed disposable mutations ignoring explicit VERSION and retaining ambient metadata fail the respective native assertions. Reviewers preserved shared source.

Parent consolidated verdict: APPROVED after integrated checks and host artifact verification, covering the unchanged frozen source above. No unresolved findings or material scoped review gaps remain. [Compact review record](iterations/01/REVIEW_SUMMARY.json) retains the empty findings set; detailed completed reports were removed after consolidation. Discovery and verification actors are distinct above.

Hosted publication and uploaded downloadable assets remain unverified: this branch is being delivered by PR for integration. No release was published, tag changed, or release PR merged. Release packaging was exercised only for Linux amd64; all five native compilation targets passed in make. Local verification logs and raw source transcripts remain outside Git.

## Telemetry

Generation is deferred to the final pre-commit step, with explicit Codex source, a literal nonce and type fix. Expected source set: this root, five native child sessions and two T3-delegated reviewer roots. Source metadata confirms all eight sessions use this exact checkout. The two independent reviewer roots were associated with FIX-16 through the supported capture request/bind commands; no hook installation or source editing occurred. There was no previous FIX-16 ledger or other-harness history to preserve. Validation results follow generation.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 10
- **Active Agent Duration:** 13m 34s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 115.2k tokens (58% of 200k healthy ceiling, 0 compactions) | Subagents peak 94.7k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0, Coder 1.0, Tester 1.0, Reviewer-behavior 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 6,774 reasoning tokens (15.1% of generated output)
- **Tool Health & Self-Correction:** 146 tool calls (13 self-corrections, 91.1% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 57,575 | 3,713 | 442,496 | 88.5% | 59.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 60,842 | 2,980 | 1,396,864 | 95.8% | 85.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 26,984 | 1,880 | 617,216 | 95.8% | 94.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 3 | 199,015 | 21,916 | 7,752,832 | 97.5% | 115.2k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-behavior** | 1 | 39,818 | 3,353 | 306,304 | 88.5% | 54.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 66,072 | 4,177 | 375,936 | 85.1% | 48.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **8** | **450,306** | **38,019** | **10,891,648** | **96.0%** | **115.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 2 | 88,318 | 5,471 | 399,232 | 81.9% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **gpt-6.1-sol** | 6 | 361,988 | 32,548 | 10,492,416 | 96.7% | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **Total** | **8** | **450,306** | **38,019** | **10,891,648** | **96.0%** | **Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations)** |

Capture observation finalized: 2026-11-03T23:03:38.517792185Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 2; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 2; reserved bytes: 938518. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1a3a1-9fc9-7682-b2ca-e91bffa1fec6: current; last successful source read 2026-11-03T23:03:38.486685968Z; accepted source checkpoint 2026-11-03T22:58:28.531Z.
Owner codex/01a1a3a2-87b8-7281-a6db-9df85e216c2b: current; last successful source read 2026-11-03T23:03:38.507116278Z; accepted source checkpoint 2026-11-03T22:59:17.227Z.

Final telemetry validation: all eight expected source session IDs are present, including the root, five native children and two independently rooted delegated reviewers. Source is Codex and ticket type is fix; no prior sessions were replaced. Generation completed without source errors; capture diagnostics report no pending, failed or unresolved child evidence. Cost evidence is unavailable. Configured Go/TypeScript code roots exclude this Python test change, so zero code counts do not mean no source changed. The installed shared launcher is absent; the freshly built native was used directly from this checkout. Native help confirms source/nonce/title flags; wrapper-only --al-offline/--al-project flags were rejected before extraction and omitted on the successful native invocation. An explicit ticket title was required and supplied. Raw logs/transcripts remain outside Git.
