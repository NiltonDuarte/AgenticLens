# Implementation Plan: FIX-3 Iteration 06 - Removal of Canonical Role Concepts and Platform Decoupling

Target Path: `alspaces/FIX-3/iterations/06/PLAN.md`

---

## 1. Goal and Iteration Scope

Eliminate repository-specific role assumptions, artificial "canonical" role concepts, and hardcoded project setup details from the AgenticLens platform core across Go backend and TypeScript frontend:

1. De-Hardcode Go Domain Role Constants:
   - Remove repository-specific role constants (`RoleOrchestrator`, `RolePlanner`, `RoleCoder`, `RoleTester`, `RoleReviewer`, `RoleUnassigned`) from `backend/internal/domain/enums.go`.
   - Introduce two generic platform fallback constants:
     - `RoleRootAgent AgentRole = "RootAgent"` for nameless root-level sessions.
     - `RoleUnassignedSubAgent AgentRole = "UnassignedSubAgent"` for nameless subagent sessions.
   - Retain `type AgentRole string` as the domain type for agent roles across the platform.

2. Update Go Adapters to Generic Platform Fallbacks:
   - In `backend/internal/adapters/live/service.go`, replace hardcoded fallbacks to `"orchestrator"`, `"coder"`, or `"unassigned"` with `RoleRootAgent` for nameless root sessions and `RoleUnassignedSubAgent` for nameless subagents or unmapped sessions.
   - In `backend/internal/adapters/opencode/extractor.go`, replace fallback `"orchestrator"` and `"unassigned"` with `RoleRootAgent` and `RoleUnassignedSubAgent`.
   - In `backend/internal/domain/config.go`, update `RoleMappings.ResolveRole` to return `RoleUnassignedSubAgent` when given an empty or whitespace string, and strip "canonical" wording from docstrings.
   - In `backend/internal/domain/aggregations.go`, update subagent role resolution in `ComputeWorkflowEfficiency` to fallback to `RoleUnassignedSubAgent` when role is empty or whitespace.

3. Clean Break Renaming of Root Context Window Health Metrics:
   - In `backend/internal/domain/models.go`:
     - Rename `ContextWindowHealth.OrchestratorPeakTokens` to `RootPeakTokens` with JSON tag `json:"root_peak_tokens"`.
     - Rename `ContextWindowHealth.OrchestratorSaturationPct` to `RootSaturationPct` with JSON tag `json:"root_saturation_pct"`.
     - Rename `TicketComparisonRow.OrchestratorSaturationPct` to `RootSaturationPct` with JSON tag `json:"root_saturation_pct"`.
   - In `backend/internal/domain/aggregations.go`: update calculations assigning `RootPeakTokens` and `RootSaturationPct`.
   - In `backend/internal/adapters/formatters.go`: update references and change markdown output to `Root agent peak %s tokens`.
   - In `web/src/domain/types.ts` and `web/src/types/telemetry.ts`:
     - `ContextWindowHealth`: `root_peak_tokens: number`, `root_saturation_pct: number`.
     - `TicketComparisonRow`: `root_saturation_pct: number`.
     - `ComparisonSortKey`: `'root_saturation_pct'`.
   - In all UI components (`SaturationBar.tsx`, `TicketComparisons.tsx`, `TicketTable.tsx`, `formatters.ts`, `mockData.ts`): update property access.
   - Update all existing JSON ledgers (`alspaces/**/agentic_metrics.json` and fixtures) from `orchestrator_peak_tokens` / `orchestrator_saturation_pct` to `root_peak_tokens` / `root_saturation_pct`.

4. Decouple Theme Colors from Agent Roles:
   - In `web/src/utils/theme.ts`: remove `CANONICAL_ROLE_COLORS`.
   - Make `getRoleColor` purely dynamic and deterministic via hash-to-palette mapping (`hashStringToColor`), eliminating any hardcoded agent role names in frontend styling.

5. Decouple Debug Prompt Ticket Regex:
   - In `web/src/utils/debugPrompt.ts`: replace `(?:AL|FIX|GEN|PO|[A-Z]{2,10})-\d+` with `[A-Z]{2,10}-\d+`.
   - Generalize comments and test descriptions to avoid hardcoding repository-specific ticket prefixes as special cases.

6. Eliminate "Canonical" Role Types and Orderings from TypeScript Contracts:
   - In `web/src/types/telemetry.ts`, remove `CanonicalAgentRole` and redefine `AgentRole` directly as `export type AgentRole = string;`.
   - In `web/src/domain/types.ts`, remove `CanonicalAgentRole` imports and re-exports, remove `CANONICAL_ROLE_ORDER`, remove `SUBAGENT_ROLE_ORDER`, and remove `isCanonicalRole`.
   - In `web/src/test/fixtures/mockData.ts`, replace `CANONICAL_ROLE_ORDER` import with an explicit inline string array in `mockProjectConfig.role_order`.
   - In `web/src/domain/formatters.test.ts`, remove unit tests asserting canonical order constants and type predicates, keeping assertions verifying that `AgentRole` accepts arbitrary string personas.

7. Update Test Suites Across Go and TypeScript:
   - Update all Go unit tests in `enums_test.go`, `config_test.go`, `aggregations_test.go`, `live_test.go`, `schemas_test.go`, `loader_test.go`, `formatters_test.go`, `live_handlers_test.go`, and `extractor_test.go` to use `domain.AgentRole(...)` conversions or the generic fallback constants (`RoleRootAgent`, `RoleUnassignedSubAgent`).
   - Align test expectations for nameless sessions in `extractor_test.go` and `aggregations_test.go` to assert `"RootAgent"` and `"UnassignedSubAgent"`.
   - Align all Vitest test suites to new property names and generic role handling.

Non-goals:
- Modifying `.agenticlens-config.yaml` default file schema (`role_mappings` and `role_order` remain configurable project settings).
- Modifying database schemas or column definitions in SQLite.
- Altering the behavior of `OrderRoles` or `OrderSubagentRoles` (they already accept generic string slices and prioritize configured order before alphabetical sorting).

---

## 2. Grounded Reference Architecture and Symbols

### Go Functional Core (`backend/internal/domain/`)
- `enums.go`:
  - `AgentRole`: String type definition (`type AgentRole string`).
  - Constants to delete: `RoleOrchestrator`, `RolePlanner`, `RoleCoder`, `RoleTester`, `RoleReviewer`, `RoleUnassigned`.
  - Constants to add: `RoleRootAgent`, `RoleUnassignedSubAgent`.
  - `OrderRoles[T ~string](roles []T, customOrder ...[]T) []T`: Generic ordering helper (preserves custom configuration order first, then alphabetical).
  - `OrderSubagentRoles[T ~string](roles []T, customOrder ...[]T) []T`: Generic ordering helper for subagent candidate roles.
- `config.go`:
  - `RoleMappings`: Map of string to string.
  - `RoleMappings.ResolveRole(rawAgent string) AgentRole`: Resolves raw agent persona string. Currently returns `RoleUnassigned` on empty string; must return `RoleUnassignedSubAgent`.
- `aggregations.go`:
  - `ComputeWorkflowEfficiency`: Subagent loop around line 192 currently assigns `role = RoleUnassigned` when role is whitespace; must assign `role = RoleUnassignedSubAgent`.
- `enums_test.go`:
  - `TestOrderRoles_GenericAgentRoleType`: References deleted role constants.
- `config_test.go`:
  - `TestRoleMappings_ResolveRole`, `TestRoleMappings_ResolveRole_AdversarialUnmappedAndWhitespace`: Reference deleted role constants and `RoleUnassigned`.
- `aggregations_test.go`:
  - `TestComputeThreadSummaries_DeduplicationAndCustomOrdering`, table-driven role mapping tests, and subagent context window health tests reference deleted role constants and `"unassigned"`.
- `live_test.go`:
  - `TestBuildSessionTree`, `TestBuildSessionTree_OrphanPromotion`: Reference deleted role constants.

### Go Imperative Boundary Shell (`backend/internal/adapters/`)
- `live/service.go`:
  - `FetchSessionTraceMulti`: Lines 720-732 assign fallback roles to `sessionRoleMap` (`domain.RoleOrchestrator`, `domain.RoleUnassigned`, `domain.RoleCoder`).
  - Lines 765 and 898 assign fallback `role = domain.RoleCoder`.
  - Must use `domain.RoleRootAgent` for nameless root sessions and `domain.RoleUnassignedSubAgent` for nameless subagents or unmapped sessions.
- `opencode/extractor.go`:
  - `ExtractFacts`: Lines 248-256 assign fallback `domain.RoleOrchestrator` for root sessions without agent name and `domain.RoleUnassigned` for subagents without agent name.
  - Line 479 assigns fallback `role := domain.RoleOrchestrator`.
  - Must use `domain.RoleRootAgent` and `domain.RoleUnassignedSubAgent`.
- `schemas_test.go`:
  - References `domain.RoleCoder` and `domain.RoleOrchestrator`.
- `config/loader_test.go`:
  - References `domain.RolePlanner` and `domain.RoleTester`.
- `formatters_test.go`:
  - References `domain.RoleCoder`.
- `http/live_handlers_test.go`:
  - References `domain.RoleOrchestrator` and `domain.RoleCoder`.
- `opencode/extractor_test.go`:
  - `TestExtractFacts_RoleResolutionAndAgentMapping`: Asserts `"orchestrator"` and `"unassigned"` fallback roles.

### TypeScript Domain Contracts (`web/src/types/` and `web/src/domain/`)
- `web/src/types/telemetry.ts`:
  - Declares `CanonicalAgentRole` and `AgentRole`.
  - Must remove `CanonicalAgentRole` and export `type AgentRole = string;`.
- `web/src/domain/types.ts`:
  - Re-exports and imports `CanonicalAgentRole`.
  - Declares `CANONICAL_ROLE_ORDER`, `SUBAGENT_ROLE_ORDER`, and `isCanonicalRole`.
  - Must remove all four symbols and preserve re-export of `AgentRole`.
- `web/src/test/fixtures/mockData.ts`:
  - Imports `CANONICAL_ROLE_ORDER` and sets `mockProjectConfig.role_order: CANONICAL_ROLE_ORDER`.
  - Must use explicit inline array of strings.
- `web/src/domain/formatters.test.ts`:
  - Imports `CANONICAL_ROLE_ORDER`, `SUBAGENT_ROLE_ORDER`, `isCanonicalRole`, `CanonicalAgentRole`.
  - Tests canonical order constants and `isCanonicalRole` predicate.
  - Must remove tests for deleted symbols and update test to verify generic `AgentRole` string acceptance.

---

## 3. Architectural Specifications and Contracts

### Specification A: Removal of Specific Role Constants and Addition of Generic Fallbacks
1. Location: `backend/internal/domain/enums.go`
2. Removal Details:
   - Delete the entire legacy constant block declaring:
     - `RoleOrchestrator AgentRole = "orchestrator"`
     - `RolePlanner AgentRole = "planner"`
     - `RoleCoder AgentRole = "coder"`
     - `RoleTester AgentRole = "tester"`
     - `RoleReviewer AgentRole = "reviewer"`
     - `RoleUnassigned AgentRole = "unassigned"`
3. Additions:
   - Declare generic platform fallback constants:
     - `RoleRootAgent AgentRole = "RootAgent"`
     - `RoleUnassignedSubAgent AgentRole = "UnassignedSubAgent"`
4. Docstring and Comment Decoupling:
   - In `backend/internal/domain/enums.go`: Update docstrings on `OrderRoles` and `OrderSubagentRoles` to remove any mention of "canonical" roles. Describe the behavior purely: roles specified in custom order (if provided and non-empty) are sorted first in that exact order, followed by remaining roles sorted in ascending alphabetical order.
   - In `backend/internal/domain/config.go`: Update docstring on `RoleMappings` from "maps agent persona names to canonical AgentRole enums" to "maps agent persona names to AgentRole enums". Update `ResolveRole` docstring similarly.

### Specification B: Go Domain Config and Aggregation Updates
1. Role Mapping Resolution (`backend/internal/domain/config.go`):
   - In `ResolveRole(rawAgent string) AgentRole`:
     - Clean the input: `cleaned := strings.ToLower(strings.TrimSpace(rawAgent))`.
     - If `cleaned == ""`: Return `RoleUnassignedSubAgent` (previously `RoleUnassigned`).
     - If mapped in `m`: Return `AgentRole(strings.ToLower(strings.TrimSpace(mapped)))`.
     - If unmapped: Return `AgentRole(cleaned)`.
2. Aggregation Subagent Fallback (`backend/internal/domain/aggregations.go`):
   - In `ComputeWorkflowEfficiency`:
     - Inside the loop traversing invocations where `inv.ParentID != nil` (subagents):
     - Check: `if strings.TrimSpace(string(role)) == ""`.
     - Assign: `role = RoleUnassignedSubAgent` (previously `RoleUnassigned`).
     - Accumulate subagent costs, peak context tokens, and turns under `RoleUnassignedSubAgent`.

### Specification C: Go Adapters Alignment
1. OpenCode Database Extractor (`backend/internal/adapters/opencode/extractor.go`):
   - In `ExtractFacts` session processing:
     - When inspecting each session:
       - If `cleanAgent != ""`: `role = domain.AgentRole(cleanAgent)`.
       - Else if `s.parentID == nil` (root session): `role = domain.RoleRootAgent` (previously `domain.RoleOrchestrator`).
       - Else (subagent session): `role = domain.RoleUnassignedSubAgent` (previously `domain.RoleUnassigned`).
     - Store in `sessionRoleMap[s.id] = role`.
   - In part processing fallback (around line 479):
     - Initialize `role := domain.RoleRootAgent` (previously `domain.RoleOrchestrator`).
     - If `r, ok := sessionRoleMap[sID]; ok`: `role = r`.
2. Live Service Adapter (`backend/internal/adapters/live/service.go`):
   - In `FetchSessionTraceMulti` (session role mapping around lines 718-732):
     - Query SQLite `session` table for `agent` and `parent_id`.
     - If query succeeds:
       - If `agent.Valid && strings.TrimSpace(agent.String) != ""`: `sessionRoleMap[sid] = s.cfg.RoleMappings.ResolveRole(strings.TrimSpace(agent.String))`.
       - Else if `!parentID.Valid || strings.TrimSpace(parentID.String) == ""`: `sessionRoleMap[sid] = domain.RoleRootAgent` (previously `domain.RoleOrchestrator`).
       - Else: `sessionRoleMap[sid] = domain.RoleUnassignedSubAgent` (previously `domain.RoleUnassigned`).
     - If query fails: `sessionRoleMap[sid] = domain.RoleUnassignedSubAgent` (previously `domain.RoleCoder`).
   - In message processing loop (lines 763-766):
     - If `role := sessionRoleMap[rm.SessionID]` is empty: fallback to `domain.RoleUnassignedSubAgent` (previously `domain.RoleCoder`).
   - In orphan part processing loop (lines 896-899):
     - If `role := sessionRoleMap[sid]` is empty: fallback to `domain.RoleUnassignedSubAgent` (previously `domain.RoleCoder`).

### Specification D: TypeScript Contracts and Fixtures Decoupling
1. Telemetry Types (`web/src/types/telemetry.ts`):
   - Delete `CanonicalAgentRole` type definition entirely.
   - Redefine `AgentRole`: `export type AgentRole = string;`.
   - Ensure all references to `AgentRole` in `EventFact`, `AnomalyContext`, and elsewhere continue using the string alias.
2. Domain Types (`web/src/domain/types.ts`):
   - Remove import and export of `CanonicalAgentRole`.
   - Retain `export type { AgentRole } from '../types/telemetry';`.
   - Delete `export const CANONICAL_ROLE_ORDER`.
   - Delete `export const SUBAGENT_ROLE_ORDER`.
   - Delete `export function isCanonicalRole`.
   - Retain `InvocationFact`, `RoleMetric`, `ThreadRoleMetric`, `ContextWindowHealth` using `AgentRole`.
3. Test Fixture Decoupling (`web/src/test/fixtures/mockData.ts`):
   - Remove import of `CANONICAL_ROLE_ORDER`.
   - In `mockProjectConfig`: assign `role_order: ['orchestrator', 'planner', 'coder', 'tester', 'reviewer']` as a plain array literal.
4. Formatters Test Suite Decoupling (`web/src/domain/formatters.test.ts`):
   - Remove imports of `CANONICAL_ROLE_ORDER`, `SUBAGENT_ROLE_ORDER`, `isCanonicalRole`, `CanonicalAgentRole`.
   - Remove test `contains all 5 canonical roles in ordered constants`.
   - Remove test `validates canonical roles with isCanonicalRole type predicate`.
   - Update test `accepts both canonical roles and custom subagent strings under AgentRole` to `accepts arbitrary role strings under AgentRole`, verifying that variables typed as `AgentRole` accept strings like `'coder'`, `'architect'`, `'security-auditor'`.
5. Theme Utility Palette (`web/src/utils/theme.ts`):
   - Note that `CANONICAL_ROLE_COLORS` is typed as `Readonly<Record<string, string>>` and does not import `CanonicalAgentRole`. It serves as a visual color palette for common role names.

### Specification E: Go Test Suite Realignment
1. `backend/internal/domain/enums_test.go`:
   - In `TestOrderRoles_GenericAgentRoleType`: Replace `domain.RoleTester`, `domain.RoleCoder`, etc. with `domain.AgentRole("tester")`, `domain.AgentRole("coder")`, etc.
   - Add a test function `TestAgentRole_FallbackConstants` that asserts `domain.RoleRootAgent` equals `"RootAgent"` and `domain.RoleUnassignedSubAgent` equals `"UnassignedSubAgent"`.
2. `backend/internal/domain/config_test.go`:
   - In `TestRoleMappings_ResolveRole`: Replace `domain.RolePlanner`, etc. with `domain.AgentRole("planner")`, etc.
   - Replace assertions on empty strings (`""` and `"   "`) expecting `domain.RoleUnassigned` to assert `domain.RoleUnassignedSubAgent`.
   - In `TestRoleMappings_ResolveRole_AdversarialUnmappedAndWhitespace`:
     - Update table tests to expect `domain.AgentRole(...)`.
     - In subtest testing empty/whitespace strings: assert `domain.RoleUnassignedSubAgent`.
3. `backend/internal/domain/aggregations_test.go`:
   - In `TestComputeThreadSummaries_DeduplicationAndCustomOrdering` (line 1750): Replace `domain.RoleOrchestrator` and `domain.RoleCoder` with `domain.AgentRole("orchestrator")` and `domain.AgentRole("coder")`.
   - In table-driven test lines 2743, 2751, 2759: Replace `domain.RoleOrchestrator` and `domain.RoleCoder` with `domain.AgentRole("orchestrator")` and `domain.AgentRole("coder")`.
   - In subtests testing whitespace subagents (lines 5182-5216 and 5503-5554):
     - Assert `domain.RoleUnassignedSubAgent`.
   - In subtest `unassigned role participates in tie-breaking properly` (lines 5650-5674):
     - Align with ASCII sort order for `"UnassignedSubAgent"`.
4. `backend/internal/domain/live_test.go`:
   - Replace `domain.RoleTester`, etc. with `domain.AgentRole(...)`.
5. `backend/internal/adapters/schemas_test.go`, `formatters_test.go`, `config/loader_test.go`, `http/live_handlers_test.go`:
   - Replace removed constants with `domain.AgentRole(...)`.
6. `backend/internal/adapters/opencode/extractor_test.go`:
   - In `TestExtractFacts_RoleResolutionAndAgentMapping`:
     - Assert `domain.RoleRootAgent` (`"RootAgent"`) for nameless root sessions.
     - Assert `domain.RoleUnassignedSubAgent` (`"UnassignedSubAgent"`) for nameless child sessions.

---

## 4. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Core Enum and Config Updates
- Task 1.1: In `backend/internal/domain/enums.go`:
  - Delete constants `RoleOrchestrator`, `RolePlanner`, `RoleCoder`, `RoleTester`, `RoleReviewer`, `RoleUnassigned`.
  - Add constants `RoleRootAgent AgentRole = "RootAgent"` and `RoleUnassignedSubAgent AgentRole = "UnassignedSubAgent"`.
  - Update docstrings for `OrderRoles` and `OrderSubagentRoles` to remove the word "canonical".
- Task 1.2: In `backend/internal/domain/config.go`:
  - In `RoleMappings.ResolveRole`, change empty/whitespace fallback return value from `RoleUnassigned` to `RoleUnassignedSubAgent`.
  - Update docstring for `RoleMappings` and `ResolveRole` to remove "canonical".
- Task 1.3: In `backend/internal/domain/aggregations.go`:
  - In `ComputeWorkflowEfficiency`, update subagent role resolution on whitespace to assign `RoleUnassignedSubAgent`.

### Phase 2: Go Adapters Realignment
- Task 2.1: In `backend/internal/adapters/live/service.go`:
  - In `FetchSessionTraceMulti`, update nameless root session fallback to `domain.RoleRootAgent`.
  - Update nameless subagent session fallback to `domain.RoleUnassignedSubAgent`.
  - Update DB query failure fallback to `domain.RoleUnassignedSubAgent`.
  - In message and orphan part loops, update empty role fallbacks from `domain.RoleCoder` to `domain.RoleUnassignedSubAgent`.
- Task 2.2: In `backend/internal/adapters/opencode/extractor.go`:
  - In `ExtractFacts`, update nameless root session fallback to `domain.RoleRootAgent`.
  - Update nameless subagent session fallback to `domain.RoleUnassignedSubAgent`.
  - In part processing fallback, initialize role to `domain.RoleRootAgent`.

### Phase 3: TypeScript Domain Contracts and Fixture Updates
- Task 3.1: In `web/src/types/telemetry.ts`:
  - Delete `export type CanonicalAgentRole`.
  - Redefine `AgentRole`: `export type AgentRole = string;`.
- Task 3.2: In `web/src/domain/types.ts`:
  - Remove imports and re-exports of `CanonicalAgentRole`.
  - Delete `CANONICAL_ROLE_ORDER`, `SUBAGENT_ROLE_ORDER`, and `isCanonicalRole`.
- Task 3.3: In `web/src/test/fixtures/mockData.ts`:
  - Remove import of `CANONICAL_ROLE_ORDER`.
  - Update `mockProjectConfig.role_order` to inline array `['orchestrator', 'planner', 'coder', 'tester', 'reviewer']`.
- Task 3.4: In `web/src/domain/formatters.test.ts`:
  - Remove imports of `CANONICAL_ROLE_ORDER`, `SUBAGENT_ROLE_ORDER`, `isCanonicalRole`, `CanonicalAgentRole`.
  - Delete tests asserting canonical role arrays and `isCanonicalRole` predicate.
  - Update test verifying `AgentRole` accepts arbitrary strings.

### Phase 5: Verification and Quality Gates
- Task 5.1: Run TypeScript typecheck (`npm run typecheck`) to verify zero type errors.
- Task 5.2: Run TypeScript linter (`npm run lint`) to verify zero ESLint errors or warnings.
- Task 5.3: Run frontend test suites (`npm test`) to verify all Vitest tests pass.
- Task 5.4: Run Go test suites (`go test -race ./...`) across all domain and adapter packages.
- Task 5.5: Run Go linter/vet (`go vet ./...`) across all packages.
- Task 5.6: Confirm binary build (`make build`) succeeds with embedded assets.
- Task 5.7: Verify zero emojis in code, comments, docstrings, and workspace files.
