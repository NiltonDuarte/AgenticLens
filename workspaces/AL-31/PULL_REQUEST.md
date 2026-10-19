# refactor(domain): decouple root session structural hierarchy from persona roles (AL-31)

## Summary

Decouple structural session hierarchy (`ParentID == nil` vs `ParentID != nil`) from agent persona role names across the AgenticLens Go domain engine, boundary adapters, formatters, and parity tests. Pure structural classification replaces persona string blacklists.

### Background

Root sessions initiated directly by users in OpenCode are assigned the persona `"general"`. Previously, the domain engine assumed the root session persona was literally `"orchestrator"`. Because `"general" != "orchestrator"`, root sessions were misclassified as delegated subagents, causing root coordination peak context tokens to drop to 0, root session peak tokens to leak into `SubagentsPeakTokens`, subagent effort distribution percentages and turn velocity to be diluted, and eight test cases in `backend/internal/adapters/parity_test.go` to fail against authentic workspace ledgers.

Following steering feedback in Iteration 02, persona string blacklists (`""`, `"orchestrator"`, `"general"`, `"build"`, `"root"`) were completely eliminated in favor of pure structural hierarchy via session parent pointers.

### Key Changes

1. **Domain Core (`backend/internal/domain/`)**:
   - `aggregations.go`: In `ComputeWorkflowEfficiency`, partition invocations strictly by structural hierarchy. Root invocations (`ParentID == nil`) exclusively govern `OrchestratorPeakTokens` and `OrchestratorSaturationPct` regardless of persona role string (`general`, `build`, `orchestrator`, or custom). Delegated child invocations (`ParentID != nil`) exclusively govern `SubagentsPeakTokens`, `SubagentsPeakRole`, `RoleEffortDistributionPct`, and `SubagentTurnVelocity`. Empty child roles normalize to `"unassigned"`.
   - Single-invocation tickets without delegated subagents cleanly evaluate to `SubagentsPeakTokens = 0`, `SubagentsPeakRole = "none"`, empty `RoleEffortDistributionPct`, and zero-valued turn velocity.
   - `enums.go`: Refactored `OrderSubagentRoles` to accept candidate roles directly from child invocations without persona blacklists, prioritizing canonical roles (`planner`, `coder`, `tester`, `reviewer`) and ordering remaining custom roles alphabetically. Marked `IsSubagentRole` as deprecated in favor of structural classification (`ParentID != nil`).
   - `anomalies.go`: In `detectSubagentReworkLoops`, mapped `EventFact.SessionID` to `InvocationFact.ParentID != nil` so that root coordination turns are structurally ignored during ping-pong cycle evaluation.
2. **Boundary Adapters (`backend/internal/adapters/`)**:
   - `formatters.go`: Updated `FormatMarkdownMetricsTable` to render `RoleEffortDistributionPct` directly without persona exclusion filters, formatting empty subagent distributions as `N/A`.
   - `parity_test.go`: Aligned test baseline fixtures for tickets with authentic `"general"` root sessions and updated fleet role costs to reflect authentic split between orchestrator and general personas. Added `TestStructuralClassificationParityOnWorkspaces` verifying 100% compliance across all 37 workspace ledgers.

### References
- Contract: `workspaces/AL-31/CONTRACT.md`
- Detailed Summary: `workspaces/AL-31/SUMMARY.md`
