# Iteration 01 Plan: Codebase, Test Suite, and Skills Refactoring (AL-16)

## Iteration Objective
Address the critical and high-priority friction items identified during the audit: decouple filesystem mutations from the formatters adapter, optimize single-ticket workspace lookup, optimize test suite execution and polling latency, align agent and skill path definitions with the actual directory structure and yaml config, and configure lightweight execution environments for pure frontend tests.

## Phase 1: Go Adapter Layering and Filesystem I/O Decoupling
1. In backend/internal/adapters/repository.go, implement LoadTicketWorkspace to load a single ticket's agentic_metrics.json, parse into domain.AgenticFactLedger, and derive domain.AgenticMetricsLedger.
2. In backend/internal/adapters/http/handlers.go, update WorkspaceDetailHandler to directly call LoadTicketWorkspace, eliminating the full directory scan of all workspaces for single-ticket lookups.
3. Relocate UpdateSummaryFile from backend/internal/adapters/formatters.go to backend/internal/adapters/repository.go.
4. Update callers in backend/internal/adapters/backfill/engine.go and corresponding tests to reference the repository method.
5. In backend/internal/adapters/formatters.go, ensure all remaining functions are pure string transformers with zero dependencies on the os package.

## Phase 2: Test Suite Performance Optimization and Concurrency
1. In backend/cmd/agentic-lens/main_test.go, revise the server polling wait loops across TestCLIServerSubcommand, TestCLIServerPortEnvironmentVariable, and TestCLIServerPortFlagPrecedence:
   - Attempt the initial HTTP GET immediately before entering sleep.
   - Reduce the retry polling sleep duration from 100 milliseconds to 10 milliseconds.
2. Add parallel execution declarations (t.Parallel()) to stateless table-driven tests in backend/internal/domain/aggregations_test.go, backend/internal/domain/anomalies_test.go, backend/internal/domain/config_test.go, backend/internal/domain/version_test.go, and backend/internal/adapters/git/analyzer_test.go.
3. In web/src/domain/progression.test.ts, web/src/domain/formatters.test.ts, and web/src/domain/events.test.ts, add per-file environment pragmas targeting the node environment rather than JSDOM.

## Phase 3: Agent Personas and Skill Definitions Modernization
1. In .opencode/agent/planner.md:
   - Remove hardcoded absolute project root references.
   - Update layer path definitions to include the backend/ prefix.
   - Update configuration file references from config.json to config.yaml.
2. In .opencode/skills/agentic-lens-dev/SKILL.md:
   - Remove absolute machine path references from description and preamble.
   - Update prompt templates and directory pointers to reflect the backend/ layout.
   - Update configuration references to .agenticlens/config.yaml.
3. In .opencode/agent/coder.md, tester.md, reviewer.md, and orchestrator.md:
   - Verify and update all layer paths to reference backend/internal/domain, backend/internal/adapters, and backend/cmd/agentic-lens.
   - Update any references to config.yaml.
   - Maintain the orchestrator non-execution mandate and mandatory user confirmation gate.

## Phase 4: Frontend Design Token Consolidation
1. Extract inline chart and saturation colors in web/src/views/MetricProgression.tsx and web/src/components/SaturationBar.tsx into a dedicated constants module or CSS custom properties.
2. Ensure consistent theme variable usage across metric charts.

## Verification Tasks
1. Execute backend static analysis and race detector suite:
   - Run go vet across all backend packages.
   - Run go test with race detection enabled across all backend packages.
2. Execute frontend verification:
   - Run typecheck in the web directory.
   - Run linter in the web directory.
   - Run test suite in the web directory.
3. Audit diff for zero emojis and verify absence of code blocks in planning documents.
