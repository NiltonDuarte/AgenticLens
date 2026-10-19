# Summary of Changes: AL-31 Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests

## Overview

Ticket AL-31 decouples structural session hierarchy (`ParentID == nil` for root coordination session vs `ParentID != nil` for delegated child subagents) from agent persona role names across the domain engine, boundary adapters, markdown formatters, and parity regression test suites.

Following user steering across Iterations 01, 02, and 03:
- The domain engine completely eliminates persona string blacklists and role shims (`IsSubagentRole` deleted entirely).
- Invocations with `ParentID == nil` exclusively govern root coordination metrics (`OrchestratorPeakTokens`, `OrchestratorSaturationPct`).
- Invocations with `ParentID != nil` exclusively govern delegated child subagent metrics (`SubagentsPeakTokens`, `SubagentsPeakRole`, `RoleEffortDistributionPct`, `SubagentTurnVelocity`).
- Candidate subagent roles extracted from child invocations are preserved in `OrderSubagentRoles` without blacklist filtering, placing canonical roles first and sorting custom roles alphabetically.
- Rework loop detection in `detectSubagentReworkLoops` indexes session identifiers to `InvocationFact.ParentID != nil`, structurally ignoring root coordination turns during cycle evaluation without relying on role heuristics or fallback shims.
- Git branch inference tests in `inference_test.go` and `analyzer_test.go` use isolated transient branch fixtures with `t.Cleanup`, resolving CI test failures on fresh checkouts.

---

## Changes Grouped by Layer

### 1. Domain Layer (`backend/internal/domain/`)

- `backend/internal/domain/aggregations.go`:
  - In `ComputeWorkflowEfficiency`, eliminated all role blacklist filtering.
  - Partitioned invocations strictly by structural hierarchy:
    - Root Invocations (`ParentID == nil`): Governs `rootPeak`, populating `OrchestratorPeakTokens` and calculating `OrchestratorSaturationPct` regardless of role persona string (`general`, `build`, `orchestrator`, or custom).
    - Delegated Child Subagents (`ParentID != nil`): Normalizes whitespace roles (defaulting to `"unassigned"` if empty), accumulating `subCosts`, `totSubCost`, `subPeaks`, and `roleTurns`.
    - Single-session tickets without child invocations evaluate to `SubagentsPeakTokens = 0`, `SubagentsPeakRole = "none"`, empty `RoleEffortDistributionPct`, and zero `SubagentTurnVelocity`.
- `backend/internal/domain/enums.go`:
  - Completely removed `IsSubagentRole(role string) bool`.
  - Refactored `OrderSubagentRoles` to accept candidate roles directly from child invocations, deduplicating non-empty trimmed roles, prioritizing canonical roles (`planner`, `coder`, `tester`, `reviewer`), and sorting remaining custom roles alphabetically without filtering against persona blacklists.
- `backend/internal/domain/anomalies.go`:
  - In `detectSubagentReworkLoops`, mapped `EventFact.SessionID` to `isChildSession` (`inv.ParentID != nil`). Root coordination turns (`ParentID == nil`) are structurally ignored from the role sequence. Eliminated all fallback calls to `IsSubagentRole`.
- `backend/internal/domain/aggregations_test.go`:
  - Added comprehensive table-driven test suite `TestComputeWorkflowEfficiency_HierarchyDecoupling`.
  - Added `TestComputeWorkflowEfficiency_AdversarialStructuralHierarchy` verifying inverted roles (root labeled `"coder"`, child labeled `"orchestrator"`), multi-root sessions without children, and whitespace child role normalization to `"unassigned"`.
  - Removed obsolete `TestIsSubagentRole`.
- `backend/internal/domain/enums_test.go`:
  - Added `TestOrderSubagentRoles_OrderingAndPreservation`.
  - Removed obsolete `TestIsSubagentRole_DeprecationAndBackwardsCompatibility`.
- `backend/internal/domain/anomalies_test.go`:
  - Synchronized all unit test fixtures in `TestDetectAnomalies_SubagentReworkLoop` with explicit `Invocations` hierarchy (`ParentID: nil` for root, `ParentID: &rootID` for child subagents) and mapped `SessionID`.
  - Updated `TestDetectAnomalies_DeterministicSortingAllRules` with root and child invocation hierarchy.
  - Added boundary tests for empty ledgers, unmapped session IDs, orphaned child sessions, and root-only sessions.

### 2. Adapters Layer (`backend/internal/adapters/`)

- `backend/internal/adapters/formatters.go`:
  - Updated `FormatMarkdownMetricsTable` to render `RoleEffortDistributionPct` directly from the decoupled domain map without hardcoded persona exclusion filters.
  - Formats empty subagent distributions as `N/A` and zero-subagent context health cleanly as `Subagents peak 0 (None)`.
- `backend/internal/adapters/formatters_test.go`:
  - Added unit test `TestFormatMarkdownMetricsTable_SingleRootAndZeroSubagents` validating markdown rendering and zero-emoji compliance.
- `backend/internal/adapters/parity_test.go`:
  - Updated reference baseline expectations for tickets with authentic `"general"` root sessions (`AL-1`, `AL-10`, `AL-11`, `AL-12`, `GEN-1`, `GEN-2`, `PO-1`, `PO-2`, `PO-3`) in `PeakByRole`.
  - Updated `TestFleetSummaryAndComparisonParity` fleet role costs to authentically partition spend between `orchestrator` ($7.8645) and `general` ($10.9622).
  - Added `TestStructuralClassificationParityOnWorkspaces` verifying 100% compliance across all 37 workspace ledgers.
- `backend/internal/adapters/git/inference_test.go`:
  - Implemented `createTestBranch(t, branchName)` helper creating transient branches at HEAD with automatic `t.Cleanup` removal.
  - Updated `TestInferTicketType_BranchPrefixResolution` and `TestInferTicketType_Precedence_BranchOverConfiguredPrefix` to use hermetic fixture branches, preventing reliance on ambient branches in fresh CI checkouts.
- `backend/internal/adapters/git/analyzer_test.go`:
  - Updated `TestInferenceOnActiveWorkspace` with transient fixture branch `feat/AL-11-ci-fixture` and `t.Cleanup` ensuring deterministic pass rates in CI.

### 3. Presentation Layer (`web/src/`)

- Preserved existing property contracts (`orchestrator_peak_tokens`, `subagents_peak_tokens`, `subagents_peak_role`) across TypeScript interfaces and widgets.
- All 398 frontend unit tests in `web/` passed with zero regressions.

---

## Verification Status

- Static Analysis: `go vet ./...` passed with zero warnings.
- Frontend Lint & Typecheck: `tsc --noEmit` and `eslint` passed with zero warnings.
- Backend Tests: `go test -race -count=1 ./...` passed across all 10 packages with zero data races.
- Frontend Tests: `npm test` passed 398/398 Vitest assertions.
- Parity Tests: `TestCrossEngineParityOnWorkspaceLedgers`, `TestFleetSummaryAndComparisonParity`, and `TestStructuralClassificationParityOnWorkspaces` passed 100%.
- Git Inference Tests: `TestInferTicketType_BranchPrefixResolution`, `TestInferTicketType_Precedence_BranchOverConfiguredPrefix`, and `TestInferenceOnActiveWorkspace` pass deterministically in isolation.
- Single Binary Build: `make build` compiled multi-platform executables into `bin/`.
- Invariants: Zero `any` in Go and TypeScript, zero emojis.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 4h 31m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 174.4k tokens (87% of 200k healthy ceiling, 2 compactions) | Subagents peak 167.7k (Coder)
- **Role Effort Distribution (Subagents):** Planner 20.5%, Coder 48.2%, Tester 19.9%, Reviewer 11.4%
- **Subagent Turn Velocity & Thrashing:** 71.7 avg turns/run (peak: 161 turns) | Planner 53.3, Coder 116.7, Tester 69.0, Reviewer 47.7
- **Test-to-Production Code Ratio:** 13.3:1 (+1,426 test lines / +107 prod lines)
- **Code Delivered:** +1,533 / -130 lines across 9 code files (excluding docs)
- **Reasoning Effort:** 298,680 reasoning tokens (61.1% of generated output)
- **Tool Health & Self-Correction:** 1,098 tool calls (19 self-corrections, 98.3% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (4 CRITICAL, 3 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 75 (session: ses_eaac24625ffeeMwQsqm70U1eLS), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k in turn 109 (session: ses_eaa823bf1ffeYqu3A3kt4HetRS), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (174.4k tokens) — Context tokens reached 174.4k in turn 118 (session: ses_eaad954a1ffek1dSxjZ0I4Cy6Y), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+86.5k tokens) — Context tokens surged by 86.5k tokens in turn 202 (session: ses_eaad954a1ffek1dSxjZ0I4Cy6Y, total: 86.5k tokens, threshold: 80000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+61.9k tokens) — Context tokens surged by 61.9k tokens in turn 83 (session: ses_eaac24625ffeeMwQsqm70U1eLS, total: 61.9k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+74.0k tokens) — Context tokens surged by 74.0k tokens in turn 119 (session: ses_eaad954a1ffek1dSxjZ0I4Cy6Y, total: 74.0k tokens, threshold: 40000)
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: mcp-gateway_mcp-find (33.3%) — Tool 'mcp-gateway_mcp-find' failed 3 out of 9 calls (33.3% failure rate, threshold: >20%)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 3,238,156 | 55,044 | 16,396,299 | 83.5% | 174.4k | ELEVATED | $4.02 |
| **planner** | 3 | 1,183,029 | 25,927 | 9,198,961 | 88.6% | 145.3k | ELEVATED | $2.01 |
| **coder** | 3 | 2,518,413 | 62,417 | 29,615,549 | 92.2% | 167.7k | ELEVATED | $4.74 |
| **tester** | 3 | 993,330 | 32,180 | 12,769,564 | 92.8% | 120.8k | ELEVATED | $1.95 |
| **reviewer** | 3 | 640,473 | 14,240 | 6,517,209 | 91.1% | 82.3k | HEALTHY | $1.12 |
| **Total** | **13** | **8,573,401** | **189,808** | **74,497,582** | **89.7%** | **174.4k** | **ELEVATED** | **$13.85** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 13 | 8,573,401 | 189,808 | 74,497,582 | 89.7% | $13.85 |
| **Total** | **13** | **8,573,401** | **189,808** | **74,497,582** | **89.7%** | **$13.85** |
