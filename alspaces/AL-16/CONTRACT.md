# Architectural Contract: Codebase, Test Suite, and Skills Refactoring (AL-16)

## Problem Statement
The AgenticLens application has reached functional maturity with complete coverage across multi-agent telemetry ingestion, CLI reporting, and TypeScript visualization. However, architectural friction, skill configuration drift, and test suite execution inefficiencies have accumulated across three distinct layers:
1. Architectural Layering: Filesystem I/O is coupled inside presentation formatters, and ticket detail queries incur full repository scans.
2. Skill and Agent Drift: Agent persona markdown files and development skills contain stale developer-specific absolute paths, missing backend directory prefixes, and outdated references to retired json config.
3. Test Performance: CLI server tests execute serial subprocesses with high unconditional sleep delays, pure Go tests lack parallel execution, and frontend pure domain tests run inside heavy JSDOM environments.

## Functional Core, Imperative Shell Invariants
1. Pure Core Invariance: All packages under backend/internal/domain must maintain zero external dependencies, zero I/O imports, and microsecond deterministic execution.
2. Boundary Discipline: All filesystem reads, writes, and mutations must reside exclusively within backend/internal/adapters/repository.go or dedicated repository writers. Presentation formatters must remain pure string functions taking value data and returning formatted text.
3. Targeted Data Access: Single-ticket lookups in the HTTP adapter must load the specific ticket workspace directly via repository functions rather than scanning all workspaces on disk.
4. Clock Injection: Any logic dependent on current time must accept an explicit timestamp parameter passed from outer HTTP handlers or CLI entrypoints.
5. Strict Zero-Code Invariance: Planning documents, architectural contracts, and iteration guides must contain zero code blocks.

## Backend Scope Boundaries
1. Refactor backend/internal/adapters/formatters.go to remove UpdateSummaryFile and relocate filesystem write operations to backend/internal/adapters/repository.go.
2. Implement LoadTicketWorkspace in backend/internal/adapters/repository.go and update WorkspaceDetailHandler in backend/internal/adapters/http/handlers.go to directly load the requested ticket path instead of executing a full ScanWorkspaces directory traversal.
3. Refactor backend/cmd/agentic-lens/main_test.go to eliminate unconditional 100ms sleep calls during server startup polling, replacing them with immediate attempts and low-latency retry intervals.
4. Introduce parallel test execution markers across stateless pure unit tests in backend/internal/domain/ and backend/internal/adapters/.

## Frontend Scope Boundaries
1. Tag pure mathematical domain tests in web/src/domain/ with node environment overrides to avoid unnecessary JSDOM initialization overhead.
2. Consolidate hardcoded hex color values across MetricProgression and SaturationBar into centralized CSS custom variables or typed token maps.
3. Ensure all TypeScript types maintain strict mode compliance with zero any declarations.

## OpenCode Agent and Skills Scope Boundaries
1. Remove all hardcoded local machine paths from .opencode/agent/planner.md and .opencode/skills/agentic-lens-dev/SKILL.md, replacing them with repository-relative references.
2. Align layer path references across .opencode/agent/planner.md, coder.md, tester.md, reviewer.md, and orchestrator.md to explicitly reflect the backend/ prefix (e.g., backend/internal/domain/, backend/internal/adapters/, backend/cmd/agentic-lens/).
3. Update configuration references to .agenticlens/config.yaml across agent personas and skill documentation.

## Non-Goals
1. No alterations to the public JSON schema of agentic_metrics.json.
2. No breaking changes to existing HTTP API response payloads (/api/fleet, /api/workspaces/{ticketID}).
3. No introduction of external caching dependencies or databases.

## Quality and Verification Gates
1. Go Static Analysis: Execution of go vet across backend packages must complete with zero warnings.
2. Go Race and Concurrency: Execution of go test with race detection across backend packages must pass all tests with reduced execution latency.
3. Frontend Type Safety and Linting: Execution of npm typecheck and lint must pass with zero warnings.
4. Frontend Test Suite: Execution of npm test must pass with verified node environment execution on pure domain tests.
5. Zero Emojis: Diff audit must show zero emojis across all modified files.
