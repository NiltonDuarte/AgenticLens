# GEN-19: Release candidates reuse Release Please's next version

## Current iteration 03 status

The user explicitly requested the proposed further reduction. Captured PR metadata now remains valid when main or the PR advances; parent/provenance/history scans and repeated identity checks are removed. One exact stable-tag lookup remains before publication. Review also found a pre-existing tag identity issue, fixed with the action's captured-SHA target input and an assertion. All focus/model cells and affected behavior rechecks approve; 14 focused tests pass. Full integrated make is pending; the earlier iteration 02 evidence below is historical until final consolidation.

## Request and accepted steering

The user reported v2.0.0-rc candidates after stable v2.0.0, although the next release was 2.1.0, and requested the next GEN ticket. GEN-19 was free across workspaces, index, remote workspace tree and branch search. Initial working tree was clean at 036d808. Branch: chore/GEN-19-rc-target-version; configured root: alspaces; conventional type: chore. The generated initial GEN-13 branch name was not used as ticket allocation evidence.

Iteration 01 implemented a second read-only Release Please calculation. The user then asked to reuse the already calculated version in the open release PR and explicitly approved the simpler approach on 2026-11-03. CONTRACT.md records that accepted supersession; [iteration 02 plan](iterations/02/PLAN.md) records changed authority/trigger, preserved guarantees and verification impact. The current implementation below supersedes the independent engine; iteration 01 approval and evidence do not certify the replacement.

## Delivered behavior

Automatic RC builds now follow successful Release Please runs on main. Both jobs check out the upstream run's head SHA, rather than workflow_run's default-branch GITHUB_SHA; manual dispatch remains available with main-commit eligibility checks. Source detection reads the actual checked-out commit message.

The RC workflow discovers the same-repository pending generated release PR, verifies its one-parent source commit, reads the root version from the manifest at its immutable head, and appends the existing run/SHA suffix. Missing/stale PRs and superseded source commits skip explicitly; API errors, ambiguous/invalid data and already stable targets fail before publication. Main/PR/stable eligibility is checked again before publish/prune.

The standalone resolver, pinned engine, lockfile and engine tests are removed. The final diff against origin/main adds no scripts/release directory or Makefile dependency changes. Existing gh, Bash, Git and runner tools perform selection; release version calculation remains entirely with Release Please. CONTRIBUTING.md describes sequencing and local fixture-test prerequisites. No Go domain/adapters, schemas, frontend behavior, stable publishing or historical releases changed.

Coder gpt-6.1-sol/high implemented the workflow/removal; tester gpt-6.1-sol/high owned executable release tests and corrections. Parent owns integration, evidence, telemetry and PR delivery.

## Verification

Reviewed/tested implementation revision: 6fb6318c34fc5847375863cddf99a755b03c3960. Final source/deletion fingerprint: 2f9057a71a4cbbb703600b4268758894363152ead629189a08c54925d3da3ab0. Frozen production/test hashes remained unchanged through integration. Fetch/merge origin/main reported already up to date.

- Parent full make passed after integration: Go vet/race tests (valid cached results retained for unchanged packages), frontend type/lint checks, 35 frontend files / 783 tests, 14 Python release tests, 12 installation tests, frontend embedding and standalone builds. No conditional stage was silently skipped.
- Tester and final test reviewers independently ran make test-release-workflow:14passed, zero skips. Six stable tests are preserved. RC tests execute actual shell on disposable Git/API fixtures and assert successful environment/output handoff, immutable manifest refs, source parent, explicit skips/errors, stable rejection, event/checkout wiring and final publication gates.
- Actual captured production gh/gojq filters execute against raw local API JSON using installed gh and a local HTTP fixture, without GitHub login. Disposable wrong-version, wrong PR filter, wrong manifest root, missing pagination and metadata-only filter mutations fail. Deleted engine calculations map to manifest-target consumption; version calculation policy belongs to the existing Release Please workflow.
- Behavior/data/architecture reviewers gpt-6.1-sol/high independently ran read-only production shell against real GitHub APIs in disposable checkouts of main/036d808. PR #55 at head 1300741296d069b1282d89a6192219bf13fbb4f9 had sole parent 036d808 and manifest version 2.1.0. Complete metadata and eligibility steps returned should_publish=true; VERSION was v2.1.0-rc.114+036d808 for run 114, with correct commit/date/PR fields. No publish/prune command was invoked. Data reviewer also verified API-route failure probes block publication.
- Parent git diff --check passed, standalone version --json ran, and web/dist exactly matched backend/internal/adapters/web/dist. Generated assets remain ignored. The development make build uses git describe by design; RC builds receive the explicit validated VERSION from workflow metadata.

## Review, fixes and attribution

Final consolidated verdict: APPROVED after required integrated make. Initial discovery covered behavior/integration, data/failure, architecture/maintenance and test effectiveness under both gpt-6.1-sol/high and gpt-6-luna/medium. All requested identities were verified in Codex turn_context records. Initial test-effectiveness Luna requested changes; the other seven initial cells approved, including one nonblocking Sol test observation. Paired test rechecks verified corrections; Sol recheck found one additional nonblocking oracle issue, which was also fixed and verified by a final pair. Production-focused approvals carry forward on unchanged workflow/deletion/Makefile hashes; corrections affected tests and two contributor prerequisite/target lines only. No blockers or unresolved findings remain.

[Iteration 02 compact findings](iterations/02/REVIEW_SUMMARY.json) retains every candidate and discovering tuple:

- TEST-02 severity 4: test-effectiveness reviewer gpt-6-luna/medium discovered that canned post-filter API values bypassed the real response-filter oracle. Tester gpt-6.1-sol/high fixed this with actual captured gh/gojq evaluation over raw local JSON.
- TEST-03 severity 3: test-effectiveness reviewer gpt-6.1-sol/high discovered that omitted stable pagination survived tests. Tester gpt-6.1-sol/high added required-pagination argument coverage; the mutation now fails.
- TEST-04 severity 3: test-effectiveness reviewer gpt-6.1-sol/high discovered during re-review that route-keyed collection overwrote an earlier metadata query. Tester gpt-6.1-sol/high retained/evaluated every distinct query; a metadata-only mutation now fails.

Verification is separate from discovery: final test reviewers gpt-6.1-sol/high and gpt-6-luna/medium approved all three corrections and 14/14 tests. Other responsibilities/profiles found no additional defects. Parent validated revision continuity, integrated checks and final disposition. Completed detailed working reports/run records were compacted only after all findings and reporter tuples were retained.

[Iteration 01 compact findings](iterations/01/REVIEW_SUMMARY.json) remains historical evidence: REL-01 severity 2 was shared discovery by behavior and architecture reviewers gpt-6.1-sol/high; its Node-support discrepancy is resolved by removing the independent dependency tree. TEST-01 severity 3 was discovery by test reviewer gpt-6.1-sol/high; its successful-output oracle gap is resolved by the replacement shell success/mutation assertions. Coder/tester gpt-6.1-sol/high implemented those resolutions; current test reviewers verified the replacement. No accepted debt from that mechanism remains.

## Remaining limits and delivery

Local shell/filter evidence does not establish hosted workflow_run scheduling, action checkout execution, artifact upload/download or retention deletion. A required-pagination flag oracle is not a hosted Link-header pagination test. Release packaging was not exercised in iteration 02; the iteration 01 Windows ZIP probe was blocked by unavailable zip and is not claimed successful. GitHub reads and publication are separate operations; final checks are best effort, not an atomic transaction.

No releases were published, renamed, retagged or deleted, and the stable release PR was not merged. No GEN-19 planning specification/index entry exists, so no lifecycle move or downstream promotion is required. PR #98 is updated with the replacement; future eligible RCs use the PR's version after merge. Raw transcripts and probe/build logs remain outside Git.

Telemetry is refreshed only at the final iteration step with source codex and ticket type chore. Previous twelve Codex sessions must be preserved; the complete source tree now contains 24 sessions (root, planner/coder/tester and 20 review sessions across both iterations). Existing ledger has no other-harness sessions. Final ledger validation follows extraction.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 6
- **Active Agent Duration:** 1h 24m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 206.9k tokens (103% of 200k healthy ceiling, 0 compactions) | Subagents peak 132.5k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.2 avg turns/run (peak: 4 turns) | Planner 2.0, Coder 2.0, Tester 4.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 75,229 reasoning tokens (27.8% of generated output)
- **Tool Health & Self-Correction:** 447 tool calls (19 self-corrections, 95.7% success rate)

### Anomalies & Health Flags
- **Status:** 9 execution anomalies detected (6 CRITICAL, 3 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.9k tokens) — Context tokens reached 200.9k, breaching the 200k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (15 cycles) — Subagent roles 'coder' and 'tester' alternated for 15 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (10 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (7 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (6 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 6 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+60.3k tokens) — Context tokens surged by 60.3k to 60.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 149,464 | 5,448 | 872,576 | 85.4% | 81.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 88,700 | 15,324 | 4,328,832 | 98.0% | 117.2k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 67,285 | 24,334 | 5,796,992 | 98.9% | 132.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 160,627 | 83,699 | 21,167,232 | 99.2% | 206.9k | BLOATED (103%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 4 | 248,639 | 12,082 | 1,272,064 | 83.6% | 69.2k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 4 | 198,093 | 15,412 | 1,652,480 | 89.3% | 67.6k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-data** | 4 | 173,280 | 13,984 | 1,452,160 | 89.3% | 63.1k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-tests** | 8 | 336,573 | 24,724 | 2,672,640 | 88.8% | 60.3k | HEALTHY | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **24** | **1,422,661** | **195,007** | **39,214,976** | **96.5%** | **206.9k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 24 unavailable / 24 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 10 | 481,818 | 23,978 | 2,185,856 | 81.9% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **gpt-6.1-sol** | 14 | 940,843 | 171,029 | 37,029,120 | 97.5% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **Total** | **24** | **1,422,661** | **195,007** | **39,214,976** | **96.5%** | **Unavailable (0 observed, 0 partial, 24 unavailable / 24 invocations)** |

Final telemetry validation: exactly 24 source-verified Codex sessions extracted: root 01a1a138-f579-7481-80b2-39db0205af01 plus 23 expected children. All twelve prior session IDs are preserved; all ledger/source IDs match exactly and source is codex throughout. No extraction errors were reported. No other-harness facts existed to preserve. Ticket type is chore. Raw transcripts remain outside Git.

Telemetry interpretation: cost evidence is unavailable. Configured production/test roots exclude CI tooling, so zero rendered code-line counts do not mean no implementation was delivered. Raw anomaly flags are retained; interleaved role activity and legitimate correction/re-review rounds do not establish the analyzer's reported number of product rework cycles. Telemetry analyzer behavior is outside this ticket.
