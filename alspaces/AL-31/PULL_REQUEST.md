# refactor(domain): decouple root session structural hierarchy from persona roles (AL-31)

## Summary

Decouple structural session hierarchy (`ParentID == nil` vs `ParentID != nil`) from agent persona role names across the AgenticLens Go domain engine, boundary adapters, formatters, and parity tests. Completely eliminates `IsSubagentRole` and role string shims, de-hardcodes role resolution to support arbitrary personas, introduces configurable dynamic role ordering defaulting to alphabetical, and removes fragile ticket title inference.

### Key Changes

1. **Structural Hierarchy Decoupling**:
   - In `ComputeWorkflowEfficiency`, partition invocations strictly by structural hierarchy. Root invocations (`ParentID == nil`) exclusively govern `OrchestratorPeakTokens` and `OrchestratorSaturationPct` regardless of persona role string (`general`, `build`, `orchestrator`, or custom).
   - Delegated child invocations (`ParentID != nil`) exclusively govern `SubagentsPeakTokens`, `SubagentsPeakRole`, `RoleEffortDistributionPct`, and `SubagentTurnVelocity`. Empty child roles normalize to `"unassigned"`.
   - Single-session tickets without child invocations evaluate to `SubagentsPeakTokens = 0`, `SubagentsPeakRole = "none"`, and empty subagent distribution.
   - In `detectSubagentReworkLoops`, index session identifiers to `inv.ParentID != nil`. Root turns are structurally ignored during ping-pong cycle evaluation without fallback role heuristics.

2. **Generic Configurable Role Resolution**:
   - In `RoleMappings.ResolveRole`, deleted all hardcoded persona switch cases (`orchestrator`, `build`, `root`, `planner`, `coder`, `tester`, `reviewer`).
   - Cleaned empty or whitespace-only inputs return `RoleUnassigned` (`"unassigned"`).
   - Configured aliases resolve via exact, raw, and case-insensitive matching; arbitrary unmapped roles return `AgentRole(cleaned)` directly (e.g., `developer`, `researcher`, `architect`).

3. **Dynamic Role Ordering**:
   - Added `role_order` configuration to `.agenticlens/config.yaml` and loaded into `ProjectConfig.RoleOrder`.
   - In `OrderRoles` and `OrderSubagentRoles`, roles sort in ascending alphabetical order by default when custom order is omitted or empty, and honor custom priority sequence when provided.
   - Introduced `DeriveMetricsLedgerWithConfig(factLedger, cfg)` propagating `cfg.RoleOrder` to metrics aggregations and formatters.

4. **Fragile Title Inference Elimination**:
   - Completely deleted `InferTicketTitle`, `cleanHeadingTitle`, and `findWorkspaceFile` from `git/inference.go`, `main.go`, and test suites.
   - Active ticket titles are resolved strictly via explicit CLI flags or cached ledger facts.

5. **Hermetic Test Isolation**:
   - Added `createTestBranch` helper registering transient fixture branches at HEAD with automatic `t.Cleanup` removal, ensuring git branch inference tests pass deterministically in fresh CI checkouts.

### References
- Contract: `workspaces/AL-31/CONTRACT.md`
- Detailed Summary: `workspaces/AL-31/SUMMARY.md`
