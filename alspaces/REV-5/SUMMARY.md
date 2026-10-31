# REV-5 audit and remediation summary

Baseline 36bea1f18f2256392428329136b3241e98d09854; initial working tree clean. User authorized repository review, especially code smells and production used only by tests. Assigned unused GEN-14 because GEN-13 belongs to an earlier documentation task. Branch GEN-14/code-smell-test-only-audit.

Completed repository-wide AST/reference screening (111 Go files; 89 frontend TS/TSX files) and focused manual inspection. Seven findings recorded in iterations/01/REVIEW.md, including one executable rollback reproduction. No application/test/runtime changes and no test deletions. Scope and confidence limits are explicit. No child agents launched. No matching planning specification/index entry exists to transition.

Verification: baseline make passed; after committing audit deliverables and fetching/merging origin/main (already up to date), integrated make passed at 236e1fd with only audit-text refinements pending. Go vet/race tests, TypeScript typecheck/ESLint, all 741 frontend tests across 35 files, frontend bundle and four platform builds passed. No gate stages skipped. The integrated Go run reused valid cache results from unchanged code. Embedded asset tree matched web/dist; host binary version command succeeded. git diff --check passed. No runtime performance claims.

Telemetry expectation: local Codex rollout metadata identifies root 01a19432-3529-73a0-a201-d8b1dbf00473 in this exact worktree. No children launched and no prior GEN-14 ledger or other-harness sessions to preserve. Final extraction uses source codex and ticket type chore. Snapshot covers through generate; later ledger verification, commit and PR delivery are outside that snapshot.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 4
- **Active Agent Duration:** 24m 59s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 127.0k tokens (64% of 200k healthy ceiling, 0 compactions) | Subagents peak 158.3k (Coder)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 1.3 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 1.7, Tester 1.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 1.8:1 (+478 test lines / +264 prod lines)
- **Code Delivered:** +742 / -2,750 lines across 48 code files (excluding docs)
- **Reasoning Effort:** 16,778 reasoning tokens (18.5% of generated output)
- **Tool Health & Self-Correction:** 218 tool calls (14 self-corrections, 93.6% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (2 CRITICAL).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+116.0k tokens) — Context tokens surged by 116.0k to 116.0k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 96,871 | 3,388 | 567,296 | 85.4% | 98.9k | HEALTHY | $0.00 |
| **coder** | 3 | 185,048 | 35,816 | 11,250,176 | 98.4% | 158.3k | ELEVATED | $0.00 |
| **tester** | 1 | 63,780 | 4,636 | 2,172,416 | 97.1% | 144.6k | ELEVATED | $0.00 |
| **reviewer** | 1 | 74,383 | 3,099 | 1,617,664 | 95.6% | 154.5k | ELEVATED | $0.00 |
| **rootagent** | 1 | 119,449 | 26,808 | 11,107,584 | 98.9% | 127.0k | ELEVATED | $0.00 |
| **Total** | **7** | **539,531** | **73,747** | **26,715,136** | **98.0%** | **158.3k** | **ELEVATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-astra** | 7 | 539,531 | 73,747 | 26,715,136 | 98.0% | $0.00 |
| **Total** | **7** | **539,531** | **73,747** | **26,715,136** | **98.0%** | **$0.00** |

## Iteration 02: REV-5 reclassification and remediation

User steering: "track this session under a REV ticket. fix the findings". This is the same task/session reclassified from GEN-14 to previously unused REV-5, not a switch to unrelated work. Historical audit artifacts retain their original labels as evidence; the workspace is relocated intact. User explicitly authorizes correction of F1-F7 and safe subtraction of confirmed unused production code. Audit-only scope is superseded for this iteration; preserve runtime contracts and useful test assertions. Original ledger session identities must survive final refresh. REV is explicitly requested by the user despite not being in current configured prefixes.

Baseline e8d9003ea6388d1a750273478ec572863e965a72; initial working tree clean. PR #91 was open at start; remediation will use a REV-5 branch/PR and supersede the audit PR.

### Implementation

CLI/adapters: configuration-pair validation and deferred cleanup/rollback now preserve originals and report restoration errors. Removed the unused live buffer, obsolete merge helpers, and unreachable adapter/formatter/schema APIs. Domain: removed test-only alternate APIs and retargeted useful assertions to active calculation, filtering and migration paths. Frontend: removed unused fetchers/formatters and added concrete validation of active HTTP responses; malformed optional diagnostics remain visible as warnings without discarding valid sibling data. Omitted artifact descriptions normalize to the Go zero value after independent tester feedback.

Candidate disposition and assertion recipients are recorded in iterations/02/BACKEND.md, DOMAIN.md and FRONTEND.md. Actual revision caching, implicit codecs, public routes, legacy wire shapes and configuration acceptance are preserved. Updated setup/operator guidance and the root configuration comment to describe the legacy trace-buffer setting accurately. Independent testing and review completed; reviewer APPROVED. Final integrated verification passed as recorded below.

### Final verification

Reviewed implementation commit eb2bba8; origin/main merge was already up to date. Integrated make exited 0: Go vet/race suite, TypeScript typecheck/ESLint, 745 frontend tests across 35 files, canonical asset build, and Linux/Darwin amd64/arm64 binaries passed. No gate stages skipped. Valid unchanged Go test cache results were reused. Only final ticket evidence and telemetry changed afterward.

Original forced-init defect reproduced against the rebuilt CLI now preserves exact original bytes and leaves zero staged/backup files; see rollback-verification.json. A temporary empty adopting workspace with unavailable database verified that the standalone binary serves byte-identical built index/JS/CSS; see embedded-verification.json. Generated source/embed asset trees match. Host version command and git diff --check passed.

F1-F7 are fixed within the approved boundaries. Candidate/test mappings and review limitations remain in iteration 02 reports; no claim of exhaustive dead-code proof, injected OS restoration failure, browser geometry, provider source hydration or runtime performance profiling. No planning specification/index entry exists to transition.

### Telemetry continuity

This is the original session renamed from GEN-14 to REV-5, not a new task. Existing root identity must remain. Local rollout metadata establishes the root plus six child sessions (planner, three coders, tester, reviewer). All are Codex; existing ledger has no other-harness sessions. Expected IDs:
- 01a19432-3529-73a0-a201-d8b1dbf00473: root
- 01a19444-c3d4-7393-82a8-cf821f617656: tester
- 01a1943e-5f6b-7473-a5b9-44dee687d643: planner
- 01a19447-3ec5-7b82-a101-b7b73ee6414a: reviewer
- 01a19440-559b-7413-b93f-78d1f15d4149: coder
- 01a1943f-cbe8-7d92-91f7-5048d8eb00d6: coder
- 01a1943f-e378-7681-9560-6a653db69307: coder

Final source extraction uses explicit codex source, REV-5 ticket and fix type. Snapshot covers through generate; subsequent ledger checks, commit and PR delivery are outside that snapshot.

Final ledger verified: REV-5/fix, original root retained, exactly all six expected children, source codex throughout, and correct parent links. No previous other-harness sessions existed to preserve.
