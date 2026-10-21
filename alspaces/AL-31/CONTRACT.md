# Contract: AL-31 Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests

## 1. Problem Statement

AgenticLens telemetry models session lifecycles as a hierarchical execution tree: a top-level root session coordinates user interactions, requirements, and steering, while child invocations represent delegated subagents assigned to specialized functional roles (such as planner, coder, tester, and reviewer).

Previously, the domain aggregation engine and boundary extraction adapters conflated structural hierarchy (the presence or absence of a parent session identifier: ParentID is nil for root vs ParentID is non-nil for subagents) with semantic persona strings (such as general, orchestrator, build, or coder). In particular:

1. The helper IsSubagentRole in backend/internal/domain/enums.go classified any non-empty role string that was not literally equal to orchestrator as a subagent.
2. When human operators interact directly with OpenCode, the interactive root session is assigned the persona general. Because general does not equal orchestrator, the root coordination session was misclassified as a delegated worker subagent.
3. In backend/internal/domain/aggregations.go, ComputeWorkflowEfficiency looked up root coordination context window metrics exclusively under the fixed key orchestrator in rolePeaks. When the root session was named general or build, OrchestratorPeakTokens and OrchestratorSaturationPct dropped to zero.
4. Concurrently, the root session peak context tokens (such as 123,350 tokens in ticket AL-1) leaked into SubagentsPeakTokens, misidentifying general as the peak subagent role.
5. Invocations from the root session diluted subagent effort distribution percentages (RoleEffortDistributionPct) and corrupted turn velocity calculations.
6. For single-invocation tickets without delegated subagents (such as GEN-1, GEN-2, PO-1, PO-2, and PO-3), SubagentsPeakTokens reported the root session tokens and SubagentsPeakRole reported general instead of correctly reporting zero tokens and none.
7. Eight test cases failed in backend/internal/adapters/parity_test.go because the test baselines expected authentic root session coordination metrics rather than zeroed values or synthetic role remappings.

Ticket AL-31 decouples structural session hierarchy from agent persona names across the domain engine, boundary adapters, formatters, and test suites. Structural hierarchy alone determines whether an invocation is a root coordination session or a delegated subagent.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must satisfy the architectural requirements defined in TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Functional Core (backend/internal/domain/):
   - Pure domain models, mathematical rollups, and enums remain completely free of external dependencies, file input/output, database drivers, system clocks, network calls, or random generators.
   - Aggregations in backend/internal/domain/aggregations.go and role helpers in backend/internal/domain/enums.go operate exclusively on immutable input structs (AgenticFactLedger, InvocationFact, ToolFact, CodeFacts) and return deterministic output structs (WorkflowEfficiency, ContextWindowHealth, GroupedMetric).
   - Defensive guards prevent division by zero across all percentages, ratios, and averages (such as zero total subagent cost, zero child invocations, zero healthy ceiling tokens, or zero tool calls).
2. Imperative Shell (backend/internal/adapters/):
   - Filesystem scanning in backend/internal/adapters/repository.go, SQLite session extraction in backend/internal/adapters/opencode/extractor.go, DTO mapping in backend/internal/adapters/schemas.go, and markdown table rendering in backend/internal/adapters/formatters.go interact with the functional core via explicit domain types.
   - Authentic persona names and structural ParentID pointers extracted from OpenCode database rows must be preserved without synthetic remapping.
3. Presentation Layer (web/src/):
   - TypeScript frontend models in web/src/domain/types.ts and presentation widgets in web/src/components/SaturationBar.tsx preserve existing property contracts (orchestrator_peak_tokens, subagents_peak_tokens, subagents_peak_role) while rendering authentic coordination and subagent metrics.

### Strict Typing and Zero Any
1. Zero any or empty interface across Go backend code and test files. Concrete structs, typed enums, and explicit pointer semantics must be used exclusively.
2. Zero any or unknown casts in TypeScript frontend definitions and test suites.
3. ContextWindowHealth and WorkflowEfficiency structures maintain explicit scalar types (int64, float64, string) and concrete maps.

### Single Binary Distribution and Offline Independence
1. Backend changes must compile cleanly into the standalone single binary via the root Makefile build target.
2. Zero external runtime dependencies introduced into production distribution.

### Zero-Emoji Mandate
1. Zero emojis permitted across all modified Go code, comments, docstrings, schema files, formatters, markdown review tables, test assertions, and workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Domain Core (backend/internal/domain/):
   - Refactor ComputeWorkflowEfficiency in backend/internal/domain/aggregations.go:
     - Classify invocations strictly by structural hierarchy: ParentID is nil defines root coordination sessions; ParentID is non-nil defines delegated subagents.
     - Attribute root coordination peak tokens (OrchestratorPeakTokens) and context saturation percentage (OrchestratorSaturationPct) from the maximum peak context tokens among root invocations (ParentID is nil), irrespective of whether the persona role is general, orchestrator, build, po, or custom.
     - Restrict subagent peak tokens (SubagentsPeakTokens), peak subagent role (SubagentsPeakRole), subagent costs (subCosts), subagent effort distribution percentages (RoleEffortDistributionPct), and subagent turn velocity (SubagentTurnVelocity) strictly to child invocations (ParentID is non-nil).
     - When no child invocations exist (ParentID is non-nil), assign SubagentsPeakTokens to zero, SubagentsPeakRole to the string literal none, and RoleEffortDistributionPct to an empty map.
     - Maintain overall context window health evaluation (GetBaseHealthStatus) based on the maximum peak context tokens observed across all invocations (root and child).
   - Refactor IsSubagentRole and OrderSubagentRoles in backend/internal/domain/enums.go:
     - Discontinue the negative role is not orchestrator test.
     - Define non-subagent root personas (orchestrator, general, build, root, and empty strings) as non-subagents, while classifying canonical worker roles (planner, coder, tester, reviewer) and custom delegated agent roles as subagents.
     - Ensure OrderSubagentRoles deduplicates and orders canonical subagents first, followed by alphabetical ordering of custom subagents, without admitting root coordination personas.
   - Refactor detectSubagentReworkLoops in backend/internal/domain/anomalies.go:
     - Ensure rework loop detection relies on the domain subagent classification rather than an ad-hoc string comparison, preventing root general turns from being misidentified as worker subagent ping-pong transitions.
2. Boundary Adapters (backend/internal/adapters/):
   - In backend/internal/adapters/formatters.go:
     - Update FormatMarkdownMetricsTable to render subagent effort distribution directly from RoleEffortDistributionPct without legacy role exclusion hacks.
     - Ensure markdown rendering cleanly displays root coordination metrics alongside subagent peaks, formatting none and zero tokens gracefully when no subagents exist.
   - In backend/internal/adapters/opencode/extractor.go:
     - Audit and verify that ExtractFacts preserves raw agent persona names and authentic ParentID pointers without synthetic mutation.
   - In backend/internal/adapters/schemas.go:
     - Audit and verify that InvocationFactDTO and AgenticFactLedgerDTO correctly serialize and deserialize ParentID and role strings.
3. Verification and Parity Tests (backend/internal/adapters/ and backend/internal/domain/):
   - Add extensive table-driven unit tests in backend/internal/domain/aggregations_test.go:
     - Single-invocation root sessions with diverse roles (general, build, orchestrator, po, custom) asserting zero subagent tokens, none subagent role, and empty effort distribution.
     - Multi-invocation hierarchical sessions asserting that root general tokens do not leak into SubagentsPeakTokens, SubagentsPeakRole, or RoleEffortDistributionPct.
     - Unit tests for IsSubagentRole asserting false for general, build, root, orchestrator, empty strings, and true for canonical and custom subagents.
     - Unit tests for OrderSubagentRoles confirming exclusion of root personas.
   - Update backend/internal/adapters/parity_test.go:
     - Adjust reference baseline fixtures for tickets with authentic general root sessions (AL-1, AL-10, AL-11, AL-12, GEN-1, GEN-2, PO-1, PO-2, PO-3) so that PeakByRole reflects authentic general personas and OrchestratorPeakTokens matches the root session context peak.
     - Adjust fleet role cost expectations in TestFleetSummaryAndComparisonParity to reflect authentic split between orchestrator and general personas.
     - Verify 100% test pass rate across all parity test suites.

### Out of Scope
1. Database Schema Alterations: No schema changes to OpenCode SQLite tables.
2. Telemetry Schema Version Bump: Schema version remains at version 2. The JSON field names (orchestrator_peak_tokens, subagents_peak_tokens, subagents_peak_role) remain stable to preserve backward compatibility with all web views and historical ledgers.
3. Frontend Component Rewrites: Presentation components in web/src/components/ and views in web/src/views/ already consume these properties and require no structural alterations.

---

## 4. Architectural Invariants and Behavioral Specifications

### Hierarchy Partitioning Rule
An invocation is classified strictly by its structural relationship in the session tree:
- Root Coordination Invocation: InvocationFact.ParentID == nil.
- Delegated Worker Subagent Invocation: InvocationFact.ParentID != nil.

### Root Coordination Metrics
- OrchestratorPeakTokens is the maximum value of Tokens.PeakContext across all invocations where ParentID == nil. If no invocations have ParentID == nil, OrchestratorPeakTokens is zero.
- OrchestratorSaturationPct is the rounded percentage of OrchestratorPeakTokens relative to HealthyCeilingTokens (200,000 tokens), defended against zero or negative ceiling values.

### Subagent Delegation Metrics
- SubagentsPeakTokens is the maximum value of Tokens.PeakContext across all invocations where ParentID != nil. If no invocations have ParentID != nil, SubagentsPeakTokens is zero.
- SubagentsPeakRole is the role string of the subagent invocation exhibiting SubagentsPeakTokens. If no invocations have ParentID != nil, SubagentsPeakRole is none. If multiple subagents share the peak value, canonical subagent order breaks ties.
- RoleEffortDistributionPct maps each subagent role to its percentage of total subagent inference spend (CostUSD across invocations where ParentID != nil). If total subagent cost is zero or no child invocations exist, the map is empty.
- SubagentTurnVelocity aggregates average turns, peak turns, and per-role turn averages strictly for invocations where ParentID != nil.

### Role Classification Rule
- Non-Subagent Root Roles: orchestrator, general, build, root, and empty/whitespace strings.
- Subagent Worker Roles: planner, coder, tester, reviewer, and any custom non-root persona.

---

## 5. Acceptance Criteria

1. Invocations with ParentID == nil are strictly excluded from SubagentsPeakTokens, SubagentsPeakRole, subagent costs, role effort distribution percentages, and subagent turn velocity.
2. Root sessions configured with persona general, build, orchestrator, or custom roles correctly populate OrchestratorPeakTokens and OrchestratorSaturationPct with the root session context window peak.
3. Sessions containing only root invocations (ParentID == nil) evaluate to SubagentsPeakTokens = 0, SubagentsPeakRole = none, RoleEffortDistributionPct = empty map, and SubagentTurnVelocity with zero average/peak turns and empty role map.
4. Delegated subagent metrics accurately aggregate child invocations (ParentID != nil) without dilution or inflation from root session turns or costs.
5. In enums.go, IsSubagentRole returns false for general, build, root, orchestrator, and whitespace strings, while returning true for canonical subagents and custom delegated worker roles.
6. In enums.go, OrderSubagentRoles discards root personas and orders canonical subagents first, followed by alphabetical order for custom subagents.
7. In formatters.go, FormatMarkdownMetricsTable outputs clean markdown metrics tables displaying Orchestrator peak, subagent peaks, and N/A for empty subagent effort distribution without panics or undefined string artifacts.
8. All unit tests in backend/internal/domain/aggregations_test.go, backend/internal/domain/enums_test.go, and backend/internal/adapters/parity_test.go pass cleanly.
9. Go static analysis (go vet ./...) and race-detected tests (go test -race ./...) pass with zero errors.
10. Single binary compilation (make build) produces an executable without errors.
11. Zero any types and zero emojis across all code, tests, documentation, and commit messages.
