# REV-5 audit and remediation summary

Baseline 36bea1f18f2256392428329136b3241e98d09854; initial working tree clean. User authorized repository review, especially code smells and production used only by tests. Assigned unused GEN-14 because GEN-13 belongs to an earlier documentation task. Branch GEN-14/code-smell-test-only-audit.

Completed repository-wide AST/reference screening (111 Go files; 89 frontend TS/TSX files) and focused manual inspection. Seven findings recorded in iterations/01/REVIEW.md, including one executable rollback reproduction. No application/test/runtime changes and no test deletions. Scope and confidence limits are explicit. No child agents launched. No matching planning specification/index entry exists to transition.

Verification: baseline make passed; after committing audit deliverables and fetching/merging origin/main (already up to date), integrated make passed at 236e1fd with only audit-text refinements pending. Go vet/race tests, TypeScript typecheck/ESLint, all 741 frontend tests across 35 files, frontend bundle and four platform builds passed. No gate stages skipped. The integrated Go run reused valid cache results from unchanged code. Embedded asset tree matched web/dist; host binary version command succeeded. git diff --check passed. No runtime performance claims.

Telemetry expectation: local Codex rollout metadata identifies root 01a19432-3529-73a0-a201-d8b1dbf00473 in this exact worktree. No children launched and no prior GEN-14 ledger or other-harness sessions to preserve. Final extraction uses source codex and ticket type chore. Snapshot covers through generate; later ledger verification, commit and PR delivery are outside that snapshot.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 5
- **Active Agent Duration:** 53m 19s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 155.6k tokens (78% of 200k healthy ceiling, 0 compactions) | Subagents peak 188.0k (Reviewer)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.2 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 1.3, Tester 1.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 1.9:1 (+532 test lines / +275 prod lines)
- **Code Delivered:** +807 / -2,545 lines across 51 code files (excluding docs)
- **Reasoning Effort:** 29,302 reasoning tokens (20.6% of generated output)
- **Tool Health & Self-Correction:** 356 tool calls (27 self-corrections, 92.4% success rate)

### Anomalies & Health Flags
- **Status:** 15 execution anomalies detected (9 CRITICAL, 6 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+116.0k tokens) — Context tokens surged by 116.0k to 116.0k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+138.3k tokens) — Context tokens surged by 138.3k to 138.3k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+140.9k tokens) — Context tokens surged by 140.9k to 140.9k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+148.5k tokens) — Context tokens surged by 148.5k to 148.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+151.1k tokens) — Context tokens surged by 151.1k to 151.1k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+145.9k tokens) — Context tokens surged by 145.9k to 145.9k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+149.6k tokens) — Context tokens surged by 149.6k to 149.6k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (11 cycles) — Subagent roles 'coder' and 'tester' alternated for 11 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.0k tokens) — Context tokens reached 160.0k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.2k tokens) — Context tokens reached 162.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.5k tokens) — Context tokens reached 163.5k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.2k tokens) — Context tokens reached 163.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (172.4k tokens) — Context tokens reached 172.4k, breaching the 160k warning ceiling.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 243,521 | 5,527 | 1,144,832 | 82.5% | 148.2k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **coder** | 6 | 659,473 | 52,011 | 20,631,552 | 96.9% | 184.5k | ELEVATED | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **tester** | 2 | 150,345 | 8,641 | 5,075,712 | 97.1% | 174.6k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer** | 2 | 131,800 | 5,729 | 3,437,952 | 96.3% | 188.0k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **rootagent** | 1 | 138,362 | 40,891 | 19,113,344 | 99.3% | 155.6k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **13** | **1,323,501** | **112,799** | **49,403,392** | **97.4%** | **188.0k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-astra** | 13 | 1,323,501 | 112,799 | 49,403,392 | 97.4% | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **Total** | **13** | **1,323,501** | **112,799** | **49,403,392** | **97.4%** | **Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations)** |

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

## Upstream integration review (iteration 02 continuation)

User requested merging main and reviewing the new code. Initial working tree clean at aa86f4b71abe6cbb91a43e855b91bbc11fb31515; PR92 open. Fetched main47af923 (AL-57 unavailable-cost/schema-v5 work) and started merge. Conflicts require preserving both the upstream explicit cost-evidence contracts and REV-5 subtraction/validation guarantees. This is routine integration under REV-5, not a new ticket or acceptance iteration. Prior verification remains valid for unchanged code; integrated cost/wire/migration changes require fresh checks. Incoming AL-57 artifacts are inherited main history, not independently edited task work.

### Integration decisions

Resolved all seven conflicted paths, including deleting the obsolete schemas.go monolith in favor of main's split DTOs. Preserved upstream unavailable/partial/observed cost evidence and v4/v5 migration behavior. Kept REV-5 removed APIs absent, removed two dead helpers reintroduced by the split and one new test-only summary wrapper, and retained meaningful assertions through active APIs. Frontend now composes general shape and cost-specific validation with one body read, accepts legitimate nullable/omitted costs, and preserves diagnostic recovery.

Independent tester found no functional integration regression. Backend/domain race checks and 113 combined client/warning cases passed. Actual merged Go HTTP responses for isolated v4 observed-zero, v4 missing-cost and v5 partial-cost fixtures passed frontend decoders. See INTEGRATION_BACKEND.md, INTEGRATION_DOMAIN.md, INTEGRATION_FRONTEND.md and INTEGRATION_TESTING.md for exact decisions, assertions and limits.

### Integration final verification

Reviewer APPROVED with no blocking findings. Merge commit 49cf350 has parents aa86f4b71abe6cbb91a43e855b91bbc11fb31515 and 47af923f5763dc937372ada776c1043c2dd3e7a6. Required integrated make exited 0: Go vet/race tests, TypeScript typecheck/ESLint, 781 frontend tests across 35 files, canonical asset build and four Linux/Darwin platform binaries passed. No gate stages skipped; valid cached Go results reused.

Generated source/embed trees match. The rebuilt standalone binary served byte-identical HTML/JS/CSS in a temporary empty adopting workspace with an unavailable database; see INTEGRATION_EMBEDDED.json. Host version and whitespace checks passed. No provider-cost inference, production source hydration, financial estimate, performance benchmark or browser geometry claim is made. Only evidence/telemetry changes follow the tested merge.

### Integration telemetry expectation

Preserve the seven previously recorded sessions. Local source metadata establishes 13 sessions: original root plus the earlier six children and six new integration children. Final refresh remains source codex, ticket REV-5, type fix; upstream persistence now uses schema v5 cost evidence. Expected source IDs:
- 01a19432-3529-73a0-a201-d8b1dbf00473: root
- 01a19444-c3d4-7393-82a8-cf821f617656: tester
- 01a19458-6b61-7d43-9f53-6a7ce3f20564: coder
- 01a19458-9464-76d3-8523-00bc5fe35d6d: coder
- 01a1945a-4b16-7300-a14f-1c23c66d2404: tester
- 01a19458-7f96-7e11-81d3-f4eaa397b4b4: coder
- 01a1943e-5f6b-7473-a5b9-44dee687d643: planner
- 01a1945c-a81a-77d0-8601-106060ab6faf: reviewer
- 01a19447-3ec5-7b82-a101-b7b73ee6414a: reviewer
- 01a19440-559b-7413-b93f-78d1f15d4149: coder
- 01a19457-610e-7da3-970c-d80161b44af2: planner
- 01a1943f-cbe8-7d92-91f7-5048d8eb00d6: coder
- 01a1943f-e378-7681-9560-6a653db69307: coder

Snapshot covers through the final generate command; subsequent verification, commit and PR push are outside that snapshot.

Final integration ledger verified: schema v5; all 13 expected Codex sessions with correct parent links, including all seven previous identities; explicit invocation cost evidence. No other-harness sessions existed to preserve.
