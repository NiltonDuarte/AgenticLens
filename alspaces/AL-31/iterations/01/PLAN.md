# Implementation Plan: AL-31 Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests

## 1. Goal and Iteration Scope

Deliver the complete decoupling of structural session hierarchy from agent persona roles across the Go domain core, boundary adapters, formatters, and test suites. Ensure that root invocations exclusively govern root coordination metrics, child invocations exclusively govern subagent metrics, single-session tickets report zero subagent tokens with none as the role, and all parity tests in parity_test.go pass against authentic workspace ledgers. Deliver this vertical slice in iteration 01.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation directly modifies and verifies the following existing packages and modules:

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/aggregations.go:
  - Modify ComputeWorkflowEfficiency to partition invocations by ParentID == nil (root coordination) versus ParentID != nil (delegated subagents).
  - Compute root coordination peak tokens and saturation percentage from root invocations.
  - Compute subagent peak tokens, peak role, costs, effort distribution, and turn velocity strictly from child invocations.
  - Return zero subagent peak tokens, none subagent peak role, and empty effort distribution when child invocations are absent.
- backend/internal/domain/enums.go:
  - Modify IsSubagentRole to return false for root personas (orchestrator, general, build, root, and whitespace strings) and true for worker subagent personas.
  - Modify OrderSubagentRoles to exclude root personas.
- backend/internal/domain/anomalies.go:
  - Update detectSubagentReworkLoops to use domain.IsSubagentRole rather than the ad-hoc local check.
- backend/internal/domain/aggregations_test.go:
  - Add comprehensive table-driven tests for ComputeWorkflowEfficiency covering diverse root personas, single-invocation sessions, and multi-tier subagent trees.
  - Expand TestIsSubagentRole and TestOrderSubagentRoles test suites.

### Go Imperative Boundary Shell (backend/internal/adapters/)
- backend/internal/adapters/formatters.go:
  - Update FormatMarkdownMetricsTable to format RoleEffortDistributionPct directly and handle none subagent role and zero tokens cleanly.
- backend/internal/adapters/parity_test.go:
  - Update reference baseline assertions for tickets AL-1, AL-10, AL-11, AL-12, GEN-1, GEN-2, PO-1, PO-2, and PO-3 to reflect authentic general root sessions in PeakByRole and correct subagent attribution.
  - Update expected fleet role costs in TestFleetSummaryAndComparisonParity to match authentic split between orchestrator and general personas.

### Quality, Verification, and Build Targets
- Makefile: check, test, build targets.
- Single binary entrypoint: backend/cmd/agentic-lens/main.go.

---

## 3. Phased Implementation Tasks

### Phase 1: Domain Core Enums Refactoring (backend/internal/domain/enums.go)
- Task 1.1: Audit all occurrences of RoleOrchestrator, IsSubagentRole, and OrderSubagentRoles in backend/internal/domain/enums.go.
- Task 1.2: Refactor IsSubagentRole:
  - Trim whitespace and convert the input string to lowercase.
  - Return false if the resulting string is empty.
  - Return false if the string matches any recognized root/coordination persona: orchestrator, general, build, or root.
  - Return true for all other personas (canonical subagents planner, coder, tester, reviewer, and custom worker roles such as advisor, researcher, po, or security-auditor).
- Task 1.3: Refactor OrderSubagentRoles:
  - Filter input candidate roles through IsSubagentRole so that root coordination personas are discarded.
  - Retain canonical subagent ordering precedence (planner, coder, tester, reviewer).
  - Sort remaining custom subagent roles alphabetically.
- Task 1.4: Update detectSubagentReworkLoops in backend/internal/domain/anomalies.go to call domain.IsSubagentRole.

### Phase 2: Domain Aggregations Hierarchy Decoupling (backend/internal/domain/aggregations.go)
- Task 2.1: In ComputeWorkflowEfficiency, introduce explicit tracking variables:
  - Root coordination tracking: rootPeak of type int64, hasRoot of type bool.
  - Subagent tracking: subCosts map from string to float64, totSubCost of type float64, subPeaks map from string to int64, roleTurns map from string to int64 slice.
- Task 2.2: Refactor the primary invocation loop over ledger.Invocations:
  - Accumulate active duration, total reasoning tokens, total output tokens, and compactions across all invocations.
  - Record peak context tokens in the overall rolePeaks map (used for ContextWindowHealth.PeakByRole).
  - Evaluate structural condition: if inv.ParentID == nil:
    - Update rootPeak with the maximum of rootPeak and inv.Tokens.PeakContext, setting hasRoot to true.
  - Otherwise (inv.ParentID != nil):
    - Accumulate child invocation cost into subCosts[inv.Role] and totSubCost.
    - Update subPeaks[inv.Role] with the maximum of subPeaks[inv.Role] and inv.Tokens.PeakContext.
    - Append inv.Turns to roleTurns[inv.Role].
- Task 2.3: Derive root coordination metrics:
  - Assign orchPeak to rootPeak if hasRoot is true, otherwise zero.
  - Calculate orchSatPct as the rounded percentage of orchPeak against HealthyCeilingTokens, defended against zero or negative ceiling values.
- Task 2.4: Derive subagent peak metrics:
  - If subPeaks is empty (no child invocations exist):
    - Set subPeak to zero.
    - Set subRoleStr to the string literal none.
  - If subPeaks is non-empty:
    - Collect all keys from subPeaks into candidateSubRoles.
    - Iterate over OrderSubagentRoles(candidateSubRoles) to find the role exhibiting the maximum peak context tokens, with tie-breaking governed by canonical role order.
    - Set subPeak to the maximum value and subRoleStr to the winning role.
- Task 2.5: Derive subagent effort distribution:
  - Initialize roleEffortDistributionPct as an empty map.
  - If totSubCost > 0 and subCosts is non-empty:
    - Calculate the percentage of totSubCost for each role in subCosts, rounding to one decimal place, only adding entries with positive percentage.
- Task 2.6: Derive subagent turn velocity:
  - Aggregate allSubTurns and calculate average and peak turns across roleTurns.
  - If roleTurns is empty, set average turns to 0.0, peak turns to 0, and by-role velocity to an empty map.
- Task 2.7: Construct ContextWindowHealth:
  - Compute overall peak as the maximum of orchPeak and subPeak.
  - Compute ContextWindowHealth.Status using GetBaseHealthStatus with overall peak.
  - Populate ContextWindowHealth with OrchestratorPeakTokens = orchPeak, OrchestratorSaturationPct = orchSatPct, SubagentsPeakTokens = subPeak, SubagentsPeakRole = subRoleStr, and PeakByRole = rolePeaks.

### Phase 3: Boundary Adapters and Markdown Formatter Synchronization (backend/internal/adapters/)
- Task 3.1: In backend/internal/adapters/formatters.go:
  - Update FormatMarkdownMetricsTable subagent effort distribution generation:
    - Collect keys from we.RoleEffortDistributionPct.
    - Order keys with domain.OrderSubagentRoles.
    - Format each subagent percentage without hardcoded role != RoleOrchestrator checks (since root sessions are structurally excluded).
    - If no subagent entries exist, maintain the display string N/A.
  - Verify that when cwh.SubagentsPeakTokens == 0 and cwh.SubagentsPeakRole == "none", the context window health bullet formats cleanly as Subagents peak 0 (None).
- Task 3.2: In backend/internal/adapters/opencode/extractor.go:
  - Audit ExtractFacts to verify that raw agent persona strings from database rows are assigned to InvocationFact.Role without synthetic remapping, and parent_id pointers are preserved faithfully.
- Task 3.3: In backend/internal/adapters/schemas.go:
  - Audit InvocationFactDTO and AgenticFactLedgerDTO to verify roundtrip fidelity of ParentID and Role.

### Phase 4: Domain Core Unit Testing (backend/internal/domain/aggregations_test.go & enums_test.go)
- Task 4.1: Expand TestIsSubagentRole in backend/internal/domain/aggregations_test.go:
  - Assert that general, General, build, Build, root, Root, orchestrator, and Orchestrator return false.
  - Assert that canonical subagents (planner, coder, tester, reviewer) return true.
  - Assert that custom subagents (advisor, researcher, po, custom-agent) return true.
- Task 4.2: Expand TestOrderSubagentRoles:
  - Verify that passing a list containing root personas (such as general, orchestrator, build) along with subagent roles filters out the root personas and orders the subagents properly.
- Task 4.3: Add table-driven tests for ComputeWorkflowEfficiency in backend/internal/domain/aggregations_test.go:
  - Test Case A: Single root invocation with role general (such as GEN-1 pattern). Assert OrchestratorPeakTokens matches root peak, OrchestratorSaturationPct matches calculation, SubagentsPeakTokens == 0, SubagentsPeakRole == "none", RoleEffortDistributionPct is empty map, and SubagentTurnVelocity has zero turns.
  - Test Case B: Single root invocation with role build or custom root persona. Assert root coordination peak populated, subagents zeroed.
  - Test Case C: Root session with role general and child subagents (planner, coder, tester, reviewer) with ParentID set to the root session ID (such as AL-1 pattern). Assert root tokens do not appear in SubagentsPeakTokens, SubagentsPeakRole matches highest child, RoleEffortDistributionPct contains only child roles summing to approximately 100%, and OrchestratorPeakTokens matches the general root session peak.
  - Test Case D: Empty ledger with zero invocations. Assert zero division defenses hold, SubagentsPeakRole == "none", and zero panic.

### Phase 5: Parity Test Baseline Alignment (backend/internal/adapters/parity_test.go)
- Task 5.1: In backend/internal/adapters/parity_test.go, update reference JSON fixtures for all tickets where authentic root sessions have persona general:
  - AL-1: Update PeakByRole to reflect authentic persona general: 123350 instead of orchestrator: 123350. Verify OrchestratorPeakTokens = 123350, SubagentsPeakTokens = 85656, and SubagentsPeakRole = "coder".
  - AL-10: Update PeakByRole to reflect authentic persona general instead of orchestrator.
  - AL-11: Update PeakByRole to reflect authentic persona general instead of orchestrator.
  - AL-12: Update PeakByRole to reflect authentic persona general instead of orchestrator.
  - GEN-1: Update PeakByRole to reflect authentic persona general: 26252. Verify SubagentsPeakTokens = 0, SubagentsPeakRole = "none", and RoleEffortDistributionPct = {}.
  - GEN-2: Update PeakByRole to reflect authentic persona general: 29125. Verify SubagentsPeakTokens = 0, SubagentsPeakRole = "none", and RoleEffortDistributionPct = {}.
  - PO-1: Update PeakByRole to reflect authentic persona general: 113824. Verify SubagentsPeakTokens = 0, SubagentsPeakRole = "none", and RoleEffortDistributionPct = {}.
  - PO-2: Update PeakByRole to reflect authentic persona general: 88541. Verify SubagentsPeakTokens = 0, SubagentsPeakRole = "none", and RoleEffortDistributionPct = {}.
  - PO-3: Update PeakByRole to reflect authentic persona general: 86075. Verify SubagentsPeakTokens = 0, SubagentsPeakRole = "none", and RoleEffortDistributionPct = {}.
- Task 5.2: In TestFleetSummaryAndComparisonParity:
  - Update expectedRoleCosts map to split root session spend authentically between orchestrator (for tickets AL-2, AL-3, AL-4) and general (for tickets AL-1, AL-10, AL-11, AL-12, GEN-1, GEN-2, PO-1, PO-2, PO-3).
- Task 5.3: Run TestCrossEngineParityOnWorkspaceLedgers and TestFleetSummaryAndComparisonParity to ensure all 12 tickets pass parity evaluation with zero failures.

### Phase 6: Automated Quality Gates and Single Binary Build
- Task 6.1: Run go vet ./... to verify static analysis compliance.
- Task 6.2: Run go test -race ./... across all backend packages.
- Task 6.3: Run frontend type checks (npm run typecheck) and tests (npm test) in web/ to verify no regressions.
- Task 6.4: Execute make build from repository root to confirm single binary compilation.
- Task 6.5: Verify zero emojis exist across all code, tests, and documentation.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Domain Enums Refactoring) -> establishes pure classification helpers and unblocks Phase 2.
2. Phase 2 (Domain Aggregations Decoupling) -> establishes structural hierarchy partitioning in functional core and unblocks Phase 3, Phase 4, and Phase 5.
3. Phase 3 (Boundary Adapters and Formatters) -> synchronizes presentation formatting with the decoupled domain model.
4. Phase 4 (Domain Unit Tests) -> verifies behavioral invariants and edge cases in pure isolation.
5. Phase 5 (Parity Test Baseline Alignment) -> resolves the 8 failing parity test cases and verifies cross-engine parity.
6. Phase 6 (Quality Gates & Build) -> confirms static typing, race safety, zero any, zero emojis, and single binary build.
