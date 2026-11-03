# GEN-19: Align release candidate versions with the upcoming release

## Request, approval and baseline

The user reported v2.0.0-rc candidates after stable v2.0.0 and requested the next GEN ticket. GEN-19 was free in the workspace, index, remote workspace tree and branch search. Initial working tree was clean at 036d808; the generated original branch name was not used as ticket allocation evidence. Dedicated branch: chore/GEN-19-rc-target-version. Configured workspace root: alspaces; conventional type: chore.

The user explicitly approved CONTRACT.md and iteration 01 PLAN.md on 2026-11-03. Scope stayed within RC resolution, CI dependency setup and relevant regression tests; stable publishing and historical releases were not changed.

## Delivered behavior

The old RC workflow selected a prior tag and stripped its suffix, preserving the already released numeric version. GitHub evidence showed stable v2.0.0, subsequent v2.0.0-rc candidates and Release Please PR #55 targeting 2.1.0.

The CI-only resolver now uses pinned Release Please 17.3.0 and the existing configuration/last-released manifest to calculate the upcoming release. It preserves v<target>-rc.<run>+<sha>, checks candidate/main identity around calculation and before publishing, skips superseded/no-candidate runs, and rejects malformed or already stable targets. The workflow retains existing build, prerelease and retention behavior, with explicit pull-request read access for GraphQL. Node tooling is not a production binary dependency.

No domain, adapter, schema or frontend behavior changed. The coder (gpt-6.1-sol/high) implemented the workflow/resolver; the tester (gpt-6.1-sol/high) owned release tests and their Makefile entry. Maintainer guidance is in scripts/release/README.md.

## Verification

Reviewed/tested revision: 2bcb689db2e74a4a88c36e7f299fb4d0970d10ee; production/test fingerprint: 3a5ea9f89c48818862332b7da06019145544ec9ab2edcb830cae9a28b752765b. All nine frozen file hashes remained unchanged through integration. Fetching and merging origin/main reported already up to date; no integration code delta invalidated review evidence.

- Parent: full make passed after integration, including Go vet/race tests, frontend type/lint checks, 35 frontend test files / 783 tests, 8 Python release tests, 12 installation tests, 13 Node release tests, frontend embedding and standalone builds. No stage was silently skipped.
- Tester: make test-release-workflow passed, 8 Python and 13 Node tests, zero skips. Real pinned Manifest fixtures distinguish feature, patch, breaking and Release-As rules; isolated adapter/workflow checks cover stale runs, stable targets, errors and metadata validation. Existing stable revision guards remain intact.
- Parent and behavior reviewer (gpt-6.1-sol/high): all 13 Node tests passed on Node v20.20.2, matching CI's major; host default is v24.22.1. Coder and behavior reviewer separately performed actual read-only API resolution at main/036d808, returning 2.1.0. Test reviewer (gpt-6.1-sol/high) executed the actual read-only CLI and verified VERSION=v2.1.0-rc.114+036d808, commit/date fields and should_publish=true in the output files.
- Behavior reviewer (gpt-6.1-sol/high): disposable Linux/Windows release builds passed and host binary JSON preserved the exact supplied RC metadata. Linux tar contents were verified. Windows ZIP creation was blocked by unavailable zip, so that probe did not establish full packaging/checksum verification.
- Parent: git diff --check passed, standalone binary version --json succeeded, and web/dist exactly matched backend/internal/adapters/web/dist. Generated assets remain ignored under the established tracking policy.

## Review and remaining limits

Final consolidated verdict: APPROVED after the required integrated make gate. All four responsibilities (behavior/integration, data/failure handling, architecture/maintenance, test effectiveness) completed independent discovery under both gpt-6.1-sol/high and gpt-6-luna/medium. Exact model/effort pairs were verified in Codex turn_context source records. All eight scoped reviews approved; no blocking correctness findings remained. Parent owns this consolidated decision, not the child verdicts alone.

Retained findings: [compact review summary](iterations/01/REVIEW_SUMMARY.json).

- REL-01, severity 2: behavior/integration and architecture reviewers, both gpt-6.1-sol/high, independently observed that locked git-file-utils 3.2.1 declares Node >=22 while CI selects Node20. The Node20 tests and actual API calculation succeed; no runtime failure was established. This nonblocking declared-support discrepancy is accepted for this fix. Release tooling maintainers own alignment on the next dependency/runtime update; README documents pin maintenance.
- TEST-01, severity 3: test-effectiveness reviewer gpt-6.1-sol/high found that removing the successful CLI environment write survives both current suites. Four core version/guard mutations were caught. Actual successful CLI output was independently verified for this revision. This bounded automated-oracle gap is accepted; release tooling maintainers own durable success-path coverage in a future tooling revision. Discovery and current-behavior verification remain distinct.

Data/failure reviewers and the four gpt-6-luna/medium reviewers reported no additional findings. Completed detailed reports/run records were compacted only after retaining both findings and all three discovering persona/model/effort tuples. Accepted debt is not hidden by cleanup.

Hosted Actions execution, release publication/upload/download and retention deletion were not exercised. The final GitHub read and publication are separate operations and cannot be made atomic by these guards. No releases were published, renamed, retagged or deleted. The open stable release PR was not merged. API or resolver errors stop publication; there is no guessed-version fallback.

## Delivery and telemetry

No GEN-19 planning specification or index entry exists, so no lifecycle move or downstream promotion is required. Parent owns commits, integration, push and PR delivery. Telemetry is generated from Codex only at the final iteration step; expected source tree is one root plus eleven children (planner, coder, tester and eight reviews). This is a new ledger with no prior other-harness sessions to preserve. Raw transcripts and probe/build logs remain outside Git.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 30m 33s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 116.2k tokens (58% of 200k healthy ceiling, 0 compactions) | Subagents peak 103.4k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0, Coder 1.0, Tester 1.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 24,832 reasoning tokens (23.0% of generated output)
- **Tool Health & Self-Correction:** 239 tool calls (13 self-corrections, 94.6% success rate)

### Anomalies & Health Flags
- **Status:** 4 execution anomalies detected (3 CRITICAL, 1 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (15 cycles) — Subagent roles 'coder' and 'tester' alternated for 15 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (10 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (7 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 7 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+60.3k tokens) — Context tokens surged by 60.3k to 60.3k tokens.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 68,451 | 3,912 | 563,328 | 89.2% | 70.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 78,907 | 8,679 | 2,640,384 | 97.1% | 103.4k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 38,899 | 8,625 | 2,073,344 | 98.2% | 95.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 104,012 | 29,521 | 7,325,952 | 98.6% | 116.2k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 2 | 150,488 | 6,845 | 834,304 | 84.7% | 69.2k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 105,994 | 8,712 | 1,056,896 | 90.9% | 67.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 80,928 | 8,079 | 926,208 | 92.0% | 63.1k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 95,995 | 8,573 | 959,104 | 90.9% | 60.0k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **12** | **723,674** | **82,946** | **16,379,520** | **95.8%** | **116.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 197,503 | 10,675 | 1,008,384 | 83.6% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 8 | 526,171 | 72,271 | 15,371,136 | 96.7% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **12** | **723,674** | **82,946** | **16,379,520** | **95.8%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

Telemetry validation: extracted exactly the source-verified Codex root 01a1a138-f579-7481-80b2-39db0205af01 plus eleven expected children (twelve sessions total), all source codex, with no extraction error reported. IDs exactly matched the rollout source metadata. GEN-19 is classified as chore. This new ledger had no existing other-harness sessions to preserve. Raw transcripts remain outside Git.

Telemetry interpretation limits: cost evidence is unavailable for all twelve invocations. Configured production/test roots exclude CI tooling, so the rendered zero code-line counts do not mean no implementation was delivered. The raw ledger retains rework-loop anomaly flags, but source metadata and invocation counts show one planner, one coder, one tester and one discovery run for each of the eight review/model cells; the flags do not establish fifteen actual implementation cycles. No telemetry analyzer change is included in this ticket.
