# Implementation Plan: AL-24 Dynamic Role and Model Aggregations and Presentation

## 1. Goal and Iteration Scope

The goal of iteration 01 is to deliver ticket AL-24: refactoring backend aggregations and frontend visualizations to support arbitrary, dynamic agent roles and model names while preserving canonical role display precedence and ensuring zero data loss or truncation.

This iteration encompasses removing hardcoded loop constraints in `backend/internal/domain/aggregations.go`, treating all non-orchestrator roles as subagents for velocity, effort distribution, and context health derivations, implementing deterministic tie-breaking for model metrics, updating markdown generation in `backend/internal/adapters/formatters.go`, widening `AgentRole` in `web/src/domain/types.ts`, implementing deterministic dynamic color hashing in `web/src/utils/theme.ts`, updating frontend components (`DistributionCard.tsx`, `SaturationBar.tsx`, `KPICard.tsx`, `TicketDeepDive.tsx`), and implementing adversarial Go and TypeScript test suites.

## 2. Grounded Reference Architecture & File Paths

The implementation maps directly to the following repository locations:

- Functional Core Domain Layer:
  - `backend/internal/domain/enums.go`: Contains `AgentRole`, `CanonicalRoleOrder`, and `SubagentRoleOrder`. Add pure role ordering helper functions.
  - `backend/internal/domain/aggregations.go`: Contains `ComputeWorkflowEfficiency`, `ComputeRoleMetrics`, `ComputeModelMetrics`, `ComputeThreadSummaries`, and `CalculateFleetSummary`. Refactor subagent iteration, peak context detection, and model sorting.
  - `backend/internal/domain/aggregations_test.go`: Add pure table-driven unit tests for dynamic roles, custom models, tie-breaking, and zero-division edge cases.
- Boundary Adapters Layer:
  - `backend/internal/adapters/schemas.go`: Verify `InvocationFactDTO` and `EventFactDTO` to ensure arbitrary role and model strings are preserved without mutation.
  - `backend/internal/adapters/formatters.go`: Update `FormatMarkdownSummary` to format role effort and turn velocity dynamically across all discovered subagent roles.
  - `backend/internal/adapters/formatters_test.go`: Add test verifying custom roles in rendered markdown summaries.
- Presentation Domain and Utilities:
  - `web/src/domain/types.ts`: Widen `AgentRole` type to string; maintain canonical order arrays.
  - `web/src/utils/theme.ts`: Implement `getRoleColor` and `getModelColor` with deterministic color hashing.
  - `web/src/utils/theme.test.ts`: Unit tests asserting deterministic color resolution, canonical token preservation, and valid hex formatting.
- Presentation Components and Views:
  - `web/src/components/DistributionCard.tsx`: Add color indicators and dynamic progress track fill styling.
  - `web/src/components/SaturationBar.tsx`: Render custom roles in the peak context breakdown and handle arbitrary subagent peak roles.
  - `web/src/components/KPICard.tsx`: Ensure accessible tooltip and text rendering for dynamic values.
  - `web/src/views/TicketDeepDive.tsx`: Update turn velocity and breakdown tables with dynamic colors and generalized tooltips.
  - `web/src/components/Components.test.tsx`: Test dynamic role/model rendering in `DistributionCard` and `SaturationBar`.
  - `web/src/views/TicketDeepDive.test.tsx`: Verify autopsy view renders custom roles and models.

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Pure Domain Aggregations Refactoring (backend/internal/domain/)

- Task 1.1: In `backend/internal/domain/enums.go`, author pure helper functions for role classification and ordering:
  - Implement `IsSubagentRole(role string) bool`: Trims whitespace and normalizes case; returns true if non-empty and not equal to `RoleOrchestrator` (`"orchestrator"`).
  - Implement `OrderRoles(roles []string) []string`: Takes arbitrary slice of role strings, deduplicates, and returns canonical roles in `CanonicalRoleOrder` followed by remaining custom roles sorted alphabetically.
  - Implement `OrderSubagentRoles(roles []string) []string`: Takes arbitrary slice of subagent role strings, deduplicates, and returns canonical subagents in `SubagentRoleOrder` followed by remaining custom subagent roles sorted alphabetically.
- Task 1.2: In `backend/internal/domain/aggregations.go`, refactor `ComputeWorkflowEfficiency`:
  - Identify all unique roles present in `ledger.Invocations`.
  - Compute subagent costs for every role satisfying `IsSubagentRole`. Calculate total subagent cost across all discovered subagent roles.
  - Calculate `RoleEffortDistributionPct` for every discovered subagent role with positive cost.
  - Aggregate subagent turns (`allSubTurns`) across all invocations where `inv.ParentID != nil` and `IsSubagentRole(inv.Role)` is true. Derive `avg_turns_per_invocation` and `peak_turns` over all subagent turns.
  - Populate `SubagentTurnVelocity.ByRole` for every subagent role having recorded turns.
  - Determine `SubagentsPeakTokens` and `SubagentsPeakRole` by inspecting all roles satisfying `IsSubagentRole`. For tie-breaking when two or more subagents have the same peak context, iterate according to `OrderSubagentRoles` so canonical subagents take precedence over custom subagents, and custom subagents tie-break alphabetically. If no subagents exist, set peak tokens to zero and peak role to `"none"`.
- Task 1.3: In `backend/internal/domain/aggregations.go`, refactor `ComputeModelMetrics`:
  - Ensure all unique models in `invocations` are aggregated without filtering.
  - Sort the resulting `ModelMetric` slice primarily by `CostUSD` descending. When `CostUSD` values are identical, sort secondarily by `Model` string ascending for deterministic results.
- Task 1.4: In `backend/internal/domain/aggregations.go`, refactor `ComputeThreadSummaries`:
  - Within each thread session group, order roles using `OrderRoles`.
  - Within each role's `models` slice, sort `ModelMetric` entries by `CostUSD` descending with secondary tie-breaker on `Model` string ascending.

### Phase 2: Markdown Telemetry Formatter Refactoring (backend/internal/adapters/)

- Task 2.1: In `backend/internal/adapters/formatters.go`, update `FormatMarkdownSummary`:
  - In the workflow efficiency section, collect all keys from `we.RoleEffortDistributionPct`. Order the roles using `domain.OrderSubagentRoles`. Format each subagent percentage as `TitleCase(role) X.X%`.
  - In the turn velocity section, collect all keys from `we.SubagentTurnVelocity.ByRole`. Order the roles using `domain.OrderSubagentRoles`. Format each turn average as `TitleCase(role) X.X`.
- Task 2.2: In `backend/internal/adapters/schemas.go`, verify boundary unmarshaling:
  - Confirm `InvocationFactDTO.ToDomain()` preserves arbitrary role and model values without falling back to defaults unless empty.
  - Confirm `EventFactDTO.ToDomain()` preserves arbitrary role values.

### Phase 3: Frontend Type Definitions and Deterministic Color Hashing (web/src/)

- Task 3.1: In `web/src/domain/types.ts`:
  - Widen `AgentRole` from literal union to open string: `export type AgentRole = string;`
  - Retain `CANONICAL_ROLE_ORDER` and `SUBAGENT_ROLE_ORDER` as readonly arrays.
  - Export helper `isCanonicalRole(role: string): boolean`.
- Task 3.2: In `web/src/utils/theme.ts`:
  - Define an accessible color palette array of vibrant hex colors with verified contrast against dark card backgrounds (e.g. pink, teal, rose, lime, violet, sky, yellow-amber, indigo, fuchsia, emerald, orange, cyan).
  - Implement pure hashing function `hashStringToColor(str: string): string` using deterministic polynomial character summation mapped modulo palette length.
  - Implement `getRoleColor(role: string): string`: Returns canonical color for canonical roles (`orchestrator` purple, `planner` cyan, `coder` green, `tester` amber, `reviewer` blue); returns `hashStringToColor` for custom roles.
  - Implement `getModelColor(model: string): string`: Returns `hashStringToColor(model)`.

### Phase 4: UI Component Updates for Dynamic Roles and Models (web/src/)

- Task 4.1: In `web/src/components/DistributionCard.tsx`:
  - Add support for color-coded progress bars and color badges.
  - Determine color for each distribution item using `getRoleColor` or `getModelColor` based on the card title or an optional `colorType` prop (`role`, `model`, `default`).
  - Render progress fill using item-specific color with defensive percentage clamping between 0 and 100, guarding against NaN when total cost is zero.
- Task 4.2: In `web/src/components/SaturationBar.tsx`:
  - Update `Peak Context by Agent Role` grid to render all role entries in `contextHealth.peak_by_role`.
  - Apply `getRoleColor(role)` to role indicators and mini-track progress bars (switching to theme red when exceeding ceiling).
  - Ensure `contextHealth.subagents_peak_role` renders cleanly for any custom role string without truncation.
- Task 4.3: In `web/src/components/KPICard.tsx`:
  - Ensure value, secondary, and tooltip props handle dynamic content without layout distortion.
- Task 4.4: In `web/src/views/TicketDeepDive.tsx`:
  - Update `SDLC Iterations` KPI card tooltip to read: `Total SDLC feedback iterations across all participating agent roles`.
  - In `Subagent Turn Velocity` card, render all roles present in `subTurns.by_role`, styling turn track fills with `getRoleColor(role)`.
  - In `Role Resource Consumption` table, render all roles from `ledger.roles` with role badges or color indicators.
  - In `Model Resource Consumption` table, render all models from `ledger.models` with model color indicators.

### Phase 5: Adversarial Unit and Boundary Testing

- Task 5.1: In `backend/internal/domain/aggregations_test.go`:
  - Add test verifying `ComputeWorkflowEfficiency` with non-canonical roles (`advisor`, `po`, `researcher`, `architect`): assert turn velocity, effort distribution percentages, and peak context tracking include all custom subagents.
  - Add test verifying peak context tie-breaking: when a canonical role and a custom role have equal peak context, canonical role is selected; when two custom roles tie, alphabetical order is selected.
  - Add test verifying `ComputeRoleMetrics` with canonical and custom roles: assert canonical roles appear first, followed by custom roles sorted alphabetically.
  - Add test verifying `ComputeModelMetrics` with non-standard model strings: assert sort order is cost descending with secondary alphabetical tie-breaking for equal costs.
  - Add test verifying edge cases: ledger with zero subagents, ledger where all invocations are orchestrator, ledger with zero cost.
- Task 5.2: In `backend/internal/adapters/formatters_test.go`:
  - Add test verifying `FormatMarkdownSummary` renders custom subagent names in both role effort distribution and turn velocity lines.
- Task 5.3: In `web/src/utils/theme.test.ts`:
  - Add unit tests for `getRoleColor`: assert canonical roles return fixed theme colors; assert custom roles return valid hex strings; assert identical custom strings return identical colors across multiple calls.
  - Add unit tests for `getModelColor`: assert deterministic hex color generation.
- Task 5.4: In `web/src/components/Components.test.tsx`:
  - Add test for `DistributionCard` rendering custom roles and arbitrary model names with correct percentages and no NaN.
  - Add test for `SaturationBar` rendering custom roles in `peak_by_role` and custom `subagents_peak_role`.
- Task 5.5: In `web/src/views/TicketDeepDive.test.tsx`:
  - Add test asserting autopsy view displays custom roles in turn velocity, role table, and model table.

## 4. Adversarial Testing Strategy for Tester

### Pure Domain Unit Tests (backend/internal/domain/aggregations_test.go)
- Test Case 1: Dynamic Subagent Effort Distribution
  - Create ledger with invocations for `orchestrator` ($1.00), `coder` ($2.00), `advisor` ($1.00), and `po` ($1.00). Total subagent cost is $4.00.
  - Assert `RoleEffortDistributionPct` has coder at 50.0%, advisor at 25.0%, and po at 25.0%. Assert orchestrator is excluded from distribution.
- Test Case 2: Dynamic Subagent Turn Velocity
  - Create subagent invocations with `ParentID` set: `advisor` with 4 turns, `po` with 8 turns, `coder` with 6 turns.
  - Assert `avg_turns_per_invocation` is 6.0 and `peak_turns` is 8.
  - Assert `by_role` contains advisor (4.0), po (8.0), and coder (6.0).
- Test Case 3: Subagent Peak Context Tracking with Custom Role
  - Create invocations where `orchestrator` has 50k tokens, `coder` has 120k tokens, and `advisor` has 180k tokens.
  - Assert `SubagentsPeakTokens` is 180,000 and `SubagentsPeakRole` is `advisor`.
- Test Case 4: Peak Context Tie-Breaking Precedence
  - Create invocations where `coder` has 100k tokens and `advisor` has 100k tokens. Assert `SubagentsPeakRole` is `coder` (canonical precedence).
  - Create invocations where `po` has 100k tokens and `advisor` has 100k tokens (no canonical subagents). Assert `SubagentsPeakRole` is `advisor` (alphabetical tie-break).
- Test Case 5: Orchestrator-Only Session
  - Create ledger containing only orchestrator invocations.
  - Assert `SubagentsPeakTokens` is 0, `SubagentsPeakRole` is `none`, `RoleEffortDistributionPct` is empty map, and `avg_turns_per_invocation` is 0.0.
- Test Case 6: Model Metrics Aggregation and Secondary Tie-Breaking
  - Create invocations for `model-z` ($1.00), `model-a` ($1.00), `model-top` ($5.00).
  - Assert `ComputeModelMetrics` returns `model-top` first ($5.00), followed by `model-a` ($1.00), and `model-z` ($1.00) due to secondary alphabetical sort.

### Adapter Formatting Tests (backend/internal/adapters/formatters_test.go)
- Test Case 7: Markdown Telemetry Formatting with Custom Subagents
  - Generate markdown summary from ledger containing `advisor` and `po`.
  - Assert output string contains `Advisor` and `Po` in `Role Effort Distribution (Subagents)` and `Subagent Turn Velocity & Thrashing`.

### Frontend Presentation Unit Tests (web/src/)
- Test Case 8: Theme Deterministic Color Hashing
  - Verify `getRoleColor('orchestrator')` matches purple theme token.
  - Verify `getRoleColor('advisor')` and `getRoleColor('advisor')` produce the identical hex color.
  - Verify `getModelColor('claude-3-7-sonnet')` returns valid 7-character hex color.
- Test Case 9: DistributionCard with Dynamic Roles and Models
  - Render `DistributionCard` with items containing `advisor`, `po`, and `researcher`.
  - Assert all three role names and percentages render cleanly with progress fill elements.
  - Render `DistributionCard` with arbitrary models (`claude-3-7-sonnet`, `gemini-2.5-pro`). Assert correct sorting and labels.
- Test Case 10: SaturationBar with Dynamic Subagent Peak
  - Render `SaturationBar` with `subagents_peak_role: 'advisor'` and `subagents_peak_tokens: 180000`.
  - Assert screen displays text `advisor (180k)`.
  - Assert `peak_by_role` renders items for all custom roles in the grid.

## 5. Review Criteria for Reviewer

### Axis 1: Specification and Contract Alignment
- Are non-canonical agent roles (e.g. `advisor`, `po`, `researcher`) included in `role_costs`, `subagent_turn_velocity`, `role_effort_distribution_pct`, and `context_window_health`?
- Are arbitrary model identifiers aggregated and displayed in `Model Cost Distribution` and telemetry tables?
- Do canonical roles maintain priority display ordering, followed by custom roles sorted alphabetically?
- Does `ComputeModelMetrics` sort by cost descending with deterministic secondary tie-breaking on model name?
- Does frontend display custom roles and models with consistent styling, deterministic colors, and accessible contrast?

### Axis 2: Engineering Standards and Technical Invariants
- FCIS Discipline: Is `backend/internal/domain/` 100% pure functional core with zero I/O, disk, network, or clock imports?
- Strong Typing: Is there zero usage of `any` or `interface{}` in Go and TypeScript? Is `AgentRole` widened to string in `web/src/domain/types.ts`?
- Value Semantics: Are domain models immutable structs?
- Error and Zero Division Defenses: Does code defensively handle empty invocations, zero subagent turns, zero cost, and empty role strings without panics or NaN?
- Zero Emojis Mandate: Are all code files, comments, docstrings, UI labels, tooltips, test names, and markdown artifacts completely free of emojis?
- Quality Gates: Do `make check` and `make test` pass with zero warnings or errors?
