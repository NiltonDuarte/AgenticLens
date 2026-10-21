# Implementation Plan: AL-36 Iteration 02 - Project-Level Workspace Artifacts Configuration, CLI Init Strictness, GitHub Release Location, and Default Config Elimination

## 1. Goal and Iteration 02 Scope

This delta plan addresses user steering feedback for Ticket AL-36 across Go backend architecture, CLI validation semantics, frontend artifact rendering, and release distribution infrastructure:

1. Dynamic Workspace Artifacts Configuration:
   Decouple workspace artifacts (CONTRACT.md, SUMMARY.md, PULL_REQUEST.md) from platform hardcoding. Elevate artifact definitions to configurable items under workspaces in project configuration. Propagate configured artifacts through domain structures, configuration loader, YAML writer, filesystem repository reader, HTTP workspace detail handler, and web dashboard ArtifactViewer and TicketDeepDive components.

2. Strict CLI Initialization Enforcement:
   Remove all fallbacks in agentic-lens init for project name, source folders, and ticket prefixes. Require explicit values for project-name, src-folders, and ticket-prefix. Terminate initialization with an exit code of 1 and an unambiguous error message if any of these mandatory options are absent.

3. GitHub Release Location Alignment:
   Re-target release asset retrieval scripts, documentation, and skill definitions to the canonical GitHub repository: https://github.com/NiltonDuarte/AgenticLens/releases.

4. Elimination of Production DefaultProjectConfig:
   Eliminate DefaultProjectConfig from production domain code entirely. Ensure that when no configuration file is found, the application halts with an explicit error directing the engineer to run agentic-lens init. Restrict project configuration fixtures to test-scoped helpers in test files.

5. Telemetry Fact Synchronization and Ticket Title:
   Assign the formal human-readable title to Ticket AL-36 in its session ledger and re-extract telemetry facts with the updated CLI binary.

---

## 2. Grounded Reference Architecture and Concrete File Paths

### Go Domain Functional Core (backend/internal/domain/)
- backend/internal/domain/config.go:
  - Define WorkspaceArtifactConfig struct containing Path, Title, and Description string fields.
  - Update WorkspacesConfig struct to include Artifacts as a slice of WorkspaceArtifactConfig.
  - Remove func DefaultProjectConfig() ProjectConfig completely from production domain code.
- backend/internal/domain/models.go:
  - Define TicketArtifactItem struct containing Path, Title, Description, and Content (pointer to string).
  - Update TicketArtifacts struct to store Items as a slice of TicketArtifactItem, retaining optional legacy pointer fields ContractContent and SummaryContent for backwards compatibility.
- backend/internal/domain/aggregations.go:
  - In DeriveMetricsLedgerWithConfig and DeriveMetricsLedger, replace calls to DefaultProjectConfig with explicit zero-value initialization and parameter passing.
- backend/internal/domain/config_test.go:
  - Introduce testProjectConfig helper function local to the test suite to supply test fixtures.
  - Replace all invocations of DefaultProjectConfig across test cases with testProjectConfig.

### Go Imperative Boundary Shell: Configuration Adapters (backend/internal/adapters/config/)
- backend/internal/adapters/config/loader.go:
  - Add rawWorkspaceArtifactConfig struct and include Artifacts field in rawWorkspacesConfig.
  - Define typed error ErrConfigNotFound indicating that .agenticlens-config.yaml was not located.
  - In Load function, return ErrConfigNotFound instead of returning a default configuration when no file is found.
  - In mergeConfig function, construct ProjectConfig from raw DTO without referencing DefaultProjectConfig.
- backend/internal/adapters/config/writer.go:
  - Remove all invocations of DefaultProjectConfig.
  - In GenerateConfigYAML, iterate over cfg.Workspaces.Artifacts and serialize structured YAML entries under workspaces.artifacts with descriptive comments.
  - In GenerateDefaultYAML, require explicit project name, source folders, and prefixes; construct ProjectConfig explicitly and populate default artifact configurations for CONTRACT.md, SUMMARY.md, and PULL_REQUEST.md.
- backend/internal/adapters/config/loader_test.go and writer_test.go:
  - Replace invocations of DefaultProjectConfig with local testProjectConfig helpers.
  - Add test coverage for missing configuration errors and workspace artifact YAML serialization/deserialization.

### Go Imperative Boundary Shell: Repository & HTTP API (backend/internal/adapters/)
- backend/internal/adapters/repository.go:
  - Update ReadTicketArtifacts to accept ticketID string, basePath string, and artifactConfigs slice of domain.WorkspaceArtifactConfig.
  - For each configured artifact, verify path containment within the ticket directory to prevent directory traversal attacks, stat the target, read content if present, and append a domain.TicketArtifactItem.
- backend/internal/adapters/http/handlers.go:
  - Update WorkspaceDetailHandler to accept configured artifact definitions, forwarding them to ReadTicketArtifacts.
- backend/internal/adapters/http/server.go:
  - Update Server struct to store configured artifact definitions.
  - Update NewServer constructor or add NewServerWithConfig accepting domain.ProjectConfig to pass configured artifacts to WorkspaceDetailHandler.
- backend/internal/adapters/http/server_test.go and repository_test.go:
  - Update tests to verify dynamic artifact loading and traversal safety.

### Go CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go:
  - In runInit:
    - Remove current directory fallback for projectName; fail if flag is empty.
    - Remove directory scanning fallback for srcFoldersFlag; fail if flag is empty.
    - Remove default prefix fallback for ticketPrefixFlag; fail if flag is empty.
    - Pass configured artifacts (CONTRACT.md, SUMMARY.md, PULL_REQUEST.md) into configuration generation.
  - In runGenerate, runServer, and runScan:
    - Check error returned by adaptersconfig.Load(""); if ErrConfigNotFound, print error to stderr advising the user to run agentic-lens init and exit with status 1.
  - In runServer:
    - Pass loaded projCfg into HTTP server initialization so configured workspace artifacts are available to the detail handler.
- backend/cmd/agentic-lens/main_test.go:
  - Update test cases to assert that runInit fails when project-name, src-folders, or ticket-prefix are missing.
  - Add test asserting that runGenerate, runServer, and runScan fail with exit 1 when run without a configuration file.

### TypeScript Presentation Layer (web/src/)
- web/src/domain/types.ts:
  - Define TicketArtifactItem interface with path, title, description, and nullable content string.
  - Update TicketArtifacts interface to include readonly items array of TicketArtifactItem alongside optional contract_content and summary_content fields.
- web/src/components/ArtifactViewer.tsx:
  - Replace static ArtifactTab union with dynamic tab selection driven by artifacts.items.
  - Render tab buttons using each artifact's title and indicate when content is empty.
  - Support Rich Document (markdown rendering) and Raw Source views for the currently active tab item.
  - Provide clipboard copy action for the currently active artifact content.
  - Display artifact description subtitle if configured.
  - Provide fallback rendering when artifacts or items are null or empty.
- web/src/views/TicketDeepDive.tsx:
  - Forward artifacts object from workspace detail response directly into ArtifactViewer.
- web/src/components/ArtifactViewer.test.tsx and Components.test.tsx:
  - Update test suites to cover multi-artifact dynamic tabs, empty state handling, copy actions, and markdown rendering.

### Release Downloader and Setup Skill Automation
- scripts/download-agentic-lens.sh and .opencode/skills/agentic-lens-setup/scripts/download.sh:
  - Update REPO variable to NiltonDuarte/AgenticLens.
  - Update user-facing error messages referencing release URLs.
- .opencode/skills/agentic-lens-setup/SKILL.md:
  - Update download URLs, repository references, and init flag explanations.
  - Emphasize that project-name, src-folders, and ticket-prefix are mandatory parameters.
- scripts/test-setup.sh:
  - Update test invocations to pass mandatory flags to agentic-lens init.

### Root Repository Configuration (.agenticlens-config.yaml)
- .agenticlens-config.yaml:
  - Add artifacts list under workspaces specifying CONTRACT.md (Technical Contract), SUMMARY.md (Work Summary), and PULL_REQUEST.md (Pull Request).

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Domain Modeling and Default Config Elimination
- Task 1.1: In backend/internal/domain/config.go, declare WorkspaceArtifactConfig with Path, Title, and Description fields. Add Artifacts field to WorkspacesConfig.
- Task 1.2: In backend/internal/domain/models.go, declare TicketArtifactItem with Path, Title, Description, and Content fields. Add Items field to TicketArtifacts.
- Task 1.3: Delete func DefaultProjectConfig() from backend/internal/domain/config.go.
- Task 1.4: In backend/internal/domain/aggregations.go, remove calls to DefaultProjectConfig from DeriveMetricsLedger and DeriveMetricsLedgerWithConfig.
- Task 1.5: In backend/internal/domain/config_test.go, create testProjectConfig() helper function and migrate all domain tests from DefaultProjectConfig to testProjectConfig.

### Phase 2: Configuration Loader and Writer Upgrades
- Task 2.1: In backend/internal/adapters/config/loader.go, add rawWorkspaceArtifactConfig struct and wire it into rawWorkspacesConfig. Define ErrConfigNotFound error. Update Load function to return ErrConfigNotFound when no config file is located. Update mergeConfig to remove references to DefaultProjectConfig.
- Task 2.2: In backend/internal/adapters/config/writer.go, eliminate DefaultProjectConfig calls. Add serialization logic for workspaces.artifacts. Update GenerateDefaultYAML to accept artifact configurations or default to CONTRACT.md, SUMMARY.md, and PULL_REQUEST.md.
- Task 2.3: In backend/internal/adapters/config/loader_test.go and writer_test.go, introduce testProjectConfig() helper and update test assertions to verify ErrConfigNotFound on missing files and valid round-tripping of workspace artifacts.
- Task 2.4: In backend/internal/adapters/live/sqlite_test.go and backend/internal/adapters/http/server_test.go, replace DefaultProjectConfig calls with testProjectConfig helpers.

### Phase 3: Strict CLI Init Validation and Missing Config Guardrails
- Task 3.1: In backend/cmd/agentic-lens/main.go inside runInit, check that projectNameFlag is non-empty after trimming; if empty, print an error and exit with status 1.
- Task 3.2: In runInit, check that srcFoldersFlag is non-empty after trimming; if empty, print an error and exit with status 1.
- Task 3.3: In runInit, check that ticketPrefixFlag is non-empty after trimming; if empty, print an error and exit with status 1.
- Task 3.4: In runGenerate, runServer, and runScan, inspect the error returned by adaptersconfig.Load(""). If an error is returned, print an error message stating that no configuration was found and instruct the user to run agentic-lens init, then exit with status 1.
- Task 3.5: In backend/cmd/agentic-lens/main_test.go, update existing init tests to supply mandatory flags, add tests verifying failure when each mandatory flag is omitted, and add test verifying that missing configuration exits with status 1 on generate/scan/server.

### Phase 4: Dynamic Repository Artifact Reader and HTTP Wiring
- Task 4.1: In backend/internal/adapters/repository.go, update ReadTicketArtifacts to accept configured artifact definitions. Iterate over configurations, validate relative path containment, read existing files, and populate TicketArtifactItem list.
- Task 4.2: In backend/internal/adapters/http/handlers.go, update WorkspaceDetailHandler to accept configured artifact definitions and pass them to ReadTicketArtifacts.
- Task 4.3: In backend/internal/adapters/http/server.go, store artifact configurations on Server and supply them when registering WorkspaceDetailHandler. Add NewServerWithConfig to accept domain.ProjectConfig.
- Task 4.4: In backend/cmd/agentic-lens/main.go inside runServer, pass loaded projCfg to NewServerWithConfig.
- Task 4.5: Update backend/internal/adapters/repository_test.go and http/server_test.go to test dynamic artifact reading and path traversal protections.

### Phase 5: TypeScript Frontend Dynamic Artifact Viewer
- Task 5.1: In web/src/domain/types.ts, define TicketArtifactItem and update TicketArtifacts.
- Task 5.2: In web/src/components/ArtifactViewer.tsx, implement dynamic tab selection using artifacts.items. Display artifact title, empty badge when content is null or whitespace, Rich Document and Raw Source views, copy action, and description tooltip or subtitle.
- Task 5.3: In web/src/views/TicketDeepDive.tsx, ensure detail.artifacts is passed cleanly to ArtifactViewer.
- Task 5.4: In web/src/components/ArtifactViewer.test.tsx, update tests to supply items array and verify multi-tab switching, empty tab states, markdown rendering, and clipboard actions.

### Phase 6: Release Automation and Script Alignment
- Task 6.1: Update REPO in scripts/download-agentic-lens.sh and .opencode/skills/agentic-lens-setup/scripts/download.sh to NiltonDuarte/AgenticLens.
- Task 6.2: Update .opencode/skills/agentic-lens-setup/SKILL.md to reflect the updated release repository URL and describe mandatory flags for agentic-lens init.
- Task 6.3: Update scripts/test-setup.sh to pass mandatory flags to agentic-lens init.
- Task 6.4: Update root .agenticlens-config.yaml to include configured artifacts under workspaces.

### Phase 7: Verification and Ticket Telemetry Synchronization
- Task 7.1: Run make check, make test, and make build to verify Go and TypeScript builds and test suites.
- Task 7.2: Update workspaces/AL-36/agentic_metrics.json with ticket_title set to "Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator" using the updated agentic-lens generate CLI.

---

## 4. Verification and Testing Strategy for Tester

The tester subagent must verify all components against edge cases and negative conditions:

1. CLI Init Strict Validation Tests:
   - Run agentic-lens init without --project-name: assert non-zero exit status and error message on stderr.
   - Run agentic-lens init without --src-folders: assert non-zero exit status and error message on stderr.
   - Run agentic-lens init without --ticket-prefix: assert non-zero exit status and error message on stderr.
   - Run agentic-lens init with all mandatory flags: assert success and valid generated YAML file.

2. Configuration File Requirement Tests:
   - In a directory without .agenticlens-config.yaml, execute agentic-lens generate, agentic-lens scan, and agentic-lens server: assert non-zero exit status and clear message directing user to run agentic-lens init.
   - Verify that no DefaultProjectConfig symbol exists in production Go binaries or packages.

3. Workspace Artifacts Configuration Tests:
   - Verify serialization and deserialization of workspaces.artifacts in .agenticlens-config.yaml.
   - Test ReadTicketArtifacts with custom artifact paths including nested relative paths, non-existent files, and directory traversal attempts (such as path containing parent directory markers). Assert that hostile paths are safely ignored.
   - Query GET /api/workspaces/{ticketId} and verify that the JSON response contains artifacts.items with correct path, title, description, and content.

4. Frontend Artifact Viewer Tests:
   - Verify rendering with multiple configured artifacts (e.g. CONTRACT.md, SUMMARY.md, PULL_REQUEST.md).
   - Verify tab switching updates visible content and active tab styling.
   - Verify that empty artifacts display an empty indicator and appropriate placeholder.
   - Verify that clipboard copy copies the content of the currently active artifact tab.
   - Verify that switching between Rich Document and Raw Source preserves the currently selected tab.

5. Release Script and Setup Tests:
   - Verify that scripts/download-agentic-lens.sh targets NiltonDuarte/AgenticLens.
   - Execute scripts/test-setup.sh and assert all tests pass.

---

## 5. Audit and Review Criteria for Reviewer

The reviewer subagent must evaluate the implementation along two strict axes:

- Axis 1: Spec and User Steering Alignment
  - Are ticket artifacts fully configurable via workspaces.artifacts in .agenticlens-config.yaml?
  - Do CONTRACT.md, SUMMARY.md, and PULL_REQUEST.md render as configured tabs in the ticket deep dive view rather than hardcoded platform assumptions?
  - Does agentic-lens init strictly require --project-name, --src-folders, and --ticket-prefix, failing without defaults?
  - Does the application fail cleanly with an actionable setup message when no configuration file exists?
  - Are all release downloader scripts and skill references pointing to https://github.com/NiltonDuarte/AgenticLens/releases?
  - Is DefaultProjectConfig completely eliminated from production source code?
  - Is the ticket title set in the telemetry ledger?

- Axis 2: Engineering Standards and Technical Invariants
  - Are pure functional core boundaries respected with zero I/O in internal/domain/?
  - Is directory traversal strictly prevented when reading configured artifact files?
  - Are Go and TypeScript free of any or interface{} types?
  - Are all source files, comments, docstrings, shell scripts, CLI output, and markdown artifacts 100% free of emojis?
  - Do make check, make test, and make build execute with zero errors?
