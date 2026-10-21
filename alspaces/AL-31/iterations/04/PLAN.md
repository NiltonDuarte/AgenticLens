# Implementation Plan: AL-31 Generic Configurable Agent Roles, Dynamic Role Ordering, and Fragile Title Inference Removal (Iteration 04)

## 1. Goal and Iteration Scope

Iteration 04 executes three critical architectural decoupling and cleanup directives:
1. De-hardcode agent role resolution in the domain core. The platform must accept any arbitrary role configured without hardcoding repo-specific personas (such as orchestrator, build, root, planner, coder, tester, or reviewer) inside domain switch cases. External personas like developer, researcher, or architect must resolve strictly via user-defined configuration mappings, falling back cleanly to their raw cleaned string or unassigned when empty.
2. Replace hardcoded role ordering with configurable role priority. Static priority slices in the domain layer are replaced by configurable role ordering loaded from project configuration, defaulting to alphabetical sorting when unconfigured.
3. Completely remove fragile ticket title inference. Heuristic title extraction from markdown headers and git commit subjects is deleted. Ticket titles will originate strictly from explicit user input or cached ledger facts.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation modifies and verifies the following Go domain models, boundary adapters, CLI commands, and test suites:

### Domain Core (backend/internal/domain/)
- backend/internal/domain/config.go:
  - Refactor RoleMappings.ResolveRole: Remove the hardcoded switch block matching orchestrator, build, root, planner, coder, tester, and reviewer. Check the mappings table using exact match, raw match, and case-insensitive match. If found, return the mapped role converted to lowercase. If not found in the mapping table, return AgentRole(cleaned) directly. If the cleaned string is empty, return RoleUnassigned (the literal string unassigned).
  - Update ProjectConfig struct: Add RoleOrder slice of strings with JSON tag role_order and YAML tag role_order.
  - Update DefaultProjectConfig: Include the baseline role ordering (orchestrator, planner, coder, tester, reviewer) in RoleOrder, while preserving the baseline mapping table in RoleMappings.
- backend/internal/domain/enums.go:
  - Define RoleUnassigned as AgentRole with value unassigned.
  - Deprecate or remove CanonicalRoleOrder and SubagentRoleOrder.
  - Refactor OrderRoles: Accept an input slice of roles and an optional variadic custom order slice. Deduplicate and trim input role strings, filtering out empty strings. If a custom order is provided and non-empty, order roles matching the custom order first in that exact sequence, followed by any remaining custom roles sorted in ascending alphabetical order. If no custom order is provided, sort all deduplicated input roles in ascending alphabetical order.
  - Refactor OrderSubagentRoles: Accept an input slice of candidate subagent roles and an optional variadic custom order slice. Deduplicate and trim candidate role strings. Apply the identical priority logic: if custom order is provided, place matching roles first in that sequence followed by remaining roles sorted alphabetically; if no custom order is provided, sort all candidate roles in ascending alphabetical order.
- backend/internal/domain/aggregations.go:
  - Refactor ComputeWorkflowEfficiency: Accept optional variadic role order slice. Pass this order to OrderSubagentRoles when determining peak subagent role tie-breaking and subagent turn velocity role ordering.
  - Refactor ComputeRoleMetrics: Accept optional variadic role order slice and pass to OrderRoles.
  - Refactor ComputeThreadSummaries: Accept optional variadic role order slice and pass to OrderRoles for thread role ordering.
  - Introduce DeriveMetricsLedgerWithConfig: Accepts an AgenticFactLedger and a complete ProjectConfig. Applies RoleMappings to invocations and events, and passes RoleOrder into ComputeWorkflowEfficiency, ComputeRoleMetrics, and ComputeThreadSummaries.
  - Maintain DeriveMetricsLedger: Preserves backward compatibility by accepting an AgenticFactLedger and optional RoleMappings, delegating to DeriveMetricsLedgerWithConfig with DefaultProjectConfig.
- backend/internal/domain/config_test.go:
  - Update TestRoleMappings_ResolveRole and TestRoleMappings_ResolveRole_CaseInsensitive: Test that unmapped roles (such as developer) return their verbatim cleaned strings rather than falling back into legacy switch branches. Test that empty and whitespace-only strings resolve to unassigned rather than orchestrator. Test custom mapping lookups.
- backend/internal/domain/enums_test.go:
  - Update TestOrderSubagentRoles_OrderingAndPreservation: Verify default alphabetical sorting when no custom order is provided. Verify custom priority sorting when custom order slices are passed.
- backend/internal/domain/aggregations_test.go:
  - Update TestOrderRoles, TestOrderSubagentRoles, and TestOrderRoles_And_OrderSubagentRoles_AdversarialEdgeCases to verify default alphabetical ordering and custom ordering.
  - Update TestComputeRoleMetrics_DynamicOrdering and TestComputeThreadSummaries_DynamicOrderingAndTieBreaking to assert alphabetical ordering by default or custom ordering when provided.
  - Update peak context subagent tie-breaking tests in TestWorkflowEfficiency_DynamicSubagents and TestComputeWorkflowEfficiency_ComprehensiveTieBreaking to test alphabetical tie-breaking by default, and custom priority tie-breaking when custom order is supplied.

### Boundary Adapters and Configuration (backend/internal/adapters/ & .agenticlens/)
- .agenticlens/config.yaml:
  - Add role_order list under configuration defining the display and priority order (orchestrator, planner, coder, tester, reviewer).
- backend/internal/adapters/config/loader.go:
  - Add RoleOrder slice of strings to rawProjectConfigDTO.
  - In mergeConfig: If dto.RoleOrder is non-nil, sanitize and assign to cfg.RoleOrder.
- backend/internal/adapters/config/loader_test.go:
  - Add test case verifying that role_order is parsed accurately from YAML and JSON project configuration files into ProjectConfig.RoleOrder.
  - Verify that when role_order is omitted from YAML, DefaultProjectConfig baseline role order is preserved.
- backend/internal/adapters/git/inference.go:
  - Delete InferTicketTitle and cleanHeadingTitle completely.
  - Delete findWorkspaceFile helper which was exclusively used by InferTicketTitle.
  - Preserve ComputeSDLCIterations, GetGitUser, and InferTicketType.
- backend/internal/adapters/git/analyzer_test.go:
  - In TestInferenceOnActiveWorkspace: Remove InferTicketTitle invocation and assertions.
  - In TestGitAnalyzer_AdversarialEdgeCases: Remove InferTicketTitle empty string assertion.
- backend/internal/adapters/repository.go:
  - In ScanWorkspaces and ScanSingleWorkspace: Update calls from domain.DeriveMetricsLedger(factLedger, cfg.RoleMappings) to domain.DeriveMetricsLedgerWithConfig(factLedger, cfg) so that both RoleMappings and RoleOrder are applied during workspace scans.
- backend/internal/adapters/formatters.go:
  - In FormatMarkdownMetricsTable: Accept optional role order slice or retrieve from ledger if attached, passing to domain.OrderSubagentRoles for subagent effort distribution and turn velocity rendering. Default to alphabetical sorting when unsupplied.
- backend/internal/adapters/formatters_test.go:
  - Update TestFormatMarkdownMetricsTable_MixedCanonicalAndCustomSubagentsPrecedence and related tests to verify alphabetical formatting by default, or provide custom order when testing priority formatting.
- backend/internal/adapters/parity_test.go:
  - Verify all 12 reference ticket ledgers pass 100% parity evaluation.

### Command Line Interface (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go:
  - In generate command handler: Remove the fallback invocation of git.InferTicketTitle(activeTicketID). Set activeTicketTitle strictly if the --ticket-title flag is non-empty, falling back to cachedFacts.TicketTitle, otherwise leaving activeTicketTitle as nil.
  - Update ledger derivation call to domain.DeriveMetricsLedgerWithConfig(factLedger, projCfg).
  - Update FormatMarkdownMetricsTable invocation to pass projCfg.RoleOrder.

### Presentation Layer (web/src/)
- web/src/domain/types.ts:
  - Retain AgentRole as string. Ensure that the web UI continues to consume backend-provided role arrays without imposing hardcoded client-side sorting filters.
- web/src/utils/theme.ts:
  - Confirm getRoleColor safely falls back to dynamic palette hashing via hashStringToColor for arbitrary custom roles (such as developer, architect, researcher).
- web/src/views/TicketDeepDive.tsx & FleetOverview.tsx:
  - Verify autopsy breakdown tables and turn velocity lists render in the exact sequence delivered by the Go backend.

---

## 3. Phased Implementation Tasks

### Phase 1: De-Hardcode Role Resolution in Domain Core (backend/internal/domain/config.go & enums.go)
- Task 1.1: Define constant RoleUnassigned as AgentRole = "unassigned" in backend/internal/domain/enums.go.
- Task 1.2: In backend/internal/domain/config.go, refactor RoleMappings.ResolveRole:
  - Clean input rawAgent by trimming whitespace and converting to lowercase.
  - If the cleaned string is empty, return RoleUnassigned immediately.
  - If the mapping table is non-nil:
    - Check exact match on cleaned string. If present and non-empty, return AgentRole of trimmed lowercase value.
    - Check exact match on raw input string. If present and non-empty, return AgentRole of trimmed lowercase value.
    - Check case-insensitive match across keys. If match found and non-empty, return AgentRole of trimmed lowercase value.
  - If no mapping matches, return AgentRole(cleaned) directly.
  - Delete the entire switch cleaned statement containing orchestrator, build, root, planner, coder, tester, reviewer.
- Task 1.3: In backend/internal/domain/config_test.go:
  - Update existing assertions where empty string or whitespace previously returned RoleOrchestrator to now assert RoleUnassigned.
  - Update unmapped persona test cases to assert that any unknown persona (such as developer, qa-specialist, or researcher) returns its verbatim cleaned name.
  - Verify mapped persona test cases continue to resolve configured aliases (such as architect to planner).

### Phase 2: Configurable Role Ordering in Domain Core (backend/internal/domain/)
- Task 2.1: In backend/internal/domain/config.go:
  - Add RoleOrder ([]string) field to ProjectConfig struct with json:"role_order" yaml:"role_order" tags.
  - In DefaultProjectConfig, set RoleOrder to a slice containing orchestrator, planner, coder, tester, reviewer.
- Task 2.2: In backend/internal/domain/enums.go:
  - Refactor OrderRoles(roles []string, customOrder ...[]string) []string:
    - Deduplicate non-empty trimmed strings from roles. If none exist, return nil.
    - If customOrder is provided, len(customOrder) > 0, and len(customOrder[0]) > 0:
      - Iterate through customOrder[0], and for each role present in the input set, append to the ordered slice and remove from the set.
      - Collect remaining unlisted roles from the input set, sort them in ascending alphabetical order, and append them to the ordered slice.
      - Return the combined slice.
    - If customOrder is not provided or empty:
      - Collect all deduplicated input roles, sort them in ascending alphabetical order, and return.
  - Refactor OrderSubagentRoles(roles []string, customOrder ...[]string) []string:
    - Implement identical logic: if customOrder is supplied, match against the custom order sequence first, then append remaining unlisted roles sorted alphabetically; if customOrder is absent or empty, sort all deduplicated input roles alphabetically.
- Task 2.3: In backend/internal/domain/aggregations.go:
  - Update ComputeWorkflowEfficiency signature to accept optional variadic roleOrder ...[]string. Pass roleOrder... to OrderSubagentRoles when calculating candidate subagent role ordering and turn velocity role ordering.
  - Update ComputeRoleMetrics signature to accept optional variadic roleOrder ...[]string. Pass roleOrder... to OrderRoles.
  - Update ComputeThreadSummaries signature to accept optional variadic roleOrder ...[]string. Pass roleOrder... to OrderRoles.
  - Introduce DeriveMetricsLedgerWithConfig(factLedger AgenticFactLedger, cfg ProjectConfig) AgenticMetricsLedger:
    - Map invocations and events using cfg.RoleMappings.
    - Pass cfg.RoleOrder into ComputeWorkflowEfficiency, ComputeRoleMetrics, and ComputeThreadSummaries.
  - Update DeriveMetricsLedger(factLedger AgenticFactLedger, roleMappings ...RoleMappings) AgenticMetricsLedger:
    - Construct baseline ProjectConfig using DefaultProjectConfig(), override RoleMappings if roleMappings parameter is provided, and delegate to DeriveMetricsLedgerWithConfig.
- Task 2.4: In backend/internal/domain/enums_test.go:
  - Refactor TestOrderSubagentRoles_OrderingAndPreservation:
    - Test default behavior without custom order: assert input slices sort alphabetically (for example, reviewer, tester, coder, planner sorts to coder, planner, reviewer, tester).
    - Test custom order behavior: pass explicit custom order slices and assert that listed roles appear in the requested order followed by unlisted roles in alphabetical order.
- Task 2.5: In backend/internal/domain/aggregations_test.go:
  - Update TestOrderRoles and TestOrderSubagentRoles to test both alphabetical default sorting and custom order priority.
  - Update TestComputeRoleMetrics_DynamicOrdering and TestComputeThreadSummaries_DynamicOrderingAndTieBreaking to assert alphabetical sorting by default, or pass custom order where priority ordering is explicitly validated.
  - In TestWorkflowEfficiency_DynamicSubagents and TestComputeWorkflowEfficiency_ComprehensiveTieBreaking: Update tie-breaker assertions so that when no custom order is passed, alphabetical order breaks ties (for example, advisor beats coder); and when a custom order slice is supplied, the custom order breaks ties.

### Phase 3: Configuration Loading and Config YAML Synchronization
- Task 3.1: In .agenticlens/config.yaml:
  - Add role_order section with the standard repository role sequence (orchestrator, planner, coder, tester, reviewer).
- Task 3.2: In backend/internal/adapters/config/loader.go:
  - Add RoleOrder ([]string) to rawProjectConfigDTO.
  - In mergeConfig: If dto.RoleOrder is non-nil, filter empty strings, trim whitespace, and assign to cfg.RoleOrder.
- Task 3.3: In backend/internal/adapters/config/loader_test.go:
  - Add unit test verifying that loading a YAML file containing role_order correctly populates cfg.RoleOrder in domain.ProjectConfig.
  - Verify that when role_order is omitted from YAML, DefaultProjectConfig baseline role order is preserved.

### Phase 4: Elimination of Fragile Ticket Title Inference
- Task 4.1: In backend/internal/adapters/git/inference.go:
  - Delete InferTicketTitle(ticketID string) *string.
  - Delete cleanHeadingTitle(line, ticketID, docType string) string.
  - Delete findWorkspaceFile(ticketID, filename string) ([]byte, bool).
  - Verify that ComputeSDLCIterations and InferTicketType remain intact.
- Task 4.2: In backend/cmd/agentic-lens/main.go:
  - In the generate command flag parsing section: Remove the else fallback branch that assigned activeTicketTitle = git.InferTicketTitle(activeTicketID).
  - Ensure activeTicketTitle is populated only if --ticket-title flag is non-empty, or if cachedFacts contains a non-nil TicketTitle; otherwise activeTicketTitle remains nil.
- Task 4.3: In backend/internal/adapters/git/analyzer_test.go:
  - In TestInferenceOnActiveWorkspace: Remove lines invoking git.InferTicketTitle("AL-11") and asserting title content.
  - In TestGitAnalyzer_AdversarialEdgeCases: Remove the assertion verifying git.InferTicketTitle("").

### Phase 5: Boundary Adapters and CLI Integration
- Task 5.1: In backend/internal/adapters/repository.go:
  - In ScanWorkspaces: Change the ledger derivation call from domain.DeriveMetricsLedger(factLedger, cfg.RoleMappings) to domain.DeriveMetricsLedgerWithConfig(factLedger, cfg).
  - In ScanSingleWorkspace: Change the ledger derivation call to domain.DeriveMetricsLedgerWithConfig(factLedger, cfg).
- Task 5.2: In backend/internal/adapters/formatters.go:
  - In FormatMarkdownMetricsTable: Accept optional role order slice variadic parameter roleOrder ...[]string. Pass roleOrder... into domain.OrderSubagentRoles when formatting Role Effort Distribution and Subagent Turn Velocity.
- Task 5.3: In backend/cmd/agentic-lens/main.go:
  - Update metrics derivation to domain.DeriveMetricsLedgerWithConfig(factLedger, projCfg).
  - Update FormatMarkdownMetricsTable call to pass projCfg.RoleOrder.
- Task 5.4: In backend/internal/adapters/formatters_test.go:
  - Update TestFormatMarkdownMetricsTable_MixedCanonicalAndCustomSubagentsPrecedence: Pass custom role order to test custom sequence rendering, or assert alphabetical rendering when no role order is passed.
  - Verify all other formatter tests pass cleanly.

### Phase 6: Comprehensive Verification and Quality Gates
- Task 6.1: Run go vet ./... across all Go packages to confirm zero static analysis issues.
- Task 6.2: Run go test -race -count=1 ./... across all Go backend packages, including parity_test.go, config_test.go, enums_test.go, aggregations_test.go, analyzer_test.go, and formatters_test.go.
- Task 6.3: In web/, execute npm run typecheck and npm test to confirm TypeScript contracts and component test suites remain green.
- Task 6.4: Execute make build from repository root to confirm single binary compilation succeeds.
- Task 6.5: Audit git diff to verify zero emojis and zero forbidden code patterns exist across all modified code, documentation, and configuration files.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (De-Hardcode Role Resolution) removes switch statement dependencies in config.go and establishes RoleUnassigned. Unblocks Phase 2.
2. Phase 2 (Configurable Role Ordering) refactors enums.go, config.go, and aggregations.go to sort alphabetically by default and respect custom order when supplied. Unblocks Phase 3 and Phase 5.
3. Phase 3 (Configuration Loading) adds role_order to config.yaml and the loader adapter, connecting user configuration to domain config structs.
4. Phase 4 (Title Inference Elimination) removes InferTicketTitle from inference.go, main.go, and analyzer_test.go.
5. Phase 5 (Boundary Adapters and CLI Integration) wires DeriveMetricsLedgerWithConfig into repository.go and main.go, and updates formatters.go.
6. Phase 6 (Quality Gates) runs comprehensive unit tests, race detection, parity tests, frontend typechecks, and single binary build.
