# GEN-20: Unique worktree directory identifiers

## Delivered behavior

Capture coordination directories use the linked Git worktree administrative name qualified by its observed identity, or a persisted 16-hex random identifier when no linked-worktree identifier is available. Concurrent first opens share the committed mapping for the same scope and isolate distinct scopes. Bounded collisions skip existing filesystem entries without adopting or overwriting them.

Adapters and CLI consumers resolve the mapping consistently for read-only opens, aliases, provisional previews, installation, policy checks, diagnostics, workers and supported relocation/resume. Internal numeric SQL IDs remain unchanged. Domain calculations, portable ledger authority and frontend behavior are unchanged.

## Scope and baseline

- Initial revision: dc44619ae15586fe1194f872743033f523635c92; working tree was clean. Ticket branch: chore/GEN-20-unique-worktree-identifiers; configured workspace root: alspaces.
- The user approved the baseline contract and iteration 01 plan, then clarified that database compatibility is unnecessary because the database-introducing version is unreleased. [Iteration 02](iterations/02/PLAN.md) supersedes numeric-path preservation, protocol upgrades and old-binary fencing in the historical [contract](CONTRACT.md). Registry protocol remains 2 with the mapping in its current schema. Existing unrelated protocol-1 recovery remains supported.
- Unsupported local layouts fail explicitly without rewriting the database. Existing numeric layouts require fresh capture state; this ticket performs no automatic reset or migration.
- Initial implementation: 67d2a8bdd7039e7638e5dfb68fc435993252433b. Corrected revision: 13a6ff38c97e9c9d3019d76a0944ce46afe1814e. Both were synchronized with origin/main; upstream remained at the baseline.

## Verification

Coder and tester: gpt-6.1-sol, high reasoning effort. The coder executed focused capture/CLI race checks; the tester inspected fixture reachability, verified test discovery, and reused those executed results. [Test evidence](iterations/02/TEST_REPORT.md) distinguishes inspection, discovery and execution.

The coordinator (gpt-6.1-sol, high reasoning effort) verified the corrections: the new direct-root and physical-alias regression failed before the path fix; affected capture/CLI race tests passed afterward. The Git naming mutation that passed the original test now fails the stronger administrative-name assertion. Unknown/foreign aliases, redirected coordinates, concurrent allocation, fallback stability, installation and relocation coverage still pass.

Integrated make passed on the original reviewed revision: backend races, frontend typecheck/lint and 790 tests, Python release/installation/distribution checks, frontend embedding and all configured platform builds. Final make also passed on corrected revision 13a6ff38c97e9c9d3019d76a0944ce46afe1814e; its native binary reports that commit.

## Review findings and attribution

Eight independent first-pass reviews covered behavior, data/failure handling, architecture and tests on 67d2a8b, each under acknowledged gpt-6.1-sol/high and gpt-6-luna/medium settings. The Sol behavior, data and architecture reviews requested changes for one shared defect; their Luna counterparts found no blocking issues. The Sol test review approved with one useful suggestion; the Luna test review approved without findings. A mistyped base SHA in two Luna reports did not change their reviewed diff: the coordinator verified that the reviewed candidate parent is the requested baseline.

- Severity 5: ordinary configured bases beneath a directory named worktrees were misclassified as coordination aliases. Independently discovered by behavior, data and architecture reviewers, each gpt-6.1-sol/high. Fixed by the coordinator, gpt-6.1-sol/high, by requiring the projects/<name>/worktrees/<directory-name> shape; direct and physical-alias before/after tests verify the fix while alias rejection remains intact.
- Severity 3 suggestion: the Git naming oracle accepted randomized substitutes. Discovered by the test reviewer, gpt-6.1-sol/high, using an isolated mutation. The coordinator, gpt-6.1-sol/high, strengthened the existing concurrency test using independently observed Git administrative names and hexadecimal qualification, then confirmed the mutation is rejected.

The host requires small review corrections to be fixed locally without further delegates. The coordinator therefore verified these narrow corrections directly and carried forward the unaffected independent review evidence. No reviewer is claimed to have reviewed 13a6ff3; final consolidated approval is APPROVED after the corrected revision passed make. Compact findings and every discovery attribution are retained in [REVIEW_SUMMARY.json](iterations/02/REVIEW_SUMMARY.json). Earlier implementation self-review also corrected an arbitrary probe-child check in root validation; focused tests passed after that correction.

## Directory-name terminology correction

The user explicitly requested replacing ambiguous worktree token terminology in the table and source. The initial follow-up working tree was clean at e2df9649de644f7dde1406757a85e9441a9a14d6. Commit 46cb6e84860e9cdb0e686559c2d502efab8e8788 renames the SQL column to directory_name, Go locals to directoryName, and associated helpers, tests, diagnostics and comments consistently. Genuine request tokens and directory allocation behavior are unchanged. This is a review correction within iteration 02; no compatibility migration is introduced.

The coder and tester (gpt-6.1-sol/high) verified complete SQL/caller consistency and preserved test reachability. Focused capture and CLI race tests passed. Narrow behavior reviewers under acknowledged gpt-6.1-sol/high and gpt-6-luna/medium both approved the six-file correction without findings. Prior data, architecture and test coverage carries forward for unchanged behavior. The integrated source revision is 29811ad56e8f4fc128ce788ede8687a3cae4b713, merged with official main 33ddada05c64cb8fa41701fd9fce3376ff83ff10; that integration changes no backend file. Final make passed on this revision, including backend races, all 790 frontend tests, static/Python checks, embedding and platform builds. Consolidated naming review: APPROVED with no new findings; the existing compact review summary retains the earlier resolved findings and all discovery credit.

## Telemetry and limits

For the naming follow-up, incremental ingestion into isolated scratch state rejected the existing ledger because scratch state has no historical ownership receipts. The follow-up therefore re-extracts the original Codex sources into fresh staged artifacts and verifies that every existing session identity and parent link is preserved before replacing generated artifacts. No receipt or provenance is fabricated, and the original ledger remains intact if fresh extraction or verification fails.

Final ingestion succeeded with an explicit source codex, ticket type chore and ticket title. The configured local registry uses the unsupported earlier layout; isolated scratch capture state was used, the temporary local configuration was removed, and a byte hash confirmed the original database was unchanged. The built native CLI takes generate flags directly; dispatcher-only --al-offline/--al-project flags are not supported by the native executable.

Source inspection identifies root 01a1a395-ebd9-7820-8931-b7c6a91a5906 and native planner, coder and tester child sessions. Ten T3 reviewer sessions are recorded by Codex as separate roots without native parent lineage. Ledger membership was verified against real Codex metadata: exactly the root and three expected native children are present, all source codex with the correct parent links. The ten separately rooted T3 reviewer sessions are omitted because their source metadata lacks native parent lineage; their usage is not included. Pricing is unavailable for the four included invocations, so cost remains unavailable. The ledger is explicitly incomplete for total ticket usage. No prior GEN-20 ledger or other-harness sessions existed at baseline. Raw rollouts and detailed working run/finding records remain outside Git.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 20
- **Active Agent Duration:** 1h 2m (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 199.9k tokens (44% of 450k healthy ceiling, 0 compactions) | Subagents peak 148.2k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 3.0 avg turns/run (peak: 4 turns) | Planner 2.0, Coder 3.0, Tester 4.0
- **Test-to-Production Code Ratio:** 1.3:1 (+410 test lines / +308 prod lines)
- **Code Delivered:** +718 / -57 lines across 17 code files (excluding docs)
- **Reasoning Effort:** 32,101 reasoning tokens (24.9% of generated output)
- **Tool Health & Self-Correction:** 257 tool calls (20 self-corrections, 92.2% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (3 WARNING).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.7k tokens) — Context tokens reached 160.7k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (4 cycles) — Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (4 cycles) — Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 134,214 | 6,216 | 795,904 | 85.6% | 81.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 240,708 | 24,173 | 5,870,464 | 96.1% | 148.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 82,971 | 9,319 | 3,637,504 | 97.8% | 136.3k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 172,421 | 57,224 | 23,136,896 | 99.3% | 199.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **4** | **630,314** | **96,932** | **33,440,768** | **98.2%** | **199.9k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 4 | 630,314 | 96,932 | 33,440,768 | 98.2% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **4** | **630,314** | **96,932** | **33,440,768** | **98.2%** | **Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations)** |
