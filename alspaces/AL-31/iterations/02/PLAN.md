# Implementation Plan: AL-31 Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests (Iteration 02)

## 1. Goal and Iteration Scope

Deliver Iteration 02 of ticket AL-31 to implement pure structural session classification across the Go domain engine and anomaly detection rules. In Iteration 01, root coordination metrics were decoupled from fixed persona names, but delegated subagent classification in aggregations, role ordering, and rework loop detection still relied on string blacklist filtering (such as checking whether a role string matched orchestrator, general, build, or root).

In Iteration 02, the domain engine transitions to pure structural classification:
- In `backend/internal/domain/aggregations.go` (`ComputeWorkflowEfficiency`), an invocation is classified exclusively by its structural relationship in the session tree: `inv.ParentID == nil` strictly designates a root coordination session, while `inv.ParentID != nil` strictly designates a delegated child subagent. The string blacklist check is eliminated.
- In `backend/internal/domain/enums.go`, `OrderSubagentRoles` is simplified to accept candidate role strings extracted from child invocations, ordering canonical subagents (`planner`, `coder`, `tester`, `reviewer`) first and sorting remaining custom roles alphabetically without filtering by a persona blacklist. `IsSubagentRole` is explicitly deprecated and documented with guidance to use structural hierarchy (`ParentID != nil`).
- In `backend/internal/domain/anomalies.go` (`detectSubagentReworkLoops`), subagent event transitions are resolved structurally by mapping event session identifiers (`EventFact.SessionID`) to invocation hierarchy (`InvocationFact.ParentID != nil`), ensuring root coordination turns are excluded structurally rather than via string blacklists.
- All tests in `aggregations_test.go`, `enums_test.go`, `anomalies_test.go`, and `parity_test.go` are aligned to verify pure structural classification.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation modifies and verifies the following existing domain packages and adapters:

### Go Functional Core (backend/internal/domain/)
- `backend/internal/domain/aggregations.go`:
  - In `ComputeWorkflowEfficiency`, eliminate `IsSubagentRole(inv.Role)` on child invocation processing. Classify every invocation strictly by whether `inv.ParentID == nil` (root coordination session) or `inv.ParentID != nil` (delegated child subagent).
  - Accumulate subagent costs, peak context tokens, and turns strictly for invocations where `inv.ParentID != nil`.
  - Maintain root coordination peak context tokens and context saturation percentage strictly for invocations where `inv.ParentID == nil`.
- `backend/internal/domain/enums.go`:
  - In `OrderSubagentRoles`, remove the dependency on `IsSubagentRole`. Deduplicate non-empty trimmed candidate role strings, order canonical subagents by `SubagentRoleOrder`, and sort custom roles alphabetically.
  - In `IsSubagentRole`, add a deprecation doc comment indicating that session classification is strictly structural (`ParentID != nil`), and retain non-empty, non-orchestrator backwards compatibility for callers lacking session hierarchy context.
- `backend/internal/domain/anomalies.go`:
  - In `detectSubagentReworkLoops`, construct a session hierarchy map from `ledger.Invocations` indexing session IDs to a boolean indicating whether `inv.ParentID != nil`.
  - Track subagent turns from `ledger.Events` by checking if the event session ID maps to a child subagent session (`ParentID != nil`). Root session events are structurally ignored.
  - When falling back to `ledger.Invocations`, include invocations where `inv.ParentID != nil`.
  - Provide a fallback to role heuristics only when no hierarchy information is present (such as in legacy or synthetic test fixtures where all invocations have nil parent pointers and empty session IDs).
- `backend/internal/domain/aggregations_test.go`:
  - Update `TestComputeWorkflowEfficiency_HierarchyDecoupling` Case H to verify that child invocations with `ParentID != nil` are classified as subagents regardless of their persona role string.
  - Expand table-driven test cases asserting that pure structural hierarchy determines root versus subagent attribution.
  - Update `TestOrderSubagentRoles` to assert that candidate roles extracted from child invocations are preserved, ordered by canonical priority, and sorted alphabetically without dropping non-canonical strings.
- `backend/internal/domain/anomalies_test.go`:
  - Add test scenarios for `detectSubagentReworkLoops` verifying that root session turns (`ParentID == nil`) interleaved with subagent turns do not disrupt ping-pong cycle detection between child subagents.

### Boundary Adapters and Parity Tests (backend/internal/adapters/)
- `backend/internal/adapters/formatters.go`:
  - Confirm `FormatMarkdownMetricsTable` cleanly consumes the structurally derived `RoleEffortDistributionPct` and `SubagentTurnVelocity.ByRole`.
- `backend/internal/adapters/parity_test.go`:
  - Verify that all 12 reference workspace ledgers pass 100% without modification, confirming zero regressions against authentic multi-agent and single-session data.

---

## 3. Phased Implementation Tasks

### Phase 1: Pure Structural Hierarchy in Domain Aggregations (backend/internal/domain/aggregations.go)
- Task 1.1: In `ComputeWorkflowEfficiency`, inspect the primary iteration loop over `ledger.Invocations`.
- Task 1.2: Refactor the invocation classification branch:
  - If `inv.ParentID == nil`:
    - Record invocation as root coordination.
    - If `!hasRoot` or `inv.Tokens.PeakContext > rootPeak`, update `rootPeak` to `inv.Tokens.PeakContext`.
    - Set `hasRoot` to true.
  - Else (`inv.ParentID != nil`):
    - Treat invocation strictly as a delegated child subagent without checking `IsSubagentRole(inv.Role)`.
    - Normalize role string: trim leading and trailing whitespace. If empty, fall back to the literal string unassigned.
    - Accumulate child invocation cost: add `inv.CostUSD` to `subCosts[role]` and `totSubCost`.
    - Update peak context tokens: if not recorded or `inv.Tokens.PeakContext > subPeaks[role]`, set `subPeaks[role] = inv.Tokens.PeakContext`.
    - Collect execution turns: append `inv.Turns` to `roleTurns[role]`.
- Task 1.3: Confirm root metrics derivation remains intact:
  - `orchPeak` is `rootPeak` if `hasRoot` is true, otherwise 0.
  - `orchSatPct` is calculated against `HealthyCeilingTokens`, defended against zero or negative ceiling values.
- Task 1.4: Confirm subagent metrics derivation remains intact:
  - If `len(subPeaks) == 0`: `subPeak` is 0, `subRoleStr` is none, `roleEffortDistributionPct` is an empty map, and `subagentTurnVelocity` has zero average and peak turns with an empty map.
  - If `len(subPeaks) > 0`: candidate roles from `subPeaks` are ordered via `OrderSubagentRoles`, highest peak context tokens determines `subPeak`, tie-breaking follows canonical order, and effort distribution percentages and turn velocities are calculated exclusively from child invocations.

### Phase 2: Role Ordering Refactoring and Deprecation in Domain Enums (backend/internal/domain/enums.go)
- Task 2.1: Refactor `OrderSubagentRoles`:
  - Remove invocation of `IsSubagentRole`.
  - Deduplicate input role strings after trimming whitespace, ignoring empty strings.
  - Iterate through `SubagentRoleOrder` (`planner`, `coder`, `tester`, `reviewer`), appending matching roles present in the input.
  - Collect remaining custom roles not in `SubagentRoleOrder`, sort them in ascending alphabetical order, and append them to the ordered slice.
  - Return the resulting slice.
- Task 2.2: Refactor and deprecate `IsSubagentRole`:
  - Add standard Go deprecation doc comment explaining that session classification is strictly determined by structural hierarchy (`ParentID != nil`) rather than persona string heuristics.
  - For callers lacking session hierarchy context, maintain backwards-compatible behavior: trim whitespace, return false for empty string or `orchestrator`, and return true for canonical and custom subagent personas.

### Phase 3: Structural Subagent Resolution in Anomaly Detection (backend/internal/domain/anomalies.go)
- Task 3.1: In `detectSubagentReworkLoops`, construct session hierarchy lookup maps from `ledger.Invocations`:
  - Build `isChildSession` mapping `inv.ID` to boolean `(inv.ParentID != nil)`.
  - Determine `hasHierarchy`: true if at least one invocation in `ledger.Invocations` has `inv.ParentID != nil`.
- Task 3.2: Refactor event processing (`ledger.Events`):
  - When `hasHierarchy` is true, check whether `isChildSession[e.SessionID]` is true. If true, extract `e.Role`, trim and lowercase, and append to the role sequence if distinct from the previous role. If false (root coordination turn), ignore the event.
  - When `hasHierarchy` is false (such as synthetic unit tests where invocations are omitted or session IDs are blank), fall back to `IsSubagentRole(e.Role)` to preserve test suite compatibility.
- Task 3.3: Refactor invocation fallback processing (`ledger.Invocations` when events produce no roles):
  - When `hasHierarchy` is true, check `inv.ParentID != nil`. If true, extract `inv.Role`, trim and lowercase, and append to the role sequence if distinct from the previous role. If `inv.ParentID == nil`, ignore the invocation.
  - When `hasHierarchy` is false, fall back to `IsSubagentRole(inv.Role)`.
- Task 3.4: Preserve existing cycle detection logic (alternating pairs for >= 3 cycles triggers WARNING, >= 5 cycles triggers CRITICAL).

### Phase 4: Test Suite Synchronization and Edge Case Coverage
- Task 4.1: In `backend/internal/domain/aggregations_test.go`:
  - Update `TestComputeWorkflowEfficiency_HierarchyDecoupling` Case H: modify the child invocation test case to verify that child invocations with `ParentID != nil` are included in subagent metrics regardless of their persona string (such as unconventional names).
  - Add a dedicated test verifying that single-invocation tickets with custom persona names have `SubagentsPeakTokens == 0`, `SubagentsPeakRole == "none"`, and `RoleEffortDistributionPct == {}`.
  - Add a dedicated test verifying that multi-tier child invocations (`root -> subagent1 -> subagent2`) are all treated as delegated subagents because their parent IDs are non-nil.
- Task 4.2: In `backend/internal/domain/enums_test.go` and `aggregations_test.go`:
  - Update `TestOrderSubagentRoles` to assert that all non-empty roles passed in are retained, canonical roles appear first, and custom roles appear in alphabetical order.
  - Verify `TestIsSubagentRole` continues to satisfy its contract for backwards compatibility.
- Task 4.3: In `backend/internal/domain/anomalies_test.go`:
  - Add test case verifying that in a ledger with authentic hierarchy (`Invocations` containing root with `ParentID == nil` and subagents with `ParentID != nil`), events from the root session are structurally ignored and do not interrupt rework loop detection between subagents.
  - Verify all existing synthetic unit tests pass without regressions.
- Task 4.4: In `backend/internal/adapters/parity_test.go`:
  - Run `TestCrossEngineParityOnWorkspaceLedgers` across all 12 reference tickets.
  - Run `TestFleetSummaryAndComparisonParity`.
  - Confirm 100% pass rate.

### Phase 5: Verification, Quality Gates, and Single Binary Build
- Task 5.1: Run `go vet ./...` across all Go packages to confirm static analysis compliance.
- Task 5.2: Run race detection test suite via `go test -race ./...`.
- Task 5.3: Run frontend type checks (`npm run typecheck` in `web/`) to confirm TypeScript contract alignment.
- Task 5.4: Execute `make build` to verify clean single binary compilation.
- Task 5.5: Verify zero emojis in code, docstrings, comments, test assertions, and workspace artifacts.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Domain Aggregations Structural Hierarchy) establishes pure `ParentID` partitioning in `ComputeWorkflowEfficiency`.
2. Phase 2 (Domain Enums Role Ordering & Deprecation) decouples `OrderSubagentRoles` from string blacklists and deprecates `IsSubagentRole`.
3. Phase 3 (Anomaly Detection Structural Mapping) updates `detectSubagentReworkLoops` to map session IDs to invocation hierarchy.
4. Phase 4 (Test Suite Synchronization) updates unit test fixtures in `aggregations_test.go`, `enums_test.go`, and `anomalies_test.go`, verifying parity across reference ledgers.
5. Phase 5 (Verification & Quality Gates) executes `go vet`, race tests, frontend checks, and single binary build.
