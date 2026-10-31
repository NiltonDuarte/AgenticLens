# REV-5 audit and remediation summary

Baseline 36bea1f18f2256392428329136b3241e98d09854; initial working tree clean. User authorized repository review, especially code smells and production used only by tests. Assigned unused GEN-14 because GEN-13 belongs to an earlier documentation task. Branch GEN-14/code-smell-test-only-audit.

Completed repository-wide AST/reference screening (111 Go files; 89 frontend TS/TSX files) and focused manual inspection. Seven findings recorded in iterations/01/REVIEW.md, including one executable rollback reproduction. No application/test/runtime changes and no test deletions. Scope and confidence limits are explicit. No child agents launched. No matching planning specification/index entry exists to transition.

Verification: baseline make passed; after committing audit deliverables and fetching/merging origin/main (already up to date), integrated make passed at 236e1fd with only audit-text refinements pending. Go vet/race tests, TypeScript typecheck/ESLint, all 741 frontend tests across 35 files, frontend bundle and four platform builds passed. No gate stages skipped. The integrated Go run reused valid cache results from unchanged code. Embedded asset tree matched web/dist; host binary version command succeeded. git diff --check passed. No runtime performance claims.

Telemetry expectation: local Codex rollout metadata identifies root 01a19432-3529-73a0-a201-d8b1dbf00473 in this exact worktree. No children launched and no prior GEN-14 ledger or other-harness sessions to preserve. Final extraction uses source codex and ticket type chore. Snapshot covers through generate; later ledger verification, commit and PR delivery are outside that snapshot.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 0m 0s (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 77.0k tokens (39% of 200k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 872 reasoning tokens (9.4% of generated output)
- **Tool Health & Self-Correction:** 27 tool calls (1 self-corrections, 96.3% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **rootagent** | 1 | 72,389 | 8,362 | 1,263,744 | 94.6% | 77.0k | HEALTHY | $0.00 |
| **Total** | **1** | **72,389** | **8,362** | **1,263,744** | **94.6%** | **77.0k** | **HEALTHY** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-astra** | 1 | 72,389 | 8,362 | 1,263,744 | 94.6% | $0.00 |
| **Total** | **1** | **72,389** | **8,362** | **1,263,744** | **94.6%** | **$0.00** |

Final ledger verified: exactly the expected Codex root, no children, no previous ledger to preserve.

## Iteration 02: REV-5 reclassification and remediation

User steering: "track this session under a REV ticket. fix the findings". This is the same task/session reclassified from GEN-14 to previously unused REV-5, not a switch to unrelated work. Historical audit artifacts retain their original labels as evidence; the workspace is relocated intact. User explicitly authorizes correction of F1-F7 and safe subtraction of confirmed unused production code. Audit-only scope is superseded for this iteration; preserve runtime contracts and useful test assertions. Original ledger session identities must survive final refresh. REV is explicitly requested by the user despite not being in current configured prefixes.

Baseline e8d9003ea6388d1a750273478ec572863e965a72; initial working tree clean. PR #91 was open at start; remediation will use a REV-5 branch/PR and supersede the audit PR.

### Implementation

CLI/adapters: configuration-pair validation and deferred cleanup/rollback now preserve originals and report restoration errors. Removed the unused live buffer, obsolete merge helpers, and unreachable adapter/formatter/schema APIs. Domain: removed test-only alternate APIs and retargeted useful assertions to active calculation, filtering and migration paths. Frontend: removed unused fetchers/formatters and added concrete validation of active HTTP responses; malformed optional diagnostics remain visible as warnings without discarding valid sibling data. Omitted artifact descriptions normalize to the Go zero value after independent tester feedback.

Candidate disposition and assertion recipients are recorded in iterations/02/BACKEND.md, DOMAIN.md and FRONTEND.md. Actual revision caching, implicit codecs, public routes, legacy wire shapes and configuration acceptance are preserved. Updated setup/operator guidance and the root configuration comment to describe the legacy trace-buffer setting accurately. Independent testing/review and integrated verification are still in progress.
