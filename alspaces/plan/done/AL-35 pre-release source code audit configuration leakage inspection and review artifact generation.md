h3. Title: AL-35: Pre-Release Source Code Audit, Configuration Leakage Inspection, and Review Artifact Generation

*TARGET*: AgenticLens (review | domain | adapters | presentation | tests)

*CONTEXT*

As AgenticLens approaches its v1.0.0 first public release, the codebase requires a dedicated, rigorous architectural and code hygiene review using the provided `agentic-lens-review` skill. Because AgenticLens was incubated within this specific repository, repository-specific conventions and hardcoded assumptions may be leaking into source code, domain adapters, and frontend views (e.g. hardcoded folder prefixes `src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/` in `analyzer.go`, hardcoded `workspaces/` in `classifier.go`, hardcoded base branch `origin/main`, lack of `/api/config` dynamic exposure, and static branding assumptions). Additionally, the test suite requires systematic profiling to detect tests exceeding 100ms, identify artificial sleeps, and diagnose execution bottlenecks.

*GOAL*

Execute a comprehensive, read-only pre-release code review using `agentic-lens-review`, specifically auditing for configuration leakage and profiling test suite execution, delivering a detailed Markdown audit report (`workspaces/AL-35/AUDIT_REPORT.md`) along with actionable, prioritized follow-up task artifacts. No modifications to application code or test suites are permitted in this ticket.

*VERTICAL SLICE SCOPE*

* Multi-Tier Audit via `agentic-lens-review`:
  - Tier 1: Layering & Boundary Discipline (Go `internal/domain` pure FCIS isolation, zero I/O imports, zero clocks, stateless adapters).
  - Tier 2: Module Depth & Dead Code (identify shallow passthroughs, speculative interfaces, unused structs, and subtraction opportunities).
  - Tier 3: Fowler Smells & Domain Modeling (primitive obsession, data clumps, missing discriminated unions, role/model mapping consistency).
  - Tier 4: Test Suite Performance Profiling (profiling test execution times, flagging tests exceeding 100ms, locating `time.Sleep`, subprocess calls, and disk churn).
* Project Configuration Leakage Deep-Dive:
  - Systematically scan backend, adapters, CLI, and frontend for hardcoded repository-specific strings, paths, and assumptions.
  - Audit git analyzer and classifier for hardcoded directory names (`backend/`, `web/`, `workspaces/`), inspecting against `ProjectConfig.Workspaces`.
  - Audit git base branch resolution for hardcoded `origin/main`.
  - Audit API endpoints and frontend views for missing dynamic project configuration exposure (`/api/config`), ensuring `AgenticLens` is preserved permanently as the platform name while allowing observed project identity to display alongside it.
* Review Artifacts & Follow-Up Decomposition:
  - Deliver comprehensive report at `workspaces/AL-35/AUDIT_REPORT.md`:
    - Executive summary of audited scope.
    - Test suite performance profiling diagnostics table.
    - Architecture and configuration leakage findings categorized by severity (High / Medium / Low) with affected files, concrete smell description, subtraction targets, and proposed target shapes.
    - Prioritized remediation roadmap.
  - Generate structured task artifacts for follow-up implementation tickets resulting from the audit findings.
* Invariant Check:
  - Strict review-only invariant: zero application code, test files, runtime scripts, or configuration files are modified.

*ACCEPTANCE CRITERIA*

* Complete codebase audit executed across Go backend, adapters, CLI subcommands, and TypeScript frontend using `agentic-lens-review`.
* Configuration leakage inspection thoroughly documents all hardcoded project-specific assumptions and paths.
* Test suite profiled with all tests exceeding 100ms identified and diagnosed.
* Detailed audit report authored at `workspaces/AL-35/AUDIT_REPORT.md` conforming to review skill standards.
* Actionable follow-up task specifications authored for subsequent implementation tickets.
* Strict review-only discipline maintained: zero lines of code modified in `backend/` or `web/src/`.
* Zero emojis across all generated review artifacts and task specifications.

*DEFINITION OF DONE*

* Review report and follow-up task artifacts authored under `workspaces/AL-35/`.
* Quality gates passing (`make check`, `make test`, `make build` with zero git diff in source code).
* Telemetry tested with `agentic-lens` CLI.
