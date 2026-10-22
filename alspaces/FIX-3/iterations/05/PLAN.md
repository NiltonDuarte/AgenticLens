# Implementation Plan: FIX-3 Iteration 05 - Primitive Obsession Remediation for Declared Domain Enums

Authoring Path: alspaces/FIX-3/iterations/05/PLAN.md

---

## 1. Goal and Iteration Scope

Resolve primitive obsession across the Go backend domain core and TypeScript frontend domain contracts for four existing declared enums identified in the audit (`alspaces/FIX-3/PRIMITIVE_OBSESSION_AUDIT.md`):

1. Context Window Health Classification:
   - Go: Replace raw `string` in domain aggregation structs and comparison rows with `domain.ContextHealthStatus`.
   - TypeScript: Replace raw `string` in domain interfaces and presentation selector props with `ContextHealthStatus`.
   - Align aggregation calculations to consistently invoke `GetBaseHealthStatus` instead of serializing formatted strings or raw string casts.

2. Agent Role Classification:
   - Go: Replace raw `string` in `InvocationFact`, `EventFact`, `LiveMessageFact`, `RoleMetric`, `ThreadRoleMetric`, and `ContextWindowHealth` with `domain.AgentRole`. Make `OrderRoles` and `OrderSubagentRoles` generic over string types so slices of `AgentRole` are accepted natively without conversion.
   - TypeScript: Define `CanonicalAgentRole` for canonical personas (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`). Define `AgentRole` as an autocomplete union `CanonicalAgentRole | (string & {})` to provide IDE autocomplete while preserving non-breaking support for custom subagent roles. Update domain interfaces, telemetry event facts, and UI props to use `AgentRole`.

3. Tool and Part Failure Severity:
   - Go: Update `LivePartFact.Severity` to use `domain.FailureSeverity`. Update `ClassifyToolFailure` signature to return `(bool, FailureSeverity, *string)`. Type live anomaly alert severity with named domain type `LiveAnomalySeverity`.
   - TypeScript: Export `FailureSeverity` union (`'none' | 'low' | 'medium' | 'high'`). Update `LivePartFact.severity` to use `FailureSeverity`. Rectify mock test payloads containing invalid severity values.

4. Conventional Ticket Type Classification:
   - Go: Replace raw `*string` with `*domain.ConventionalTicketType` in `AgenticFactLedger`, `AgenticMetricsLedger`, `TicketComparisonRow`, and Git branch inference helpers. Retain raw `*string` on external boundary DTOs (`schemas.go`) with explicit domain mapping.
   - TypeScript: Align `StandardTicketType` to encompass canonical conventional commit types. Update `ticket_type` fields across `AgenticFactLedger`, `AgenticMetricsLedger`, and `TicketComparisonRow` to `StandardTicketType | string | null`.

Non-goals:
- Introducing branded or nominal entity identifier types (`TicketID`, `SessionID`, `MessageID`, `PartID`); these are designated for future architectural ticket `REF-1`.
- Modifying SQLite schema migrations or OpenCode database column types.
- Altering CLI flag names or HTTP REST query parameter names.

---

## 2. Grounded Reference Architecture and Symbols

### Go Functional Core (`backend/internal/domain/`)
- `enums.go`:
  - `ContextHealthStatus`: Existing string type (`ContextHealthHealthy`, `ContextHealthElevated`, `ContextHealthBloated`, `ContextHealthCritical`).
  - `AgentRole`: Existing string type (`RoleOrchestrator`, `RolePlanner`, `RoleCoder`, `RoleTester`, `RoleReviewer`, `RoleUnassigned`).
  - `ConventionalTicketType`: Existing string type (`TicketTypeFeat`, `TicketTypeFix`, `TicketTypeDocs`, `TicketTypeChore`, `TicketTypeRefactor`, `TicketTypeTest`, `TicketTypeCI`, `TicketTypePerf`, `TicketTypeBuild`).
  - `OrderRoles` and `OrderSubagentRoles`: Sorting helpers for agent role sequences.
- `constants.go`:
  - `GetBaseHealthStatus(toks int64, ceiling int64) ContextHealthStatus`: Pure function returning base context health enum.
  - `GetHealthLabel(toks int64, ceiling int64) string`: Pure presentation helper formatting labels with percentages for CLI markdown rendering.
- `models.go`:
  - `InvocationFact`: Contains `Role string`.
  - `EventFact`: Contains `Role string`.
  - `GroupedMetric`: Contains `ContextHealth string`.
  - `ContextWindowHealth`: Contains `Status string`, `SubagentsPeakRole string`, `PeakByRole map[string]int64`.
  - `SubagentTurnVelocity`: Contains `ByRole map[string]float64`.
  - `ModelMetric`: Contains `ContextHealth string`.
  - `RoleMetric`: Contains `Role string`, `ContextHealth string`.
  - `ThreadRoleMetric`: Contains `Role string`.
  - `FleetSummary`: Contains `RoleCosts map[string]float64`.
  - `AgenticFactLedger`: Contains `TicketType *string`.
  - `AgenticMetricsLedger`: Contains `TicketType *string`.
  - `TicketComparisonRow`: Contains `ContextHealthStatus string`, `TicketType *string`.
- `live.go`:
  - `FailureSeverity`: Existing string type (`SeverityNone`, `SeverityLow`, `SeverityMedium`, `SeverityHigh`).
  - `LivePartFact`: Contains `Severity string`.
  - `LiveMessageFact`: Contains `Role string`.
  - `LiveSessionNode`: Already uses `Role AgentRole`.
  - `ClassifyToolFailure`: Currently returns `(bool, string, *string)`.
- `live_anomaly.go`:
  - Constants `LiveSeverityWarning`, `LiveSeverityHigh`, `LiveSeverityCritical`.
  - `LiveAnomalyAlert`: Contains `Severity string`.
- `anomalies.go`:
  - `AnomalyContext`: Contains `Role *string`, `SecondaryRole *string`.
- `aggregations.go`:
  - Computes `GroupedMetric`, `WorkflowEfficiency`, `RoleMetric`, `ModelMetric`, `ThreadExecutionRecord`, `TicketComparisonRow`.

### Go Imperative Shell (`backend/internal/adapters/` and `backend/cmd/`)
- `adapters/schemas.go`:
  - Boundary DTOs: `InvocationFactDTO`, `EventFactDTO`, `AgenticFactLedgerDTO`, `AgenticMetricsLedgerDTO`, `AnomalyContextDTO`.
  - Mapping functions: `ToDomain`, `FromDomain`, `SerializeLedger`.
- `adapters/repository.go`:
  - `ExistingLedgerFacts`: Contains `TicketType *string`.
  - `LoadExistingFactLedger`: Parses on-disk JSON into `ExistingLedgerFacts`.
- `adapters/git/inference.go`:
  - `InferTicketTypeFromBranchList`, `InferTicketType`: Return conventional ticket type pointers.
- `adapters/live/service.go`:
  - `processRawPart`, `convertToEventFact`: Map SQLite rows into `LivePartFact` and `EventFact`.
- `adapters/opencode/extractor.go`:
  - `ExtractFacts`: Maps SQLite messages and parts into `InvocationFact` and `EventFact`.
- `adapters/formatters.go`:
  - Formats summary tables using `totals.ContextHealth` and `r.PeakContextTokens`.
- `cmd/agentic-lens/main.go`:
  - CLI flag `--ticket-type` resolution and ledger synthesis.

### TypeScript Domain Contracts (`web/src/domain/` and `web/src/types/`)
- `web/src/types/telemetry.ts`:
  - `EventFact`: Contains `role: string`.
  - `AnomalyContext`: Contains `role?: string`, `secondary_role?: string`.
- `web/src/domain/types.ts`:
  - `ContextHealthStatus`: Existing string union `'HEALTHY' | 'ELEVATED' | 'BLOATED' | 'CRITICAL'`.
  - `AgentRole`: Currently `type AgentRole = string`.
  - `StandardTicketType`: Existing string union.
  - Struct interfaces: `InvocationFact`, `GroupedMetric`, `ContextWindowHealth`, `ModelMetric`, `RoleMetric`, `ThreadRoleMetric`, `AgenticFactLedger`, `AgenticMetricsLedger`, `TicketComparisonRow`.
- `web/src/domain/live.ts`:
  - `LiveAnomalySeverity`: Existing string union `'WARNING' | 'HIGH' | 'CRITICAL'`.
  - `LivePartFact`: Contains `severity: string`.
  - `LiveMessageFact`: Contains `role: string`.
  - `LiveSessionNode`: Contains `role: string`.
- `web/src/domain/formatters.ts`:
  - `canonicalTicketType`: Alias resolution function.
  - `filterLedgersByTicketTypes`, `filterComparisonRows`.
- `web/src/components/`:
  - `TicketSelector.tsx`: Contains `context_health_status?: string`.
  - `FailedToolAlertCard.tsx`: Uses `part.severity` and `role?: string`.
  - `ThreadViewer.tsx`: Uses `m.context_health` for CSS class mapping.
  - `TicketTable.tsx`, `LandingPage.tsx`, `TicketComparisons.tsx`: Display `context_health_status`.

---

## 3. Architectural Specifications and Contracts

### Specification A: Context Health Status Typings
1. Backend Go Domain:
   - In `backend/internal/domain/models.go`:
     - `GroupedMetric.ContextHealth`: Substitute type from `string` to `ContextHealthStatus`.
     - `ContextWindowHealth.Status`: Substitute type from `string` to `ContextHealthStatus`.
     - `ModelMetric.ContextHealth`: Substitute type from `string` to `ContextHealthStatus`.
     - `RoleMetric.ContextHealth`: Substitute type from `string` to `ContextHealthStatus`.
     - `TicketComparisonRow.ContextHealthStatus`: Substitute type from `string` to `ContextHealthStatus`.
   - In `backend/internal/domain/aggregations.go`:
     - In `computeGroupedMetric`: Assign `ContextHealth: healthStatus` directly (value returned by `GetBaseHealthStatus`), eliminating `string(...)` cast.
     - In `ComputeWorkflowEfficiency`: Assign `Status: healthStatus` directly without cast.
     - In `ComputeRoleMetrics`: Change assignment from `GetHealthLabel` to `GetBaseHealthStatus(acc.peakContextTokens, HealthyCeilingTokens)`, directly assigning `ContextHealthStatus`.
     - In `ComputeModelMetrics`: Change assignment from `GetHealthLabel` to `GetBaseHealthStatus(acc.peakContextTokens, HealthyCeilingTokens)`, directly assigning `ContextHealthStatus`.
     - In `ComputeThreadSummaries`: Change assignment from `GetHealthLabel` to `GetBaseHealthStatus(peak, HealthyCeilingTokens)`, directly assigning `ContextHealthStatus`.
     - In `BuildComparisonRows`: Assign `ContextHealthStatus: cwh.Status` directly.
     - In `CalculateRoleEffortRollup`: Change assignment from `GetHealthLabel` to `GetBaseHealthStatus`.
2. Frontend TypeScript Contracts:
   - In `web/src/domain/types.ts`:
     - `GroupedMetric.context_health`: Change from `string` to `ContextHealthStatus`.
     - `ContextWindowHealth.status`: Change from `string` to `ContextHealthStatus`.
     - `ModelMetric.context_health`: Change from `string` to `ContextHealthStatus`.
     - `RoleMetric.context_health`: Change from `string` to `ContextHealthStatus`.
     - `TicketComparisonRow.context_health_status`: Change from `string` to `ContextHealthStatus`.
   - In `web/src/components/TicketSelector.tsx`:
     - `TicketSelectorProps.context_health_status`: Change from `string | undefined` to `ContextHealthStatus | undefined`.

### Specification B: Agent Role Typings
1. Frontend TypeScript Contracts:
   - In `web/src/types/telemetry.ts`:
     - Export union `CanonicalAgentRole`: `'orchestrator' | 'planner' | 'coder' | 'tester' | 'reviewer'`.
     - Export autocomplete union `AgentRole`: `CanonicalAgentRole | (string & {})`.
     - In `EventFact`: Change `role: string` to `role: AgentRole`.
     - In `AnomalyContext`: Change `role?: string` to `role?: AgentRole | undefined`, and `secondary_role?: string` to `secondary_role?: AgentRole | undefined`.
   - In `web/src/domain/types.ts`:
     - Re-export `CanonicalAgentRole` and `AgentRole` from `../types/telemetry`.
     - Update `CANONICAL_ROLE_ORDER` to `readonly CanonicalAgentRole[]`.
     - Update `SUBAGENT_ROLE_ORDER` to `readonly CanonicalAgentRole[]`.
     - Update `isCanonicalRole` signature to type predicate: `(role: string): role is CanonicalAgentRole`.
     - In `InvocationFact`: Change `role: string` to `role: AgentRole`.
     - In `RoleMetric`: Change `role: string` to `role: AgentRole`.
     - In `ThreadRoleMetric`: Change `role: string` to `role: AgentRole`.
     - In `ContextWindowHealth`: Change `subagents_peak_role: string` to `subagents_peak_role: AgentRole`. Change `peak_by_role` to `Readonly<Record<AgentRole, number>>`.
     - In `SubagentTurnVelocity`: Change `by_role` to `Readonly<Record<AgentRole, number>>`.
     - In `FleetSummary`: Change `role_costs` to `Readonly<Record<AgentRole, number>>`.
   - In `web/src/domain/live.ts`:
     - In `LiveMessageFact`: Change `role: string` to `role: AgentRole`.
     - In `LiveSessionNode`: Change `role: string` to `role: AgentRole`.
   - In `web/src/components/FailedToolAlertCard.tsx`:
     - Update `role?: string | undefined` to `role?: AgentRole | undefined`.
2. Backend Go Domain:
   - In `backend/internal/domain/enums.go`:
     - Retain `type AgentRole string` and canonical constants (`RoleOrchestrator`, etc.).
     - Update generic signatures: `func OrderRoles[T ~string](roles []T, customOrder ...[]string) []T` and `func OrderSubagentRoles[T ~string](roles []T, customOrder ...[]string) []T`. This allows accepting `[]AgentRole` and `[]string` without conversions or allocations.
   - In `backend/internal/domain/models.go`:
     - `InvocationFact.Role`: Change from `string` to `AgentRole`.
     - `EventFact.Role`: Change from `string` to `AgentRole`.
     - `RoleMetric.Role`: Change from `string` to `AgentRole`.
     - `ThreadRoleMetric.Role`: Change from `string` to `AgentRole`.
     - `ContextWindowHealth.SubagentsPeakRole`: Change from `string` to `AgentRole`.
     - `ContextWindowHealth.PeakByRole`: Change from `map[string]int64` to `map[AgentRole]int64`.
     - `SubagentTurnVelocity.ByRole`: Change from `map[string]float64` to `map[AgentRole]float64`.
     - `FleetSummary.RoleCosts`: Change from `map[string]float64` to `map[AgentRole]float64`.
   - In `backend/internal/domain/live.go`:
     - `LiveMessageFact.Role`: Change from `string` to `AgentRole`.
   - In `backend/internal/domain/anomalies.go`:
     - `AnomalyContext.Role`: Change from `*string` to `*AgentRole`.
     - `AnomalyContext.SecondaryRole`: Change from `*string` to `*AgentRole`.
3. Backend Go Adapters:
   - In `backend/internal/adapters/schemas.go`:
     - `InvocationFactDTO.ToDomain`: Map `Role: domain.AgentRole(role)`.
     - `InvocationFactDTOFromDomain`: Map `Role: string(inv.Role)`.
     - `EventFactDTO.ToDomain`: Map `Role: domain.AgentRole(role)`.
     - `EventFactDTOFromDomain`: Map `Role: string(fact.Role)`.
     - `AnomalyContextDTO.ToDomain`: Convert `Role` and `SecondaryRole` pointer values to `*domain.AgentRole`.
     - `AnomalyContextDTOFromDomain`: Convert `Role` and `SecondaryRole` pointer values to `*string`.
   - In `backend/internal/adapters/live/service.go`:
     - Change `sessionRoleMap` to `map[string]domain.AgentRole`. Assign directly to `EventFact.Role` and `LiveMessageFact.Role`.
   - In `backend/internal/adapters/opencode/extractor.go`:
     - Change `sessionRoleMap` to `map[string]domain.AgentRole`. Assign directly to `EventFact.Role` and `InvocationFact.Role`.

### Specification C: Failure Severity Typings
1. Backend Go Domain:
   - In `backend/internal/domain/live.go`:
     - `LivePartFact.Severity`: Change from `string` to `FailureSeverity`.
     - `ClassifyToolFailure`: Update signature return type to `(bool, FailureSeverity, *string)`. Return domain constants `SeverityNone`, `SeverityHigh`, `SeverityMedium`, `SeverityLow` without string casting.
   - In `backend/internal/domain/live_anomaly.go`:
     - Define `type LiveAnomalySeverity string`.
     - Type constants `LiveSeverityWarning`, `LiveSeverityHigh`, and `LiveSeverityCritical` with `LiveAnomalySeverity`.
     - Update `LiveAnomalyAlert.Severity` from `string` to `LiveAnomalySeverity`.
2. Backend Go Adapters:
   - In `backend/internal/adapters/live/service.go`:
     - Update local `severity` variable in `processRawPart` to type `domain.FailureSeverity`.
     - Assign fallback `severity = domain.SeverityNone` without string casting.
3. Frontend TypeScript Contracts:
   - In `web/src/domain/live.ts`:
     - Export union `FailureSeverity`: `'none' | 'low' | 'medium' | 'high'`.
     - In `LivePartFact`: Change `readonly severity: string` to `readonly severity: FailureSeverity`.
     - Re-export `FailureSeverity` in `web/src/domain/types.ts`.
   - In `web/src/components/FailedToolAlertCard.tsx`:
     - Adjust default severity fallback to `'high'` before uppercase conversion.
   - In `web/src/domain/live.test.ts`:
     - In mock test payload (around line 752 and 761), update invalid tool severity `'critical'` to `'high'`. Update line 780 expectation from `'critical'` to `'high'` to align with domain invariants.

### Specification D: Conventional Ticket Type Typings
1. Backend Go Domain:
   - In `backend/internal/domain/models.go`:
     - `AgenticFactLedger.TicketType`: Change from `*string` to `*ConventionalTicketType`.
     - `AgenticMetricsLedger.TicketType`: Change from `*string` to `*ConventionalTicketType`.
     - `TicketComparisonRow.TicketType`: Change from `*string` to `*ConventionalTicketType`.
   - In `backend/internal/domain/migration/validator.go`:
     - Define generic helper `ptrEqual[T comparable](a, b *T) bool` to compare pointer equality across `TicketType`, `TicketID`, and integer fields without type casting.
2. Backend Go Adapters:
   - In `backend/internal/adapters/git/inference.go`:
     - Update `InferTicketTypeFromBranchList` return type from `*string` to `*domain.ConventionalTicketType`.
     - Update `InferTicketType` return type from `*string` to `*domain.ConventionalTicketType`.
   - In `backend/internal/adapters/repository.go`:
     - In `ExistingLedgerFacts`: Change `TicketType` from `*string` to `*domain.ConventionalTicketType`.
     - In `LoadExistingFactLedger`: Convert raw `raw.TicketType` string pointer to `*domain.ConventionalTicketType`.
   - In `backend/internal/adapters/schemas.go`:
     - In `AgenticFactLedgerDTO.ToDomain`: Map `*dto.TicketType` to `*domain.ConventionalTicketType`.
     - In `AgenticFactLedgerDTOFromDomain`: Map `*ledger.TicketType` to `*string`.
     - In `AgenticMetricsLedgerDTO.ToDomain`: Map `*dto.TicketType` to `*domain.ConventionalTicketType`.
     - In `SerializeLedger`: Map `*ledger.TicketType` to `*string`.
   - In `backend/cmd/agentic-lens/main.go`:
     - Change local variable `activeTicketType` to `*domain.ConventionalTicketType`. Assign return value of `git.InferTicketType` and normalized values directly.
3. Frontend TypeScript Contracts:
   - In `web/src/domain/types.ts`:
     - Update `StandardTicketType` union to encompass canonical conventional commit types: `'feat' | 'fix' | 'feature' | 'bug' | 'chore' | 'docs' | 'refactor' | 'test' | 'ci' | 'perf' | 'build'`.
     - Define `TicketType = StandardTicketType | (string & {})`.
     - In `AgenticFactLedger`: Update `ticket_type` from `string | null` to `TicketType | null`.
     - In `AgenticMetricsLedger`: Update `ticket_type` from `string | null` to `TicketType | null`.
     - In `TicketComparisonRow`: Update `ticket_type` to `TicketType | null`.
   - In `web/src/domain/formatters.ts`:
     - `canonicalTicketType`: Update return type to `StandardTicketType | string`.
     - `filterLedgersByTicketTypes`: Accept `types: readonly (TicketType | string)[]`.
     - `filterComparisonRows`: Accept `types: readonly (TicketType | string)[]`.

---

## 4. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Core Enum Substitutions
- Task 1.1: In `backend/internal/domain/enums.go`:
  - Update `OrderRoles` and `OrderSubagentRoles` to generic function signature `func OrderRoles[T ~string](roles []T, customOrder ...[]string) []T`.
- Task 1.2: In `backend/internal/domain/models.go`:
  - Update `GroupedMetric.ContextHealth` to `ContextHealthStatus`.
  - Update `ContextWindowHealth.Status` to `ContextHealthStatus`.
  - Update `ContextWindowHealth.SubagentsPeakRole` to `AgentRole`.
  - Update `ContextWindowHealth.PeakByRole` to `map[AgentRole]int64`.
  - Update `SubagentTurnVelocity.ByRole` to `map[AgentRole]float64`.
  - Update `ModelMetric.ContextHealth` to `ContextHealthStatus`.
  - Update `RoleMetric.Role` to `AgentRole`, and `ContextHealth` to `ContextHealthStatus`.
  - Update `ThreadRoleMetric.Role` to `AgentRole`.
  - Update `FleetSummary.RoleCosts` to `map[AgentRole]float64`.
  - Update `InvocationFact.Role` to `AgentRole`.
  - Update `EventFact.Role` to `AgentRole`.
  - Update `AgenticFactLedger.TicketType` to `*ConventionalTicketType`.
  - Update `AgenticMetricsLedger.TicketType` to `*ConventionalTicketType`.
  - Update `TicketComparisonRow.ContextHealthStatus` to `ContextHealthStatus`, and `TicketType` to `*ConventionalTicketType`.
- Task 1.3: In `backend/internal/domain/live.go`:
  - Update `LivePartFact.Severity` to `FailureSeverity`.
  - Update `LiveMessageFact.Role` to `AgentRole`.
  - Update `ClassifyToolFailure` return signature to `(bool, FailureSeverity, *string)` and return uncast enum constants.
- Task 1.4: In `backend/internal/domain/live_anomaly.go`:
  - Define `type LiveAnomalySeverity string`.
  - Type constants `LiveSeverityWarning`, `LiveSeverityHigh`, and `LiveSeverityCritical` with `LiveAnomalySeverity`.
  - Update `LiveAnomalyAlert.Severity` to `LiveAnomalySeverity`.
- Task 1.5: In `backend/internal/domain/anomalies.go`:
  - Update `AnomalyContext.Role` and `SecondaryRole` to `*AgentRole`.
- Task 1.6: In `backend/internal/domain/aggregations.go`:
  - Replace `GetHealthLabel` with `GetBaseHealthStatus` across role, model, thread, and rollup metrics.
  - Eliminate all `string(...)` enum wrapper casts when setting `ContextHealth` and `ContextHealthStatus`.

### Phase 2: Go Boundary Adapters and Entrypoint Alignment
- Task 2.1: In `backend/internal/adapters/schemas.go`:
  - Update `ToDomain` and `FromDomain` converters for `InvocationFactDTO`, `EventFactDTO`, `AgenticFactLedgerDTO`, `AgenticMetricsLedgerDTO`, and `AnomalyContextDTO` to perform clean type conversions between wire strings and domain enums.
- Task 2.2: In `backend/internal/adapters/git/inference.go`:
  - Update return types of `InferTicketTypeFromBranchList` and `InferTicketType` to `*domain.ConventionalTicketType`.
- Task 2.3: In `backend/internal/adapters/repository.go`:
  - Update `ExistingLedgerFacts.TicketType` to `*domain.ConventionalTicketType`.
  - Convert parsed JSON string pointers in `LoadExistingFactLedger` to `*domain.ConventionalTicketType`.
- Task 2.4: In `backend/internal/adapters/live/service.go`:
  - Update `sessionRoleMap` to use `domain.AgentRole`.
  - Update `severity` in `processRawPart` to `domain.FailureSeverity`.
- Task 2.5: In `backend/internal/adapters/opencode/extractor.go`:
  - Update `sessionRoleMap` to use `domain.AgentRole`.
- Task 2.6: In `backend/internal/domain/migration/validator.go`:
  - Implement generic pointer comparison helper `ptrEqual` and update ticket type invariant check.
- Task 2.7: In `backend/cmd/agentic-lens/main.go`:
  - Align `activeTicketType` variable to `*domain.ConventionalTicketType`.

### Phase 3: TypeScript Domain Interfaces and Component Alignment
- Task 3.1: In `web/src/types/telemetry.ts`:
  - Define and export `CanonicalAgentRole` and `AgentRole`.
  - Update `EventFact.role` to `AgentRole`.
  - Update `AnomalyContext.role` and `secondary_role` to `AgentRole | undefined`.
- Task 3.2: In `web/src/domain/live.ts`:
  - Export `FailureSeverity = 'none' | 'low' | 'medium' | 'high'`.
  - Update `LivePartFact.severity` to `FailureSeverity`.
  - Update `LiveMessageFact.role` and `LiveSessionNode.role` to `AgentRole`.
- Task 3.3: In `web/src/domain/types.ts`:
  - Re-export `CanonicalAgentRole` and `AgentRole` from `../types/telemetry`.
  - Re-export `FailureSeverity` from `./live`.
  - Update `CANONICAL_ROLE_ORDER` and `SUBAGENT_ROLE_ORDER` to `readonly CanonicalAgentRole[]`.
  - Update `isCanonicalRole` to type guard `role is CanonicalAgentRole`.
  - Update `StandardTicketType` and define `TicketType = StandardTicketType | (string & {})`.
  - Update `GroupedMetric.context_health`, `ContextWindowHealth.status`, `ModelMetric.context_health`, `RoleMetric.context_health`, and `TicketComparisonRow.context_health_status` to `ContextHealthStatus`.
  - Update `InvocationFact.role`, `RoleMetric.role`, `ThreadRoleMetric.role`, and `ContextWindowHealth.subagents_peak_role` to `AgentRole`.
  - Update `ContextWindowHealth.peak_by_role`, `SubagentTurnVelocity.by_role`, and `FleetSummary.role_costs` to `Readonly<Record<AgentRole, number>>`.
  - Update `AgenticFactLedger.ticket_type`, `AgenticMetricsLedger.ticket_type`, and `TicketComparisonRow.ticket_type` to `TicketType | null`.
- Task 3.4: In `web/src/domain/formatters.ts`:
  - Update `canonicalTicketType` return type and filter function parameter types.
- Task 3.5: In `web/src/components/TicketSelector.tsx`:
  - Update `context_health_status` to `ContextHealthStatus | undefined`.
- Task 3.6: In `web/src/components/FailedToolAlertCard.tsx`:
  - Update `role` prop to `AgentRole | undefined`.
  - Ensure severity fallback defaults to `'high'`.

### Phase 4: Test Suite Updates and Assertions Alignment
- Task 4.1: In `backend/internal/domain/aggregations_test.go`:
  - Update comparisons to assert directly against enum constants (`domain.ContextHealthHealthy`, `domain.RoleCoder`, etc.) without `string(...)` wrapping.
  - Update `TestBuildComparisonRows` ticket type fixtures.
- Task 4.2: In `backend/internal/domain/live_test.go`:
  - Update severity assertions to compare against `domain.SeverityNone`, `domain.SeverityHigh`, etc.
- Task 4.3: In `backend/internal/adapters/live/sqlite_test.go` and `backend/internal/adapters/opencode/extractor_test.go`:
  - Align role assertions and fixtures.
- Task 4.4: In `backend/internal/adapters/formatters_test.go`:
  - Use `domain.ContextHealthElevated` and `domain.ContextHealthHealthy` in test ledger totals and role metrics.
- Task 4.5: In `backend/internal/adapters/schemas_test.go`:
  - Verify boundary DTO round-trip tests continue to pass.
- Task 4.6: In `backend/cmd/agentic-lens/main_test.go`:
  - Ensure ticket type classification CLI tests pass with typed pointer fields.
- Task 4.7: In `web/src/domain/live.test.ts`:
  - Update mock JSON payload lines 752 and 761 to use `'high'` instead of invalid `'critical'`, and update line 780 expectation to `'high'`.
- Task 4.8: In `web/src/views/LandingPage.test.tsx`:
  - Update unrecognized status test cast to `as unknown as ContextHealthStatus`.

### Phase 5: Verification and Quality Gates
- Task 5.1: Run `npm run typecheck` to verify zero TypeScript errors under strict mode.
- Task 5.2: Run `npm run lint` to verify zero ESLint errors or warnings.
- Task 5.3: Run `npm test` across all 30 Vitest frontend suites.
- Task 5.4: Run `go test -race ./...` across backend packages.
- Task 5.5: Run `go vet ./...` across backend packages.
- Task 5.6: Run `make build` to confirm binary compilation with embedded frontend assets succeeds.
- Task 5.7: Verify zero emojis exist anywhere in the code or modified files.

---

## 5. Concise Plan Summary

Iteration 05 resolves primitive obsession by replacing raw `string` and `*string` types with existing declared domain enums:
1. `ContextHealthStatus`: Substituted into `GroupedMetric.ContextHealth`, `ContextWindowHealth.Status`, `ModelMetric.ContextHealth`, `RoleMetric.ContextHealth`, and `TicketComparisonRow.ContextHealthStatus` in Go and TypeScript. Aggregations now consistently invoke `GetBaseHealthStatus`.
2. `AgentRole`: Substituted into `InvocationFact.Role`, `EventFact.Role`, `LiveMessageFact.Role`, `RoleMetric.Role`, `ThreadRoleMetric.Role`, and `ContextWindowHealth.SubagentsPeakRole`. Go `OrderRoles` is generalized with generic type constraints, while TypeScript adopts an autocomplete union `CanonicalAgentRole | (string & {})`.
3. `FailureSeverity`: Substituted into `LivePartFact.Severity` and returned directly by `ClassifyToolFailure` without string conversion. `LiveAnomalyAlert.Severity` receives a dedicated named domain type. In TypeScript, invalid `'critical'` tool severities in test fixtures are aligned to `'high'`.
4. `ConventionalTicketType` and `StandardTicketType`: Substituted into `AgenticFactLedger`, `AgenticMetricsLedger`, `TicketComparisonRow`, and Git branch inference helpers. Boundary DTOs retain raw strings to safely parse external wire formats.

The plan contains zero code blocks, enforces strict FCIS layer isolation, preserves zero `any` / zero `interface{}` invariants, and contains zero emojis.
