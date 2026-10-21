# feat(aggregations): dynamic role and model aggregations and presentation (AL-24)

## Overview

Refactors backend domain aggregations and frontend visualizations to support arbitrary, dynamic agent roles and model names without truncation or data loss. Eliminates static role enumerations from rollup loops, treating all non-orchestrator roles as subagents. Preserves raw agent personas from session traces in `agentic_metrics.json`, while applying configured `role_mappings` during domain transformation (`DeriveMetricsLedger`) and presentation (telemetry summary and web dashboard). Implements deterministic secondary model tie-breaking and accessible high-contrast color hashing for dark theme presentation.

- Contract: [workspaces/AL-24/CONTRACT.md](workspaces/AL-24/CONTRACT.md)
- Telemetry: [workspaces/AL-24/SUMMARY.md](workspaces/AL-24/SUMMARY.md)

## Summary of Changes

### Domain Core Layer (`backend/internal/domain/`)
- Role Classification & Ordering: Added pure helper functions `IsSubagentRole`, `OrderRoles`, and `OrderSubagentRoles` in `enums.go`.
- Dynamic Workflow Efficiency: Refactored `ComputeWorkflowEfficiency` in `aggregations.go` to compute subagent effort distribution and turn velocity across all discovered non-orchestrator roles.
- Context Health Tie-Breaking: Implemented deterministic peak subagent context selection with canonical priority followed by alphabetical custom role ordering.
- Deterministic Model Sorting: Updated `ComputeModelMetrics` to sort primarily by `CostUSD` descending with secondary tie-breaking on `Model` name ascending.
- Role Mapping Transformation: Extended `DeriveMetricsLedger` to accept configured `RoleMappings`, transforming roles during projection while preserving raw unmutated facts in `AgenticMetricsLedger.Facts`.
- Orchestrator Rollup: Mapped roles resolving to `RoleOrchestrator` are automatically excluded from subagent turn velocity and effort distributions.

### Boundary Adapters Layer (`backend/internal/adapters/`, `backend/cmd/`)
- Raw Persona Preservation: Updated `ExtractFacts` in `opencode/extractor.go` to record exact session personas into `InvocationFact.Role` and `EventFact.Role` without premature coercion to `orchestrator`.
- Boundary Serialization: Verified DTO conversion in `schemas.go` preserves non-canonical roles and arbitrary model strings.
- Repository & CLI Integration: Updated `ScanWorkspaces`, `LoadTicketWorkspace`, and `runGenerate` to load `role_mappings` from `.agenticlens/config.yaml`, deriving mapped metrics while saving raw facts to JSON.
- Markdown Telemetry Formatter: Refactored `FormatMarkdownMetricsTable` in `formatters.go` to dynamically format effort distribution and turn velocity using `domain.OrderSubagentRoles`.

### Presentation Layer (`web/src/`)
- Open Agent Role Type: Widened `AgentRole` from closed union to `type AgentRole = string` in `domain/types.ts` with helper `isCanonicalRole`.
- Deterministic Theming: Added polynomial string hashing (`hashStringToColor`) and palette mappings (`getRoleColor`, `getModelColor`) in `utils/theme.ts`.
- Component Enhancements:
  - `DistributionCard`: Added role and model color indicators and styled track fills with defensive zero-division checks.
  - `SaturationBar`: Rendered peak context breakdown across all active roles and supported dynamic subagent peak role labels.
  - `TicketDeepDive`: Generalized iterations tooltip and rendered dynamic color badges in turn velocity and resource consumption tables.
