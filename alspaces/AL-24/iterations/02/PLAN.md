# Implementation Plan: AL-24 Raw Fact Preservation and Configured Role Mapping Transformation

## 1. Goal and Iteration Scope

Iteration 02 refines the extraction, domain transformation, and presentation pipeline for ticket AL-24:
1. Preserving raw agent personas in OpenCode SQLite fact extraction: The extractor must record the real, unmutated persona string from session.agent into InvocationFact.Role and EventFact.Role (such as build, coder, or custom persona names), rather than prematurely coercing root sessions or sessions with agent build into the canonical orchestrator role. When saved to agentic_metrics.json, the raw fact ledger contains pure, unmutated data.
2. Applying role_mappings from .agenticlens/config.yaml during domain transformation and presentation: In the functional core, domain.DeriveMetricsLedger must apply RoleMappings.ResolveRole during metric derivation to map raw agent personas to canonical or configured domain roles. Derived metrics (including role metrics, workflow efficiency distributions, subagent turn velocity, context window health, and thread summaries) must reflect the mapped roles.
3. Presentation alignment:
   - The markdown summary table in SUMMARY.md generated via FormatMarkdownMetricsTable reflects the mapped roles (for example, build rolled up under orchestrator).
   - The web dashboard (served via the HTTP API /api/workspaces and /api/workspaces/{ticketId}) receives the mapped rollups in the top-level metrics ledger while preserving the untouched raw data in the facts payload.
4. Scope Boundaries: UI layout and styling changes are explicitly excluded from iteration 02. The focus is strictly backend fact extraction, pure domain transformation with role mappings, and presentation formatting.

## 2. Grounded Reference Architecture & File Paths

The implementation touches the following files across domain, adapters, and CLI entrypoint:

- OpenCode Fact Extraction Boundary:
  - backend/internal/adapters/opencode/extractor.go: In function ExtractFacts, remove the hardcoded conditions that force root sessions or sessions with agent build to orchestrator. Extract and assign the real raw string from session.agent into InvocationFact.Role and EventFact.Role.
  - backend/internal/adapters/opencode/extractor_test.go: Update extraction unit tests to verify that raw session agent values (such as build) are preserved in invocation and event facts rather than coerced to orchestrator.

- Functional Core Domain Layer:
  - backend/internal/domain/config.go: Houses the pure RoleMappings type and ResolveRole method mapping raw agent personas to canonical AgentRole values.
  - backend/internal/domain/aggregations.go: Update DeriveMetricsLedger to accept optional RoleMappings. During transformation, map invocation and event roles using ResolveRole to compute totals, role metrics, workflow efficiency, and thread summaries, while keeping the untouched raw AgenticFactLedger in the Facts field.
  - backend/internal/domain/aggregations_test.go: Add pure unit tests asserting DeriveMetricsLedger transforms raw roles into mapped aggregates while keeping raw facts unmutated.

- Imperative Shell Repository and CLI Pipeline:
  - backend/internal/adapters/config/loader.go: Load project configuration from .agenticlens/config.yaml to obtain RoleMappings.
  - backend/internal/adapters/repository.go: In ScanWorkspaces and LoadTicketWorkspace, load project configuration and pass RoleMappings to DeriveMetricsLedger.
  - backend/cmd/agentic-lens/main.go: In runGenerate, load project configuration, derive metrics with RoleMappings, write raw facts to agentic_metrics.json, and format the markdown summary with the mapped metrics.
  - backend/internal/adapters/repository_test.go: Test that workspace loading applies configured role mappings during metrics derivation.
  - backend/cmd/agentic-lens/main_test.go: Test that generate command writes raw data to JSON while outputting mapped roles in markdown summaries.
  - backend/internal/adapters/formatters.go: Verify FormatMarkdownMetricsTable accurately renders mapped metrics.

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Raw Fact Preservation in OpenCode Extractor (backend/internal/adapters/opencode/)

- Task 1.1: Refactor session agent extraction in backend/internal/adapters/opencode/extractor.go:
  - In ExtractFacts step 3 (message processing loop), inspect the session agent string (s.agent).
  - If strings.TrimSpace(s.agent) is non-empty, use the exact raw trimmed string as the session role in sessionRoleMap (e.g. build, custom-agent, orchestrator).
  - Do not check whether s.parentID is nil or whether s.agent equals build.
  - If strings.TrimSpace(s.agent) is completely empty or null, provide a safe fallback role string (orchestrator for root session if empty, unassigned for child).
  - Populate InvocationFact.Role and EventFact.Role using this raw extracted string.
  - Verify zero role mapping or enum coercion occurs during fact extraction.
- Task 1.2: Update backend/internal/adapters/opencode/extractor_test.go:
  - Update TestExtractFacts and related test scenarios where a root session was created with agent build.
  - Assert that rootInv.Role is build, not orchestrator.
  - Assert that turn events from the root session have Role equal to build.

### Phase 2: Domain Transformation with Role Mappings (backend/internal/domain/)

- Task 2.1: Update DeriveMetricsLedger in backend/internal/domain/aggregations.go:
  - Expand the function signature to accept variadic role mappings: DeriveMetricsLedger(factLedger AgenticFactLedger, roleMappings ...RoleMappings) AgenticMetricsLedger.
  - If roleMappings is supplied and non-empty, use the provided mappings; otherwise fallback to DefaultProjectConfig().RoleMappings.
  - For aggregation calculations, construct a mapped copy of invocations where each invocation role is replaced by string(mappings.ResolveRole(inv.Role)).
  - For workflow efficiency calculations, construct a mapped copy of the fact ledger where invocations use the mapped roles.
  - Pass the mapped invocations into ComputeTotals, ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, ComputeThreadSummaries, and DetectAnomalies.
  - In the returned AgenticMetricsLedger, assign the original, untouched factLedger to the Facts field so raw data is preserved.
  - Invocations whose raw role maps to RoleOrchestrator are grouped under orchestrator and excluded from subagent velocity and effort distribution.
- Task 2.2: Add pure domain unit tests in backend/internal/domain/aggregations_test.go:
  - Create a fact ledger with raw roles build, coder, and custom-qa.
  - Configure RoleMappings with build mapped to orchestrator and custom-qa mapped to tester.
  - Assert DeriveMetricsLedger returns Roles containing orchestrator and tester.
  - Assert Facts.Invocations retains the original raw roles build and custom-qa.
  - Assert SubagentTurnVelocity and RoleEffortDistributionPct reflect the mapped subagent roles.

### Phase 3: Repository, Handlers, and CLI Pipeline Integration (backend/internal/adapters/ & backend/cmd/)

- Task 3.1: In backend/internal/adapters/repository.go:
  - In LoadTicketWorkspace, load project configuration using config.Load(basePath) (falling back to domain defaults on error). Pass cfg.RoleMappings into DeriveMetricsLedger.
  - In ScanWorkspaces, load project configuration once using config.Load(basePath) and pass cfg.RoleMappings into DeriveMetricsLedger for all scanned ledgers.
- Task 3.2: In backend/cmd/agentic-lens/main.go:
  - In runGenerate, load project configuration using config.Load("").
  - Pass cfg.RoleMappings into DeriveMetricsLedger(factLedger, cfg.RoleMappings).
  - Write raw factLedger to the path specified by --json-output via adapters.SaveFactLedger.
  - Generate the markdown summary table using the derived metricsLedger (which reflects mapped roles) and update the file specified by --append.
- Task 3.3: In backend/cmd/agentic-lens/main_test.go and backend/internal/adapters/repository_test.go:
  - Add a test confirming the generate command writes raw agent strings to the JSON output file and mapped role names to the summary markdown file.
  - Verify repository workspace scan produces mapped roles across tickets.

## 4. Adversarial Testing Strategy for Tester

### Extractor Boundary Tests (backend/internal/adapters/opencode/extractor_test.go)
- Test Case 1: Raw Persona Preservation
  - Insert sessions into SQLite with agent set to build for the root session and security-scanner for a child session.
  - Call ExtractFacts.
  - Assert root invocation has Role equal to build, not orchestrator.
  - Assert child invocation has Role equal to security-scanner.
  - Assert chronological turn events have Role matching the respective raw session agent.

### Pure Domain Transformation Tests (backend/internal/domain/aggregations_test.go)
- Test Case 2: Mapping Execution Invariance
  - Create a fact ledger where invocation 1 has Role build ($2.00) and invocation 2 has Role orchestrator ($1.00).
  - Apply RoleMappings mapping build to orchestrator.
  - Derive metrics ledger.
  - Assert Roles slice contains exactly one entry for orchestrator with invocation count 2 and total cost $3.00.
  - Assert Facts.Invocations still contains invocation 1 with Role build.
- Test Case 3: Subagent Classification Under Role Mapping
  - Create a fact ledger where an invocation has Role custom-bot ($1.50).
  - Test scenario A: custom-bot mapped to orchestrator. Assert custom-bot is excluded from RoleEffortDistributionPct and SubagentTurnVelocity.
  - Test scenario B: custom-bot mapped to coder. Assert coder is included in RoleEffortDistributionPct with 100% effort.
  - Test scenario C: custom-bot unmapped. Assert custom-bot is treated as a dynamic subagent and included in subagent metrics.

### End-to-End CLI Integration Tests (backend/cmd/agentic-lens/main_test.go)
- Test Case 4: JSON Raw vs Markdown Mapped Separation
  - Run generate CLI against a SQLite database containing a session with agent build.
  - Parse the generated agentic_metrics.json file: assert invocation role is build.
  - Read the generated summary markdown file: assert the table contains a row for orchestrator and no row for build.

### Repository Boundary Tests (backend/internal/adapters/repository_test.go)
- Test Case 5: Workspace Load with Configured Mappings
  - Write a test agentic_metrics.json with raw role build.
  - Call LoadTicketWorkspace with a project directory containing a config.yaml that maps build to orchestrator.
  - Assert the loaded metrics ledger has Roles containing orchestrator and Facts containing build.

## 5. Review Criteria for Reviewer

### Axis 1: Specification and Contract Alignment
- Does agentic_metrics.json store raw, unmapped agent personas extracted from the OpenCode database without coercing build or root to orchestrator?
- Does SUMMARY.md display canonical mapped roles (e.g. build mapped to orchestrator) according to role_mappings in .agenticlens/config.yaml?
- Does the web dashboard API (/api/workspaces and /api/workspaces/{ticketId}) return mapped roles across top-level summaries while preserving raw facts in the underlying ledger?
- Are UI layout and styling files completely untouched, maintaining zero scope creep?

### Axis 2: Engineering Standards and Technical Invariants
- FCIS Discipline: Is role mapping in DeriveMetricsLedger a pure domain transformation without I/O or global state?
- Immutability: Are the original fact ledger invocations and events in Facts preserved without in-place mutation?
- Strong Typing: Zero usage of any or interface{} across Go and TypeScript?
- Zero Emojis: Are code, comments, UI text, logs, and markdown files completely free of emojis?
- Quality Gates: Do make check and make test pass cleanly with zero warnings or errors?
