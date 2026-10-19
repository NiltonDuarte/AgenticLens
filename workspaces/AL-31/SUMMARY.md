# Summary of Changes: AL-31 Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests

## Overview

Ticket AL-31 decouples structural session hierarchy (`ParentID == nil` for root coordination session vs `ParentID != nil` for delegated child subagents) from agent persona role names across the domain engine, boundary adapters, markdown formatters, and parity regression test suites.

Following user steering in Iteration 02, the domain engine completely eliminates persona string blacklists (`""`, `"orchestrator"`, `"general"`, `"build"`, `"root"`) in favor of pure structural hierarchy:
- Invocations with `ParentID == nil` exclusively govern root coordination metrics (`OrchestratorPeakTokens`, `OrchestratorSaturationPct`).
- Invocations with `ParentID != nil` exclusively govern delegated child subagent metrics (`SubagentsPeakTokens`, `SubagentsPeakRole`, `RoleEffortDistributionPct`, `SubagentTurnVelocity`).
- Candidate subagent roles extracted from child invocations are preserved in `OrderSubagentRoles` without blacklist filtering, placing canonical roles first and sorting custom roles alphabetically.
- Rework loop detection in `detectSubagentReworkLoops` indexes session identifiers to `InvocationFact.ParentID != nil`, structurally ignoring root coordination turns during cycle evaluation.

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
  - Refactored `OrderSubagentRoles` to accept candidate roles directly from child invocations, deduplicating non-empty trimmed roles, prioritizing canonical roles (`planner`, `coder`, `tester`, `reviewer`), and sorting remaining custom roles alphabetically without filtering against persona blacklists.
  - Added deprecation notice on `IsSubagentRole` indicating that session classification is purely structural via `ParentID != nil`.
- `backend/internal/domain/anomalies.go`:
  - In `detectSubagentReworkLoops`, mapped `EventFact.SessionID` to `InvocationFact.ParentID != nil`. Root coordination turns (`ParentID == nil`) are structurally ignored from the role sequence, avoiding false cycle disruption between worker subagents.
- `backend/internal/domain/aggregations_test.go`:
  - Added table-driven unit tests for `IsSubagentRole` and `OrderSubagentRoles`.
  - Added comprehensive table-driven test suite `TestComputeWorkflowEfficiency_HierarchyDecoupling`.
  - Added `TestComputeWorkflowEfficiency_AdversarialStructuralHierarchy` verifying inverted roles (root labeled `"coder"`, child labeled `"orchestrator"`), multi-root sessions without children, and whitespace child role normalization to `"unassigned"`.
- `backend/internal/domain/enums_test.go`:
  - Added `TestOrderSubagentRoles_OrderingAndPreservation` and `TestIsSubagentRole_DeprecationAndBackwardsCompatibility`.
- `backend/internal/domain/anomalies_test.go`:
  - Added adversarial test scenarios verifying that interleaved root events and fallback root invocations are structurally ignored during ping-pong rework loop detection.

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
- Single Binary Build: `make build` compiled multi-platform executables into `bin/`.
- Invariants: Zero `any` in Go and TypeScript, zero emojis.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 3h 17m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 174.4k tokens (87% of 200k healthy ceiling, 2 compactions) | Subagents peak 167.7k (Coder)
- **Role Effort Distribution (Subagents):** Planner 19.6%, Coder 50.0%, Tester 19.4%, Reviewer 11.0%
- **Subagent Turn Velocity & Thrashing:** 81.3 avg turns/run (peak: 161 turns) | Planner 56.5, Coder 142.5, Tester 72.5, Reviewer 53.5
- **Test-to-Production Code Ratio:** 13.3:1 (+1,426 test lines / +107 prod lines)
- **Code Delivered:** +1,533 / -130 lines across 9 code files (excluding docs)
- **Reasoning Effort:** 208,699 reasoning tokens (59.4% of generated output)
- **Tool Health & Self-Correction:** 826 tool calls (14 self-corrections, 98.3% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (3 CRITICAL, 3 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 75 (session: ses_eaac24625ffeeMwQsqm70U1eLS), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k in turn 109 (session: ses_eaa823bf1ffeYqu3A3kt4HetRS), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (174.4k tokens) — Context tokens reached 174.4k in turn 118 (session: ses_eaad954a1ffek1dSxjZ0I4Cy6Y), breaching 160k critical ceiling
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+61.9k tokens) — Context tokens surged by 61.9k tokens in turn 83 (session: ses_eaac24625ffeeMwQsqm70U1eLS, total: 61.9k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+74.0k tokens) — Context tokens surged by 74.0k tokens in turn 119 (session: ses_eaad954a1ffek1dSxjZ0I4Cy6Y, total: 74.0k tokens, threshold: 40000)
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: mcp-gateway_mcp-find (33.3%) — Tool 'mcp-gateway_mcp-find' failed 2 out of 6 calls (33.3% failure rate, threshold: >20%)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 2,144,153 | 40,884 | 11,382,749 | 84.1% | 174.4k | ELEVATED | $2.71 |
| **planner** | 2 | 887,207 | 20,817 | 6,960,050 | 88.7% | 145.3k | ELEVATED | $1.47 |
| **coder** | 2 | 1,819,001 | 49,942 | 24,885,313 | 93.2% | 167.7k | ELEVATED | $3.74 |
| **tester** | 2 | 715,865 | 20,586 | 9,971,184 | 93.3% | 120.8k | ELEVATED | $1.45 |
| **reviewer** | 2 | 466,353 | 10,602 | 4,810,963 | 91.2% | 82.3k | HEALTHY | $0.82 |
| **Total** | **9** | **6,032,579** | **142,831** | **58,010,259** | **90.6%** | **174.4k** | **ELEVATED** | **$10.19** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 9 | 6,032,579 | 142,831 | 58,010,259 | 90.6% | $10.19 |
| **Total** | **9** | **6,032,579** | **142,831** | **58,010,259** | **90.6%** | **$10.19** |
