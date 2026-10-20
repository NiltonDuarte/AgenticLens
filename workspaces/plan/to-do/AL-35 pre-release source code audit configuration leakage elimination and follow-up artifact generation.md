h3. Title: AL-35: Pre-Release Source Code Audit, Configuration Leakage Elimination, and Follow-Up Artifact Generation

*TARGET*: AgenticLens (domain | adapters | presentation | cli | tests)

*CONTEXT*

As AgenticLens prepares for its v1.0.0 first public release, the codebase must undergo a thorough architectural and performance audit using the provided `agentic-lens-review` skill. Additionally, because the tool was incubated within this specific repository, repository-specific conventions and hardcoded assumptions have leaked into core components:
1. `backend/internal/adapters/git/analyzer.go` checks hardcoded directory prefixes (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`) and diffs against hardcoded `origin/main` rather than dynamically inspecting `ProjectConfig.Workspaces.SrcFolders` and repository git configuration.
2. `backend/internal/adapters/git/classifier.go` hardcodes `workspaces/` in universal ignore rules rather than referencing `cfg.Workspaces.Dir`.
3. The Go HTTP server lacks a `GET /api/config` endpoint, preventing client awareness of the observed project name, configured ticket prefixes, and workspace parameters.
4. The test suite needs systematic profiling to eliminate tests exceeding 100ms, banish artificial sleeps, and enforce microsecond execution in `internal/domain`.

*GOAL*

Execute a comprehensive pre-release audit with `agentic-lens-review`, eliminate all project-specific configuration leaks across backend adapters, expose `/api/config` in the HTTP server, display observed project identity alongside permanent AgenticLens platform branding, profile test suite bottlenecks, and deliver a structured audit report with actionable follow-up tasks.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Verify `internal/domain` remains strictly pure with zero I/O imports, zero clocks, and zero hardcoded repo assumptions.
  - Audit domain structs and enums for primitive obsession or missing sum types.
* Boundary Adapter:
  - `backend/internal/adapters/git/analyzer.go`: Replace hardcoded directory prefixes with dynamic evaluation of configured `SrcFolders`. Replace hardcoded `origin/main` with dynamic base branch resolution (`git symbolic-ref`, `origin/main`, `origin/master`, `main`, `master`).
  - `backend/internal/adapters/git/classifier.go`: Replace hardcoded `"workspaces/"` string with dynamic `cfg.Workspaces.Dir`.
  - `backend/internal/adapters/http/handlers.go` & `server.go`: Implement `GET /api/config` handler returning sanitized `ProjectConfig` DTO.
* Presentation:
  - `web/src/api/client.ts`: Add `fetchConfig()` client call with error handling and fallback defaults.
  - `web/src/components/Navbar.tsx`: Retain `AgenticLens` as the permanent platform brand. Render the configured `project_name` alongside the brand title (e.g. `AgenticLens` with an adjacent observed project pill `Observing: <project_name>`). Never remove or replace the platform name `AgenticLens`.
  - `web/src/views/LandingPage.tsx`: Retain `AgenticLens Observability` as the hero title, rendering dynamic observed target project context.
* Review Artifacts:
  - Produce comprehensive review report (`workspaces/AL-35/AUDIT_REPORT.md`) documenting test profiling findings, FCIS compliance, dead code subtraction targets, and prioritized follow-up tickets.
* Quality:
  - Go table-driven tests for `/api/config`, dynamic branch detection, and config-aware git classification.
  - Vitest tests for dynamic project name rendering in `Navbar.tsx` and API boundary handling.
  - All tests passing with `go test -race ./...` and `npm test`.

*ACCEPTANCE CRITERIA*

* Codebase audit completed adhering to `agentic-lens-review` multi-tier methodology.
* `GET /api/config` endpoint returns 200 OK with valid JSON matching `ProjectConfig`.
* Hardcoded source folder prefixes and `origin/main` replaced with dynamic configuration in `analyzer.go`.
* Universal ignore rules in `classifier.go` dynamically respect `cfg.Workspaces.Dir`.
* Web dashboard displays dynamic observed project name alongside permanent AgenticLens platform branding.
* Audit report artifact generated under `workspaces/AL-35/AUDIT_REPORT.md` with test profiling diagnostics and follow-up recommendations.
* `make check`, `make test`, and `make build` pass cleanly with zero `any` and zero emojis.

*DEFINITION OF DONE*

* Code, tests, and audit report merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
