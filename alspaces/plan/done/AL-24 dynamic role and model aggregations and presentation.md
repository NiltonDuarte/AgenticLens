h3. Title: AL-24: Dynamic Role and Model Aggregations and Presentation

*TARGET*: AgenticLens (domain | presentation | adapters | tests)

*CONTEXT*

AgenticLens currently hardcodes agent roles to a fixed five-member enumeration (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`). Any custom role present in invocation traces (such as `advisor`, `po`, `researcher`, `architect`, `security-auditor`) is filtered out or ignored when computing subagent turn velocity, peak context tokens, and role effort distribution. Similarly, models are assumed to conform to fixed sets. As multi-agent setups become increasingly specialized, the platform must dynamically discover, aggregate, and present arbitrary roles and models without data loss or truncation.

*GOAL*

Refactor backend aggregations and frontend visualizations to support arbitrary, dynamic agent roles and model names while preserving canonical role display precedence and ensuring zero data truncation.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Remove fixed loop constraints in `aggregations.go` that ignore roles outside `SubagentRoleOrder`.
  - Treat all non-orchestrator roles as subagents when deriving `SubagentsPeakTokens`, `SubagentsPeakRole`, `SubagentTurnVelocity`, and `RoleEffortDistributionPct`.
  - Preserve `CanonicalRoleOrder` for priority display of known roles, dynamically appending all discovered custom roles in alphabetical order.
  - Dynamically aggregate `ComputeModelMetrics` for all models encountered in session facts.
* Boundary Adapter:
  - Ensure DTO serialization accommodates arbitrary role and model string values without schema validation rejection.
* Presentation:
  - Widen `AgentRole` type in `web/src/domain/types.ts` from fixed union to open string type (`type AgentRole = string`).
  - Update `SaturationBar.tsx`, `DistributionCard.tsx`, `KPICard.tsx`, and `TicketDeepDive.tsx` to dynamically render custom roles and models.
  - Implement deterministic dynamic color hashing for custom role/model bars and badges.
* Quality:
  - Go unit tests verifying aggregation behavior with custom roles (`advisor`, `po`, `custom-subagent`) and non-standard model identifiers.
  - TypeScript tests asserting that custom roles render cleanly in `DistributionCard` and `SaturationBar` without crashing or NaN values.

*ACCEPTANCE CRITERIA*

* Non-canonical agent roles (e.g. `advisor`, `po`, `researcher`) are included in `role_costs`, `subagent_turn_velocity`, `role_effort_distribution_pct`, and `context_window_health`.
* Arbitrary model identifiers (e.g. `claude-3-7-sonnet`, `custom-fine-tuned-model`) are aggregated and displayed in `Model Cost Distribution`.
* Canonical roles maintain top display ordering, followed by custom roles sorted alphabetically.
* Frontend displays custom roles and models with consistent styling and accessible contrast.
* Zero `any` types introduced; all Go and TypeScript tests pass cleanly.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
