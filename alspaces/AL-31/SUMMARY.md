# Summary of Changes: AL-31 Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests

## Overview

Ticket AL-31 decouples structural session hierarchy (`ParentID == nil` for root coordination session vs `ParentID != nil` for delegated child subagents) from agent persona role names across the domain engine, boundary adapters, markdown formatters, and parity regression test suites.

Across Iterations 01, 02, 03, and 04:
- The domain engine completely eliminates persona string blacklists and role shims (`IsSubagentRole` deleted entirely).
- Invocations with `ParentID == nil` exclusively govern root coordination metrics (`OrchestratorPeakTokens`, `OrchestratorSaturationPct`).
- Invocations with `ParentID != nil` exclusively govern delegated child subagent metrics (`SubagentsPeakTokens`, `SubagentsPeakRole`, `RoleEffortDistributionPct`, `SubagentTurnVelocity`).
- Rework loop detection in `detectSubagentReworkLoops` indexes session identifiers to `InvocationFact.ParentID != nil`, structurally ignoring root coordination turns during cycle evaluation without relying on role heuristics or fallback shims.
- Role resolution (`RoleMappings.ResolveRole`) is fully de-hardcoded: all hardcoded switch cases were eliminated. Configured mappings are applied, empty/whitespace strings resolve to `RoleUnassigned` (`"unassigned"`), and unmapped arbitrary roles return `AgentRole(cleaned)` directly (e.g. `developer`, `researcher`, `architect`).
- Role ordering (`OrderRoles`, `OrderSubagentRoles`) is configurable: `role_order` is loaded from `.agenticlens/config.yaml` into `ProjectConfig.RoleOrder`. By default, roles sort in ascending alphabetical order whenever a custom order is omitted or unconfigured.
- Fragile ticket title inference (`InferTicketTitle`, `cleanHeadingTitle`, `findWorkspaceFile`) was completely removed from the git adapter, CLI commands, and test suites. Ticket titles originate strictly from explicit user input or cached ledger facts.
- Git branch inference tests in `inference_test.go` and `analyzer_test.go` use isolated transient branch fixtures with `t.Cleanup`, resolving CI test failures on fresh checkouts.

---

## Changes Grouped by Layer

### 1. Domain Layer (`backend/internal/domain/`)

- `backend/internal/domain/config.go`:
  - In `RoleMappings.ResolveRole`, removed hardcoded switch block matching orchestrator, build, root, planner, coder, tester, reviewer.
  - Resolves configured aliases via exact, raw, and case-insensitive lookups.
  - Returns `RoleUnassigned` for empty or whitespace-only inputs.
  - Returns `AgentRole(cleaned)` directly for all unmapped roles.
  - Added `RoleOrder []string` to `ProjectConfig` (with `role_order` JSON/YAML tags) and baseline default in `DefaultProjectConfig()`.
- `backend/internal/domain/enums.go`:
  - Defined `RoleUnassigned AgentRole = "unassigned"`.
  - Completely removed `IsSubagentRole(role string) bool`.
  - Refactored `OrderRoles` and `OrderSubagentRoles` to accept candidate roles and an optional variadic `customOrder ...[]string`.
  - When `customOrder` is omitted or empty, all non-empty deduplicated roles sort in ascending alphabetical order.
  - When `customOrder` is supplied, matching roles appear first in the exact custom sequence (deduplicated), followed by remaining unlisted roles sorted in ascending alphabetical order.
- `backend/internal/domain/aggregations.go`:
  - In `ComputeWorkflowEfficiency`, partitioned invocations strictly by structural hierarchy:
    - Root Invocations (`ParentID == nil`): Governs `rootPeak`, populating `OrchestratorPeakTokens` and calculating `OrchestratorSaturationPct` regardless of role persona string (`general`, `build`, `orchestrator`, or custom).
    - Delegated Child Subagents (`ParentID != nil`): Normalizes whitespace roles (defaulting to `"unassigned"` if empty), accumulating `subCosts`, `totSubCost`, `subPeaks`, and `roleTurns`.
    - Single-session tickets without child invocations evaluate to `SubagentsPeakTokens = 0`, `SubagentsPeakRole = "none"`, empty `RoleEffortDistributionPct`, and zero `SubagentTurnVelocity`.
  - Updated `ComputeWorkflowEfficiency`, `ComputeRoleMetrics`, and `ComputeThreadSummaries` to accept optional `roleOrder ...[]string`.
  - Implemented `DeriveMetricsLedgerWithConfig(factLedger, cfg)` propagating `cfg.RoleMappings` and `cfg.RoleOrder` to metrics aggregations.
  - Maintained `DeriveMetricsLedger(factLedger, roleMappings...)` for backward compatibility.
- `backend/internal/domain/anomalies.go`:
  - In `detectSubagentReworkLoops`, mapped `EventFact.SessionID` to `isChildSession` (`inv.ParentID != nil`). Root coordination turns (`ParentID == nil`) are structurally ignored from the role sequence. Eliminated all fallback calls to `IsSubagentRole`.
- `backend/internal/domain/config_test.go`:
  - Updated test cases asserting `RoleUnassigned` for empty/whitespace inputs.
  - Added `TestRoleMappings_ResolveRole_AdversarialUnmappedAndWhitespace` testing arbitrary unmapped roles, whitespace padding, case variations, and nil/empty mapping resilience.
- `backend/internal/domain/enums_test.go`:
  - Added `TestOrderRoles_And_OrderSubagentRoles_AdversarialComprehensive` verifying alphabetical default sorting, custom priority order, whitespace trimming, duplicates, and boundary edge cases.
- `backend/internal/domain/aggregations_test.go`:
  - Added `TestComputeWorkflowEfficiency_HierarchyDecoupling` and `TestComputeWorkflowEfficiency_AdversarialStructuralHierarchy`.
  - Added `TestComputeWorkflowEfficiency_AdversarialTieBreaking` verifying alphabetical tie-breaking by default, custom order tie-breaking, and root session isolation.
  - Updated `TestOrderRoles`, `TestOrderSubagentRoles`, `TestComputeRoleMetrics_DynamicOrdering`, and `TestComputeThreadSummaries_DynamicOrderingAndTieBreaking`.
- `backend/internal/domain/anomalies_test.go`:
  - Synchronized all unit test fixtures in `TestDetectAnomalies_SubagentReworkLoop` with explicit `Invocations` hierarchy (`ParentID: nil` for root, `ParentID: &rootID` for child subagents) and mapped `SessionID`.
  - Added boundary tests for empty ledgers, unmapped session IDs, orphaned child sessions, and root-only sessions.

### 2. Adapters Layer (`backend/internal/adapters/`)

- `.agenticlens/config.yaml`:
  - Added `role_order: [orchestrator, planner, coder, tester, reviewer]` configuration.
- `backend/internal/adapters/config/loader.go`:
  - Added `RoleOrder []string` to `rawProjectConfigDTO`.
  - In `mergeConfig`, trimmed, lowercased, and assigned `dto.RoleOrder` to `cfg.RoleOrder`, falling back to default when omitted.
- `backend/internal/adapters/config/loader_test.go`:
  - Added `TestLoad_RoleOrderParsingAndFallback` testing YAML and JSON parsing, omission fallback, and whitespace sanitization.
- `backend/internal/adapters/git/inference.go`:
  - Completely deleted `InferTicketTitle`, `cleanHeadingTitle`, and `findWorkspaceFile`.
  - Preserved `ComputeSDLCIterations`, `GetGitUser`, and `InferTicketType`.
- `backend/internal/adapters/repository.go`:
  - Updated `ScanWorkspaces` and `ScanSingleWorkspace` to call `domain.DeriveMetricsLedgerWithConfig(factLedger, cfg)`.
- `backend/internal/adapters/formatters.go`:
  - Updated `FormatMarkdownMetricsTable` to accept optional `roleOrder ...[]string` and pass to `domain.OrderSubagentRoles`.
- `backend/internal/adapters/formatters_test.go`:
  - Added unit test `TestFormatMarkdownMetricsTable_SingleRootAndZeroSubagents`.
  - Updated `TestFormatMarkdownMetricsTable_MixedCanonicalAndCustomSubagentsPrecedence` verifying custom priority and alphabetical default rendering.
- `backend/internal/adapters/parity_test.go`:
  - Updated reference baseline expectations for tickets with authentic `"general"` root sessions (`AL-1`, `AL-10`, `AL-11`, `AL-12`, `GEN-1`, `GEN-2`, `PO-1`, `PO-2`, `PO-3`) in `PeakByRole`.
  - Updated `TestFleetSummaryAndComparisonParity` fleet role costs to authentically partition spend between `orchestrator` ($7.8645) and `general` ($10.9622).
  - Added `TestStructuralClassificationParityOnWorkspaces` verifying 100% compliance across all 37 workspace ledgers.
- `backend/internal/adapters/git/inference_test.go` & `backend/internal/adapters/git/analyzer_test.go`:
  - Implemented `createTestBranch` helper creating transient branches at HEAD with automatic `t.Cleanup` removal.
  - Removed obsolete `InferTicketTitle` test invocations.

### 3. CLI & Presentation Layer

- `backend/cmd/agentic-lens/main.go`:
  - In `generate` command handler, removed `git.InferTicketTitle` fallback. Active ticket titles are resolved strictly via CLI flags or cached ledger facts.
  - Updated metrics derivation to `domain.DeriveMetricsLedgerWithConfig(factLedger, projCfg)`.
  - Passed `projCfg.RoleOrder` to `FormatMarkdownMetricsTable`.
- `web/`:
  - Preserved existing property contracts across TypeScript interfaces.
  - Confirmed `getRoleColor` supports dynamic palette hashing for arbitrary custom roles.
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
- **Software Development Life Cycle Iterations:** 4
- **Active Agent Duration:** 6h 16m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 174.4k tokens (87% of 200k healthy ceiling, 5 compactions) | Subagents peak 168.1k (Coder)
- **Role Effort Distribution (Subagents):** Planner 19.8%, Coder 49.5%, Tester 19.3%, Reviewer 11.4%
- **Subagent Turn Velocity & Thrashing:** 81.8 avg turns/run (peak: 217 turns) | Planner 59.3, Coder 141.8, Tester 73.5, Reviewer 52.5
- **Test-to-Production Code Ratio:** 15.1:1 (+2,687 test lines / +178 prod lines)
- **Code Delivered:** +2,874 / -462 lines across 19 code files (excluding docs)
- **Reasoning Effort:** 398,211 reasoning tokens (59.9% of generated output)
- **Tool Health & Self-Correction:** 0 tool calls (0 self-corrections, 100.0% success rate)

### Anomalies & Health Flags
- **Status:** 21 execution anomalies detected (15 CRITICAL, 6 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.6k tokens) — Context tokens reached 160.6k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+93.5k tokens) — Context tokens surged by 93.5k to 160.1k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+109.9k tokens) — Context tokens surged by 109.9k to 174.4k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (174.4k tokens) — Context tokens reached 174.4k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+92.5k tokens) — Context tokens surged by 92.5k to 92.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (174.4k tokens) — Context tokens reached 174.4k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (167.7k tokens) — Context tokens reached 167.7k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (166.0k tokens) — Context tokens reached 166.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (168.1k tokens) — Context tokens reached 168.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (174.4k tokens) — Context tokens reached 174.4k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (167.7k tokens) — Context tokens reached 167.7k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (166.0k tokens) — Context tokens reached 166.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (168.1k tokens) — Context tokens reached 168.1k, breaching the 160k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+49.6k tokens) — Context tokens surged by 49.6k to 49.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+63.1k tokens) — Context tokens surged by 63.1k to 161.0k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+61.9k tokens) — Context tokens surged by 61.9k to 61.9k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+74.0k tokens) — Context tokens surged by 74.0k to 74.0k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+74.8k tokens) — Context tokens surged by 74.8k to 74.8k tokens.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: read (3 consecutive errors) — Tool 'read' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 4,717,443 | 69,530 | 21,655,886 | 82.1% | 174.4k | ELEVATED | $5.62 |
| **planner** | 4 | 1,750,272 | 33,222 | 14,063,095 | 88.9% | 145.3k | ELEVATED | $2.93 |
| **coder** | 4 | 3,943,159 | 96,604 | 46,673,871 | 92.2% | 168.1k | ELEVATED | $7.34 |
| **tester** | 4 | 1,361,690 | 47,777 | 19,367,528 | 93.4% | 132.2k | ELEVATED | $2.85 |
| **reviewer** | 4 | 922,537 | 19,896 | 10,517,625 | 91.9% | 94.7k | HEALTHY | $1.69 |
| **Total** | **17** | **12,695,101** | **267,029** | **112,278,005** | **89.8%** | **174.4k** | **ELEVATED** | **$20.44** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 17 | 12,695,101 | 267,029 | 112,278,005 | 89.8% | $20.44 |
| **Total** | **17** | **12,695,101** | **267,029** | **112,278,005** | **89.8%** | **$20.44** |
