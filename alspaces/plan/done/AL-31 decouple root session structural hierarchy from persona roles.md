h3. Title: AL-31: Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests

*TARGET*: AgenticLens (domain | adapters | tests)

*CONTEXT*

In AgenticLens, session telemetry is organized as a hierarchy: a root session coordinates the user workflow, and child invocations (`ParentID != nil`) represent delegated subagents (`planner`, `coder`, `tester`, `reviewer`).
However, the domain metrics engine and extraction adapters conflated session hierarchy (structural: `ParentID == nil` vs `ParentID != nil`) with agent persona (semantic: `general`, `orchestrator`, `coder`, etc.). Specifically, `domain.IsSubagentRole()` checked `role != "orchestrator"`.
When a user interacts directly with OpenCode, the root session is assigned the persona `"general"`. Because `"general" != "orchestrator"`, the root session was misclassified as a delegated subagent. Consequently:
1. Root session peak context tokens (e.g., 123.3k in AL-1) were attributed to `SubagentsPeakTokens` with `SubagentsPeakRole = "general"`.
2. Root coordination metrics (`OrchestratorPeakTokens`, `OrchestratorSaturationPct`) dropped to 0 because they looked up the literal key `"orchestrator"`.
3. Subagent effort distribution percentages were distorted because root session tokens diluted actual subagents.
4. Eight test cases failed in `backend/internal/adapters/parity_test.go` when evaluated against authentic workspace ledgers.

*GOAL*

Decouple structural session hierarchy from agent persona roles across the Go domain engine and adapters, ensuring that root invocations (`ParentID == nil`) exclusively drive root coordination metrics regardless of persona name (`general`, `orchestrator`, `build`, etc.), child invocations (`ParentID != nil`) exclusively drive subagent metrics, and `parity_test.go` test baselines match authentic root session calculations.

*VERTICAL SLICE SCOPE*

* Domain Core (`internal/domain/`):
  - Refactor `DeriveWorkflowEfficiency` and aggregation routines in `internal/domain/aggregations.go`:
    - Determine root vs subagent status strictly by structural hierarchy (`ParentID == nil` for root session; `ParentID != nil` for subagents).
    - Attribute root coordination peak tokens and saturation percentage from root invocations (`ParentID == nil`) rather than looking up fixed role name `"orchestrator"`.
    - Restrict `SubagentsPeakTokens`, `SubagentsPeakRole`, `subCosts`, `roleEffortDistributionPct`, and `SubagentTurnVelocity` strictly to invocations where `ParentID != nil`.
    - When no child invocations exist (`ParentID != nil`), set `SubagentsPeakTokens` to 0, `SubagentsPeakRole` to `"none"`, and `roleEffortDistributionPct` to empty map.
  - Refactor `IsSubagentRole` and `OrderSubagentRoles` in `internal/domain/enums.go`:
    - Decouple subagent classification from the `"orchestrator"` string check. Role ordering must handle any custom or standard role without treating root `"general"` sessions as subagents.
* Boundary Adapter (`internal/adapters/`):
  - Ensure `ExtractFacts` in `internal/adapters/opencode/extractor.go` and schema types in `internal/adapters/schemas.go` consistently preserve `ParentID` and authentic persona names without synthetic remapping.
  - Update `internal/adapters/formatters.go` to cleanly format root coordination metrics regardless of whether the root persona is `"general"`, `"orchestrator"`, or custom.
* Quality & Tests:
  - Table-driven unit tests in `internal/domain/aggregations_test.go` covering root sessions with diverse roles (`general`, `build`, `orchestrator`, `po`, and custom roles).
  - Verify single-invocation sessions (e.g. `GEN-1`, `PO-1`) produce 0 subagent peak tokens and `"none"` subagent peak role.
  - Update `internal/adapters/parity_test.go` baselines and fixtures to match authentic `"general"` root sessions and correct subagent attributions, ensuring all parity test cases pass.
  - Pass `go test -race ./...` and static checks with zero `any`.

*ACCEPTANCE CRITERIA*

* Invocations with `ParentID == nil` are never included in `SubagentsPeakTokens`, `SubagentsPeakRole`, or subagent effort distribution.
* Root sessions with persona `"general"`, `"build"`, or custom roles correctly report peak context tokens in `OrchestratorPeakTokens` (root coordination peak) and saturation percentage.
* Single-invocation sessions with no subagents report `SubagentsPeakTokens: 0` and `SubagentsPeakRole: "none"`.
* Subagent effort distribution percentages reflect strictly child invocations (`ParentID != nil`).
* All tests in `backend/internal/adapters/parity_test.go` and `backend/internal/domain/aggregations_test.go` pass.
* Zero `any` and zero emojis across all modified code and documentation.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
