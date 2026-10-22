# Contract: AL-24 Dynamic Role and Model Aggregations and Presentation

## 1. Problem Statement

AgenticLens currently restricts agent role tracking to a static five-member enumeration (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`). In `backend/internal/domain/aggregations.go`, rollup routines calculate subagent metrics by iterating strictly over `SubagentRoleOrder`. Consequently, custom roles present in invocation traces (such as `advisor`, `po`, `researcher`, `architect`, `security-auditor`) are omitted from:
1. `RoleEffortDistributionPct`: Custom role costs are excluded from `subCosts` and `totSubCost`, resulting in incomplete distribution metrics.
2. `SubagentTurnVelocity`: Invocation turns from non-canonical roles are excluded from `allSubTurns`, `avg_turns_per_invocation`, `peak_turns`, and `by_role`.
3. `ContextWindowHealth`: `subagents_peak_tokens` and `subagents_peak_role` only inspect `SubagentRoleOrder`, hiding peak context spikes if a custom role consumes the largest context window.
4. Telemetry markdown generation in `backend/internal/adapters/formatters.go`: Markdown summaries loop over `domain.SubagentRoleOrder`, omitting custom roles from reported summaries.

On the frontend:
1. The `AgentRole` type in `web/src/domain/types.ts` is a closed union of five literal strings, creating typing mismatches for dynamic roles.
2. Visual components (`DistributionCard.tsx`, `SaturationBar.tsx`, `KPICard.tsx`, `TicketDeepDive.tsx`) lack dynamic color assignment and contrast management for non-canonical roles and arbitrary model identifiers.
3. Model sorting in `ComputeModelMetrics` lacks deterministic tie-breaking when two models report identical costs.

Ticket AL-24 refactors backend aggregations and frontend presentations to discover, aggregate, and render arbitrary agent roles and model names dynamically without data loss or truncation, while preserving canonical role display precedence.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must satisfy `TECHNICAL_GUIDELINES.md` and `AGENTS.md`:

### Inward Dependency Flow
1. Functional Core (`backend/internal/domain/`):
   - Pure functional domain logic in `backend/internal/domain/aggregations.go` and `backend/internal/domain/enums.go`.
   - Zero filesystem, network, database, clock, or external dependencies.
   - Pure mathematical calculations for role effort distributions, subagent turn velocities, context window peaks, and model metric aggregations.
   - Pure deterministic ordering algorithms for canonical and custom roles and models.
2. Boundary Adapters (`backend/internal/adapters/`):
   - `backend/internal/adapters/schemas.go`: Validates that DTO serialization and deserialization accept arbitrary role and model strings without truncation or fallback coercion to default roles.
   - `backend/internal/adapters/formatters.go`: Formats markdown summaries dynamically incorporating all encountered subagent roles.
3. Presentation Layer (`web/src/`):
   - `web/src/domain/types.ts`: Pure TypeScript types and canonical ordering constants.
   - `web/src/utils/theme.ts`: Pure deterministic color hashing utilities for roles and models.
   - `web/src/components/` and `web/src/views/`: Presentation components rendering dynamic roles and models with accessible contrast and zero runtime exceptions.

### Strong Typing and Zero Any Invariant
1. Zero usage of `any` or `interface{}` across Go backend and TypeScript frontend.
2. `AgentRole` in `web/src/domain/types.ts` widened to `type AgentRole = string`.
3. Preserved typed string constants for canonical roles across Go and TypeScript.

### Zero Emojis Mandate
No emojis permitted in code, comments, docstrings, UI labels, tooltips, chart legends, logs, or workspace markdown artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope for AL-24
1. Go Domain Aggregations Core (`backend/internal/domain/`):
   - Refactor `ComputeWorkflowEfficiency` in `aggregations.go` to treat every non-orchestrator role (`role != RoleOrchestrator` and `role != ""`) as a subagent.
   - Remove hardcoded iteration loops over `SubagentRoleOrder` when deriving `subCosts`, `totSubCost`, `RoleEffortDistributionPct`, `SubagentTurnVelocity` (`allSubTurns`, `avg_turns_per_invocation`, `peak_turns`, `by_role`), and `ContextWindowHealth` (`subagents_peak_tokens`, `subagents_peak_role`).
   - Implement pure ordering helper functions `OrderRoles` and `OrderSubagentRoles` that prioritize canonical roles (`CanonicalRoleOrder` / `SubagentRoleOrder`), followed by custom roles sorted in lexicographical order.
   - Ensure `ContextWindowHealth` tie-breaking favors canonical subagent order first, then alphabetical custom roles.
   - Refactor `ComputeModelMetrics` to deterministically tie-break models with identical `CostUSD` by sorting model identifiers alphabetically ascending.
   - Apply identical deterministic model tie-breaking inside `ComputeThreadSummaries`.
2. Go Adapters and Boundary Serialization (`backend/internal/adapters/`):
   - Verify `InvocationFactDTO` and `EventFactDTO` in `schemas.go` cleanly preserve arbitrary non-empty role strings and model strings.
   - Update `FormatMarkdownSummary` in `formatters.go` to iterate over all dynamically discovered subagent roles when formatting `Role Effort Distribution (Subagents)` and `Subagent Turn Velocity & Thrashing`.
3. TypeScript Domain and Presentation (`web/src/`):
   - Widen `AgentRole` in `web/src/domain/types.ts` from fixed union to open `string` type (`type AgentRole = string`).
   - Retain `CANONICAL_ROLE_ORDER` and `SUBAGENT_ROLE_ORDER` as readonly arrays for precedence rendering.
   - Implement deterministic color hashing in `web/src/utils/theme.ts` (`getRoleColor` and `getModelColor`) mapping canonical roles to established theme tokens and arbitrary strings to accessible, distinct palette colors.
   - Update `DistributionCard.tsx` to support item color indicators and dynamic progress track fills.
   - Update `SaturationBar.tsx` to render all custom roles in the `Peak Context by Agent Role` grid and support arbitrary `subagents_peak_role` labels.
   - Update `KPICard.tsx` and `TicketDeepDive.tsx` tooltips to reflect dynamic subagent roles rather than fixed four-subagent text.
   - Update `TicketDeepDive.tsx` to render dynamic turn velocity bars and tables with dynamic role and model colors.
4. Quality & Verification:
   - Pure table-driven Go tests in `backend/internal/domain/aggregations_test.go` verifying custom roles (`advisor`, `po`, `researcher`, `security-auditor`), arbitrary model strings (`claude-3-7-sonnet`, `gemini-2.5-pro`), tie-breaking, and zero-cost/zero-turn edge cases.
   - Go adapter tests in `backend/internal/adapters/formatters_test.go` asserting custom roles appear in markdown output.
   - TypeScript unit tests in `web/src/utils/theme.test.ts`, `web/src/components/Components.test.tsx`, and `web/src/views/TicketDeepDive.test.tsx` verifying stable color hashing, distribution card rendering with custom roles/models, and saturation bar rendering without NaN or crash states.

### Out of Scope for AL-24
1. Natural ticket identifier parsing and sortable tables (governed by AL-23).
2. Contextual hover information icons (governed by AL-25).
3. Fleet Overview completion date filtering and autopsy sorting (governed by AL-26).
4. Live SQLite hook and SSE streaming engine (governed by AL-17 and AL-18).
5. Modification of database extraction SQL queries in `backend/internal/adapters/opencode/extractor.go`.

---

## 4. Architectural Contracts and Specifications

### Subagent Classification Contract
A role is classified as a subagent if and only if:
- Normalized role string is non-empty.
- Normalized role string does not equal `domain.RoleOrchestrator` (`"orchestrator"`).

### Role and Model Ordering Specification
1. Full Role Ordering:
   - Canonical roles present in the dataset appear first in fixed order: `orchestrator`, `planner`, `coder`, `tester`, `reviewer`.
   - All non-canonical roles present in the dataset follow, sorted in alphabetical (lexicographical) order.
2. Subagent Role Ordering:
   - Canonical subagent roles present appear first in fixed order: `planner`, `coder`, `tester`, `reviewer`.
   - All non-canonical subagent roles present follow, sorted in alphabetical order.
3. Model Metric Ordering:
   - Primary sort: `CostUSD` descending.
   - Secondary sort (tie-breaker): `Model` name ascending.

### Peak Subagent Context Tie-Breaking Specification
When evaluating `ContextWindowHealth.SubagentsPeakTokens` and `SubagentsPeakRole`:
- Find the maximum peak context tokens across all subagent roles.
- If multiple subagent roles share the exact same peak token value, selection order follows the Subagent Role Ordering: canonical subagents (`planner`, `coder`, `tester`, `reviewer`) have precedence over custom subagents; custom subagents tie-break alphabetically.
- If no subagents exist, `SubagentsPeakTokens` is `0` and `SubagentsPeakRole` is `"none"`.

### Frontend Color Token Contract
- Canonical role colors:
  - `orchestrator`: `#8b5cf6` (purple)
  - `planner`: `#06b6d4` (cyan)
  - `coder`: `#10b981` (green)
  - `tester`: `#f59e0b` (amber)
  - `reviewer`: `#3b82f6` (blue)
- Custom roles and models:
  - Derived via deterministic polynomial string hash (`djb2` or prime multiplier over character codes).
  - Mapped into a high-contrast accessible color palette suitable for dark theme backgrounds.
  - Idempotent: identical input string always produces identical hex color.

---

## 5. Acceptance Criteria

1. Non-canonical agent roles (e.g. `advisor`, `po`, `researcher`, `security-auditor`) are included in `RoleCosts`, `SubagentTurnVelocity`, `RoleEffortDistributionPct`, and `ContextWindowHealth`.
2. Arbitrary model identifiers (e.g. `claude-3-7-sonnet`, `gemini-2.5-pro`, `custom-fine-tuned-model`) are aggregated and displayed in `Model Cost Distribution` and telemetry tables without omission or truncation.
3. Canonical roles maintain top display ordering across all summaries and tables, followed by custom roles sorted alphabetically.
4. Frontend displays custom roles and models with consistent styling, deterministic colors, and accessible contrast.
5. Zero `any` or `interface{}` types introduced across Go and TypeScript.
6. All automated checks pass cleanly via `make check` and `make test`.
